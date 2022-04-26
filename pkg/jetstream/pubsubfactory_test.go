package jetstream_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/jetstream"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats/wmpb"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func getTestFeatures() tests.Features {
	return tests.Features{
		ConsumerGroups:                      true,
		ExactlyOnceDelivery:                 false,
		GuaranteedOrder:                     true,
		GuaranteedOrderWithSingleSubscriber: true,
		Persistent:                          true,
		RequireSingleInstance:               false,
		NewSubscriberReceivesOldMessages:    true,
	}
}

func newPubSub(t *testing.T, clientID string, queueName string, exactlyOnce bool) (message.Publisher, message.Subscriber) {
	trace := os.Getenv("WATERMILL_TEST_NATS_TRACE")
	debug := os.Getenv("WATERMILL_TEST_NATS_DEBUG")

	format := os.Getenv("WATERMILL_TEST_NATS_FORMAT")

	var marshaler jetstream.MarshalerUnmarshaler

	switch strings.ToLower(format) {
	case "nats":
		marshaler = &jetstream.NATSMarshaler{}
	case "proto":
		marshaler = &wmpb.NATSMarshaler{}
	case "json":
		marshaler = &jetstream.JSONMarshaler{}
	default:
		marshaler = &jetstream.GobMarshaler{}
	}

	logger := watermill.NewStdLogger(strings.ToLower(debug) == "true", strings.ToLower(trace) == "true")

	natsURL := os.Getenv("WATERMILL_TEST_NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	options := []nats.Option{
		nats.RetryOnFailedConnect(true),
		nats.Timeout(30 * time.Second),
		nats.ReconnectWait(1 * time.Second),
	}

	subscriberCount := 1

	if queueName != "" {
		subscriberCount = 2
	}

	subscribeOptions := []nats.SubOpt{
		nats.DeliverAll(),
		nats.AckExplicit(),
	}

	c, err := nats.Connect(natsURL, options...)
	require.NoError(t, err)

	defer c.Close()

	jetstreamOptions := make([]nats.JSOpt, 0)

	_, err = c.JetStream()
	require.NoError(t, err)

	pub, err := jetstream.NewPublisher(jetstream.PublisherConfig{
		URL:              natsURL,
		Marshaler:        marshaler,
		NatsOptions:      options,
		JetstreamOptions: jetstreamOptions,
		AutoProvision:    true,
		TrackMsgId:       exactlyOnce,
	}, logger)
	require.NoError(t, err)

	sub, err := jetstream.NewSubscriber(jetstream.SubscriberConfig{
		URL:              natsURL,
		QueueGroup:       queueName,
		DurableName:      queueName,
		SubscribersCount: subscriberCount, //multiple only works if a queue group specified
		AckWaitTimeout:   30 * time.Second,
		Unmarshaler:      marshaler,
		NatsOptions:      options,
		SubscribeOptions: subscribeOptions,
		JetstreamOptions: jetstreamOptions,
		CloseTimeout:     30 * time.Second,
		AutoProvision:    false, // tests use SubscribeInitialize
		AckSync:          exactlyOnce,
	}, logger)
	require.NoError(t, err)

	return pub, sub
}

func createPubSub(t *testing.T) (message.Publisher, message.Subscriber) {
	return newPubSub(t, watermill.NewUUID(), "", false)
}

func createPubSubWithConsumerGroup(t *testing.T, consumerGroup string) (message.Publisher, message.Subscriber) {
	return newPubSub(t, watermill.NewUUID(), consumerGroup, false)
}

//nolint:deadcode,unused
func createPubSubWithExactlyOnce(t *testing.T) (message.Publisher, message.Subscriber) {
	return newPubSub(t, watermill.NewUUID(), "", true)
}

//nolint:deadcode,unused
func createPubSubWithConsumerGroupWithExactlyOnce(t *testing.T, consumerGroup string) (message.Publisher, message.Subscriber) {
	return newPubSub(t, watermill.NewUUID(), consumerGroup, true)
}
