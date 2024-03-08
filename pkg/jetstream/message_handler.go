package jetstream

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go/jetstream"
)

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
		if s.ackAsync {
			err = msg.Ack()
		} else {
			err = msg.DoubleAck(ctx)
		}

		if err != nil {
			s.logger.Error("Cannot send ack", err, messageLogFields)
			return
		}
		s.logger.Trace("Message Acked", messageLogFields)
	case <-m.Nacked():
		var nakDelay time.Duration

		if s.nakDelay != nil {
			metadata, err := msg.Metadata()
			if err != nil {
				s.logger.Error("Cannot parse nats message metadata, use nak without delay", err, messageLogFields)
			} else {
				nakDelay = s.nakDelay.WaitTime(metadata.NumDelivered)
				messageLogFields = messageLogFields.Add(watermill.LogFields{
					"delay":    nakDelay.String(),
					"retryNum": metadata.NumDelivered,
				})
			}
		}

		if nakDelay == TermSignal {
			if err := msg.Term(); err != nil {
				s.logger.Error("Cannot send term", err, messageLogFields)
			} else {
				s.logger.Trace("Message Termed via -1 NakDelay calculation", messageLogFields)
			}
			return
		}

		if nakDelay > 0 {
			if err := msg.NakWithDelay(nakDelay); err != nil {
				s.logger.Error("Cannot send nak", err, messageLogFields)
				return
			}
		} else {
			if err := msg.Nak(); err != nil {
				s.logger.Error("Cannot send nak", err, messageLogFields)
				return
			}
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
