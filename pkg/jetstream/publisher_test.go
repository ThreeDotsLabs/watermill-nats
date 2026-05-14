package jetstream_test

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/jetstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go"
	natsJS "github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublisher_TrackMsgId_SetsNatsMsgIdHeader(t *testing.T) {
	ctx := context.Background()

	topic := "watermill_test_" + watermill.NewShortUUID()

	ncMgmt, err := nats.Connect(nats.DefaultURL)
	require.NoError(t, err)
	defer ncMgmt.Close()

	js, err := natsJS.New(ncMgmt)
	require.NoError(t, err)

	_, err = js.CreateStream(ctx, natsJS.StreamConfig{
		Name:     topic,
		Subjects: []string{topic},
	})
	require.NoError(t, err)
	defer func() { _ = js.DeleteStream(ctx, topic) }()

	ncSub, err := nats.Connect(nats.DefaultURL)
	require.NoError(t, err)
	defer ncSub.Close()

	sub, err := ncSub.SubscribeSync(topic)
	require.NoError(t, err)
	require.NoError(t, ncSub.Flush())

	pub, err := jetstream.NewPublisher(jetstream.PublisherConfig{
		URL:        nats.DefaultURL,
		TrackMsgId: true,
	})
	require.NoError(t, err)
	defer func() { _ = pub.Close() }()

	wmMsg := message.NewMessage(watermill.NewUUID(), []byte("test"))
	require.NoError(t, pub.Publish(topic, wmMsg))

	natsMsg, err := sub.NextMsg(5 * time.Second)
	require.NoError(t, err)

	assert.Equal(t, wmMsg.UUID, natsMsg.Header.Get(nats.MsgIdHdr))
}

func TestPublisher_TrackMsgId_Disabled_DoesNotSetNatsMsgIdHeader(t *testing.T) {
	ctx := context.Background()

	topic := "watermill_test_" + watermill.NewShortUUID()

	ncMgmt, err := nats.Connect(nats.DefaultURL)
	require.NoError(t, err)
	defer ncMgmt.Close()

	js, err := natsJS.New(ncMgmt)
	require.NoError(t, err)

	_, err = js.CreateStream(ctx, natsJS.StreamConfig{
		Name:     topic,
		Subjects: []string{topic},
	})
	require.NoError(t, err)
	defer func() { _ = js.DeleteStream(ctx, topic) }()

	ncSub, err := nats.Connect(nats.DefaultURL)
	require.NoError(t, err)
	defer ncSub.Close()

	sub, err := ncSub.SubscribeSync(topic)
	require.NoError(t, err)
	require.NoError(t, ncSub.Flush())

	pub, err := jetstream.NewPublisher(jetstream.PublisherConfig{
		URL: nats.DefaultURL,
	})
	require.NoError(t, err)
	defer func() { _ = pub.Close() }()

	wmMsg := message.NewMessage(watermill.NewUUID(), []byte("test"))
	require.NoError(t, pub.Publish(topic, wmMsg))

	natsMsg, err := sub.NextMsg(5 * time.Second)
	require.NoError(t, err)

	assert.Empty(t, natsMsg.Header.Get(nats.MsgIdHdr))
}
