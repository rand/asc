# ASC Project Structure

This document provides an overview of the project organization and where to find things.

## Directory Layout

```
asc/
├── agent/                      # Python agent adapter framework
│   ├── agent_adapter.py        # Main entry point for agents
│   ├── llm_client.py           # LLM provider abstraction
│   ├── phase_loop.py           # Hephaestus task execution loop
│   ├── ace.py                  # ACE playbook learning system
│   ├── heartbeat.py            # Agent status reporting
│   ├── tests/                  # Agent unit tests
│   ├── requirements.txt        # Python dependencies
│   ├── setup.py                # Package installation
│   ├── README.md               # Agent documentation
│   └── VALIDATION.md           # Validation report
│
├── cmd/                        # CLI command implementations
│   ├── root.go                 # Root command and global flags
│   ├── init.go                 # Interactive setup wizard
│   ├── up.go                   # Start agent stack
│   ├── down.go                 # Stop agent stack
│   ├── check.go                # Dependency checker
│   ├── test.go                 # Test runner
│   ├── services.go             # Service management
│   └── secrets.go              # Secrets management
│
├── internal/                   # Internal Go packages
│   ├── config/                 # Configuration management
│   │   ├── config.go           # Config parsing and validation
│   │   └── config_test.go      # Config tests
│   │
│   ├── process/                # Process management
│   │   ├── manager.go          # Process lifecycle management
│   │   └── manager_test.go     # Process tests
│   │
│   ├── check/                  # Dependency checking
│   │   ├── checker.go          # Dependency validation
│   │   └── checker_test.go     # Checker tests
│   │
│   ├── beads/                  # Beads integration
│   │   └── client.go           # Beads client
│   │
│   ├── mcp/                    # MCP integration
│   │   ├── client.go           # MCP HTTP client
│   │   └── client_test.go      # MCP tests
│   │
│   ├── tui/                    # Terminal UI
│   │   ├── model.go            # Bubbletea model
│   │   ├── view.go             # UI rendering
│   │   ├── agents.go           # Agent status pane
│   │   ├── tasks.go            # Task stream pane
│   │   ├── logs.go             # Log pane
│   │   └── wizard.go           # Setup wizard
│   │
│   ├── secrets/                # Secrets management
│   │   ├── secrets.go          # Age encryption
│   │   └── secrets_test.go     # Secrets tests
│   │
│   ├── logger/                 # Logging
│   │   └── logger.go           # Structured logging
│   │
│   └── errors/                 # Error handling
│       └── errors.go           # Custom error types
│
├── test/                       # Integration tests
│   ├── integration_test.go     # Integration test suite
│   └── e2e_test.go             # End-to-end tests
│
├── docs/                       # Documentation
│   ├── README.md               # Documentation index
│   ├── PROJECT_STRUCTURE.md    # This file
│   │
│   ├── specs/                  # Specifications
│   │   └── asc-spec.md         # Original specification
│   │
│   ├── security/               # Security documentation
│   │   ├── SECURITY.md         # Security overview
│   │   ├── SECURITY_IMPROVEMENTS.md
│   │   └── STREAMLINED_SECURITY.md
│   │
│   ├── testing/                # Test documentation
│   │   ├── TEST_REPORT.md      # Test results
│   │   └── TESTING_SUMMARY.md  # Coverage summary
│   │
│   └── archive/                # Historical documents
│       ├── GAP_ANALYSIS.md
│       ├── IMPLEMENTATION_STATUS.md
│       └── NEXT_PHASE_TASKS.md
│
├── .kiro/                      # Kiro spec (active development)
│   └── specs/
│       └── agent-stack-controller/
│           ├── requirements.md # System requirements
│           ├── design.md       # Architecture design
│           └── tasks.md        # Implementation tasks
│
├── build/                      # Build artifacts (gitignored)
├── .env                        # Environment variables (gitignored)
├── .env.example                # Environment template
├── asc.toml                    # Configuration file
├── go.mod                      # Go module definition
├── go.sum                      # Go dependency checksums
├── main.go                     # Application entry point
├── Makefile                    # Build automation
└── README.md                   # Project overview
```

## Key Files

### Configuration
- **asc.toml** - Main configuration file defining agents and services
- **.env** - API keys and secrets (not committed to git)
- **.env.example** - Template for environment variables

### Entry Points
- **main.go** - Go application entry point
- **agent/agent_adapter.py** - Python agent entry point

### Documentation
- **README.md** - Project overview and quick start
- **docs/README.md** - Documentation index
- **agent/README.md** - Agent framework documentation

