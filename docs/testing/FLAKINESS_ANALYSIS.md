# Test Flakiness Analysis

**Date:** 2025-11-10  
**Status:** Analysis Complete

## Overview

This document identifies potential sources of test flakiness in the Agent Stack Controller test suite and provides recommendations for fixes.

## Identified Flaky Patterns

### 1. Time-Based Delays (time.Sleep)

#### internal/process/error_handling_test.go

**Location:** Multiple occurrences
**Pattern:** Using `time.Sleep()` to wait for process state changes

**Examples:**
```go
// Line 132: Waiting for process to be killed
syscall.Kill(pid, syscall.SIGKILL)
time.Sleep(100 * time.Millisecond)

// Line 214: Waiting for process to exit
time.Sleep(100 * time.Millisecond)

// Line 257: Waiting for process to be killed
syscall.Kill(pid1, syscall.SIGKILL)
time.Sleep(100 * time.Millisecond)
```

**Risk Level:** Medium
**Failure Mode:** On slow systems or under load, 100ms may not be enough for process termination

**Recommended Fix:**
```go
// Replace time.Sleep with proper wait condition
func waitForProcessExit(pid int, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if !isProcessRunning(pid) {
            return nil
        }
        time.Sleep(10 * time.Millisecond)
    }
    return fmt.Errorf("process %d did not exit within %v", pid, timeout)
}

// Usage:
syscall.Kill(pid, syscall.SIGKILL)
if err := waitForProcessExit(pid, 1*time.Second); err != nil {
    t.Fatal(err)
}
```

#### internal/mcp/error_handling_test.go

**Location:** Line 114
**Pattern:** Using `time.Sleep()` to simulate server timeout

**Example:**
```go
serverFunc: func(w http.ResponseWriter, r *http.Request) {
    time.Sleep(3 * time.Second)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("[]"))
}
```

**Risk Level:** Low (intentional timeout test)
**Failure Mode:** Test takes 3 seconds to run, slowing down test suite

**Recommended Fix:**
```go
// Use context with timeout instead
serverFunc: func(w http.ResponseWriter, r *http.Request) {
    select {
    case <-r.Context().Done():
        // Client timed out as expected
        return
    case <-time.After(3 * time.Second):
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("[]"))
    }
}
```

### 2. Race Conditions

#### internal/process/error_handling_test.go

**Location:** Line 418 (TestConcurrentAccess)
**Pattern:** Concurrent operations without proper synchronization

**Example:**
```go
_, err := mgr.Start(name, "sleep", []string{"5"}, nil)
// Multiple goroutines may access manager concurrently
```

**Risk Level:** High
**Failure Mode:** Race detector may catch issues, or tests may fail intermittently

**Recommended Fix:**
```go
// Add proper synchronization
var wg sync.WaitGroup
var mu sync.Mutex
errors := []error{}

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        name := fmt.Sprintf("concurrent-%d", id)
        _, err := mgr.Start(name, "sleep", []string{"1"}, nil)
        if err != nil {
            mu.Lock()
            errors = append(errors, err)
            mu.Unlock()
        }
    }(i)
}

wg.Wait()
if len(errors) > 0 {
    t.Errorf("Concurrent operations failed: %v", errors)
}
```

### 3. External Dependencies

#### test/e2e_test.go and test/integration_test.go

**Pattern:** Tests depend on external services (beads, MCP server)
**Risk Level:** High
**Failure Mode:** Tests fail if services are not running or are slow to respond

**Recommended Fix:**
- Use test fixtures and mocks for unit tests
- Reserve E2E tests for actual integration testing
- Add retry logic with exponential backoff for network operations
- Use health checks before running tests

### 4. File System Operations

**Pattern:** Tests create temporary files and directories
**Risk Level:** Low
**Failure Mode:** Tests may fail if file system is slow or full

**Current Mitigation:** Using `t.TempDir()` which auto-cleans up
**Status:** ✅ Good practice already in use

### 5. Process Lifecycle Timing

**Pattern:** Tests start processes and immediately check their status
**Risk Level:** Medium
**Failure Mode:** Process may not be fully started when status is checked

**Example:**
```go
pid, err := mgr.Start("test", "sleep", []string{"10"}, nil)
if !mgr.IsRunning(pid) {  // May fail if process not fully started
    t.Error("Process should be running")
}
```

**Recommended Fix:**
```go
pid, err := mgr.Start("test", "sleep", []string{"10"}, nil)
// Wait for process to be fully started
if err := waitForProcessRunning(pid, 1*time.Second); err != nil {
    t.Fatal(err)
}
if !mgr.IsRunning(pid) {
    t.Error("Process should be running")
}
```

## Flakiness Metrics

### Current State
- **Total time.Sleep calls in tests:** 5
- **Tests with race conditions:** 1 (identified)
- **Tests with external dependencies:** Multiple (E2E and integration)
- **Average test duration:** Unknown (need to run timing analysis)

### Target State
- **time.Sleep calls:** 0 (replace with proper wait conditions)
- **Race conditions:** 0 (add proper synchronization)
- **External dependencies:** Isolated to E2E tests only
- **Test suite duration:** <2 minutes

## Recommended Fixes by Priority

### High Priority (Blocking)

1. **Replace time.Sleep with wait conditions in process tests**
   - Estimated effort: 2 hours
   - Impact: Eliminates most common source of flakiness
   - Files: `internal/process/error_handling_test.go`

2. **Add synchronization to concurrent tests**
   - Estimated effort: 1 hour
   - Impact: Prevents race conditions
   - Files: `internal/process/error_handling_test.go`

