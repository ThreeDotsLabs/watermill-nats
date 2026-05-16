package nats

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscriberSubscriptionConfig_Validate(t *testing.T) {
	tests := []struct {
		name                   string
		unmarshaler            Unmarshaler
		queueGroup             string
		subscribersCount       int
		subjectCalculator      SubjectCalculator
		subjectDetailGenerator SubjectDetailGenerator
		wantErr                bool
	}{
		{name: "OK - 1 Subscriber", unmarshaler: &GobMarshaler{}, subscribersCount: 1, wantErr: false, subjectCalculator: DefaultSubjectCalculator},
		{name: "OK - Multi Subscriber + Queue Group", unmarshaler: &GobMarshaler{}, subscribersCount: 3, queueGroup: "not empty", wantErr: false, subjectCalculator: DefaultSubjectCalculator},
		// TODO: revisit this validation
		//{name: "Invalid - Multi Subscriber no QueueGroupPrefix", unmarshaler: &GobMarshaler{}, subscribersCount: 3, wantErr: true, SubjectCalculator: DefaultSubjectCalculator("")},
		{name: "Invalid - No Unmarshaler", unmarshaler: nil, subscribersCount: 3, queueGroup: "not empty", wantErr: true, subjectCalculator: DefaultSubjectCalculator},
		{name: "Invalid - No Subject Calculator, no Subject Detailer", unmarshaler: &GobMarshaler{}, subscribersCount: 3, queueGroup: "not empty", wantErr: true, subjectCalculator: nil, subjectDetailGenerator: nil},
		{name: "Invalid - Subject Calculator, no Subject Detailer", unmarshaler: &GobMarshaler{}, subscribersCount: 3, queueGroup: "not empty", wantErr: false, subjectCalculator: DefaultSubjectCalculator, subjectDetailGenerator: nil},
		{name: "Invalid - No Subject Calculator, Subject Detailer", unmarshaler: &GobMarshaler{}, subscribersCount: 3, queueGroup: "not empty", wantErr: false, subjectCalculator: nil, subjectDetailGenerator: NewDefaultSubjectDetailer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SubscriberSubscriptionConfig{
				Unmarshaler:            tt.unmarshaler,
				SubscribersCount:       tt.subscribersCount,
				SubjectCalculator:      tt.subjectCalculator,
				SubjectDetailGenerator: tt.subjectDetailGenerator,
			}

			if tt.wantErr {
				require.Error(t, c.Validate())
			} else {
				require.NoError(t, c.Validate())
			}
		})
	}
}
