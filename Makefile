all: generate fmt build vet staticcheck test

fmt:
	gofmt -s -w .
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

install-devtools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/shurcooL/markdownfmt@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

vet:
	go vet ./...

staticcheck:
	staticcheck ./...

bench:
	go test ./... -bench=.

clean:
	rm -rf aretext
	rm -rf dist
	go clean ./...
