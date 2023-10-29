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
		Port:      42222,
		JetStream: true,
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
	subscribeOptions := []nc.SubOpt{
		nc.DeliverAll(),
		nc.AckExplicit(),
	}

	jsConfig := nats.JetStreamConfig{
		Disabled:         false,
		AutoProvision:    true,
		ConnectOptions:   nil,
		SubscribeOptions: subscribeOptions,
		PublishOptions:   nil,
		TrackMsgId:       false,
		AckAsync:         false,
		DurablePrefix:    "",
	}

	subConfig := nats.SubscriberConfig{
		URL:              svr.ClientURL(),
		CloseTimeout:     30 * time.Second,
		AckWaitTimeout:   30 * time.Second,
		NatsOptions:      options,
		Unmarshaler:      marshaler,
		JetStream:        jsConfig,
		QueueGroupPrefix: "testqg",
	}

	subscriber, err := nats.NewSubscriber(
		subConfig,
		logger,
	)
	if err != nil {
		panic(err)
	}

	subscriber2, err := nats.NewSubscriber(
		subConfig,
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

	messages, err := subscriber.Subscribe(context.Background(), "example_topic")
	if err != nil {
		panic(err)
	}

	messages2, err := subscriber2.Subscribe(context.Background(), "example_topic")
	if err != nil {
		panic(err)
	}

	go read(messages, 1)
	go read(messages2, 2)

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

	for {
		msg := message.NewMessage(watermill.NewUUID(), []byte("Hello, world!"))

		if err := publisher.Publish("example_topic", msg); err != nil {
			panic(err)
		}

		time.Sleep(time.Second)
	}
}

func read(messages <-chan *message.Message, subNum int) {
	for msg := range messages {
		log.Printf("received message: %s on %d, payload: %s", msg.UUID, subNum, string(msg.Payload))

		// we need to Acknowledge that we received and processed the message,
		// otherwise, it will be resent over and over again.
		msg.Ack()
	}
}
