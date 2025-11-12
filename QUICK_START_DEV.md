# Quick Start for Developers

Get up and running with asc development in 5 minutes.

## Prerequisites

- Go 1.21+ installed
- Git installed
- Make installed
- Docker (optional, for containerized features)

## Setup (First Time)

```bash
# 1. Clone the repository
git clone https://github.com/yourusername/asc.git
cd asc

# 2. Set up development environment (installs hooks and tools)
make setup-dev

# 3. Build the project
make build

# 4. Run tests
make test
```

That's it! You're ready to develop.

## Daily Workflow

### Before Starting Work

```bash
# Update your local main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/my-feature
```

### While Developing

```bash
# Run tests frequently
make test

# Check your code
make check  # Runs fmt, vet, and test

# Build and test
make build
./build/asc --help
```

### Before Committing

The pre-commit hook will automatically run, but you can run checks manually:

```bash
# Format code
make fmt

# Run linter
make vet

# Run tests
make test

# Or run all checks at once
make check
```

### Committing

```bash
# Stage your changes
git add .

# Commit (pre-commit hook runs automatically)
git commit -m "feat: add awesome feature"

# Push to your fork
git push origin feature/my-feature
```

### Creating a Pull Request

1. Go to GitHub and create a pull request
2. Fill out the PR template
3. Wait for CI checks to pass
4. Address review feedback
5. Merge when approved!

## Common Commands

```bash
# Build
make build              # Build for current platform
make build-all          # Build for all platforms

# Testing
make test               # Run unit tests
make test-coverage      # Run tests with coverage report
make test-e2e           # Run end-to-end tests
make test-all           # Run all tests

# Code Quality
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golangci-lint
make check              # Run all checks

# Development
make run                # Build and run
make dev                # Run with race detector
make clean              # Clean build artifacts

# Setup
make setup-dev          # Set up development environment
make setup-hooks        # Install git hooks
make deps               # Download dependencies
```

## Project Structure

```
asc/
├── cmd/                # CLI commands
├── internal/           # Internal packages
│   ├── config/        # Configuration
│   ├── process/       # Process management
│   ├── tui/           # Terminal UI
│   ├── beads/         # Beads client
│   └── mcp/           # MCP client
├── agent/             # Python agent code
├── test/              # Integration and e2e tests
├── docs/              # Documentation
└── build/             # Build output
```

## Key Files

- `main.go` - Entry point
- `asc.toml` - Configuration file
- `.env` - API keys (gitignored)
- `Makefile` - Build automation
- `go.mod` - Go dependencies

## Testing

```bash
# Run specific test
go test -v -run TestName ./internal/config

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./internal/config
```

## Debugging

```bash
# Run with debug logging
export ASC_LOG_LEVEL=debug
./build/asc up

# View logs
tail -f ~/.asc/logs/asc.log

# Use delve debugger
dlv debug . -- up
```

## Getting Help

- **Documentation**: See [docs/README.md](docs/README.md)
- **Contributing Guide**: See [CONTRIBUTING.md](CONTRIBUTING.md)
- **Troubleshooting**: See [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- **Debugging**: See [DEBUGGING.md](DEBUGGING.md)
- **Testing**: See [TESTING.md](TESTING.md)

## Tips

1. **Run tests frequently** - Catch issues early
2. **Use the pre-commit hook** - Prevents bad commits
3. **Keep commits small** - Easier to review
4. **Write tests first** - TDD helps design better code
5. **Read existing code** - Learn the patterns
6. **Ask questions** - Use GitHub Discussions

## Optional: Docker Setup

Docker is optional but useful for testing containerized features:

```bash
# macOS
brew install --cask docker
open -a Docker

# Linux
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Verify
docker --version
```

See [README.md#docker-setup](README.md#docker-setup) for detailed instructions.

## Common Issues

### "Tests failing"
```bash
go clean -testcache
make test
```

### "Build failing"
```bash
make clean
go mod tidy
make build
```

### "Linter errors"
```bash
make fmt
make vet
golangci-lint run --fix ./...
```

### "Pre-commit hook not running"
```bash
make setup-hooks
```

### "Docker not found"
```bash
# Install Docker (optional)
brew install --cask docker  # macOS
# or see README.md#docker-setup for other platforms
```

## Next Steps

1. Read [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines
2. Check [CODE_REVIEW_CHECKLIST.md](CODE_REVIEW_CHECKLIST.md) to understand what reviewers look for
3. Browse existing issues for good first issues
4. Join discussions to ask questions

Happy coding! 🚀
