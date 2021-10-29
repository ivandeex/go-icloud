all: build

build:
	go build -o build/icloud ./cmd/icloud

test:
	go test -count=1 -v ./...

lint:
	golangci-lint run ./...

.PHONY: all build
