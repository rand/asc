# Repository Organization

This document describes the organization and structure of the Agent Stack Controller repository.

## Directory Structure

```
asc/
├── .github/              # GitHub workflows and templates
│   ├── workflows/        # CI/CD workflows
│   ├── ISSUE_TEMPLATE/   # Issue templates
│   └── PULL_REQUEST_TEMPLATE.md
├── .githooks/            # Git hooks
├── .kiro/                # Kiro IDE configuration
│   └── specs/            # Feature specifications
├── agent/                # Python agent framework
│   ├── tests/            # Agent tests
│   └── *.py              # Agent implementation
├── cmd/                  # CLI commands
├── docs/                 # Documentation
│   ├── adr/              # Architecture Decision Records
│   ├── archive/          # Archived documentation
│   ├── completion-reports/ # Historical completion reports
│   ├── quality/          # Quality documentation
│   ├── security/         # Security documentation
│   └── testing/          # Testing documentation
├── internal/             # Internal Go packages
│   ├── beads/            # Beads client
│   ├── check/            # Health checks
│   ├── config/           # Configuration
│   ├── doctor/           # Diagnostics
│   ├── errors/           # Error handling
│   ├── health/           # Health monitoring
│   ├── logger/           # Logging
│   ├── mcp/              # MCP client
│   ├── process/          # Process management
│   ├── secrets/          # Secrets management
│   └── tui/              # Terminal UI
├── scripts/              # Automation scripts
├── test/                 # Integration and E2E tests
├── CHANGELOG.md          # Version history
├── CODE_REVIEW_CHECKLIST.md
├── CONTRIBUTING.md       # Contribution guide
├── DEBUGGING.md          # Debugging guide
├── go.mod                # Go dependencies
├── go.sum                # Go dependency checksums
├── LICENSE               # MIT License
├── Makefile              # Build automation
├── PROJECT_ROADMAP.md    # Project roadmap
├── QUICK_START_DEV.md    # Developer quick start
├── README.md             # Project overview
├── SECURITY.md           # Security policy
├── TESTING.md            # Testing guide
├── TROUBLESHOOTING.md    # Troubleshooting guide
└── VERSIONING.md         # Versioning policy
```

## Documentation Organization

### Root Level Documentation

**User-Facing:**
- `README.md` - Project overview, features, quick start
- `SECURITY.md` - Security policy and reporting
- `TROUBLESHOOTING.md` - Common issues and solutions
- `DEBUGGING.md` - Debugging techniques

**Developer-Facing:**
- `CONTRIBUTING.md` - How to contribute
- `CODE_REVIEW_CHECKLIST.md` - Code review guidelines
- `TESTING.md` - Testing best practices
- `QUICK_START_DEV.md` - Developer quick start
- `VERSIONING.md` - Versioning policy
- `CHANGELOG.md` - Version history
- `PROJECT_ROADMAP.md` - Project roadmap

### docs/ Directory

**Main Documentation:**
- `INDEX.md` - Complete documentation index
- `README.md` - Documentation overview and standards
- `CONFIGURATION.md` - Configuration reference
- `API_REFERENCE.md` - API documentation
- `CODE_EXAMPLES.md` - Practical examples
- `FAQ.md` - Frequently asked questions
- `OPERATORS_HANDBOOK.md` - Operations guide

**Feature Documentation:**
- `HOT_RELOAD.md` - Hot-reload feature
- `INTERACTIVE_TUI_FEATURES.md` - TUI features
- `LOGGING_AND_DEBUGGING.md` - Logging details
- `TEMPLATES.md` - Project templates
- `WEBSOCKET_IMPLEMENTATION.md` - WebSocket details

**Reference Documentation:**
- `DEPENDENCIES.md` - Dependency documentation
- `DEPENDENCY_COMPATIBILITY.md` - Compatibility matrix
- `DEPENDENCY_QUICK_REFERENCE.md` - Quick reference
- `BREAKING_CHANGES.md` - Breaking changes
- `KNOWN_ISSUES.md` - Known issues
- `COMMON_USER_ISSUES.md` - User issues

**Development Documentation:**
- `DEVELOPER_EXPERIENCE.md` - Developer experience
- `PROJECT_STRUCTURE.md` - Code organization
- `ORGANIZATION_SUMMARY.md` - Organization details
- `INTEGRATION_TESTING.md` - Integration testing
- `PERFORMANCE.md` - Performance guide
- `PERFORMANCE_VALIDATION.md` - Performance validation
- `UPGRADE_GUIDE.md` - Upgrade instructions
- `QUICK_VALIDATION_GUIDE.md` - Quick validation
- `TEST_SUITE_REVIEW_PLAN.md` - Test suite review
- `VAPORWAVE_DESIGN.md` - TUI design system
- `VAPORWAVE_IMPLEMENTATION_SUMMARY.md` - TUI implementation

### docs/adr/ - Architecture Decision Records

