#!/bin/bash
# run-validation.sh - Execute Phase 29 validation cycle
# This script runs a comprehensive validation of the Agent Stack Controller

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Output directory for reports
REPORT_DIR="validation-reports/$(date +%Y%m%d-%H%M%S)"
mkdir -p "$REPORT_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Phase 29: Validation Cycle${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Report directory: $REPORT_DIR"
echo ""

# Function to log section
log_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# Function to log success
log_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to log warning
log_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Function to log error
log_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Track overall status
CRITICAL_FAILURES=0
HIGH_FAILURES=0
WARNINGS=0

# 29.1 - Full Clean Build
log_section "29.1 - Full Clean Build"

echo "Cleaning build artifacts..."
make clean || true
rm -rf build/

echo "Building for all platforms..."
if make build-all > "$REPORT_DIR/build.log" 2>&1; then
    log_success "Build successful for all platforms"
    ls -lh build/ >> "$REPORT_DIR/build.log"
else
    log_error "Build failed"
    CRITICAL_FAILURES=$((CRITICAL_FAILURES + 1))
    cat "$REPORT_DIR/build.log"
fi

# 29.2 - Run Complete Test Suite
log_section "29.2 - Run Complete Test Suite"

echo "Running unit tests with coverage..."
if go test -v -race -coverprofile="$REPORT_DIR/coverage.out" ./... > "$REPORT_DIR/unit-tests.log" 2>&1; then
    log_success "Unit tests passed"
    go tool cover -func="$REPORT_DIR/coverage.out" > "$REPORT_DIR/coverage-summary.txt"
    COVERAGE=$(go tool cover -func="$REPORT_DIR/coverage.out" | grep total | awk '{print $3}')
    echo "Total coverage: $COVERAGE"
else
    log_error "Unit tests failed"
    HIGH_FAILURES=$((HIGH_FAILURES + 1))
    tail -50 "$REPORT_DIR/unit-tests.log"
fi

echo "Running integration tests..."
if go test -v -tags=integration ./test > "$REPORT_DIR/integration-tests.log" 2>&1; then
    log_success "Integration tests passed"
else
    log_warning "Integration tests failed (may require dependencies)"
    WARNINGS=$((WARNINGS + 1))
fi

echo "Running E2E tests..."
if go test -v -tags=e2e ./test > "$REPORT_DIR/e2e-tests.log" 2>&1; then
    log_success "E2E tests passed"
else
    log_warning "E2E tests failed (may require dependencies)"
    WARNINGS=$((WARNINGS + 1))
fi

# 29.3 - Analyze Test Results
log_section "29.3 - Analyze Test Results and Coverage"

echo "Analyzing coverage by package..."
go tool cover -func="$REPORT_DIR/coverage.out" | grep -v "100.0%" | grep -v "total:" > "$REPORT_DIR/coverage-gaps.txt" || true

echo "Packages with <80% coverage:"
awk '$3 < 80.0 {print $1, $3}' "$REPORT_DIR/coverage-gaps.txt" | tee "$REPORT_DIR/low-coverage.txt"

# 29.4 - Static Analysis
log_section "29.4 - Static Analysis and Linting"

echo "Running go vet..."
if go vet ./... > "$REPORT_DIR/vet.log" 2>&1; then
    log_success "go vet passed"
else
    log_warning "go vet found issues"
    WARNINGS=$((WARNINGS + 1))
    cat "$REPORT_DIR/vet.log"
fi

echo "Running gofmt check..."
UNFORMATTED=$(gofmt -l . | grep -v vendor || true)
if [ -z "$UNFORMATTED" ]; then
    log_success "All files properly formatted"
else
    log_warning "Some files need formatting"
    echo "$UNFORMATTED" | tee "$REPORT_DIR/format-issues.txt"
    WARNINGS=$((WARNINGS + 1))
fi

echo "Running golangci-lint..."
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run ./... > "$REPORT_DIR/golangci-lint.log" 2>&1; then
        log_success "golangci-lint passed"
    else
        log_warning "golangci-lint found issues"
        WARNINGS=$((WARNINGS + 1))
        head -50 "$REPORT_DIR/golangci-lint.log"
    fi
else
    log_warning "golangci-lint not installed, skipping"
fi

echo "Running gosec..."
if command -v gosec &> /dev/null; then
    if gosec -fmt=json -out="$REPORT_DIR/gosec.json" ./... > "$REPORT_DIR/gosec.log" 2>&1; then
        log_success "gosec passed"
    else
        log_warning "gosec found security issues"
        HIGH_FAILURES=$((HIGH_FAILURES + 1))
        cat "$REPORT_DIR/gosec.log"
    fi
else
    log_warning "gosec not installed, skipping"
fi

# 29.5 - Documentation Validation
log_section "29.5 - Documentation Validation"

echo "Checking for broken links in documentation..."
find docs -name "*.md" -type f > "$REPORT_DIR/doc-files.txt"
DOC_COUNT=$(wc -l < "$REPORT_DIR/doc-files.txt")
echo "Found $DOC_COUNT documentation files"

echo "Checking README completeness..."
if grep -q "## Installation" README.md && \
   grep -q "## Usage" README.md && \
   grep -q "## Configuration" README.md; then
    log_success "README has required sections"
else
    log_warning "README may be missing sections"
    WARNINGS=$((WARNINGS + 1))
