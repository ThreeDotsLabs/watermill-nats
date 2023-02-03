//go:build stress
// +build stress

package nats_test

import (
	"testing"

	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
)

func TestPublishSubscribe_stress(t *testing.T) {
	factory := pubSubFactory(false)

	tests.TestPubSubStressTest(
		t,
		getTestFeatures(bool(factory)),
		factory.createPubSub,
		factory.createPubSubWithConsumerGroup,
	)
}
