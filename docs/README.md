# ASC Documentation

This directory contains all project documentation organized by category.

## Directory Structure

```
docs/
├── README.md              # This file - documentation index
├── specs/                 # Specifications and design documents
├── security/              # Security documentation
├── testing/               # Test reports and summaries
└── archive/               # Historical/deprecated documents
```

## Active Documentation

### Project Overview
- [Main README](../README.md) - Project overview, installation, and usage

### Specifications
- [ASC Specification](specs/asc-spec.md) - Original project specification
- [Kiro Spec](../.kiro/specs/agent-stack-controller/) - Current implementation spec
  - [Requirements](../.kiro/specs/agent-stack-controller/requirements.md)
  - [Design](../.kiro/specs/agent-stack-controller/design.md)
  - [Tasks](../.kiro/specs/agent-stack-controller/tasks.md)

### API and Reference
- [API Reference](API_REFERENCE.md) - Complete API documentation
- [Configuration Reference](CONFIGURATION.md) - Configuration guide
- [Architecture Decision Records](ADR_INDEX.md) - Design decisions and rationale

### Operations
- [Operator's Handbook](OPERATORS_HANDBOOK.md) - Day-to-day operations guide
- [Upgrade Guide](UPGRADE_GUIDE.md) - Version upgrades and migrations
- [FAQ](FAQ.md) - Frequently asked questions

### Component Documentation
- [Agent Adapter](../agent/README.md) - Python agent framework documentation
- [Agent Validation](../agent/VALIDATION.md) - Agent implementation validation report

### Developer Experience
- [Developer Experience Overview](DEVELOPER_EXPERIENCE.md) - Complete DX documentation
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute to the project
- [Quick Start for Developers](../QUICK_START_DEV.md) - Get started in 5 minutes
- [Testing Best Practices](../TESTING.md) - How to write good tests
- [Debugging Guide](../DEBUGGING.md) - Tools and techniques for debugging
- [Troubleshooting](../TROUBLESHOOTING.md) - Solutions to common issues
- [Code Review Checklist](../CODE_REVIEW_CHECKLIST.md) - What reviewers look for
- [Security Policy](../SECURITY.md) - Security practices and reporting

### Dependency Management
- [Dependencies Guide](DEPENDENCIES.md) - Comprehensive dependency documentation
- [Breaking Changes Log](BREAKING_CHANGES.md) - Dependency breaking changes tracking
- [Known Issues](KNOWN_ISSUES.md) - Known dependency issues and workarounds

### Security
- [Security Overview](security/SECURITY.md) - Security features and best practices
- [Security Best Practices](security/SECURITY_BEST_PRACTICES.md) - Security guidelines
- [Security Improvements](security/SECURITY_IMPROVEMENTS.md) - Implemented security enhancements
- [Incident Response Plan](security/INCIDENT_RESPONSE_PLAN.md) - Security incident procedures

### Testing
- [Test Report](testing/TEST_REPORT.md) - Comprehensive test results
- [Testing Summary](testing/TESTING_SUMMARY.md) - Test coverage summary
- [Usability Testing Guide](testing/USABILITY_TESTING_GUIDE.md) - User testing procedures

## Archived Documentation

Historical documents that may be useful for reference but are no longer actively maintained:

- [Gap Analysis](archive/GAP_ANALYSIS.md) - Initial gap analysis
- [Implementation Status](archive/IMPLEMENTATION_STATUS.md) - Historical implementation tracking
- [Next Phase Tasks](archive/NEXT_PHASE_TASKS.md) - Previous phase planning

## Documentation Guidelines

### When Adding New Documentation

1. **Choose the right location:**
   - `specs/` - Design documents, specifications, architecture
   - `security/` - Security-related documentation
   - `testing/` - Test reports, coverage, validation
   - `archive/` - Deprecated or historical documents

2. **Update this index** - Add links to new documents in the appropriate section

3. **Use clear naming:**
   - Use UPPERCASE for standalone documents (e.g., `SECURITY.md`)
   - Use lowercase for component docs (e.g., `agent/README.md`)
   - Use descriptive names that indicate content

4. **Cross-reference:**
   - Link to related documents
   - Update the main README if needed
   - Keep the Kiro spec up to date

### When Deprecating Documentation

1. Move to `archive/` directory
2. Update this index to reflect the change
3. Add a note in the archived document explaining why it was deprecated
4. Keep a reference in the "Archived Documentation" section

## Quick Links

### For New Contributors
1. Start with [Main README](../README.md)
2. Read [Quick Start for Developers](../QUICK_START_DEV.md)
3. Review [Contributing Guide](../CONTRIBUTING.md)
4. Check [ASC Specification](specs/asc-spec.md)

### For Developers
1. [Developer Experience Overview](DEVELOPER_EXPERIENCE.md)
2. [Design Document](../.kiro/specs/agent-stack-controller/design.md)
3. [Task List](../.kiro/specs/agent-stack-controller/tasks.md)
4. [Testing Best Practices](../TESTING.md)
5. [Debugging Guide](../DEBUGGING.md)
6. [Agent Documentation](../agent/README.md)

### For Security Review
1. [Security Overview](security/SECURITY.md)
2. [Security Improvements](security/SECURITY_IMPROVEMENTS.md)
3. [Agent Validation](../agent/VALIDATION.md)

## Maintenance

This documentation structure should be maintained as the project evolves:

- **Weekly**: Review and update test reports
- **Per Feature**: Update specs and design docs
- **Per Release**: Archive outdated documents
- **Quarterly**: Review and consolidate documentation

## Contributing to Documentation

When contributing documentation:

1. Follow the existing structure
2. Use Markdown formatting
3. Include code examples where appropriate
4. Add diagrams for complex concepts (use Mermaid)
5. Keep language clear and concise
6. Update this index

## Questions?

If you can't find what you're looking for:
1. Check the [Main README](../README.md)
2. Search the codebase for inline documentation
3. Review the [Kiro spec](../.kiro/specs/agent-stack-controller/)
4. Open an issue for missing documentation
