# Claude Development Guide

This document provides comprehensive information about the monhang project for AI coding agents and developers.

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture Guidelines](#architecture-guidelines)
- [Code Organization](#code-organization)
- [Development Workflow](#development-workflow)
- [Important Information for Code Agents](#important-information-for-code-agents)
- [Validation Tools](#validation-tools)

---

## Project Overview

**Monhang** is a component management tool designed to simplify dependency management for multi-component projects. It handles fetching, versioning, and organizing components and their dependencies.

### Key Features

- **Component bootstrapping**: Automatically fetch components and all their dependencies
- **Git-based repositories**: Uses Git as the component repository backend
- **Dependency management**: Supports build, runtime, and install-time dependencies
- **Multi-format configuration**: Supports both JSON and TOML configuration files
- **Dependency graph resolution**: Uses topological sorting for correct dependency ordering

### Configuration Format

Components are defined in `monhang.json` or `monhang.toml` files with:
- **name**: Component identification
- **version**: Version/tag to checkout
- **repo**: Git repository URL
- **deps**: Dependencies (build, runtime, install)

### Current Status

This is a development version. All commands and APIs may change without notice.

---

## Architecture Guidelines

### Design Principles

1. **Standard Go Project Layout**: Follows Go community conventions for project structure
2. **Clean Separation**: Command-line interface (cmd/) separate from core logic (internal/)
3. **Graph-Based Dependencies**: Uses directed graph for dependency resolution
4. **Format Flexibility**: Automatic detection of JSON vs TOML configuration
5. **Git-Centric**: Leverages Git for component versioning and fetching

### Core Components

#### Dependency Graph (`internal/monhang/component.go`)

The project uses a directed graph to model component dependencies:
- **Nodes**: Represent components (Project, ComponentRef)
- **Edges**: Represent dependency relationships
- **Topological Sort**: Ensures dependencies are processed in the correct order

Key types:
- `Project`: Top-level configuration with embedded ComponentRef and dependency graph
- `ComponentRef`: References to individual components
- `Dependency`: Container for build/runtime/install dependencies
- `RepoConfig`: Repository configuration (type, base URL)

#### Command System (`internal/monhang/command.go`)

Commands follow a consistent pattern:
- Each command is a `Command` struct with Name, Args, Short, Long description
- Commands have their own flag sets
- `Run` function executes the command logic

#### Bootstrap Process (`internal/monhang/bootstrap.go`)

The `boot` command workflow:
1. Parse configuration file (JSON or TOML)
2. Build dependency graph via `ProcessDeps()`
3. Topologically sort dependencies via `Sort()`
4. Fetch components in dependency order

### Dependency Management

The project uses minimal external dependencies:
- `github.com/op/go-logging`: Structured logging
- `github.com/twmb/algoimpl`: Graph algorithms (topological sort)
- `github.com/BurntSushi/toml`: TOML parsing

### Error Handling

- Fatal errors use `mglog.Fatal()` for logging and exit
- Non-fatal errors are logged and returned
- Git command errors include stderr output in error messages

---

## Code Organization

### Directory Structure

```
monhang/
├── cmd/
│   └── monhang/           # Command-line entry point
│       └── main.go        # Main application, CLI setup, command routing
├── internal/
│   └── monhang/           # Core business logic (not importable by external projects)
│       ├── bootstrap.go   # Boot command implementation
│       ├── build.go       # Build-related functionality
│       ├── command.go     # Command infrastructure
│       ├── component.go   # Component, dependency, and project types
│       └── component_test.go # Component tests
├── test/                  # Test fixtures and data
│   ├── monhang.json       # Example JSON configuration
│   └── monhang.toml       # Example TOML configuration
├── .github/
│   └── workflows/         # CI/CD pipelines
├── .golangci.yml          # Linter configuration
├── Makefile               # Build and validation automation
├── go.mod                 # Go module definition
├── LICENSE                # GPL v3 license
├── README.md              # User documentation
└── claude.md              # This file - developer/agent documentation
```

### Package Structure

#### `cmd/monhang`

**Purpose**: Application entry point and CLI setup

**Key responsibilities**:
- Argument parsing
- Command routing
- Logging configuration
- User interface (help, version)

**Important functions**:
- `main()`: Entry point, parses flags and routes to commands
- `setupLog()`: Configures structured logging with colors
- `usageExit()`: Prints usage information

#### `internal/monhang`

**Purpose**: Core component management logic

**Key files**:
- `component.go`: Data structures for components, dependencies, projects
- `bootstrap.go`: Workspace bootstrapping logic
- `command.go`: Command infrastructure and utilities
- `build.go`: Build-related functionality

**Key functions**:
- `ParseProjectFile()`: Parses JSON/TOML configuration files
- `ProcessDeps()`: Builds dependency graph
- `Sort()`: Topologically sorts dependencies
- `Fetch()`: Clones Git repositories

### File Naming Conventions

- `*_test.go`: Test files (run with `go test`)
- `*.go`: Go source files
- `*.json` / `*.toml`: Configuration files

---

## Development Workflow

### Before Making Changes

1. Understand the dependency graph implications of your changes
2. Check if configuration file parsing needs updates (both JSON and TOML)
3. Consider impact on topological sorting

### Making Changes

1. **Format your code**: `make fmt`
2. **Run static analysis**: `make vet`
3. **Run linter**: `make lint`
4. **Run tests**: `make test`
5. **Build**: `make build`

### Before Committing

**REQUIRED: All code changes MUST pass `make` validation before committing.**

Run the full validation suite locally:

```bash
make
```

This will:
- Format code with `gofmt`
- Run `go vet` for static analysis
- Run `golangci-lint` with 30+ linters
- Run all tests
- Build the binary

For full CI validation (includes additional checks):

```bash
make ci
```

This ensures your changes will pass CI before pushing.

### Recommended Workflow

1. Create a feature branch
2. Make your changes to the code
3. **Run `make` to validate changes** (REQUIRED)
4. Fix any linting or test failures
5. Run CI checks: `make ci`
6. Commit with a descriptive message
7. Push your changes

**Note**: The build will fail if:
- Code is not properly formatted (`gofmt`)
- Linting issues exist (complexity, style, security, etc.)
- Tests fail
- Build errors occur

---

## Important Information for Code Agents

### Critical Guidelines

1. **Always preserve configuration format flexibility**
   - Support both JSON and TOML formats
   - Test changes with both `test/monhang.json` and `test/monhang.toml`
   - Use `filepath.Ext()` for format detection

2. **Maintain graph correctness**
   - Dependency graph must remain acyclic (directed acyclic graph)
   - Topological sort is essential for correct component ordering
   - Test dependency resolution with complex scenarios

3. **Git operations are external**
   - Git commands use `exec.Command()` - ensure proper error handling
   - Include stderr in error messages for debugging
   - Consider Git availability in target environments

4. **Security considerations**
   - Be cautious with file path operations (note `#nosec G304` comments)
   - Validate repository URLs before Git operations
   - Don't expose sensitive information in logs

5. **Logging practices**
   - Use `mglog` (module logger) for all logging
   - Fatal errors: `mglog.Fatal()`
   - Notices: `mglog.Noticef()`
   - Debug: `mglog.Debug()`
   - Errors (non-fatal): `mglog.Error()`

### Linting Configuration

The project uses strict linting with `.golangci.yml`:
- **30+ enabled linters** including security (gosec), complexity (gocyclo, gocognit), and style
- **Function length limits**: Max 100 lines, 50 statements
- **Cyclomatic complexity**: Max 15
- **All govet analyzers enabled**
- **Type assertion checking**
- **Exhaustive enum checking**

When modifying code:
- Keep functions short and focused
- Reduce complexity (if-else chains, nested loops)
- Handle all errors explicitly
- Use meaningful variable names (no single-letter except loop indices)
- Add comments for exported functions and types
- End comments with periods (godot linter)

### Testing Practices

- **Test files**: Use `*_test.go` naming convention
- **Coverage target**: Aim for high coverage (run `make coverage`)
- **Race detection**: Tests run with `-race` flag in coverage mode
- **Test fixtures**: Use `test/` directory for example configurations

### Common Patterns

#### Adding a new command

1. Create command var in appropriate file (e.g., `bootstrap.go`)
2. Define `Command` struct with Name, Args, Short, Long
3. Create flags using `Command.Flag`
4. Implement `Run` function
5. Register in `commands` slice in `main.go`

Example:
```go
var CmdNewFeature = &Command{
    Name:  "feature",
    Args:  "[args]",
    Short: "short description",
    Long:  `Long description`,
}

var featureFlag = CmdNewFeature.Flag.String("f", "default", "flag description")

func runFeature(_ *Command, _ []string) {
    // Implementation
}

func init() {
    CmdNewFeature.Run = runFeature
}
```

#### Adding configuration fields

1. Add to relevant struct (ComponentRef, Dependency, Project) with both JSON and TOML tags
2. Update test fixtures in `test/monhang.json` and `test/monhang.toml`
3. Handle in `ProcessDeps()` if it affects dependency resolution
4. Update README.md with new configuration option

Example:
```go
type ComponentRef struct {
    Name    string `json:"name" toml:"name"`
    Version string `json:"version" toml:"version"`
    NewField string `json:"new_field" toml:"new_field"`  // Add both tags
}
```

### File References

When working with the codebase, key locations:

- **Main entry point**: `cmd/monhang/main.go:79` (main function)
- **Command routing**: `cmd/monhang/main.go:86-96`
- **Configuration parsing**: `internal/monhang/component.go:86-109` (ParseProjectFile)
- **Dependency graph building**: `internal/monhang/component.go:112-133` (ProcessDeps)
- **Topological sort**: `internal/monhang/component.go:136-139` (Sort)
- **Git operations**: `internal/monhang/component.go:53-64` (git function)
- **Component fetching**: `internal/monhang/component.go:77-81` (Fetch)

### Known TODOs

Check the codebase for TODO comments:
- `cmd/monhang/main.go:65`: Implement command-specific help

---

## Validation Tools

### Prerequisites

Before running validation tools, ensure you have the required dependencies installed:

```bash
# Install golangci-lint (required for linting)
# Visit https://golangci-lint.run/usage/install/ for installation instructions

# Install gosec (optional, for security scanning)
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Install govulncheck (optional, for vulnerability scanning)
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### Quick Validation

To run all standard checks before committing changes:

```bash
make all
```

This runs: `fmt`, `vet`, `lint`, `test`, and `build`

### Individual Validation Commands

#### Format Code

Format all Go files according to standard conventions:

```bash
make fmt
```

Check if code is properly formatted (without modifying files):

```bash
make fmt-check
```

#### Static Analysis

Run Go's built-in static analysis tool:

```bash
make vet
```

#### Linting

Run comprehensive linting checks with golangci-lint:

```bash
make lint
```

#### Testing

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

#### Build

Build the project to ensure it compiles:

```bash
make build
```

### Security and Vulnerability Checks

#### Security Scanning

Run gosec security scanner to detect common security issues:

```bash
make security
```

#### Vulnerability Check

Check for known vulnerabilities in dependencies:

```bash
make vuln-check
```

### CI Validation

To run the same checks that CI runs (recommended before pushing):

```bash
make ci
```

This runs: `deps`, `fmt-check`, `vet`, `lint`, `test`, and `build`

### Workflow Summary

1. Make your changes to the code
2. Format the code: `make fmt`
3. Run all validation checks: `make all`
4. If everything passes, run CI checks: `make ci`
5. Commit and push your changes

### Other Useful Commands

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

### Troubleshooting

If `make lint` fails because golangci-lint is not installed, follow the installation instructions at: https://golangci-lint.run/usage/install/

If `make security` or `make vuln-check` fail due to missing tools, install them using the commands shown in the Prerequisites section above.
