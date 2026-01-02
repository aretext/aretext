.PHONY: all fmt generate build build-debug release test install install-devtools vet staticcheck bench clean

VERSION := $(shell git describe --tags --always --dirty)
GO_OS ?= $(shell go env GOOS)
GO_ARCH ?= $(shell go env GOARCH)
GO_BUILD_FLAGS ?=
GO_LDFLAGS := -ldflags="-X 'main.version=$(VERSION)'"
GO_OUTPUT := aretext

RELEASE_PLATFORMS := linux_amd64 linux_arm64 darwin_arm64 freebsd_amd64 freebsd_arm64
RELEASE_DIR := dist
RELEASE_ARCHIVES := $(foreach platform,$(RELEASE_PLATFORMS),$(RELEASE_DIR)/aretext_$(VERSION)_$(platform).tar.gz)
RELEASE_CHECKSUM := $(RELEASE_DIR)/aretext_$(VERSION)_checksums.txt

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
	$(MAKE) build --no-print-directory GO_BUILD_FLAGS="-gcflags \"all=-N -l\""

release: $(RELEASE_ARCHIVES) $(RELEASE_CHECKSUM)

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)

$(RELEASE_DIR)/aretext_$(VERSION)_%.tar.gz: $(RELEASE_DIR)
	@goos=$(word 1,$(subst _, ,$*)); \
	goarch=$(word 2,$(subst _, ,$*)); \
	dir=$(RELEASE_DIR)/aretext_$(VERSION)_$${goos}_$${goarch}; \
	mkdir -p $$dir; \
	$(MAKE) build --no-print-directory GO_OS=$$goos GO_ARCH=$$goarch GO_BUILD_FLAGS="-trimpath" GO_OUTPUT=$$dir/aretext; \
	cp LICENSE $$dir/; \
	cp -r docs $$dir/; \
	archive_cwd=$$(dirname $$dir); \
	archive_src=$$(basename $$dir); \
	archive_dst=$@; \
	archive_ts=$$(git log -1 --format=%ct); \
	tar --mtime "@$$archive_ts" -czf $$archive_dst -C $$archive_cwd $$archive_src; \
	echo "Created tarball: $$archive_dst"

$(RELEASE_CHECKSUM): $(RELEASE_ARCHIVES)
	@(cd $(RELEASE_DIR) && shasum -a 256 $(notdir $^) > $(notdir $@))
	@echo "Generated checksums: $@"

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
