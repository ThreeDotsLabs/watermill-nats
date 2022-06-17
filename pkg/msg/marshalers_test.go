package msg_test

import (
	"crypto/rand"
	"fmt"
	"sync"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/msg"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type marshalerCase struct {
	name      string
	marshaler msg.MarshalerUnmarshaler
}

var marshalerCases = []marshalerCase{
	{"gob", &msg.GobMarshaler{}},
	{"json", &msg.JSONMarshaler{}},
	{"proto", &msg.PBMarshaler{}},
	{"nats", &msg.NATSMarshaler{}},
}

func TestMarshalers(t *testing.T) {
	for _, tc := range marshalerCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := sampleMessage(100)

			marshaler := tc.marshaler

			b, err := marshaler.Marshal("topic", msg)
			require.NoError(t, err)

			unmarshaledMsg, err := marshaler.Unmarshal(b)
			require.NoError(t, err)

			assert.True(t, msg.Equals(unmarshaledMsg))

			unmarshaledMsg.Ack()

			select {
			case <-unmarshaledMsg.Acked():
				// ok
			default:
				t.Fatal("ack is not working")
			}
		})
	}
}

func TestMarshalers_multiple_messages_async(t *testing.T) {
	for _, tc := range marshalerCases {
		t.Run(tc.name, func(t *testing.T) {

			messagesCount := 1000
			wg := sync.WaitGroup{}
			wg.Add(messagesCount)

			for i := 0; i < messagesCount; i++ {
				go func(msgNum int) {
					defer wg.Done()

					msg := sampleMessage(100)

					b, err := tc.marshaler.Marshal("topic", msg)
					require.NoError(t, err)

					unmarshaledMsg, err := tc.marshaler.Unmarshal(b)

					require.NoError(t, err)

					assert.True(t, msg.Equals(unmarshaledMsg))
				}(i)
			}

			wg.Wait()
		})
	}
}

func BenchmarkMarshalers(b *testing.B) {
	for _, plSize := range []int{100, 1000, 10000, 20000} {
		for _, tc := range marshalerCases {
			b.Run(fmt.Sprintf("%db-%s", plSize, tc.name), func(b *testing.B) {
				msg := sampleMessage(plSize)

				for i := 0; i < b.N; i++ {
					marshaled, err := tc.marshaler.Marshal("topic", msg)
					require.NoError(b, err)

					unmarshaledMsg, err := tc.marshaler.Unmarshal(marshaled)
					require.NoError(b, err)

					assert.True(b, msg.Equals(unmarshaledMsg))

					unmarshaledMsg.Ack()

					select {
					case <-unmarshaledMsg.Acked():
						// ok
					default:
						b.Fatal("ack is not working")
					}
				}
			})
		}
	}
}

func TestNatsMarshaler_Error_Multiple_Values_In_Single_Header(t *testing.T) {
	b := nats.NewMsg("fizz")
	b.Header.Add("foo", "bar")
	b.Header.Add("foo", "baz")

	marshaler := msg.NATSMarshaler{}

	_, err := marshaler.Unmarshal(b)

	require.Error(t, err)
}

func TestNatsMarshaler_Skips_Reserved_Headers(t *testing.T) {
	natsMsg := nats.NewMsg("fizz")
	natsMsg.Header.Add("foo", "bar")
	assert.Equal(t, 1, len(natsMsg.Header))

	unmarshaler := &msg.NATSMarshaler{}

	reserved := []string{
		nats.MsgIdHdr,
		nats.ExpectedLastMsgIdHdr,
		nats.ExpectedLastSeqHdr,
		nats.ExpectedLastSubjSeqHdr,
		nats.ExpectedStreamHdr,
		msg.WatermillUUIDHdr,
	}

	for _, v := range reserved {
		t.Run(v, func(t *testing.T) {
			assertReservedKey(t, natsMsg, v, unmarshaler)
		})
	}
}

func assertReservedKey(t *testing.T, natsMsg *nats.Msg, hdr string, unmarshaler *msg.NATSMarshaler) {
	natsMsg.Header.Add(hdr, uuid.NewString())
	defer natsMsg.Header.Del(hdr)
	assert.Equal(t, 2, len(natsMsg.Header))
	msg, err := unmarshaler.Unmarshal(natsMsg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(msg.Metadata))
}

func sampleMessage(plSize int) *message.Message {
	pl := make([]byte, plSize)

	_, _ = rand.Read(pl)

	msg := message.NewMessage(watermill.NewUUID(), pl)

	// add some metadata
	for i := 0; i < 10; i++ {
		msg.Metadata.Set(uuid.NewString(), uuid.NewString())
	}

	return msg
}
