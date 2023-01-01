package nats

import (
	"github.com/nats-io/nats.go"
)

type topicInterpreter struct {
	js                nats.JetStreamManager
	subjectCalculator SubjectCalculator
}

func newTopicInterpreter(js nats.JetStreamManager, formatter SubjectCalculator) *topicInterpreter {
	if formatter == nil {
		formatter = DefaultSubjectCalculator
	}

	return &topicInterpreter{
		js:                js,
		subjectCalculator: formatter,
	}
}

func (b *topicInterpreter) ensureStream(topic string) error {
	_, err := b.js.StreamInfo(topic)

	if err != nil {
		_, err = b.js.AddStream(&nats.StreamConfig{
			Name:        topic,
			Description: "",
			Subjects:    b.subjectCalculator(topic).All(),
		})

		if err != nil {
			return err
		}
	}

	return err
}
