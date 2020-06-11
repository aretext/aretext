all: fmt build test

fmt:
	goimports -w ./internal/..
	goimports -w ./cmd/..

build:
	go build ./...

test:
	go test ./...

clean:
	rm -rf aretext
	go clean ./...
