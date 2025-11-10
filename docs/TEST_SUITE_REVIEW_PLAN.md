# Test Suite Review and Gap Remediation Plan

**Task**: 28.7 Review test suite outcomes and address gaps  
**Status**: Not Started  
**Created**: 2025-11-10

## Overview

This document outlines the comprehensive plan to review the current test suite, identify gaps and issues, and systematically address them to achieve high-quality, reliable test coverage across the Agent Stack Controller (asc) project.

## Objectives

1. **Achieve 80%+ code coverage** across all packages
2. **Eliminate all failing tests** or document justification for skips
3. **Fix all flaky tests** to ensure reliable CI/CD
4. **Improve test quality** and maintainability
5. **Expand test scenarios** to cover all critical user workflows
6. **Optimize test performance** to keep suite under 2 minutes
7. **Document all gaps** and create ongoing maintenance plan

## Current State

Based on the latest metrics (as of 2025-11-10):

- **Code Coverage**: 14.8% (target: 80%)
- **Test Files**: 24
- **Test Functions**: 279
- **Production Code**: 14,707 lines
- **Test Code**: 11,483 lines
- **Test/Code Ratio**: 78%

### Key Observations

- ✅ Good test volume (78% test/code ratio)
- ❌ Low code coverage (14.8% vs 80% target)
- ⚠️ Coverage gap of 65.2% needs to be addressed
- ✅ Quality gates infrastructure in place
- ✅ Flakiness detection automated

## Task Breakdown

### 28.7.1 Analyze Current Test Coverage and Identify Gaps

**Objective**: Understand where test coverage is lacking and prioritize improvements.

**Actions**:
1. Run comprehensive coverage analysis:
   ```bash
   make test-coverage
   go tool cover -html=coverage.out
   ```