### Specifications
- **.kiro/specs/agent-stack-controller/** - Active development spec
  - **requirements.md** - What the system must do
  - **design.md** - How the system is architected
  - **tasks.md** - Implementation task list

## Finding Things

### "Where do I find...?"

**CLI commands?**
→ `cmd/` directory - each command has its own file

**Agent implementation?**
→ `agent/` directory - Python agent framework

**Configuration parsing?**
→ `internal/config/config.go`

**Process management?**
→ `internal/process/manager.go`

**TUI dashboard?**
→ `internal/tui/` directory

**Tests?**
→ `*_test.go` files throughout the codebase
→ `agent/tests/` for Python tests
→ `test/` for integration tests

**Documentation?**
→ `docs/` directory
→ `README.md` files in component directories

**Requirements?**
→ `.kiro/specs/agent-stack-controller/requirements.md`

**Design decisions?**
→ `.kiro/specs/agent-stack-controller/design.md`

**Task list?**
→ `.kiro/specs/agent-stack-controller/tasks.md`

**Security info?**
→ `docs/security/` directory

**Test reports?**
→ `docs/testing/` directory

## Code Organization Principles

### Go Code (internal/)

1. **Package per concern**: Each subdirectory is a focused package
2. **Internal only**: Code in `internal/` cannot be imported by external projects
3. **Tests alongside code**: `*_test.go` files in the same directory
4. **Interfaces for testability**: Use interfaces for dependencies

### Python Code (agent/)

1. **Flat structure**: All modules at the top level
2. **Clear naming**: Module names match their purpose
3. **Tests separate**: Tests in `tests/` subdirectory
4. **Package exports**: `__init__.py` defines public API

### Documentation (docs/)

1. **Organized by type**: specs/, security/, testing/, archive/
2. **Index at root**: `docs/README.md` is the entry point
3. **Cross-references**: Documents link to related docs
4. **Archive old docs**: Move outdated docs to archive/

## Development Workflow

### Adding a New Feature

1. **Update spec**: Add requirements to `.kiro/specs/agent-stack-controller/requirements.md`
2. **Design**: Document in `.kiro/specs/agent-stack-controller/design.md`
3. **Create tasks**: Add to `.kiro/specs/agent-stack-controller/tasks.md`
4. **Implement**: Write code in appropriate directory
5. **Test**: Add tests alongside implementation
6. **Document**: Update relevant README files
7. **Validate**: Run tests and update validation reports

### Adding Documentation

1. **Choose location**: specs/, security/, testing/, or component directory
2. **Create file**: Use descriptive name
3. **Update index**: Add to `docs/README.md`
4. **Cross-reference**: Link from related documents
5. **Update main README**: If it's a major document

### Archiving Documents

1. **Move to archive**: `mv doc.md docs/archive/`
2. **Add deprecation notice**: At the top of the document
3. **Update index**: Move to "Archived" section in `docs/README.md`
4. **Update references**: Fix links in other documents

## Build Artifacts

### Generated Files (gitignored)

- `build/` - Compiled binaries
- `coverage.out` - Test coverage data
- `agent/.venv/` - Python virtual environment
- `agent/__pycache__/` - Python bytecode
- `.env` - Local environment variables

### Committed Files

- `go.sum` - Go dependency checksums
- `agent/requirements.txt` - Python dependencies
- `.env.example` - Environment template

## Configuration Files

### Go
- `go.mod` - Module definition and dependencies
- `go.sum` - Dependency checksums

### Python
- `agent/requirements.txt` - Python dependencies
- `agent/setup.py` - Package installation

### Build
- `Makefile` - Build automation
- `.gitignore` - Git ignore rules

### Application
- `asc.toml` - Agent and service configuration
- `.env` - API keys and secrets

## Testing Structure

### Go Tests
- Unit tests: `*_test.go` alongside source files
- Integration tests: `test/integration_test.go`
- E2E tests: `test/e2e_test.go`

### Python Tests
- Unit tests: `agent/tests/test_*.py`
- Test fixtures: `agent/tests/fixtures/`

### Test Reports
- `docs/testing/TEST_REPORT.md` - Comprehensive results
- `docs/testing/TESTING_SUMMARY.md` - Coverage summary
- `agent/VALIDATION.md` - Agent validation

## Quick Navigation

### For New Contributors
1. Start: [README.md](../README.md)
2. Understand: [Requirements](../.kiro/specs/agent-stack-controller/requirements.md)
3. Learn: [Design](../.kiro/specs/agent-stack-controller/design.md)
4. Explore: [Project Structure](PROJECT_STRUCTURE.md) (this file)

### For Developers
1. Tasks: [Task List](../.kiro/specs/agent-stack-controller/tasks.md)
2. Code: `internal/` and `agent/` directories
3. Tests: `*_test.go` and `agent/tests/`
4. Docs: Component README files

### For Reviewers
1. Spec: `.kiro/specs/agent-stack-controller/`
2. Tests: `docs/testing/`
3. Security: `docs/security/`
4. Validation: `agent/VALIDATION.md`

## Maintenance

This document should be updated when:
- New directories are added
- Major files are moved or renamed
- Documentation structure changes
- New development workflows are established

Last updated: 2024-11-09
