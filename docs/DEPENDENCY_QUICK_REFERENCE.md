# Dependency Compatibility Quick Reference

## Minimum Requirements

| Component | Minimum Version | Recommended | Tested With |
|-----------|----------------|-------------|-------------|
| **Go** | 1.21 | 1.22+ | 1.25.4 ✅ |
| **Python** | 3.8 | 3.10+ | 3.14.0 ✅ |
| **git** | 2.20+ | Latest | 2.51.2 ✅ |
| **docker** | 20.10+ | Latest | Optional |
| **bd (beads)** | Latest | Latest | 0.22.1 ✅ |

## Quick Test Commands

```bash
# Test all dependencies
./scripts/test-dependency-compatibility.sh

# Test Go compatibility
go test -v ./test -run TestGoVersionCompatibility

# Test Python compatibility
go test -v ./test -run TestPythonVersionCompatibility

# Test external dependencies
go test -v ./test -run TestExternalDependencies

# Test cross-compilation
go test -v ./test -run TestCrossCompilation

# Run all dependency tests
go test -v ./test -run TestDependency
```

## Python Dependencies

```
anthropic>=0.34.0          # Claude API
google-generativeai>=0.3.0 # Gemini API
openai>=1.0.0              # OpenAI API
requests>=2.31.0           # HTTP client
python-dotenv>=1.0.0       # Environment variables
```

## Installation

### Go
```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
```

### Python
```bash
# macOS
brew install python3

# Linux
sudo apt-get install python3 python3-pip
```

### Python Dependencies
```bash
cd agent
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Status Indicators

✅ **Compatible** - Tested and working  
⚠️ **Warning** - Works but has minor issues  
❌ **Incompatible** - Does not work  
ℹ️ **Info** - Additional information available

## Common Issues

### Issue: "go version too old"
**Solution:** Upgrade Go to 1.21+
```bash
brew upgrade go  # macOS
```

### Issue: "python3 not found"
**Solution:** Install Python 3.8+
```bash
brew install python3  # macOS
```

### Issue: "externally-managed-environment" (Python)
**Solution:** Use virtual environment
```bash
python3 -m venv .venv
source .venv/bin/activate
```

### Issue: "docker not found"
**Solution:** Docker is optional, not required for core functionality

## Documentation

- **Full Guide:** `docs/DEPENDENCY_COMPATIBILITY.md`
- **Test Report:** `DEPENDENCY_COMPATIBILITY_REPORT.md`
- **Completion Summary:** `TASK_29.6_COMPLETION.md`

## CI/CD Integration

```yaml
# GitHub Actions example
- name: Check Dependencies
  run: ./scripts/test-dependency-compatibility.sh

- name: Run Dependency Tests
  run: go test -v ./test -run TestDependency
```

## Update Commands

```bash
# Update Go dependencies
go get -u ./...
go mod tidy

# Update Python dependencies
pip install --upgrade -r agent/requirements.txt

# Check for updates
go list -u -m all
```

## Support Matrix

| Go Version | Status | Notes |
|------------|--------|-------|
| 1.20 | ❌ | Too old |
| 1.21 | ✅ | Minimum |
| 1.22 | ✅ | Recommended |
| 1.23+ | ✅ | Latest |

| Python Version | Status | Notes |
|----------------|--------|-------|
| 3.7 | ❌ | Too old |
| 3.8 | ✅ | Minimum |
| 3.9-3.11 | ✅ | Recommended |
| 3.12+ | ✅ | Latest |

## Platform Support

| Platform | Architecture | Status |
|----------|-------------|--------|
| Linux | amd64 | ✅ |
| macOS | amd64 | ✅ |
| macOS | arm64 | ✅ |
| Windows | amd64 | ⚠️ Experimental |

---

**Last Updated:** November 10, 2025  
**Task:** 29.6 Test dependency compatibility  
**Status:** ✅ COMPLETED
