# Upgrade and Migration Guide

Guide for upgrading between versions of the Agent Stack Controller and migrating configurations.

## Table of Contents

- [Version Compatibility](#version-compatibility)
- [Upgrade Process](#upgrade-process)
- [Migration Guides](#migration-guides)
- [Breaking Changes](#breaking-changes)
- [Rollback Procedures](#rollback-procedures)
- [Data Migration](#data-migration)

---

## Version Compatibility

### Current Version: 1.0.0

### Compatibility Matrix

| asc Version | Go Version | Python Version | beads Version | mcp_agent_mail Version |
|-------------|------------|----------------|---------------|------------------------|
| 1.0.x       | 1.21+      | 3.8+           | 1.0+          | 1.0+                   |
| 0.9.x       | 1.21+      | 3.8+           | 0.9+          | 0.9+                   |
| 0.8.x       | 1.20+      | 3.7+           | 0.8+          | 0.8+                   |

### Deprecation Policy

- **Minor versions** (1.x): Deprecated features supported for 2 releases
- **Major versions** (x.0): Breaking changes allowed, migration guide provided
- **Security patches**: Applied to current and previous minor version

---

## Upgrade Process

### General Upgrade Steps

1. **Backup your configuration**
   ```bash
   cp asc.toml asc.toml.backup
   cp .env .env.backup
   cp -r ~/.asc ~/.asc.backup
   ```

2. **Stop the current stack**
   ```bash
   asc down
   ```

3. **Upgrade the binary**
   ```bash
   # Using go install
   go install github.com/yourusername/asc@latest
   
   # Or download new binary
   curl -L https://github.com/yourusername/asc/releases/latest/download/asc-$(uname -s)-$(uname -m) -o asc
   chmod +x asc
   sudo mv asc /usr/local/bin/
   ```

4. **Check for breaking changes**
   ```bash
   asc --version
   # Read CHANGELOG.md for your version
   ```

5. **Validate configuration**
   ```bash
   asc check
   ```

6. **Migrate configuration if needed**
   ```bash
   # See version-specific migration below
   ```

7. **Test the upgrade**
   ```bash
   asc test
   ```

8. **Start the stack**
   ```bash
   asc up
   ```

### Automated Upgrade

```bash
# Download and run upgrade script
curl -L https://github.com/yourusername/asc/raw/main/scripts/upgrade.sh | bash
```

The script will:
- Backup current installation
- Download new version
- Migrate configuration
- Validate setup
- Restart services

---

## Migration Guides

### Migrating from 0.9.x to 1.0.0

#### Breaking Changes

1. **Configuration format changed**
   - Old: `[agents]` section with array
   - New: Individual `[agent.name]` sections

2. **Secrets management added**
   - New: age encryption required
   - Old: Plain `.env` files

3. **Command changes**
   - Removed: `asc start` (use `asc up`)
   - Removed: `asc stop` (use `asc down`)
   - Added: `asc secrets`
   - Added: `asc doctor`

#### Migration Steps

**1. Update configuration format**

Old format (0.9.x):
```toml
[core]
beads_db_path = "./project"

[services]
mcp_url = "http://localhost:8765"

[agents]
agents = [
  { name = "planner", model = "gemini", phases = ["planning"] },
  { name = "coder", model = "claude", phases = ["implementation"] }
]
```

New format (1.0.0):
```toml
[core]
beads_db_path = "./project"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.planner]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning"]

[agent.coder]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation"]
```

**Automated migration:**
```bash
asc migrate config --from=0.9 --to=1.0
```

**2. Set up secrets encryption**

```bash
# Install age if not present
brew install age  # macOS
# or
apt install age   # Linux

# Encrypt existing .env
asc secrets encrypt
```

**3. Update commands in scripts**

Replace:
- `asc start` → `asc up`
- `asc stop` → `asc down`

**4. Test migration**

```bash
asc check
asc test
asc up
```

### Migrating from 0.8.x to 0.9.x

#### Breaking Changes

1. **Python 3.7 → 3.8 minimum**
2. **New agent phases added**
3. **Configuration validation stricter**

#### Migration Steps

**1. Upgrade Python**

```bash
# Check version
python3 --version

# Upgrade if needed (macOS)
brew upgrade python3

# Upgrade if needed (Linux)
apt update && apt upgrade python3
```

**2. Update phase names**

Old phases:
- `plan` → `planning`
- `code` → `implementation`
- `test` → `testing`

**3. Validate configuration**

```bash
asc check
```

---

## Breaking Changes

### Version 1.0.0

**Configuration Format**
- Changed from array-based to section-based agent definitions
- Added `command` field to agent config
- Split `services` into `services.mcp_agent_mail`

**Commands**
- Removed `asc start` (use `asc up`)
- Removed `asc stop` (use `asc down`)
- Added `asc secrets` command group
- Added `asc doctor` command

**Secrets Management**
- Introduced age encryption
- `.env` now gitignored by default
- Added `.env.age` for encrypted secrets

**Dependencies**
- Go 1.21+ required (was 1.20+)
- age binary required for secrets

**Migration:** See [Migrating from 0.9.x to 1.0.0](#migrating-from-09x-to-100)

### Version 0.9.0

**Python Version**
- Minimum Python version: 3.8 (was 3.7)

**Phase Names**
- Standardized phase names (planning, implementation, testing)

**Configuration Validation**
- Stricter validation of model names
- Required fields enforced

**Migration:** See [Migrating from 0.8.x to 0.9.x](#migrating-from-08x-to-09x)

---

## Rollback Procedures

### Rolling Back to Previous Version

If an upgrade fails or causes issues:

**1. Stop the current stack**
```bash
asc down
```

**2. Restore previous binary**
```bash
# If you backed up the binary
sudo cp /usr/local/bin/asc.backup /usr/local/bin/asc

# Or download specific version
curl -L https://github.com/yourusername/asc/releases/download/v0.9.0/asc-$(uname -s)-$(uname -m) -o asc
chmod +x asc
sudo mv asc /usr/local/bin/
```

**3. Restore configuration**
```bash
cp asc.toml.backup asc.toml
cp .env.backup .env
cp -r ~/.asc.backup ~/.asc
```

**4. Verify rollback**
```bash
asc --version
asc check
```

**5. Restart**
```bash
asc up
```

### Rollback Checklist

- [ ] Stop current stack
- [ ] Restore binary
- [ ] Restore configuration files
- [ ] Restore state directory (~/.asc)
- [ ] Verify version
- [ ] Run checks
- [ ] Test functionality
- [ ] Restart stack

---

## Data Migration

### Migrating Agent State

Agent state is stored in `~/.asc/`:

```
~/.asc/
├── pids/          # Process IDs
├── logs/          # Log files
├── playbooks/     # ACE playbooks
├── templates/     # Configuration templates
└── age.key        # Encryption key
```

**Backup:**
```bash
tar -czf asc-state-backup.tar.gz ~/.asc
```

**Restore:**
```bash
tar -xzf asc-state-backup.tar.gz -C ~/
```

### Migrating Between Machines

**Export from old machine:**
```bash
# Backup configuration
tar -czf asc-config.tar.gz asc.toml .env.age

# Backup state
tar -czf asc-state.tar.gz ~/.asc

# Transfer files to new machine
scp asc-config.tar.gz user@newmachine:~/
scp asc-state.tar.gz user@newmachine:~/
```

**Import on new machine:**
```bash
# Install asc
go install github.com/yourusername/asc@latest

# Restore configuration
tar -xzf asc-config.tar.gz

# Restore state
tar -xzf asc-state.tar.gz -C ~/

# Decrypt secrets
asc secrets decrypt

# Verify
asc check
asc test
```

### Migrating Playbooks

ACE playbooks are stored per-agent in `~/.asc/playbooks/`:

**Export playbooks:**
```bash
tar -czf playbooks-backup.tar.gz ~/.asc/playbooks/
```

**Import playbooks:**
```bash
tar -xzf playbooks-backup.tar.gz -C ~/
```

**Merge playbooks:**
```bash
# Copy playbooks from old agent to new agent
cp -r ~/.asc/playbooks/old-agent ~/.asc/playbooks/new-agent
```

### Migrating Beads Database

The beads database is a git repository:

**Clone to new location:**
```bash
git clone /path/to/old/beads /path/to/new/beads
```

**Update configuration:**
```toml
[core]
beads_db_path = "/path/to/new/beads"
```

---

## Version-Specific Notes

### Version 1.0.0

**New Features:**
- Age encryption for secrets
- Configuration templates
- Hot-reload configuration
- Health monitoring and auto-recovery
- WebSocket real-time updates
- Vaporwave aesthetic TUI

**Improvements:**
- Better error messages
- Faster startup
- Lower memory usage
- More robust process management

**Known Issues:**
- WebSocket reconnection may take up to 30s
- Hot-reload doesn't work with .env changes

### Version 0.9.0

**New Features:**
- Interactive setup wizard
- Configuration validation
- Health check command

**Improvements:**
- Better TUI rendering
- Improved error handling

**Known Issues:**
- Configuration migration from 0.8.x requires manual steps

---

## Upgrade Checklist

Before upgrading:

- [ ] Read CHANGELOG for your version
- [ ] Check breaking changes
- [ ] Backup configuration files
- [ ] Backup state directory
- [ ] Note current version
- [ ] Plan downtime window

During upgrade:

- [ ] Stop current stack
- [ ] Upgrade binary
- [ ] Migrate configuration
- [ ] Validate configuration
- [ ] Run health check
- [ ] Test functionality

After upgrade:

- [ ] Verify all agents start
- [ ] Check logs for errors
- [ ] Test task execution
- [ ] Monitor for issues
- [ ] Update documentation
- [ ] Notify team

---

## Getting Help

If you encounter issues during upgrade:

1. **Check the logs**
   ```bash
   tail -f ~/.asc/logs/asc.log
   ```

2. **Run diagnostics**
   ```bash
   asc doctor --verbose
   ```

3. **Search known issues**
   - [GitHub Issues](https://github.com/yourusername/asc/issues)
   - [Known Issues](KNOWN_ISSUES.md)

4. **Ask for help**
   - [GitHub Discussions](https://github.com/yourusername/asc/discussions)
   - [Discord Community](https://discord.gg/asc)

5. **Report bugs**
   - [File an issue](https://github.com/yourusername/asc/issues/new)

---

## See Also

- [CHANGELOG](../CHANGELOG.md)
- [Breaking Changes](BREAKING_CHANGES.md)
- [Known Issues](KNOWN_ISSUES.md)
- [Troubleshooting](../TROUBLESHOOTING.md)
- [Configuration Reference](CONFIGURATION.md)
