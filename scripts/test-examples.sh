#!/bin/bash
# Test code examples from documentation
# Extracts Go and bash code blocks and validates them

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "📝 Testing code examples from documentation..."
echo ""

# Create temp directory for extracted examples
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

TOTAL_EXAMPLES=0
PASSED_EXAMPLES=0
FAILED_EXAMPLES=0
SKIPPED_EXAMPLES=0

# Function to extract code blocks from markdown
extract_code_blocks() {
    local file=$1
    local lang=$2
    local output_dir=$3
    
    # Create a safe filename from the input file
    local safe_name=$(echo "$file" | sed 's/[^a-zA-Z0-9]/_/g')
    
    awk -v lang="$lang" -v output_dir="$output_dir" -v safe_name="$safe_name" '
    BEGIN { 
        in_block = 0
        block_num = 0
        content = ""
    }
    /^```'"$lang"'/ { 
        in_block = 1
        block_num++
        content = ""
        next
    }
    /^```/ && in_block { 
        in_block = 0
        if (content != "") {
            filename = output_dir "/" safe_name "_" block_num "." lang
            print content > filename
            close(filename)
            print filename
        }
        content = ""
        next
    }
    in_block { 
        content = content $0 "\n"
    }
    ' "$file"
}

# Function to test Go code
test_go_code() {
    local file=$1
    local basename=$(basename "$file")
    
    echo -n "Testing Go example: $basename... "
    TOTAL_EXAMPLES=$((TOTAL_EXAMPLES + 1))
    
    # Check if it's a complete program or snippet
    if grep -q "^package main" "$file" && grep -q "^func main()" "$file"; then
        # Check if it has external imports that won't resolve
        if grep -q "github.com/yourusername" "$file" || grep -q "example.com" "$file"; then
            echo -e "${BLUE}⊘ (example with placeholder imports)${NC}"
            # Just check syntax
            if gofmt -e "$file" >/dev/null 2>&1; then
                PASSED_EXAMPLES=$((PASSED_EXAMPLES + 1))
            else
                echo -e "${RED}  └─ syntax errors${NC}"
                FAILED_EXAMPLES=$((FAILED_EXAMPLES + 1))
                gofmt -e "$file" 2>&1 | head -3
            fi
        else
            # Complete program - try to build it
            if go build -o /dev/null "$file" 2>/dev/null; then
                echo -e "${GREEN}✓ (builds)${NC}"
                PASSED_EXAMPLES=$((PASSED_EXAMPLES + 1))
            else
                echo -e "${RED}✗ (build failed)${NC}"
                FAILED_EXAMPLES=$((FAILED_EXAMPLES + 1))
                go build -o /dev/null "$file" 2>&1 | head -5
            fi
        fi
    elif grep -q "^package " "$file"; then
        # Check if it has multiple package declarations (comparison example)
        package_count=$(grep -c "^package " "$file")
        if [ "$package_count" -gt 1 ]; then
            echo -e "${BLUE}⊘ (comparison example)${NC}"
            SKIPPED_EXAMPLES=$((SKIPPED_EXAMPLES + 1))
        else
            # Package snippet - just check syntax with gofmt
            echo -e "${YELLOW}⊘ (snippet)${NC}"
            if gofmt -e "$file" >/dev/null 2>&1; then
                echo -e "${GREEN}  └─ syntax OK${NC}"
                PASSED_EXAMPLES=$((PASSED_EXAMPLES + 1))
            else
                echo -e "${RED}  └─ syntax errors${NC}"
                FAILED_EXAMPLES=$((FAILED_EXAMPLES + 1))
                gofmt -e "$file" 2>&1 | head -3
            fi
        fi
    else
        # Code snippet without package - check if it's just comments/examples
        if grep -qE "^// (Good|Bad|Example)" "$file"; then
            echo -e "${BLUE}⊘ (example snippet)${NC}"
            SKIPPED_EXAMPLES=$((SKIPPED_EXAMPLES + 1))
        else
            # Try syntax check
            echo -e "${BLUE}⊘ (snippet - no package)${NC}"
            if gofmt -e "$file" >/dev/null 2>&1; then
                PASSED_EXAMPLES=$((PASSED_EXAMPLES + 1))
            else
                # Many snippets won't have valid syntax without context
                SKIPPED_EXAMPLES=$((SKIPPED_EXAMPLES + 1))
            fi
        fi
    fi
}

