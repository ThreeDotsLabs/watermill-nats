module main

go 1.18

replace github.com/ThreeDotsLabs/watermill-nats/v2 => ../

require (
	github.com/ThreeDotsLabs/watermill v1.2.0-rc.10
	github.com/ThreeDotsLabs/watermill-nats/v2 v2.0.0
	github.com/nats-io/nats-server/v2 v2.9.8
	github.com/nats-io/nats.go v1.20.0
)

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/nats-io/jwt/v2 v2.3.0 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.7.1 // indirect
	golang.org/x/crypto v0.0.0-20220926161630-eccd6366d1be // indirect
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
	golang.org/x/time v0.0.0-20220922220347-f3bd1da661af // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)
