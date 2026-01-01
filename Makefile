.PHONY: all fmt generate build build-debug test install install-devtools vet staticcheck bench clean

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags="-X 'main.version=$(VERSION)'"

all: generate fmt build vet staticcheck test

fmt:
	gofmt -s -w .
	goimports -w -local "github.com/aretext" .
	markdownfmt -w *.md ./docs

generate:
	go generate ./...

build:
	go build $(LDFLAGS) -o aretext github.com/aretext/aretext

build-debug:
	go build $(LDFLAGS) -o aretext -gcflags "all=-N -l" github.com/aretext/aretext

test:
	go test ./...

install:
	go install

install-devtools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/shurcooL/markdownfmt@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

vet:
	go vet ./...

staticcheck:
	staticcheck --checks inherit,-ST1005 ./...

bench:
	go test ./... -bench=.

clean:
	rm -rf aretext
	rm -rf dist
	go clean ./...
