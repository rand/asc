# Test Quality Improvements

**Date:** 2025-11-10  
**Status:** Recommendations

## Overview

This document provides recommendations for improving test quality and maintainability across the Agent Stack Controller test suite.

## Current State Analysis

### Test Organization
- ✅ Tests are organized by package
- ✅ Separate files for error handling tests
- ⚠️ Some duplication across test files
- ⚠️ Inconsistent naming conventions
- ❌ Limited use of table-driven tests

### Test Documentation
- ⚠️ Some tests have comments
- ❌ Many tests lack explanation of what they're testing
- ❌ No documentation of test setup requirements
- ❌ Edge cases not clearly documented

### Test Maintainability
- ⚠️ Some helper functions exist
- ❌ Significant code duplication
- ❌ Hard-coded values scattered throughout
- ❌ Brittle assertions (exact string matching)

## Improvement Recommendations

### 1. Adopt Table-Driven Test Pattern

**Current State:**
```go
func TestCheckBinary_ErrorPaths(t *testing.T) {
    checker := NewChecker("asc.toml", ".env")
    
    // Test 1
    result := checker.CheckBinary("nonexistent-binary-12345")
    if result.Status != CheckFail {
        t.Error("Expected fail")
    }
    
    // Test 2
    result = checker.CheckBinary("")
    if result.Status != CheckFail {
        t.Error("Expected fail")
    }
    
    // ... more tests
}
```

**Improved:**
```go
func TestCheckBinary(t *testing.T) {
    tests := []struct {
        name         string
        binary       string
        wantStatus   CheckStatus
        wantContains string
    }{
        {
            name:         "nonexistent binary",
            binary:       "nonexistent-binary-12345",
            wantStatus:   CheckFail,
            wantContains: "not found",
        },
        {
            name:         "empty binary name",
            binary:       "",
            wantStatus:   CheckFail,
            wantContains: "empty",
        },
        {
            name:       "valid binary",
            binary:     "go",
            wantStatus: CheckPass,
        },
    }
    
    checker := NewChecker("asc.toml", ".env")
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := checker.CheckBinary(tt.binary)
            
            if result.Status != tt.wantStatus {
                t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
            }
            
            if tt.wantContains != "" && !strings.Contains(result.Message, tt.wantContains) {
                t.Errorf("Message %q does not contain %q", result.Message, tt.wantContains)
            }
        })
    }
}
```

**Benefits:**
- Easier to add new test cases
- Clear test case documentation
- Consistent test structure
- Better failure messages

### 2. Extract Common Test Helpers

**Create:** `internal/testing/helpers.go`

```go
package testing

import (
    "os"
    "path/filepath"
    "testing"
)

// CreateTempConfig creates a temporary config file with the given content
func CreateTempConfig(t *testing.T, content string) string {
    t.Helper()
    dir := t.TempDir()
    path := filepath.Join(dir, "asc.toml")
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        t.Fatal(err)
    }
    return path
}

// CreateTempEnv creates a temporary .env file with the given content
func CreateTempEnv(t *testing.T, content string) string {
    t.Helper()
    dir := t.TempDir()
    path := filepath.Join(dir, ".env")
    if err := os.WriteFile(path, []byte(content), 0600); err != nil {
        t.Fatal(err)
    }
    return path
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, got, want string) {
    t.Helper()
    if !strings.Contains(got, want) {
        t.Errorf("String %q does not contain %q", got, want)
    }
}

// AssertError checks if an error occurred when expected
func AssertError(t *testing.T, err error, wantErr bool) {
    t.Helper()
    if (err != nil) != wantErr {
        t.Errorf("Error = %v, wantErr %v", err, wantErr)
    }
}

// AssertEqual checks if two values are equal
func AssertEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()
    if got != want {
        t.Errorf("Got %v, want %v", got, want)
    }
}
```

**Usage:**
```go
func TestLoadConfig(t *testing.T) {
    configPath := testhelpers.CreateTempConfig(t, `
