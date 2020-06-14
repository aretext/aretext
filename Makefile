all: fmt build test

fmt:
	goimports -w ./internal/..
	goimports -w ./cmd/..

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf aretext
	go clean ./...
