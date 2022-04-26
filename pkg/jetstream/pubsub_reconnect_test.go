//go:build reconnect
// +build reconnect

package jetstream_test

import (
	"os"
	"testing"

	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
)

func TestPublishSubscribe_reconnect(t *testing.T) {
	features := getTestFeatures()

	containerName := "watermill-jetstream_nats_1" //default on linux
	if cn, found := os.LookupEnv("WATERMILL_TEST_NATS_CONTAINERNAME"); found {
		containerName = cn
	}

	// only provide this on reconnect test
	// the reconnect test itself will introduce a data race
	features.RestartServiceCommand = []string{"docker", "restart", containerName}

	tests.TestPubSub(
		t,
		features,
		createPubSub,
		createPubSubWithConsumerGroup,
	)
}
