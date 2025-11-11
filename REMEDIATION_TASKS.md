# Remediation Task Checklist

**Generated:** November 10, 2025  
**Source:** Phase 29 Gap Analysis Report  
**Plan:** PHASE_29_REMEDIATION_PLAN.md

## Quick Reference

Use this checklist to track remediation progress. Check off tasks as they are completed.

---

## Phase 1: Critical Issues (Week 1) - MUST FIX BEFORE ANY RELEASE

### Day 1 - Morning
- [ ] **C-1:** Fix compilation errors in `internal/beads/error_handling_test.go` (1h)
  - [ ] Update `NewClient` call to include `time.Duration` parameter (line 41)
  - [ ] Fix type mismatch for string constant (line 234)
  - [ ] Remove duplicate `contains` function (line 582)
  - [ ] Verify tests compile and run

- [ ] **C-2:** Fix compilation errors in `internal/mcp/error_handling_test.go` (1h)
  - [ ] Locate correct `NewClient` function or import
  - [ ] Update all function calls to use correct signature
  - [ ] Verify tests compile and run

- [ ] **C-3:** Fix compilation errors in `internal/process/error_handling_test.go` (1h)
  - [ ] Fix variable declaration on line 289 (use `=` instead of `:=`)
  - [ ] Fix PID type mismatches (lines 300, 310)
  - [ ] Verify tests compile and run

### Day 1 - Afternoon
- [ ] **H-1:** Fix 5 test assertion failures in `internal/check` (2h)
  - [ ] Update expected error messages to match actual implementation
  - [ ] Update expected status levels (warn vs fail)
  - [ ] Run tests to verify fixes
  - [ ] Document any intentional behavior changes

- [ ] **H-2:** Fix 5 test assertion failures in `internal/config` (2h)
  - [ ] Update expected error messages to match actual implementation
  - [ ] Update validation order expectations
  - [ ] Run tests to verify fixes
  - [ ] Document any intentional behavior changes

### Day 1 - End of Day
- [ ] **H-3:** Format all code with gofmt (5min)
  - [ ] Run `gofmt -w .`
  - [ ] Verify `gofmt -l .` returns no files
  - [ ] Commit formatted code

### Day 2
- [ ] **H-4:** Install and run golangci-lint (2h)
  - [ ] Install golangci-lint
  - [ ] Run `golangci-lint run ./...`
  - [ ] Address all high-severity issues
  - [ ] Document accepted warnings

- [ ] **H-5:** Install and run gosec (1h)
  - [ ] Install gosec
  - [ ] Run `gosec ./...`
  - [ ] Address all critical security issues
  - [ ] Document accepted warnings

### Phase 1 Gate
- [ ] All code compiles without errors
- [ ] All tests pass (no failures)
- [ ] Code properly formatted
- [ ] No high-severity linting issues
- [ ] No critical security issues

**Estimated Completion:** End of Week 1  
**Go/No-Go Decision:** Proceed to Phase 2?

---

## Phase 2: High Priority Coverage (Weeks 2-4) - SHOULD FIX BEFORE PRODUCTION

### Week 2: TUI Testing Part 1
- [ ] **H-6.1:** Set up TUI integration test framework (2 days)
  - [ ] Research bubbletea testing approaches
  - [ ] Create mock terminal for testing
  - [ ] Set up test fixtures and helpers

- [ ] **H-6.2:** Add wizard flow tests (3 days)
  - [ ] Test welcome screen rendering
  - [ ] Test dependency check display
  - [ ] Test API key input and validation
  - [ ] Test config generation
  - [ ] Test validation step
  - [ ] Test complete screen
  - [ ] Target: 60%+ coverage for wizard.go

### Week 3: TUI Testing Part 2 + CLI Testing Part 1
- [ ] **H-6.3:** Add TUI rendering tests (2 days)
  - [ ] Test agent pane rendering
  - [ ] Test task pane rendering
  - [ ] Test log pane rendering
  - [ ] Test footer rendering
  - [ ] Test layout calculations
  - [ ] Target: 40%+ coverage for view.go, agents.go, tasks.go, logs.go

- [ ] **H-6.4:** Add TUI interaction tests (1 day)
  - [ ] Test keyboard event handling
  - [ ] Test modal interactions
  - [ ] Test navigation and search
  - [ ] Test state transitions
  - [ ] Target: 40%+ coverage for update.go, modals.go

- [ ] **H-6.5:** Add theme and styling tests (1 day)
  - [ ] Test theme application
  - [ ] Test color calculations
  - [ ] Test animation rendering
  - [ ] Target: 30%+ coverage for theme.go, animations.go

- [ ] **H-7.1:** Set up CLI integration test framework (1 day)
  - [ ] Create test environment setup/teardown
  - [ ] Mock file system operations
  - [ ] Mock process execution
  - [ ] Set up test fixtures

### Week 4: CLI Testing Part 2 + Coverage Improvements
- [ ] **H-7.2:** Add CLI command tests (5 days)
  - [ ] Test `asc init` command (1 day)
  - [ ] Test `asc up` command (1 day)
  - [ ] Test `asc down` command (0.5 day)
  - [ ] Test `asc check` command (0.5 day)
  - [ ] Test `asc test` command (0.5 day)
  - [ ] Test `asc services` command (0.5 day)
  - [ ] Test `asc secrets` command (0.5 day)
  - [ ] Test `asc doctor` command (0.5 day)
  - [ ] Test `asc cleanup` command (0.5 day)

