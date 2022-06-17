package msg

import "fmt"

// Subjects contains nats subject detail (primary + all additional) for a given watermill topic.
type Subjects struct {
	Primary    string
	Additional []string
}

// All combines the primary and all additional subjects for use by the nats client on creation.
func (s *Subjects) All() []string {
	return append([]string{s.Primary}, s.Additional...)
}

func PublishSubject(topic string, uuid string) string {
	return fmt.Sprintf("%s.%s", topic, uuid)
}

// SubjectCalculator is a function used to calculate nats subject(s) for the given topic.
type SubjectCalculator func(topic string) *Subjects

func DefaultSubjectCalculator(topic string) *Subjects {
	return &Subjects{
		Primary: fmt.Sprintf("%s.*", topic),
	}
}
