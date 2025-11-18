# Claude Development Guide

This document provides instructions for validating changes to the monhang project using the available development tools.

## Prerequisites

Before running validation tools, ensure you have the required dependencies installed:

```bash
# Install golangci-lint (required for linting)
# Visit https://golangci-lint.run/usage/install/ for installation instructions

# Install gosec (optional, for security scanning)
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Install govulncheck (optional, for vulnerability scanning)
go install golang.org/x/vuln/cmd/govulncheck@latest
```

## Quick Validation

To run all standard checks before committing changes:

```bash
make all
```

This runs: `fmt`, `vet`, `lint`, `test`, and `build`

## Individual Validation Commands

### Format Code

Format all Go files according to standard conventions:

```bash
make fmt
```

Check if code is properly formatted (without modifying files):

```bash
make fmt-check
```

### Static Analysis

Run Go's built-in static analysis tool:

```bash
make vet
```

### Linting

Run comprehensive linting checks with golangci-lint:

```bash
make lint
```

### Testing

Run all tests:

```bash
make test
```

Run tests with verbose output:

```bash
make test-verbose
```

Run tests with coverage report:

```bash
make coverage
```

This generates `coverage.html` that you can open in a browser to view detailed coverage information.

### Build

Build the project to ensure it compiles:

```bash
make build
```

## Security and Vulnerability Checks

### Security Scanning

Run gosec security scanner to detect common security issues:

```bash
make security
```

### Vulnerability Check

Check for known vulnerabilities in dependencies:

```bash
make vuln-check
```

## CI Validation

To run the same checks that CI runs (recommended before pushing):

```bash
make ci
```

This runs: `deps`, `fmt-check`, `vet`, `lint`, `test`, and `build`

## Recommended Workflow

1. Make your changes to the code
2. Format the code: `make fmt`
3. Run all validation checks: `make all`
4. If everything passes, run CI checks: `make ci`
5. Commit and push your changes

## Other Useful Commands

Download and verify dependencies:

```bash
make deps
```

Clean build artifacts:

```bash
make clean
```

View all available make targets:

```bash
make help
```

## Troubleshooting

If `make lint` fails because golangci-lint is not installed, follow the installation instructions at: https://golangci-lint.run/usage/install/

If `make security` or `make vuln-check` fail due to missing tools, install them using the commands shown in the Prerequisites section above.
