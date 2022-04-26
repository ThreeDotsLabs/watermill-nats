package main

import (
	"context"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	wmnats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func main() {
	svr, err := server.NewServer(&server.Options{
		Port: 42222,
	})

	if err != nil {
		panic(err)
	}

	svr.Start()
	defer svr.Shutdown()

	marshaler := &wmnats.GobMarshaler{}
	logger := watermill.NewStdLogger(false, false)
	options := []nats.Option{
		nats.RetryOnFailedConnect(true),
		nats.Timeout(30 * time.Second),
		nats.ReconnectWait(1 * time.Second),
	}

	subscriber, err := wmnats.NewSubscriber(
		wmnats.SubscriberConfig{
			URL:            svr.ClientURL(),
			CloseTimeout:   30 * time.Second,
			AckWaitTimeout: 30 * time.Second,
			NatsOptions:    options,
			Unmarshaler:    marshaler,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	messages, err := subscriber.Subscribe(context.Background(), "example_topic_nats")
	if err != nil {
		panic(err)
	}

	go process(messages)

	publisher, err := wmnats.NewPublisher(
		wmnats.PublisherConfig{
			URL:         svr.ClientURL(),
			NatsOptions: options,
			Marshaler:   marshaler,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	publishMessages(publisher)
}

func publishMessages(publisher message.Publisher) {
	for {
		msg := message.NewMessage(watermill.NewUUID(), []byte("Hello, world!"))

		if err := publisher.Publish("example_topic_nats", msg); err != nil {
			panic(err)
		}

		time.Sleep(time.Second)
	}
}

func process(messages <-chan *message.Message) {
	for msg := range messages {
		log.Printf("received message: %s, payload: %s", msg.UUID, string(msg.Payload))

		// we need to Acknowledge that we received and processed the message,
		// otherwise, it will be resent over and over again.
		msg.Ack()
	}
}
