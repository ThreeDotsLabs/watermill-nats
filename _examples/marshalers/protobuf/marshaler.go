package protobuf

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type Marshaler struct{}

func (*Marshaler) Marshal(topic string, m *message.Message) (*nats.Msg, error) {
	pbMsg := &Message{
		Uuid:     m.UUID,
		Metadata: m.Metadata,
		Payload:  m.Payload,
	}

	data, err := proto.Marshal(pbMsg)

	if err != nil {
		return nil, err
	}

	natsMsg := nats.NewMsg(topic)
	natsMsg.Data = data

	return natsMsg, nil
}

func (*Marshaler) Unmarshal(msg *nats.Msg) (*message.Message, error) {
	pbMsg := &Message{}

	err := proto.Unmarshal(msg.Data, pbMsg)

	if err != nil {
		return nil, err
	}

	wmMsg := message.NewMessage(pbMsg.GetUuid(), pbMsg.GetPayload())
	wmMsg.Metadata = pbMsg.GetMetadata()

	return wmMsg, nil
}