3. **Add retry logic to network operations**
   - Estimated effort: 3 hours
   - Impact: Makes tests resilient to transient failures
   - Files: `internal/mcp/*_test.go`, `internal/beads/*_test.go`

### Medium Priority

1. **Optimize timeout test in MCP**
   - Estimated effort: 30 minutes
   - Impact: Reduces test duration by 3 seconds
   - Files: `internal/mcp/error_handling_test.go`

2. **Add health checks before E2E tests**
   - Estimated effort: 2 hours
   - Impact: Prevents E2E test failures due to service unavailability
   - Files: `test/e2e_test.go`, `test/integration_test.go`

### Low Priority

1. **Add test timing analysis**
   - Estimated effort: 1 hour
   - Impact: Identifies slow tests for optimization
   - Files: New test infrastructure

2. **Add flakiness monitoring**
   - Estimated effort: 2 hours
   - Impact: Detects flaky tests automatically
   - Files: CI/CD configuration

## Implementation Plan

### Phase 1: Fix Critical Flakiness (Week 1)
1. Create helper functions for wait conditions
2. Replace all time.Sleep calls with proper waits
3. Add synchronization to concurrent tests
4. Run tests 20+ times to verify fixes

### Phase 2: Add Resilience (Week 2)
1. Add retry logic to network operations
2. Add health checks to E2E tests
3. Optimize timeout tests
4. Document testing best practices

### Phase 3: Monitor and Maintain (Ongoing)
1. Set up flakiness monitoring in CI
2. Add test timing analysis
3. Review and update tests regularly
4. Track flakiness metrics

## Helper Functions to Add

### Wait Conditions

```go
// internal/process/testing_helpers.go

package process

import (
    "fmt"
    "time"
)

// WaitForProcessExit waits for a process to exit within the timeout
func WaitForProcessExit(pid int, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        // Check if process exists
        process, err := os.FindProcess(pid)
        if err != nil {
            return nil // Process doesn't exist
        }
        
        // Try to signal the process
        err = process.Signal(syscall.Signal(0))
        if err != nil {
            return nil // Process is dead
        }
        
        time.Sleep(10 * time.Millisecond)
    }
    return fmt.Errorf("process %d did not exit within %v", pid, timeout)
}

// WaitForProcessRunning waits for a process to be running within the timeout
func WaitForProcessRunning(pid int, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        process, err := os.FindProcess(pid)
        if err != nil {
            time.Sleep(10 * time.Millisecond)
            continue
        }
        
        // Try to signal the process
        err = process.Signal(syscall.Signal(0))
        if err == nil {
            return nil // Process is running
        }
        
        time.Sleep(10 * time.Millisecond)
    }
    return fmt.Errorf("process %d did not start within %v", pid, timeout)
}

// WaitForCondition waits for a condition to be true within the timeout
func WaitForCondition(condition func() bool, timeout time.Duration, checkInterval time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if condition() {
            return nil
        }
        time.Sleep(checkInterval)
    }
    return fmt.Errorf("condition not met within %v", timeout)
}
```

### Retry Logic

```go
// internal/testing/retry.go

package testing

import (
    "fmt"
    "time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay time.Duration
    Multiplier float64
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() RetryConfig {
    return RetryConfig{
        MaxAttempts: 3,
        InitialDelay: 100 * time.Millisecond,
        MaxDelay: 5 * time.Second,
        Multiplier: 2.0,
    }
}

// Retry executes fn with exponential backoff
func Retry(config RetryConfig, fn func() error) error {
    var lastErr error
    delay := config.InitialDelay
    
    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        if attempt > 0 {
            time.Sleep(delay)
            delay = time.Duration(float64(delay) * config.Multiplier)
            if delay > config.MaxDelay {
                delay = config.MaxDelay
            }
        }
        
        if err := fn(); err == nil {
            return nil
        } else {
            lastErr = err
        }
    }
    
    return fmt.Errorf("failed after %d attempts: %w", config.MaxAttempts, lastErr)
}
```

## Testing Best Practices

### DO:
- ✅ Use `t.TempDir()` for temporary files
- ✅ Use wait conditions instead of time.Sleep
- ✅ Add proper synchronization for concurrent tests
- ✅ Use retry logic for network operations
- ✅ Clean up resources in defer statements
- ✅ Use table-driven tests for multiple scenarios
- ✅ Run tests with `-race` flag to detect race conditions

### DON'T:
- ❌ Use fixed time.Sleep for synchronization
- ❌ Assume operations complete instantly
- ❌ Share state between test cases
- ❌ Depend on test execution order
- ❌ Use global variables without synchronization
- ❌ Leave processes or files after test completion

## Monitoring and Metrics

### CI/CD Integration
```yaml
# .github/workflows/test-quality.yml
- name: Run flakiness check
  run: |
    for i in {1..20}; do
      go test -race ./... || echo "Run $i failed" >> failures.txt
    done
    if [ -f failures.txt ]; then
      echo "Flaky tests detected:"
      cat failures.txt
      exit 1
    fi
```

### Metrics to Track
- Test execution time (per test and total)
- Flakiness rate (failures per 100 runs)
- Race condition detections
- Timeout occurrences
- External dependency failures

## Next Steps

1. ✅ Identify all sources of flakiness (COMPLETE)
2. [ ] Implement helper functions for wait conditions
3. [ ] Replace time.Sleep calls with proper waits
4. [ ] Add synchronization to concurrent tests
5. [ ] Run tests 20+ times to verify fixes
6. [ ] Add retry logic to network operations
7. [ ] Set up flakiness monitoring in CI
8. [ ] Document testing best practices

---

**Last Updated:** 2025-11-10  
**Next Review:** 2025-11-11  
**Owner:** Test Infrastructure Team
