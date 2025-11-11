# Integration Testing Guide

## Overview

This guide explains how to run integration validation tests for the Agent Stack Controller (asc). Integration tests validate that all components work together correctly in real-world scenarios.

## Quick Start

### Run All Tests

```bash
./scripts/run-integration-validation.sh
```

This will run all integration tests that can execute in your current environment and skip tests that require additional setup.

### Run Specific Test

```bash
go test -tags=integration ./test/integration_validation_test.go -v -run TestName
```

Replace `TestName` with the specific test you want to run.

## Test Categories

### Phase 1: Basic Workflow Tests

Tests fundamental workflows like initialization and configuration management.

```bash
# Run init workflow test
go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_InitWorkflow

# Run config hot-reload test
go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_ConfigHotReload

# Run config templates test
go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_ConfigTemplates
```

### Phase 2: Process Management Tests

Tests agent lifecycle and process management.

```bash
# Run multi-agent coordination test
go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_MultiAgentCoordination

# Run error recovery test
go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_ErrorRecovery
```

### Phase 3: Security Tests

Tests secrets encryption and security features.

**Requirements:** `age` and `age-keygen` binaries must be installed.

```bash
# Install age (macOS)
brew install age

# Install age (Linux)
# See https://github.com/FiloSottile/age

# Run secrets test
go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_SecretsEncryptionDecryption
```

### Phase 4: External Integration Tests

Tests integration with external tools and services.

```bash
# Test beads integration (requires bd CLI)
go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_RealBeadsRepository

# Test MCP server integration (requires running MCP server)
INTEGRATION_MCP=true go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_RealMCPServer
```

### Phase 5: Complete Workflow Tests

Tests end-to-end workflows with all components.

**Requirements:** All dependencies installed and services running.

```bash
# Run complete workflow test
INTEGRATION_FULL=true go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_CompleteWorkflow

# Run up → work → down workflow
INTEGRATION_FULL=true go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_UpWorkDown

# Run health monitoring test
INTEGRATION_FULL=true go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_HealthMonitoringAndRecovery
```

### Phase 6: Stress Tests

Tests system behavior under load.

```bash
# Run stress test (starts 20 agents)
INTEGRATION_STRESS=true go test -tags=integration ./test/integration_validation_test.go -v -run TestIntegrationValidation_StressTest
```

## Environment Variables

### INTEGRATION_FULL

Enables full integration tests that require all dependencies and services.

```bash
INTEGRATION_FULL=true ./scripts/run-integration-validation.sh
```

**Requirements:**
- All binaries installed (git, python3, bd, age)
- MCP server running
- Beads repository initialized

### INTEGRATION_MCP

Enables MCP server integration tests.

```bash
INTEGRATION_MCP=true ./scripts/run-integration-validation.sh
```

**Requirements:**
- MCP Agent Mail server running on http://localhost:8765

### INTEGRATION_STRESS

Enables stress tests that create many agents and processes.

```bash
INTEGRATION_STRESS=true ./scripts/run-integration-validation.sh
```

**Note:** Stress tests may take several minutes to complete.

## Prerequisites

### Required Binaries

| Binary | Required For | Installation |
|--------|-------------|--------------|
| `go` | All tests | https://golang.org/dl/ |
| `git` | All tests | Pre-installed on most systems |
| `bd` | Beads tests | https://github.com/steveyegge/beads |
| `age` | Secrets tests | `brew install age` (macOS) |
| `python3` | Agent tests | Pre-installed on most systems |

### Optional Services

| Service | Required For | Setup |
|---------|-------------|-------|
| MCP Agent Mail | MCP tests | `python -m mcp_agent_mail.server` |
| Beads Repository | Beads tests | `bd init` in a git repo |

## Test Structure

### Integration Test File

**Location:** `test/integration_validation_test.go`

**Tests Included:**

