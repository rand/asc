# Troubleshooting Guide

This guide provides solutions to common issues you may encounter when using or developing asc.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Installation Issues](#installation-issues)
- [Configuration Issues](#configuration-issues)
- [Runtime Issues](#runtime-issues)
- [Agent Issues](#agent-issues)
- [TUI Issues](#tui-issues)
- [Performance Issues](#performance-issues)
- [Development Issues](#development-issues)

## Quick Diagnostics

Before diving into specific issues, run the built-in diagnostic tool:

```bash
# Run comprehensive diagnostics
asc doctor

# Get detailed information
asc doctor --verbose

# Automatically fix common issues
asc doctor --fix

# Output as JSON for automation
asc doctor --json
```

The doctor command checks for:
- Configuration problems
- Corrupted state (PIDs, logs)
- Permission issues
- Resource problems
- Network connectivity
- Agent health issues

Many common issues can be automatically fixed with the `--fix` flag.

## Installation Issues

### "go: command not found"

**Problem:** Go is not installed or not in PATH.

**Solution:**

1. Install Go from https://golang.org/dl/
2. Add Go to your PATH:

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$(go env GOPATH)/bin

# Reload shell
source ~/.bashrc  # or ~/.zshrc
```

3. Verify installation:

```bash
go version
```

### "make: command not found"

**Problem:** Make is not installed.

**Solution:**

**macOS:**
```bash
xcode-select --install
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install build-essential
```

**Linux (Fedora/RHEL):**
```bash
sudo dnf install make
```

### "Build fails with missing dependencies"

**Problem:** Go modules not downloaded.

**Solution:**

```bash
# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Clean and rebuild
make clean
make build
```

### "Permission denied when installing"

**Problem:** Insufficient permissions to install to system directories.

**Solution:**

```bash
# Option 1: Install to user directory
make build
cp build/asc ~/bin/asc  # Ensure ~/bin is in PATH

# Option 2: Use sudo (not recommended)
sudo make install

# Option 3: Use go install (installs to $GOPATH/bin)
go install .
```

## Configuration Issues

### "asc.toml not found"

**Problem:** Configuration file doesn't exist.

**Solution:**

```bash
# Run init to create default config
asc init

# Or create manually
cat > asc.toml << 'EOF'
[core]
beads_db_path = "./project-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.main-agent]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning", "implementation"]
EOF
```

### "Invalid TOML syntax"

**Problem:** Configuration file has syntax errors.

**Solution:**

1. Validate TOML syntax:

```bash
# Use a TOML validator
go install github.com/pelletier/go-toml/v2/cmd/tomljson@latest
tomljson asc.toml

# Or use Python
python -c "import tomli; tomli.load(open('asc.toml', 'rb'))"
```

2. Common TOML mistakes:

```toml
# Wrong: Missing quotes
beads_db_path = ./path

# Right: Quoted strings
beads_db_path = "./path"

# Wrong: Invalid array syntax
phases = [planning, implementation]

# Right: Quoted array elements
phases = ["planning", "implementation"]
```

### "API key not found"

**Problem:** .env file missing or API keys not set.

**Solution:**

1. Create .env file:

```bash
cat > .env << 'EOF'
CLAUDE_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=...
EOF

# Set restrictive permissions
chmod 600 .env
```

2. Verify keys are loaded:

```bash
# Check if .env exists
ls -la .env

# Verify content (be careful not to expose keys)
grep -c "API_KEY" .env
```

3. If using encrypted secrets:

```bash
# Decrypt secrets
asc secrets decrypt

# Verify decryption
cat .env
```

### "beads_db_path not accessible"

**Problem:** Path to beads database is incorrect or doesn't exist.

**Solution:**

```bash
# Check if path exists
ls -la ./project-repo

# Initialize beads if needed
cd ./project-repo
bd init

# Update asc.toml with correct path
# Use absolute path if relative path doesn't work
[core]
beads_db_path = "/full/path/to/project-repo"
```

## Runtime Issues

### "Failed to start mcp_agent_mail"

**Problem:** MCP server won't start.

**Solution:**

1. Check if Python and mcp_agent_mail are installed:

```bash
python3 --version
python3 -m pip list | grep mcp-agent-mail
```

2. Install if missing:

```bash
pip install mcp-agent-mail
# or
uv pip install mcp-agent-mail
```

3. Check if port is already in use:

```bash
lsof -i :8765
# If something is using the port, kill it:
kill -9 <PID>
```

4. Test manual start:

```bash
python -m mcp_agent_mail.server
```

5. Check logs:

```bash
tail -f ~/.asc/logs/mcp_agent_mail.log
```

### "Port already in use"

**Problem:** Another process is using the required port.

**Solution:**

1. Find what's using the port:

```bash
# macOS/Linux
lsof -i :8765

# Or use netstat
netstat -an | grep 8765
```

2. Kill the process:

```bash
kill -9 <PID>
```

3. Or change the port in asc.toml:

```toml
[services.mcp_agent_mail]
url = "http://localhost:8766"  # Use different port
```

### "Agent processes won't start"

**Problem:** Agents fail to launch.

**Solution:**

1. Check agent command exists:

```bash
which python
ls -la agent_adapter.py
```

2. Test agent manually:

```bash
export AGENT_NAME=test
export AGENT_MODEL=claude
export AGENT_PHASES=planning
export MCP_MAIL_URL=http://localhost:8765
export BEADS_DB_PATH=./project-repo
export CLAUDE_API_KEY=your-key

python agent_adapter.py
```

3. Check agent logs:

```bash
tail -f ~/.asc/logs/test.log
```

4. Verify environment variables:

```bash
# In agent code, print env vars
import os
print(os.environ)
```

### "Zombie processes after shutdown"

**Problem:** Processes not cleaned up properly.

**Solution:**

1. Find zombie processes:

```bash
ps aux | grep -E '(asc|agent_adapter|mcp_agent_mail)' | grep -v grep
```

2. Kill them:

```bash
pkill -9 -f agent_adapter
pkill -9 -f mcp_agent_mail
```

3. Clean up PID files:

```bash
rm -rf ~/.asc/pids/*
```

4. Restart cleanly:

```bash
asc down
asc up
```

## Agent Issues

### "Agent stuck in 'Working' state"

**Problem:** Agent appears to be working but not making progress.

**Solution:**

1. Check agent logs:

```bash
tail -f ~/.asc/logs/agent-name.log
```

2. Check if process is alive:

```bash
ps aux | grep agent-name
```

3. Check file leases:

```bash
curl http://localhost:8765/leases | jq
```

4. Restart the agent:

```bash
# Find PID
ps aux | grep agent-name

# Kill it
kill <PID>

# asc will restart it automatically (if health monitoring is enabled)
# Or restart manually:
asc down
asc up
```

### "Agent not picking up tasks"

**Problem:** Tasks exist but agent doesn't claim them.

**Solution:**

1. Verify agent phases match task phases:

```bash
# Check agent config
grep -A 3 "agent.name" asc.toml

# Check task phases
bd list --json | jq '.[] | {id, phase}'
```

2. Check if agent is running:

```bash
ps aux | grep agent-name
```

3. Check agent logs for errors:

```bash
grep -i error ~/.asc/logs/agent-name.log
```

4. Verify beads database is accessible:

```bash
cd $BEADS_DB_PATH
bd list
```

5. Check MCP connectivity:

```bash
curl http://localhost:8765/health
```

### "Agent crashes repeatedly"

**Problem:** Agent keeps crashing and restarting.

**Solution:**

1. Check crash logs:

```bash
tail -n 100 ~/.asc/logs/agent-name.log
```

2. Common causes:
   - Invalid API key
   - Network issues
   - Out of memory
   - Python dependency issues

3. Test agent in isolation:

```bash
# Run with debug output
export ASC_LOG_LEVEL=debug
python agent_adapter.py
```

4. Check Python dependencies:

```bash
cd agent/
pip install -r requirements.txt
```

### "LLM API errors"

**Problem:** Agent can't communicate with LLM API.

**Solution:**

1. Verify API key:

```bash
# Test Claude API
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $CLAUDE_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{"model":"claude-3-sonnet-20240229","max_tokens":1024,"messages":[{"role":"user","content":"Hello"}]}'

# Test OpenAI API
curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}]}'
```

2. Check rate limits:
   - Claude: 50 requests/minute
   - OpenAI: Varies by tier
   - Gemini: 60 requests/minute

3. Check for API outages:
   - https://status.anthropic.com/
   - https://status.openai.com/
   - https://status.cloud.google.com/

## TUI Issues

### "TUI not rendering correctly"

**Problem:** Display is garbled or not showing properly.

**Solution:**

1. Check terminal compatibility:

```bash
echo $TERM
# Should be xterm-256color or similar

# Set if needed
export TERM=xterm-256color
```

2. Check terminal size:

```bash
echo $COLUMNS $LINES
# Should be at least 80x24

# Resize terminal if needed
```

3. Test color support:

```bash
tput colors
# Should output 256 or higher
```

4. Try different terminal emulator:
   - iTerm2 (macOS)
   - Alacritty (cross-platform)
   - Windows Terminal (Windows)

### "Colors not showing"

**Problem:** TUI displays without colors.

**Solution:**

1. Enable 256 color support:

```bash
export TERM=xterm-256color
```

2. Check terminal capabilities:

```bash
tput colors
```

3. Test with simple color output:

```bash
printf "\033[38;5;82mGreen\033[0m\n"
```

### "Keyboard shortcuts not working"

**Problem:** Key presses don't trigger actions.

**Solution:**

1. Check if another process is capturing input:

```bash
# Kill any other TUI apps
pkill -f bubbletea
```

2. Verify terminal is in raw mode (should be automatic)

3. Try different keys:
   - Some terminals intercept certain key combinations
   - Try alternative shortcuts if available

4. Check logs for key events:

```bash
tail -f /tmp/asc-tui-debug.log
```

### "TUI freezes or becomes unresponsive"

**Problem:** TUI stops responding to input.

**Solution:**

1. Check if process is hung:

```bash
ps aux | grep asc
```

2. Send SIGQUIT to get stack trace:

```bash
kill -QUIT <PID>
# Check logs for stack trace
```

3. Force quit and restart:

```bash
# Press Ctrl+C or Ctrl+\
# Or kill from another terminal
pkill -9 asc

# Restart
asc up
```

## Performance Issues

### "High CPU usage"

**Problem:** asc or agents consuming too much CPU.

**Solution:**

1. Identify the culprit:

```bash
top -o cpu
# or
htop
```

2. Profile the process:

```bash
# For Go code
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# For Python agents
py-spy top --pid <PID>
```

3. Common causes:
   - Tight polling loops (reduce poll frequency)
   - Inefficient rendering (optimize TUI updates)
   - Too many goroutines (check for leaks)

4. Adjust polling intervals in asc.toml:

```toml
[core]
refresh_interval = 5  # Increase from default 2 seconds
```

### "High memory usage"

**Problem:** Memory consumption grows over time.

**Solution:**

1. Check memory usage:

```bash
ps aux | grep asc
# Look at RSS column
```

2. Profile memory:

```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

3. Common causes:
   - Log buffer not being cleared
   - Task list growing unbounded
   - Goroutine leaks

4. Restart to clear memory:

```bash
asc down
asc up
```

### "Slow startup"

**Problem:** asc takes a long time to start.

**Solution:**

1. Check what's slow:

```bash
# Add timing logs
time asc up
```

2. Common causes:
   - Large beads database (optimize git operations)
   - Many agents to start (start in parallel)
   - Slow dependency checks (cache results)

3. Skip checks if you know they pass:

```bash
asc up --skip-checks  # If this flag exists
```

## Development Issues

### "Tests failing"

**Problem:** Tests don't pass.

**Solution:**

1. Run tests with verbose output:

```bash
go test -v ./...
```

2. Run specific failing test:

```bash
go test -v -run TestName ./package
```

3. Check for race conditions:

```bash
go test -race ./...
```

4. Clean test cache:

```bash
go clean -testcache
go test ./...
```

### "Build fails"

**Problem:** Code doesn't compile.

**Solution:**

1. Check Go version:

```bash
go version
# Should be 1.21 or later
```

2. Update dependencies:

```bash
go mod tidy
go mod download
```

3. Clean and rebuild:

```bash
make clean
make build
```

4. Check for syntax errors:

```bash
go vet ./...
```

### "Linter errors"

**Problem:** golangci-lint reports issues.

**Solution:**

1. Run linter:

```bash
golangci-lint run ./...
```

2. Auto-fix some issues:

```bash
golangci-lint run --fix ./...
```

3. Format code:

```bash
make fmt
```

4. If linter is too strict, configure .golangci.yml

### "Import cycle detected"

**Problem:** Circular dependency between packages.

**Solution:**

1. Identify the cycle:

```bash
go build ./...
# Error message will show the cycle
```

2. Refactor to break the cycle:
   - Extract shared code to a new package
   - Use interfaces to invert dependencies
   - Move code to eliminate the dependency

## Getting More Help

If you can't find a solution here:

1. **Check the logs** - Most issues leave traces in logs
2. **Search issues** - https://github.com/yourusername/asc/issues
3. **Ask in discussions** - https://github.com/yourusername/asc/discussions
4. **Read the docs** - Check [docs/README.md](docs/README.md)
5. **File an issue** - If it's a bug, open an issue with:
   - Steps to reproduce
   - Expected vs actual behavior
   - Logs and error messages
   - Environment (OS, Go version, etc.)

## Diagnostic Commands

Run these to gather information for bug reports:

```bash
# System information
uname -a
go version
python3 --version

# asc version
asc --version

# Configuration
cat asc.toml

# Process status
ps aux | grep -E '(asc|agent|mcp)'

# Port usage
lsof -i :8765

# Logs (last 50 lines)
tail -n 50 ~/.asc/logs/asc.log

# Disk space
df -h ~/.asc

# File permissions
ls -la asc.toml .env ~/.asc/
```

---

Still stuck? Don't hesitate to ask for help! The community is here to support you.
