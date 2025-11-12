# Production Security Checklist

This document provides a comprehensive security checklist for deploying and operating the Agent Stack Controller (asc) in production environments.

## Pre-Deployment Security Checklist

### 1. File Permissions

**Critical Files:**
- [ ] `.env` file has permissions `0600` (rw-------)
  ```bash
  chmod 600 .env
  ls -la .env  # Should show: -rw-------
  ```

- [ ] `~/.asc/age.key` has permissions `0600` (rw-------)
  ```bash
  chmod 600 ~/.asc/age.key
  ls -la ~/.asc/age.key  # Should show: -rw-------
  ```

**Directories:**
- [ ] `~/.asc/logs/` has permissions `0700` (rwx------)
  ```bash
  chmod 700 ~/.asc/logs
  ls -ld ~/.asc/logs  # Should show: drwx------
  ```

- [ ] `~/.asc/pids/` has permissions `0700` (rwx------)
  ```bash
  chmod 700 ~/.asc/pids
  ls -ld ~/.asc/pids  # Should show: drwx------
  ```

- [ ] `~/.asc/playbooks/` has permissions `0700` (rwx------)
  ```bash
  chmod 700 ~/.asc/playbooks
  ls -ld ~/.asc/playbooks  # Should show: drwx------
  ```

**Configuration Files:**
- [ ] `asc.toml` has permissions `0644` (rw-r--r--)
  ```bash
  chmod 644 asc.toml
  ls -la asc.toml  # Should show: -rw-r--r--
  ```

### 2. Git Repository Security

**Verify .gitignore:**
- [ ] `.env` is in `.gitignore`
- [ ] `.env.*` (except `.env.example`) is in `.gitignore`
- [ ] `*.key` is in `.gitignore`
- [ ] `secrets/` and `.secrets/` are in `.gitignore`

**Remove Tracked Secrets:**
- [ ] Remove `.env` from git tracking if accidentally committed:
  ```bash
  git rm --cached .env
  git commit -m "Remove .env from tracking"
  ```

- [ ] Check git history for leaked secrets:
  ```bash
  git log --all --full-history -- .env
  git log -p --all -S "sk-ant-" -S "sk-" -S "AIza"
  ```

- [ ] If secrets were committed, rotate all API keys immediately

**Verify Encrypted Files:**
- [ ] `.env.age` is committed (encrypted secrets)
- [ ] Age key (`~/.asc/age.key`) is NOT committed
- [ ] Age key is backed up securely offline

### 3. API Key Security

**Key Validation:**
- [ ] All API keys are valid and active
- [ ] API keys have appropriate rate limits configured
- [ ] API keys are scoped to minimum required permissions

**Key Rotation:**
- [ ] Document key rotation schedule (recommended: every 90 days)
- [ ] Test key rotation procedure
- [ ] Have backup keys ready for zero-downtime rotation

**Key Storage:**
- [ ] API keys are stored in `.env` file only
- [ ] API keys are encrypted in `.env.age` for version control
- [ ] API keys are never logged or displayed in TUI
- [ ] API keys are never passed as command-line arguments

### 4. Process Security

**Process Isolation:**
- [ ] Agents run with same user permissions as asc (no privilege escalation)
- [ ] Process groups are properly configured for cleanup
- [ ] PID files are stored in secure directory (`~/.asc/pids/`)

**Resource Limits:**
- [ ] Configure ulimits for agent processes
  ```bash
  ulimit -n 1024  # Max open files
  ulimit -u 512   # Max user processes
  ```

- [ ] Monitor agent resource usage
- [ ] Set up alerts for excessive resource consumption

### 5. Network Security

**MCP Server:**
- [ ] MCP server binds to localhost only (not 0.0.0.0)
- [ ] MCP server uses authentication if exposed
- [ ] MCP server uses TLS if exposed over network
- [ ] Firewall rules restrict MCP server access

**External APIs:**
- [ ] All API calls use HTTPS
- [ ] Certificate validation is enabled
- [ ] Timeout values are configured
- [ ] Retry logic has exponential backoff

### 6. Logging Security

**Log Content:**
- [ ] API keys are never logged
- [ ] Sensitive data is redacted from logs
- [ ] User data is anonymized in logs
- [ ] Log level is appropriate for environment (INFO in production)

