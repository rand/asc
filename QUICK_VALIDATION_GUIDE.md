# Quick Validation Guide

Fast reference for running Phase 29 validation.

## TL;DR

```bash
# Run automated validation
./scripts/run-validation.sh

# Check results
cat validation-reports/*/SUMMARY.md
```

## What Gets Validated

✅ **Build** - All platforms compile  
✅ **Tests** - All tests pass with good coverage  
✅ **Code Quality** - Linting and static analysis  
✅ **Documentation** - Complete and accurate  
✅ **Dependencies** - Compatible versions  
✅ **Integration** - End-to-end workflows  
✅ **Security** - No vulnerabilities or leaks  
✅ **Performance** - Meets targets  

## Quick Commands

### Full Validation
```bash
./scripts/run-validation.sh
```

### Individual Checks

**Build**:
```bash
make clean && make build-all
```

**Tests**:
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**Linting**:
```bash
go vet ./...
golangci-lint run ./...
gosec ./...
```

**Security**:
```bash
grep -r "api[_-]key.*=.*['\"]sk-" . --exclude-dir=vendor
```

**Performance**:
```bash
go test -bench=. -benchmem ./...
```

## Interpreting Results

### Exit Codes
- `0` - All checks passed ✅
- `1` - Critical failures (must fix) ❌
- `2` - High-priority issues (should fix) ⚠️

### Report Location
```
validation-reports/YYYYMMDD-HHMMSS/
├── SUMMARY.md              # Start here
├── build.log
├── unit-tests.log
├── coverage.out
├── coverage-summary.txt
├── coverage-gaps.txt
├── low-coverage.txt
├── vet.log
├── golangci-lint.log
├── gosec.json
├── secret-scan.txt
└── benchmarks.txt
```

## Common Issues

### Build Fails
```bash
# Clean and retry
make clean
go mod tidy
make build
```

### Tests Fail
```bash
# Run specific test
go test -v ./internal/config -run TestConfigLoad

# Check test dependencies
go mod verify
```

### Coverage Too Low
```bash
# See what's not covered
go tool cover -html=coverage.out
```

### Linting Issues
```bash
# Auto-fix formatting
gofmt -w .

# See specific issues
golangci-lint run --verbose
```

## Decision Matrix

| Critical | High | Medium | Low | Decision |
|----------|------|--------|-----|----------|
| 0 | 0 | Any | Any | ✅ GO |
| 0 | 1-3 | Any | Any | ⚠️ CONDITIONAL GO |
| 1+ | Any | Any | Any | ❌ NO-GO |

## Next Steps After Validation

### If GO ✅
1. Review summary report
2. Document any accepted warnings
3. Proceed with release preparation

### If CONDITIONAL GO ⚠️
1. Review high-priority issues
2. Create tasks to fix them
3. Fix issues
4. Re-run validation
5. Proceed if clear

### If NO-GO ❌
1. Review critical failures
2. Create urgent fix tasks
3. Fix critical issues
4. Re-run validation
5. Do not proceed until clear

## Tips

- Run validation on a clean checkout
- Ensure all dependencies are installed
- Run on the same platform as production
- Save reports for documentation
- Re-run after any fixes

## Help

**Validation script not working?**
```bash
chmod +x scripts/run-validation.sh
```

**Missing tools?**
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

**Need more details?**
- See: `PHASE_29_VALIDATION_PLAN.md`
- See: `.kiro/specs/agent-stack-controller/tasks.md` (Phase 29)

---

**Remember**: Validation is discovery, not judgment. The goal is to know what needs work, not to achieve perfection immediately.
