# Known Issues and Workarounds

This document tracks known issues with dependencies, their workarounds, and resolution status.

## Active Issues

### None Currently

No active dependency issues at this time.

## Self-Healing Capabilities

The `asc doctor` command can automatically detect and fix many common issues:

### Auto-Fixable Issues

The following issues can be automatically remediated with `asc doctor --fix`:

1. **Insecure .env permissions** - Sets file permissions to 0600
2. **Corrupted PID files** - Removes invalid JSON files
3. **Orphaned PID files** - Cleans up stale process tracking
4. **Missing directories** - Creates required ~/.asc subdirectories
5. **Large log directories** - Removes logs older than 7 days
6. **Incorrect ~/.asc permissions** - Fixes directory permissions

### Manual Remediation Required

Some issues require manual intervention:

1. **Missing configuration files** - Run `asc init` to create
2. **Invalid TOML syntax** - Edit asc.toml to fix syntax errors
3. **Missing API keys** - Add keys to .env file
4. **Missing binaries** - Install required dependencies
5. **Invalid agent configuration** - Update agent settings in asc.toml

### Running Diagnostics

```bash
# Check for issues
asc doctor

# Check with detailed output
asc doctor --verbose

# Automatically fix issues
asc doctor --fix

# Output as JSON for automation
asc doctor --json
```

---

## Resolved Issues

### Example Entry (Template)

#### Issue: [Dependency Name] v[Version] - [Brief Description]

**Status**: Resolved ✅  
**Affected Versions**: asc v1.0.0 - v1.0.5  
**Severity**: Critical | High | Medium | Low  
**Reported**: YYYY-MM-DD  
**Resolved**: YYYY-MM-DD  

**Description**:
[Detailed description of the issue]

**Impact**:
- [Impact point 1]
- [Impact point 2]

**Root Cause**:
[Explanation of what caused the issue]

**Workaround**:
```bash
# Commands or configuration changes
```

**Resolution**:
[How the issue was resolved]

**Related**:
- Issue: #123
- PR: #456
- Upstream: [link]

---

## Monitoring

We actively monitor the following sources for dependency issues:

### Go Dependencies

- [Charm.sh Discussions](https://github.com/charmbracelet/bubbletea/discussions)
- [Cobra Issues](https://github.com/spf13/cobra/issues)
- [Viper Issues](https://github.com/spf13/viper/issues)
- [Gorilla WebSocket Issues](https://github.com/gorilla/websocket/issues)

### Python Dependencies

- [Anthropic SDK Issues](https://github.com/anthropics/anthropic-sdk-python/issues)
- [OpenAI SDK Issues](https://github.com/openai/openai-python/issues)
- [Google AI SDK Issues](https://github.com/google/generative-ai-python/issues)

### Security Advisories

- [GitHub Advisory Database](https://github.com/advisories)
- [Go Vulnerability Database](https://pkg.go.dev/vuln/)
- [PyPI Advisory Database](https://pypi.org/security/)
- [CVE Database](https://cve.mitre.org/)

## Reporting Issues

### Internal Issues (asc-specific)

If you encounter an issue with asc:

1. Check this document for known issues
2. Search [existing issues](https://github.com/yourusername/asc/issues)
3. Create a new issue using the bug report template
4. Include:
   - asc version (`./asc --version`)
   - Go version (`go version`)
   - Python version (`python --version`)
   - Operating system and version
   - Dependency versions (`go list -m all`, `pip list`)
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs

### Upstream Issues (dependency-related)

If you believe an issue is in a dependency:

1. Verify the issue is not in our code
2. Check the dependency's issue tracker
3. Create a minimal reproduction case
4. Report to the upstream project
5. Document the issue here with workaround
6. Link to upstream issue

## Workaround Patterns

### Pinning Dependency Versions

If a new version introduces issues:

```bash
# Go: Edit go.mod
require github.com/example/package v1.2.3

# Python: Edit agent/requirements.txt
package==1.2.3

# Then update
go mod tidy
cd agent && pip install -r requirements.txt
```

### Using Replace Directives (Go)

For temporary fixes or forks:

```go
// In go.mod
replace github.com/example/package => github.com/yourfork/package v1.2.4-fix
```

### Conditional Imports (Python)

For version-specific code:

```python
try:
    from new_module import new_function
except ImportError:
    from old_module import old_function as new_function
```

### Feature Flags

Disable problematic features:

```toml
# In asc.toml
[experimental]
websocket_enabled = false  # Fallback to polling
hot_reload_enabled = false  # Disable config watching
```

## Testing for Issues

### Before Updating Dependencies

```bash
# Run full test suite
make test-all

# Run E2E tests
make test-e2e

# Test critical workflows
./asc init
./asc check
./asc up
# Interact with TUI
./asc down

# Check for warnings
go build -v ./... 2>&1 | grep -i "deprecated\|warning"
cd agent && python -W default -m pytest
```

### After Encountering Issues

```bash
# Isolate the problematic dependency
go mod graph | grep problematic-package
pip show problematic-package

# Test with previous version
go get github.com/example/package@v1.2.3
cd agent && pip install package==1.2.3

# Verify fix
make test
```

## Communication

### User Notification

When a known issue affects users:

1. Update this document immediately
2. Add notice to README.md if critical
3. Post in GitHub Discussions
4. Include in release notes
5. Update documentation with workaround

### Internal Tracking

- Label issues with `known-issue` tag
- Link to this document in issue comments
- Update status regularly
- Close issues when resolved

## Prevention

### Pre-merge Checks

- Run dependency compatibility tests
- Review changelogs thoroughly
- Test on multiple platforms
- Check for deprecation warnings
- Verify performance benchmarks

### Monitoring

- Daily security scans
- Weekly dependency updates review
- Monthly dependency audit
- Quarterly major version planning

## Resources

- [Dependency Management Guide](DEPENDENCIES.md)
- [Breaking Changes Log](BREAKING_CHANGES.md)
- [Troubleshooting Guide](../TROUBLESHOOTING.md)
- [Contributing Guide](../CONTRIBUTING.md)

## Maintenance

This document should be updated:
- When a new issue is discovered
- When a workaround is found
- When an issue is resolved
- During release preparation
- After major dependency updates

**Last Updated**: 2025-11-10