[core]
beads_db_path = "./test"
`)
    
    cfg, err := LoadConfig(configPath)
    testhelpers.AssertError(t, err, false)
    testhelpers.AssertEqual(t, cfg.Core.BeadsDBPath, "./test")
}
```

### 3. Improve Test Naming

**Current:**
```go
func TestCheckBinary_ErrorPaths(t *testing.T)
func TestCheckFile_ErrorPaths(t *testing.T)
func TestInvalidInput(t *testing.T)
```

**Improved:**
```go
func TestChecker_CheckBinary(t *testing.T)
func TestChecker_CheckFile(t *testing.T)
func TestChecker_InvalidInput(t *testing.T)
```

**Pattern:** `Test<Type>_<Method>` or `Test<Function>`

**Benefits:**
- Clear what is being tested
- Groups related tests together
- Easier to find tests for specific functionality

### 4. Add Test Documentation

**Template:**
```go
// TestChecker_CheckBinary verifies that CheckBinary correctly identifies
// binaries in the system PATH and returns appropriate status codes.
//
// Test cases:
// - Valid binary (go, git) should return CheckPass
// - Nonexistent binary should return CheckFail
// - Empty binary name should return CheckFail
// - Binary with special characters should return CheckFail
//
// Edge cases:
// - Binary with path traversal attempts
// - Binary with null bytes
// - Very long binary names
func TestChecker_CheckBinary(t *testing.T) {
    // ...
}
```

### 5. Use Subtests for Better Organization

**Current:**
```go
func TestMultipleScenarios(t *testing.T) {
    // Test scenario 1
    // Test scenario 2
    // Test scenario 3
}
```

**Improved:**
```go
func TestMultipleScenarios(t *testing.T) {
    t.Run("scenario 1", func(t *testing.T) {
        // Test scenario 1
    })
    
    t.Run("scenario 2", func(t *testing.T) {
        // Test scenario 2
    })
    
    t.Run("scenario 3", func(t *testing.T) {
        // Test scenario 3
    })
}
```

**Benefits:**
- Can run individual scenarios: `go test -run TestMultipleScenarios/scenario_1`
- Better failure reporting
- Parallel execution possible with `t.Parallel()`

### 6. Reduce Assertion Brittleness

**Current (Brittle):**
```go
if err.Error() != "configuration file not found: /path/to/file" {
    t.Error("Wrong error message")
}
```

**Improved (Flexible):**
```go
if !strings.Contains(err.Error(), "configuration file not found") {
    t.Errorf("Expected error about missing config, got: %v", err)
}

// Or use error wrapping
if !errors.Is(err, ErrConfigNotFound) {
    t.Errorf("Expected ErrConfigNotFound, got: %v", err)
}
```

### 7. Add Test Constants

**Create:** `internal/testing/constants.go`

```go
package testing

const (
    // Test timeouts
    ShortTimeout  = 1 * time.Second
    MediumTimeout = 5 * time.Second
    LongTimeout   = 30 * time.Second
    
    // Test data
    ValidConfigTOML = `
[core]
beads_db_path = "./test"

[services.mcp_agent_mail]
url = "http://localhost:8765"

