package jetstream

import (
	wmnats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"
)

type Unmarshaler interface {
	Unmarshal(jetstream.Msg) (*message.Message, error)
}

type DefaultUnmarshaler struct{}

// Unmarshal turns a given JetStream message into a watermill message.
// Reserved NATS headers are not propagated in the resulting message Metadata.
func (DefaultUnmarshaler) Unmarshal(msg jetstream.Msg) (*message.Message, error) {
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
