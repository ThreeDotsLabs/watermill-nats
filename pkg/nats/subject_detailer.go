package nats

import "github.com/pkg/errors"

type SubjectDetailGenerator func(streamName string, queueGroupPrefix string) SubjectDetailer

type SubjectDetailer interface {
	Subject(topic string) string
	AllSubjects(topic string) []string
	QueueGroup(topic string) string
	StreamName() string
}

type defaultSubjectDetail struct {
	streamName       string
	queueGroupPrefix string
}

func (d *defaultSubjectDetail) Subject(topic string) string {
	if len(d.streamName) > 0 {
		return d.streamName + "." + topic
	}
	return topic

}

func (d *defaultSubjectDetail) AllSubjects(_ string) []string {
	if len(d.streamName) > 0 {
		return []string{d.streamName, d.streamName + ".*"}
	} else {
		return nil
	}
}

func (d *defaultSubjectDetail) QueueGroup(_ string) string {
	return d.queueGroupPrefix
}

func (d *defaultSubjectDetail) StreamName() string {
	return d.streamName
}
func NewDefaultSubjectDetailer(
	streamName string,
	queueGroupPrefix string,
) SubjectDetailer {
	return &defaultSubjectDetail{streamName, queueGroupPrefix}
}

type subjectDetailFromCalculator struct {
	calculator       SubjectCalculator
	queueGroupPrefix string
}

func (d *subjectDetailFromCalculator) Subject(topic string) string {
	return d.calculator(d.queueGroupPrefix, topic).Primary
}

func (d *subjectDetailFromCalculator) AllSubjects(topic string) []string {
	return d.calculator(d.queueGroupPrefix, topic).All()
}

func (d *subjectDetailFromCalculator) QueueGroup(topic string) string {
	return d.calculator(d.queueGroupPrefix, topic).QueueGroup
}

func (d *subjectDetailFromCalculator) StreamName() string {
	return ""
}

func newSubjectDetailFromCalculator(
	formatter SubjectCalculator,
	queueGroupPrefix string,
) (SubjectDetailer, error) {
	if formatter == nil {
		return nil, errors.New("No subject calculator")
	}
	return &subjectDetailFromCalculator{formatter, queueGroupPrefix}, nil
}