- [ ] **H-8:** Fix secrets tests (1 day)
  - [ ] Install age binary
  - [ ] Remove skip conditions from tests
  - [ ] Add tests for missing coverage
  - [ ] Target: 80%+ coverage

- [ ] **H-9:** Improve doctor coverage (3 days)
  - [ ] Add tests for `checkAgents` function (1 day)
  - [ ] Add tests for report generation (1 day)
  - [ ] Improve coverage for other functions (1 day)
  - [ ] Target: 80%+ coverage

- [ ] **H-10:** Improve logger coverage (2 days)
  - [ ] Add log rotation tests (1 day)
  - [ ] Add concurrent logging tests (0.5 day)
  - [ ] Add structured logging tests (0.5 day)
  - [ ] Target: 80%+ coverage

### Phase 2 Gate
- [ ] TUI coverage > 40%
- [ ] CLI coverage > 60%
- [ ] Secrets coverage > 80%
- [ ] Doctor coverage > 80%
- [ ] Logger coverage > 80%
- [ ] Overall coverage > 60%

**Estimated Completion:** End of Week 4  
**Go/No-Go Decision:** Proceed to Phase 3 or release?

---

## Phase 3: Medium Priority Issues (Week 5) - NICE TO HAVE

### Week 5: Documentation and Tooling
- [ ] **M-1:** Improve config coverage (2 days)
  - [ ] Add tests for default path functions
  - [ ] Improve watcher `Start` function coverage
  - [ ] Test agent lifecycle functions
  - [ ] Test template functions
  - [ ] Target: 80%+ coverage

- [ ] **M-2, M-3, M-6:** Add CHANGELOG and versioning (3 hours)
  - [ ] Create CHANGELOG.md
  - [ ] Create VERSIONING.md
  - [ ] Update go.mod from 1.25.4 to 1.21
  - [ ] Run `go mod tidy`

- [ ] **M-4:** Add link validation (4 hours)
  - [ ] Install markdown-link-check
  - [ ] Create link validation script
  - [ ] Add to CI/CD pipeline
  - [ ] Fix any broken links

- [ ] **M-5:** Add example testing (4 hours)
  - [ ] Extract code examples from documentation
  - [ ] Create test script to compile/run examples
  - [ ] Add to CI/CD pipeline
  - [ ] Fix any broken examples

- [ ] **M-7, M-8:** Install Python linters (2 hours)
  - [ ] Install pylint
  - [ ] Run pylint on agent code
  - [ ] Address high-severity issues
  - [ ] Install flake8
  - [ ] Run flake8 on agent code
  - [ ] Add to CI/CD pipeline

### Phase 3 Gate
- [ ] Config coverage > 80%
- [ ] CHANGELOG.md created
- [ ] VERSIONING.md created
- [ ] go.mod updated to 1.21
- [ ] Link validation automated
- [ ] Example testing automated
- [ ] Python code linted

**Estimated Completion:** End of Week 5  
**Go/No-Go Decision:** Proceed to Phase 4 or release?

---

## Phase 4: Low Priority Issues (Week 6+) - OPTIONAL

### Week 6: Polish and Optional Features
- [ ] **L-1:** Add screenshots to README (2 hours)
  - [ ] Capture screenshots of TUI
  - [ ] Add to README
  - [ ] Optimize image sizes

- [ ] **L-2:** Fix dev environment security (1 hour)
  - [ ] Fix .env permissions (chmod 600)
  - [ ] Remove .env from git tracking
  - [ ] Fix log directory permissions
  - [ ] Fix PID directory permissions

- [ ] **L-3:** Install Docker (optional) (30 minutes)
  - [ ] Install Docker Desktop
  - [ ] Verify installation
  - [ ] Update documentation

- [ ] **L-4:** Update dependencies (4 hours)
  - [ ] Review available updates
  - [ ] Test updates in staging
  - [ ] Apply safe updates
  - [ ] Run full test suite

### Phase 4 Gate
- [ ] README has screenshots
- [ ] Dev environment secure
- [ ] Docker installed (optional)
- [ ] Dependencies updated

**Estimated Completion:** End of Week 6  
**Decision:** Release to production

---

## Summary Progress

### Overall Status
- [ ] Phase 1: Critical Issues (Week 1)
- [ ] Phase 2: High Priority Coverage (Weeks 2-4)
- [ ] Phase 3: Medium Priority Issues (Week 5)
- [ ] Phase 4: Low Priority Issues (Week 6+)

### Coverage Metrics
- Current: 23.7%
- After Phase 1: ~25%
- After Phase 2: ~60%
- After Phase 3: ~65%
- After Phase 4: ~70%

### Release Readiness
- [ ] **Minimum (Beta):** Phase 1 complete (2-3 days)
- [ ] **Recommended (Production):** Phase 1 + Phase 2 complete (4 weeks)
- [ ] **Full (Gold):** All phases complete (5-6 weeks)

---

## Quick Commands

### Verification Commands
```bash
# Compile all code
go build ./...

# Run all tests
go test ./... -v

# Check formatting
gofmt -l .

# Run linters
golangci-lint run ./...
gosec ./...

# Check coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# Python linting
pylint agent/*.py
flake8 agent/
```

### Fix Commands
```bash
# Format all code
gofmt -w .

# Update go.mod
sed -i '' 's/go 1.25.4/go 1.21/' go.mod
go mod tidy

# Fix .env permissions
chmod 600 .env
git rm --cached .env
```

---

**Last Updated:** November 10, 2025  
**Next Review:** After Phase 1 completion
