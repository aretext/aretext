all: generate fmt build vet test

fmt:
	goimports -w .

generate:
	go generate ./...

build:
	go build -o aretext main.go

test:
	go test ./...

vet:
	go vet ./...

bench:
	go test ./... -bench=.

clean:
	rm -rf aretext
	go clean ./...
