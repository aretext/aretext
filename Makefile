all: generate fmt build vet test

fmt:
	goimports -w -local "github.com/aretext" .
	markdownfmt -w *.md ./docs

generate:
	go generate ./...

build:
	go build -o aretext github.com/aretext/aretext

build-debug:
	go build -o aretext -gcflags "all=-N -l" github.com/aretext/aretext

test:
	go test ./...

install:
	go install

vet:
	go vet ./...

bench:
	go test ./... -bench=.

clean:
	rm -rf aretext
	rm -rf dist
	go clean ./...
