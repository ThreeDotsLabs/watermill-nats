package nats_test

import (
	"testing"

	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
)

func TestPublishSubscribe(t *testing.T) {
	for name, factory := range map[string]pubSubFactory{
		"normal":       false,
		"exactly_once": true,
	} {
		t.Run(name, func(t *testing.T) {
			tests.TestPubSub(
				t,
				getTestFeatures(bool(factory)),
				factory.createPubSub,
				factory.createPubSubWithConsumerGroup,
			)
		})
	}
}
