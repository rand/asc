# Performance Validation Guide

This guide explains how to run and interpret performance validation tests for the Agent Stack Controller (asc).

## Quick Start

### Run All Performance Tests

```bash
# Run all performance validation tests
go test -v -run "TestPerformanceValidation" ./test

# Run with automated report generation
./scripts/run-performance-validation.sh
```

### Run Specific Performance Tests

```bash
# Startup time only
go test -v -run "TestPerformanceValidation_StartupTime" ./test

# Memory usage only
go test -v -run "TestPerformanceValidation_MemoryUsage" ./test

# Shutdown time only
go test -v -run "TestPerformanceValidation_ShutdownTime" ./test

# TUI responsiveness only
go test -v -run "TestPerformanceValidation_TUIResponsiveness" ./test

# Task throughput only
go test -v -run "TestPerformanceValidation_TaskProcessingThroughput" ./test

# Large log files only
go test -v -run "TestPerformanceValidation_LargeLogFiles" ./test

# Many tasks only
go test -v -run "TestPerformanceValidation_ManyTasks" ./test

# Concurrent operations only
go test -v -run "TestPerformanceValidation_ConcurrentOperations" ./test

# Memory leak detection only
go test -v -run "TestPerformanceValidation_MemoryLeaks" ./test
```

## Test Coverage

### 1. Startup Time Validation
Tests application startup time with varying numbers of agents.

**Configurations:**
- 1 agent (target: < 200ms)
- 3 agents (target: < 300ms)
- 5 agents (target: < 400ms)
- 10 agents (target: < 500ms)

**What it measures:**
- Configuration loading time
- Process manager initialization
- Agent setup overhead

### 2. Shutdown Time Validation
Tests graceful shutdown time with varying numbers of agents.

**Configurations:**
- 1 agent (target: < 500ms)
- 3 agents (target: < 1s)
- 5 agents (target: < 1.5s)
- 10 agents (target: < 2s)

**What it measures:**
- Process termination time
- Cleanup operations
- Resource deallocation

### 3. Memory Usage Validation
Tests memory consumption with different agent configurations.

**Configurations:**
- 1 agent (target: < 5 MB)
- 3 agents (target: < 10 MB)
- 5 agents (target: < 15 MB)
- 10 agents (target: < 25 MB)

**What it measures:**
- Total memory allocated
- Heap memory usage
- Memory scaling with agent count

### 4. TUI Responsiveness Validation
Tests TUI responsiveness under different load conditions.

**Load Levels:**
- Light (5 agents, target: < 5ms avg)
- Medium (10 agents, target: < 10ms avg)
- Heavy (20 agents, target: < 20ms avg)

**What it measures:**
- Configuration reload time
- UI update latency
- Rendering performance

### 5. Task Processing Throughput
Tests task processing throughput with varying workloads.

**Workloads:**
- Small (10 tasks, 1 agent, target: > 5 tasks/s)
- Medium (50 tasks, 3 agents, target: > 10 tasks/s)
- Large (100 tasks, 5 agents, target: > 15 tasks/s)

**What it measures:**
- Task assignment speed
- Agent coordination overhead
- Overall system throughput

### 6. Large Log File Handling
Tests performance with large log files.

**File Sizes:**
- 10 MB (target: < 500ms)
- 50 MB (target: < 2s)
- 100 MB (target: < 4s)

**What it measures:**
- File I/O performance
- Read operation speed
- Memory efficiency during file operations

### 7. Many Tasks Handling
Tests performance with large numbers of tasks.

**Task Counts:**
- 100 tasks (target: < 100ms)
- 500 tasks (target: < 300ms)
- 1000 tasks (target: < 500ms)

**What it measures:**
- Task parsing speed
- Data structure efficiency
- Scaling with task count

### 8. Concurrent Operations
Tests performance under concurrent load.

**Concurrency Levels:**
- Low (5 goroutines, 100 ops, target: < 1s)
- Medium (10 goroutines, 200 ops, target: < 2s)
- High (20 goroutines, 400 ops, target: < 3s)

**What it measures:**
- Thread safety overhead
- Lock contention
- Concurrent operation throughput

### 9. Memory Leak Detection
Tests for memory leaks over extended operations.

**Test Parameters:**
- 1000 iterations
- Target: < 10 MB heap growth

**What it measures:**
- Heap memory growth
- Resource cleanup
- Memory stability

## Performance Baselines

