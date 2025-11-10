# Agent Stack Controller (asc) - Makefile

# Binary name
BINARY_NAME=asc

# Version information
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"

# Build directory
BUILD_DIR=build

# Platforms for cross-compilation
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64

.PHONY: all build test clean install help deps tidy fmt vet lint build-all

# Default target
all: test build

## help: Display this help message
help:
	@echo "Agent Stack Controller (asc) - Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Development Setup:"
	@echo "  make setup-dev    - Set up development environment (install hooks, tools)"
	@echo "  make setup-hooks  - Install git pre-commit hooks"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build binaries for all platforms
build-all: clean
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} . ; \
		echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}" ; \
	done
	@echo "All builds complete!"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete!"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## test-e2e: Run end-to-end tests (requires build)
test-e2e: build
	@echo "Running end-to-end tests..."
	$(GOTEST) -tags=e2e -v ./test
	@echo "E2E tests complete!"

## test-e2e-full: Run comprehensive e2e tests (requires all dependencies)
test-e2e-full: build
	@echo "Running comprehensive e2e tests..."
	E2E_FULL=true $(GOTEST) -tags=e2e -v ./test -timeout 30m
	@echo "Comprehensive e2e tests complete!"

## test-e2e-stress: Run stress tests
test-e2e-stress: build
	@echo "Running stress tests..."
	E2E_STRESS=true $(GOTEST) -tags=e2e -v ./test -run TestE2EStress
	@echo "Stress tests complete!"

## test-all: Run all tests including e2e
test-all: test test-e2e
	@echo "All tests complete!"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

## install: Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOINSTALL) $(LDFLAGS) .
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) -v ./...
	@echo "Dependencies downloaded!"

## tidy: Tidy and verify dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "Dependencies tidied!"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Formatting complete!"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "Vet complete!"

## lint: Run golangci-lint (requires golangci-lint installed)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## setup-hooks: Install git hooks for development
setup-hooks:
	@echo "Installing git hooks..."
	@mkdir -p .git/hooks
	@cp .githooks/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed successfully!"
	@echo "Pre-commit hook will run: format check, vet, and tests"

## setup-dev: Set up development environment
setup-dev: deps setup-hooks
	@echo "Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "Development environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make build' to build the project"
	@echo "  2. Run 'make test' to run tests"
	@echo "  3. Run 'make check' to run all checks"
	@echo "  4. See 'make help' for more commands"

## run: Build and run the binary
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

## dev: Run in development mode (with race detector)
dev:
	@echo "Running in development mode..."
	$(GOCMD) run -race . up

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed!"

## quality: Run comprehensive quality checks
quality: fmt vet lint test test-coverage security vuln-check
	@echo "All quality checks passed!"

## security: Run security checks with gosec
security:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -fmt=text ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

## vuln-check: Check for known vulnerabilities
vuln-check:
	@echo "Checking for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./...; \
	fi

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem -benchtime=5s -run=^$$ ./...

## bench-compare: Run benchmarks and compare with previous results
bench-compare:
	@echo "Running benchmarks with comparison..."
	@if [ -f benchmark-old.txt ]; then \
		$(GOTEST) -bench=. -benchmem -benchtime=5s -run=^$$ ./... | tee benchmark-new.txt; \
		if command -v benchcmp >/dev/null 2>&1; then \
			benchcmp benchmark-old.txt benchmark-new.txt; \
		else \
			echo "benchcmp not installed. Install with: go install golang.org/x/tools/cmd/benchcmp@latest"; \
		fi; \
	else \
		echo "No previous benchmark results found. Creating baseline..."; \
		$(GOTEST) -bench=. -benchmem -benchtime=5s -run=^$$ ./... | tee benchmark-old.txt; \
	fi

## profile-cpu: Run CPU profiling
profile-cpu:
	@echo "Running CPU profiling..."
	@mkdir -p profiles
	$(GOTEST) -cpuprofile=profiles/cpu.prof -bench=. ./internal/tui/...
	@echo "CPU profile saved to profiles/cpu.prof"
	@echo "View with: go tool pprof profiles/cpu.prof"

## profile-mem: Run memory profiling
profile-mem:
	@echo "Running memory profiling..."
	@mkdir -p profiles
	$(GOTEST) -memprofile=profiles/mem.prof -bench=. ./internal/tui/...
	@echo "Memory profile saved to profiles/mem.prof"
	@echo "View with: go tool pprof profiles/mem.prof"

## license-check: Check dependency licenses
license-check:
	@echo "Checking dependency licenses..."
	@if command -v go-licenses >/dev/null 2>&1; then \
		go-licenses check ./... --disallowed_types=forbidden,restricted,reciprocal,unknown; \
		go-licenses report ./... > licenses-report.txt; \
		echo "License report saved to licenses-report.txt"; \
	else \
		echo "go-licenses not installed. Install with: go install github.com/google/go-licenses@latest"; \
	fi

## metrics: Generate quality metrics report
metrics:
	@echo "Generating quality metrics..."
	@echo "=== Code Coverage ===" > metrics-report.txt
	@if [ -f coverage.out ]; then \
		go tool cover -func=coverage.out | tail -1 >> metrics-report.txt; \
	else \
		echo "No coverage data. Run 'make test-coverage' first." >> metrics-report.txt; \
	fi
	@echo "" >> metrics-report.txt
	@echo "=== Test Count ===" >> metrics-report.txt
	@find . -name "*_test.go" -not -path "./vendor/*" | wc -l | xargs echo "Test files:" >> metrics-report.txt
	@grep -r "func Test" --include="*_test.go" --exclude-dir=vendor | wc -l | xargs echo "Test functions:" >> metrics-report.txt
	@echo "" >> metrics-report.txt
	@echo "=== Code Statistics ===" >> metrics-report.txt
	@find . -name "*.go" -not -path "./vendor/*" -not -name "*_test.go" | xargs wc -l | tail -1 | xargs echo "Production code:" >> metrics-report.txt
	@find . -name "*_test.go" -not -path "./vendor/*" | xargs wc -l | tail -1 | xargs echo "Test code:" >> metrics-report.txt
	@cat metrics-report.txt
	@echo ""
	@echo "Full report saved to metrics-report.txt"

## test-timing: Analyze test execution times
test-timing:
	@echo "Analyzing test timing..."
	@./scripts/analyze-test-timing.sh

## test-flakiness: Check for flaky tests (runs tests 10 times by default)
test-flakiness:
	@echo "Checking for flaky tests..."
	@./scripts/check-flakiness.sh $(RUNS)

## release: Prepare a release (build all platforms, run tests)
release: check build-all
	@echo "Release artifacts ready in $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/
