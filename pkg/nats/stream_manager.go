package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

// TODO: Add additional stream configuration here
type StreamConfig struct {
	AllowRollup bool
}

type streamManager struct {
	SubjectDetailer
	config StreamConfig
	js     nats.JetStreamManager
}

func newStreamManager(
	js nats.JetStreamManager,
	config StreamConfig,
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
		config,
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
	info, err := s.js.StreamInfo(streamName)

	if err != nil {
		if errors.Is(err, nats.ErrStreamNotFound) {
			// TODO: provision durable as well
			// or simply provide override capability
			// TODO: Ensure that stream names do not contain disallowed characters
			_, err = s.js.AddStream(&nats.StreamConfig{
				Name:        streamName,
				Description: "",
				Subjects:    s.AllSubjects(streamName),
				AllowRollup: s.config.AllowRollup,
			})

			if err != nil {
				return err
			}
		} else {
			_, err = s.js.UpdateStream(&nats.StreamConfig{
				Name:        streamName,
				Description: "",
				Subjects:    s.AllSubjects(streamName),
				AllowRollup: s.config.AllowRollup,
			})

			if err != nil {
				return err
			}
		}
	} else if info.Config.AllowRollup != s.config.AllowRollup {

	}

	return nil
}
