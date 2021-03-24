all: generate fmt build vet test

fmt:
	goimports -w .

generate:
	go generate ./...

build:
	go build -o aretext $(shell ./ldflags.sh) main.go

build-debug:
	go build -o aretext $(shell ./ldflags.sh) -gcflags "all=-N -l" main.go

test:
	go test ./...

vet:
	go vet ./...

bench:
	go test ./... -bench=.

clean:
	rm -rf aretext
	go clean ./...
