# Quality Assurance Scripts

This directory contains scripts for quality assurance, testing, and monitoring.

## Available Scripts

### Test Quality Scripts

#### `check-flakiness.sh`

Detects flaky tests by running the test suite multiple times and analyzing failures.

**Usage:**

```bash
# Run tests 10 times (default)
./scripts/check-flakiness.sh

# Run tests 20 times
./scripts/check-flakiness.sh 20

# Via Make
make test-flakiness RUNS=20
```

**Output:**
- Creates `test-flakiness-results/` directory with test run data
- Generates a flakiness report with:
  - Flaky tests (fail sometimes but not always)
  - Consistently failing tests (fail every time)
  - Recommendations for fixing issues

**Exit Codes:**
- `0`: No flaky or failing tests detected
- `1`: Flaky or consistently failing tests found

**Example Report:**

```markdown
# Test Flakiness Report

**Generated**: 2025-11-10 14:30:00
**Total Runs**: 10
**Failed Runs**: 2

## Flaky Tests

⚠️  **TestProcessManager_StartStop**
   - Failed: 2/10 runs
   - Flakiness rate: 20%

## Recommendations

1. Investigate race conditions: Run with `go test -race`
2. Check timing assumptions: Look for hardcoded timeouts
3. Review external dependencies: Mock or stub external services
```

---

#### `profile-performance.sh`

Comprehensive performance profiling and benchmarking tool for asc.

**Usage:**

```bash
# Interactive mode (menu-driven)
./scripts/profile-performance.sh

# Command line mode
./scripts/profile-performance.sh all          # Run everything
./scripts/profile-performance.sh bench        # Run benchmarks only
./scripts/profile-performance.sh cpu          # Generate CPU profiles
./scripts/profile-performance.sh mem          # Generate memory profiles
./scripts/profile-performance.sh test         # Run performance tests
./scripts/profile-performance.sh analyze      # Analyze existing profiles
./scripts/profile-performance.sh report       # Generate report
./scripts/profile-performance.sh interactive  # Open interactive profiler
./scripts/profile-performance.sh clean        # Clean profile directory
```

**Output:**
- Creates `profiles/` directory with:
  - CPU profiles (`cpu-*.prof`)
  - Memory profiles (`mem-*.prof`)
  - Benchmark results (`benchmark-results.txt`)
  - Performance test results (`performance-tests.txt`)
  - Comprehensive report (`performance-report.md`)

**Features:**
- Benchmarks all packages with memory statistics
- Generates CPU profiles for hot paths (TUI, config, process management)
- Generates memory profiles for memory-intensive operations
- Runs performance regression tests
- Analyzes profiles and identifies bottlenecks
- Creates detailed performance report

**Interactive Profiler:**

After generating profiles, use the interactive mode to explore:

```bash
./scripts/profile-performance.sh interactive
```

This opens `go tool pprof` with the selected profile, where you can use commands like:
- `top`: Show top CPU/memory consumers
- `list <function>`: Show source code with annotations
- `web`: Generate visual graph (requires graphviz)
- `peek <function>`: Show callers and callees

**Example Report:**

```markdown
# Performance Report

Generated: 2025-11-10 14:30:00

## Benchmark Results

BenchmarkTUIRendering-8           1000    1234567 ns/op    12345 B/op    123 allocs/op
BenchmarkConfigLoad-8            10000     123456 ns/op     1234 B/op     12 allocs/op

## Performance Tests

PASS: TestPerformanceRegression
PASS: TestMemoryUsageUnderLoad
PASS: TestStartupTime (elapsed: 245ms)
PASS: TestShutdownTime (elapsed: 1.2s)

## CPU Profile Analysis

Top 10 CPU consumers:
  45.2%  renderAgentsPane
  23.1%  renderTaskPane
  12.3%  renderLogPane
  ...

## Recommendations

1. Optimize renderAgentsPane (45% of CPU time)
2. Consider caching for task rendering
3. Memory usage within acceptable limits
```

---

#### `analyze-test-timing.sh`

Analyzes test execution times and identifies slow tests.

**Usage:**

```bash
# Analyze test timing
./scripts/analyze-test-timing.sh

# Via Make
make test-timing
```

**Output:**
- Generates `test-timing-analysis.md` with:
  - Summary statistics (total time, average, slowest/fastest)
  - Top 20 slowest tests
  - Package timing breakdown
  - Performance warnings for slow tests
  - Time distribution histogram

**Exit Codes:**
- `0`: All tests complete in acceptable time
- `1`: Very slow tests detected (>10s) or many slow tests (>5 tests >5s)

**Example Report:**

