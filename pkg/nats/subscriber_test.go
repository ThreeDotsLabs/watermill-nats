package nats

import (
	"testing"

	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/msg"
	"github.com/stretchr/testify/require"
)

func TestSubscriberSubscriptionConfig_Validate(t *testing.T) {
	tests := []struct {
		name              string
		unmarshaler       msg.Unmarshaler
		queueGroup        string
		subscribersCount  int
		SubjectCalculator func(string) *msg.Subjects
		wantErr           bool
	}{
		{name: "OK - 1 Subscriber", unmarshaler: &msg.GobMarshaler{}, subscribersCount: 1, wantErr: false, SubjectCalculator: msg.DefaultSubjectCalculator},
		{name: "OK - Multi Subscriber + Queue Group", unmarshaler: &msg.GobMarshaler{}, subscribersCount: 3, queueGroup: "not empty", wantErr: false, SubjectCalculator: msg.DefaultSubjectCalculator},
		{name: "Invalid - Multi Subscriber no QueueGroup", unmarshaler: &msg.GobMarshaler{}, subscribersCount: 3, wantErr: true, SubjectCalculator: msg.DefaultSubjectCalculator},
		{name: "Invalid - No Unmarshaler", unmarshaler: nil, subscribersCount: 3, queueGroup: "not empty", wantErr: true, SubjectCalculator: msg.DefaultSubjectCalculator},
		{name: "Invalid - No Subject Calculator", unmarshaler: &msg.GobMarshaler{}, subscribersCount: 3, queueGroup: "not empty", wantErr: true, SubjectCalculator: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SubscriberSubscriptionConfig{
				Unmarshaler:       tt.unmarshaler,
				QueueGroup:        tt.queueGroup,
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
