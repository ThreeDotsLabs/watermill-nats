package jetstream

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go/jetstream"
)

type ConsumerBuilder func(ctx context.Context, js jetstream.JetStream, topic string) (
	jetstream.Consumer,
	func(context.Context, watermill.LoggerAdapter),
	error)

type ConsumerNamer func(string, string) string

func defaultConsumerNamer(topic string, group string) string {
	if group != "" {
		return fmt.Sprintf("watermill__%s__%s", group, topic)
	}
	return fmt.Sprintf("watermill__%s", topic)
}

// ExistingConsumer is used to connect to a stream/consumer that already exist with the given topic name - it will not attempt creation of any broker-managed resources.
// It takes as an argument a function to transform the topic into a consumer name, passing nil will invoke the default behavior
// consumerName := fmt.Sprintf("watermill__%s", topic)
func ExistingConsumer(consumerNamer ConsumerNamer, group string) ConsumerBuilder {
	return func(ctx context.Context, js jetstream.JetStream, topic string) (jetstream.Consumer, func(context.Context, watermill.LoggerAdapter), error) {
		stream, err := js.Stream(ctx, topic)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get stream for topic %s: %w", topic, err)
		}

		var consumerName string
		if consumerNamer == nil {
			consumerName = fmt.Sprintf("watermill__%s", topic)
		} else {
			consumerName = consumerNamer(topic, group)
		}

		consumer, err := stream.Consumer(ctx, consumerName)

		return consumer, nil, err
	}
}

// GroupedConsumer builds a callback to create a consumer in the given group.
// The closing function is not returned since a single subscription in the group cannot know when the backing consumer should be deleted.
func GroupedConsumer(groupName string) ConsumerBuilder {
	return func(ctx context.Context, js jetstream.JetStream, topic string) (jetstream.Consumer, func(context.Context, watermill.LoggerAdapter), error) {
		consumer, _, err := createOrUpdateConsumerWithCloser(ctx, js, topic, groupName, defaultConsumerNamer)
		return consumer, nil, err
	}
}

// EphemeralConsumer builds a callback to create a consumer, returning a function that will be used to delete the broker-managed consumer.
func EphemeralConsumer() ConsumerBuilder {
	return func(ctx context.Context, js jetstream.JetStream, topic string) (jetstream.Consumer,
		func(context.Context, watermill.LoggerAdapter),
		error) {
		return createOrUpdateConsumerWithCloser(ctx, js, topic, watermill.NewShortUUID(), defaultConsumerNamer)
	}
}

func createOrUpdateConsumerWithCloser(ctx context.Context, js jetstream.JetStream, topic, group string, namer ConsumerNamer) (
	jetstream.Consumer,
	func(context.Context, watermill.LoggerAdapter),
	error) {
	stream, err := js.Stream(ctx, topic)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stream for topic %s: %w", topic, err)
	}

	name := namer(topic, group)

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:      name,
		AckPolicy: jetstream.AckExplicitPolicy,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get consumer for topic %s: %w", topic, err)
	}

	return consumer, func(cctx context.Context, logger watermill.LoggerAdapter) {
		deleteErr := stream.DeleteConsumer(cctx, name)
		if deleteErr != nil {
			logger.Error("failed to delete consumer", deleteErr, watermill.LogFields{})
		}
	}, err
}

// consume manages callback-based consumption of data from the provided JetStream consumer
func consume(ctx context.Context,
	closing chan struct{},
	consumer jetstream.Consumer,
	cb handleFunc,
	deferred func(),
) (chan *message.Message, error) {
	output := make(chan *message.Message)

	// TODO: this is the closest analog to callback based subscriptions in watermill-nats/pkg/nats
	// add support for batching pull consumers using consumer.Fetch / FetchNoWait
	cc, err := consumer.Consume(func(msg jetstream.Msg) {
		cb(ctx, msg, output)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start jetstream consumer: %w", err)
	}

	go monitor(ctx, closing, output, func() {
		defer deferred()
		cc.Stop()
	})

	return output, nil
}

// monitor will watch for context / subscriber closure, close the output channel and apply any other necessary cleanup.
func monitor(ctx context.Context,
	closing chan struct{},
	output chan *message.Message,
	after func()) {
	defer after()
	select {
	case <-closing:
		//unblock
	case <-ctx.Done():
		//unblock
	}
	close(output)
}
