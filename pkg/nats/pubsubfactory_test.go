package nats_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
	nc "github.com/nats-io/nats.go"
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

func getMarshaler(format string) nats.MarshalerUnmarshaler {
	switch strings.ToLower(format) {
	case "gob":
		return &nats.GobMarshaler{}
	case "json":
		return &nats.JSONMarshaler{}
	default:
		return &nats.NATSMarshaler{}
	}
}

func newPubSub(t *testing.T, clientID string, queueName string, exactlyOnce bool) (message.Publisher, message.Subscriber) {
	trace := os.Getenv("WATERMILL_TEST_NATS_TRACE")
	debug := os.Getenv("WATERMILL_TEST_NATS_DEBUG")

	format := os.Getenv("WATERMILL_TEST_NATS_FORMAT")
	marshaler := getMarshaler(format)

	logger := watermill.NewStdLogger(strings.ToLower(debug) == "true", strings.ToLower(trace) == "true")

	natsURL := os.Getenv("WATERMILL_TEST_NATS_URL")
	if natsURL == "" {
		natsURL = nc.DefaultURL
	}

	options := []nc.Option{
		nc.Name(clientID),
		nc.RetryOnFailedConnect(true),
		nc.Timeout(30 * time.Second),
		nc.ReconnectWait(1 * time.Second),
	}

	subscriberCount := 1

	if queueName != "" {
		subscriberCount = 2
	}

	subscribeOptions := []nc.SubOpt{
		nc.DeliverAll(),
		nc.AckExplicit(),
	}

	c, err := nc.Connect(natsURL, options...)
	require.NoError(t, err)

	defer c.Close()

	jetstreamOptions := make([]nc.JSOpt, 0)

	_, err = c.JetStream()
	require.NoError(t, err)

	jsConfig := nats.JetStreamConfig{
		Enabled:          true,
		AutoProvision:    false, // tests use SubscribeInitialize
		ConnectOptions:   jetstreamOptions,
		SubscribeOptions: subscribeOptions,
		PublishOptions:   nil,
		TrackMsgId:       exactlyOnce,
		AckAsync:         exactlyOnce,
		DurablePrefix:    queueName,
	}
	pub, err := nats.NewPublisher(nats.PublisherConfig{
		URL:         natsURL,
		Marshaler:   marshaler,
		NatsOptions: options,
		JetStream:   jsConfig,
		// SubjectCalculator: nats.DefaultSubjectCalculator("", queueName),
	}, logger)
	require.NoError(t, err)

	sub, err := nats.NewSubscriber(nats.SubscriberConfig{
		URL:              natsURL,
		QueueGroupPrefix: queueName,
		SubscribersCount: subscriberCount, //multiple only works if a queue group specified
		AckWaitTimeout:   30 * time.Second,
		Unmarshaler:      marshaler,
		NatsOptions:      options,
		CloseTimeout:     30 * time.Second,
		// SubjectCalculator: nats.DefaultSubjectCalculator(queueName, queueName),
		JetStream: jsConfig,
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
