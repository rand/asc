# Dependency Compatibility Guide

This document details the dependency requirements, compatibility testing, and version-specific issues for the Agent Stack Controller (asc).

## Version Requirements

### Go

**Minimum Supported Version:** Go 1.21

**Recommended Version:** Go 1.22+

**Tested Versions:**
- Go 1.21.x ✓
- Go 1.22.x ✓
- Go 1.23.x ✓
- Go 1.24.x ✓
- Go 1.25.x ✓

**Note:** The current `go.mod` file specifies `go 1.25.4`. For maximum compatibility, this should be updated to `go 1.21` to match the minimum supported version stated in requirements.

### Python

**Minimum Supported Version:** Python 3.8

**Recommended Version:** Python 3.10+

**Tested Versions:**
- Python 3.8.x ✓
- Python 3.9.x ✓
- Python 3.10.x ✓
- Python 3.11.x ✓
- Python 3.12.x ✓
- Python 3.13.x ✓
- Python 3.14.x ✓

## Go Dependencies

### Core Dependencies

| Package | Version | Purpose | Notes |
|---------|---------|---------|-------|
| github.com/charmbracelet/bubbles | v0.21.0 | TUI components | Stable |
| github.com/charmbracelet/bubbletea | v1.3.10 | TUI framework | Stable |
| github.com/charmbracelet/lipgloss | v1.1.0 | TUI styling | Stable |
| github.com/fsnotify/fsnotify | v1.9.0 | File watching | Stable |
| github.com/google/uuid | v1.6.0 | UUID generation | Stable |
| github.com/gorilla/websocket | v1.5.3 | WebSocket client | Stable |
| github.com/spf13/cobra | v1.10.1 | CLI framework | Stable |
| github.com/spf13/viper | v1.21.0 | Configuration | Stable |

### Dependency Update Policy

- **Security updates:** Applied immediately
- **Minor updates:** Reviewed and applied quarterly
- **Major updates:** Reviewed carefully for breaking changes

### Known Compatibility Issues

#### Go 1.20 and Earlier
- **Issue:** Some TUI rendering features may not work correctly
- **Reason:** Uses features introduced in Go 1.21
- **Solution:** Upgrade to Go 1.21 or later

#### Go 1.26+ (Future)
- **Status:** Not yet tested
- **Action:** Will be tested when released

## Python Dependencies

### Core Dependencies

| Package | Minimum Version | Purpose | Python Version |
|---------|----------------|---------|----------------|
| anthropic | ≥0.34.0 | Claude API client | 3.8+ |
| google-generativeai | ≥0.3.0 | Gemini API client | 3.8+ |
| openai | ≥1.0.0 | OpenAI API client | 3.8+ |
| requests | ≥2.31.0 | HTTP client | 3.8+ |
| python-dotenv | ≥1.0.0 | Environment variables | 3.8+ |

### Python Version-Specific Notes

#### Python 3.8
- **Status:** Minimum supported version
- **Notes:** All features work correctly
- **EOL:** October 2024 (consider upgrading)

#### Python 3.9-3.11
- **Status:** Fully supported
- **Notes:** Recommended for production use

#### Python 3.12+
- **Status:** Fully supported
- **Notes:** Latest features and performance improvements
- **Testing:** Verified with Python 3.12, 3.13, 3.14

### Known Python Compatibility Issues

#### anthropic Package
- **Version:** 0.34.0+
- **Issue:** Earlier versions (<0.34.0) have different API
- **Solution:** Use version 0.34.0 or later

#### openai Package
- **Version:** 1.0.0+
- **Issue:** Version 1.0.0 introduced breaking changes from 0.x
- **Solution:** Use version 1.0.0 or later (not 0.x)

## External Dependencies

### Required

#### Git
- **Minimum Version:** 2.20+
- **Purpose:** Version control and beads repository management
- **Installation:** 
  - macOS: `brew install git`
  - Linux: `apt-get install git` or `yum install git`

### Optional

