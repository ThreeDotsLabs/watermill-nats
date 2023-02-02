//go:build reconnect
// +build reconnect

package nats_test

import (
	"os"
	"testing"

	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
)

func TestPublishSubscribe_reconnect(t *testing.T) {
	factory := pubSubFactory(false)

	features := getTestFeatures(bool(factory))

	containerName := "watermill-nats_nats_1" //default on linux
	if cn, found := os.LookupEnv("WATERMILL_TEST_NATS_CONTAINERNAME"); found {
		containerName = cn
	}

	// only provide this on reconnect test
	// the reconnect test itself will introduce a data race
	features.RestartServiceCommand = []string{"docker", "restart", containerName}

	tests.TestPubSub(
		t,
		features,
		factory.createPubSub,
		factory.createPubSubWithConsumerGroup,
	)
}
