#!/bin/bash
# Link validation script for markdown documentation
# Uses markdown-link-check to validate all links in markdown files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔗 Validating links in markdown documentation..."
echo ""

# Create config file for markdown-link-check
cat > /tmp/markdown-link-check-config.json <<EOF
{
  "ignorePatterns": [
    {
      "pattern": "^http://localhost"
    },
    {
      "pattern": "^https://localhost"
    },
    {
      "pattern": "^http://127.0.0.1"
    }
  ],
  "timeout": "10s",
  "retryOn429": true,
  "retryCount": 3,
  "fallbackRetryDelay": "5s",
  "aliveStatusCodes": [200, 206, 301, 302, 307, 308]
}
EOF

# Find all markdown files
MARKDOWN_FILES=$(find . -name "*.md" \
  -not -path "./node_modules/*" \
  -not -path "./.git/*" \
  -not -path "./vendor/*" \
  -not -path "./build/*" \
  -not -path "./agent/.venv/*" \
  -not -path "./agent/__pycache__/*" \
  -not -path "./.kiro/*")

TOTAL_FILES=0
FAILED_FILES=0
PASSED_FILES=0

# Check if npx is available
if ! command -v npx &> /dev/null; then
    echo -e "${RED}Error: npx is not installed. Please install Node.js and npm.${NC}"
    exit 1
fi

# Process each markdown file
for file in $MARKDOWN_FILES; do
    TOTAL_FILES=$((TOTAL_FILES + 1))
    echo -n "Checking $file... "
    
    if npx --yes markdown-link-check "$file" --config /tmp/markdown-link-check-config.json --quiet > /tmp/link-check-output.txt 2>&1; then
        echo -e "${GREEN}✓${NC}"
        PASSED_FILES=$((PASSED_FILES + 1))
    else
        echo -e "${RED}✗${NC}"
        FAILED_FILES=$((FAILED_FILES + 1))
        cat /tmp/link-check-output.txt
        echo ""
    fi
done

# Clean up temp files
rm -f /tmp/markdown-link-check-config.json /tmp/link-check-output.txt

# Print summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Link Validation Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Total files checked: $TOTAL_FILES"
echo -e "Passed: ${GREEN}$PASSED_FILES${NC}"
echo -e "Failed: ${RED}$FAILED_FILES${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $FAILED_FILES -gt 0 ]; then
    echo -e "${RED}Link validation failed. Please fix broken links.${NC}"
    exit 1
else
    echo -e "${GREEN}All links are valid!${NC}"
    exit 0
fi
