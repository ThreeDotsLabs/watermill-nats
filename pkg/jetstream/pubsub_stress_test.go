//go:build stress
// +build stress

package jetstream_test

import (
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/jetstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
	"github.com/nats-io/nats.go"
)

func TestPublishSubscribe_stress(t *testing.T) {
	tests.TestPubSubStressTest(
		t,
		tests.Features{
			ConsumerGroups:                      false,
			ExactlyOnceDelivery:                 false,
			GuaranteedOrder:                     true,
			GuaranteedOrderWithSingleSubscriber: true,
			Persistent:                          true,
			RestartServiceCommand:               nil,
			RequireSingleInstance:               false,
			NewSubscriberReceivesOldMessages:    true,
		},
		func(t *testing.T) (message.Publisher, message.Subscriber) {
			url, logger := nats.DefaultURL, watermill.NewStdLogger(true, true)
			p, err := jetstream.NewPublisher(&jetstream.PublisherConfig{URL: url, Logger: logger})
			if err != nil {
				t.Fatalf("creating publisher: %s", err)
			}
			s, err := jetstream.NewSubscriber(&jetstream.SubscriberConfig{URL: url, Logger: logger, AckWaitTimeout: 1 * time.Second})
			if err != nil {
				t.Fatalf("creating subscriber: %s", err)
			}
			return p, s
		},
		nil,
	)
}