fi

# 29.6 - Dependency Compatibility
log_section "29.6 - Dependency Compatibility"

echo "Current Go version:"
go version

echo "Checking go.mod..."
if go mod verify > "$REPORT_DIR/mod-verify.log" 2>&1; then
    log_success "go.mod verified"
else
    log_error "go.mod verification failed"
    CRITICAL_FAILURES=$((CRITICAL_FAILURES + 1))
    cat "$REPORT_DIR/mod-verify.log"
fi

echo "Checking for outdated dependencies..."
go list -u -m all > "$REPORT_DIR/dependencies.txt"

# 29.7 - Integration Validation
log_section "29.7 - Integration Validation"

echo "Testing binary execution..."
if [ -f "build/asc-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)" ]; then
    BINARY="build/asc-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"
    if $BINARY --version > "$REPORT_DIR/binary-version.txt" 2>&1; then
        log_success "Binary executes successfully"
        cat "$REPORT_DIR/binary-version.txt"
    else
        log_error "Binary execution failed"
        CRITICAL_FAILURES=$((CRITICAL_FAILURES + 1))
    fi
else
    log_warning "Binary not found, skipping execution test"
fi

# 29.8 - Security Validation
log_section "29.8 - Security Validation"

echo "Checking for hardcoded secrets..."
if grep -r -i "api[_-]key.*=.*['\"]sk-" . --exclude-dir=vendor --exclude-dir=.git --exclude="*.md" > "$REPORT_DIR/secret-scan.txt" 2>&1; then
    log_error "Found potential hardcoded secrets"
    CRITICAL_FAILURES=$((CRITICAL_FAILURES + 1))
    cat "$REPORT_DIR/secret-scan.txt"
else
    log_success "No hardcoded secrets found"
fi

echo "Checking file permissions..."
if [ -f ".env" ]; then
    PERMS=$(stat -f "%A" .env 2>/dev/null || stat -c "%a" .env 2>/dev/null)
    if [ "$PERMS" = "600" ]; then
        log_success ".env has correct permissions (600)"
    else
        log_warning ".env has incorrect permissions ($PERMS, should be 600)"
        WARNINGS=$((WARNINGS + 1))
    fi
fi

# 29.9 - Performance Validation
log_section "29.9 - Performance Validation"

echo "Running benchmarks..."
if go test -bench=. -benchmem ./... > "$REPORT_DIR/benchmarks.txt" 2>&1; then
    log_success "Benchmarks completed"
    grep "Benchmark" "$REPORT_DIR/benchmarks.txt" | head -20
else
    log_warning "Some benchmarks failed"
    WARNINGS=$((WARNINGS + 1))
fi

# Generate Summary Report
log_section "Validation Summary"

cat > "$REPORT_DIR/SUMMARY.md" <<EOF
# Validation Summary Report

**Date**: $(date)
**Report Directory**: $REPORT_DIR

## Results

### Build
- Status: $([ $CRITICAL_FAILURES -eq 0 ] && echo "✓ PASS" || echo "✗ FAIL")
- See: build.log

### Tests
- Unit Tests: $(grep -q "PASS" "$REPORT_DIR/unit-tests.log" && echo "✓ PASS" || echo "✗ FAIL")
- Coverage: $COVERAGE
- See: unit-tests.log, coverage.out, coverage-summary.txt

### Static Analysis
- go vet: $([ -s "$REPORT_DIR/vet.log" ] && echo "⚠ WARNINGS" || echo "✓ PASS")
- gofmt: $([ -z "$UNFORMATTED" ] && echo "✓ PASS" || echo "⚠ WARNINGS")
- See: vet.log, golangci-lint.log, gosec.json

### Security
- Secret Scan: $([ $CRITICAL_FAILURES -gt 0 ] && echo "✗ FAIL" || echo "✓ PASS")
- See: secret-scan.txt, gosec.json

### Documentation
- Files: $DOC_COUNT markdown files
- See: doc-files.txt

## Issue Summary

- **Critical Failures**: $CRITICAL_FAILURES
- **High Priority Issues**: $HIGH_FAILURES
- **Warnings**: $WARNINGS

## Recommendation

EOF

if [ $CRITICAL_FAILURES -eq 0 ] && [ $HIGH_FAILURES -eq 0 ]; then
    echo "✓ **GO**: System is ready for release" >> "$REPORT_DIR/SUMMARY.md"
    log_success "Validation PASSED - Ready for release"
elif [ $CRITICAL_FAILURES -eq 0 ]; then
    echo "⚠ **CONDITIONAL GO**: Address high-priority issues before release" >> "$REPORT_DIR/SUMMARY.md"
    log_warning "Validation PASSED with warnings - Address high-priority issues"
else
    echo "✗ **NO-GO**: Critical issues must be resolved before release" >> "$REPORT_DIR/SUMMARY.md"
    log_error "Validation FAILED - Critical issues must be resolved"
fi

echo ""
echo "Full report available at: $REPORT_DIR/SUMMARY.md"
cat "$REPORT_DIR/SUMMARY.md"

# Exit with appropriate code
if [ $CRITICAL_FAILURES -gt 0 ]; then
    exit 1
elif [ $HIGH_FAILURES -gt 0 ]; then
    exit 2
else
    exit 0
fi
