# Task 28.10 Completion: Issue Detection and Remediation

## Overview

Task 28.10 has been successfully completed, implementing comprehensive issue detection and remediation capabilities for the Agent Stack Controller (asc).

## Completed Sub-tasks

### ✅ 1. Health Check Diagnostics (asc doctor command)

**Implementation**: `cmd/doctor.go` and `internal/doctor/doctor.go`

The `asc doctor` command provides comprehensive diagnostics with the following features:

- **Multiple diagnostic categories**:
  - Configuration validation
  - State integrity (PIDs, logs)
  - File permissions
  - System resources
  - Network connectivity
  - Agent configuration

- **Command-line options**:
  - `asc doctor` - Run diagnostics
  - `asc doctor --verbose` - Show detailed information
  - `asc doctor --fix` - Automatically fix issues
  - `asc doctor --json` - Output in JSON format

- **Severity levels**: Critical, High, Medium, Low, Info
- **Issue categories**: Configuration, State, Permissions, Resources, Network, Agent

### ✅ 2. Automatic Issue Detection

**Implementation**: `internal/doctor/doctor.go` (checkConfiguration, checkState, checkPermissions, checkResources, checkNetwork, checkAgents)

The doctor automatically detects:

**Configuration Issues**:
- Missing configuration file (asc.toml)
- Invalid TOML syntax
- Missing .env file
- Insecure .env permissions (not 0600)

**State Issues**:
- Corrupted PID files (invalid JSON)
- Orphaned PID files (process not running)
- Large log directories (>100MB)

**Permission Issues**:
- ~/.asc not a directory
- ~/.asc not writable
- Missing subdirectories (pids, logs, playbooks)

**Resource Issues**:
- High disk usage
- Missing required binaries (git, python3, uv, bd)

**Network Issues**:
- MCP server connectivity (informational)

**Agent Issues**:
- Missing agent command
- Invalid model configuration
- Missing phases configuration

### ✅ 3. Actionable Remediation Steps

**Implementation**: Each `Issue` struct includes:
- Clear title and description
- Impact explanation
- Specific remediation instructions
- Auto-fixable flag

**Example remediation messages**:
- "Run 'chmod 600 .env' to secure the file"
- "Delete the corrupted file: rm ~/.asc/pids/corrupted.json"
- "Create directory: mkdir -p ~/.asc/logs"
- "Install git and ensure it's in your PATH"

### ✅ 4. Test Recovery from Corrupted State

**Implementation**: `internal/doctor/doctor_test.go`

Comprehensive test coverage for recovery scenarios:

1. **TestRecoveryFromCorruptedPID**
   - Creates corrupted PID file with invalid JSON
   - Verifies detection
   - Applies fix
   - Confirms file removal

2. **TestRecoveryFromOrphanedPID**
   - Creates PID file for non-existent process
   - Verifies detection
   - Applies fix
   - Confirms cleanup

3. **TestRecoveryFromLargeLogs**
   - Creates old log files (>7 days)
   - Verifies detection when size exceeds threshold
   - Applies fix
   - Confirms old logs removed, recent logs preserved

4. **TestRecoveryFromPermissionIssues**
   - Creates .env with insecure permissions (0644)
   - Verifies detection
   - Applies fix
   - Confirms permissions set to 0600

5. **TestRecoveryFromMissingDirectories**
   - Tests with missing ~/.asc subdirectories
   - Verifies detection of all missing dirs
   - Applies fixes
   - Confirms directories created

6. **TestDoctorWithMultipleIssues**
   - Creates multiple issues simultaneously
   - Verifies all detected
   - Applies all fixes
   - Confirms all resolved
   - Re-runs diagnostics to verify clean state

**Test Results**:
- All tests passing ✅
- Coverage: 69.8% of statements
- 13 test cases total

### ✅ 5. Self-Healing Capabilities

**Implementation**: `internal/doctor/doctor.go` (ApplyFixes method)

Auto-fixable issues with `--fix` flag:

1. **fixEnvPermissions**: Sets .env to 0600
2. **fixAscNotDir**: Removes file, creates directory
3. **fixAscNotWritable**: Sets ~/.asc to 0755
4. **fixLargeLogs**: Deletes logs older than 7 days
5. **fixCorruptedPID**: Removes corrupted PID files
6. **fixOrphanedPID**: Removes orphaned PID files
7. **fixMissingDir**: Creates missing subdirectories

**Fix reporting**:
- Each fix tracked with success/failure status
- Detailed messages for each fix attempt
- Timestamp of when fix was applied
- Logged to ~/.asc/logs/asc.log

### ✅ 6. Issue Reporting Template

**Implementation**: `.github/ISSUE_TEMPLATE/bug_report.md`

Comprehensive bug report template includes:
- Bug description
- Steps to reproduce
- Expected vs actual behavior
- Environment information (OS, versions, terminal)
- Configuration (with PII removal reminder)
- Logs section
- Screenshots
- Possible solution
- Checklist for completeness

