package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/jetstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	natsJS "github.com/nats-io/nats.go/jetstream"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svr, err := server.NewServer(&server.Options{
		Port:      42223,
		JetStream: true,
	})

	if err != nil {
		panic(err)
	}

	svr.Start()
	defer svr.Shutdown()

	topic := "example_topic"
	// this is the default watermill will look for if no namer func passed
	consumer := fmt.Sprintf("watermill__%s", topic)
	var namer jetstream.ConsumerConfigurator

	// test a custom namer
	/*
		consumer := fmt.Sprintf("test__%s", topic)
		namer := func(s, _ string) string {
			return fmt.Sprintf("test__%s", s)
		}
	*/

	logger := watermill.NewStdLogger(true, true)
	mainLogFields := watermill.LogFields{"topic": topic, "consumer": consumer, "url": svr.ClientURL()}

	// create an existing jetstream consumer and connect explicitly
	conn, natsConnectErr := nats.Connect(svr.ClientURL())
	if natsConnectErr != nil {
		panic(err)
	}
	js, jsErr := natsJS.New(conn)
	if jsErr != nil {
		panic(jsErr)
	}
	s, se := js.CreateStream(ctx, natsJS.StreamConfig{
		Name:     topic,
		Subjects: []string{topic},
	})
	if se != nil {
		panic(se)
	}
	_, ce := s.CreateOrUpdateConsumer(ctx, natsJS.ConsumerConfig{
		Name:      consumer,
		AckPolicy: natsJS.AckExplicitPolicy,
	})
	if ce != nil {
		panic(ce)
	}

	subscriber, err := jetstream.NewSubscriber(jetstream.SubscriberConfig{
		URL:                 svr.ClientURL(),
		Logger:              logger,
		AckWaitTimeout:      5 * time.Second,
		ResourceInitializer: jetstream.ExistingConsumer(namer, ""),
	})
	if err != nil {
		panic(err)
	}

	stopPublisher := make(chan struct{}, 1)

	messages, err := subscriber.Subscribe(ctx, "example_topic")
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range messages {
			logger.Info(fmt.Sprintf("received message: %s (test: %s), payload: %s", msg.UUID, msg.Metadata.Get("test"), string(msg.Payload)), mainLogFields)

			// we need to Acknowledge that we received and processed the message,
			// otherwise, it will be resent over and over again.
			msg.Ack()
		}
		logger.Info("subscriber stopped", mainLogFields)
	}()

	publisher, err := jetstream.NewPublisher(jetstream.PublisherConfig{
		URL:    svr.ClientURL(),
		Logger: logger,
	})
	if err != nil {
		panic(err)
	}

	logger.Info("starting", mainLogFields)

	go func() {
		for {
			select {
			case <-stopPublisher:
				if closeErr := publisher.Close(); err != nil {
					logger.Error("failed to close publisher", closeErr, mainLogFields)
				}
				logger.Info("publisher stopped", mainLogFields)

			case <-time.After(time.Second):
				msg := message.NewMessage(watermill.NewUUID(), []byte("Hello, world!"))
				msg.Metadata.Set("test", watermill.NewShortUUID())
				if err := publisher.Publish(topic, msg); err != nil {
					logger.Error("publish failed", err, mainLogFields)
				}
			}
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	for {
		select {

		case sig := <-c:
			signal.Ignore(sig)
			logger.Info("Ctrl+C pressed in Terminal - shutting down", mainLogFields)
			stopPublisher <- struct{}{}
			if closeErr := subscriber.Close(); closeErr != nil {
				logger.Error("failed to close subscriber", closeErr, mainLogFields)
			} else {
				logger.Info("subscriber closed", mainLogFields)
			}
			if cde := s.DeleteConsumer(ctx, consumer); cde != nil {
				logger.Error("failed to delete consumer", cde, mainLogFields)
			} else {
				logger.Info("consumer deleted", mainLogFields)
			}
			if sde := js.DeleteStream(ctx, topic); sde != nil {
				logger.Error("failed to delete stream", sde, mainLogFields)
			} else {
				logger.Info("stream deleted", mainLogFields)
			}

			logger.Info("done", mainLogFields)
			return
		}
	}
}