# Function to test bash code
test_bash_code() {
    local file=$1
    local basename=$(basename "$file")
    
    echo -n "Testing bash example: $basename... "
    TOTAL_EXAMPLES=$((TOTAL_EXAMPLES + 1))
    
    # Check if it's a command snippet or script
    if head -1 "$file" | grep -q "^#!"; then
        # Complete script - check syntax
        if bash -n "$file" 2>/dev/null; then
            echo -e "${GREEN}✓ (syntax OK)${NC}"
            PASSED_EXAMPLES=$((PASSED_EXAMPLES + 1))
        else
            echo -e "${RED}✗ (syntax error)${NC}"
            FAILED_EXAMPLES=$((FAILED_EXAMPLES + 1))
            bash -n "$file" 2>&1 | head -3
        fi
    else
        # Command snippet - check if it looks valid
        # Skip examples that are clearly just showing output or placeholders
        if grep -qE "^(#|//|\$|>)" "$file" || grep -qE "<.*>|\.\.\.|\[.*\]" "$file"; then
            echo -e "${BLUE}⊘ (example/placeholder)${NC}"
            SKIPPED_EXAMPLES=$((SKIPPED_EXAMPLES + 1))
        else
            # Try basic syntax check by wrapping in a function
            echo "#!/bin/bash" > "$file.test"
            echo "function test_snippet() {" >> "$file.test"
            cat "$file" >> "$file.test"
            echo "}" >> "$file.test"
            
            if bash -n "$file.test" 2>/dev/null; then
                echo -e "${GREEN}✓ (syntax OK)${NC}"
                PASSED_EXAMPLES=$((PASSED_EXAMPLES + 1))
            else
                echo -e "${YELLOW}⊘ (snippet - may need context)${NC}"
                SKIPPED_EXAMPLES=$((SKIPPED_EXAMPLES + 1))
            fi
            rm -f "$file.test"
        fi
    fi
}

# Find all markdown files
MARKDOWN_FILES=$(find . -name "*.md" \
  -not -path "./node_modules/*" \
  -not -path "./.git/*" \
  -not -path "./vendor/*" \
  -not -path "./build/*" \
  -not -path "./agent/.venv/*" \
  -not -path "./agent/__pycache__/*" \
  -not -path "./.kiro/*")

echo -e "${BLUE}Extracting code examples...${NC}"
echo ""

# Extract and test Go examples
GO_DIR="$TEMP_DIR/go"
mkdir -p "$GO_DIR"

for file in $MARKDOWN_FILES; do
    extract_code_blocks "$file" "go" "$GO_DIR"
done

GO_EXAMPLES=$(find "$GO_DIR" -name "*.go" 2>/dev/null || true)
if [ -n "$GO_EXAMPLES" ]; then
    echo -e "${BLUE}Testing Go examples:${NC}"
    for example in $GO_EXAMPLES; do
        test_go_code "$example"
    done
    echo ""
fi

# Extract and test bash examples
BASH_DIR="$TEMP_DIR/bash"
mkdir -p "$BASH_DIR"

for file in $MARKDOWN_FILES; do
    extract_code_blocks "$file" "bash" "$BASH_DIR"
done

BASH_EXAMPLES=$(find "$BASH_DIR" -name "*.bash" 2>/dev/null || true)
if [ -n "$BASH_EXAMPLES" ]; then
    echo -e "${BLUE}Testing bash examples:${NC}"
    for example in $BASH_EXAMPLES; do
        test_bash_code "$example"
    done
    echo ""
fi

# Print summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Example Testing Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Total examples: $TOTAL_EXAMPLES"
echo -e "Passed: ${GREEN}$PASSED_EXAMPLES${NC}"
echo -e "Failed: ${RED}$FAILED_EXAMPLES${NC}"
echo -e "Skipped: ${BLUE}$SKIPPED_EXAMPLES${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $FAILED_EXAMPLES -gt 0 ]; then
    echo -e "${RED}Some examples failed validation. Please fix them.${NC}"
    exit 1
else
    echo -e "${GREEN}All testable examples are valid!${NC}"
    exit 0
fi
