package msg

import (
	"strings"
)

func GetMarshaler(format string) MarshalerUnmarshaler {
	switch strings.ToLower(format) {
	case "nats":
		return &NATSMarshaler{}
	case "proto":
		return &PBMarshaler{}
	case "json":
		return &JSONMarshaler{}
	default:
		return &GobMarshaler{}
	}
}
