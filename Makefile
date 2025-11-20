.PHONY: all build build-all test lint fmt vet security vuln-check coverage clean help install smoke-test integration-test test-setup test-cleanup

# Variables
BINARY_NAME=monhang
DIST_DIR=dist
CMD_PATH=./cmd/monhang
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: fmt vet lint test build

# Help target
help:
	@echo "Available targets:"
	@echo "  all              - Run fmt, vet, lint, test, and build (default)"
	@echo "  build            - Build the binary for current platform (outputs to dist/)"
	@echo "  build-all        - Build binaries for all platforms (Linux, macOS, Windows)"
	@echo "  install          - Install the binary to GOPATH/bin"
	@echo "  test             - Run all tests"
	@echo "  test-verbose     - Run tests with verbose output"
	@echo "  coverage         - Run tests with coverage report"
	@echo "  smoke-test       - Run smoke tests (builds and tests monhang end-to-end)"
	@echo "  integration-test - Run integration tests"
	@echo "  test-setup       - Setup test environment (git repos)"
	@echo "  test-cleanup     - Cleanup test environment"
	@echo "  lint             - Run golangci-lint"
	@echo "  fmt              - Format code with gofmt"
	@echo "  fmt-check        - Check if code is formatted (CI mode)"
	@echo "  vet              - Run go vet"
	@echo "  security         - Run gosec security scanner"
	@echo "  vuln-check       - Check for vulnerabilities in dependencies"
	@echo "  clean            - Remove build artifacts"
	@echo "  deps             - Download and verify dependencies"
	@echo "  ci               - Run all CI checks locally"

# Build the project for current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(DIST_DIR)
	go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "✓ Binary built: $(DIST_DIR)/$(BINARY_NAME)"

# Build for all platforms
build-all:
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(DIST_DIR)
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_PATH)
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "✓ All binaries built in $(DIST_DIR)/"

# Install binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(CMD_PATH)
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .

# Check if code is formatted (for CI)
fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "The following files are not formatted:"; \
		gofmt -s -l .; \
		echo "Please run: make fmt"; \
		exit 1; \
	fi
	@echo "All files are properly formatted"

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m ./...; \
	else \
		echo "golangci-lint is not installed. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# Run security scanner
security:
	@echo "Running gosec security scanner..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec is not installed. Install it with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
		exit 1; \
	fi

# Check for vulnerabilities
vuln-check:
	@echo "Checking for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck is not installed. Install it with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
		exit 1; \
	fi

# Download and verify dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	@echo "Verifying dependencies..."
	go mod verify

# Clean build artifacts
clean:
	@echo "Cleaning..."
	go clean
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html
	@echo "✓ Cleaned build artifacts"

# Setup test environment
test-setup:
	@echo "Setting up test environment..."
	@bash scripts/setup-test-repos.sh test-workspace

# Cleanup test environment
test-cleanup:
	@echo "Cleaning up test environment..."
	@rm -rf test-workspace test-workspace-integration
	@echo "✓ Test environment cleaned"

# Run smoke tests
smoke-test: build
	@echo "Running smoke tests..."
	@bash scripts/smoke-test.sh

# Run integration tests
integration-test: build
	@echo "Running integration tests..."
	@bash scripts/integration-test.sh all

# Run all CI checks locally
ci: deps fmt-check vet lint test build
	@echo ""
	@echo "✓ All CI checks passed!"
