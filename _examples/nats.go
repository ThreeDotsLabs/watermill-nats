package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats-server/v2/server"
	nc "github.com/nats-io/nats.go"
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

	marshaler := &nats.GobMarshaler{}
	logger := watermill.NewStdLogger(false, false)
	options := []nc.Option{
		nc.RetryOnFailedConnect(true),
		nc.Timeout(30 * time.Second),
		nc.ReconnectWait(1 * time.Second),
	}

	jsConfig := nats.JetStreamConfig{Disabled: true}

	subscriber, err := nats.NewSubscriber(
		nats.SubscriberConfig{
			URL:            svr.ClientURL(),
			CloseTimeout:   30 * time.Second,
			AckWaitTimeout: 30 * time.Second,
			NatsOptions:    options,
			Unmarshaler:    marshaler,
			JetStream:      jsConfig,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal - closing subscriber")
		subscriber.Close()
		os.Exit(0)
	}()

	messages, err := subscriber.Subscribe(context.Background(), "example_topic_nats")
	if err != nil {
		panic(err)
	}

	go process(messages)

	publisher, err := nats.NewPublisher(
		nats.PublisherConfig{
			URL:         svr.ClientURL(),
			NatsOptions: options,
			Marshaler:   marshaler,
			JetStream:   jsConfig,
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
