package nats

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublisherConfig_Validate(t *testing.T) {
	tests := []struct {
		name                   string
		marshaler              Marshaler
		subjectCalculator      SubjectCalculator
		subjectDetailGenerator SubjectDetailGenerator
		wantErr                bool
	}{
		{name: "OK", marshaler: &GobMarshaler{}, wantErr: false, subjectCalculator: DefaultSubjectCalculator},
		{name: "Invalid - No Marshaler", marshaler: nil, wantErr: true, subjectCalculator: DefaultSubjectCalculator},
		{name: "Invalid - No Subject Calculator, no Subject Detailer", marshaler: &GobMarshaler{}, wantErr: true, subjectCalculator: nil, subjectDetailGenerator: nil},
		{name: "Invalid - Subject Calculator, no Subject Detailer", marshaler: &GobMarshaler{}, wantErr: false, subjectCalculator: DefaultSubjectCalculator, subjectDetailGenerator: nil},
		{name: "Invalid - No Subject Calculator, Subject Detailer", marshaler: &GobMarshaler{}, wantErr: false, subjectCalculator: nil, subjectDetailGenerator: NewDefaultSubjectDetailer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PublisherConfig{
				SubjectCalculator:      tt.subjectCalculator,
				SubjectDetailGenerator: tt.subjectDetailGenerator,
				Marshaler:              tt.marshaler,
			}

			if tt.wantErr {
				require.Error(t, c.Validate())
			} else {
				require.NoError(t, c.Validate())
			}
		})
	}
}
