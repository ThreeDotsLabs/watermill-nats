package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

type streamManager struct {
	SubjectDetailer
	js nats.JetStreamManager
}

func newStreamManager(
	js nats.JetStreamManager,
	detailer SubjectDetailer,
	formatter SubjectCalculator,
	queueGroupPrefix string,
) (*streamManager, error) {
	if detailer == nil {
		//Look to remove queuegroupprefix and formatter from here when subjectcalculator is removed
		var err error
		detailer, err = newSubjectDetailFromCalculator(formatter, queueGroupPrefix)
		if err != nil {
			return nil, errors.Wrap(err, "No detailer supplied")
		}
	}

	return &streamManager{
		detailer,
		js,
	}, nil
}

func (s *streamManager) ensureStream() error {
	streamName := s.StreamName()
	if len(streamName) > 0 {
		return s.ensureStreamForStreamName(streamName)
	}

	return nil
}

func (s *streamManager) ensureStreamForTopic(topic string) error {
	streamName := s.StreamName()
	if len(streamName) > 0 {
		return s.ensureStreamForStreamName(streamName)
	} else {
		return s.ensureStreamForStreamName(topic)
	}
}

func (s *streamManager) ensureStreamForStreamName(streamName string) error {
	_, err := s.js.StreamInfo(streamName)

	if err != nil {
		// TODO: provision durable as well
		// or simply provide override capability
		// TODO: Ensure that stream names do not contain disallowed characters
		// TODO: Add additional stream configuration here
		_, err = s.js.AddStream(&nats.StreamConfig{
			Name:        streamName,
			Description: "",
			Subjects:    s.AllSubjects(streamName),
		})

		if err != nil {
			return err
		}
	}

	return err
}
