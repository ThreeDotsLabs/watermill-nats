package jetstream

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

// SubjectCalculator is a function used to calculate nats subject(s) for the given topic.
type SubjectCalculator func(topic string) *Subjects

// Subjects contains nats subject detail (primary + all additional) for a given watermill topic.
type Subjects struct {
	Primary    string
	Additional []string
}

// All combines the primary and all additional subjects for use by the nats client on creation.
func (s *Subjects) All() []string {
	return append([]string{s.Primary}, s.Additional...)
}

type topicInterpreter struct {
	js                nats.JetStreamManager
	subjectCalculator SubjectCalculator
}

func defaultSubjectCalculator(topic string) *Subjects {
	return &Subjects{
		Primary: fmt.Sprintf("%s.*", topic),
	}
}

func newTopicInterpreter(js nats.JetStreamManager, formatter SubjectCalculator) *topicInterpreter {
	if formatter == nil {
		formatter = defaultSubjectCalculator
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

func PublishSubject(topic string, uuid string) string {
	return fmt.Sprintf("%s.%s", topic, uuid)
}
