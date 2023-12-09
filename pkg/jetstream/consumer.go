package jetstream

import (
	"context"
	"fmt"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go/jetstream"
)

type ResourceInitializer func(ctx context.Context, js jetstream.JetStream, topic string) (
	jetstream.Consumer,
	func(context.Context, watermill.LoggerAdapter),
	error)

type StreamConfigurator func(string) jetstream.StreamConfig

func defaultStreamConfigurator(topic string) jetstream.StreamConfig {
	return jetstream.StreamConfig{
		Name:     topic,
		Subjects: []string{topic},
	}
}

func workQueueStreamConfigurator(topic string) jetstream.StreamConfig {
	return jetstream.StreamConfig{
		Name:      topic,
		Subjects:  []string{topic},
		Retention: jetstream.WorkQueuePolicy,
	}
}

type ConsumerConfigurator func(string, string) jetstream.ConsumerConfig

func defaultConsumerConfigurator(topic string, group string) jetstream.ConsumerConfig {
	var name string
	if group != "" {
		name = fmt.Sprintf("watermill__%s__%s", group, topic)
	} else {
		name = fmt.Sprintf("watermill__%s", topic)
	}

	return jetstream.ConsumerConfig{Name: name, AckPolicy: jetstream.AckExplicitPolicy}
}

// ExistingConsumer is used to connect to a stream/consumer that already exist with the given topic name - it will not attempt creation of any broker-managed resources.
// It takes as an argument a function to transform the topic into a consumer name, passing nil will invoke the default behavior
// consumerName := fmt.Sprintf("watermill__%s", topic)
func ExistingConsumer(consumerNamer ConsumerConfigurator, group string) ResourceInitializer {
	return func(ctx context.Context, js jetstream.JetStream, topic string) (jetstream.Consumer, func(context.Context, watermill.LoggerAdapter), error) {
		streamConfig := defaultStreamConfigurator(topic)

		stream, err := js.Stream(ctx, streamConfig.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get stream for topic %s: %w", topic, err)
		}

		if consumerNamer == nil {
			consumerNamer = defaultConsumerConfigurator
		}

		consumerName := consumerNamer(topic, group).Name

		consumer, err := stream.Consumer(ctx, consumerName)

		return consumer, nil, err
	}
}

// GroupedConsumer builds a callback to create a consumer in the given group.
// The closing function is not returned since a single subscription in the group cannot know when the backing consumer should be deleted.
func GroupedConsumer(groupName string) ResourceInitializer {
	return func(ctx context.Context, js jetstream.JetStream, topic string) (jetstream.Consumer, func(context.Context, watermill.LoggerAdapter), error) {
		consumer, _, err := createOrUpdateConsumerWithCloser(ctx, js, topic, groupName, workQueueStreamConfigurator, defaultConsumerConfigurator)
		return consumer, nil, err
	}
}

// EphemeralConsumer builds a callback to create a consumer, returning a function that will be used to delete the broker-managed consumer.
func EphemeralConsumer() ResourceInitializer {
	return func(ctx context.Context, js jetstream.JetStream, topic string) (jetstream.Consumer,
		func(context.Context, watermill.LoggerAdapter),
		error) {
		return createOrUpdateConsumerWithCloser(ctx, js, topic, watermill.NewShortUUID(), defaultStreamConfigurator, defaultConsumerConfigurator)
	}
}

func createOrUpdateConsumerWithCloser(ctx context.Context,
	js jetstream.JetStream,
	topic, group string,
	configureStream StreamConfigurator,
	configureConsumer ConsumerConfigurator) (
	jetstream.Consumer,
	func(context.Context, watermill.LoggerAdapter),
	error) {
	streamConfig := configureStream(topic)
	stream, err := js.Stream(ctx, streamConfig.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stream for topic %s: %w", topic, err)
	}

	cfg := configureConsumer(topic, group)

	consumer, err := stream.CreateOrUpdateConsumer(ctx, cfg)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get consumer for topic %s: %w", topic, err)
	}

	return consumer, func(cctx context.Context, logger watermill.LoggerAdapter) {
		deleteErr := stream.DeleteConsumer(cctx, cfg.Name)
		if deleteErr != nil {
			logger.Error("failed to delete consumer", deleteErr, watermill.LogFields{})
		}
	}, err
}

// consume manages callback-based consumption of data from the provided JetStream consumer

// consume manages callback-based consumption of data from the provided JetStream consumer
func consume(ctx context.Context,
	s *Subscriber,
	consumer jetstream.Consumer,
	deferred func(),
) (chan *message.Message, error) {
	output := make(chan *message.Message)

	// TODO: this is the closest analog to callback based subscriptions in watermill-nats/pkg/nats
	// add support for batching pull consumers using consumer.Fetch / FetchNoWait
	cc, err := consumer.Consume(func(msg jetstream.Msg) {
		s.handleMsg(ctx, msg, output)
	}, s.consumeOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to start jetstream consumer: %w", err)
	}

	go monitor(ctx, s.closing, output, cc, s.messagesWG, deferred)

	return output, nil
}

// monitor will watch for context / subscriber closure, close the output channel and apply any other necessary cleanup.
func monitor(ctx context.Context,
	closing chan struct{},
	output chan *message.Message,
	consumeContext jetstream.ConsumeContext,
	messageWg *sync.WaitGroup,
	after func()) {
	select {
	case <-closing:
		//unblock
	case <-ctx.Done():
		//unblock
	}
	consumeContext.Stop()
	messageWg.Wait()
	close(output)
	after()
}