1. `TestIntegrationValidation_InitWorkflow` - Init command validation
2. `TestIntegrationValidation_UpWorkDown` - Complete lifecycle
3. `TestIntegrationValidation_ConfigHotReload` - Config watching
4. `TestIntegrationValidation_SecretsEncryptionDecryption` - Age encryption
5. `TestIntegrationValidation_HealthMonitoringAndRecovery` - Health checks
6. `TestIntegrationValidation_RealBeadsRepository` - Beads integration
7. `TestIntegrationValidation_RealMCPServer` - MCP integration
8. `TestIntegrationValidation_MultiAgentCoordination` - Multi-agent scenarios
9. `TestIntegrationValidation_CompleteWorkflow` - End-to-end workflow
10. `TestIntegrationValidation_ErrorRecovery` - Error handling
11. `TestIntegrationValidation_ConfigTemplates` - Template system
12. `TestIntegrationValidation_StressTest` - Load testing

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: |
          # Install age for secrets tests
          sudo apt-get update
          sudo apt-get install -y age
          
          # Install bd for beads tests
          go install github.com/steveyegge/beads/cmd/bd@latest
      
      - name: Build asc
        run: make build
      
      - name: Run integration tests
        run: ./scripts/run-integration-validation.sh
```

### GitLab CI Example

```yaml
integration-tests:
  stage: test
  image: golang:1.21
  before_script:
    - apt-get update && apt-get install -y age git
    - go install github.com/steveyegge/beads/cmd/bd@latest
  script:
    - make build
    - ./scripts/run-integration-validation.sh
```

## Troubleshooting

### Test Fails: "asc binary not built"

**Solution:** Build the binary first:
```bash
make build
```

### Test Skipped: "age not installed"

**Solution:** Install age:
```bash
# macOS
brew install age

# Linux
sudo apt-get install age
```

### Test Skipped: "bd not installed"

**Solution:** Install beads:
```bash
go install github.com/steveyegge/beads/cmd/bd@latest
```

### Test Fails: "MCP server not running"

**Solution:** Start the MCP server:
```bash
python -m mcp_agent_mail.server
```

### Test Timeout

Some tests may take longer on slower systems. Increase the timeout:
```bash
go test -tags=integration -timeout 30m ./test/integration_validation_test.go -v
```

## Best Practices

### Before Committing

Always run integration tests before committing:
```bash
./scripts/run-integration-validation.sh
```

### Before Releasing

Run full integration tests including stress tests:
```bash
INTEGRATION_FULL=true INTEGRATION_STRESS=true ./scripts/run-integration-validation.sh
```

### Continuous Testing

Set up a cron job or CI/CD pipeline to run integration tests regularly:
```bash
# Run daily at 2 AM
0 2 * * * cd /path/to/asc && ./scripts/run-integration-validation.sh
```

## Test Output

### Successful Run

```
=========================================
ASC Integration Validation Suite
=========================================

Phase 1: Basic Workflow Tests
------------------------------
Running: Init Workflow
  ✓ PASSED

Running: Config Hot-Reload
  ✓ PASSED

...

=========================================
Integration Validation Summary
=========================================

✓ All tests passed!

Integration validation completed successfully.
```

### Failed Run

```
Running: Init Workflow
  ✗ FAILED
--- FAIL: TestIntegrationValidation_InitWorkflow (0.03s)
    integration_validation_test.go:74: Failed to load config: ...

=========================================
Integration Validation Summary
=========================================

✗ 1 test(s) failed:
  - Init Workflow

Please review the failures above and fix any issues.
```

## Additional Resources

- [Testing Documentation](./TESTING.md)
- [Development Guide](./DEVELOPER_EXPERIENCE.md)
- [Troubleshooting Guide](../TROUBLESHOOTING.md)
- [Task 29.7 Validation Report](../TASK_29.7_INTEGRATION_VALIDATION_REPORT.md)

## Support

If you encounter issues with integration tests:

1. Check the [Troubleshooting](#troubleshooting) section above
2. Review test output for specific error messages
3. Ensure all prerequisites are installed
4. Check that services are running (if required)
5. Open an issue with test output and environment details
