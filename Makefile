all: generate fmt build vet test

fmt:
	goimports -w ./internal/..
	goimports -w ./cmd/..
	pipenv run black .

generate:
	go generate ./...

build:
	go build ./...
	go build ./cmd/aretext

test:
	go test ./...

vet:
	go vet ./...

bench:
	go test ./... -bench=.

clean:
	rm -rf aretext
	go clean ./...
