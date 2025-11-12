# Security Best Practices

This document outlines security best practices for developing, deploying, and using the Agent Stack Controller (asc).

> **For Production Deployments:** See the comprehensive [Production Security Checklist](./PRODUCTION_SECURITY_CHECKLIST.md) for a complete pre-deployment and operational security guide.

## Table of Contents

- [For Users](#for-users)
- [For Developers](#for-developers)
- [API Key Management](#api-key-management)
- [File Permissions](#file-permissions)
- [Input Validation](#input-validation)
- [Process Security](#process-security)
- [Network Security](#network-security)
- [Logging and Monitoring](#logging-and-monitoring)
- [Incident Response](#incident-response)

## For Users

### 1. Protect Your API Keys

**Never commit `.env` files to version control:**

```bash
# Ensure .env is in .gitignore
echo ".env" >> .gitignore
```

**Use encrypted secrets for team sharing:**

```bash
# Initialize secrets management
asc secrets init

# Encrypt your .env file
asc secrets encrypt

# Commit the encrypted file
git add .env.age
git commit -m "Add encrypted secrets"

# Team members can decrypt
asc secrets decrypt
```

**Verify file permissions:**

```bash
# .env should be readable only by you
chmod 600 .env

# Verify permissions
ls -la .env
# Should show: -rw------- (600)
```

### 2. Keep asc Updated

**Check for updates regularly:**

```bash
# Check current version
asc --version

# Update to latest version
go install github.com/yourusername/asc@latest
```

**Subscribe to security advisories:**
- Watch the GitHub repository for security advisories
- Enable notifications for releases
- Review CHANGELOG.md for security fixes

### 3. Audit Access

**Review who has access to your secrets:**

```bash
# Check file permissions
ls -la .env .env.age ~/.asc/age.key

# Review git history for accidental commits
git log --all --full-history -- .env
```

**Rotate keys regularly:**

```bash
# Rotate encryption key
asc secrets rotate

# Update API keys in .env
# Then re-encrypt
asc secrets encrypt
```

### 4. Monitor Logs

**Check logs for suspicious activity:**

```bash
# View asc logs
tail -f ~/.asc/logs/asc.log

# Check for failed authentication attempts
grep "authentication failed" ~/.asc/logs/*.log

# Check for unusual file access
grep "permission denied" ~/.asc/logs/*.log
```

### 5. Secure Your Environment

**Use a secure development environment:**
- Keep your OS and tools updated
- Use full disk encryption
- Enable firewall
- Use strong passwords/passphrases
- Enable 2FA on all accounts

**Isolate development environments:**
- Use separate API keys for dev/staging/prod
- Don't use production keys in development
- Use separate beads databases for each environment

## For Developers

### 1. Never Commit Secrets

**Use environment variables:**

```go
// Good: Load from environment
apiKey := os.Getenv("CLAUDE_API_KEY")

// Bad: Hardcoded secret
apiKey := "sk-ant-123456789" // NEVER DO THIS
```

**Check for secrets before committing:**

```bash
# Install pre-commit hook
cp .githooks/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit

# Manually check for secrets
git diff --cached | grep -i "api_key\|secret\|password"
```

### 2. Validate All Input

**Validate configuration files:**

```go
func LoadConfig(path string) (*Config, error) {
    // Validate path
    if !isValidPath(path) {
        return nil, fmt.Errorf("invalid config path: %s", path)
    }
    
    // Clean path to prevent traversal
    cleanPath := filepath.Clean(path)
    
    // Ensure path is within allowed directory
    if !strings.HasPrefix(cleanPath, allowedDir) {
        return nil, fmt.Errorf("config path outside allowed directory")
    }
    
    // Load and validate
    cfg, err := viper.ReadInConfig()
    if err != nil {
        return nil, err
    }
    
    // Validate required fields
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return cfg, nil
}
```

**Sanitize user input:**

```go
func ValidateAgentName(name string) error {
    // Only allow alphanumeric, dash, underscore
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
    if !matched {
        return fmt.Errorf("invalid agent name: %s", name)
    }
    
    // Limit length
    if len(name) > 64 {
        return fmt.Errorf("agent name too long")
    }
    
    return nil
}
```

### 3. Prevent Command Injection

**Use exec.Command properly:**

```go
// Good: Arguments are properly separated
cmd := exec.Command("python", "agent.py", "--name", agentName)

// Bad: Shell execution allows injection
cmd := exec.Command("sh", "-c", fmt.Sprintf("python agent.py --name %s", agentName))
```

**Never use shell execution with user input:**

```go
// NEVER DO THIS
func runCommand(userInput string) error {
    cmd := exec.Command("sh", "-c", userInput)
    return cmd.Run()
}

// DO THIS INSTEAD
func runCommand(program string, args []string) error {
    // Validate program
    if !isValidProgram(program) {
        return fmt.Errorf("invalid program")
    }
    
    // Validate args
    for _, arg := range args {
        if !isValidArg(arg) {
            return fmt.Errorf("invalid argument")
        }
    }
    
    cmd := exec.Command(program, args...)
    return cmd.Run()
}
```

### 4. Prevent Path Traversal

**Validate file paths:**

```go
func SafeFilePath(baseDir, userPath string) (string, error) {
    // Clean the path
    cleanPath := filepath.Clean(userPath)
    
    // Join with base directory
    fullPath := filepath.Join(baseDir, cleanPath)
    
    // Resolve symlinks
    realPath, err := filepath.EvalSymlinks(fullPath)
    if err != nil {
        return "", err
    }
    
    // Ensure path is within base directory
    if !strings.HasPrefix(realPath, baseDir) {
        return "", fmt.Errorf("path traversal detected")
    }
    
    return realPath, nil
}
```

**Check for symlink attacks:**

```go
func SafeReadFile(path string) ([]byte, error) {
    // Get file info without following symlinks
    info, err := os.Lstat(path)
    if err != nil {
        return nil, err
    }
    
    // Reject symlinks
    if info.Mode()&os.ModeSymlink != 0 {
        return nil, fmt.Errorf("symlinks not allowed")
    }
    
    // Read file
    return os.ReadFile(path)
}
```

### 5. Set Proper File Permissions

**Create files with restrictive permissions:**

```go
// Sensitive files (secrets, keys)
err := os.WriteFile(path, data, 0600) // rw-------

// Directories
err := os.MkdirAll(path, 0700) // rwx------

// Regular files
err := os.WriteFile(path, data, 0644) // rw-r--r--
```

**Check and fix permissions:**

```go
func EnsureSecurePermissions(path string) error {
    info, err := os.Stat(path)
    if err != nil {
        return err
    }
    
    // Check if world-readable
    if info.Mode().Perm()&0004 != 0 {
        // Fix permissions
        if err := os.Chmod(path, 0600); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 6. Handle Errors Securely

**Don't leak sensitive information in errors:**

```go
// Bad: Leaks API key
return fmt.Errorf("authentication failed with key: %s", apiKey)

// Good: Generic error message
return fmt.Errorf("authentication failed")

// Good: Log details securely, return generic error
logger.Error("authentication failed", "key_prefix", apiKey[:8])
return fmt.Errorf("authentication failed")
```

**Validate error messages before displaying:**

```go
func SafeErrorMessage(err error) string {
    msg := err.Error()
    
    // Remove any API keys
    msg = regexp.MustCompile(`sk-[a-zA-Z0-9]+`).ReplaceAllString(msg, "[REDACTED]")
    
    // Remove file paths
    msg = regexp.MustCompile(`/[a-zA-Z0-9/_-]+`).ReplaceAllString(msg, "[PATH]")
    
    return msg
}
```

## API Key Management

### Storage

**Store API keys in `.env` file:**

```bash
# .env
CLAUDE_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=...
```

**Encrypt for version control:**

```bash
# Encrypt
asc secrets encrypt

# Commit encrypted version
git add .env.age
git commit -m "Add encrypted secrets"
```

### Rotation

**Rotate API keys regularly:**

1. Generate new API keys from provider dashboards
2. Update `.env` file with new keys
3. Test with new keys
4. Revoke old keys
5. Re-encrypt: `asc secrets encrypt`

**Rotate encryption keys:**

```bash
# Rotate age key and re-encrypt all files
asc secrets rotate
```

### Access Control

**Limit who can access keys:**

- Use separate keys for each team member
- Use separate keys for each environment
- Revoke keys when team members leave
- Audit key usage regularly

## File Permissions

### Sensitive Files

| File/Directory | Permissions | Description |
|---------------|-------------|-------------|
| `.env` | 600 (rw-------) | API keys and secrets |
| `~/.asc/age.key` | 600 (rw-------) | Encryption key |
| `~/.asc/logs/` | 700 (rwx------) | Log directory |
| `~/.asc/pids/` | 700 (rwx------) | PID directory |
| `asc.toml` | 644 (rw-r--r--) | Configuration file |

### Checking Permissions

```bash
# Check file permissions
ls -la .env ~/.asc/age.key

# Fix permissions if needed
chmod 600 .env
chmod 600 ~/.asc/age.key
chmod 700 ~/.asc/logs
chmod 700 ~/.asc/pids
```

### Automated Checks

```bash
# Add to pre-commit hook
if [ -f .env ]; then
    perms=$(stat -f "%Lp" .env 2>/dev/null || stat -c "%a" .env 2>/dev/null)
    if [ "$perms" != "600" ]; then
        echo "Error: .env has insecure permissions: $perms"
        echo "Fix with: chmod 600 .env"
        exit 1
    fi
fi
```

## Input Validation

### Configuration Files

**Validate TOML syntax:**

```go
func ValidateConfig(cfg *Config) error {
    // Check required fields
    if cfg.Core.BeadsDBPath == "" {
        return fmt.Errorf("beads_db_path is required")
    }
    
    // Validate paths
    if !isValidPath(cfg.Core.BeadsDBPath) {
        return fmt.Errorf("invalid beads_db_path")
    }
    
    // Validate agents
    for name, agent := range cfg.Agents {
        if err := ValidateAgent(name, agent); err != nil {
            return err
        }
    }
    
    return nil
}
```

### User Input

**Validate all user input:**

```go
func ValidateUserInput(input string) error {
    // Check length
    if len(input) > maxLength {
        return fmt.Errorf("input too long")
    }
    
    // Check for dangerous characters
    if containsDangerousChars(input) {
        return fmt.Errorf("input contains invalid characters")
    }
    
    // Check format
    if !matchesExpectedFormat(input) {
        return fmt.Errorf("input format invalid")
    }
    
    return nil
}
```

### Environment Variables

**Validate environment variables:**

```go
func ValidateEnvVar(key, value string) error {
    // Check key format
    if !regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`).MatchString(key) {
        return fmt.Errorf("invalid env var key: %s", key)
    }
    
    // Check for null bytes
    if strings.Contains(value, "\x00") {
        return fmt.Errorf("env var contains null byte")
    }
    
    // Check for newlines (can break env format)
    if strings.ContainsAny(value, "\n\r") {
        return fmt.Errorf("env var contains newline")
    }
    
    return nil
}
```

## Process Security

### Process Isolation

**Run agents with minimal privileges:**

```go
func StartAgent(name string, cmd string) error {
    // Don't elevate privileges
    command := exec.Command(cmd)
    
    // Set process group for cleanup
    command.SysProcAttr = &syscall.SysProcAttr{
        Setpgid: true,
    }
    
    // Limit resources
    command.SysProcAttr.Rlimit = []syscall.Rlimit{
        {Cur: maxMemory, Max: maxMemory}, // Memory limit
    }
    
    return command.Start()
}
```

### Process Cleanup

**Clean up child processes:**

```go
func StopAgent(pid int) error {
    // Send SIGTERM for graceful shutdown
    if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
        return err
    }
    
    // Wait for shutdown
    done := make(chan bool)
    go func() {
        syscall.Wait4(pid, nil, 0, nil)
        done <- true
    }()
    
    // Timeout after 5 seconds
    select {
    case <-done:
        return nil
    case <-time.After(5 * time.Second):
        // Force kill
        return syscall.Kill(-pid, syscall.SIGKILL)
    }
}
```

## Network Security

### MCP Server

**Bind to localhost only:**

```go
// Good: Only accessible locally
listener, err := net.Listen("tcp", "127.0.0.1:8765")

// Bad: Accessible from network
listener, err := net.Listen("tcp", ":8765")
```

**Use authentication:**

```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Check authentication token
    token := r.Header.Get("Authorization")
    if !isValidToken(token) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Handle request
    // ...
}
```

### API Calls

**Use HTTPS for API calls:**

```go
// Configure HTTP client with TLS
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
    },
}
```

**Validate certificates:**

```go
// Don't skip certificate verification
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: false, // Always verify
        },
    },
}
```

## Logging and Monitoring

### Secure Logging

**Never log sensitive data:**

```go
// Bad: Logs API key
logger.Info("Starting agent", "api_key", apiKey)

// Good: Logs only prefix
logger.Info("Starting agent", "api_key_prefix", apiKey[:8])

// Good: Doesn't log key at all
logger.Info("Starting agent", "agent_name", name)
```

**Sanitize log messages:**

```go
func SanitizeLogMessage(msg string) string {
    // Remove API keys
    msg = regexp.MustCompile(`sk-[a-zA-Z0-9]+`).ReplaceAllString(msg, "[REDACTED]")
    
    // Remove passwords
    msg = regexp.MustCompile(`password[=:]\s*\S+`).ReplaceAllString(msg, "password=[REDACTED]")
    
    // Remove tokens
    msg = regexp.MustCompile(`token[=:]\s*\S+`).ReplaceAllString(msg, "token=[REDACTED]")
    
    return msg
}
```

### Monitoring

**Monitor for security events:**

```go
// Failed authentication attempts
if authFailed {
    logger.Warn("Authentication failed",
        "agent", agentName,
        "attempts", failedAttempts,
        "source_ip", sourceIP)
}

// Unusual file access
if accessDenied {
    logger.Warn("File access denied",
        "path", path,
        "user", user,
        "operation", operation)
}

// Suspicious activity
if suspiciousActivity {
    logger.Error("Suspicious activity detected",
        "type", activityType,
        "details", details)
}
```

## Incident Response

### Detection

**Monitor for security incidents:**

1. Failed authentication attempts
2. Unusual file access patterns
3. Unexpected process terminations
4. Network connection failures
5. Configuration changes
6. API rate limit errors

### Response Plan

**When a security incident is detected:**

1. **Contain**
   - Stop affected agents: `asc down`
   - Revoke compromised API keys
   - Isolate affected systems

2. **Investigate**
   - Review logs: `~/.asc/logs/`
   - Check file modifications
   - Review process history
   - Identify scope of compromise

3. **Remediate**
   - Rotate all API keys
   - Rotate encryption keys: `asc secrets rotate`
   - Update passwords
   - Patch vulnerabilities
   - Update asc to latest version

4. **Recover**
   - Restore from clean backup
   - Verify system integrity
   - Restart services: `asc up`
   - Monitor for recurrence

5. **Document**
   - Document incident timeline
   - Record actions taken
   - Identify root cause
   - Update security procedures

### Reporting

**Report security incidents:**

1. Email: security@yourdomain.com
2. Include:
   - Description of incident
   - Timeline of events
   - Systems affected
   - Actions taken
   - Logs and evidence

## Security Checklist

### Before Deployment

- [ ] All secrets encrypted
- [ ] `.env` in `.gitignore`
- [ ] File permissions set correctly
- [ ] Security scans passed (gosec, govulncheck)
- [ ] Dependencies updated
- [ ] Security tests passed
- [ ] Code review completed
- [ ] Documentation updated

### Regular Maintenance

- [ ] Update asc monthly
- [ ] Rotate API keys quarterly
- [ ] Review access logs monthly
- [ ] Update dependencies weekly
- [ ] Run security scans weekly
- [ ] Review file permissions monthly
- [ ] Audit team access quarterly

### After Incident

- [ ] Incident documented
- [ ] Root cause identified
- [ ] Vulnerabilities patched
- [ ] Keys rotated
- [ ] Systems verified clean
- [ ] Monitoring enhanced
- [ ] Procedures updated
- [ ] Team notified

## Resources

### Tools

- **gosec**: Go security checker
- **govulncheck**: Go vulnerability checker
- **age**: File encryption tool
- **gitleaks**: Secret scanner
- **nancy**: Dependency scanner

### Documentation

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [Go Security Best Practices](https://golang.org/doc/security/)
- [age Encryption](https://github.com/FiloSottile/age)

### Training

- OWASP Security Training
- Secure Coding in Go
- Cryptography Basics
- Incident Response Training

## Contact

For security questions or concerns:
- Email: security@yourdomain.com
- GitHub: Open a security advisory
- Documentation: See SECURITY.md

---

**Remember**: Security is everyone's responsibility. When in doubt, ask!
