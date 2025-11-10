# Testing Best Practices

This document outlines testing standards and best practices for the asc project.

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Types](#test-types)
- [Writing Good Tests](#writing-good-tests)
- [Test Organization](#test-organization)
- [Mocking and Fakes](#mocking-and-fakes)
- [Test Coverage](#test-coverage)
- [Running Tests](#running-tests)
- [Common Patterns](#common-patterns)

## Testing Philosophy

### Core Principles

1. **Tests are documentation** - Tests show how code should be used
2. **Fast feedback** - Tests should run quickly to encourage frequent execution
3. **Reliable** - Tests should be deterministic and not flaky
4. **Maintainable** - Tests should be easy to understand and update
5. **Comprehensive** - Test happy paths, edge cases, and error conditions

### Test Pyramid

We follow the test pyramid approach:

```
        /\
       /  \
      / E2E \      ← Few, slow, high-level
     /------\
    /  Integ \     ← Some, medium speed
   /----------\
  /    Unit    \   ← Many, fast, focused
 /--------------\
```

- **70% Unit tests** - Fast, focused, test individual functions
- **20% Integration tests** - Test component interactions
- **10% E2E tests** - Test complete user workflows

## Test Types

### Unit Tests

Test individual functions or methods in isolation.

**Characteristics:**
- Fast (< 1 second)
- No external dependencies
- Deterministic
- Test one thing

**Example:**

```go
func TestParseConfig_ValidTOML_ReturnsConfig(t *testing.T) {
    input := `[core]
beads_db_path = "./test"`
    
    got, err := ParseConfig(input)
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    want := Config{
        Core: CoreConfig{
            BeadsDBPath: "./test",
        },
    }
    
    if !reflect.DeepEqual(got, want) {
        t.Errorf("got %+v, want %+v", got, want)
    }
}
```

### Integration Tests

Test multiple components working together.

**Characteristics:**
- Medium speed (1-10 seconds)
- May use real dependencies (files, databases)
- Test component interactions
- More complex setup

**Example:**

```go
func TestProcessManager_StartStop_Integration(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    pm := NewProcessManager(tmpDir)
    
    // Start a process
    pid, err := pm.Start("test", "sleep", []string{"10"}, nil)
    if err != nil {
        t.Fatalf("failed to start: %v", err)
    }
    
    // Verify it's running
    if !pm.IsRunning(pid) {
        t.Error("process should be running")
    }
    
    // Stop it
    if err := pm.Stop(pid); err != nil {
        t.Errorf("failed to stop: %v", err)
    }
    
    // Verify it stopped
    if pm.IsRunning(pid) {
        t.Error("process should be stopped")
    }
}
```

### End-to-End Tests

Test complete user workflows from start to finish.

**Characteristics:**
- Slow (10+ seconds)
- Use real system
- Test user scenarios
- Complex setup and teardown

**Example:**

```go
// +build e2e

func TestE2E_InitUpDown(t *testing.T) {
    // Setup test environment
    tmpDir := t.TempDir()
    os.Chdir(tmpDir)
    
    // Run init
    cmd := exec.Command("asc", "init", "--non-interactive")
    if err := cmd.Run(); err != nil {
        t.Fatalf("init failed: %v", err)
    }
    
    // Run up
    cmd = exec.Command("asc", "up")
    if err := cmd.Start(); err != nil {
        t.Fatalf("up failed: %v", err)
    }
    defer cmd.Process.Kill()
    
    // Wait for startup
    time.Sleep(5 * time.Second)
    
    // Verify agents are running
    // ... verification code ...
    
    // Run down
    cmd = exec.Command("asc", "down")
    if err := cmd.Run(); err != nil {
        t.Errorf("down failed: %v", err)
    }
}
```

## Writing Good Tests

### Test Naming

Use descriptive names that explain what's being tested:

```go
// Good
func TestParseConfig_EmptyInput_ReturnsError(t *testing.T)
func TestProcessManager_StopNonExistent_ReturnsError(t *testing.T)
func TestTUI_Resize_UpdatesDimensions(t *testing.T)

// Bad
func TestParse(t *testing.T)
func TestStop(t *testing.T)
func TestResize(t *testing.T)
```

**Pattern:** `Test<Function>_<Scenario>_<ExpectedBehavior>`

### Table-Driven Tests

Use table-driven tests for multiple test cases:

```go
func TestValidateConfig(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid config",
            config: Config{
                Core: CoreConfig{BeadsDBPath: "./test"},
            },
            wantErr: false,
        },
        {
            name: "missing beads path",
            config: Config{
                Core: CoreConfig{BeadsDBPath: ""},
            },
            wantErr: true,
            errMsg:  "beads_db_path is required",
        },
        {
            name: "invalid path",
            config: Config{
                Core: CoreConfig{BeadsDBPath: "/nonexistent"},
            },
            wantErr: true,
            errMsg:  "path does not exist",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateConfig(tt.config)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
            }
        })
    }
}
```

### Test Helpers

Extract common setup into helper functions:

```go
func setupTestConfig(t *testing.T) Config {
    t.Helper()  // Mark as helper for better error reporting
    
    return Config{
        Core: CoreConfig{
            BeadsDBPath: t.TempDir(),
        },
    }
}

func TestSomething(t *testing.T) {
    cfg := setupTestConfig(t)
    // ... test code ...
}
```

### Cleanup

Use `t.Cleanup()` or `defer` for cleanup:

```go
func TestWithTempFile(t *testing.T) {
    f, err := os.CreateTemp("", "test")
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() {
        os.Remove(f.Name())
    })
    
    // Or use t.TempDir() which auto-cleans
    dir := t.TempDir()
    
    // ... test code ...
}
```

### Error Messages

Provide clear error messages:

```go
// Good
if got != want {
    t.Errorf("ParseConfig() = %v, want %v", got, want)
}

if err == nil {
    t.Fatal("expected error but got nil")
}

// Bad
if got != want {
    t.Error("wrong value")
}

if err == nil {
    t.Fatal("error")
}
```

## Test Organization

### File Structure

```
package/
├── file.go           # Implementation
├── file_test.go      # Unit tests
├── file_bench_test.go # Benchmarks (optional)
└── testdata/         # Test fixtures
    ├── valid.toml
    └── invalid.toml
```

### Test Package Names

```go
// Same package - can test private functions
package config

func TestPrivateFunction(t *testing.T) { }

// External package - tests public API only
package config_test

import "github.com/yourusername/asc/internal/config"

func TestPublicAPI(t *testing.T) { }
```

### Test Data

Use `testdata/` directory for test fixtures:

```go
func TestParseConfigFile(t *testing.T) {
    data, err := os.ReadFile("testdata/valid.toml")
    if err != nil {
        t.Fatal(err)
    }
    
    cfg, err := ParseConfig(string(data))
    // ... test code ...
}
```

## Mocking and Fakes

### Interface-Based Mocking

Define interfaces for dependencies:

```go
// Interface
type BeadsClient interface {
    GetTasks() ([]Task, error)
    CreateTask(title string) (Task, error)
}

// Mock implementation
type MockBeadsClient struct {
    GetTasksFunc    func() ([]Task, error)
    CreateTaskFunc  func(string) (Task, error)
}

func (m *MockBeadsClient) GetTasks() ([]Task, error) {
    if m.GetTasksFunc != nil {
        return m.GetTasksFunc()
    }
    return nil, nil
}

func (m *MockBeadsClient) CreateTask(title string) (Task, error) {
    if m.CreateTaskFunc != nil {
        return m.CreateTaskFunc(title)
    }
    return Task{}, nil
}

// Use in tests
func TestSomething(t *testing.T) {
    mock := &MockBeadsClient{
        GetTasksFunc: func() ([]Task, error) {
            return []Task{{ID: "1", Title: "Test"}}, nil
        },
    }
    
    // Test code using mock
}
```

### Fake Implementations

For complex dependencies, create fake implementations:

```go
// Fake in-memory beads client
type FakeBeadsClient struct {
    tasks map[string]Task
    mu    sync.Mutex
}

func NewFakeBeadsClient() *FakeBeadsClient {
    return &FakeBeadsClient{
        tasks: make(map[string]Task),
    }
}

func (f *FakeBeadsClient) GetTasks() ([]Task, error) {
    f.mu.Lock()
    defer f.mu.Unlock()
    
    tasks := make([]Task, 0, len(f.tasks))
    for _, t := range f.tasks {
        tasks = append(tasks, t)
    }
    return tasks, nil
}

func (f *FakeBeadsClient) CreateTask(title string) (Task, error) {
    f.mu.Lock()
    defer f.mu.Unlock()
    
    task := Task{
        ID:    fmt.Sprintf("%d", len(f.tasks)+1),
        Title: title,
    }
    f.tasks[task.ID] = task
    return task, nil
}
```

## Test Coverage

### Measuring Coverage

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out
```

### Coverage Goals

- **Overall:** 80%+ coverage
- **Critical paths:** 100% coverage (config parsing, process management)
- **Error handling:** All error paths tested
- **New code:** 80%+ coverage required

### What to Test

**Always test:**
- Public APIs
- Error conditions
- Edge cases (empty input, nil values, boundary conditions)
- Critical business logic

**Don't obsess over:**
- Trivial getters/setters
- Generated code
- Third-party library wrappers (test your usage, not the library)

## Running Tests

### Basic Commands

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/config

# Run specific test
go test -run TestParseConfig ./internal/config

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
```

### Using Make

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run e2e tests
make test-e2e

# Run all tests including e2e
make test-all
```

### Continuous Testing

Use `entr` or similar for continuous testing:

```bash
# Install entr
brew install entr  # macOS
apt install entr   # Linux

# Watch for changes and run tests
find . -name '*.go' | entr -c go test ./...
```

## Common Patterns

### Testing Errors

```go
func TestFunction_Error(t *testing.T) {
    _, err := Function(invalidInput)
    
    // Check error occurred
    if err == nil {
        t.Fatal("expected error but got nil")
    }
    
    // Check error message
    if !strings.Contains(err.Error(), "expected message") {
        t.Errorf("error = %q, want to contain %q", err.Error(), "expected message")
    }
    
    // Check error type
    var targetErr *CustomError
    if !errors.As(err, &targetErr) {
        t.Errorf("error type = %T, want *CustomError", err)
    }
}
```

### Testing Timeouts

```go
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    done := make(chan bool)
    go func() {
        // Long-running operation
        time.Sleep(1 * time.Second)
        done <- true
    }()
    
    select {
    case <-done:
        // Success
    case <-ctx.Done():
        t.Fatal("test timed out")
    }
}
```

### Testing Concurrency

```go
func TestConcurrent(t *testing.T) {
    const numGoroutines = 100
    
    var wg sync.WaitGroup
    wg.Add(numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        go func(id int) {
            defer wg.Done()
            // Concurrent operation
            result := DoSomething(id)
            if result == nil {
                t.Errorf("goroutine %d: got nil result", id)
            }
        }(i)
    }
    
    wg.Wait()
}
```

### Testing File Operations

```go
func TestFileOperation(t *testing.T) {
    // Use t.TempDir() for automatic cleanup
    dir := t.TempDir()
    
    file := filepath.Join(dir, "test.txt")
    
    // Write file
    if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
        t.Fatal(err)
    }
    
    // Test operation
    result, err := ReadAndProcess(file)
    if err != nil {
        t.Fatalf("ReadAndProcess() error = %v", err)
    }
    
    // Verify result
    if result != "expected" {
        t.Errorf("got %q, want %q", result, "expected")
    }
}
```

### Benchmarking

```go
func BenchmarkParseConfig(b *testing.B) {
    input := `[core]
beads_db_path = "./test"`
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = ParseConfig(input)
    }
}

// Run benchmarks
// go test -bench=. -benchmem ./internal/config
```

## Test Checklist

Before submitting code, ensure:

- [ ] All tests pass
- [ ] New functionality has tests
- [ ] Tests are focused and fast
- [ ] Error paths are tested
- [ ] Edge cases are covered
- [ ] Tests are deterministic (no flaky tests)
- [ ] Test names are descriptive
- [ ] Coverage is adequate (80%+)
- [ ] No commented-out tests
- [ ] Tests are well-organized

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Test Comments](https://github.com/golang/go/wiki/CodeReviewComments#tests)
- [Advanced Testing with Go](https://www.youtube.com/watch?v=8hQG7QlcLBk)

---

Remember: Good tests are an investment in code quality and maintainability. Take the time to write them well!
