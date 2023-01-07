package nats

import (
	"time"
)

// StopTime if this duration was returned, event will be term`ed
const StopTime = time.Duration(-1)

type Delay interface {
	// WaitTime return time.Duration that we need to wait.
	// retryNum is how many times WaitTime was called for
	// specific message
	WaitTime(retryNum uint64) time.Duration
}

// StaticDelay delay that always return the same time.Duration
type StaticDelay struct {
	Delay time.Duration
}

func NewStaticDelay(delay time.Duration) StaticDelay {
	return StaticDelay{Delay: delay}
}

func (s StaticDelay) WaitTime(retryNum int) time.Duration {
	return s.Delay
}