**Log Storage:**
- [ ] Log files have secure permissions (0600)
- [ ] Log directory has secure permissions (0700)
- [ ] Log rotation is configured (max 10MB, keep 5 files)
- [ ] Old logs are securely deleted

**Log Monitoring:**
- [ ] Set up alerts for ERROR level logs
- [ ] Monitor for suspicious patterns (repeated failures, unusual API usage)
- [ ] Regularly review logs for security incidents

### 7. Dependency Security

**Dependency Scanning:**
- [ ] Run `go list -m all` to list all dependencies
- [ ] Check for known vulnerabilities with `govulncheck`
  ```bash
  go install golang.org/x/vuln/cmd/govulncheck@latest
  govulncheck ./...
  ```

- [ ] Review Python agent dependencies
  ```bash
  pip list --outdated
  safety check -r agent/requirements.txt
  ```

**Dependency Updates:**
- [ ] Keep Go dependencies up to date
- [ ] Keep Python dependencies up to date
- [ ] Test updates in staging before production
- [ ] Subscribe to security advisories for key dependencies

### 8. Configuration Security

**Configuration Validation:**
- [ ] Run `asc check` to validate configuration
- [ ] Run `asc doctor` to check for security issues
- [ ] Review agent commands for injection vulnerabilities
- [ ] Validate all file paths in configuration

**Configuration Backup:**
- [ ] Backup `asc.toml` configuration
- [ ] Backup `.env.age` encrypted secrets
- [ ] Store backups securely offline
- [ ] Test restore procedure

### 9. Access Control

**User Permissions:**
- [ ] asc runs as non-root user
- [ ] User has minimum required permissions
- [ ] Beads repository access is restricted
- [ ] Playbook directory is user-only

**Multi-User Environments:**
- [ ] Each user has separate `~/.asc/` directory
- [ ] Users cannot access each other's logs or PIDs
- [ ] Shared beads repository has appropriate permissions
- [ ] Consider using separate API keys per user

## Runtime Security Checklist

### 1. Startup Checks

Before starting the agent stack:
- [ ] Run `asc check` to verify dependencies
- [ ] Run `asc doctor` to check for issues
- [ ] Verify file permissions with `scripts/check-security.sh`
- [ ] Check available disk space for logs
- [ ] Verify network connectivity to required services

### 2. Monitoring

During operation:
- [ ] Monitor agent health in TUI dashboard
- [ ] Watch for repeated errors in logs
- [ ] Monitor API rate limit usage
- [ ] Check for stuck or crashed agents
- [ ] Monitor disk usage for logs and PIDs

### 3. Incident Response

If security incident detected:
- [ ] Stop all agents immediately: `asc down`
- [ ] Preserve logs for forensics:
  ```bash
  cp -r ~/.asc/logs ~/incident-logs-$(date +%Y%m%d-%H%M%S)
  cp -r ~/.asc/pids ~/incident-pids-$(date +%Y%m%d-%H%M%S)
  ```
- [ ] Rotate all API keys
- [ ] Review git history for leaked secrets
- [ ] Analyze logs for unauthorized access
- [ ] Document incident and remediation steps

## Post-Deployment Security Checklist

### 1. Regular Maintenance

**Weekly:**
- [ ] Review error logs for security issues
- [ ] Check disk usage for logs and PIDs
- [ ] Verify agent health and performance
- [ ] Monitor API usage and costs

**Monthly:**
- [ ] Run security scan: `scripts/check-security.sh`
- [ ] Update dependencies
- [ ] Review and rotate API keys if needed
- [ ] Test backup and restore procedures

**Quarterly:**
- [ ] Conduct security audit
- [ ] Review access controls
- [ ] Update security documentation
- [ ] Train team on security best practices

### 2. Cleanup

**Regular Cleanup:**
- [ ] Remove old log files:
  ```bash
  find ~/.asc/logs -name "*.log" -mtime +30 -delete
  ```

- [ ] Clean up orphaned PID files:
  ```bash
  asc cleanup
  ```

- [ ] Archive old playbooks if needed

