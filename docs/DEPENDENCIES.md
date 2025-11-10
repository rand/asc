# Dependency Management

This document provides a comprehensive overview of all dependencies used in the Agent Stack Controller (asc) project, their purposes, version constraints, and update policies.

## Table of Contents

- [Go Dependencies](#go-dependencies)
- [Python Dependencies](#python-dependencies)
- [GitHub Actions Dependencies](#github-actions-dependencies)
- [Dependency Update Policy](#dependency-update-policy)
- [Security Monitoring](#security-monitoring)
- [Breaking Changes](#breaking-changes)
- [Testing Compatibility](#testing-compatibility)

## Go Dependencies

### Direct Dependencies

#### UI Framework (Charm.sh Ecosystem)

| Package | Version | Purpose | Update Policy |
|---------|---------|---------|---------------|
| `github.com/charmbracelet/bubbletea` | v1.3.10+ | TUI framework for interactive terminal applications | Minor/patch updates weekly |
| `github.com/charmbracelet/lipgloss` | v1.1.0+ | Terminal styling and layout library | Minor/patch updates weekly |
| `github.com/charmbracelet/bubbles` | v0.21.0+ | Reusable TUI components (textinput, viewport, etc.) | Minor/patch updates weekly |

**Rationale**: The Charm.sh ecosystem provides the foundation for our TUI. These libraries are actively maintained and follow semantic versioning. We stay current with minor/patch updates to benefit from bug fixes and new features.

**Critical**: Yes - Core functionality depends on these libraries.

#### CLI Framework

| Package | Version | Purpose | Update Policy |
|---------|---------|---------|---------------|
| `github.com/spf13/cobra` | v1.10.1+ | Command-line interface framework | Minor/patch updates weekly |
| `github.com/spf13/viper` | v1.21.0+ | Configuration management (TOML parsing) | Minor/patch updates weekly |

**Rationale**: Cobra and Viper are industry-standard libraries for Go CLI applications. They are stable and well-maintained.

**Critical**: Yes - Core CLI functionality depends on these libraries.

### Indirect Dependencies

#### Terminal and System

- `github.com/erikgeiser/coninput` - Windows console input handling
- `github.com/mattn/go-isatty` - TTY detection
- `github.com/mattn/go-localereader` - Locale-aware input reading
- `github.com/muesli/cancelreader` - Cancellable reader for terminal input
- `github.com/xo/terminfo` - Terminal capability database

**Purpose**: Cross-platform terminal compatibility and input handling.

#### WebSocket Communication

- `github.com/gorilla/websocket` v1.5.3+ - WebSocket client for real-time MCP updates

**Purpose**: Real-time communication with mcp_agent_mail server.

**Critical**: Yes - Required for real-time agent status updates.

#### File System Watching

- `github.com/fsnotify/fsnotify` v1.9.0+ - File system event notifications

**Purpose**: Configuration hot-reload functionality.

**Critical**: No - Fallback to manual reload if unavailable.

#### Utilities

- `github.com/google/uuid` - UUID generation for correlation IDs
- `github.com/lucasb-eyer/go-colorful` - Color manipulation
- `github.com/rivo/uniseg` - Unicode text segmentation
- `github.com/atotto/clipboard` - Clipboard operations

**Purpose**: Supporting utilities for various features.

### Version Pinning Strategy

**Critical Dependencies** (pinned to minor version):
- `bubbletea` >= v1.3.10, < v2.0.0
- `lipgloss` >= v1.1.0, < v2.0.0
- `cobra` >= v1.10.1, < v2.0.0
- `viper` >= v1.21.0, < v2.0.0
- `gorilla/websocket` >= v1.5.3, < v2.0.0

**Non-Critical Dependencies**: Allow patch updates automatically.

## Python Dependencies

### Agent Runtime Dependencies

| Package | Version | Purpose | Update Policy |
|---------|---------|---------|---------------|
| `anthropic` | >=0.34.0 | Claude API client | Minor/patch updates weekly |
| `google-generativeai` | >=0.3.0 | Gemini API client | Minor/patch updates weekly |
| `openai` | >=1.0.0 | OpenAI API client (GPT-4, Codex) | Minor/patch updates weekly |
| `requests` | >=2.31.0 | HTTP client for MCP communication | Patch updates only |
| `python-dotenv` | >=1.0.0 | Environment variable loading | Patch updates only |

### LLM Provider SDKs

**Anthropic (Claude)**
- **Purpose**: Interface with Claude models (claude-3-opus, claude-3-sonnet, etc.)
- **Update Policy**: Follow official SDK releases, test thoroughly before updating
- **Breaking Changes**: Monitor Anthropic changelog for API changes
- **Critical**: Yes - Required for Claude-based agents

**Google Generative AI (Gemini)**
- **Purpose**: Interface with Gemini models (gemini-pro, gemini-ultra, etc.)
- **Update Policy**: Follow official SDK releases, test thoroughly before updating
- **Breaking Changes**: Monitor Google AI changelog for API changes
- **Critical**: Yes - Required for Gemini-based agents

**OpenAI**
- **Purpose**: Interface with OpenAI models (GPT-4, GPT-4-turbo, etc.)
- **Update Policy**: Follow official SDK releases, test thoroughly before updating
- **Breaking Changes**: Monitor OpenAI changelog for API changes
- **Critical**: Yes - Required for OpenAI-based agents

### Version Pinning Strategy

**LLM SDKs** (minimum version specified):
- Allow minor and patch updates to benefit from bug fixes and new model support
- Test compatibility before merging Dependabot PRs
- Pin to specific versions if breaking changes are introduced

**Utilities** (minimum version specified):
- `requests` >= 2.31.0 (security fixes)
- `python-dotenv` >= 1.0.0 (stable API)

## GitHub Actions Dependencies

### Workflow Actions

| Action | Version | Purpose | Update Policy |
|--------|---------|---------|---------------|
| `actions/checkout` | v4 | Repository checkout | Major version pinned |
| `actions/setup-go` | v5 | Go environment setup | Major version pinned |
| `actions/setup-python` | v5 | Python environment setup | Major version pinned |
| `actions/cache` | v4 | Dependency caching | Major version pinned |
| `codecov/codecov-action` | v5 | Code coverage reporting | Major version pinned |
| `github/codeql-action` | v3 | Security scanning | Major version pinned |
| `golangci/golangci-lint-action` | v6 | Go linting | Major version pinned |

**Update Policy**: Dependabot monitors and creates PRs for new major versions. Review release notes before merging.

## Dependency Update Policy

### Automated Updates (Dependabot)

**Schedule**: Weekly on Mondays at 09:00 UTC

**Configuration**: `.github/dependabot.yml`

**Process**:
1. Dependabot creates PRs for dependency updates
2. CI/CD pipeline runs full test suite
3. Dependency compatibility tests run automatically
4. Manual review required before merging
5. Group minor/patch updates to reduce PR volume

### Manual Review Checklist

Before merging Dependabot PRs:

- [ ] Review changelog/release notes for breaking changes
- [ ] Verify all CI checks pass
- [ ] Run dependency compatibility tests locally
- [ ] Test critical user workflows (init, up, down)
- [ ] Check for deprecation warnings in logs
- [ ] Update documentation if API changes
- [ ] Test on multiple platforms (Linux, macOS)

### Major Version Updates

**Process**:
1. Create feature branch for major update
2. Update dependency version
3. Fix breaking changes
4. Update tests
5. Update documentation
6. Run full test suite including E2E tests
7. Test on all supported platforms
8. Create PR with detailed changelog
9. Request team review
10. Merge after approval

## Security Monitoring

### Vulnerability Scanning

**Tools**:
- GitHub Dependabot Security Alerts (enabled)
- `go list -m -json all | nancy sleuth` (Go dependencies)
- `pip-audit` (Python dependencies)
- CodeQL security scanning (GitHub Actions)

**Process**:
1. Security alerts trigger immediate notification
2. Assess severity and impact
3. Create hotfix branch if critical
4. Update vulnerable dependency
5. Run security tests
6. Deploy patch release

### Security Advisory Sources

**Go**:
- GitHub Advisory Database
- Go Vulnerability Database (govulncheck)
- NIST National Vulnerability Database

**Python**:
- PyPI Advisory Database
- GitHub Advisory Database
- Safety DB

**Monitoring Frequency**: Continuous (GitHub alerts) + Weekly manual review

### Response Time SLAs

| Severity | Response Time | Patch Release |
|----------|---------------|---------------|
| Critical | 24 hours | 48 hours |
| High | 3 days | 1 week |
| Medium | 1 week | 2 weeks |
| Low | 2 weeks | Next release |

## Breaking Changes

### Tracking Breaking Changes

**Documentation Location**: `docs/BREAKING_CHANGES.md`

**Format**:
```markdown
## [Dependency Name] v[Old Version] → v[New Version]

**Date**: YYYY-MM-DD
**Severity**: Critical | High | Medium | Low
**Impact**: [Description of what breaks]

### Changes Required

1. [Change 1]
2. [Change 2]

### Migration Guide

[Step-by-step instructions]

### Rollback Plan

[How to revert if needed]
```

### Recent Breaking Changes

See `docs/BREAKING_CHANGES.md` for historical record.

### Deprecation Policy

**Notice Period**: 2 releases (minimum 4 weeks)

**Process**:
1. Add deprecation warning in code
2. Document in CHANGELOG.md
3. Update documentation with migration path
4. Remove in next major version

## Testing Compatibility

### Dependency Compatibility Test Workflow

**File**: `.github/workflows/dependency-compatibility.yml`

**Triggers**:
- Dependabot PRs
- Manual workflow dispatch
- Weekly scheduled run

**Tests**:
1. Unit tests with new dependencies
2. Integration tests
3. E2E tests
4. Performance benchmarks
5. Cross-platform tests (Linux, macOS)
6. Multiple Go versions (1.21, 1.22, 1.23)
7. Multiple Python versions (3.10, 3.11, 3.12)

### Compatibility Matrix

**Go Versions**:
- Minimum: 1.21
- Recommended: 1.23+
- Tested: 1.21, 1.22, 1.23

**Python Versions**:
- Minimum: 3.10
- Recommended: 3.11+
- Tested: 3.10, 3.11, 3.12

**Operating Systems**:
- Linux (Ubuntu 22.04, 24.04)
- macOS (13, 14, 15)
- Windows (experimental support)

### Regression Testing

**Scope**:
- All unit tests must pass
- All integration tests must pass
- All E2E tests must pass
- Performance benchmarks within 10% of baseline
- No new linting errors
- No new security vulnerabilities

**Failure Handling**:
1. Investigate root cause
2. Determine if issue is in our code or dependency
3. File issue with dependency maintainer if needed
4. Apply workaround or pin to previous version
5. Document in `docs/KNOWN_ISSUES.md`

## Dependency Upgrade Workflow

### Step-by-Step Process

1. **Receive Dependabot PR**
   - Review PR description and changelog
   - Check for breaking changes

2. **Automated Testing**
   - CI runs full test suite
   - Dependency compatibility workflow runs
   - Security scans execute

3. **Manual Testing**
   ```bash
   # Checkout PR branch
   git fetch origin pull/[PR_NUMBER]/head:dependabot-update
   git checkout dependabot-update
   
   # Run tests locally
   make test
   make test-integration
   make test-e2e
   
   # Test critical workflows
   ./asc init
   ./asc check
   ./asc up
   # Verify TUI functionality
   ./asc down
   ```

4. **Review Checklist**
   - [ ] All tests pass
   - [ ] No breaking changes
   - [ ] Documentation updated if needed
   - [ ] Performance acceptable
   - [ ] Security scan clean

5. **Merge**
   - Approve PR
   - Merge with squash commit
   - Monitor for issues

6. **Post-Merge**
   - Verify main branch CI passes
   - Monitor error tracking for new issues
   - Update CHANGELOG.md if significant

## Troubleshooting

### Common Issues

**Issue**: Dependabot PR fails CI
- **Solution**: Check test logs, may need code changes for compatibility

**Issue**: Dependency conflict
- **Solution**: Run `go mod tidy` or update conflicting dependencies together

**Issue**: Security vulnerability in transitive dependency
- **Solution**: Update direct dependency that pulls in vulnerable version

**Issue**: Breaking change in minor version
- **Solution**: Pin to previous version, file issue with maintainer

### Getting Help

- Check dependency documentation and changelog
- Search GitHub issues for similar problems
- Ask in project discussions
- Contact dependency maintainers

## References

- [Go Modules Documentation](https://go.dev/doc/modules/managing-dependencies)
- [Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)
- [Semantic Versioning](https://semver.org/)
- [Go Vulnerability Database](https://pkg.go.dev/vuln/)
- [Python Package Index](https://pypi.org/)
