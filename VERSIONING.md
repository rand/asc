# Versioning Policy

This document describes the versioning policy for the Agent Stack Controller (asc) project.

## Semantic Versioning

The Agent Stack Controller follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).

Given a version number `MAJOR.MINOR.PATCH`, we increment:

- **MAJOR** version when we make incompatible API changes
- **MINOR** version when we add functionality in a backward compatible manner
- **PATCH** version when we make backward compatible bug fixes

Additional labels for pre-release and build metadata are available as extensions to the `MAJOR.MINOR.PATCH` format.

### Version Format

```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

**Examples:**
- `1.0.0` - Stable release
- `1.2.3` - Stable release with patches
- `2.0.0-alpha.1` - Pre-release alpha version
- `2.0.0-beta.2` - Pre-release beta version
- `2.0.0-rc.1` - Release candidate
- `1.0.0+20250111` - Build metadata

## Version Numbering Rules

### Major Version (X.0.0)

Increment the MAJOR version when making **incompatible changes** such as:

- Breaking changes to CLI command syntax or flags
- Removal of CLI commands or subcommands
- Breaking changes to configuration file format (asc.toml)
- Removal of configuration options without backward compatibility
- Breaking changes to agent environment variables
- Changes to process management that break existing workflows
- Incompatible changes to log formats or file locations
- Breaking changes to TUI keyboard shortcuts or behavior

**Example:** Changing `asc up` to require a configuration file path as an argument instead of auto-discovering it.

### Minor Version (0.X.0)

Increment the MINOR version when adding **new functionality** in a backward compatible manner:

- New CLI commands or subcommands
- New configuration options (with sensible defaults)
- New TUI features or panes
- New agent capabilities or LLM provider support
- New health monitoring or diagnostic features
- Performance improvements without breaking changes
- New documentation or examples
- Deprecation warnings (before removal in next major version)

**Example:** Adding a new `asc logs` command to view agent logs without opening the TUI.

### Patch Version (0.0.X)

Increment the PATCH version when making **backward compatible bug fixes**:

- Bug fixes that don't change functionality
- Security patches
- Documentation corrections
- Dependency updates (non-breaking)
- Performance optimizations without API changes
- Test improvements
- Build or CI/CD fixes

**Example:** Fixing a bug where the TUI crashes on terminal resize.

## Pre-Release Versions

Pre-release versions are denoted by appending a hyphen and a series of dot-separated identifiers:

### Alpha (X.Y.Z-alpha.N)

- Early development versions
- Features may be incomplete or unstable
- Breaking changes may occur between alpha releases
- Not recommended for production use

**Example:** `2.0.0-alpha.1`, `2.0.0-alpha.2`

### Beta (X.Y.Z-beta.N)

- Feature-complete but may contain bugs
- API is stabilizing but minor changes may occur
- Suitable for testing and feedback
- Not recommended for production use

**Example:** `2.0.0-beta.1`, `2.0.0-beta.2`

### Release Candidate (X.Y.Z-rc.N)

- Final testing before stable release
- No new features, only bug fixes
- API is frozen
- Suitable for production testing
- Will become stable release if no critical issues found

**Example:** `2.0.0-rc.1`, `2.0.0-rc.2`

## Build Metadata

Build metadata can be appended with a plus sign:

```
1.0.0+20250111
1.0.0+build.123
1.0.0+sha.5114f85
```

Build metadata:
- Does NOT affect version precedence
- Used for CI/CD tracking
- Can include build date, commit hash, or build number

## Compatibility Guarantees

### Within Major Versions

**Guaranteed Compatible:**
- Configuration files (asc.toml) from earlier minor/patch versions
- CLI commands and flags (may add new ones, won't remove)
- Agent environment variables
- Log file formats and locations
- Process management behavior

**May Change (with deprecation warnings):**
- Internal APIs (not documented as public)
- Experimental features (marked as such in documentation)
- Default values for new configuration options

### Across Major Versions

**No Compatibility Guarantees:**
- Configuration file format may change
- CLI commands or flags may be removed or changed
- Agent environment variables may change
- Log formats may change
- Process management may change

**Migration Support:**
- Migration guides provided in CHANGELOG
- Automated migration tools when possible
- Deprecation warnings in previous major version

## Deprecation Policy

When deprecating features:

1. **Announce** deprecation in CHANGELOG and documentation
2. **Warn** users at runtime when using deprecated features
3. **Maintain** deprecated features for at least one minor version
4. **Remove** in next major version with migration guide

**Example Timeline:**
- `1.5.0` - Feature X deprecated, warning added
- `1.6.0` - Feature X still works with warning
- `2.0.0` - Feature X removed, migration guide provided

## Release Process

### 1. Version Planning

- Review issues and pull requests
- Determine version bump (major, minor, or patch)
- Create milestone for target version
- Update project roadmap

### 2. Development

- Create feature branches from `main`
- Implement changes with tests
- Update documentation
- Submit pull requests for review

### 3. Pre-Release Testing

- Merge approved PRs to `main`
- Create pre-release tag (alpha, beta, or rc)
- Run comprehensive test suite
- Perform manual testing
- Gather feedback from testers

### 4. Release Preparation

- Update CHANGELOG.md with all changes
- Update version in relevant files
- Update documentation with new features
- Create migration guide if needed
- Review and update README

### 5. Release

- Create release tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
- Push tag: `git push origin vX.Y.Z`
- Build binaries for all platforms
- Create GitHub release with binaries
- Update package registries (if applicable)
- Announce release

### 6. Post-Release

- Monitor for critical issues
- Prepare patch releases if needed
- Update documentation based on feedback
- Plan next version

## Version Tagging

### Git Tags

All releases are tagged in git with the format `vX.Y.Z`:

```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push tag to remote
git push origin v1.0.0

