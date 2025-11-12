# Quality Metrics Dashboard

This document provides an overview of the quality gates, monitoring systems, and metrics tracked for the Agent Stack Controller (asc) project.

## Overview

The asc project maintains high quality standards through automated quality gates, continuous monitoring, and comprehensive testing. This dashboard tracks key quality metrics and provides visibility into the health of the codebase.

## Quality Gates

### 1. Code Coverage

**Target**: ≥ 80% coverage for all packages

**Monitoring**:
- Automated coverage reports on every PR
- Coverage trends tracked over time
- Codecov integration for detailed analysis

**Status Indicators**:
- 🟢 Green: ≥ 80% coverage
- 🟡 Yellow: 60-79% coverage
- 🔴 Red: < 60% coverage

**Actions**:
- PRs with coverage decrease > 2% require justification
- New code should maintain or improve coverage
- Critical paths require 90%+ coverage

### 2. Static Analysis

**Tools**:
- `golangci-lint`: Comprehensive Go linting
- `gosec`: Security-focused static analysis
- `go vet`: Go's built-in static analyzer

**Checks**:
- Code formatting (gofmt, goimports)
- Code complexity (gocyclo)
- Security vulnerabilities (gosec)
- Code duplication (dupl)
- Unused code detection
- Error handling verification

**Configuration**: See `.golangci.yml`

**Failure Criteria**:
- Any high-severity security issue
- Cyclomatic complexity > 15
- Unchecked errors in critical paths

### 3. Dependency Security

**Tools**:
- `govulncheck`: Go vulnerability scanner
- `dependabot`: Automated dependency updates
- `pip-licenses`: Python license compliance

**Monitoring**:
- Weekly vulnerability scans
- Automated security advisories
- Dependency update PRs

**Actions**:
- Critical vulnerabilities: Immediate fix required
- High vulnerabilities: Fix within 7 days
- Medium/Low: Fix in next release cycle

### 4. License Compliance

**Allowed Licenses**:
- MIT
- Apache-2.0
- BSD-2-Clause, BSD-3-Clause
- ISC
- MPL-2.0

**Restricted Licenses** (require review):
- GPL (any version)
- AGPL (any version)
- LGPL (any version)

**Monitoring**:
- Weekly license scans
- Automated license reports
- New dependency license checks

**Workflow**: See `.github/workflows/license-check.yml`

### 5. Performance Regression

**Benchmarks**:
- TUI rendering performance
- Configuration parsing speed
- Process management operations
- Client API response times

**Thresholds**:
- Alert on > 50% performance regression
- Fail on > 150% performance regression
- Track memory allocation trends

**Monitoring**:
- Nightly benchmark runs
- PR-based benchmark comparisons
- Historical trend analysis

**Workflow**: See `.github/workflows/performance.yml`

### 6. Test Quality

**Metrics Tracked**:
- Test execution time
- Test flakiness rate
- Test coverage trends
- Test failure patterns

**Quality Standards**:
- No test should take > 5 seconds (unit tests)
- Flakiness rate < 1%
- All tests must be deterministic
- Integration tests < 30 seconds

**Monitoring**:
- Daily test timing analysis
- 5x flakiness detection runs
- Test execution trend tracking

**Workflow**: See `.github/workflows/test-quality.yml`

## Continuous Integration Pipeline

### PR Checks (Required)

1. **Linting** (`golangci-lint`)
   - Must pass all enabled linters
   - No new warnings introduced

2. **Unit Tests**
   - All tests must pass
   - Coverage must not decrease > 2%
   - Tests run on multiple platforms (Linux, macOS)
   - Tests run on multiple Go versions (1.21, 1.22)

3. **Security Scan** (`gosec`)
   - No high-severity issues
   - SARIF results uploaded to GitHub

4. **Dependency Check** (`govulncheck`)
   - No known vulnerabilities in dependencies

5. **Build Verification**
   - Successful builds for all target platforms
   - Binary artifacts generated

6. **Integration Tests**
   - End-to-end workflows verified
   - Multi-component integration tested

### Scheduled Checks

1. **Nightly Performance Tests** (2:00 AM UTC)
   - Full benchmark suite
   - Memory and CPU profiling
   - Load testing

2. **Daily Test Quality** (6:00 AM UTC)
   - Test timing analysis
   - Flakiness detection
   - Coverage trend tracking

