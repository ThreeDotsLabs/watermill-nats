package nats

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var _ jetstream.Msg = &JetstreamCompat{}

func NewJetstreamCompat(nm *nats.Msg) *JetstreamCompat {
	return &JetstreamCompat{nm: nm}
}

type JetstreamCompat struct {
	nm *nats.Msg
}

func (j JetstreamCompat) Metadata() (*jetstream.MsgMetadata, error) {
	md, err := j.nm.Metadata()
	if err != nil {
		return nil, err
	}
	return &jetstream.MsgMetadata{
		Sequence: jetstream.SequencePair{
			Consumer: md.Sequence.Consumer,
			Stream:   md.Sequence.Stream,
		},
		NumDelivered: md.NumDelivered,
		NumPending:   md.NumPending,
		Timestamp:    md.Timestamp,
		Stream:       md.Stream,
		Consumer:     md.Consumer,
		Domain:       md.Domain,
	}, nil
}

func (j JetstreamCompat) Data() []byte {
	return j.nm.Data
}

func (j JetstreamCompat) Headers() nats.Header {
	return j.nm.Header
}

func (j JetstreamCompat) Subject() string {
	return j.nm.Subject
}

func (j JetstreamCompat) Reply() string {
	return j.nm.Reply
}

func (j JetstreamCompat) Ack() error {
	if j.nm.Reply == "" {
		return nil
	}
	return j.nm.Ack()
}

func (j JetstreamCompat) DoubleAck(ctx context.Context) error {
	if j.nm.Reply == "" {
		return nil
	}
	return j.nm.AckSync(nats.Context(ctx))
}

func (j JetstreamCompat) Nak() error {
	if j.nm.Reply == "" {
		return nil
	}
	return j.nm.Nak()
}

func (j JetstreamCompat) NakWithDelay(delay time.Duration) error {
	if j.nm.Reply == "" {
		return nil
	}
	return j.nm.NakWithDelay(delay)
}

func (j JetstreamCompat) InProgress() error {
	if j.nm.Reply == "" {
		return nil
	}
	return j.nm.InProgress()
}

func (j JetstreamCompat) Term() error {
	if j.nm.Reply == "" {
		return nil
	}
	return j.nm.Term()
}
