package nats

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscriberSubscriptionConfig_Validate(t *testing.T) {
	tests := []struct {
		name              string
		unmarshaler       Unmarshaler
		queueGroup        string
		subscribersCount  int
		SubjectCalculator SubjectCalculator
		wantErr           bool
	}{
		{name: "OK - 1 Subscriber", unmarshaler: &GobMarshaler{}, subscribersCount: 1, wantErr: false, SubjectCalculator: DefaultSubjectCalculator},
		{name: "OK - Multi Subscriber + Queue Group", unmarshaler: &GobMarshaler{}, subscribersCount: 3, queueGroup: "not empty", wantErr: false, SubjectCalculator: DefaultSubjectCalculator},
		// TODO: revisit this validation
		//{name: "Invalid - Multi Subscriber no QueueGroupPrefix", unmarshaler: &GobMarshaler{}, subscribersCount: 3, wantErr: true, SubjectCalculator: DefaultSubjectCalculator("")},
		{name: "Invalid - No Unmarshaler", unmarshaler: nil, subscribersCount: 3, queueGroup: "not empty", wantErr: true, SubjectCalculator: DefaultSubjectCalculator},
		{name: "Invalid - No Subject Calculator", unmarshaler: &GobMarshaler{}, subscribersCount: 3, queueGroup: "not empty", wantErr: true, SubjectCalculator: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SubscriberSubscriptionConfig{
				Unmarshaler:       tt.unmarshaler,
				SubscribersCount:  tt.subscribersCount,
				SubjectCalculator: tt.SubjectCalculator,
			}

			if tt.wantErr {
				require.Error(t, c.Validate())
			} else {
				require.NoError(t, c.Validate())
			}
		})
	}
}
