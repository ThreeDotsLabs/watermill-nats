package nats_test

import (
	"testing"

	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
)

func TestPublishSubscribe(t *testing.T) {
	tests.TestPubSub(
		t,
		getTestFeatures(),
		createPubSub,
		createPubSubWithConsumerGroup,
	)
}
