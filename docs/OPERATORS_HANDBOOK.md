# Operator's Handbook

Day-to-day operations guide for running and maintaining the Agent Stack Controller.

## Table of Contents

- [Daily Operations](#daily-operations)
- [Monitoring](#monitoring)
- [Maintenance](#maintenance)
- [Incident Response](#incident-response)
- [Performance Tuning](#performance-tuning)
- [Backup and Recovery](#backup-and-recovery)
- [Security Operations](#security-operations)
- [Troubleshooting Runbook](#troubleshooting-runbook)

---

## Daily Operations

### Starting Your Day

**Morning Checklist:**

1. **Start the stack**
   ```bash
   cd /path/to/project
   asc up
   ```

2. **Verify all agents are running**
   - Check Agent Status pane in TUI
   - All agents should show green ● (idle) or blue ⟳ (working)

3. **Check for overnight issues**
   ```bash
   # Review logs
   tail -100 ~/.asc/logs/asc.log
   
   # Check for errors
   grep ERROR ~/.asc/logs/*.log
   ```

4. **Review task queue**
   - Check Task Stream pane
   - Verify tasks are being picked up

5. **Monitor system health**
   ```bash
   asc doctor
   ```

### During the Day

**Periodic Checks (every 2-4 hours):**

- Monitor agent status in TUI
- Check for stuck agents (same task >30 min)
- Review error logs
- Verify task progress

**Quick Health Check:**
```bash
# From another terminal
asc test
```

### End of Day

**Shutdown Checklist:**

1. **Check for in-progress tasks**
   - Review Task Stream pane
   - Wait for critical tasks to complete

2. **Graceful shutdown**
   ```bash
   # In TUI: press 'q'
   # Or from terminal:
   asc down
   ```

3. **Verify clean shutdown**
   ```bash
   # Check no orphaned processes
   ps aux | grep agent_adapter
   ps aux | grep mcp_agent_mail
   ```

4. **Review day's logs**
   ```bash
   # Check for patterns or issues
   grep -i error ~/.asc/logs/asc.log | tail -50
   ```

---

## Monitoring

### Real-Time Monitoring

**TUI Dashboard:**

The TUI provides real-time monitoring:

```
┌─────────────────┬─────────────────────────────────┐
│ Agent Status    │ Task Stream                     │
│                 │                                 │
│ ● planner       │ #42 [open] Implement auth       │
│   Idle          │ #43 [in_progress] Fix bug       │
│                 │                                 │
│ ⟳ coder         │                                 │
│   Working #43   │                                 │
│                 │                                 │
│ ! tester        │                                 │
│   Error         │                                 │
├─────────────────┴─────────────────────────────────┤
│ MCP Interaction Log                               │
│ [10:30:15] [lease] coder → src/auth.go           │
└───────────────────────────────────────────────────┘
```

**Status Indicators:**

- ● Green: Agent idle, ready for work
- ⟳ Blue: Agent working on task
- ! Red: Agent error state
- ○ Gray: Agent offline

### Log Monitoring

**Watch logs in real-time:**

```bash
# All logs
tail -f ~/.asc/logs/*.log

# Specific agent
tail -f ~/.asc/logs/agent-name.log

# System log
tail -f ~/.asc/logs/asc.log

# Errors only
tail -f ~/.asc/logs/*.log | grep ERROR
```

**Log locations:**

```
~/.asc/logs/
├── asc.log              # System log
├── agent-name.log       # Per-agent logs
├── mcp_agent_mail.log   # MCP server log
└── health.log           # Health monitoring log
```

### Metrics to Monitor

**Agent Metrics:**

- **Uptime**: How long agents have been running
- **Task completion rate**: Tasks completed per hour
- **Error rate**: Errors per hour
- **Average task time**: Time to complete tasks

**System Metrics:**

- **Memory usage**: Per agent and total
- **CPU usage**: Per agent and total
- **Disk usage**: Log files and state
- **Network**: MCP server connections

**Check metrics:**

```bash
# Process stats
ps aux | grep agent_adapter

# Memory usage
ps aux | grep agent_adapter | awk '{sum+=$6} END {print sum/1024 " MB"}'

# Disk usage
du -sh ~/.asc/logs
```

### Alerting

**Set up alerts for:**

- Agent crashes (process exits)
- High error rates (>10 errors/hour)
- Stuck agents (same task >30 min)
- Disk space low (<1GB free)
- Memory usage high (>80%)

**Example monitoring script:**

```bash
#!/bin/bash
# monitor-asc.sh

# Check if agents are running
if ! pgrep -f agent_adapter > /dev/null; then
    echo "ALERT: No agents running!"
    # Send notification
fi

# Check error rate
ERROR_COUNT=$(grep ERROR ~/.asc/logs/asc.log | grep "$(date +%Y-%m-%d)" | wc -l)
if [ $ERROR_COUNT -gt 10 ]; then
    echo "ALERT: High error rate: $ERROR_COUNT errors today"
fi

# Check disk space
DISK_FREE=$(df -h ~/.asc | tail -1 | awk '{print $4}' | sed 's/G//')
if [ $(echo "$DISK_FREE < 1" | bc) -eq 1 ]; then
    echo "ALERT: Low disk space: ${DISK_FREE}GB free"
fi
```

---

## Maintenance

### Daily Maintenance

**Log Rotation:**

Logs are automatically rotated when they reach 10MB. Old logs are kept for 5 days.

**Manual rotation:**
```bash
# Rotate logs now
find ~/.asc/logs -name "*.log" -size +10M -exec mv {} {}.old \;
```

**Cleanup old logs:**
```bash
# Delete logs older than 7 days
find ~/.asc/logs -name "*.log.*" -mtime +7 -delete
```

### Weekly Maintenance

**1. Review and clean logs**
```bash
# Archive old logs
tar -czf logs-$(date +%Y%m%d).tar.gz ~/.asc/logs/*.log.*
mv logs-*.tar.gz ~/backups/

# Clean up
find ~/.asc/logs -name "*.log.*" -delete
```

**2. Update dependencies**
```bash
# Check for updates
asc check

# Update Python packages
pip install --upgrade mcp-agent-mail beads-cli

# Update asc
go install github.com/yourusername/asc@latest
```

**3. Backup configuration**
```bash
# Backup config and state
tar -czf asc-backup-$(date +%Y%m%d).tar.gz \
    asc.toml .env.age ~/.asc/

mv asc-backup-*.tar.gz ~/backups/
```

**4. Review agent performance**
```bash
# Check task completion times
grep "Task completed" ~/.asc/logs/*.log | \
    awk '{print $NF}' | \
    awk '{sum+=$1; count++} END {print "Avg:", sum/count, "seconds"}'
```

### Monthly Maintenance

**1. Rotate encryption keys**
```bash
asc secrets rotate
```

**2. Clean up playbooks**
```bash
# Review playbook sizes
du -sh ~/.asc/playbooks/*

# Archive old playbooks
tar -czf playbooks-$(date +%Y%m).tar.gz ~/.asc/playbooks/
```

**3. Update API keys**
```bash
# Decrypt secrets
asc secrets decrypt

# Edit .env with new keys
vim .env

# Re-encrypt
asc secrets encrypt
```

**4. Review and optimize configuration**
```bash
# Check for unused agents
# Review agent performance
# Adjust agent count if needed
```

### Quarterly Maintenance

**1. Major version upgrades**
```bash
# Check for new versions
asc --version

# Read upgrade guide
cat docs/UPGRADE_GUIDE.md

# Perform upgrade
# See UPGRADE_GUIDE.md
```

**2. Security audit**
```bash
# Check file permissions
ls -la .env .env.age ~/.asc/age.key

# Review access logs
grep "API key" ~/.asc/logs/*.log

# Rotate all secrets
asc secrets rotate
```

**3. Performance review**
```bash
# Analyze logs for patterns
# Review task completion rates
# Optimize agent configuration
```

---

## Incident Response

### Agent Crash

**Symptoms:**
- Agent shows ○ (offline) in TUI
- Process not in `ps aux` output
- Log shows exit message

**Response:**

1. **Check logs**
   ```bash
   tail -100 ~/.asc/logs/agent-name.log
   ```

2. **Identify cause**
   - API rate limit?
   - Out of memory?
   - Code error?

3. **Fix issue**
   - Wait for rate limit reset
   - Increase memory
   - Report bug

4. **Restart agent**
   ```bash
   # Configuration hot-reloads automatically
   # Or restart stack:
   asc down && asc up
   ```

### Stuck Agent

**Symptoms:**
- Agent shows ⟳ (working) for >30 minutes
- Same task ID in status
- No log activity

**Response:**

1. **Check what agent is doing**
   ```bash
   tail -50 ~/.asc/logs/agent-name.log
   ```

2. **Check if task is actually stuck**
   - Large task may take time
   - Check beads for task updates

3. **If truly stuck, restart agent**
   ```bash
   # In TUI: select agent, press 'k'
   # Or manually:
   kill <pid>
   ```

4. **Release file leases**
   ```bash
   # MCP server will auto-release after timeout
   # Or manually via MCP API
   ```

### MCP Server Down

**Symptoms:**
- Connection status shows red in TUI
- Agents can't communicate
- "Connection refused" in logs

**Response:**

1. **Check if server is running**
   ```bash
   ps aux | grep mcp_agent_mail
   ```

2. **Check server logs**
   ```bash
   tail -100 ~/.asc/logs/mcp_agent_mail.log
   ```

3. **Restart server**
   ```bash
   asc services restart
   ```

4. **Verify connectivity**
   ```bash
   curl http://localhost:8765/health
   ```

### Beads Database Issues

**Symptoms:**
- Tasks not showing in TUI
- "Git error" in logs
- Agents can't claim tasks

**Response:**

1. **Check beads repository**
   ```bash
   cd /path/to/beads/repo
   git status
   ```

2. **Fix git issues**
   ```bash
   # If merge conflict
   git merge --abort
   git pull --rebase
   
   # If corrupted
   git fsck
   ```

3. **Verify beads CLI**
   ```bash
   bd list
   ```

4. **Restart stack**
   ```bash
   asc down && asc up
   ```

### High Memory Usage

**Symptoms:**
- System slow
- OOM errors in logs
- Agents crashing

**Response:**

1. **Check memory usage**
   ```bash
   ps aux | grep agent_adapter | awk '{print $6, $11}'
   ```

2. **Identify memory hog**
   ```bash
   # Sort by memory
   ps aux | grep agent_adapter | sort -k6 -rn
   ```

3. **Restart high-memory agent**
   ```bash
   # In TUI: select agent, press 'k'
   ```

4. **Reduce agent count if needed**
   ```bash
   # Edit asc.toml, remove some agents
   # Hot-reload will stop them
   ```

### Disk Full

**Symptoms:**
- "No space left" errors
- Can't write logs
- Agents failing

**Response:**

1. **Check disk usage**
   ```bash
   df -h
   du -sh ~/.asc/logs
   ```

2. **Clean up logs**
   ```bash
   # Delete old logs
   find ~/.asc/logs -name "*.log.*" -delete
   
   # Truncate current logs
   truncate -s 0 ~/.asc/logs/*.log
   ```

3. **Archive if needed**
   ```bash
   tar -czf logs-archive.tar.gz ~/.asc/logs/
   mv logs-archive.tar.gz ~/backups/
   rm ~/.asc/logs/*.log
   ```

---

## Performance Tuning

### Agent Count Optimization

**Guidelines:**

- **1 agent**: Simple projects, learning
- **2-3 agents**: Most projects
- **4-6 agents**: Large projects
- **7+ agents**: High throughput needs

**Tuning:**

```toml
# Start conservative
[agent.main]
model = "claude"
phases = ["planning", "implementation", "testing"]

# Add more as needed
[agent.planner]
model = "gemini"
phases = ["planning"]

[agent.coder-1]
model = "claude"
phases = ["implementation"]

[agent.coder-2]
model = "claude"
phases = ["implementation"]
```

### Model Selection

**Performance characteristics:**

| Model | Speed | Quality | Cost | Best For |
|-------|-------|---------|------|----------|
| Gemini | Fast | Good | Low | Planning, docs |
| Claude | Medium | Excellent | Medium | Implementation |
| GPT-4 | Slow | Excellent | High | Testing, review |

**Optimization:**

```toml
# Fast planning
[agent.planner]
model = "gemini"
phases = ["planning"]

# Quality implementation
[agent.coder]
model = "claude"
phases = ["implementation"]

# Thorough testing
[agent.tester]
model = "gpt-4"
phases = ["testing"]
```

### Resource Limits

**Set limits per agent:**

```bash
# Memory limit (Linux)
systemd-run --scope -p MemoryMax=500M python agent_adapter.py

# CPU limit
systemd-run --scope -p CPUQuota=50% python agent_adapter.py
```

**Monitor resource usage:**

```bash
# Real-time monitoring
watch -n 1 'ps aux | grep agent_adapter'

# Memory usage
ps aux | grep agent_adapter | awk '{sum+=$6} END {print sum/1024 " MB"}'
```

### Polling Intervals

**Adjust in code if needed:**

- **Beads refresh**: 5s (default) - increase for less load
- **MCP polling**: 2s (default) - decrease for faster updates
- **Health checks**: 10s (default) - increase for less overhead

---

## Backup and Recovery

### What to Backup

**Critical:**
- `asc.toml` - Configuration
- `.env.age` - Encrypted secrets
- `~/.asc/age.key` - Encryption key
- `~/.asc/playbooks/` - Agent learning

**Important:**
- `~/.asc/templates/` - Custom templates
- `~/.asc/logs/` - Recent logs (for debugging)

**Not needed:**
- `~/.asc/pids/` - Regenerated on start
- `.env` - Can be decrypted from `.env.age`

### Backup Procedures

**Daily backup:**

```bash
#!/bin/bash
# backup-asc.sh

BACKUP_DIR=~/backups/asc
DATE=$(date +%Y%m%d)

mkdir -p $BACKUP_DIR

# Backup configuration
tar -czf $BACKUP_DIR/config-$DATE.tar.gz \
    asc.toml .env.age

# Backup state
tar -czf $BACKUP_DIR/state-$DATE.tar.gz \
    ~/.asc/playbooks/ \
    ~/.asc/templates/ \
    ~/.asc/age.key

# Keep only last 7 days
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete
```

**Run daily:**
```bash
# Add to crontab
crontab -e

# Add line:
0 2 * * * /path/to/backup-asc.sh
```

### Recovery Procedures

**Restore from backup:**

```bash
# Stop stack
asc down

# Restore configuration
tar -xzf config-20241109.tar.gz

# Restore state
tar -xzf state-20241109.tar.gz -C ~/

# Decrypt secrets
asc secrets decrypt

# Verify
asc check

# Restart
asc up
```

**Disaster recovery:**

If everything is lost:

1. **Reinstall asc**
   ```bash
   go install github.com/yourusername/asc@latest
   ```

2. **Restore from backup**
   ```bash
   # Restore files as above
   ```

3. **If no backup, start fresh**
   ```bash
   asc init
   # Reconfigure manually
   ```

---

## Security Operations

### Access Control

**File permissions:**

```bash
# Check permissions
ls -la .env .env.age ~/.asc/age.key

# Fix if needed
chmod 600 .env
chmod 600 ~/.asc/age.key
chmod 644 .env.age
```

**Process isolation:**

```bash
# Run as non-root user
# Never use sudo for asc
```

### Secret Rotation

**Rotate encryption key:**

```bash
# Monthly rotation
asc secrets rotate
```

**Rotate API keys:**

```bash
# 1. Generate new keys from providers
# 2. Update .env
asc secrets decrypt
vim .env
asc secrets encrypt

# 3. Test
asc test

# 4. Restart
asc down && asc up
```

### Audit Logging

**Review access:**

```bash
# Check who accessed secrets
grep "decrypt" ~/.asc/logs/asc.log

# Check API usage
grep "API" ~/.asc/logs/*.log
```

### Security Monitoring

**Daily checks:**

```bash
# Check for suspicious activity
grep -i "unauthorized\|failed\|denied" ~/.asc/logs/*.log

# Check file integrity
sha256sum asc.toml .env.age
```

---

## Troubleshooting Runbook

### Quick Diagnostics

```bash
# Run full diagnostics
asc doctor --verbose

# Check dependencies
asc check

# Test connectivity
asc test

# View logs
tail -100 ~/.asc/logs/asc.log
```

### Common Issues

See [Troubleshooting Guide](../TROUBLESHOOTING.md) for detailed solutions.

**Quick fixes:**

```bash
# Restart everything
asc down && asc up

# Clear state
rm -rf ~/.asc/pids/*

# Reset logs
truncate -s 0 ~/.asc/logs/*.log

# Decrypt secrets
asc secrets decrypt
```

---

## See Also

- [Troubleshooting Guide](../TROUBLESHOOTING.md)
- [Security Best Practices](security/SECURITY_BEST_PRACTICES.md)
- [Configuration Reference](CONFIGURATION.md)
- [API Reference](API_REFERENCE.md)
- [Upgrade Guide](UPGRADE_GUIDE.md)
