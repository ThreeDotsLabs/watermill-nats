package jetstream

import (
	"context"
	"fmt"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	wmnats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var _ message.Publisher = &Publisher{}

// Publisher provides a watermill publisher interface to NATS JetStream
type Publisher struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	m      wmnats.Marshaler
	logger watermill.LoggerAdapter
	known  sync.Map
}

// NewPublisher creates a new watermill JetStream publisher.
func NewPublisher(config *PublisherConfig) (*Publisher, error) {
	config.setDefaults()

	nc, err := nats.Connect(config.URL)

	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return newPublisher(nc, config)
}

func newPublisher(nc *nats.Conn, config *PublisherConfig) (*Publisher, error) {
	jsapi, err := jetstream.New(nc)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize jetstream: %w", err)
	}

	return &Publisher{
		nc:     nc,
		js:     jsapi,
		m:      &wmnats.NATSMarshaler{},
		logger: config.Logger,
		known:  sync.Map{},
	}, nil
}

// Publish sends provided watermill messages to the given topic.
func (p *Publisher) Publish(topic string, messages ...*message.Message) error {
	for _, m := range messages {
		// TODO: how can we handle eg routing metadata without fallback to *nats.Msg
		nm, err := p.m.Marshal(topic, m)
		if err != nil {
			return fmt.Errorf("failed to marshal: %w", err)
		}

		_, err = p.js.PublishMsg(context.Background(), nm) //marshal
		if err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
	}

	return nil
}

// Close closes the publisher and its underlying connection
func (p *Publisher) Close() error {
	// TODO: if we support shared connections don't always close
	p.nc.Close()
	return nil
}
