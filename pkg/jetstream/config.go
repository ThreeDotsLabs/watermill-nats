// Package jetstream is in early stages and should not be considered a stable API at this time.
// Targeting a first stable release in v2.1
package jetstream

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
)

// PublisherConfig defines the watermill configuration for a JetStream publisher
type PublisherConfig struct {
	URL    string
	Logger watermill.LoggerAdapter
}

// setDefaults sets default values needed for a publisher if unset
func (p *PublisherConfig) setDefaults() {
	if p.Logger == nil {
		p.Logger = watermill.NewStdLogger(false, false)
	}
}

// SubscriberConfig defines the watermill configuration for a JetStream subscriber
type SubscriberConfig struct {
	URL                 string
	Logger              watermill.LoggerAdapter
	AckWaitTimeout      time.Duration
	ResourceInitializer ConsumerBuilder
	NakDelay            Delay
}

type ResourceInitializerOpt func(config *SubscriberConfig) ConsumerBuilder

// setDefaults sets default values needed for a subscriber if unset
func (s *SubscriberConfig) setDefaults() {
	if s.Logger == nil {
		s.Logger = watermill.NewStdLogger(false, false)
	}

	if s.AckWaitTimeout == 0*time.Second {
		s.AckWaitTimeout = 30 * time.Second
	}

	if s.ResourceInitializer == nil {
		s.ResourceInitializer = EphemeralConsumer()
	}
}
