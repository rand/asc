# End-to-End Testing Guide

This document describes the comprehensive end-to-end tests for the Agent Stack Controller (asc).

## Overview

The e2e tests validate the complete system behavior including:
- Complete agent stack startup and shutdown
- Agent task execution from beads to completion
- Multi-agent coordination and file lease conflicts
- Error recovery scenarios (agent crash, MCP disconnect)
- Long-running stability (24+ hour runs)
- Resource cleanup (PIDs, logs, temp files)
- Graceful degradation (missing dependencies, network issues)

## Test Files

- `e2e_test.go` - Basic end-to-end tests
- `e2e_comprehensive_test.go` - Comprehensive system tests

## Running Tests

### Prerequisites

1. Build the asc binary:
```bash
make build
```

2. Ensure dependencies are installed:
- git
- python3
- bd (beads CLI) - optional for full tests
- mcp_agent_mail - optional for full tests

### Basic E2E Tests

Run basic e2e tests (no external dependencies required):
```bash
go test -tags=e2e ./test -v
```

### Full E2E Tests

Run full e2e tests (requires all dependencies):
```bash
E2E_FULL=true go test -tags=e2e ./test -v -run TestE2EComplete
```

### Long-Running Stability Tests

Run 24-hour stability test:
```bash
E2E_LONG=true go test -tags=e2e ./test -v -run TestE2ELongRunning -timeout 25h
```

For shorter stability test:
```bash
E2E_LONG=true go test -tags=e2e ./test -v -run TestE2ELongRunning -timeout 10m -short
```

### Stress Tests

Run stress tests with rapid operations:
```bash
E2E_STRESS=true go test -tags=e2e ./test -v -run TestE2EStress
```

### Performance Tests

Run performance baseline tests:
```bash
E2E_PERF=true go test -tags=e2e ./test -v -run TestE2EPerformance
```

### All Comprehensive Tests

Run all comprehensive tests:
```bash
E2E_FULL=true E2E_STRESS=true E2E_PERF=true go test -tags=e2e ./test -v -run TestE2E
```

## Test Categories

### 1. Complete Stack Lifecycle
- **TestE2ECompleteStackStartupShutdown**: Tests full startup and shutdown sequence
- Validates service startup, agent launching, and graceful shutdown
- Verifies resource cleanup after shutdown

### 2. Agent Task Execution
- **TestE2EAgentTaskExecution**: Tests agent task workflow
- Creates tasks in beads
- Monitors agent task pickup and completion
- Validates task state transitions

### 3. Multi-Agent Coordination
- **TestE2EMultiAgentCoordination**: Tests multiple agents working together
- Validates file lease conflict resolution
- Tests task distribution among agents
- Verifies coordination through MCP

### 4. Error Recovery
- **TestE2EErrorRecovery**: Tests error handling and recovery
- Agent crash recovery
- MCP server disconnect and reconnect
- Beads sync failure handling

### 5. Long-Running Stability
- **TestE2ELongRunningStability**: Tests system stability over time
- Runs for 24 hours (or 5 minutes with -short flag)
- Performs periodic health checks
- Monitors memory usage and orphaned processes

### 6. Resource Cleanup
- **TestE2EResourceCleanup**: Tests resource management
- PID file cleanup
- Log file management
- Temporary file cleanup

### 7. Graceful Degradation
- **TestE2EGracefulDegradation**: Tests system behavior with issues
- Missing dependencies
- Network connectivity issues
- Corrupted configuration files

### 8. Stress Testing
- **TestE2EStressTest**: Tests system under load
- Rapid start/stop cycles
- Concurrent command execution
- Resource exhaustion scenarios

### 9. Data Integrity
- **TestE2EDataIntegrity**: Tests data consistency
- Configuration file persistence
- Environment file security
- Log rotation

### 10. Security Validation
- **TestE2ESecurityValidation**: Tests security features
- API key protection in logs
- File permission validation
- Command injection prevention

### 11. Documentation Validation
- **TestE2EDocumentation**: Tests documentation examples
- README examples work correctly
- Help output is accurate

## Environment Variables

- `E2E_FULL=true` - Enable full e2e tests (requires all dependencies)
- `E2E_LONG=true` - Enable long-running stability tests
- `E2E_STRESS=true` - Enable stress tests
- `E2E_PERF=true` - Enable performance tests

## Test Output

Tests log detailed information about:
- Command execution and output
- Resource usage
- Error conditions
- Performance metrics

Use `-v` flag for verbose output:
```bash
go test -tags=e2e ./test -v
```

## Continuous Integration

For CI environments, run a subset of tests:
```bash
# Quick validation (no external dependencies)
go test -tags=e2e ./test -v -short

# Full validation (with dependencies)
E2E_FULL=true go test -tags=e2e ./test -v -timeout 30m
```

## Troubleshooting

### Tests Skip

If tests are skipped, check:
1. Build the asc binary: `make build`
2. Set appropriate environment variables
3. Install required dependencies

### Tests Fail

Common issues:
1. **Port conflicts**: MCP server port 8765 may be in use
2. **Missing dependencies**: Install git, python3, bd
3. **Permissions**: Ensure write access to test directories
4. **Timeouts**: Increase timeout with `-timeout` flag

### Cleanup After Failed Tests

If tests fail and leave resources:
```bash
# Kill any orphaned processes
pkill -f "asc"

# Clean up test directories
rm -rf /tmp/go-build*
rm -rf ~/.asc/pids/*
```

## Adding New Tests

When adding new e2e tests:

1. Use the `// +build e2e` tag
2. Add appropriate skip conditions for optional dependencies
3. Use helper functions for common setup
4. Clean up resources in defer or cleanup functions
5. Log detailed information for debugging
6. Use descriptive test names

Example:
```go
func TestE2ENewFeature(t *testing.T) {
    if os.Getenv("E2E_FULL") != "true" {
        t.Skip("Skipping test (set E2E_FULL=true to run)")
    }

    tmpDir := t.TempDir()
    setupTestEnvironment(t, tmpDir)
    defer cleanup(t, tmpDir)

    // Test implementation
}
```

## Performance Benchmarks

Expected performance baselines:
- Check command: < 5 seconds
- Service startup: < 10 seconds
- Config load (50 agents): < 10 seconds
- Graceful shutdown: < 5 seconds

## Test Coverage

The e2e tests cover:
- ✅ Complete stack lifecycle
- ✅ Agent task execution workflow
- ✅ Multi-agent coordination
- ✅ Error recovery scenarios
- ✅ Long-running stability
- ✅ Resource cleanup
- ✅ Graceful degradation
- ✅ Stress testing
- ✅ Data integrity
- ✅ Security validation
- ✅ Documentation accuracy

## Future Enhancements

Planned test additions:
- WebSocket real-time updates
- Configuration hot-reload
- Health monitoring and auto-recovery
- Agent template system
- Distributed agent coordination
