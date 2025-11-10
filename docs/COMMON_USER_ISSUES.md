# Common User Issues and Solutions

This document catalogs common issues users encounter with the Agent Stack Controller (asc) and provides clear solutions.

## Table of Contents

1. [Installation and Setup](#installation-and-setup)
2. [Configuration Issues](#configuration-issues)
3. [Agent Problems](#agent-problems)
4. [TUI Display Issues](#tui-display-issues)
5. [Performance Issues](#performance-issues)
6. [Network and Connectivity](#network-and-connectivity)
7. [File and Permission Issues](#file-and-permission-issues)

---

## Installation and Setup

### Issue: Command Not Found After Installation

**Symptoms**: Running `asc` returns "command not found"

**Causes**:
- Binary not in PATH
- Installation incomplete
- Wrong installation directory

**Solutions**:

1. **Check installation location**:
   ```bash
   which asc
   ```

2. **Add to PATH** (if installed in custom location):
   ```bash
   export PATH=$PATH:/path/to/asc
   ```

3. **Reinstall using go install**:
   ```bash
   go install github.com/yourusername/asc@latest
   ```

4. **Verify GOPATH is in PATH**:
   ```bash
   echo $GOPATH/bin
   export PATH=$PATH:$GOPATH/bin
   ```

### Issue: Init Command Fails

**Symptoms**: `asc init` exits with error

**Causes**:
- Missing dependencies
- Permission issues
- Existing configuration conflicts

**Solutions**:

1. **Check dependencies first**:
   ```bash
   asc check
   ```

2. **Run with verbose output**:
   ```bash
   asc init --verbose
   ```

3. **Backup and remove existing config**:
   ```bash
   mv asc.toml asc.toml.backup
   mv .env .env.backup
   asc init
   ```

4. **Check directory permissions**:
   ```bash
   ls -la
   # Ensure you have write permissions
   ```

---

## Configuration Issues

### Issue: API Keys Not Loaded

**Symptoms**: Agents fail with authentication errors

**Causes**:
- .env file not in correct location
- Wrong file permissions
- Incorrect key format
- Keys not exported to environment

**Solutions**:

1. **Verify .env location**:
   ```bash
   ls -la .env
   # Should be in project root
   ```

2. **Check file permissions**:
   ```bash
   chmod 600 .env
   # .env should be readable only by owner
   ```

3. **Verify .env format**:
   ```bash
   cat .env
   # Should be:
   # CLAUDE_API_KEY=sk-ant-...
   # OPENAI_API_KEY=sk-...
   # GOOGLE_API_KEY=...
   ```

4. **Test with check command**:
   ```bash
   asc check
   # Should show .env status
   ```

5. **Manually export for testing**:
   ```bash
   export CLAUDE_API_KEY=your-key-here
   asc up
   ```

### Issue: Invalid Configuration Syntax

**Symptoms**: `asc check` reports config parse error

**Causes**:
- TOML syntax error
- Missing required fields
- Incorrect data types

**Solutions**:

1. **Check TOML syntax**:
   ```bash
   # Use online TOML validator or:
   python3 -c "import toml; toml.load('asc.toml')"
   ```

2. **Common syntax errors**:
   ```toml
   # Wrong: Missing closing bracket
   [core
   beads_db_path = "./repo"
   
   # Right:
   [core]
   beads_db_path = "./repo"
   
   # Wrong: Unquoted string with spaces
   command = python agent.py
   
   # Right:
   command = "python agent.py"
   ```

3. **Validate required fields**:
   ```toml
   [core]
   beads_db_path = "./repo"  # Required
   
   [agent.name]
   command = "..."  # Required
   model = "..."    # Required
   phases = [...]   # Required
   ```

4. **Use a template**:
   ```bash
   asc init --template=solo
   # Generates valid config
   ```

### Issue: Agent Configuration Not Applied

**Symptoms**: Changes to asc.toml don't take effect

**Causes**:
- Config not reloaded
- Agents already running with old config
- Syntax error preventing reload

**Solutions**:

1. **Restart agents**:
   ```bash
   asc down
   asc up
   ```

2. **Check for syntax errors**:
   ```bash
   asc check
   ```

3. **Verify hot-reload is working**:
   ```bash
   # Check logs for reload messages
   tail -f ~/.asc/logs/asc.log
   ```

4. **Force reload** (if implemented):
   ```bash
   asc reload
   ```

---

## Agent Problems

### Issue: Agents Not Starting

**Symptoms**: `asc up` completes but agents don't run

**Causes**:
- Agent command not found
- Missing dependencies
- Permission issues
- Port conflicts

**Solutions**:

1. **Check agent command exists**:
   ```bash
   which python  # or whatever command is configured
   ```

2. **Test agent command manually**:
   ```bash
   python agent_adapter.py
   # Should start without errors
   ```

3. **Check agent logs**:
   ```bash
   tail -f ~/.asc/logs/agent-name.log
   ```

4. **Verify environment variables**:
   ```bash
   # Agent should receive:
   # AGENT_NAME, AGENT_MODEL, AGENT_PHASES
   # MCP_MAIL_URL, BEADS_DB_PATH
   # API keys
   ```

5. **Check process status**:
   ```bash
   asc services status
   ps aux | grep agent
   ```

### Issue: Agent Stuck on Task

**Symptoms**: Agent shows "working" for extended period

**Causes**:
- Agent in infinite loop
- Waiting for external resource
- File lease not released
- LLM API timeout

**Solutions**:

1. **Check agent logs**:
   ```bash
   tail -f ~/.asc/logs/agent-name.log
   # Look for repeated errors or hangs
   ```

2. **Check file leases**:
   ```bash
   # Query MCP server for active leases
   curl http://localhost:8765/leases
   ```

3. **Restart stuck agent**:
   ```bash
   asc down
   asc up
   ```

4. **Check LLM API status**:
   ```bash
   # Verify API is responding
   curl https://api.anthropic.com/v1/messages \
     -H "x-api-key: $CLAUDE_API_KEY"
   ```

5. **Increase timeout** (if configurable):
   ```toml
   [agent.name]
   timeout = "30m"  # Increase from default
   ```

### Issue: Agent Crashes Repeatedly

**Symptoms**: Agent starts then immediately stops

**Causes**:
- Invalid API key
- Missing Python dependencies
- Code error in agent
- Resource exhaustion

**Solutions**:

1. **Check crash logs**:
   ```bash
   tail -100 ~/.asc/logs/agent-name.log
   # Look for Python tracebacks
   ```

2. **Test agent manually**:
   ```bash
   cd agent/
   python agent_adapter.py
   # See error directly
   ```

3. **Verify Python dependencies**:
   ```bash
   cd agent/
   pip install -r requirements.txt
   ```

4. **Check API key validity**:
   ```bash
   # Test Claude API
   curl https://api.anthropic.com/v1/messages \
     -H "x-api-key: $CLAUDE_API_KEY" \
     -H "content-type: application/json" \
     -d '{"model":"claude-3-sonnet-20240229","messages":[{"role":"user","content":"test"}],"max_tokens":10}'
   ```

5. **Check system resources**:
   ```bash
   free -h  # Memory
   df -h    # Disk space
   ```

### Issue: Agents Not Picking Up Tasks

**Symptoms**: Tasks exist but agents remain idle

**Causes**:
- Phase mismatch
- MCP connectivity issue
- Beads sync failure
- Agent filter misconfiguration

**Solutions**:

1. **Verify phase matching**:
   ```bash
   # Check agent phases in asc.toml
   [agent.name]
   phases = ["planning", "implementation"]
   
   # Check task phases in beads
   bd list --phase=planning
   ```

2. **Check MCP connectivity**:
   ```bash
   curl http://localhost:8765/status
   # Should return 200 OK
   ```

3. **Verify beads sync**:
   ```bash
   cd /path/to/beads/repo
   git pull
   bd list
   ```

4. **Check agent logs for errors**:
   ```bash
   tail -f ~/.asc/logs/agent-name.log
   # Look for connection errors
   ```

5. **Test end-to-end**:
   ```bash
   asc test
   # Should verify full stack
   ```

---

## TUI Display Issues

### Issue: TUI Display Corrupted or Garbled

**Symptoms**: TUI shows strange characters or broken layout

**Causes**:
- Terminal doesn't support required features
- Wrong TERM environment variable
- Terminal size too small
- Font issues

**Solutions**:

1. **Check TERM variable**:
   ```bash
   echo $TERM
   # Should be xterm-256color or similar
   ```

2. **Set correct TERM**:
   ```bash
   export TERM=xterm-256color
   asc up
   ```

3. **Use modern terminal**:
   - macOS: iTerm2, Alacritty
   - Linux: Alacritty, Kitty, GNOME Terminal
   - Windows: Windows Terminal

4. **Check terminal size**:
   ```bash
   tput cols  # Should be >= 80
   tput lines # Should be >= 24
   ```

5. **Resize terminal**:
   - Minimum: 80x24
   - Recommended: 120x40 or larger

6. **Try high contrast mode** (if available):
   ```bash
   asc up --theme=high-contrast
   ```

### Issue: Colors Not Showing

**Symptoms**: TUI is monochrome or colors are wrong

**Causes**:
- Terminal doesn't support colors
- Color scheme incompatibility
- TERM variable incorrect

**Solutions**:

1. **Test color support**:
   ```bash
   tput colors
   # Should return 256 or more
   ```

2. **Enable 256 colors**:
   ```bash
   export TERM=xterm-256color
   ```

3. **Test with simple color**:
   ```bash
   echo -e "\033[31mRed\033[0m"
   # Should show red text
   ```

4. **Use different theme**:
   ```bash
   asc up --theme=minimal
   ```

### Issue: TUI Not Updating

**Symptoms**: TUI shows stale data, doesn't refresh

**Causes**:
- Refresh interval too long
- Connection to MCP/beads lost
- TUI frozen or crashed

**Solutions**:

1. **Manual refresh**:
   ```bash
   # Press 'r' in TUI to force refresh
   ```

2. **Check refresh interval**:
   ```toml
   [core]
   refresh_interval = "2s"  # Decrease if too slow
   ```

3. **Check connectivity**:
   ```bash
   # In another terminal:
   curl http://localhost:8765/status
   ```

4. **Restart TUI**:
   ```bash
   # Press 'q' to quit
   asc up
   ```

5. **Check for errors**:
   ```bash
   tail -f ~/.asc/logs/asc.log
   ```

---

## Performance Issues

### Issue: High CPU Usage

**Symptoms**: asc process using excessive CPU

**Causes**:
- Polling intervals too aggressive
- Too many agents
- Rendering inefficiency
- Infinite loop in agent

**Solutions**:

1. **Increase polling intervals**:
   ```toml
   [core]
   refresh_interval = "5s"  # Increase from 2s
   beads_poll_interval = "10s"
   mcp_poll_interval = "5s"
   ```

2. **Use WebSocket instead of polling**:
   ```toml
   [services.mcp_agent_mail]
   use_websocket = true
   ```

3. **Reduce number of agents**:
   ```toml
   # Comment out unused agents
   # [agent.unused]
   # ...
   ```

4. **Check for runaway agents**:
   ```bash
   top -p $(pgrep -d',' -f agent)
   # Identify high CPU agents
   ```

5. **Profile the application**:
   ```bash
   # If built with profiling:
   go tool pprof http://localhost:6060/debug/pprof/profile
   ```

### Issue: High Memory Usage

**Symptoms**: asc using excessive memory

**Causes**:
- Log buffer too large
- Memory leak
- Too many cached tasks/messages
- Large agent playbooks

**Solutions**:

1. **Limit log buffer**:
   ```toml
   [core]
   max_log_entries = 1000  # Reduce from default
   ```

2. **Enable log rotation**:
   ```toml
   [logging]
   max_size = "10MB"
   max_files = 5
   ```

3. **Clear old logs**:
   ```bash
   rm ~/.asc/logs/*.log.old
   ```

4. **Restart periodically**:
   ```bash
   # Add to cron for long-running instances
   0 0 * * * asc down && asc up
   ```

5. **Monitor memory**:
   ```bash
   watch -n 5 'ps aux | grep asc'
   ```

### Issue: Slow Startup

**Symptoms**: `asc up` takes long time to start

**Causes**:
- Large beads repository
- Many agents to start
- Slow dependency checks
- Network latency

**Solutions**:

1. **Skip checks on startup**:
   ```bash
   asc up --skip-check
   ```

2. **Optimize beads repo**:
   ```bash
   cd /path/to/beads/repo
   git gc --aggressive
   ```

3. **Start agents in parallel** (if not already):
   ```toml
   [core]
   parallel_start = true
   ```

4. **Reduce startup tasks**:
   ```toml
   [core]
   skip_initial_sync = true
   ```

---

## Network and Connectivity

### Issue: MCP Server Won't Start

**Symptoms**: Error starting mcp_agent_mail

**Causes**:
- Port already in use
- Python not found
- Missing Python dependencies
- Permission issues

**Solutions**:

1. **Check if port is in use**:
   ```bash
   lsof -i :8765
   # or
   netstat -an | grep 8765
   ```

2. **Kill process using port**:
   ```bash
   kill $(lsof -t -i:8765)
   ```

3. **Change port in config**:
   ```toml
   [services.mcp_agent_mail]
   url = "http://localhost:9000"  # Use different port
   ```

4. **Check Python and dependencies**:
   ```bash
   which python3
   pip list | grep mcp
   ```

5. **Install MCP server**:
   ```bash
   pip install mcp-agent-mail
   ```

### Issue: Can't Connect to MCP Server

**Symptoms**: Agents can't reach MCP server

**Causes**:
- Server not running
- Wrong URL in config
- Firewall blocking connection
- Network issue

**Solutions**:

1. **Verify server is running**:
   ```bash
   asc services status
   curl http://localhost:8765/status
   ```

2. **Check URL in config**:
   ```toml
   [services.mcp_agent_mail]
   url = "http://localhost:8765"  # Verify correct
   ```

3. **Test connectivity**:
   ```bash
   curl -v http://localhost:8765/status
   # Should return 200 OK
   ```

4. **Check firewall**:
   ```bash
   # macOS
   sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate
   
   # Linux
   sudo ufw status
   ```

5. **Check logs**:
   ```bash
   tail -f ~/.asc/logs/mcp_agent_mail.log
   ```

### Issue: Beads Sync Failures

**Symptoms**: Can't sync with beads repository

**Causes**:
- Git conflicts
- Network issues
- Repository corruption
- Permission issues

**Solutions**:

1. **Check git status**:
   ```bash
   cd /path/to/beads/repo
   git status
   ```

2. **Resolve conflicts**:
   ```bash
   git pull --rebase
   # Resolve any conflicts
   git rebase --continue
   ```

3. **Reset if corrupted**:
   ```bash
   git fetch origin
   git reset --hard origin/main
   ```

4. **Check remote connectivity**:
   ```bash
   git remote -v
   git fetch --dry-run
   ```

5. **Verify permissions**:
   ```bash
   ls -la /path/to/beads/repo
   # Ensure you have write access
   ```

---

## File and Permission Issues

### Issue: Permission Denied Errors

**Symptoms**: Can't write to logs, PIDs, or config

**Causes**:
- Insufficient permissions
- Files owned by different user
- Directory doesn't exist
- SELinux/AppArmor restrictions

**Solutions**:

1. **Check file permissions**:
   ```bash
   ls -la ~/.asc/
   ls -la asc.toml .env
   ```

2. **Fix ownership**:
   ```bash
   sudo chown -R $USER:$USER ~/.asc/
   ```

3. **Fix permissions**:
   ```bash
   chmod 755 ~/.asc/
   chmod 755 ~/.asc/logs ~/.asc/pids
   chmod 600 .env
   chmod 644 asc.toml
   ```

4. **Create missing directories**:
   ```bash
   mkdir -p ~/.asc/logs ~/.asc/pids ~/.asc/playbooks
   ```

5. **Check SELinux** (Linux):
   ```bash
   getenforce
   # If enforcing, may need to adjust policies
   ```

### Issue: Logs Filling Disk

**Symptoms**: ~/.asc/logs directory very large

**Causes**:
- No log rotation
- Verbose logging
- Long-running agents
- Not cleaning old logs

**Solutions**:

1. **Enable log rotation**:
   ```toml
   [logging]
   max_size = "10MB"
   max_files = 5
   rotate = true
   ```

2. **Clean old logs**:
   ```bash
   find ~/.asc/logs -name "*.log" -mtime +7 -delete
   # Delete logs older than 7 days
   ```

3. **Reduce log verbosity**:
   ```toml
   [logging]
   level = "info"  # Change from "debug"
   ```

4. **Set up log rotation with logrotate**:
   ```bash
   # Create /etc/logrotate.d/asc
   /home/user/.asc/logs/*.log {
       daily
       rotate 7
       compress
       missingok
       notifempty
   }
   ```

5. **Monitor disk usage**:
   ```bash
   du -sh ~/.asc/logs
   ```

### Issue: PID Files Not Cleaned Up

**Symptoms**: Old PID files remain after shutdown

**Causes**:
- Unclean shutdown
- Process killed forcefully
- Bug in cleanup code

**Solutions**:

1. **Manual cleanup**:
   ```bash
   rm ~/.asc/pids/*.json
   ```

2. **Verify processes stopped**:
   ```bash
   ps aux | grep asc
   ps aux | grep agent
   # Kill any remaining processes
   ```

3. **Clean shutdown**:
   ```bash
   asc down
   # Wait for completion
   # Don't Ctrl+C or kill
   ```

4. **Add cleanup to shell exit**:
   ```bash
   # In ~/.bashrc or ~/.zshrc
   trap 'asc down' EXIT
   ```

---

## Getting More Help

If you've tried these solutions and still have issues:

1. **Check the logs**:
   ```bash
   tail -100 ~/.asc/logs/asc.log
   tail -100 ~/.asc/logs/agent-name.log
   ```

2. **Run with debug mode**:
   ```bash
   asc up --debug
   ```

3. **Check documentation**:
   - README.md
   - TROUBLESHOOTING.md
   - docs/

4. **Search existing issues**:
   - GitHub Issues
   - Discussion forums

5. **Report a bug**:
   - Include asc version: `asc --version`
   - Include OS and terminal info
   - Include relevant logs
   - Describe steps to reproduce

6. **Ask for help**:
   - GitHub Discussions
   - Community chat
   - Stack Overflow (tag: asc)

---

## Prevention Tips

### Best Practices

1. **Always run check before up**:
   ```bash
   asc check && asc up
   ```

2. **Keep backups of working configs**:
   ```bash
   cp asc.toml asc.toml.backup
   ```

3. **Monitor logs regularly**:
   ```bash
   tail -f ~/.asc/logs/asc.log
   ```

4. **Update regularly**:
   ```bash
   go install github.com/yourusername/asc@latest
   ```

5. **Use version control for configs**:
   ```bash
   git add asc.toml
   git commit -m "Update agent config"
   ```

6. **Test changes in isolation**:
   ```bash
   # Test new config before applying
   asc check
   ```

7. **Document your setup**:
   - Keep notes on customizations
   - Document agent purposes
   - Track configuration changes

### Maintenance Tasks

**Daily**:
- Check agent status
- Review error logs
- Monitor resource usage

**Weekly**:
- Clean old logs
- Update dependencies
- Review agent performance

**Monthly**:
- Update asc version
- Review and optimize config
- Backup configurations

---

## Quick Reference

### Common Commands

```bash
# Check system
asc check

# Start agents
asc up

# Stop agents
asc down

# View status
asc services status

# Test connectivity
asc test

# View logs
tail -f ~/.asc/logs/asc.log

# Clean up
rm ~/.asc/pids/*.json
rm ~/.asc/logs/*.log.old
```

### Common File Locations

```
~/.asc/                    # asc home directory
~/.asc/logs/               # Log files
~/.asc/pids/               # Process ID files
~/.asc/playbooks/          # Agent playbooks
./asc.toml                 # Configuration file
./.env                     # Environment variables
```

### Emergency Recovery

```bash
# Nuclear option: clean slate
asc down
rm -rf ~/.asc/
rm asc.toml .env
asc init
```

---

*Last updated: 2025-11-10*
*Version: 1.0*
