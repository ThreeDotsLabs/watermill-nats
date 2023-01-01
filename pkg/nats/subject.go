package nats

// Subjects contains nats-core subject detail (primary + all additional) for a given watermill topic.
type Subjects struct {
	Primary    string
	Additional []string
}

// All combines the primary and all additional subjects for use by the nats-core client on creation.
func (s *Subjects) All() []string {
	return append([]string{s.Primary}, s.Additional...)
}

// SubjectCalculator is a function used to calculate nats-core subject(s) for the given topic.
type SubjectCalculator func(topic string) *Subjects

func DefaultSubjectCalculator(topic string) *Subjects {
	return &Subjects{
		Primary: topic,
	}
}
