package msg

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type PBMarshaler struct{}

func (*PBMarshaler) Marshal(topic string, m *message.Message) (*nats.Msg, error) {
	pbMsg := &Message{
		Uuid:     m.UUID,
		Metadata: m.Metadata,
		Payload:  m.Payload,
	}

	data, err := proto.Marshal(pbMsg)

	if err != nil {
		return nil, err
	}

	natsMsg := nats.NewMsg(PublishSubject(topic, m.UUID))
	natsMsg.Data = data

	return natsMsg, nil
}

func (*PBMarshaler) Unmarshal(msg *nats.Msg) (*message.Message, error) {
	pbMsg := &Message{}

	err := proto.Unmarshal(msg.Data, pbMsg)

	if err != nil {
		return nil, err
	}

	wmMsg := message.NewMessage(pbMsg.GetUuid(), pbMsg.GetPayload())
	wmMsg.Metadata = pbMsg.GetMetadata()

	return wmMsg, nil
}
