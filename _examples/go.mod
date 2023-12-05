module main

go 1.21

toolchain go1.21.0

replace github.com/ThreeDotsLabs/watermill-nats/v2 => ../

require (
	github.com/ThreeDotsLabs/watermill v1.2.0
	github.com/ThreeDotsLabs/watermill-nats/v2 v2.0.0
	github.com/google/uuid v1.3.0
	github.com/nats-io/nats-server/v2 v2.10.4
	github.com/nats-io/nats.go v1.31.1-0.20231201130123-4af26aae2522
	github.com/stretchr/testify v1.8.1
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/nats-io/jwt/v2 v2.5.2 // indirect
	github.com/nats-io/nkeys v0.4.6 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/automaxprocs v1.5.3 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
