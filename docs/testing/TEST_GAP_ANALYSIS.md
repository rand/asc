# Test Gap Analysis Report

**Generated:** 2025-11-10  
**Overall Coverage:** 18.4%  
**Target Coverage:** 80%

## Executive Summary

This report documents the current state of test coverage across the Agent Stack Controller (asc) project, identifies critical gaps, and provides an action plan for remediation.

### Key Findings

1. **Build Failures:** 4 packages have failing tests that prevent compilation
2. **Low Coverage:** Multiple packages are significantly below the 80% target
3. **Untested Code:** Large portions of TUI and wizard code have 0% coverage
4. **Test Quality Issues:** Several tests have outdated signatures and assumptions

## Package-Level Coverage Analysis

### ✅ Passing Packages

| Package | Coverage | Status | Priority |
|---------|----------|--------|----------|
| `internal/errors` | 100.0% | ✅ Excellent | Low |
| `internal/health` | 72.0% | ⚠️ Below Target | Medium |
| `internal/logger` | 67.7% | ⚠️ Below Target | Medium |
| `internal/secrets` | 47.4% | ❌ Critical Gap | High |
| `internal/tui` | 2.5% | ❌ Critical Gap | High |
| `cmd/*` | 0.0% | ❌ Critical Gap | High |
| `main.go` | 0.0% | ❌ Critical Gap | Medium |

### ❌ Failing Packages (Build Errors)

| Package | Issue | Impact |
|---------|-------|--------|
| `internal/check` | Outdated test signatures for `NewChecker()` | Cannot measure coverage |
| `internal/process` | Type mismatches in error handling tests | Cannot measure coverage |
| `internal/beads` | Missing constructor parameters | Cannot measure coverage |
| `internal/mcp` | Undefined `NewClient` function | Cannot measure coverage |
| `internal/config` | Test assertion failures | Partial coverage (76.6%) |

## Critical Gaps by Category

### 1. Build Failures (Highest Priority)

#### internal/check/error_handling_test.go
**Issue:** `NewChecker()` called without required arguments
```
not enough arguments in call to NewChecker
    have ()
    want (string, string)
```

**Root Cause:** Test file not updated after API changes to `NewChecker` constructor

**Impact:** 
- Cannot run any tests in `internal/check` package
- Zero coverage visibility for dependency checking logic
- Blocks CI/CD pipeline

**Action Items:**
- [ ] Update all `NewChecker()` calls to include `configPath` and `envPath` parameters
- [ ] Update `CheckConfig()` and `CheckEnv()` calls to match current signatures
- [ ] Update `RunAll()` calls to remove parameters

#### internal/process/error_handling_test.go
**Issue:** Type mismatches between string process names and int PIDs
```
cannot use tt.processName (variable of type string) as int value in argument to mgr.Stop
```

**Root Cause:** Process manager API changed from string-based to PID-based identification

**Impact:**
- Cannot run any tests in `internal/process` package
- Zero coverage for critical process lifecycle management
- No validation of graceful shutdown logic

**Action Items:**
- [ ] Update all `Stop()` calls to use PIDs instead of process names
- [ ] Update `IsRunning()` calls to use PIDs
- [ ] Remove references to non-existent `GetPID()` method
- [ ] Fix error handling logic that assumes `error` is iterable

#### internal/beads/error_handling_test.go
**Issue:** Missing `time.Duration` parameter in `NewClient()` calls
```
not enough arguments in call to NewClient
    have (string)
    want (string, time.Duration)
```

**Root Cause:** Client constructor updated to include refresh interval parameter

**Impact:**
- Cannot run beads client tests
- No coverage for Git-backed task database integration
- Cannot validate task CRUD operations

**Action Items:**
- [ ] Add `time.Duration` parameter to all `NewClient()` calls
- [ ] Fix task status type mismatches (string vs *string)
- [ ] Resolve duplicate `contains` function declaration

#### internal/mcp/error_handling_test.go
**Issue:** `NewClient` function is undefined
```
undefined: NewClient
```

**Root Cause:** MCP client uses different constructor pattern or is not exported

**Impact:**
- Cannot run MCP client tests
- No coverage for agent communication layer
- Cannot validate WebSocket and HTTP client logic

**Action Items:**
- [ ] Identify correct constructor function (likely `NewHTTPClient` or `NewWebSocketClient`)
- [ ] Update all test instantiations
- [ ] Verify client interface matches test expectations

### 2. Test Assertion Failures