**Decommissioning:**
- [ ] Stop all agents: `asc down`
- [ ] Revoke all API keys
- [ ] Securely delete all logs and PIDs:
  ```bash
  rm -rf ~/.asc/logs/*
  rm -rf ~/.asc/pids/*
  ```
- [ ] Remove age encryption key:
  ```bash
  shred -u ~/.asc/age.key  # Linux
  rm -P ~/.asc/age.key     # macOS
  ```

## Security Automation

### Automated Security Checks

Add to CI/CD pipeline:

```bash
# Run security checks
./scripts/check-security.sh

# Run vulnerability scanning
govulncheck ./...

# Run static analysis
golangci-lint run

# Run security-focused tests
go test -v ./test/security_test.go
```

### Pre-commit Hook

Install security checks in git pre-commit hook:

```bash
#!/bin/bash
# .githooks/pre-commit

# Check for secrets in staged files
if git diff --cached --name-only | grep -E '\.(go|py|sh|toml)$' | xargs grep -E 'sk-ant-|sk-[a-zA-Z0-9-]{20,}|AIza[a-zA-Z0-9_-]{35}'; then
    echo "ERROR: Potential API key found in staged files"
    exit 1
fi

# Check for .env file
if git diff --cached --name-only | grep -E '^\.env$'; then
    echo "ERROR: Attempting to commit .env file"
    exit 1
fi

# Check file permissions
if [ -f .env ] && [ "$(stat -f '%Lp' .env 2>/dev/null || stat -c '%a' .env 2>/dev/null)" != "600" ]; then
    echo "WARNING: .env file has insecure permissions"
    chmod 600 .env
fi

exit 0
```

## Security Resources

### Documentation
- [Security Best Practices](./SECURITY_BEST_PRACTICES.md)
- [Incident Response Plan](./INCIDENT_RESPONSE_PLAN.md)
- [Age Encryption Guide](../adr/ADR-0011-age-encryption.md)

### Tools
- `asc check` - Dependency verification
- `asc doctor` - Health and security diagnostics
- `scripts/check-security.sh` - Comprehensive security scan
- `govulncheck` - Go vulnerability scanner
- `safety` - Python dependency security checker

### External Resources
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)

## Compliance Considerations

### Data Protection
- [ ] Understand what data agents process
- [ ] Ensure compliance with GDPR/CCPA if applicable
- [ ] Document data retention policies
- [ ] Implement data deletion procedures

### Audit Trail
- [ ] Enable structured logging for audit trail
- [ ] Retain logs for required compliance period
- [ ] Implement log integrity verification
- [ ] Document incident response procedures

### Third-Party Services
- [ ] Review LLM provider terms of service
- [ ] Understand data processing agreements
- [ ] Document third-party dependencies
- [ ] Maintain vendor security assessments

## Quick Reference

### Fix Common Security Issues

```bash
# Fix .env permissions
chmod 600 .env

# Fix directory permissions
chmod 700 ~/.asc/logs ~/.asc/pids ~/.asc/playbooks

# Remove .env from git
git rm --cached .env
git commit -m "Remove .env from tracking"

# Rotate API keys
# 1. Generate new keys from provider dashboards
# 2. Update .env file
# 3. Encrypt: asc secrets encrypt
# 4. Restart agents: asc down && asc up

# Run security scan
./scripts/check-security.sh

# Check for vulnerabilities
govulncheck ./...
```

### Emergency Procedures

```bash
# Emergency shutdown
asc down
pkill -f "python agent_adapter.py"
pkill -f "mcp_agent_mail"

# Secure sensitive files
chmod 600 .env ~/.asc/age.key
chmod 700 ~/.asc/logs ~/.asc/pids

# Preserve evidence
tar -czf incident-$(date +%Y%m%d-%H%M%S).tar.gz \
    ~/.asc/logs ~/.asc/pids asc.toml .env.age

# Rotate all keys immediately
# (Visit provider dashboards to generate new keys)
```

## Conclusion

Security is an ongoing process, not a one-time checklist. Regularly review and update your security practices, stay informed about new vulnerabilities, and maintain a security-first mindset when operating the Agent Stack Controller.

For questions or to report security issues, see [SECURITY.md](../../SECURITY.md).