3. **Weekly License Compliance** (Monday 9:00 AM UTC)
   - Dependency license scanning
   - License report generation
   - Compliance verification

4. **Weekly Dependency Updates** (Monday 9:00 AM UTC)
   - Automated dependency PRs
   - Security advisory checks
   - Version compatibility testing

## Quality Metrics

### Code Quality Metrics

| Metric | Target | Current | Trend |
|--------|--------|---------|-------|
| Code Coverage | ≥ 80% | Check Codecov | 📊 |
| Cyclomatic Complexity | ≤ 15 | Check golangci-lint | 📊 |
| Code Duplication | < 5% | Check dupl | 📊 |
| Security Issues | 0 high | Check gosec | 📊 |
| Linter Warnings | 0 | Check CI | 📊 |

### Test Quality Metrics

| Metric | Target | Monitoring |
|--------|--------|------------|
| Test Execution Time | < 2 min | Daily |
| Slowest Unit Test | < 5s | Daily |
| Test Flakiness Rate | < 1% | Daily |
| Integration Test Time | < 10 min | Per PR |
| E2E Test Time | < 30 min | Per PR |

### Performance Metrics

| Metric | Target | Monitoring |
|--------|--------|------------|
| TUI Render Time | < 16ms (60fps) | Nightly |
| Config Parse Time | < 100ms | Nightly |
| Process Start Time | < 500ms | Nightly |
| Memory Usage (idle) | < 50MB | Nightly |
| Memory Usage (load) | < 200MB | Nightly |

### Dependency Metrics

| Metric | Target | Monitoring |
|--------|--------|------------|
| Known Vulnerabilities | 0 | Weekly |
| Outdated Dependencies | < 10 | Weekly |
| License Compliance | 100% | Weekly |
| Dependency Count (Go) | Track | Weekly |
| Dependency Count (Python) | Track | Weekly |

## Quality Reports

### Available Reports

1. **Coverage Report**
   - Location: Codecov dashboard
   - Frequency: Every PR
   - Details: Line-by-line coverage, file coverage, package coverage

2. **Benchmark Report**
   - Location: GitHub Actions artifacts
   - Frequency: Nightly
   - Details: Performance trends, regression alerts

3. **License Report**
   - Location: GitHub Actions artifacts
   - Frequency: Weekly
   - Details: All dependency licenses, compliance status

4. **Test Timing Report**
   - Location: GitHub Actions artifacts
   - Frequency: Daily
   - Details: Slowest tests, timing trends

5. **Flakiness Report**
   - Location: GitHub Actions artifacts
   - Frequency: Daily
   - Details: Flaky test detection, failure patterns

6. **Security Report**
   - Location: GitHub Security tab
   - Frequency: Weekly + on-demand
   - Details: Vulnerability scan results, SARIF analysis

## Accessing Quality Metrics

### GitHub Actions

All quality checks run automatically in GitHub Actions:

```bash
# View workflow runs
https://github.com/yourusername/asc/actions

# Download artifacts
# Navigate to workflow run → Artifacts section
```

### Local Quality Checks

Run quality checks locally before pushing:

```bash
# Run all checks
make check

# Run linter
make lint

# Run tests with coverage
make test-coverage

# Run benchmarks
go test -bench=. -benchmem ./...

# Run security scan
gosec ./...

# Check for vulnerabilities
govulncheck ./...
```

### Codecov Dashboard

View detailed coverage reports:

```bash
# Visit Codecov dashboard
https://codecov.io/gh/yourusername/asc

# View coverage trends
# View file-level coverage
# View PR coverage impact
```

## Quality Improvement Process

### 1. Identify Issues

- Review quality metrics weekly
- Monitor CI failures and warnings
- Track performance regressions
- Analyze test flakiness

### 2. Prioritize

**P0 (Critical)**:
- Security vulnerabilities (high/critical)
- Consistently failing tests
- Major performance regressions (> 100%)
- License compliance violations

**P1 (High)**:
- Coverage decreases > 5%
- Flaky tests (> 5% failure rate)
- Medium security vulnerabilities
- Performance regressions (50-100%)

**P2 (Medium)**:
- Code complexity issues
- Minor performance regressions (< 50%)
- Low security vulnerabilities
- Code duplication