#### internal/config/error_handling_test.go
**Failing Tests:**
- `TestLoadConfig_ErrorPaths/missing_config_file`
- `TestLoadConfig_ErrorPaths/invalid_TOML_syntax`
- `TestLoadConfig_ErrorPaths/empty_config_file`
- `TestLoadConfig_ErrorPaths/missing_required_fields`
- `TestValidate_ErrorPaths/agent_with_empty_model`
- `TestValidate_ErrorPaths/agent_with_empty_phases`
- `TestLoadEnv_ErrorPaths/missing_env_file`
- `TestLoadEnv_ErrorPaths/malformed_env_file`
- `TestRecoveryFromTransientErrors`

**Issue:** Error message format changes and validation order changes

**Example:**
```
Expected error to contain "no such file", got: "configuration file not found: /path/to/file"
```

**Root Cause:** Error messages were improved for user-friendliness but tests expect old format

**Impact:**
- 9 failing test cases
- Coverage reported as 76.6% but with failures
- False confidence in error handling paths

**Action Items:**
- [ ] Update error message assertions to match new format
- [ ] Fix validation order expectations (PATH checks before model checks)
- [ ] Update recovery test to handle validation failures

### 3. Coverage Gaps (<80% Target)

#### internal/health (72.0% coverage)
**Missing Coverage:**
- Edge cases in health check logic
- Recovery action execution paths
- Concurrent health monitoring scenarios

**Action Items:**
- [ ] Add tests for health check timeout scenarios
- [ ] Test recovery action failures
- [ ] Add concurrent monitoring tests

#### internal/logger (67.7% coverage)
**Missing Coverage:**
- Log rotation logic
- Aggregation edge cases
- Concurrent logging scenarios

**Action Items:**
- [ ] Add tests for log file rotation
- [ ] Test log aggregation with high volume
- [ ] Add concurrent logging tests

#### internal/secrets (47.4% coverage)
**Missing Coverage:**
- Encryption/decryption error paths
- Key derivation edge cases
- File permission handling

**Action Items:**
- [ ] Add tests for encryption failures
- [ ] Test key derivation with various inputs
- [ ] Add file permission validation tests

### 4. Critical Untested Code

#### internal/tui (2.5% coverage)
**Untested Components:**
- All wizard screens (0% coverage)
- Most rendering functions (0% coverage)
- Layout calculations (0% coverage)
- Event handling (minimal coverage)

**Impact:**
- No validation of user-facing UI
- Cannot detect rendering regressions
- No coverage for interactive flows

**Action Items:**
- [ ] Add unit tests for layout calculations
- [ ] Add tests for wizard state machine
- [ ] Test rendering functions with mock data
- [ ] Add integration tests for user flows

#### cmd/* (0% coverage)
**Untested Components:**
- All CLI commands (init, up, down, check, test, services)
- Command-line argument parsing
- Command execution flows

**Impact:**
- No validation of primary user interface
- Cannot detect CLI regressions
- No coverage for command orchestration

**Action Items:**
- [ ] Add integration tests for each command
- [ ] Test command-line argument parsing
- [ ] Test command error handling
- [ ] Add E2E tests for command workflows

## Detailed Gap Analysis by Package

### internal/check
**Current Coverage:** Unknown (build failed)  
**Target Coverage:** 80%  
**Gap:** 80%+

**Untested Functionality:**
- Binary existence checks
- File accessibility validation
- Configuration validation
- Environment variable checks
- Check result formatting
- Error path handling

**Critical Paths:**
- Dependency verification before startup
- Configuration validation
- Environment setup verification

**Recommended Tests:**
- [ ] Test binary checks with missing binaries
- [ ] Test file checks with various permission scenarios
- [ ] Test config validation with malformed TOML
- [ ] Test env checks with missing keys
- [ ] Test check result formatting and styling

### internal/process
**Current Coverage:** Unknown (build failed)  
**Target Coverage:** 80%  
**Gap:** 80%+

**Untested Functionality:**
- Process startup and PID tracking
- Graceful shutdown with SIGTERM
- Force kill with SIGKILL
- Process status monitoring
- PID file management
- Log file capture

**Critical Paths:**
- Agent process lifecycle
- Graceful shutdown sequence
- Process health monitoring

**Recommended Tests:**
- [ ] Test process start with valid/invalid commands
- [ ] Test graceful shutdown timeout
- [ ] Test force kill after timeout
- [ ] Test PID file creation and cleanup
- [ ] Test process status detection
- [ ] Test concurrent process management

### internal/beads
**Current Coverage:** Unknown (build failed)  
**Target Coverage:** 80%  
**Gap:** 80%+

