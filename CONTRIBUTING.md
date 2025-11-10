# Contributing to Agent Stack Controller (asc)

Thank you for your interest in contributing to asc! This guide will help you get started with development, testing, and submitting contributions.

## Table of Contents

- [Quick Start](#quick-start)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Standards](#code-standards)
- [Testing Guidelines](#testing-guidelines)
- [Submitting Changes](#submitting-changes)
- [Code Review Process](#code-review-process)
- [Getting Help](#getting-help)

## Quick Start

**Want to get started quickly?** See [QUICK_START_DEV.md](QUICK_START_DEV.md) for a 5-minute setup guide.

For detailed information, continue reading below.

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21 or later** - [Download](https://golang.org/dl/)
- **Git** - [Download](https://git-scm.com/downloads)
- **Python 3.9+** - For agent development
- **Make** - For build automation (usually pre-installed on macOS/Linux)

Optional but recommended:
- **golangci-lint** - For linting: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **age** - For secrets management: [Installation guide](https://github.com/FiloSottile/age#installation)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

```bash
git clone https://github.com/YOUR_USERNAME/asc.git
cd asc
```

3. Add the upstream repository:

```bash
git remote add upstream https://github.com/ORIGINAL_OWNER/asc.git
```

4. Verify your remotes:

```bash
git remote -v
# origin    https://github.com/YOUR_USERNAME/asc.git (fetch)
# origin    https://github.com/YOUR_USERNAME/asc.git (push)
# upstream  https://github.com/ORIGINAL_OWNER/asc.git (fetch)
# upstream  https://github.com/ORIGINAL_OWNER/asc.git (push)
```

## Development Setup

### 1. Install Dependencies

```bash
# Download Go dependencies
make deps

# Tidy and verify dependencies
make tidy
```

### 2. Build the Project

```bash
# Build for your current platform
make build

# The binary will be in build/asc
./build/asc --version
```

### 3. Run Tests

```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage

# Run end-to-end tests (requires dependencies)
make test-e2e
```

### 4. Set Up Pre-commit Hooks (Recommended)

Install the pre-commit hooks to automatically check your code before committing:

```bash
# Copy the pre-commit hook
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

The pre-commit hook will:
- Format your code with `go fmt`
- Run `go vet` to catch common issues
- Run unit tests
- Check for common mistakes

## Development Workflow

### 1. Create a Feature Branch

Always create a new branch for your work:

```bash
# Update your local main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test improvements

### 2. Make Your Changes

Follow these guidelines:

- **Write clear, focused commits** - Each commit should represent a single logical change
- **Write descriptive commit messages** - See [Commit Message Guidelines](#commit-message-guidelines)
- **Add tests** - All new features and bug fixes should include tests
- **Update documentation** - Keep docs in sync with code changes
- **Follow code standards** - See [Code Standards](#code-standards)

### 3. Test Your Changes

Before committing, ensure all tests pass:

```bash
# Run all checks (format, vet, test)
make check

# Run specific tests
go test ./internal/config -v

# Run tests with race detector
go test -race ./...

# Run e2e tests
make test-e2e
```

### 4. Commit Your Changes

```bash
# Stage your changes
git add .

# Commit with a descriptive message
git commit -m "feat: add hot-reload configuration support"

# Push to your fork
git push origin feature/your-feature-name
```

### 5. Keep Your Branch Updated

Regularly sync with upstream to avoid conflicts:

```bash
# Fetch upstream changes
git fetch upstream

# Rebase your branch on upstream/main
git rebase upstream/main

# Force push to your fork (if needed after rebase)
git push origin feature/your-feature-name --force-with-lease
```

## Code Standards

### Go Code Style

We follow standard Go conventions:

- **Use `gofmt`** - All code must be formatted with `gofmt` (run `make fmt`)
- **Follow Go Code Review Comments** - [Read the guide](https://github.com/golang/go/wiki/CodeReviewComments)
- **Use meaningful names** - Variables, functions, and types should have clear, descriptive names
- **Keep functions small** - Aim for functions that do one thing well
- **Document exported symbols** - All exported functions, types, and constants must have godoc comments

### Code Organization

```go
// Package example provides utilities for example functionality.
//
// This package implements the core example logic used throughout
// the application.
package example

import (
    "context"
    "fmt"
    
    "github.com/yourusername/asc/internal/errors"
)

// ExampleConfig holds configuration for the example system.
type ExampleConfig struct {
    // Name is the identifier for this example
    Name string
    
    // Timeout specifies how long to wait
    Timeout time.Duration
}

// NewExample creates a new Example instance with the given configuration.
//
// Returns an error if the configuration is invalid.
func NewExample(cfg ExampleConfig) (*Example, error) {
    if cfg.Name == "" {
        return nil, errors.New("name is required")
    }
    
    return &Example{
        name: cfg.Name,
        timeout: cfg.Timeout,
    }, nil
}
```

### Error Handling

- **Use error wrapping** - Provide context with `fmt.Errorf("context: %w", err)`
- **Create custom errors** - Use the `internal/errors` package for domain-specific errors
- **Check all errors** - Never ignore errors without explicit reason
- **Return early** - Use early returns to reduce nesting

```go
// Good
func processData(data []byte) error {
    if len(data) == 0 {
        return errors.New("data cannot be empty")
    }
    
    result, err := parse(data)
    if err != nil {
        return fmt.Errorf("failed to parse data: %w", err)
    }
    
    return store(result)
}

// Bad - ignoring errors
func processData(data []byte) {
    result, _ := parse(data)  // Don't do this!
    store(result)
}
```

### Testing Standards

- **Table-driven tests** - Use table-driven tests for multiple test cases
- **Test names** - Use descriptive test names: `TestFunctionName_Scenario_ExpectedBehavior`
- **Test coverage** - Aim for 80%+ coverage for new code
- **Mock external dependencies** - Don't make real network calls or file system operations in unit tests
- **Test error paths** - Always test error conditions

```go
func TestParseConfig_ValidTOML_ReturnsConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Config
        wantErr bool
    }{
        {
            name: "basic config",
            input: `[core]
beads_db_path = "./test"`,
            want: Config{
                Core: CoreConfig{
                    BeadsDBPath: "./test",
                },
            },
            wantErr: false,
        },
        {
            name:    "empty input",
            input:   "",
            want:    Config{},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseConfig(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseConfig() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Documentation Standards

- **Package documentation** - Every package should have a package comment
- **Function documentation** - All exported functions must have godoc comments
- **Example code** - Include examples for complex functionality
- **Keep docs updated** - Update documentation when changing behavior

## Testing Guidelines

### Unit Tests

Unit tests should:
- Test a single unit of functionality
- Run quickly (< 1 second per test)
- Not depend on external services
- Be deterministic (same input = same output)

```bash
# Run unit tests
go test ./...

# Run specific package tests
go test ./internal/config -v

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests verify multiple components working together:

```bash
# Run integration tests
go test ./test -v -run Integration
```

### End-to-End Tests

E2E tests verify the complete system:

```bash
# Run basic e2e tests
make test-e2e

# Run comprehensive e2e tests (requires all dependencies)
make test-e2e-full

# Run stress tests
make test-e2e-stress
```

See [test/E2E_TESTING.md](test/E2E_TESTING.md) for detailed e2e testing documentation.

### Test Coverage

We aim for:
- **80%+ coverage** for new code
- **100% coverage** for critical paths (config parsing, process management)
- **Error path coverage** - All error conditions should be tested

Check coverage:

```bash
make test-coverage
# Opens coverage.html in your browser
```

## Submitting Changes

### Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, no logic change)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks (dependencies, build)
- `perf:` - Performance improvements

**Examples:**

```
feat(tui): add vaporwave theme support

Implement a new vaporwave aesthetic theme with gradient colors,
animated borders, and retro-futuristic styling.

Closes #123
```

```
fix(process): prevent zombie processes on shutdown

Ensure all child processes receive SIGTERM before SIGKILL.
Add timeout handling for graceful shutdown.

Fixes #456
```

```
docs(readme): update installation instructions

Add instructions for installing via Homebrew and update
the quick start guide.
```

### Pull Request Process

1. **Update your branch** with the latest upstream changes:

```bash
git fetch upstream
git rebase upstream/main
```

2. **Push to your fork**:

```bash
git push origin feature/your-feature-name
```

3. **Create a Pull Request** on GitHub:
   - Use a clear, descriptive title
   - Fill out the PR template completely
   - Reference any related issues
   - Add screenshots for UI changes
   - Mark as draft if work is in progress

4. **Respond to feedback**:
   - Address all review comments
   - Push additional commits to your branch
   - Request re-review when ready

5. **Merge**:
   - Once approved, a maintainer will merge your PR
   - Delete your feature branch after merge

### Pull Request Checklist

Before submitting, ensure:

- [ ] Code follows the style guidelines
- [ ] All tests pass (`make check`)
- [ ] New tests added for new functionality
- [ ] Documentation updated
- [ ] Commit messages follow conventions
- [ ] No merge conflicts with main
- [ ] PR description is clear and complete

## Code Review Process

### For Contributors

When your PR is under review:

- **Be responsive** - Reply to comments within 48 hours
- **Be open to feedback** - Reviews help improve code quality
- **Ask questions** - If feedback is unclear, ask for clarification
- **Make requested changes** - Address all review comments
- **Keep discussions professional** - Focus on the code, not the person

### For Reviewers

When reviewing PRs:

- **Be constructive** - Suggest improvements, don't just criticize
- **Be specific** - Point to exact lines and explain why
- **Be timely** - Review within 2-3 business days
- **Approve when ready** - Don't block on minor style issues
- **Use the checklist** - See [CODE_REVIEW_CHECKLIST.md](CODE_REVIEW_CHECKLIST.md)

## Getting Help

### Resources

- **Documentation** - Check [docs/README.md](docs/README.md)
- **Troubleshooting** - See [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- **Debugging Guide** - See [DEBUGGING.md](DEBUGGING.md)
- **Architecture** - See [.kiro/specs/agent-stack-controller/design.md](.kiro/specs/agent-stack-controller/design.md)

### Communication

- **GitHub Issues** - For bug reports and feature requests
- **GitHub Discussions** - For questions and general discussion
- **Pull Requests** - For code review and collaboration

### Common Issues

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for solutions to common development issues:

- Build failures
- Test failures
- Dependency issues
- Environment setup problems

## Development Tips

### Useful Commands

```bash
# Quick development cycle
make fmt && make vet && make test && make build

# Run with race detector
go run -race . up

# Profile CPU usage
go test -cpuprofile=cpu.prof ./internal/tui
go tool pprof cpu.prof

# Profile memory usage
go test -memprofile=mem.prof ./internal/tui
go tool pprof mem.prof

# Check for common mistakes
go vet ./...

# Find inefficient code
golangci-lint run --enable=ineffassign,unused,staticcheck
```

### IDE Setup

**VS Code:**
- Install the Go extension
- Enable format on save
- Enable organize imports on save
- Configure gopls for better IntelliSense

**GoLand/IntelliJ:**
- Enable gofmt on save
- Enable optimize imports on save
- Configure file watchers for automatic formatting

### Debugging

See [DEBUGGING.md](DEBUGGING.md) for detailed debugging guides:

- Using Delve debugger
- Debugging TUI applications
- Debugging agent processes
- Analyzing logs
- Profiling performance

## License

By contributing to asc, you agree that your contributions will be licensed under the MIT License.

## Thank You!

Thank you for contributing to asc! Your efforts help make this project better for everyone.