2. Generate package-level coverage report:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -func=coverage.out | sort -k3 -n
   ```

3. Identify critical paths with <80% coverage:
   - Configuration parsing and validation
   - Process lifecycle management
   - TUI rendering and event handling
   - Client implementations (beads, MCP)
   - Error handling paths

4. Document gaps by priority:
   - **P0 (Critical)**: Core functionality, security-critical paths
   - **P1 (High)**: User-facing features, error handling
   - **P2 (Medium)**: Edge cases, less common paths
   - **P3 (Low)**: Nice-to-have coverage

5. Create action plan with estimated effort

**Deliverables**:
- Coverage gap analysis report
- Prioritized list of packages/functions needing tests
- Action plan with timeline

---

### 28.7.2 Review and Fix Failing Tests

**Objective**: Ensure all tests pass reliably or are properly documented.

**Actions**:
1. Identify all failing tests:
   ```bash
   go test -v ./... 2>&1 | grep -E "FAIL|--- FAIL"
   ```

2. Categorize failures:
   - **Bugs**: Tests failing due to actual bugs in code
   - **Outdated**: Tests need updating for new behavior
   - **Environment**: Tests fail due to missing dependencies
   - **Flaky**: Tests fail intermittently

3. Fix each category:
   - **Bugs**: Fix the underlying code issue
   - **Outdated**: Update test expectations
   - **Environment**: Document dependencies, add setup
   - **Flaky**: Address in task 28.7.3

4. Document any tests that must be skipped:
   - Add clear comments explaining why
   - Create issues to track re-enabling
   - Use `t.Skip()` with justification

**Deliverables**:
- All tests passing or properly skipped
- Documentation of skipped tests
- Bug fixes for failing code

---

### 28.7.3 Address Flaky Tests Identified by Monitoring

**Objective**: Eliminate test flakiness to ensure reliable CI/CD.

**Actions**:
1. Review flakiness reports:
   ```bash
   make test-flakiness RUNS=20
   cat test-flakiness-results/flakiness-report-*.md
   ```

2. Investigate root causes:
   - **Race conditions**: Use `go test -race`
   - **Timing issues**: Look for hardcoded sleeps
   - **External dependencies**: Check for network/filesystem deps
   - **Shared state**: Look for global variables

3. Fix flaky tests:
   - Add proper synchronization (channels, mutexes)
   - Replace `time.Sleep()` with condition waits
   - Mock external dependencies
   - Isolate test state

4. Common fixes:
   ```go
   // Bad: Hardcoded sleep
   time.Sleep(100 * time.Millisecond)
   
   // Good: Wait for condition
   require.Eventually(t, func() bool {
       return condition()
   }, 5*time.Second, 10*time.Millisecond)
   ```

5. Verify fixes with multiple runs:
   ```bash
   go test -count=50 -run TestFlakyTest ./...
   ```

**Deliverables**:
- All flaky tests fixed
- Flakiness rate <1%
- Documentation of fixes

---

### 28.7.4 Improve Test Quality and Maintainability

**Objective**: Make tests easier to understand, maintain, and extend.

**Actions**:
1. Refactor duplicated test code:
   - Extract common setup into helper functions
   - Create test fixtures for shared data
   - Use `TestMain` for package-level setup

2. Convert to table-driven tests:
   ```go
   func TestConfigParsing(t *testing.T) {
       tests := []struct {
           name    string
           input   string
           want    Config
           wantErr bool
       }{
           // Test cases...
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Test logic...
           })
       }
   }
   ```

3. Improve test naming:
   - Use descriptive names: `TestConfigParser_InvalidTOML_ReturnsError`
   - Follow pattern: `Test<Function>_<Scenario>_<ExpectedResult>`

4. Add test documentation:
   - Document complex test scenarios
   - Explain why certain mocks are used
   - Add comments for non-obvious assertions

5. Follow best practices:
   - Use `t.Parallel()` for independent tests
   - Use `t.Cleanup()` for resource cleanup
   - Use `require` for fatal assertions, `assert` for non-fatal

**Deliverables**:
- Refactored tests with less duplication
- Improved test naming and documentation
- Test helper functions and fixtures

---

### 28.7.5 Add Missing Unit Tests for Core Functionality

**Objective**: Achieve comprehensive unit test coverage for all core packages.

**Priority Packages** (based on coverage gaps):

1. **internal/config**:
   - Configuration parsing edge cases
   - Validation logic for all fields
   - Default value handling
   - Error path coverage

2. **internal/process**:
   - Process start/stop/restart
   - PID tracking and cleanup
   - Signal handling
   - Error recovery

3. **internal/tui**:
   - Model state updates
   - Event handling
   - View rendering logic
   - Layout calculations

4. **internal/beads**:
   - Task parsing and filtering
   - Git operations
   - Error handling

5. **internal/mcp**:
   - HTTP client operations
   - WebSocket handling
   - Message parsing
   - Reconnection logic

**Actions**:
1. For each package, identify uncovered functions:
   ```bash
   go test -coverprofile=coverage.out ./internal/config
   go tool cover -func=coverage.out | grep "0.0%"
   ```

2. Write unit tests for each uncovered function:
   - Test happy path
   - Test error paths
   - Test edge cases
   - Test boundary conditions

3. Ensure all exported functions have tests

4. Aim for 80%+ coverage per package

**Deliverables**:
- Unit tests for all core packages
- 80%+ coverage for critical packages
- Documentation of any intentionally uncovered code

---

### 28.7.6 Enhance Integration Test Coverage

**Objective**: Test multi-component interactions and workflows.

**Integration Test Scenarios**:

1. **Configuration Hot-Reload**:
   - Modify asc.toml while running
   - Verify config reloads without restart
   - Test adding/removing agents
   - Test updating agent configurations

2. **Health Monitoring and Auto-Recovery**:
   - Simulate agent crashes
   - Verify health monitor detects failure
   - Test auto-recovery kicks in
   - Verify recovery success

3. **WebSocket Reconnection**:
   - Disconnect WebSocket connection
   - Verify reconnection logic
   - Test fallback to polling
   - Verify data consistency

4. **Agent Lifecycle Management**:
   - Start multiple agents
   - Monitor agent status
   - Stop individual agents
   - Stop all agents
   - Verify cleanup

5. **Error Recovery and Graceful Degradation**:
   - Test with missing dependencies
   - Test with invalid configuration
   - Test with network failures
   - Verify graceful error handling

**Actions**:
1. Create integration test suite in `test/integration_test.go`
2. Use real components (not mocks) where possible
3. Test component interactions
4. Verify end-to-end workflows
5. Add cleanup to prevent test pollution

**Deliverables**:
- Comprehensive integration test suite
- Tests for all critical workflows
- Documentation of integration test setup

---

### 28.7.7 Expand E2E Test Scenarios

**Objective**: Test complete user workflows from start to finish.

**E2E Test Scenarios**:

1. **First-Time Setup**:
   ```bash
   asc init
   # Verify: config created, dependencies checked, test passed
   ```

2. **Start and Monitor**:
   ```bash
   asc up
   # Verify: agents started, TUI launched, status visible
   ```

3. **Agent Task Execution**:
   - Create task in beads
   - Verify agent picks up task
   - Monitor task progress
   - Verify task completion

4. **Multi-Agent Coordination**:
   - Multiple agents working on different phases
   - File lease coordination
   - Task handoff between agents

5. **Failure and Recovery**:
   - Kill agent process
   - Verify health monitor detects
   - Verify auto-recovery
   - Verify task resumption

6. **Graceful Shutdown**:
   ```bash
   asc down
   # Verify: agents stopped, cleanup complete
   ```

7. **Stress Testing**:
   - Many agents (10+)
   - Many tasks (1000+)
   - High message volume
   - Long-running stability

**Actions**:
1. Expand `test/e2e_comprehensive_test.go`
2. Add new E2E test scenarios
3. Use real dependencies (beads, MCP server)
4. Test on multiple platforms
5. Add stress test suite

**Deliverables**:
- Comprehensive E2E test suite
- Stress tests for high load
- Platform-specific test validation

---

### 28.7.8 Review and Improve Test Performance

**Objective**: Keep test suite fast (<2 minutes) while maintaining coverage.

**Actions**:
1. Identify slow tests:
   ```bash
   make test-timing
   cat test-timing-analysis.md
   ```

2. Optimize slow tests (>5s):
   - Add `t.Parallel()` for independent tests
   - Mock expensive I/O operations
   - Reduce test data size
   - Cache expensive setup

3. Reduce test setup overhead:
   - Use `TestMain` for package-level setup
   - Share fixtures across tests
   - Lazy-load test data

4. Mock expensive operations:
   ```go
   // Mock time for faster tests
   type Clock interface {
       Now() time.Time
   }
   
   // Mock filesystem operations
   type FileSystem interface {
       ReadFile(path string) ([]byte, error)
   }
   ```

5. Optimize test data generation:
   - Use smaller datasets
   - Generate data on-demand
   - Reuse test fixtures

6. Parallelize tests:
   ```go
   func TestSomething(t *testing.T) {
       t.Parallel() // Run in parallel with other tests
       // Test logic...
   }
   ```

**Deliverables**:
- Test suite completes in <2 minutes
- No tests >5 seconds (unit tests)
- Optimized test setup and teardown

---

### 28.7.9 Validate Test Environment and Dependencies

**Objective**: Ensure tests work reliably across all environments.

**Actions**:
1. Document all test dependencies:
   - Go version requirements
   - External tools (git, docker, etc.)
   - Environment variables
   - Platform-specific requirements

2. Verify tests in CI:
   - Check all CI test runs
   - Identify environment-specific failures
   - Fix or document platform differences

3. Test on multiple platforms:
   - Linux (Ubuntu latest)
   - macOS (latest)
   - Different Go versions (1.21, 1.22)

4. Add setup instructions:
   - Update TESTING.md
   - Document local test setup
   - Add troubleshooting guide

5. Handle platform differences:
   ```go
   // Skip tests on unsupported platforms
   if runtime.GOOS == "windows" {
       t.Skip("Test not supported on Windows")
   }
   ```

**Deliverables**:
- Documented test dependencies
- Tests passing on all platforms
- Setup instructions for local testing

---

### 28.7.10 Create Test Gap Remediation Report

**Objective**: Document all work done and create ongoing maintenance plan.

**Report Contents**:

1. **Executive Summary**:
   - Starting coverage: 14.8%
   - Final coverage: [target 80%+]
   - Tests added: [count]
   - Tests fixed: [count]
   - Flaky tests resolved: [count]

2. **Coverage Improvements**:
   - Package-by-package coverage gains
   - Critical paths now covered
   - Remaining gaps with justification

3. **Test Quality Improvements**:
   - Tests refactored
   - Performance improvements
   - Maintainability enhancements

4. **Issues Resolved**:
   - Failing tests fixed
   - Flaky tests eliminated
   - Environment issues resolved

5. **Remaining Gaps**:
   - Intentionally uncovered code
   - Justification for each gap
   - Plan to address in future

6. **Ongoing Maintenance Plan**:
   - Coverage monitoring strategy
   - Flakiness detection schedule
   - Test review cadence
   - Quality standards

7. **Lessons Learned**:
   - What worked well
   - What was challenging
   - Recommendations for future

**Deliverables**:
- Comprehensive remediation report
- Updated testing documentation
- Ongoing maintenance plan

---

## Success Criteria

### Coverage Targets

- ✅ Overall code coverage ≥80%
- ✅ Critical packages ≥90% coverage
- ✅ All exported functions have tests
- ✅ Error paths covered

### Test Quality

- ✅ All tests passing or properly skipped
- ✅ Flakiness rate <1%
- ✅ Test suite completes in <2 minutes
- ✅ Tests follow best practices

### Test Coverage

- ✅ Comprehensive unit tests
- ✅ Integration tests for workflows
- ✅ E2E tests for user scenarios
- ✅ Stress tests for high load

### Documentation

- ✅ All gaps documented
- ✅ Test setup instructions complete
- ✅ Ongoing maintenance plan created
- ✅ Lessons learned captured

---

## Timeline

**Estimated Duration**: 2-3 weeks

### Week 1: Analysis and Planning
- Days 1-2: Coverage analysis and gap identification (28.7.1)
- Days 3-4: Review and fix failing tests (28.7.2)
- Day 5: Address flaky tests (28.7.3)

### Week 2: Implementation
- Days 1-2: Add missing unit tests (28.7.5)
- Days 3-4: Enhance integration tests (28.7.6)
- Day 5: Expand E2E tests (28.7.7)

### Week 3: Optimization and Documentation
- Days 1-2: Improve test quality and performance (28.7.4, 28.7.8)
- Day 3: Validate test environment (28.7.9)
- Days 4-5: Create remediation report (28.7.10)

---

## Tools and Resources

### Coverage Analysis
```bash
# Generate coverage report
make test-coverage