# List all tags
git tag -l

# Show tag details
git show v1.0.0
```

### Tag Naming Convention

- Stable releases: `v1.0.0`, `v1.2.3`, `v2.0.0`
- Pre-releases: `v2.0.0-alpha.1`, `v2.0.0-beta.1`, `v2.0.0-rc.1`
- Always prefix with `v`
- Use lowercase for pre-release identifiers

## Version in Code

### Go Module Version

The `go.mod` file specifies the minimum Go version required:

```go
module github.com/yourusername/asc

go 1.24.0
```

**Note:** The minimum Go version is determined by our dependencies. Currently, `github.com/charmbracelet/bubbletea@v1.3.10` requires Go 1.24.0.

This should be updated when:
- Using new Go language features
- Requiring newer standard library APIs
- Updating minimum supported Go version
- Upgrading dependencies that require newer Go versions

### Application Version

The application version is embedded at build time:

```bash
# Build with version
go build -ldflags "-X main.Version=1.0.0" -o asc

# Display version
asc --version
```

Version information should be displayed by:
- `asc --version` command
- `asc version` command (if implemented)
- TUI footer or about dialog

## Backward Compatibility

### Configuration Files

- Old configuration files must work with new versions (within major version)
- New options should have sensible defaults
- Deprecated options should warn but still work
- Provide migration tools for major version upgrades

### CLI Interface

- Existing commands and flags must continue to work
- New flags should be optional with defaults
- Deprecated commands should warn but still work
- Breaking changes only in major versions

### Agent Environment

- Environment variables must remain stable
- New variables can be added
- Existing variables should not change meaning
- Deprecated variables should warn but still work

## Version Support Policy

### Active Support

- **Current Major Version**: Full support with new features and bug fixes
- **Previous Major Version**: Security fixes and critical bug fixes for 6 months after new major release

### End of Life

- Versions older than previous major version are end-of-life
- No updates or support provided
- Users encouraged to upgrade

**Example:**
- v2.0.0 released on 2025-06-01
- v1.x.x receives security fixes until 2025-12-01
- v0.x.x is end-of-life immediately

## Breaking Changes

All breaking changes must be:

1. **Documented** in CHANGELOG under "Breaking Changes" section
2. **Announced** with deprecation warnings in previous version
3. **Explained** with migration guide
4. **Justified** with clear reasoning

### Breaking Change Checklist

- [ ] Is this change absolutely necessary?
- [ ] Can it be done in a backward compatible way?
- [ ] Has deprecation warning been added in previous version?
- [ ] Is migration guide prepared?
- [ ] Are all affected users notified?
- [ ] Is the change documented in CHANGELOG?

## Version Queries

Users can check version compatibility:

```bash
# Check current version
asc --version

# Check if configuration is compatible
asc check --config-version

# Validate configuration for upgrade
asc doctor --check-upgrade
```

## Release Checklist

Use this checklist for each release:

### Pre-Release
- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped in code
- [ ] Migration guide prepared (if breaking changes)
- [ ] Security audit completed
- [ ] Performance benchmarks run

### Release
- [ ] Git tag created and pushed
- [ ] Binaries built for all platforms
- [ ] GitHub release created
- [ ] Release notes published
- [ ] Documentation site updated
- [ ] Package registries updated

### Post-Release
- [ ] Release announced (blog, social media, etc.)
- [ ] Monitor for critical issues
- [ ] Update project roadmap
- [ ] Close milestone
- [ ] Thank contributors

## Questions?

For questions about versioning:
- Open an issue on GitHub
- Check the CHANGELOG for version history
- Review the project roadmap for planned versions
- Join community discussions

## References

- [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html)
- [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
- [Go Modules Version Numbers](https://go.dev/doc/modules/version-numbers)
- [GitHub Release Best Practices](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases)

