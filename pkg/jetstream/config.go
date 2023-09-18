// Package jetstream is in early stages and should not be considered a stable API at this time.
// Targeting a first stable release in v2.1
package jetstream

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/nats-io/nats.go"
)

// PublisherConfig defines the watermill configuration for a JetStream publisher
type PublisherConfig struct {
	// URL is the path to the NATS jetstream-enabled broker
	URL string
	// Conn is an optional *nats.Conn that can be provided instead of using a watermill-managed connection to URL
	Conn *nats.Conn
	// Logger is a watermill logger (defaults to stdout with debug / trace disabled)
	Logger watermill.LoggerAdapter
	// ConfigureStream is a custom function that can be used to define stream configuration from a topic.  Publisher uses it to calculate publish destination from topic.
	ConfigureStream StreamConfigurator
}

// setDefaults sets default values needed for a publisher if unset
func (p *PublisherConfig) setDefaults() {
	if p.Logger == nil {
		p.Logger = watermill.NewStdLogger(false, false)
	}
	if p.ConfigureStream == nil {
		p.ConfigureStream = defaultStreamConfigurator
	}
}

// SubscriberConfig defines the watermill configuration for a JetStream subscriber
type SubscriberConfig struct {
	// URL is the path to the NATS jetstream-enabled broker
	URL string
	// Conn is an optional *nats.Conn that can be provided instead of using a watermill-managed connection to URL
	// TODO: should we expose jetstream here?  Currently need the NATS conn for graceful subscription shutdown:
	// https://github.com/nats-io/nats.go/issues/1328
	Conn *nats.Conn
	// Logger is a watermill logger (defaults to stdout with debug / trace disabled)
	Logger watermill.LoggerAdapter
	// AckWaitTimeout is how long watermill should wait for your application to finish processing a given message
	AckWaitTimeout time.Duration
	// ResourceInitializer is a custom function to turn a topic and consumer group into the necessary jetstream resources
	ResourceInitializer ResourceInitializer
	// NakDelay provides a delay function that can be used to delay reprocessing and eventually terminate
	NakDelay Delay
	// ConfigureStream is a custom function that can be used to define stream configuration from a topic.  Publisher uses it to calculate publish destination from topic.
	ConfigureStream StreamConfigurator
	// ConfigureConsumer is a custom function that can be used to define consumer configuration from a topic.  Publisher uses it to calculate publish destination from topic.
	ConfigureConsumer ConsumerConfigurator
}

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

	if s.ConfigureStream == nil {
		s.ConfigureStream = defaultStreamConfigurator
	}

	if s.ConfigureConsumer == nil {
		s.ConfigureConsumer = defaultConsumerConfigurator
	}
}
