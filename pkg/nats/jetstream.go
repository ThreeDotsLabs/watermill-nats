package nats

import "github.com/nats-io/nats.go"

// JetStreamConfig contains configuration settings specific to running in JetStream mode
type JetStreamConfig struct {
	// Enabled controls whether JetStream semantics should be used
	Enabled bool

	// AutoProvision indicates the application should create the configured stream if missing on the broker
	AutoProvision bool

	// ConnectOptions contains JetStream-specific options to be used when establishing context
	ConnectOptions []nats.JSOpt

	// SubscribeOptions contains options to be used when establishing subscriptions
	SubscribeOptions []nats.SubOpt

	// PublishOptions contains options to be sent on every publish operation
	PublishOptions []nats.PubOpt

	// TrackMsgId uses the Nats.MsgId option with the msg UUID to prevent duplication (needed for exactly once processing)
	TrackMsgId bool

	// AckSync enables synchronous acknowledgement (needed for exactly once processing)
	AckSync bool

	// DurableName is the JetStream durable name.
	//
	// Subscriptions may also specify a “durable name” which will survive client restarts.
	// Durable subscriptions cause the server to track the last acknowledged message
	// sequence number for a client and durable name. When the client restarts/resubscribes,
	// and uses the same client ID and durable name, the server will resume delivery beginning
	// with the earliest unacknowledged message for this durable subscription.
	//
	// Doing this causes the JetStream server to track
	// the last acknowledged message for that ClientID + DurableName.
	DurableName string
}
