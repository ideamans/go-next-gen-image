# Makefile for go-next-gen-image

# Variables
BINARY_NAME := nextgenimage
CMD_PATH := ./cmd/$(BINARY_NAME)
GO := go
GOFLAGS := -v
BUILD_FLAGS := -ldflags="-s -w"
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS_VERSION := -ldflags="-s -w -X main.version=$(VERSION)"

# Platform detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_S),Darwin)
    PLATFORM := darwin
else ifeq ($(UNAME_S),Linux)
    PLATFORM := linux
else
    PLATFORM := windows
endif

ifeq ($(UNAME_M),arm64)
    ARCH := arm64
else
    ARCH := amd64
endif

# Default target
.PHONY: all
all: test build

# Build the CLI binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for $(PLATFORM)/$(ARCH)..."
	@$(GO) build $(BUILD_FLAGS_VERSION) -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: ./$(BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 $(GO) build $(BUILD_FLAGS_VERSION) -o $(BINARY_NAME)-linux-amd64 $(CMD_PATH)

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 $(GO) build $(BUILD_FLAGS_VERSION) -o $(BINARY_NAME)-darwin-amd64 $(CMD_PATH)

.PHONY: build-darwin-arm64
build-darwin-arm64:
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 $(GO) build $(BUILD_FLAGS_VERSION) -o $(BINARY_NAME)-darwin-arm64 $(CMD_PATH)

.PHONY: build-windows
build-windows:
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 $(GO) build $(BUILD_FLAGS_VERSION) -o $(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@$(GO) test $(GOFLAGS) ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@$(GO) test -race -coverprofile=coverage.out -covermode=atomic ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests for CLI specifically
.PHONY: test-cli
test-cli:
	@echo "Running CLI tests..."
	@$(GO) test $(GOFLAGS) $(CMD_PATH)/...

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...

# Run go mod tidy
.PHONY: tidy
tidy:
	@echo "Tidying modules..."
	@$(GO) mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME) $(BINARY_NAME)-* coverage.out coverage.html
	@echo "Clean complete"

# Install the binary to GOBIN or PATH
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@$(GO) install $(CMD_PATH)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

# Uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)
	@echo "Uninstall complete"

# Development build (with race detector)
.PHONY: dev
dev:
	@echo "Building development version with race detector..."
	@$(GO) build -race -o $(BINARY_NAME)-dev $(CMD_PATH)
	@echo "Development build complete: ./$(BINARY_NAME)-dev"

# Quick test - run a sample conversion
.PHONY: test-webp
test-webp: build
	@echo "Testing WebP conversion..."
	@if [ -f testdata/test_original.jpg ]; then \
		./$(BINARY_NAME) webp testdata/test_original.jpg test_output.webp && \
		echo "WebP conversion successful: test_output.webp" && \
		rm -f test_output.webp; \
	else \
		echo "Test image not found: testdata/test_original.jpg"; \
	fi

.PHONY: test-avif
test-avif: build
	@echo "Testing AVIF conversion..."
	@if [ -f testdata/test_original.jpg ]; then \
		./$(BINARY_NAME) avif testdata/test_original.jpg test_output.avif && \
		echo "AVIF conversion successful: test_output.avif" && \
		rm -f test_output.avif; \
	else \
		echo "Test image not found: testdata/test_original.jpg"; \
	fi

# Check dependencies
.PHONY: check-deps
check-deps:
	@echo "Checking dependencies..."
	@echo -n "Go version: "
	@$(GO) version
	@echo -n "libvips: "
	@if command -v vips >/dev/null 2>&1; then \
		vips --version | head -1; \
	else \
		echo "NOT FOUND - Please install libvips"; \
		echo "  macOS: brew install vips"; \
		echo "  Ubuntu: sudo apt-get install libvips-dev"; \
	fi
	@echo "Go modules:"
	@$(GO) list -m all | grep -E "(govips|cobra)"

# Generate test data index
.PHONY: gen-testdata
gen-testdata:
	@echo "Checking testdata index..."
	@if [ -f testdata/index.json ]; then \
		echo "testdata/index.json exists"; \
		jq -r '.[] | "\(.format): \(.path)"' testdata/index.json | head -5; \
		echo "...and $(shell jq '. | length' testdata/index.json) more entries"; \
	else \
		echo "testdata/index.json not found"; \
	fi

# CI simulation - run what CI would run
.PHONY: ci
ci: check-deps lint test test-coverage

# Help
.PHONY: help
help:
	@echo "go-next-gen-image Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            - Run tests and build (default)"
	@echo "  build          - Build the CLI binary for current platform"
	@echo "  build-all      - Build for all supported platforms"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-cli       - Run CLI tests only"
	@echo "  lint           - Run linter (golangci-lint)"
	@echo "  fmt            - Format code"
	@echo "  tidy           - Run go mod tidy"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  uninstall      - Remove binary from GOPATH/bin"
	@echo "  dev            - Build with race detector"
	@echo "  test-webp      - Quick test WebP conversion"
	@echo "  test-avif      - Quick test AVIF conversion"
	@echo "  check-deps     - Check required dependencies"
	@echo "  ci             - Run CI simulation (lint + test + coverage)"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make                    # Run tests and build"
	@echo "  make test              # Run tests only"
	@echo "  make build             # Build for current platform"
	@echo "  make build-all         # Build for all platforms"
	@echo "  make clean build       # Clean and rebuild"