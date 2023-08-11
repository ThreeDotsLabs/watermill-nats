package nats

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJetStreamConfig_ShouldAutoProvision(t *testing.T) {
	tests := []struct {
		name          string
		disabled      bool
		autoProvision bool
		want          bool
	}{
		{"disabled", true, false, false},
		{"enabled without auto-provision", false, false, false},
		{"enabled with auto-provision", false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := JetStreamConfig{
				Disabled:      tt.disabled,
				AutoProvision: tt.autoProvision,
			}
			assert.Equalf(t, tt.want, c.ShouldAutoProvision(), "ShouldAutoProvision()")
		})
	}
}

func TestJetStreamConfig_CalculateDurableName(t *testing.T) {
	tests := []struct {
		name              string
		durablePrefix     string
		durableCalculator DurableCalculator
		topic             string
		want              string
	}{
		{"none", "", nil, "topic", ""},
		{"prefix", "durable", nil, "topic", "durable"},
		{"calculated", "durable", func(s string, s2 string) string {
			return fmt.Sprintf("%s+%s", s, s2)
		}, "topic", "durable+topic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := JetStreamConfig{
				DurablePrefix:     tt.durablePrefix,
				DurableCalculator: tt.durableCalculator,
			}
			assert.Equalf(t, tt.want, c.CalculateDurableName(tt.topic), "CalculateDurableName(%v)", tt.topic)
		})
	}
}
