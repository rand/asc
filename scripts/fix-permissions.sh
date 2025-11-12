#!/bin/bash
# Fix file and directory permissions for Agent Stack Controller
# This script ensures all sensitive files and directories have secure permissions

set -e

echo "🔒 Fixing Agent Stack Controller Permissions"
echo "=============================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track if any changes were made
CHANGES_MADE=0

# Function to fix file permissions
fix_file_perms() {
    local file=$1
    local target_perms=$2
    
    if [ -f "$file" ]; then
        current_perms=$(stat -f "%Lp" "$file" 2>/dev/null || stat -c "%a" "$file" 2>/dev/null)
        if [ "$current_perms" != "$target_perms" ]; then
            echo -e "${YELLOW}Fixing:${NC} $file (${current_perms} → ${target_perms})"
            chmod "$target_perms" "$file"
            CHANGES_MADE=1
        else
            echo -e "${GREEN}OK:${NC} $file ($current_perms)"
        fi
    else
        echo -e "${YELLOW}Skip:${NC} $file (not found)"
    fi
}

# Function to fix directory permissions
fix_dir_perms() {
    local dir=$1
    local target_perms=$2
    
    if [ -d "$dir" ]; then
        current_perms=$(stat -f "%Lp" "$dir" 2>/dev/null || stat -c "%a" "$dir" 2>/dev/null)
        if [ "$current_perms" != "$target_perms" ]; then
            echo -e "${YELLOW}Fixing:${NC} $dir (${current_perms} → ${target_perms})"
            chmod "$target_perms" "$dir"
            CHANGES_MADE=1
        else
            echo -e "${GREEN}OK:${NC} $dir ($current_perms)"
        fi
    else
        echo -e "${YELLOW}Skip:${NC} $dir (not found)"
    fi
}

# Fix .env file
echo "Checking .env file..."
fix_file_perms ".env" "600"
echo ""

# Fix age key
echo "Checking age encryption key..."
fix_file_perms "$HOME/.asc/age.key" "600"
echo ""

# Fix directories
echo "Checking directories..."
fix_dir_perms "$HOME/.asc" "700"
fix_dir_perms "$HOME/.asc/logs" "700"
fix_dir_perms "$HOME/.asc/pids" "700"
fix_dir_perms "$HOME/.asc/playbooks" "700"
echo ""

# Fix log files
echo "Checking log files..."
if [ -d "$HOME/.asc/logs" ]; then
    for logfile in "$HOME/.asc/logs"/*.log; do
        if [ -f "$logfile" ]; then
            fix_file_perms "$logfile" "600"
        fi
    done
else
    echo -e "${YELLOW}Skip:${NC} No log directory found"
fi
echo ""

# Fix PID files
echo "Checking PID files..."
if [ -d "$HOME/.asc/pids" ]; then
    for pidfile in "$HOME/.asc/pids"/*.json; do
        if [ -f "$pidfile" ]; then
            fix_file_perms "$pidfile" "600"
        fi
    done
else
    echo -e "${YELLOW}Skip:${NC} No PID directory found"
fi
echo ""

# Fix encrypted files
echo "Checking encrypted files..."
fix_file_perms ".env.age" "600"
echo ""

# Check if .env is tracked by git
echo "Checking git tracking..."
if [ -f ".env" ] && git ls-files --error-unmatch .env >/dev/null 2>&1; then
    echo -e "${RED}WARNING:${NC} .env is tracked by git!"
    echo "Run the following to remove it:"
    echo "  git rm --cached .env"
    echo "  git commit -m 'Remove .env from tracking'"
    CHANGES_MADE=1
else
    echo -e "${GREEN}OK:${NC} .env is not tracked by git"
fi
echo ""

# Summary
echo "=============================================="
if [ $CHANGES_MADE -eq 1 ]; then
    echo -e "${GREEN}✓${NC} Permissions fixed!"
    echo ""
    echo "Run 'scripts/check-security.sh' to verify security settings."
else
    echo -e "${GREEN}✓${NC} All permissions are already correct!"
fi
echo ""
