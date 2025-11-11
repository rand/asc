#!/bin/bash
# Security validation script for Agent Stack Controller
# This script checks for common security issues and provides remediation steps

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ISSUES_FOUND=0

echo "🔒 Agent Stack Controller - Security Validation"
echo "================================================"
echo ""

# Check 1: .env file permissions
echo "Checking .env file permissions..."
if [ -f .env ]; then
    PERMS=$(stat -f "%Lp" .env 2>/dev/null || stat -c "%a" .env 2>/dev/null)
    if [ "$PERMS" != "600" ]; then
        echo -e "${RED}✗ FAIL${NC}: .env has insecure permissions: $PERMS"
        echo "  Fix with: chmod 600 .env"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    else
        echo -e "${GREEN}✓ PASS${NC}: .env has secure permissions (600)"
    fi
else
    echo -e "${YELLOW}⊘ SKIP${NC}: .env file not found"
fi
echo ""

# Check 2: age.key permissions
echo "Checking age.key permissions..."
AGE_KEY="$HOME/.asc/age.key"
if [ -f "$AGE_KEY" ]; then
    PERMS=$(stat -f "%Lp" "$AGE_KEY" 2>/dev/null || stat -c "%a" "$AGE_KEY" 2>/dev/null)
    if [ "$PERMS" != "600" ]; then
        echo -e "${RED}✗ FAIL${NC}: age.key has insecure permissions: $PERMS"
        echo "  Fix with: chmod 600 $AGE_KEY"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    else
        echo -e "${GREEN}✓ PASS${NC}: age.key has secure permissions (600)"
    fi
else
    echo -e "${YELLOW}⊘ SKIP${NC}: age.key not found"
fi
echo ""

# Check 3: .env in .gitignore
echo "Checking .gitignore..."
if grep -q "^\.env$" .gitignore 2>/dev/null; then
    echo -e "${GREEN}✓ PASS${NC}: .env is in .gitignore"
else
    echo -e "${RED}✗ FAIL${NC}: .env not found in .gitignore"
    echo "  Fix with: echo '.env' >> .gitignore"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi
echo ""

# Check 4: .env not tracked by git
echo "Checking if .env is tracked by git..."
if git ls-files --error-unmatch .env >/dev/null 2>&1; then
    echo -e "${RED}✗ FAIL${NC}: .env is tracked by git"
    echo "  Fix with: git rm --cached .env && git commit -m 'Remove .env from git'"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
else
    echo -e "${GREEN}✓ PASS${NC}: .env is not tracked by git"
fi
echo ""

# Check 5: No secrets in logs
echo "Checking log files for secrets..."
LOG_DIR="$HOME/.asc/logs"
if [ -d "$LOG_DIR" ]; then
    SECRET_PATTERNS="sk-ant-|sk-[a-zA-Z0-9-]{20,}|AIza[a-zA-Z0-9_-]{35}"
    if grep -r -E "$SECRET_PATTERNS" "$LOG_DIR" >/dev/null 2>&1; then
        echo -e "${RED}✗ FAIL${NC}: Potential secrets found in log files"
        echo "  Review logs in: $LOG_DIR"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    else
        echo -e "${GREEN}✓ PASS${NC}: No secrets found in log files"
    fi
else
    echo -e "${YELLOW}⊘ SKIP${NC}: Log directory not found"
fi
echo ""

# Check 6: No hardcoded secrets in code
echo "Checking for hardcoded secrets in code..."
if grep -r -E "(sk-ant-[a-zA-Z0-9]{20,}|sk-[a-zA-Z0-9-]{40,}|AIza[a-zA-Z0-9_-]{35})" \
    --include="*.go" --include="*.py" \
    --exclude-dir=vendor --exclude-dir=.git \
    --exclude-dir=node_modules . >/dev/null 2>&1; then
    echo -e "${RED}✗ FAIL${NC}: Potential hardcoded secrets found in code"
    echo "  Review code for hardcoded API keys"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
else
    echo -e "${GREEN}✓ PASS${NC}: No hardcoded secrets found"
fi
echo ""

# Check 7: Directory permissions
echo "Checking directory permissions..."
for DIR in "$HOME/.asc/logs" "$HOME/.asc/pids"; do
    if [ -d "$DIR" ]; then
        PERMS=$(stat -f "%Lp" "$DIR" 2>/dev/null || stat -c "%a" "$DIR" 2>/dev/null)
        # Check if world-writable
        if [ $((PERMS & 2)) -ne 0 ]; then
            echo -e "${RED}✗ FAIL${NC}: $DIR is world-writable: $PERMS"
            echo "  Fix with: chmod 700 $DIR"
            ISSUES_FOUND=$((ISSUES_FOUND + 1))
        else
            echo -e "${GREEN}✓ PASS${NC}: $DIR has secure permissions ($PERMS)"
        fi
    fi
done
echo ""

# Check 8: Encrypted files permissions
echo "Checking encrypted file permissions..."
if ls *.age >/dev/null 2>&1; then
    for FILE in *.age; do
        PERMS=$(stat -f "%Lp" "$FILE" 2>/dev/null || stat -c "%a" "$FILE" 2>/dev/null)
        # Check if group or world readable
        if [ $((PERMS & 44)) -ne 0 ]; then
            echo -e "${YELLOW}⚠ WARN${NC}: $FILE has permissive permissions: $PERMS"
            echo "  Consider: chmod 600 $FILE"
        else
            echo -e "${GREEN}✓ PASS${NC}: $FILE has secure permissions ($PERMS)"
        fi
    done
else
    echo -e "${YELLOW}⊘ SKIP${NC}: No encrypted files found"
fi
echo ""

# Summary
echo "================================================"
if [ $ISSUES_FOUND -eq 0 ]; then
    echo -e "${GREEN}✓ Security validation passed!${NC}"
    echo "No security issues found."
    exit 0
else
    echo -e "${RED}✗ Security validation failed!${NC}"
    echo "Found $ISSUES_FOUND security issue(s)."
    echo ""
    echo "Please review and fix the issues above."
    exit 1
fi