**P3 (Low)**:
- Code style issues
- Documentation gaps
- Minor optimizations

### 3. Fix and Verify

- Create issue for tracking
- Implement fix with tests
- Verify fix in CI
- Update documentation if needed

### 4. Monitor

- Track metric improvements
- Verify no regressions
- Update quality targets if needed

## Quality Gates Configuration

### Pre-commit Hooks

Installed via `make setup-hooks`:

```bash
# Runs on every commit
- go fmt
- go vet
- golangci-lint (fast mode)
- go test (changed packages)
```

### PR Requirements

All PRs must pass:

1. ✅ All CI checks pass
2. ✅ Code review approved
3. ✅ Coverage maintained or improved
4. ✅ No new security issues
5. ✅ No new linter warnings
6. ✅ Documentation updated (if needed)

### Release Requirements

All releases must pass:

1. ✅ All PR requirements
2. ✅ Full test suite passes
3. ✅ E2E tests pass
4. ✅ Performance benchmarks acceptable
5. ✅ Security scan clean
6. ✅ License compliance verified
7. ✅ Documentation complete
8. ✅ CHANGELOG updated

## Troubleshooting Quality Issues

### Coverage Decreased

```bash
# Identify uncovered code
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Add tests for uncovered code
# Focus on critical paths first
```

### Linter Failures

```bash
# Run linter locally
make lint

# Auto-fix issues
golangci-lint run --fix ./...

# Review remaining issues
# Fix manually or add exceptions (with justification)
```

### Performance Regression

```bash
# Run benchmarks locally
go test -bench=. -benchmem -cpuprofile=cpu.prof ./...

# Analyze profile
go tool pprof cpu.prof

# Identify hot paths
# Optimize critical sections
```

### Flaky Tests

```bash
# Run test multiple times
go test -count=100 -run TestName ./...

# Add debugging
# Check for race conditions
go test -race ./...

# Fix timing issues
# Add proper synchronization
```

### Security Vulnerabilities

```bash
# Check for vulnerabilities
govulncheck ./...

# Update dependencies
go get -u ./...
go mod tidy

# Verify fix
govulncheck ./...
```

## Best Practices

### Writing Quality Code

1. **Test First**: Write tests before implementation
2. **Keep It Simple**: Avoid unnecessary complexity
3. **Document**: Add godoc comments for exported items
4. **Handle Errors**: Check and handle all errors
5. **Benchmark**: Profile performance-critical code

### Maintaining Quality

1. **Review Metrics**: Check quality dashboard weekly
2. **Fix Issues Promptly**: Don't let technical debt accumulate
3. **Update Dependencies**: Keep dependencies current
4. **Monitor Trends**: Watch for gradual degradation
5. **Continuous Improvement**: Regularly raise quality bars

### Contributing

1. **Run Local Checks**: Use `make check` before pushing
2. **Write Tests**: Maintain or improve coverage
3. **Follow Style**: Use `gofmt` and `goimports`
4. **Document Changes**: Update docs and comments
5. **Review Quality**: Check CI results on your PR

## Resources

### Tools

- [golangci-lint](https://golangci-lint.run/)
- [gosec](https://github.com/securego/gosec)
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- [Codecov](https://codecov.io/)
- [Dependabot](https://github.com/dependabot)

### Documentation

- [Testing Guide](../TESTING.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Code Review Checklist](../CODE_REVIEW_CHECKLIST.md)
- [Debugging Guide](../DEBUGGING.md)

### Workflows

- [CI Workflow](../.github/workflows/ci.yml)
- [Performance Testing](../.github/workflows/performance.yml)
- [Test Quality](../.github/workflows/test-quality.yml)
- [License Check](../.github/workflows/license-check.yml)

## Changelog

### 2025-11-10

- ✅ Initial quality metrics dashboard created
- ✅ Code coverage reporting configured (Codecov)
- ✅ Static analysis tools configured (golangci-lint, gosec)
- ✅ Dependency vulnerability scanning implemented (govulncheck)
- ✅ License compliance checking added
- ✅ Performance regression testing configured
- ✅ Test execution time monitoring implemented
- ✅ Test flakiness detection added
- ✅ Quality metrics dashboard documented

---

**Last Updated**: 2025-11-10  
**Maintained By**: Development Team  
**Review Frequency**: Weekly