**Untested Functionality:**
- Task CRUD operations
- Git refresh mechanism
- JSON parsing
- Error handling for bd CLI failures
- Task filtering by status

**Critical Paths:**
- Task retrieval for agent assignment
- Task status updates
- Git synchronization

**Recommended Tests:**
- [ ] Test task creation and retrieval
- [ ] Test task updates
- [ ] Test task deletion
- [ ] Test git refresh with conflicts
- [ ] Test JSON parsing errors
- [ ] Test bd CLI failures

### internal/mcp
**Current Coverage:** Unknown (build failed)  
**Target Coverage:** 80%  
**Gap:** 80%+

**Untested Functionality:**
- HTTP client operations
- WebSocket client operations
- Message sending and receiving
- Agent status tracking
- Connection error handling
- Retry logic

**Critical Paths:**
- Agent communication
- Status monitoring
- WebSocket reconnection

**Recommended Tests:**
- [ ] Test HTTP client with various responses
- [ ] Test WebSocket connection and reconnection
- [ ] Test message parsing
- [ ] Test agent status updates
- [ ] Test connection failures and retries
- [ ] Test concurrent message handling

### internal/config
**Current Coverage:** 76.6% (with failures)  
**Target Coverage:** 80%  
**Gap:** 3.4% + fix failures

**Untested Functionality:**
- Some edge cases in validation
- Template generation edge cases
- Hot-reload edge cases

**Critical Paths:**
- Configuration loading and validation
- Template-based config generation
- Configuration hot-reload

**Recommended Tests:**
- [ ] Fix all failing error message assertions
- [ ] Add tests for template edge cases
- [ ] Add tests for hot-reload race conditions
- [ ] Test validation with complex configs

### internal/tui
**Current Coverage:** 2.5%  
**Target Coverage:** 80%  
**Gap:** 77.5%

**Untested Functionality:**
- Wizard screens (0%)
- Layout calculations (0%)
- Rendering functions (0%)
- Event handling (minimal)
- Theme application (0%)
- Animation logic (0%)

**Critical Paths:**
- User onboarding (wizard)
- Dashboard rendering
- Interactive controls

**Recommended Tests:**
- [ ] Test wizard state transitions
- [ ] Test layout calculations with various sizes
- [ ] Test rendering with mock data
- [ ] Test keyboard event handling
- [ ] Test theme application
- [ ] Test animation frame generation

### cmd/*
**Current Coverage:** 0%  
**Target Coverage:** 80%  
**Gap:** 80%

**Untested Functionality:**
- All commands (init, up, down, check, test, services)
- Argument parsing
- Command orchestration
- Error handling

**Critical Paths:**
- User command execution
- System initialization
- Agent lifecycle management

**Recommended Tests:**
- [ ] Test each command with valid inputs
- [ ] Test each command with invalid inputs
- [ ] Test command error handling
- [ ] Test command orchestration
- [ ] Add E2E tests for workflows

## Action Plan

### Phase 1: Fix Build Failures (Week 1)
**Priority:** Critical  
**Estimated Effort:** 2-3 days

1. Fix `internal/check` test signatures
2. Fix `internal/process` test type mismatches
3. Fix `internal/beads` constructor calls
4. Fix `internal/mcp` client instantiation
5. Fix `internal/config` error message assertions

**Success Criteria:**
- All tests compile and run
- No build failures in CI
- Baseline coverage established for all packages

### Phase 2: Address Critical Gaps (Week 1-2)
**Priority:** High  
**Estimated Effort:** 3-4 days

1. Increase `internal/secrets` coverage to 80%
2. Add core tests for `internal/tui` (target 30%)
3. Add integration tests for `cmd/*` (target 50%)
4. Fix remaining test failures in `internal/config`

**Success Criteria:**
- `internal/secrets` reaches 80% coverage
- `internal/tui` reaches 30% coverage
- `cmd/*` reaches 50% coverage
- All tests pass

### Phase 3: Reach Target Coverage (Week 2-3)
**Priority:** Medium  
**Estimated Effort:** 4-5 days

1. Increase `internal/health` to 80%
2. Increase `internal/logger` to 80%
3. Increase `internal/tui` to 80%
4. Increase `cmd/*` to 80%
5. Add missing unit tests for all packages

**Success Criteria:**
- All packages reach 80% coverage
- Overall project coverage reaches 80%
- All critical paths tested

### Phase 4: Enhance Test Quality (Week 3-4)
**Priority:** Medium  
**Estimated Effort:** 2-3 days

