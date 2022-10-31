up:
	docker-compose up -d

test:
	go test -parallel 20 ./...

test_v:
	go test -parallel 20 -v ./...

test_short:
	go test -parallel 20 ./... -short

test_race:
	go test ./... -short -race

test_stress:
	STRESS_TEST_COUNT=4 go test -tags=stress -parallel 30 -timeout=45m ./...

test_reconnect:
	go test -tags=reconnect ./...

wait:
	go run github.com/ThreeDotsLabs/watermill/dev/wait-for@github-actions nats-streaming:4222

build:
	go build ./...

fmt:
	go fmt ./...
	goimports -l -w .

update_watermill:
	go get -u github.com/ThreeDotsLabs/watermill
	go mod tidy

	sed -i '\|go 1\.|d' go.mod
	go mod edit -fmt

