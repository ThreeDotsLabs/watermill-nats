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
	nc              *nats.Conn
	js              jetstream.JetStream
	logger          watermill.LoggerAdapter
	closed          bool
	closing         chan struct{}
	ackWait         time.Duration
	outputsWg       *sync.WaitGroup
	closeTimeout    time.Duration
	subsLock        *sync.RWMutex
	consumerBuilder ConsumerBuilder
}

// NewSubscriber creates a new watermill JetStream subscriber.
func NewSubscriber(config *SubscriberConfig) (*Subscriber, error) {
	config.setDefaults()
	nc, err := nats.Connect(config.URL)
	if err != nil {
		return nil, err
	}
	return newSubscriber(nc, config)
}

func newSubscriber(nc *nats.Conn, config *SubscriberConfig) (*Subscriber, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("initializing jetstream: %w", err)
	}
	return &Subscriber{
		nc:              nc,
		js:              js,
		closing:         make(chan struct{}),
		logger:          config.Logger,
		ackWait:         config.AckWaitTimeout,
		outputsWg:       &sync.WaitGroup{},
		closeTimeout:    5 * time.Second,
		subsLock:        &sync.RWMutex{},
		consumerBuilder: config.ResourceInitializer,
	}, nil
}

// SubscribeInitialize offers a way to ensure the stream for a topic exists prior to subscribe
func (s *Subscriber) SubscribeInitialize(topic string) error {
	_, err := s.js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     topic,
		Subjects: []string{topic},
	})
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

type handleFunc = func(ctx context.Context, msg jetstream.Msg, output chan *message.Message)

// handleMsg provides the core processing logic for a single JetStream message
func (s *Subscriber) handleMsg(ctx context.Context, msg jetstream.Msg, output chan *message.Message) {
	m, err := unmarshal(msg)
	if err != nil {
		s.logger.Error("cannot unmarshal message", err, nil)
	}

	messageLogFields := watermill.LogFields{
		"ID":              m.UUID,
		"ReceivedSubject": msg.Subject(),
		"Len":             len(msg.Data()),
	}

	s.logger.Info("got msg", messageLogFields)

	ctx, cancelCtx := context.WithCancel(ctx)
	m.SetContext(ctx)
	defer cancelCtx()

	timeout := time.NewTimer(s.ackWait)
	defer timeout.Stop()

	s.subsLock.RLock()
	closed := s.closed
	s.subsLock.RUnlock()
	if closed {
		s.logger.Trace("Closed, message discarded", messageLogFields)
		return
	}

	select {
	case <-s.closing:
		s.logger.Trace("Closing, message discarded", messageLogFields)
		return
	case <-ctx.Done():
		s.logger.Trace("Context cancelled, message discarded", messageLogFields)
		return
	// if this is first can risk 'send on closed channel' errors
	case output <- m:
		s.logger.Trace("Message sent to consumer", messageLogFields)
	}

	select {
	case <-m.Acked():
		// TODO: enable async ack via config
		err = msg.DoubleAck(ctx)

		if err != nil {
			s.logger.Error("Cannot send ack", err, messageLogFields)
			return
		}
		s.logger.Trace("Message Acked", messageLogFields)
	case <-m.Nacked():
		//TODO: nak delay
		if err := msg.Nak(); err != nil {
			s.logger.Error("Cannot send nak", err, messageLogFields)
			return
		}
		s.logger.Trace("Message Nacked", messageLogFields)
		return
	case <-timeout.C:
		s.logger.Trace("Ack timeout", messageLogFields)
		return
	case <-s.closing:
		s.logger.Trace("Closing, message discarded before ack", messageLogFields)
		return
	case <-ctx.Done():
		s.logger.Trace("Context cancelled, message discarded before ack", messageLogFields)
		return
	}
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
