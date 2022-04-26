package main

import (
	"context"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/jetstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go"
)

func main() {
	marshaler := &jetstream.GobMarshaler{}
	logger := watermill.NewStdLogger(false, false)
	options := []nats.Option{
		nats.RetryOnFailedConnect(true),
		nats.Timeout(30 * time.Second),
		nats.ReconnectWait(1 * time.Second),
	}
	subscribeOptions := []nats.SubOpt{
		nats.DeliverAll(),
		nats.AckExplicit(),
	}

	subscriber, err := jetstream.NewSubscriber(
		jetstream.SubscriberConfig{
			URL:              nats.DefaultURL,
			CloseTimeout:     30 * time.Second,
			AckWaitTimeout:   30 * time.Second,
			NatsOptions:      options,
			Unmarshaler:      marshaler,
			SubscribeOptions: subscribeOptions,
			AutoProvision:    true,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	messages, err := subscriber.Subscribe(context.Background(), "example_topic")
	if err != nil {
		panic(err)
	}

	go processJS(messages)

	publisher, err := jetstream.NewPublisher(
		jetstream.PublisherConfig{
			URL:         nats.DefaultURL,
			NatsOptions: options,
			Marshaler:   marshaler,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	publishMessagesJS(publisher)
}

func publishMessagesJS(publisher message.Publisher) {
	for {
		msg := message.NewMessage(watermill.NewUUID(), []byte("Hello, world!"))

		if err := publisher.Publish("example_topic", msg); err != nil {
			panic(err)
		}

		time.Sleep(time.Second)
	}
}

func processJS(messages <-chan *message.Message) {
	for msg := range messages {
		log.Printf("received message: %s, payload: %s", msg.UUID, string(msg.Payload))

		// we need to Acknowledge that we received and processed the message,
		// otherwise, it will be resent over and over again.
		msg.Ack()
	}
}
