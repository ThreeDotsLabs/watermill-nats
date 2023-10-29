package jetstream

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	wmnats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	watermillSync "github.com/ThreeDotsLabs/watermill/pubsub/sync"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"
)

var _ message.Subscriber = &Subscriber{}

// Subscriber provides a watermill subscriber interface to NATS JetStream
type Subscriber struct {
	nc                *nats.Conn
	js                jetstream.JetStream
	logger            watermill.LoggerAdapter
	closed            bool
	closing           chan struct{}
	ackWait           time.Duration
	outputsWg         *sync.WaitGroup
	closeTimeout      time.Duration
	subsLock          *sync.RWMutex
	consumerBuilder   ResourceInitializer
	nakDelay          Delay
	configureStream   StreamConfigurator
	configureConsumer ConsumerConfigurator
}

// NewSubscriber creates a new watermill JetStream subscriber.
// This middleware is currently considered an experimental / beta release - for production use
// it is recommended to use watermill-nats/pkg/nats.Subscriber with JetStream enabled.
func NewSubscriber(config SubscriberConfig) (*Subscriber, error) {
	config.setDefaults()

	nc := config.Conn
	if nc == nil {
		var err error
		nc, err = nats.Connect(config.URL)

		if err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
	}

	return newSubscriber(nc, &config)
}

func newSubscriber(nc *nats.Conn, config *SubscriberConfig) (*Subscriber, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("initializing jetstream: %w", err)
	}
	return &Subscriber{
		nc:                nc,
		js:                js,
		closing:           make(chan struct{}),
		logger:            config.Logger,
		ackWait:           config.AckWaitTimeout,
		outputsWg:         &sync.WaitGroup{},
		closeTimeout:      5 * time.Second,
		subsLock:          &sync.RWMutex{},
		consumerBuilder:   config.ResourceInitializer,
		configureStream:   config.ConfigureStream,
		configureConsumer: config.ConfigureConsumer,
	}, nil
}

// SubscribeInitialize offers a way to ensure the stream for a topic exists prior to subscribe
func (s *Subscriber) SubscribeInitialize(topic string) error {
	// TODO: how much should we allow customization here
	// do stream and consumer creator functions need to be separately overrideable?
	// or would config builders suffice?
	_, err := s.js.CreateStream(context.Background(), s.configureStream(topic))
	if err != nil {
		return fmt.Errorf("cannot initialize subscribe: %w", err)
	}
	return nil
}

// Subscribe establishes a JetStream subscription to the given topic.
func (s *Subscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	consumer, closer, err := s.consumerBuilder(ctx, s.js, topic)

	cleanup := func() {
		if closer != nil {
			defer closer(ctx, s.logger)
		}
		s.outputsWg.Done()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize jetstream consumer: %w", err)
	}

	s.outputsWg.Add(1)

	return consume(ctx, s.closing, consumer, s.handleMsg, cleanup)
}

// Close closes the subscriber and signals to close any subscriptions it created along with the underlying connection.
func (s *Subscriber) Close() error {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	close(s.closing)

	// TODO: if we support shared connections don't always close
	if err := s.nc.Drain(); err != nil {
		return fmt.Errorf("failed to drain connection: %w", err)
	}

	if watermillSync.WaitGroupTimeout(s.outputsWg, s.closeTimeout) {
		return fmt.Errorf("output wait group did not finish within alloted %s", s.closeTimeout.String())
	}

	return nil
}

// unmarshal turns a given JetStream message into a watermill message.
// Reserved NATS headers are not propagated in the resulting message Metadata.
func unmarshal(msg jetstream.Msg) (*message.Message, error) {
	data := msg.Data()

	hdr := msg.Headers()

	id := hdr.Get(wmnats.WatermillUUIDHdr)

	md := make(message.Metadata)

	for k, v := range hdr {
		switch k {
		case wmnats.WatermillUUIDHdr, nats.MsgIdHdr, nats.ExpectedLastMsgIdHdr, nats.ExpectedStreamHdr, nats.ExpectedLastSubjSeqHdr, nats.ExpectedLastSeqHdr:
			continue
		default:
			if len(v) == 1 {
				md.Set(k, v[0])
			} else {
				return nil, errors.Errorf("multiple values received in NATS header for %q: (%+v)", k, v)
			}
		}
	}

	wm := message.NewMessage(id, data)
	wm.Metadata = md

	return wm, nil
}