# View coverage by package
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

### Flakiness Detection
```bash
# Check for flaky tests
make test-flakiness RUNS=20

# Run specific test multiple times
go test -count=50 -run TestName ./...
```

### Performance Analysis
```bash
# Analyze test timing
make test-timing

# Profile tests
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

### Quality Checks
```bash
# Run all quality checks
make quality

# Run tests with race detector
go test -race ./...
```

---

## Best Practices

### Writing Tests

1. **Follow AAA Pattern**: Arrange, Act, Assert
2. **Use Table-Driven Tests**: For multiple scenarios
3. **Test One Thing**: Each test should verify one behavior
4. **Use Descriptive Names**: Clear test names explain intent
5. **Avoid Test Interdependence**: Tests should be isolated

### Test Organization

1. **Group Related Tests**: Use subtests with `t.Run()`
2. **Use Test Helpers**: Extract common setup/teardown
3. **Keep Tests Close**: Test files next to code files
4. **Use Test Fixtures**: For complex test data

### Test Maintenance

1. **Review Tests in PRs**: Treat tests as production code
2. **Update Tests with Code**: Keep tests in sync
3. **Monitor Coverage**: Track coverage trends
4. **Fix Flaky Tests**: Don't ignore intermittent failures

---

## References

- [Testing Guide](../TESTING.md)
- [Quality Metrics Dashboard](./QUALITY_METRICS.md)
- [Quality Gates Implementation](./QUALITY_GATES_IMPLEMENTATION.md)
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)

---

**Created**: 2025-11-10  
**Status**: Not Started  
**Next Review**: After task completion

