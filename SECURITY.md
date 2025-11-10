# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of asc seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Please Do Not

- **Do not** open a public GitHub issue for security vulnerabilities
- **Do not** disclose the vulnerability publicly until it has been addressed

### Please Do

1. **Email us** at security@yourdomain.com with:
   - A description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact
   - Any suggested fixes (if you have them)

2. **Allow time** for us to respond and address the issue before public disclosure

3. **Provide your contact information** so we can follow up with you

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Updates**: We will provide regular updates on our progress
- **Timeline**: We aim to address critical vulnerabilities within 7 days
- **Credit**: We will credit you in the security advisory (unless you prefer to remain anonymous)

## Security Best Practices

### For Users

1. **Keep asc updated** - Always use the latest version
2. **Protect your API keys** - Never commit `.env` files to git
3. **Use encrypted secrets** - Use `asc secrets encrypt` for team sharing
4. **Review permissions** - Ensure `.env` has 0600 permissions
5. **Audit access** - Regularly review who has access to your secrets
6. **Rotate keys** - Rotate API keys and encryption keys regularly
7. **Monitor logs** - Check logs for suspicious activity

### For Developers

1. **Never commit secrets** - Use `.env` files (gitignored)
2. **Validate input** - Always validate and sanitize user input
3. **Use parameterized queries** - Prevent SQL injection
4. **Escape shell commands** - Prevent command injection
5. **Check file paths** - Prevent path traversal attacks
6. **Set proper permissions** - Use restrictive file permissions
7. **Review dependencies** - Keep dependencies updated and secure
8. **Run security scans** - Use `gosec` and `govulncheck`

## Security Features

### Secrets Management

asc uses age encryption for secure secrets management:

- **Encryption at rest** - API keys encrypted with age
- **Automatic gitignore** - `.env` files automatically ignored
- **Restrictive permissions** - Files set to 0600 automatically
- **Key rotation** - Easy key rotation with `asc secrets rotate`

### Process Isolation

- **User-level permissions** - Agents run with same permissions as user
- **No privilege escalation** - Never requires root/admin
- **Process cleanup** - Proper cleanup of child processes
- **Resource limits** - Configurable resource limits

### Input Validation

- **Configuration validation** - TOML syntax and semantic validation
- **Path validation** - File paths checked for traversal attacks
- **Command validation** - Shell commands properly escaped
- **API input validation** - All API inputs validated

## Known Security Considerations

### API Keys

- API keys are stored in `.env` files (plaintext when decrypted)
- Use `asc secrets encrypt` to encrypt for version control
- Keys are passed to agents via environment variables
- Never logged or displayed in output

### File System Access

- asc requires read/write access to:
  - Configuration files (`asc.toml`, `.env`)
  - Log directory (`~/.asc/logs/`)
  - PID directory (`~/.asc/pids/`)
  - Beads database (configured path)
- Agents have same file system access as the user running asc

### Network Access

- MCP server listens on localhost by default (not exposed externally)
- Agents make outbound connections to LLM APIs
- No inbound network connections required

### Process Management

- asc spawns child processes (agents, MCP server)
- Child processes inherit user permissions
- Processes are tracked via PID files
- Graceful shutdown with SIGTERM, forced with SIGKILL

## Security Checklist for Contributors

When contributing code, ensure:

- [ ] No hardcoded secrets or credentials
- [ ] Input validation for all user input
- [ ] Proper error handling (no information leakage)
- [ ] No SQL injection vulnerabilities
- [ ] No command injection vulnerabilities
- [ ] No path traversal vulnerabilities
- [ ] Proper file permissions set
- [ ] Dependencies are up to date
- [ ] Security tests added for security-sensitive code
- [ ] Documentation updated for security-related changes

## Security Scanning

We use automated security scanning:

- **gosec** - Go security checker (runs in CI)
- **govulncheck** - Go vulnerability checker (runs in CI)
- **Dependabot** - Automated dependency updates
- **CodeQL** - Semantic code analysis (GitHub)

Run security scans locally:

```bash
# Install tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run scans
gosec ./...
govulncheck ./...
```

## Vulnerability Disclosure Timeline

1. **Day 0**: Vulnerability reported
2. **Day 1-2**: Acknowledgment sent, investigation begins
3. **Day 3-7**: Fix developed and tested
4. **Day 7-14**: Fix released, security advisory published
5. **Day 14+**: Public disclosure (if not already public)

## Security Updates

Security updates are released as:

- **Patch releases** (x.x.X) for minor vulnerabilities
- **Minor releases** (x.X.x) for moderate vulnerabilities
- **Immediate hotfixes** for critical vulnerabilities

Subscribe to security advisories:
- Watch the repository for security advisories
- Follow releases on GitHub
- Check the CHANGELOG for security fixes

## Contact

- **Security issues**: security@yourdomain.com
- **General questions**: GitHub Discussions
- **Bug reports**: GitHub Issues (non-security only)

## Acknowledgments

We thank the following security researchers for responsibly disclosing vulnerabilities:

<!-- List will be updated as vulnerabilities are reported and fixed -->

---

Thank you for helping keep asc and its users safe!
