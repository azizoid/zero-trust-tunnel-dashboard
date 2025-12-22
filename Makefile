.PHONY: build test bench lint vulncheck clean install

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS = -X github.com/azizoid/zero-trust-tunnel-dashboard/pkg/version.Version=$(VERSION) \
          -X github.com/azizoid/zero-trust-tunnel-dashboard/pkg/version.Commit=$(COMMIT) \
          -X github.com/azizoid/zero-trust-tunnel-dashboard/pkg/version.BuildDate=$(BUILD_DATE)

# Build flags for reproducible builds
BUILD_FLAGS = -trimpath -ldflags "$(LDFLAGS)" -buildvcs=false

# Binary name
BINARY_NAME = tunnel-dash

build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(BUILD_FLAGS) -o $(BINARY_NAME) ./cmd/tunnel-dash

build-reproducible:
	@echo "Building $(BINARY_NAME) with reproducible build flags..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME) ./cmd/tunnel-dash

test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./pkg/...

bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./pkg/...

lint:
	@echo "Running linter..."
	@golangci-lint run

vulncheck:
	@echo "Running vulnerability check..."
	@govulncheck ./...

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out
	@go clean -cache

install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(BUILD_FLAGS) ./cmd/tunnel-dash

# Release build with checksums
release:
	@echo "Building release binaries..."
	@mkdir -p dist
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then \
				ext=".exe"; \
			else \
				ext=""; \
			fi; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-$$os-$$arch$$ext ./cmd/tunnel-dash; \
		done; \
	done
	@echo "Generating checksums..."
	@cd dist && sha256sum * > checksums.txt

help:
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  build-reproducible - Build with reproducible flags"
	@echo "  test               - Run tests"
	@echo "  bench              - Run benchmarks"
	@echo "  lint               - Run linter"
	@echo "  vulncheck          - Run vulnerability check"
	@echo "  clean              - Clean build artifacts"
	@echo "  install            - Install to GOPATH/bin"
	@echo "  release            - Build release binaries with checksums"
	@echo "  help               - Show this help message"


