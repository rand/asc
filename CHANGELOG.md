# Changelog

All notable changes to the Agent Stack Controller (asc) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive documentation suite including API reference, configuration guide, and troubleshooting
- Security best practices and incident response plan
- Performance monitoring and validation tools
- Dependency compatibility testing framework
- Usability testing guide and common user issues documentation
- Quality gates and metrics tracking
- Developer experience improvements with quick start guide
- CHANGELOG.md following Keep a Changelog format
- VERSIONING.md documenting semantic versioning policy and release process
- Docker setup documentation and installation instructions for optional containerized features
- Docker verification in `asc check` command (shows as optional dependency)
- Comprehensive Docker troubleshooting guide in README and DEPENDENCIES.md

### Changed
- Enhanced error handling across all packages with comprehensive test coverage
- Improved test suite with reduced flakiness and better reliability
- Updated CI/CD workflows for better quality assurance
- Minimum Go version set to 1.24.0 (required by bubbletea dependency)

### Fixed
- Various test flakiness issues identified and resolved
- Error handling edge cases in configuration, process management, and client packages

## [0.1.0] - 2025-11-11

### Added
- Initial release of Agent Stack Controller (asc)
- CLI commands: `init`, `up`, `down`, `check`, `test`, `services`, `doctor`, `secrets`, `cleanup`
- Interactive TUI dashboard with real-time agent monitoring
- Vaporwave aesthetic design system with elegant borders, animations, and typography
- Configuration system with TOML parsing and validation
- Configuration templates (solo, team, swarm) for quick setup
- Hot-reload configuration support with file watching
- Process management for agent lifecycle control
- Dependency checker for environment verification
- Beads client for Git-backed task management integration
- MCP client with HTTP and WebSocket support for real-time updates
- Health monitoring system with automatic recovery
- Structured logging with aggregation and rotation
- Secrets management with age encryption
- Interactive setup wizard with dependency installation guidance
- Agent adapter framework in Python supporting Claude, Gemini, and OpenAI
- Hephaestus phase loop for agent task execution
- ACE (Agentic Context Engineering) for agent learning and playbook management
- Agent heartbeat system for status tracking
- End-to-end testing framework
- Integration and performance testing suites
- Security scanning and validation
- Comprehensive documentation including:
  - User guides and quick start
  - API reference and configuration guide
  - Architecture and design documents
  - Security best practices
  - Troubleshooting and debugging guides
  - Contributing guidelines and code review checklist

### Security
- Age encryption for API keys and sensitive configuration
- Automatic .gitignore setup to prevent secret leakage
- File permission management (0600 for sensitive files)
- Secrets rotation support
- Security scanning with gosec integration
- Incident response plan documentation

## [0.0.1] - 2025-11-09

### Added
- Initial project structure and Go module setup
- Basic CLI framework with cobra
- Core configuration structures
- Process manager interface
- Dependency checker interface
- Beads and MCP client interfaces
- TUI framework with bubbletea
- Basic command implementations

---

## Release Notes

### Version 0.1.0 - Initial Public Release

This is the first public release of the Agent Stack Controller (asc), a command-line orchestration tool for managing a local colony of AI coding agents.

**Highlights:**
- Complete CLI and TUI implementation with vaporwave aesthetic
- Support for multiple LLM providers (Claude, Gemini, OpenAI)
- Real-time agent monitoring and coordination
- Secure secrets management with age encryption
- Hot-reload configuration for dynamic agent management
- Comprehensive testing and documentation

**Breaking Changes:**
- None (initial release)

**Known Issues:**
- WebSocket reconnection may occasionally require manual refresh
- Long-running agents (24+ hours) may experience memory growth
- Terminal resize during modal display may cause layout issues

**Upgrade Notes:**
- This is the initial release, no upgrade path needed

**Dependencies:**
- Go 1.21 or higher
- Python 3.8 or higher
- git, bd (beads CLI), age (for secrets)
- Optional: uv (Python package manager), docker

For detailed documentation, see the [README](README.md) and [docs/](docs/) directory.