- `ADR_INDEX.md` - Index of all ADRs
- `ADR-TEMPLATE.md` - Template for new ADRs
- `ADR-0001-go-for-cli-tui.md` - Go language choice
- `ADR-0011-age-encryption.md` - Age encryption choice

### docs/security/ - Security Documentation

- `SECURITY_BEST_PRACTICES.md` - Security best practices
- `SECURITY_IMPROVEMENTS.md` - Security improvements
- `STREAMLINED_SECURITY.md` - Streamlined security
- `INCIDENT_RESPONSE_PLAN.md` - Incident response

### docs/testing/ - Testing Documentation

- `TESTING_SUMMARY.md` - Testing overview
- `TEST_REPORT.md` - Test reports
- `TEST_GAP_ANALYSIS.md` - Coverage gaps
- `TEST_FIX_SUMMARY.md` - Test fixes
- `TEST_QUALITY_IMPROVEMENTS.md` - Quality improvements
- `TEST_REMEDIATION_REPORT.md` - Remediation report
- `FLAKINESS_ANALYSIS.md` - Flaky test analysis
- `USABILITY_TESTING_GUIDE.md` - Usability testing
- `USABILITY_TEST_SUMMARY.md` - Usability results

### docs/quality/ - Quality Documentation

- `QUALITY_GATES_IMPLEMENTATION.md` - Quality gates
- `QUALITY_GATES_VERIFICATION.md` - Gate verification
- `QUALITY_METRICS.md` - Quality metrics

### docs/completion-reports/ - Historical Reports

Archived completion reports from development phases:
- `phase-28/` - Testing and quality assurance
- `phase-29/` - Validation and gap analysis
- `phase-30/` - Remediation work
- Various validation and compatibility reports

### docs/archive/ - Archived Documentation

Historical documentation that's no longer current but kept for reference.

## Scripts Organization

All automation scripts are in `scripts/` with a README:

**Testing & Validation:**
- `test-examples.sh` - Test code examples in documentation
- `validate-links.sh` - Validate markdown links
- `check-flakiness.sh` - Check for flaky tests
- `analyze-test-timing.sh` - Analyze test performance

**Security & Quality:**
- `check-security.sh` - Run security scans
- `run-validation.sh` - Run full validation suite

**Performance:**
- `profile-performance.sh` - Profile performance
- `run-performance-validation.sh` - Validate performance

**Integration:**
- `run-integration-validation.sh` - Integration validation
- `test-dependency-compatibility.sh` - Test dependencies

## Code Organization

### cmd/ - CLI Commands

Each command has its own file:
- `check.go` - Health checks
- `cleanup.go` - Cleanup operations
- `doctor.go` - Diagnostics
- `down.go` - Stop agents
- `init.go` - Initialize project
- `secrets.go` - Secrets management
- `services.go` - Service management
- `test.go` - Testing
- `up.go` - Start agents

### internal/ - Internal Packages

**Core Packages:**
- `config/` - Configuration management
- `process/` - Process management
- `logger/` - Logging infrastructure
- `errors/` - Error handling

**Integration Packages:**
- `beads/` - Beads task database client
- `mcp/` - MCP agent mail client

**Feature Packages:**
- `tui/` - Terminal user interface
- `check/` - Health checking
- `doctor/` - Diagnostics
- `health/` - Health monitoring
- `secrets/` - Secrets encryption

### test/ - Integration Tests

- `e2e_test.go` - End-to-end tests
- `e2e_comprehensive_test.go` - Comprehensive E2E
- `integration_test.go` - Integration tests
- `performance_test.go` - Performance tests
- `security_test.go` - Security tests
- `usability_test.go` - Usability tests
- Various validation test files

## Maintenance

### Adding New Documentation

1. Determine the appropriate location:
   - User guides → `docs/`
   - Developer guides → `docs/` or root
   - Architecture decisions → `docs/adr/`
   - Security → `docs/security/`
   - Testing → `docs/testing/`
   - Quality → `docs/quality/`

2. Create the document following existing patterns

3. Update `docs/INDEX.md` with a link to the new document

4. Update `README.md` if it's a major user-facing document

### Adding Completion Reports

New completion reports should go in `docs/completion-reports/` organized by phase.

### Archiving Documentation

When documentation becomes outdated but should be kept for reference:
1. Move to `docs/archive/`
2. Add a note at the top indicating it's archived
3. Update any links pointing to it

## Navigation

- Start with `README.md` for project overview
- Use `docs/INDEX.md` for complete documentation navigation
- Check `PROJECT_ROADMAP.md` for project status
- See `.kiro/specs/agent-stack-controller/tasks.md` for current tasks

## Recent Changes

### November 2024 - Repository Organization

- Created `docs/completion-reports/` for historical reports
- Moved 50+ completion reports from root to organized directories
- Created `docs/quality/` for quality documentation
- Created `docs/INDEX.md` for complete documentation navigation
- Updated `README.md` with better documentation links
- Moved quick reference guides to `docs/`
- Created this organization document
