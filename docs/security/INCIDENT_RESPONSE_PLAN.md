# Security Incident Response Plan

This document outlines the procedures for responding to security incidents in the Agent Stack Controller (asc) project.

## Table of Contents

- [Overview](#overview)
- [Incident Classification](#incident-classification)
- [Response Team](#response-team)
- [Response Procedures](#response-procedures)
- [Communication Plan](#communication-plan)
- [Post-Incident Activities](#post-incident-activities)
- [Appendices](#appendices)

## Overview

### Purpose

This plan provides a structured approach to:
- Detect and respond to security incidents
- Minimize impact and recovery time
- Prevent future incidents
- Comply with security best practices

### Scope

This plan covers security incidents related to:
- Compromised API keys or credentials
- Unauthorized access to systems or data
- Malware or malicious code
- Data breaches or leaks
- Denial of service attacks
- Vulnerabilities in asc or dependencies

### Definitions

- **Security Incident**: Any event that compromises the confidentiality, integrity, or availability of asc or its data
- **Incident Response**: The process of detecting, analyzing, containing, and recovering from security incidents
- **Incident Commander**: The person responsible for coordinating the incident response
- **Stakeholders**: Users, contributors, and anyone affected by the incident

## Incident Classification

### Severity Levels

#### Critical (P0)
- **Impact**: Severe impact on multiple users or systems
- **Examples**:
  - Widespread API key compromise
  - Remote code execution vulnerability
  - Data breach affecting user data
  - Active exploitation in the wild
- **Response Time**: Immediate (within 1 hour)
- **Escalation**: Notify all stakeholders immediately

#### High (P1)
- **Impact**: Significant impact on users or systems
- **Examples**:
  - Single API key compromise
  - Privilege escalation vulnerability
  - Unauthorized access to logs or PIDs
  - Vulnerability with public exploit
- **Response Time**: Within 4 hours
- **Escalation**: Notify security team and maintainers

#### Medium (P2)
- **Impact**: Limited impact on users or systems
- **Examples**:
  - Information disclosure
  - Denial of service vulnerability
  - Dependency vulnerability (no exploit)
  - Configuration weakness
- **Response Time**: Within 24 hours
- **Escalation**: Notify security team

#### Low (P3)
- **Impact**: Minimal impact
- **Examples**:
  - Theoretical vulnerability
  - Security best practice violation
  - Minor information leak
  - Low-risk dependency issue
- **Response Time**: Within 1 week
- **Escalation**: Track in issue tracker

## Response Team

### Roles and Responsibilities

#### Incident Commander
- **Responsibilities**:
  - Coordinate incident response
  - Make critical decisions
  - Communicate with stakeholders
  - Ensure documentation
- **Contact**: security@yourdomain.com

#### Security Lead
- **Responsibilities**:
  - Analyze security implications
  - Recommend remediation steps
  - Review security patches
  - Update security documentation
- **Contact**: security-lead@yourdomain.com

#### Development Lead
- **Responsibilities**:
  - Implement security fixes
  - Deploy patches
  - Test remediation
  - Update code and dependencies
- **Contact**: dev-lead@yourdomain.com

#### Communications Lead
- **Responsibilities**:
  - Draft security advisories
  - Notify affected users
  - Update public documentation
  - Coordinate with media (if needed)
- **Contact**: comms@yourdomain.com

### Contact Information

| Role | Primary Contact | Backup Contact |
|------|----------------|----------------|
| Incident Commander | security@yourdomain.com | backup@yourdomain.com |
| Security Lead | security-lead@yourdomain.com | - |
| Development Lead | dev-lead@yourdomain.com | - |
| Communications Lead | comms@yourdomain.com | - |

### Escalation Path

1. **Initial Report** → Security Team
2. **Confirmed Incident** → Incident Commander
3. **Critical Incident** → All Stakeholders
4. **Public Disclosure** → Communications Lead

## Response Procedures

### Phase 1: Detection and Analysis

#### 1.1 Incident Detection

**Automated Detection:**
- Security scan failures in CI/CD
- Vulnerability alerts from Dependabot
- Gitleaks secret detection
- Unusual log patterns
- Failed authentication attempts

**Manual Detection:**
- User reports
- Security researcher reports
- Code review findings
- Penetration test results

#### 1.2 Initial Assessment

**Gather Information:**
```bash
# Check recent changes
git log --since="24 hours ago" --oneline

# Review logs
tail -n 1000 ~/.asc/logs/asc.log

# Check running processes
ps aux | grep asc

# Review file modifications
find ~/.asc -type f -mtime -1 -ls

# Check network connections
netstat -an | grep ESTABLISHED
```

**Document:**
- Time of detection
- How detected
- Initial symptoms
- Affected systems
- Potential impact

#### 1.3 Classification

**Determine Severity:**
- Assess impact (confidentiality, integrity, availability)
- Identify affected users/systems
- Evaluate exploitability
- Check for active exploitation
- Classify using severity levels (P0-P3)

**Assign Incident Commander:**
- P0/P1: Senior security lead
- P2: Security team member
- P3: Development team member

### Phase 2: Containment

#### 2.1 Short-term Containment

**Immediate Actions:**

```bash
# Stop all agents
asc down

# Disable affected services
systemctl stop mcp_agent_mail  # If using systemd

# Block network access (if needed)
# iptables -A OUTPUT -p tcp --dport 443 -j DROP

# Preserve evidence
cp -r ~/.asc/logs ~/incident-logs-$(date +%Y%m%d-%H%M%S)
cp -r ~/.asc/pids ~/incident-pids-$(date +%Y%m%d-%H%M%S)
```

**For Compromised API Keys:**

```bash
# Immediately revoke keys at provider
# - Claude: https://console.anthropic.com/settings/keys
# - OpenAI: https://platform.openai.com/api-keys
# - Google: https://console.cloud.google.com/apis/credentials

# Remove from local system
rm .env
rm .env.age

# Generate new keys
# Update .env with new keys
# Re-encrypt
asc secrets encrypt
```

**For Unauthorized Access:**

```bash
# Change all passwords
# Rotate encryption keys
asc secrets rotate

# Review access logs
grep "authentication" ~/.asc/logs/*.log

# Check for backdoors
find ~/.asc -type f -name "*.py" -o -name "*.go" | xargs grep -l "backdoor\|malicious"
```

#### 2.2 Long-term Containment

**System Isolation:**
- Disconnect affected systems from network
- Disable remote access
- Implement additional monitoring
- Apply temporary security controls

**Evidence Preservation:**
```bash
# Create forensic image
tar -czf incident-evidence-$(date +%Y%m%d-%H%M%S).tar.gz \
    ~/.asc/logs \
    ~/.asc/pids \
    ~/.asc/age.key \
    asc.toml \
    .env.age

# Calculate checksums
sha256sum incident-evidence-*.tar.gz > checksums.txt

# Store securely
# Move to secure location, not on affected system
```

### Phase 3: Eradication

#### 3.1 Root Cause Analysis

**Investigate:**
- How did the incident occur?
- What vulnerabilities were exploited?
- When did it start?
- What systems were affected?
- What data was accessed?

**Analysis Tools:**
```bash
# Review git history
git log --all --full-history --source --

# Check for malicious commits
git log --all --grep="backdoor\|malicious\|hack"

# Review dependency changes
git log --all -- go.mod go.sum

# Analyze logs
grep -i "error\|fail\|unauthorized" ~/.asc/logs/*.log
```

#### 3.2 Remove Threat

**For Malicious Code:**
```bash
# Identify malicious commits
git log --all --oneline

# Revert malicious changes
git revert <commit-hash>

# Or reset to known good state
git reset --hard <good-commit-hash>

# Force push (if needed)
git push --force origin main
```

**For Vulnerabilities:**
```bash
# Update dependencies
go get -u ./...
go mod tidy

# Apply security patches
git cherry-pick <security-fix-commit>

# Rebuild
make clean
make build

# Run security scans
gosec ./...
govulncheck ./...
```

**For Compromised Credentials:**
```bash
# Rotate all keys
asc secrets rotate

# Update API keys
# Edit .env with new keys

# Re-encrypt
asc secrets encrypt

# Verify old keys are revoked
# Test with old keys (should fail)
```

### Phase 4: Recovery

#### 4.1 System Restoration

**Restore from Clean State:**
```bash
# Backup current state
cp -r ~/.asc ~/.asc.backup-$(date +%Y%m%d-%H%M%S)

# Remove potentially compromised files
rm -rf ~/.asc/logs/*
rm -rf ~/.asc/pids/*

# Reinstall asc
go install github.com/yourusername/asc@latest

# Verify installation
asc --version
asc check

# Restore configuration (from clean backup)
cp ~/backups/asc.toml.clean asc.toml

# Generate new secrets
asc secrets init
# Add API keys to .env
asc secrets encrypt
```

**Restart Services:**
```bash
# Start with monitoring
asc up

# Verify functionality
asc test

# Monitor logs
tail -f ~/.asc/logs/asc.log
```

#### 4.2 Verification

**Security Checks:**
```bash
# Run full security scan
gosec ./...
govulncheck ./...

# Check file permissions
ls -la .env ~/.asc/age.key
# Should be 600

# Verify no backdoors
grep -r "backdoor\|malicious" .

# Test authentication
# Verify API keys work
# Verify old keys don't work
```

**Functional Testing:**
```bash
# Run test suite
go test ./...

# Run integration tests
go test ./test/integration_test.go

# Run E2E tests
go test ./test/e2e_test.go

# Manual testing
asc init
asc up
asc test
asc down
```

#### 4.3 Monitoring

**Enhanced Monitoring:**
```bash
# Monitor logs continuously
tail -f ~/.asc/logs/*.log | grep -i "error\|fail\|unauthorized"

# Check for unusual activity
watch -n 60 'ps aux | grep asc'

# Monitor network connections
watch -n 60 'netstat -an | grep ESTABLISHED'

# Set up alerts (example with cron)
# */5 * * * * grep -i "error" ~/.asc/logs/asc.log | mail -s "ASC Errors" admin@example.com
```

### Phase 5: Post-Incident Activities

#### 5.1 Documentation

**Incident Report Template:**

```markdown
# Security Incident Report

## Incident Summary
- **Incident ID**: INC-YYYY-MM-DD-NNN
- **Date Detected**: YYYY-MM-DD HH:MM UTC
- **Date Resolved**: YYYY-MM-DD HH:MM UTC
- **Severity**: P0/P1/P2/P3
- **Status**: Resolved/Ongoing

## Description
[Brief description of the incident]

## Timeline
- **YYYY-MM-DD HH:MM**: Incident detected
- **YYYY-MM-DD HH:MM**: Initial containment
- **YYYY-MM-DD HH:MM**: Root cause identified
- **YYYY-MM-DD HH:MM**: Threat eradicated
- **YYYY-MM-DD HH:MM**: Systems restored
- **YYYY-MM-DD HH:MM**: Incident closed

## Impact
- **Systems Affected**: [List]
- **Users Affected**: [Number/List]
- **Data Compromised**: [Yes/No, Details]
- **Downtime**: [Duration]

## Root Cause
[Detailed analysis of how the incident occurred]

## Response Actions
[List of actions taken during response]

## Lessons Learned
[What went well, what could be improved]

## Recommendations
[Preventive measures for the future]

## Attachments
- Logs
- Screenshots
- Evidence files
```

#### 5.2 Lessons Learned

**Post-Incident Review Meeting:**
- Schedule within 1 week of resolution
- Include all response team members
- Review timeline and actions
- Identify improvements
- Update procedures

**Questions to Address:**
- What happened and why?
- What worked well?
- What could be improved?
- What should we do differently?
- What preventive measures should we implement?

#### 5.3 Preventive Measures

**Update Security Controls:**
```bash
# Add new security checks
# Update .golangci.yml
# Update .gosec.json

# Add new tests
# Create test/security_regression_test.go

# Update documentation
# Update SECURITY.md
# Update SECURITY_BEST_PRACTICES.md
```

**Improve Monitoring:**
```bash
# Add new log patterns
# Update alerting rules
# Enhance security scanning
# Add new CI/CD checks
```

**Training and Awareness:**
- Share incident report with team
- Conduct security training
- Update security guidelines
- Review with contributors

## Communication Plan

### Internal Communication

**During Incident:**
- Use dedicated Slack channel: #security-incident
- Regular status updates (every 2-4 hours for P0/P1)
- Document all decisions and actions
- Keep team informed of progress

**After Resolution:**
- Send incident summary to team
- Schedule post-incident review
- Share lessons learned
- Update documentation

### External Communication

#### Security Advisory Template

```markdown
# Security Advisory: [Title]

**Advisory ID**: ASC-YYYY-NNN
**Date**: YYYY-MM-DD
**Severity**: Critical/High/Medium/Low

## Summary
[Brief description of the vulnerability]

## Affected Versions
- asc versions: X.X.X - Y.Y.Y

## Impact
[Description of potential impact]

## Mitigation
[Immediate steps users should take]

## Resolution
[How the issue was fixed]

## Upgrade Instructions
```bash
# Update to latest version
go install github.com/yourusername/asc@latest

# Verify version
asc --version

# Rotate keys (if needed)
asc secrets rotate
```

## Timeline
- **YYYY-MM-DD**: Vulnerability discovered
- **YYYY-MM-DD**: Fix developed
- **YYYY-MM-DD**: Fix released
- **YYYY-MM-DD**: Public disclosure

## Credit
[Credit to reporter, if applicable]

## References
- CVE-YYYY-NNNNN (if assigned)
- GitHub Security Advisory: GHSA-XXXX-XXXX-XXXX
```

#### User Notification

**For Critical Incidents (P0):**
- Immediate notification via:
  - GitHub Security Advisory
  - Email to users (if available)
  - Social media
  - Project website
- Include:
  - What happened
  - What users should do
  - How to get help

**For High Incidents (P1):**
- Notification within 24 hours via:
  - GitHub Security Advisory
  - Release notes
  - Documentation update

**For Medium/Low Incidents (P2/P3):**
- Include in next release notes
- Update security documentation

### Media Communication

**If Media Inquiry:**
1. Refer to Communications Lead
2. Use prepared statement
3. Don't speculate
4. Focus on facts
5. Emphasize user safety

**Prepared Statement Template:**
```
We are aware of [brief description] affecting asc. We take security 
seriously and are actively investigating. We have implemented [immediate 
actions] to protect users. We will provide updates as more information 
becomes available. Users should [recommended actions]. For more 
information, see [link to advisory].
```

## Appendices

### Appendix A: Contact Lists

#### Internal Contacts
- Security Team: security@yourdomain.com
- Development Team: dev@yourdomain.com
- Management: management@yourdomain.com

#### External Contacts
- GitHub Security: security@github.com
- CERT: cert@cert.org
- CVE Program: cve-assign@mitre.org

### Appendix B: Tools and Resources

#### Security Tools
```bash
# Install security tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/sonatype-nexus-community/nancy@latest

# Run scans
gosec ./...
govulncheck ./...
go list -json -deps ./... | nancy sleuth
```

#### Forensic Tools
```bash
# Log analysis
grep -i "error\|fail\|unauthorized" ~/.asc/logs/*.log

# File integrity
find ~/.asc -type f -exec sha256sum {} \; > checksums.txt

# Network analysis
netstat -an
lsof -i
tcpdump -i any port 8765
```

### Appendix C: Incident Response Checklist

#### Detection Phase
- [ ] Incident detected and logged
- [ ] Initial assessment completed
- [ ] Severity classified
- [ ] Incident Commander assigned
- [ ] Response team notified

#### Containment Phase
- [ ] Affected systems identified
- [ ] Short-term containment implemented
- [ ] Evidence preserved
- [ ] Stakeholders notified
- [ ] Long-term containment planned

#### Eradication Phase
- [ ] Root cause identified
- [ ] Threat removed
- [ ] Vulnerabilities patched
- [ ] Systems cleaned
- [ ] Security controls updated

#### Recovery Phase
- [ ] Systems restored
- [ ] Functionality verified
- [ ] Security verified
- [ ] Monitoring enhanced
- [ ] Normal operations resumed

#### Post-Incident Phase
- [ ] Incident documented
- [ ] Post-incident review completed
- [ ] Lessons learned documented
- [ ] Preventive measures implemented
- [ ] Procedures updated
- [ ] Team trained
- [ ] Users notified

### Appendix D: Incident Severity Matrix

| Factor | Critical (P0) | High (P1) | Medium (P2) | Low (P3) |
|--------|--------------|-----------|-------------|----------|
| **Confidentiality** | Widespread data breach | Limited data breach | Information disclosure | Minimal disclosure |
| **Integrity** | System compromise | Data modification | Configuration change | Minor inconsistency |
| **Availability** | Complete outage | Partial outage | Degraded performance | Minor disruption |
| **Exploitability** | Active exploitation | Public exploit | Proof of concept | Theoretical |
| **Scope** | All users | Multiple users | Single user | Internal only |

### Appendix E: Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2025-11-10 | Initial version | Security Team |

---

**Document Owner**: Security Team  
**Last Updated**: 2025-11-10  
**Next Review**: 2026-11-10  
**Classification**: Internal Use Only
