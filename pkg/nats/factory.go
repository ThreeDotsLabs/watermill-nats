package nats

import (
	"strings"
)

func GetMarshaler(format string) MarshalerUnmarshaler {
	switch strings.ToLower(format) {
	case "nats-core":
		return &NATSMarshaler{}
	case "proto":
		return &PBMarshaler{}
	case "json":
		return &JSONMarshaler{}
	default:
		return &GobMarshaler{}
	}
}