[agent.test]
command = "python"
model = "claude"
phases = ["planning"]
`
    
    InvalidConfigTOML = `[invalid`
    
    ValidEnv = `
CLAUDE_API_KEY=test-key
OPENAI_API_KEY=test-key
GOOGLE_API_KEY=test-key
`
)
```

### 8. Use Test Fixtures

**Create:** `testdata/` directory structure

```
testdata/
├── configs/
│   ├── valid.toml
│   ├── invalid.toml
│   ├── minimal.toml
│   └── complete.toml
├── envs/
│   ├── valid.env
│   ├── missing-keys.env
│   └── malformed.env
└── tasks/
    ├── sample-tasks.jsonl
    └── empty-tasks.jsonl
```

**Usage:**
```go
func TestLoadConfig_ValidFile(t *testing.T) {
    cfg, err := LoadConfig("testdata/configs/valid.toml")
    if err != nil {
        t.Fatalf("Failed to load valid config: %v", err)
    }
    // Assertions...
}
```

### 9. Add Parallel Test Execution

**Pattern:**
```go
func TestIndependentOperation(t *testing.T) {
    t.Parallel() // Mark test as safe to run in parallel
    
    // Test code...
}

func TestWithSubtests(t *testing.T) {
    tests := []struct{
        name string
        // ...
    }{
        // test cases...
    }
    
    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Each subtest runs in parallel
            // Test code...
        })
    }
}
```

**Benefits:**
- Faster test execution
- Identifies tests with shared state issues
- Better resource utilization

### 10. Add Test Coverage Helpers

**Create:** `internal/testing/coverage.go`

```go
package testing

import (
    "testing"
)

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T, reason string) {
    t.Helper()
    if testing.Short() {
        t.Skipf("Skipping in short mode: %s", reason)
    }
}

// RequireEnv skips the test if an environment variable is not set
func RequireEnv(t *testing.T, key string) string {
    t.Helper()
    value := os.Getenv(key)
    if value == "" {
        t.Skipf("Skipping: %s environment variable not set", key)
    }
    return value
}

// MarkIntegration marks a test as an integration test
func MarkIntegration(t *testing.T) {
    t.Helper()
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
}
```

## Implementation Plan

### Phase 1: Foundation (Week 1)
1. Create `internal/testing/` package with helpers
2. Add test constants and fixtures
3. Document testing best practices
4. Create test templates

### Phase 2: Refactor Existing Tests (Week 2-3)
1. Convert tests to table-driven pattern
2. Extract common setup into helpers
3. Improve test naming
4. Add test documentation

### Phase 3: Enhance Quality (Week 4)
1. Add parallel execution where safe
2. Reduce assertion brittleness
3. Add coverage helpers
4. Review and optimize

## Testing Best Practices Document

**Create:** `TESTING.md` (update existing)

```markdown
# Testing Best Practices

## Test Structure

### Use Table-Driven Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name string
        input string
        want string
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test code
        })
    }
}
```

### Use Helper Functions
- Mark helpers with `t.Helper()`
- Extract common setup/teardown
- Create assertion helpers

### Name Tests Clearly
- Pattern: `Test<Type>_<Method>`
- Use descriptive subtest names
- Document what is being tested

## Test Organization

### File Structure
- `*_test.go` in same package
- `testdata/` for fixtures
- `internal/testing/` for helpers

### Test Categories
- Unit tests: Fast, isolated
- Integration tests: Multiple components
- E2E tests: Full system

## Test Quality

### DO:
- ✅ Use `t.TempDir()` for temp files
- ✅ Use `t.Cleanup()` for cleanup
- ✅ Use `t.Helper()` in helpers
- ✅ Use `t.Parallel()` when safe
- ✅ Use table-driven tests
- ✅ Test error paths
- ✅ Document edge cases

### DON'T:
- ❌ Use `time.Sleep()` for sync
- ❌ Share state between tests
- ❌ Depend on test order
- ❌ Use exact string matching
- ❌ Leave resources uncleaned
- ❌ Skip error checking

## Running Tests

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestName ./package

# Run in short mode (skip slow tests)
go test -short ./...

# Run with verbose output
go test -v ./...
```

## Coverage Goals

- Overall: 80%+
- Critical paths: 90%+
- Error handling: 100%
```

## Metrics and Tracking

### Test Quality Metrics
- Test coverage percentage
- Number of table-driven tests
- Test execution time
- Test flakiness rate
- Code duplication in tests

### Goals
- 80%+ coverage across all packages
- 90%+ of tests use table-driven pattern
- <2 minute test suite execution
- <1% flakiness rate
- <10% code duplication in tests

## Next Steps

1. ✅ Document current state (COMPLETE)
2. ✅ Create improvement recommendations (COMPLETE)
3. [ ] Create testing helpers package
4. [ ] Add test constants and fixtures
5. [ ] Refactor high-priority tests
6. [ ] Update TESTING.md documentation
7. [ ] Train team on best practices
8. [ ] Set up quality metrics tracking

---

**Last Updated:** 2025-11-10  
**Next Review:** 2025-11-17  
**Owner:** Test Infrastructure Team
