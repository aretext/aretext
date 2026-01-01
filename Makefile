.PHONY: all fmt generate build build-debug test install install-devtools vet staticcheck bench clean

VERSION := $(shell git describe --tags --always --dirty)
GO_OS ?= $(shell go env GOOS)
GO_ARCH ?= $(shell go env GOARCH)
GO_BUILD_FLAGS ?=
GO_LDFLAGS := -ldflags="-X 'main.version=$(VERSION)'"
GO_OUTPUT := aretext

all: generate fmt build vet staticcheck test

fmt:
	gofmt -s -w .
	goimports -w -local "github.com/aretext" .
	markdownfmt -w *.md ./docs

generate:
	go generate ./...

build:
	GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go build $(GO_BUILD_FLAGS) $(GO_LDFLAGS) -o $(GO_OUTPUT) github.com/aretext/aretext

build-debug:
	GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go build $(GO_BUILD_FLAGS) $(GO_LDFLAGS) -o $(GO_OUTPUT) -gcflags "all=-N -l" github.com/aretext/aretext

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