1. Refactor tests with duplication
2. Add table-driven tests
3. Improve test documentation
4. Add helper functions
5. Optimize test performance

**Success Criteria:**
- Tests follow best practices
- Test suite runs in <2 minutes
- Tests are maintainable and clear

## Coverage Targets by Package

| Package | Current | Phase 1 | Phase 2 | Phase 3 | Final Target |
|---------|---------|---------|---------|---------|--------------|
| `internal/errors` | 100.0% | 100.0% | 100.0% | 100.0% | 100.0% |
| `internal/health` | 72.0% | 72.0% | 72.0% | 80.0% | 80.0% |
| `internal/logger` | 67.7% | 67.7% | 67.7% | 80.0% | 80.0% |
| `internal/secrets` | 47.4% | 47.4% | 80.0% | 80.0% | 80.0% |
| `internal/check` | 0.0%* | 60.0% | 70.0% | 80.0% | 80.0% |
| `internal/process` | 0.0%* | 60.0% | 70.0% | 80.0% | 80.0% |
| `internal/beads` | 0.0%* | 60.0% | 70.0% | 80.0% | 80.0% |
| `internal/mcp` | 0.0%* | 60.0% | 70.0% | 80.0% | 80.0% |
| `internal/config` | 76.6% | 80.0% | 80.0% | 80.0% | 80.0% |
| `internal/tui` | 2.5% | 2.5% | 30.0% | 80.0% | 80.0% |
| `cmd/*` | 0.0% | 0.0% | 50.0% | 80.0% | 80.0% |
| **Overall** | **18.4%** | **50.0%** | **65.0%** | **80.0%** | **80.0%** |

*Build failures prevent coverage measurement

## Risk Assessment

### High Risk Areas

1. **Process Management** - Critical for agent lifecycle, currently untested
2. **MCP Communication** - Core coordination layer, currently untested
3. **TUI Wizard** - Primary user onboarding, 0% coverage
4. **Command Orchestration** - Main user interface, 0% coverage

### Medium Risk Areas

1. **Secrets Management** - Security-critical, only 47.4% coverage
2. **Health Monitoring** - Important for reliability, 72% coverage
3. **Configuration Hot-Reload** - Complex feature, needs more testing

### Low Risk Areas

1. **Error Handling** - 100% coverage, well-tested
2. **Logger** - 67.7% coverage, core paths tested

## Recommendations

### Immediate Actions (This Week)

1. **Fix all build failures** - Blocks all other testing work
2. **Fix failing config tests** - Restore confidence in error handling
3. **Establish baseline coverage** - Get accurate measurements for all packages

### Short-Term Actions (Next 2 Weeks)

1. **Focus on critical paths** - Process management, MCP, commands
2. **Add integration tests** - Test component interactions
3. **Increase secrets coverage** - Security-critical code

### Long-Term Actions (Next Month)

1. **Reach 80% coverage target** - All packages
2. **Improve test quality** - Refactor, document, optimize
3. **Add E2E tests** - Complete user workflows
4. **Set up coverage monitoring** - Prevent regressions

### Process Improvements

1. **Pre-commit hooks** - Run tests before commit
2. **Coverage gates** - Require 80% for new code
3. **Regular reviews** - Weekly test quality reviews
4. **Documentation** - Keep test docs up to date

## Appendix A: Test Execution Summary

```
Test Execution Date: 2025-11-10
Go Version: 1.21+
Platform: macOS (darwin)

Package Results:
✅ internal/errors     - 100.0% coverage - PASS
✅ internal/health     - 72.0% coverage  - PASS
✅ internal/logger     - 67.7% coverage  - PASS
✅ internal/secrets    - 47.4% coverage  - PASS
✅ internal/tui        - 2.5% coverage   - PASS
❌ internal/check      - Build failed
❌ internal/process    - Build failed
❌ internal/beads      - Build failed
❌ internal/mcp        - Build failed
⚠️  internal/config    - 76.6% coverage  - FAIL (9 tests)
❌ cmd/*               - 0.0% coverage   - Not tested
❌ main.go             - 0.0% coverage   - Not tested

Overall: 18.4% coverage
```

## Appendix B: Detailed Error Log

See full test output in CI logs or run locally:
```bash
go test -v -coverprofile=coverage.out ./...
```

## Next Steps

1. Review this analysis with the team
2. Prioritize action items based on risk and effort
3. Assign owners for each phase
4. Set up tracking for coverage improvements
5. Schedule weekly progress reviews

---

**Report Prepared By:** Kiro AI Assistant  
**Last Updated:** 2025-11-10  
**Next Review:** 2025-11-17