#### Docker
- **Minimum Version:** 20.10+
- **Purpose:** Container runtime (optional, for containerized agents)
- **Installation:**
  - macOS: `brew install docker`
  - Linux: Follow [Docker installation guide](https://docs.docker.com/engine/install/)

#### bd (Beads CLI)
- **Purpose:** Task management (required for full functionality)
- **Installation:** Follow [beads installation guide](https://github.com/steveyegge/beads)

#### uv (Python Package Manager)
- **Purpose:** Fast Python package management (optional)
- **Installation:** `curl -LsSf https://astral.sh/uv/install.sh | sh`

## Platform Compatibility

### Supported Platforms

| Platform | Architecture | Status | Notes |
|----------|-------------|--------|-------|
| Linux | amd64 | ✓ Supported | Primary development platform |
| macOS | amd64 | ✓ Supported | Intel Macs |
| macOS | arm64 | ✓ Supported | Apple Silicon (M1/M2/M3) |
| Windows | amd64 | ⚠ Experimental | Limited testing |

### Cross-Compilation

The project supports cross-compilation for all supported platforms:

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -o asc-linux-amd64

# macOS amd64
GOOS=darwin GOARCH=amd64 go build -o asc-darwin-amd64

# macOS arm64
GOOS=darwin GOARCH=arm64 go build -o asc-darwin-arm64
```

## Testing Dependency Compatibility

### Automated Tests

Run the dependency compatibility test suite:

```bash
# Run all dependency tests
go test -v ./test -run TestDependency

# Run specific tests
go test -v ./test -run TestGoVersionCompatibility
go test -v ./test -run TestPythonVersionCompatibility
go test -v ./test -run TestExternalDependencies
```

### Manual Testing Script

Run the comprehensive compatibility check script:

```bash
./scripts/test-dependency-compatibility.sh
```

This script checks:
1. Go version compatibility
2. Python version compatibility
3. Go build success
4. Go module integrity
5. External dependencies
6. Python dependencies
7. Cross-compilation
8. Deprecated dependencies
9. Available updates

## Dependency Update Scenarios

### Updating Go Dependencies

```bash
# Update all dependencies to latest minor/patch versions
go get -u ./...

# Update specific dependency
go get -u github.com/spf13/cobra@latest

# Tidy up go.mod and go.sum
go mod tidy

# Verify integrity
go mod verify
```

### Updating Python Dependencies

```bash
# Update all Python dependencies
cd agent
pip install --upgrade -r requirements.txt

# Update specific dependency
pip install --upgrade anthropic

# Generate new requirements with versions
pip freeze > requirements.txt
```

### Testing After Updates

After updating dependencies:

1. Run full test suite: `make test`
2. Run dependency compatibility tests: `go test -v ./test -run TestDependency`
3. Run integration tests: `go test -v ./test -run TestIntegration`
4. Test build: `go build ./...`
5. Test cross-compilation: `./scripts/test-dependency-compatibility.sh`

## Deprecated Dependencies

### Currently None

No deprecated dependencies are currently in use.

### Monitoring for Deprecations

We monitor for deprecated dependencies through:
- Automated dependency scanning (Dependabot)
- Regular dependency audits
- Go module updates
- Security advisories

### Replacement Strategy

If a dependency becomes deprecated:
1. Identify replacement package
2. Create migration plan
3. Update code to use new package
4. Test thoroughly
5. Update documentation
6. Communicate in release notes

## Version-Specific Issues

### Go 1.21 Specific

**Issue:** None known

**Workaround:** N/A

### Go 1.22+ Specific

**Issue:** None known

**Workaround:** N/A

### Python 3.8 Specific

**Issue:** Some type hints may not work (introduced in 3.9+)

**Workaround:** Use `from __future__ import annotations` if needed

### Python 3.12+ Specific

**Issue:** None known

**Workaround:** N/A

## Continuous Integration

### GitHub Actions

Our CI pipeline tests against multiple versions:

```yaml
strategy:
  matrix:
    go-version: ['1.21', '1.22', '1.23']
    python-version: ['3.8', '3.9', '3.10', '3.11', '3.12']
    os: [ubuntu-latest, macos-latest]
```

### Dependency Monitoring

- **Dependabot:** Automated dependency updates
- **Security scanning:** Weekly security audits
- **License checking:** Automated license compliance

## Troubleshooting

### Go Version Issues

**Problem:** Build fails with "go version too old"

**Solution:**
```bash
# Check current version
go version

# Upgrade Go (macOS)
brew upgrade go

# Upgrade Go (Linux)
# Download from https://go.dev/dl/
```

### Python Version Issues

**Problem:** Import errors or syntax errors

**Solution:**
```bash
# Check current version
python3 --version

# Upgrade Python (macOS)
brew upgrade python3

# Upgrade Python (Linux)
sudo apt-get update
sudo apt-get install python3.11
```

### Dependency Installation Issues

**Problem:** `go get` or `pip install` fails

**Solution:**
```bash
# Clear Go cache
go clean -modcache

# Clear pip cache
pip cache purge

# Retry installation
go mod download
pip install -r requirements.txt
```

## Best Practices

1. **Pin major versions** in go.mod and requirements.txt
2. **Test before updating** dependencies in production
3. **Review changelogs** for breaking changes
4. **Use semantic versioning** for your own releases
5. **Monitor security advisories** for all dependencies
6. **Keep dependencies up to date** but not bleeding edge
7. **Document version-specific issues** as they arise
8. **Test on multiple platforms** before releasing

## Resources

- [Go Release Policy](https://go.dev/doc/devel/release)
- [Python Release Schedule](https://www.python.org/downloads/)
- [Semantic Versioning](https://semver.org/)
- [Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)

## Maintenance Schedule

- **Weekly:** Security updates
- **Monthly:** Dependency review
- **Quarterly:** Minor version updates
- **Annually:** Major version updates (with testing)

## Contact

For dependency-related issues or questions:
- Open an issue on GitHub
- Check existing issues for similar problems
- Review this documentation first

---

Last Updated: 2025-11-10
