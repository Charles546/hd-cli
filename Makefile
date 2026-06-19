.PHONY: build test lint install clean

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Build flags
LDFLAGS := -X github.com/Charles546/hd-cli/internal/cli.Version=$(VERSION) \
           -X github.com/Charles546/hd-cli/internal/cli.Commit=$(COMMIT) \
           -X github.com/Charles546/hd-cli/internal/cli.BuildDate=$(DATE)

# Binary name
BINARY_NAME := hd

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/hd

test:
	go test ./...

lint:
	golangci-lint run

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

clean:
	rm -f $(BINARY_NAME)
	go clean
