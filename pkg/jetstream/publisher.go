package jetstream

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

// PublisherConfig is the configuration to create a publisher
type PublisherConfig struct {
	// URL is the NATS URL.
	URL string

	// NatsOptions are custom options for a connection.
	NatsOptions []nats.Option

	// JetstreamOptions are custom Jetstream options for a connection.
	JetstreamOptions []nats.JSOpt

	// Marshaler is marshaler used to marshal messages between watermill and wire formats
	Marshaler Marshaler

	// SubjectCalculator is a function used to transform a topic to an array of subjects on creation (defaults to "{topic}.*")
	SubjectCalculator SubjectCalculator

	// AutoProvision bypasses client validation and provisioning of streams
	AutoProvision bool

	// PublishOptions are custom publish option to be used on all publication
	PublishOptions []nats.PubOpt

	// TrackMsgId uses the Nats.MsgId option with the msg UUID to prevent duplication
	TrackMsgId bool
}

// PublisherPublishConfig is the configuration subset needed for an individual publish call
type PublisherPublishConfig struct {
	// Marshaler is marshaler used to marshal messages between watermill and wire formats
	Marshaler Marshaler

	// SubjectCalculator is a function used to transform a topic to an array of subjects on creation (defaults to "{topic}.*")
	SubjectCalculator SubjectCalculator

	// AutoProvision bypasses client validation and provisioning of streams
	AutoProvision bool

	// JetstreamOptions are custom Jetstream options for a connection.
	JetstreamOptions []nats.JSOpt

	// PublishOptions are custom publish option to be used on all publication
	PublishOptions []nats.PubOpt

	// TrackMsgId uses the Nats.MsgId option with the msg UUID to prevent duplication
	TrackMsgId bool
}

func (c *PublisherConfig) setDefaults() {
	if c.SubjectCalculator == nil {
		c.SubjectCalculator = defaultSubjectCalculator
	}
}

// Validate ensures configuration is valid before use
func (c PublisherConfig) Validate() error {
	if c.Marshaler == nil {
		return errors.New("PublisherConfig.Marshaler is missing")
	}

	if c.SubjectCalculator == nil {
		return errors.New("PublisherConfig.SubjectCalculator is missing")
	}
	return nil
}

// GetPublisherPublishConfig gets the configuration subset needed for individual publish calls once a connection has been established
func (c PublisherConfig) GetPublisherPublishConfig() PublisherPublishConfig {
	return PublisherPublishConfig{
		Marshaler:         c.Marshaler,
		SubjectCalculator: c.SubjectCalculator,
		AutoProvision:     c.AutoProvision,
		JetstreamOptions:  c.JetstreamOptions,
		PublishOptions:    c.PublishOptions,
		TrackMsgId:        c.TrackMsgId,
	}
}

// Publisher provides the jetstream implementation for watermill publish operations
type Publisher struct {
	conn             *nats.Conn
	config           PublisherPublishConfig
	logger           watermill.LoggerAdapter
	js               nats.JetStream
	topicInterpreter *topicInterpreter
}

// NewPublisher creates a new Publisher.
func NewPublisher(config PublisherConfig, logger watermill.LoggerAdapter) (*Publisher, error) {
	config.setDefaults()

	if err := config.Validate(); err != nil {
		return nil, err
	}

	conn, err := nats.Connect(config.URL, config.NatsOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to nats")
	}

	return NewPublisherWithNatsConn(conn, config.GetPublisherPublishConfig(), logger)
}

// NewPublisherWithNatsConn creates a new Publisher with the provided nats connection.
func NewPublisherWithNatsConn(conn *nats.Conn, config PublisherPublishConfig, logger watermill.LoggerAdapter) (*Publisher, error) {
	if logger == nil {
		logger = watermill.NopLogger{}
	}

	js, err := conn.JetStream(config.JetstreamOptions...)

	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn:             conn,
		config:           config,
		logger:           logger,
		js:               js,
		topicInterpreter: newTopicInterpreter(js, config.SubjectCalculator),
	}, nil
}

// Publish publishes message to NATS.
//
// Publish will not return until an ack has been received from JetStream.
// When one of messages delivery fails - function is interrupted.
func (p *Publisher) Publish(topic string, messages ...*message.Message) error {
	if p.config.AutoProvision {
		err := p.topicInterpreter.ensureStream(topic)
		if err != nil {
			return err
		}
	}

	for _, msg := range messages {
		messageFields := watermill.LogFields{
			"message_uuid": msg.UUID,
			"topic_name":   topic,
		}

		p.logger.Trace("Publishing message", messageFields)

		natsMsg, err := p.config.Marshaler.Marshal(topic, msg)
		if err != nil {
			return err
		}

		publishOpts := p.config.PublishOptions

		if p.config.TrackMsgId {
			publishOpts = append(publishOpts, nats.MsgId(msg.UUID))
		}

		if _, err := p.js.PublishMsg(natsMsg, publishOpts...); err != nil {
			return errors.Wrap(err, "sending message failed")
		}
	}

	return nil
}

// Close closes the publisher and the underlying connection
func (p *Publisher) Close() error {
	p.logger.Trace("Closing publisher", nil)
	defer p.logger.Trace("Publisher closed", nil)

	p.conn.Close()

	return nil
}