```markdown
# Test Timing Analysis

**Generated**: 2025-11-10 14:30:00

## Summary

- **Total tests**: 156
- **Total time**: 45.23s
- **Average time**: 0.290s
- **Slowest test**: 8.234s (TestE2EComprehensive)
- **Fastest test**: 0.001s (TestConfigValidation)

## Slowest Tests (Top 20)

| Rank | Test | Package | Duration |
|------|------|---------|----------|
| 1 | TestE2EComprehensive | ./test | 8.234s |
| 2 | TestIntegrationFull | ./test | 5.123s |
...

## Performance Warnings

🟡 **Slow test**: TestE2EComprehensive (8.23s)
🟡 **Slow test**: TestIntegrationFull (5.12s)

## Recommendations

1. Profile the test: Use `-cpuprofile` to identify bottlenecks
2. Reduce test scope: Break large tests into smaller units
3. Mock expensive operations: Replace real I/O with mocks
```

---

## Integration with CI/CD

These scripts are integrated into the CI/CD pipeline:

### GitHub Actions Workflows

1. **Test Quality Monitoring** (`.github/workflows/test-quality.yml`)
   - Runs daily at 6:00 AM UTC
   - Analyzes test timing
   - Detects flaky tests (5 runs)
   - Tracks coverage trends

2. **Performance Testing** (`.github/workflows/performance.yml`)
   - Runs nightly at 2:00 AM UTC
   - Executes benchmarks
   - Profiles memory and CPU usage
   - Detects performance regressions

3. **License Compliance** (`.github/workflows/license-check.yml`)
   - Runs weekly on Monday at 9:00 AM UTC
   - Scans dependency licenses
   - Generates compliance reports

### Pull Request Checks

All PRs automatically run:
- Linting (golangci-lint)
- Unit tests with coverage
- Security scanning (gosec)
- Dependency vulnerability checks (govulncheck)
- Build verification
- Integration tests

## Local Development

### Pre-commit Checks

Install pre-commit hooks:

```bash
make setup-hooks
```

This installs hooks that run on every commit:
- Code formatting (gofmt)
- Static analysis (go vet)
- Fast linting (golangci-lint)
- Tests for changed packages

### Quality Checks

Run comprehensive quality checks before pushing:

```bash
# Run all quality checks
make quality

# Individual checks
make fmt          # Format code
make vet          # Run go vet
make lint         # Run golangci-lint
make test         # Run tests
make security     # Run security scan
make vuln-check   # Check vulnerabilities
```

### Performance Analysis

Profile and benchmark your code:

```bash
# Run benchmarks
make bench

# Compare with previous results
make bench-compare

# CPU profiling
make profile-cpu

# Memory profiling
make profile-mem
```

### Metrics and Reports

Generate quality metrics:

```bash
# Generate metrics report
make metrics

# Check test timing
make test-timing

# Check for flaky tests
make test-flakiness RUNS=10

# Check licenses
make license-check
```

## Best Practices

### Writing Tests

1. **Keep tests fast**: Unit tests should complete in < 1s
2. **Make tests deterministic**: Avoid random data, use fixed seeds
3. **Isolate tests**: Don't share state between tests
4. **Use table-driven tests**: Reduce setup overhead
5. **Mock external dependencies**: Don't rely on network, filesystem, etc.

### Avoiding Flaky Tests

1. **Don't use `time.Sleep()`**: Use proper synchronization
2. **Mock time**: Use `time.Now()` through an interface
3. **Avoid race conditions**: Use proper locking or channels
4. **Clean up resources**: Use `defer` or `t.Cleanup()`
5. **Test with `-race`**: Always run race detector

### Performance Optimization

1. **Profile first**: Don't optimize without data
2. **Focus on hot paths**: Optimize the 20% that matters
3. **Benchmark changes**: Verify improvements with benchmarks
4. **Consider memory**: Reduce allocations in hot paths
5. **Use caching**: Cache expensive computations

## Troubleshooting

### Flaky Test Detection Issues

**Problem**: Script reports false positives

**Solution**:
- Increase number of runs: `./scripts/check-flakiness.sh 50`
- Check for environmental factors (disk space, network)
- Run on different machines to verify

**Problem**: Script takes too long

**Solution**:
- Reduce number of runs for quick checks
- Run only specific packages: Modify script to accept package filter
- Use parallel test execution

### Test Timing Issues

**Problem**: Tests are slow but script doesn't identify them

**Solution**:
- Check if tests are running in parallel (may hide individual slowness)
- Run with `-p=1` to disable parallelism
- Look at package-level timing

**Problem**: JSON parsing errors

**Solution**:
- Ensure Go version supports `-json` flag (Go 1.10+)
- Check for corrupted test output
- Verify `jq` is installed for analysis scripts

## Contributing

When adding new quality scripts:

1. **Document usage**: Add clear usage instructions
2. **Add to Makefile**: Create a make target for easy access
3. **Integrate with CI**: Add to appropriate GitHub Actions workflow
4. **Test thoroughly**: Verify script works on different platforms
5. **Handle errors**: Provide clear error messages and exit codes

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [golangci-lint](https://golangci-lint.run/)
- [gosec](https://github.com/securego/gosec)
- [Quality Metrics Dashboard](../docs/QUALITY_METRICS.md)
- [Testing Guide](../TESTING.md)

---

**Last Updated**: 2025-11-10  
**Maintained By**: Development Team