### ✅ 7. Documentation of Known Issues and Workarounds

**Implementation**: Multiple documentation files

1. **docs/KNOWN_ISSUES.md**
   - Active issues tracking
   - Self-healing capabilities documentation
   - Running diagnostics guide
   - Resolved issues archive
   - Monitoring sources
   - Reporting guidelines
   - Workaround patterns
   - Testing procedures
   - Communication protocols

2. **docs/COMMON_USER_ISSUES.md**
   - Comprehensive troubleshooting guide
   - 7 major categories:
     - Installation and Setup
     - Configuration Issues
     - Agent Problems
     - TUI Display Issues
     - Performance Issues
     - Network and Connectivity
     - File and Permission Issues
   - Each issue includes:
     - Symptoms
     - Causes
     - Multiple solutions
     - Prevention tips
   - Quick reference section
   - Emergency recovery procedures

## Test Coverage

### Unit Tests
- 13 test cases in `internal/doctor/doctor_test.go`
- Coverage: 69.8% of statements
- All tests passing

### Test Categories
1. Basic functionality tests (NewDoctor, HasCriticalIssues, ToJSON, Format)
2. Helper function tests (isProcessRunning, getDirSize, generateHealthSummary)
3. Recovery scenario tests (6 comprehensive tests)

### Test Quality
- Uses temporary directories for isolation
- Tests both detection and remediation
- Verifies fix success
- Confirms clean state after fixes
- Tests multiple issues simultaneously

## Usage Examples

### Basic Diagnostics
```bash
# Run diagnostics
asc doctor

# Verbose output
asc doctor --verbose

# JSON output for automation
asc doctor --json
```

### Automatic Fixes
```bash
# Fix all auto-fixable issues
asc doctor --fix

# Fix with verbose output
asc doctor --fix --verbose
```

### Example Output
```
╔════════════════════════════════════════════════════════════════╗
║              ASC DOCTOR - DIAGNOSTIC REPORT                    ║
╚════════════════════════════════════════════════════════════════╝

Run at: 2025-11-10 15:30:45
Status: Found 3 issue(s): 1 high, 2 medium

─── HIGH SEVERITY (1) ───

⚠ Environment file not found
  Category: configuration
  Remediation: Create a .env file with required API keys

─── MEDIUM SEVERITY (2) ───

! Insecure .env file permissions
  Category: permissions
  Remediation: Run 'chmod 600 .env' to secure the file
  ✓ Auto-fixable with --fix flag

! Orphaned PID file
  Category: state
  Remediation: Delete the orphaned file: rm ~/.asc/pids/agent.json
  ✓ Auto-fixable with --fix flag
```

## Integration with Existing Systems

### Works With
- **Check command**: `asc check` for dependency verification
- **Process manager**: Detects orphaned processes
- **Logger**: All diagnostics logged
- **Config system**: Validates configuration files

### Complements
- **TROUBLESHOOTING.md**: Links to doctor command
- **COMMON_USER_ISSUES.md**: References doctor for automated fixes
- **KNOWN_ISSUES.md**: Documents self-healing capabilities

## Benefits

1. **Reduced Support Burden**: Users can self-diagnose and fix common issues
2. **Faster Problem Resolution**: Automated detection and remediation
3. **Better User Experience**: Clear, actionable error messages
4. **Preventive Maintenance**: Detects issues before they cause failures
5. **Comprehensive Documentation**: Multiple levels of help available
6. **Automation-Friendly**: JSON output for scripting and monitoring

## Future Enhancements

Potential improvements for future iterations:

1. **Additional Checks**:
   - Python dependency verification
   - API key validity testing
   - Beads repository health
   - MCP server version compatibility

2. **Enhanced Fixes**:
   - Automatic dependency installation
   - Configuration migration/upgrade
   - Backup and restore functionality

3. **Monitoring Integration**:
   - Periodic health checks
   - Alert on critical issues
   - Health metrics dashboard

4. **Advanced Diagnostics**:
   - Performance profiling
   - Resource usage analysis
   - Network latency testing

## Verification

To verify the implementation:

```bash
# Run all doctor tests
go test -v ./internal/doctor

# Check coverage
go test -coverprofile=coverage.out ./internal/doctor
go tool cover -html=coverage.out

# Test the command
go build -o asc
./asc doctor
./asc doctor --fix
./asc doctor --json
```

## Conclusion

Task 28.10 has been successfully completed with comprehensive issue detection and remediation capabilities. The implementation includes:

- ✅ Fully functional `asc doctor` command
- ✅ Automatic detection of 15+ issue types
- ✅ Self-healing for 7 common issues
- ✅ Comprehensive test coverage (69.8%)
- ✅ Complete documentation
- ✅ Issue reporting template
- ✅ Known issues tracking

The system is production-ready and provides users with powerful tools to diagnose and fix common problems automatically.

---

**Completed**: 2025-11-10
**Test Status**: All tests passing ✅
**Coverage**: 69.8%
**Documentation**: Complete ✅
