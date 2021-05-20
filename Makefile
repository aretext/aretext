all: generate fmt build vet test

fmt:
	goimports -w -local "github.com/aretext" .
	markdownfmt -w *.md ./docs

generate:
	go generate ./...

build:
	go build -o aretext $(shell ./ldflags.sh) main.go

build-debug:
	go build -o aretext $(shell ./ldflags.sh) -gcflags "all=-N -l" main.go

test:
	go test ./...

install:
	go install $(shell ./ldflags.sh)

vet:
	go vet ./...

bench:
	go test ./... -bench=.

clean:
	rm -rf aretext
	go clean ./...
