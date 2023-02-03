package nats

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublisherConfig_Validate(t *testing.T) {
	tests := []struct {
		name              string
		marshaler         Marshaler
		subjectCalculator SubjectCalculator
		wantErr           bool
	}{
		{name: "OK", marshaler: &GobMarshaler{}, wantErr: false, subjectCalculator: DefaultSubjectCalculator},
		{name: "Invalid - No Marshaler", marshaler: nil, wantErr: true, subjectCalculator: DefaultSubjectCalculator},
		{name: "Invalid - No Subject Calculator", marshaler: &GobMarshaler{}, wantErr: true, subjectCalculator: nil},
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
