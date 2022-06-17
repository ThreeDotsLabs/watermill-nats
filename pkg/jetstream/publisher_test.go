package jetstream

import (
	"testing"

	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/msg"
	"github.com/stretchr/testify/require"
)

func TestPublisherConfig_Validate(t *testing.T) {
	tests := []struct {
		name              string
		marshaler         msg.Marshaler
		subjectCalculator func(string) *msg.Subjects
		wantErr           bool
	}{
		{name: "OK", marshaler: &msg.GobMarshaler{}, wantErr: false, subjectCalculator: msg.DefaultSubjectCalculator},
		{name: "Invalid - No Marshaler", marshaler: nil, wantErr: true, subjectCalculator: msg.DefaultSubjectCalculator},
		{name: "Invalid - No Subject Calculator", marshaler: &msg.GobMarshaler{}, wantErr: true, subjectCalculator: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PublisherConfig{
				SubjectCalculator: tt.subjectCalculator,
				Marshaler:         tt.marshaler,
			}

			if tt.wantErr {
				require.Error(t, c.Validate())
			} else {
				require.NoError(t, c.Validate())
			}
		})
	}
}