### Established Baselines
- **Startup time:** < 1ms per agent
- **Shutdown time:** < 0.5ms per agent
- **Memory usage:** < 1MB per agent
- **TUI response:** < 1ms average
- **Task throughput:** > 500 tasks/second
- **File I/O:** < 1ms per 10MB

### Alert Thresholds
Trigger alerts if:
- Startup time increases by > 50%
- Memory usage increases by > 100%
- Throughput decreases by > 30%
- Response time increases by > 100%

## Interpreting Results

### Success Criteria
All tests should **PASS** with metrics meeting or exceeding targets.

### Common Issues

#### Slow Startup
**Symptoms:** Startup time exceeds targets
**Possible Causes:**
- Slow disk I/O
- Configuration file too large
- Network latency (if loading remote configs)

**Solutions:**
- Optimize configuration parsing
- Cache configuration data
- Use SSD storage

#### High Memory Usage
**Symptoms:** Memory usage exceeds targets
**Possible Causes:**
- Memory leaks
- Inefficient data structures
- Large configuration files

**Solutions:**
- Profile memory usage
- Optimize data structures
- Implement memory pooling

#### Low Throughput
**Symptoms:** Task throughput below targets
**Possible Causes:**
- CPU bottlenecks
- Lock contention
- Inefficient algorithms

**Solutions:**
- Profile CPU usage
- Reduce lock contention
- Optimize hot paths

#### Memory Leaks
**Symptoms:** Heap growth over time
**Possible Causes:**
- Unclosed resources
- Circular references
- Goroutine leaks

**Solutions:**
- Use defer for cleanup
- Profile memory allocations
- Check for goroutine leaks

## Continuous Monitoring

### CI/CD Integration
Add performance tests to your CI/CD pipeline:

```yaml
# .github/workflows/performance.yml
name: Performance Tests
on: [push, pull_request]
jobs:
  performance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - name: Run performance tests
        run: go test -v -run "TestPerformanceValidation" ./test
```

### Performance Regression Detection
Set up automated alerts for performance regressions:

```bash
# Run tests and compare with baseline
go test -bench=. -benchmem ./test > current.txt
# Compare with baseline.txt
# Alert if metrics degrade by > threshold
```

## Benchmarking

### Run Benchmarks
```bash
# Run all benchmarks
go test -bench=. -benchmem ./test ./internal/...

# Run specific benchmark
go test -bench=BenchmarkConfigLoad -benchmem ./test

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./test

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./test
```

### Analyze Profiles
```bash
# Analyze CPU profile
go tool pprof cpu.prof

# Analyze memory profile
go tool pprof mem.prof

# Generate flame graph
go tool pprof -http=:8080 cpu.prof
```

## Best Practices

### Running Tests
1. Run tests on a quiet system (no other heavy processes)
2. Run multiple times to account for variance
3. Use consistent hardware for comparisons
4. Disable CPU throttling if possible

### Interpreting Results
1. Look for trends, not absolute values
2. Compare against baselines, not arbitrary numbers
3. Consider variance and outliers
4. Focus on regressions, not minor fluctuations

### Maintaining Performance
1. Run performance tests regularly
2. Set up automated alerts
3. Profile before optimizing
4. Document performance characteristics
5. Review performance in code reviews

## Troubleshooting

### Tests Fail Intermittently
**Cause:** System load, timing issues
**Solution:** Run tests multiple times, increase timeouts

### Tests Fail on CI but Pass Locally
**Cause:** Different hardware, system load
**Solution:** Adjust targets for CI environment

### Memory Tests Show High Variance
**Cause:** GC timing, background processes
**Solution:** Force GC before measurements, run multiple times

### Throughput Tests Inconsistent
**Cause:** CPU throttling, system load
**Solution:** Disable throttling, run on dedicated hardware

## Resources

- **Test Suite:** `test/performance_validation_test.go`
- **Validation Script:** `scripts/run-performance-validation.sh`
- **Latest Report:** `PERFORMANCE_VALIDATION_REPORT.md`
- **Completion Summary:** `TASK_29.9_COMPLETION.md`

## Support

For questions or issues with performance validation:
1. Check the latest performance report
2. Review test output for specific failures
3. Compare with baseline metrics
4. Profile the application if needed
5. Consult the troubleshooting section

---

**Last Updated:** 2025-11-10  
**Test Suite Version:** 1.0  
**Maintained By:** Development Team
