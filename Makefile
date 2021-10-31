all: fmt lint build

build:
	go build -o build/icloud ./cmd/icloud

test:
	go test -count=1 -v ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w $$(find ./cmd ./icloud -type d)

.PHONY: all build
