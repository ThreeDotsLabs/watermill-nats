package jetstream_test

import (
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/stretchr/testify/assert"
)

func TestStaticDelay(t *testing.T) {
	delay := time.Minute

	sd := nats.NewStaticDelay(delay)

	assert.Equal(t, sd.WaitTime(0), delay)
	assert.Equal(t, sd.WaitTime(1), delay)
	assert.Equal(t, sd.WaitTime(2000000), delay)
}

func TestMaxRetryDelay(t *testing.T) {
	delay := time.Minute
	maxRetry := uint64(5)

	sd := nats.NewMaxRetryDelay(delay, maxRetry)

	assert.Equal(t, sd.WaitTime(0), delay)
	assert.Equal(t, sd.WaitTime(4), delay)
	assert.Equal(t, sd.WaitTime(5), nats.TermSignal)
}
