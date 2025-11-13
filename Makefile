.PHONY: all build test lint fmt vet security vuln-check coverage clean help

# Default target
all: fmt vet lint test build

# Help target
help:
	@echo "Available targets:"
	@echo "  all           - Run fmt, vet, lint, test, and build (default)"
	@echo "  build         - Build the project"
	@echo "  test          - Run all tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  coverage      - Run tests with coverage report"
	@echo "  lint          - Run golangci-lint"
	@echo "  fmt           - Format code with gofmt"
	@echo "  fmt-check     - Check if code is formatted (CI mode)"
	@echo "  vet           - Run go vet"
	@echo "  security      - Run gosec security scanner"
	@echo "  vuln-check    - Check for vulnerabilities in dependencies"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Download and verify dependencies"
	@echo "  ci            - Run all CI checks locally"

# Build the project
build:
	@echo "Building..."
	go build -v ./...

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
	rm -f coverage.out coverage.html

# Run all CI checks locally
ci: deps fmt-check vet lint test build
	@echo ""
	@echo "✓ All CI checks passed!"
