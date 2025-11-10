# Security Policy

## Secrets Management

### Overview

Agent Stack Controller (asc) uses **age encryption** ([FiloSottile/age](https://github.com/FiloSottile/age)) for secure secrets management. This prevents accidental exposure of API keys and sensitive configuration in version control.

### Why age?

- **Simple**: Easy to use, hard to misuse
- **Secure**: Modern cryptography (X25519, ChaCha20-Poly1305)
- **Audited**: Small codebase, easy to audit
- **Standard**: Growing adoption in the security community

### Threat Model

**What we protect against:**
- ✅ Accidental commits of plaintext secrets to git
- ✅ Secrets exposure in public repositories
- ✅ Unauthorized access to secrets in shared repositories
- ✅ Secrets leakage through log files or error messages

**What we don't protect against:**
- ❌ Compromised developer machines (if your machine is compromised, your decrypted secrets are accessible)
- ❌ Malicious code execution (if malicious code runs with your permissions, it can decrypt secrets)
- ❌ Physical access to unlocked machines
- ❌ Keyloggers or screen capture malware

### Security Architecture

```
┌─────────────────────────────────────────────────────────┐
│ Developer Machine                                       │
│                                                         │
│  ~/.asc/age.key (private key, 0600)                    │
│       │                                                 │
│       ├─ Encrypts ──→ .env.age (safe to commit)       │
│       │                                                 │
│       └─ Decrypts ──→ .env (gitignored, 0600)         │
│                                                         │
└─────────────────────────────────────────────────────────┘
                          │
                          │ git push
                          ▼
┌─────────────────────────────────────────────────────────┐
│ Git Repository (GitHub/GitLab)                          │
│                                                         │
│  ✓ .env.age (encrypted, safe)                          │
│  ✗ .env (never committed, gitignored)                  │
│  ✗ ~/.asc/age.key (never committed)                    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Setup and Usage

#### Initial Setup

```bash
# 1. Install age
brew install age  # macOS
apt install age   # Debian/Ubuntu
scoop install age # Windows

# 2. Initialize encryption
asc secrets init

# 3. Create .env file
cp .env.example .env
# Edit .env with your API keys

# 4. Encrypt secrets
asc secrets encrypt

# 5. Commit encrypted file
git add .env.age
git commit -m "Add encrypted secrets"
```

#### Daily Workflow

```bash
# Decrypt when starting work
asc secrets decrypt

# Work with decrypted .env
asc up

# Re-encrypt after changes
asc secrets encrypt
git add .env.age
git commit -m "Update secrets"
```

### Key Management

#### Key Storage

- **Location**: `~/.asc/age.key`
- **Permissions**: 0600 (owner read/write only)
- **Format**: age private key with embedded public key

#### Key Backup

**CRITICAL**: Back up your age key securely!

Options:
1. **Password Manager**: Store in 1Password, LastPass, Bitwarden
2. **Encrypted Backup**: Use another encryption layer
3. **Hardware Token**: Store on YubiKey or similar
4. **Paper Backup**: Print and store in safe (for disaster recovery)

```bash
# View your key for backup
cat ~/.asc/age.key

# Backup to encrypted archive
tar czf - ~/.asc/age.key | age -p > age-key-backup.tar.gz.age
```

#### Key Rotation

Rotate keys every 90 days or immediately if compromised:

```bash
# Rotate key and re-encrypt all files
asc secrets rotate

# Old key is backed up to ~/.asc/age.key.old
```

### Team Collaboration

#### Sharing Encrypted Secrets

**Option 1: Shared Key (Simple)**
- Share the age key securely (encrypted channel)
- All team members use the same key
- ⚠️ If one person's machine is compromised, all secrets are at risk

**Option 2: Multiple Recipients (Recommended)**
```bash
# Each team member generates their own key
asc secrets init

# Share public keys
asc secrets status  # Shows public key

# Encrypt for multiple recipients
age -r age1alice... -r age1bob... -r age1carol... -o .env.age .env

# Each person decrypts with their own key
asc secrets decrypt
```

**Option 3: External Secrets Manager**
- Use HashiCorp Vault, AWS Secrets Manager, or similar
- asc can integrate with external secret stores
- Best for production environments

### File Permissions

asc automatically sets restrictive permissions:

| File | Permissions | Reason |
|------|-------------|--------|
| `.env` | 0600 | Plaintext secrets, owner only |
| `~/.asc/age.key` | 0600 | Private key, owner only |
| `.env.age` | 0644 | Encrypted, safe to share |
| PID files | 0644 | Process info, readable |
| Log files | 0644 | Logs, readable |

### Git Configuration

Ensure these files are never committed:

```gitignore
# .gitignore
.env
.env.*
!.env.example
!.env.*.age
*.key
secrets/
.secrets/
```

### Security Checklist

Before committing:
- [ ] `.env` is in `.gitignore`
- [ ] Only `.env.age` (encrypted) is committed
- [ ] `~/.asc/age.key` is backed up securely
- [ ] `.env` has 0600 permissions
- [ ] No secrets in log files or error messages
- [ ] No secrets in commit messages

### Incident Response

#### If secrets are accidentally committed:

1. **Immediately rotate all exposed secrets**
   ```bash
   # Rotate API keys at providers
   # - Claude: https://console.anthropic.com
   # - OpenAI: https://platform.openai.com/api-keys
   # - Google: https://console.cloud.google.com
   ```

2. **Remove from git history**
   ```bash
   # Use BFG Repo-Cleaner or git-filter-repo
   git filter-repo --path .env --invert-paths
   git push --force
   ```

3. **Notify team**
   - Inform all team members
   - Ensure everyone updates their keys

4. **Audit access**
   - Check for unauthorized API usage
   - Review access logs

#### If age key is compromised:

1. **Generate new key immediately**
   ```bash
   asc secrets rotate
   ```

2. **Re-encrypt all secrets**
   - Rotation command handles this automatically

3. **Distribute new public key**
   - Share with team members securely

4. **Revoke old key**
   - Delete `~/.asc/age.key.old` after verification

### Best Practices

1. **Never commit plaintext secrets**
   - Always use `.env.age` (encrypted)
   - Add `.env` to `.gitignore`

2. **Use environment-specific files**
   ```
   .env.age           # Development
   .env.prod.age      # Production
   .env.staging.age   # Staging
   ```

3. **Rotate keys regularly**
   - Every 90 days minimum
   - Immediately if compromised
   - After team member departure

4. **Audit regularly**
   ```bash
   # Check secrets status
   asc secrets status
   
   # Verify no plaintext secrets in git
   git log --all --full-history -- .env
   ```

5. **Use strong API keys**
   - Enable key rotation at provider
   - Use scoped/limited permissions
   - Monitor API usage

6. **Secure your development environment**
   - Use full disk encryption
   - Lock screen when away
   - Keep OS and tools updated
   - Use antivirus/antimalware

### Compliance

This secrets management approach helps meet:

- **SOC 2**: Encryption of sensitive data
- **GDPR**: Protection of personal data
- **HIPAA**: Encryption requirements
- **PCI DSS**: Key management standards

### Reporting Security Issues

If you discover a security vulnerability:

1. **Do NOT open a public issue**
2. Email: security@yourcompany.com
3. Include:
   - Description of vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will respond within 48 hours.

### Security Updates

- Subscribe to age security advisories: https://github.com/FiloSottile/age/security
- Monitor asc releases for security patches
- Enable GitHub security alerts for dependencies

### Additional Resources

- [age specification](https://age-encryption.org/)
- [age GitHub repository](https://github.com/FiloSottile/age)
- [OWASP Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [Git Secrets Prevention](https://git-scm.com/book/en/v2/Git-Tools-Credential-Storage)

### License

This security policy is part of the Agent Stack Controller project and is licensed under the same terms.

---

**Last Updated**: 2025-11-09
**Version**: 1.0
