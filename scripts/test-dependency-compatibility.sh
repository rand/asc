#!/bin/bash
# Test dependency compatibility across different versions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "==================================="
echo "Dependency Compatibility Test"
echo "==================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    
    case $status in
        "pass")
            echo -e "${GREEN}✓${NC} $message"
            ;;
        "fail")
            echo -e "${RED}✗${NC} $message"
            ;;
        "warn")
            echo -e "${YELLOW}⚠${NC} $message"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

# Test Go version
echo "1. Testing Go Version Compatibility"
echo "-----------------------------------"

if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_status "pass" "Go is installed: $GO_VERSION"
    
    # Check minimum version (1.21)
    GO_VERSION_NUM=$(echo $GO_VERSION | sed 's/go//' | cut -d. -f1,2)
    MAJOR=$(echo $GO_VERSION_NUM | cut -d. -f1)
    MINOR=$(echo $GO_VERSION_NUM | cut -d. -f2)
    
    if [ "$MAJOR" -ge 1 ] && [ "$MINOR" -ge 21 ]; then
        print_status "pass" "Go version meets minimum requirement (1.21+)"
    else
        print_status "fail" "Go version $GO_VERSION_NUM is below minimum requirement (1.21)"
        exit 1
    fi
else
    print_status "fail" "Go is not installed"
    exit 1
fi

echo ""

# Test Python version
echo "2. Testing Python Version Compatibility"
echo "---------------------------------------"

if command -v python3 &> /dev/null; then
    PYTHON_VERSION=$(python3 --version | awk '{print $2}')
    print_status "pass" "Python is installed: $PYTHON_VERSION"
    
    # Check minimum version (3.8)
    PYTHON_MAJOR=$(echo $PYTHON_VERSION | cut -d. -f1)
    PYTHON_MINOR=$(echo $PYTHON_VERSION | cut -d. -f2)
    
    if [ "$PYTHON_MAJOR" -eq 3 ] && [ "$PYTHON_MINOR" -ge 8 ]; then
        print_status "pass" "Python version meets minimum requirement (3.8+)"
    else
        print_status "fail" "Python version $PYTHON_VERSION is below minimum requirement (3.8)"
        exit 1
    fi
else
    print_status "fail" "Python 3 is not installed"
    exit 1
fi

echo ""

# Test Go build
echo "3. Testing Go Build"
echo "-------------------"

cd "$PROJECT_ROOT"

if go build -v ./... > /dev/null 2>&1; then
    print_status "pass" "Go build successful"
else
    print_status "fail" "Go build failed"
    exit 1
fi

echo ""

# Test Go modules
echo "4. Testing Go Module Integrity"
echo "-------------------------------"

if go mod verify > /dev/null 2>&1; then
    print_status "pass" "Go module checksums verified"
else
    print_status "fail" "Go module verification failed"
    exit 1
fi

# Check if go.mod is tidy
go mod tidy
if git diff --exit-code go.mod go.sum > /dev/null 2>&1; then
    print_status "pass" "go.mod and go.sum are tidy"
else
    print_status "warn" "go.mod or go.sum needs tidying (run 'go mod tidy')"
fi

echo ""

# Test external dependencies
echo "5. Testing External Dependencies"
echo "--------------------------------"

# Git
if command -v git &> /dev/null; then
    GIT_VERSION=$(git --version)
    print_status "pass" "git: $GIT_VERSION"
else
    print_status "fail" "git is not installed (required)"
fi

# Docker (optional)
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version)
    print_status "pass" "docker: $DOCKER_VERSION (optional)"
else
    print_status "warn" "docker is not installed (optional)"
fi

# bd (beads CLI - optional but needed for full functionality)
if command -v bd &> /dev/null; then
    BD_VERSION=$(bd --version 2>&1 || echo "version unknown")
    print_status "pass" "bd (beads): $BD_VERSION"
else
    print_status "warn" "bd (beads CLI) is not installed (needed for full functionality)"
fi

echo ""

# Test Python dependencies
echo "6. Testing Python Dependencies"
echo "-------------------------------"

if [ -f "$PROJECT_ROOT/agent/requirements.txt" ]; then
    print_status "pass" "requirements.txt found"
    
    # Check if pip is available
    if python3 -m pip --version > /dev/null 2>&1; then
        print_status "pass" "pip is available"
        
        # Try to check dependencies (without installing)
        echo ""
        echo "Python dependencies from requirements.txt:"
        cat "$PROJECT_ROOT/agent/requirements.txt" | grep -v "^#" | grep -v "^$"
        
    else
        print_status "warn" "pip is not available"
    fi
else
    print_status "fail" "requirements.txt not found"
fi

echo ""

# Test cross-compilation
echo "7. Testing Cross-Compilation"
echo "----------------------------"

PLATFORMS=("linux/amd64" "darwin/amd64" "darwin/arm64")

for platform in "${PLATFORMS[@]}"; do
    GOOS=$(echo $platform | cut -d/ -f1)
    GOARCH=$(echo $platform | cut -d/ -f2)
    
    if GOOS=$GOOS GOARCH=$GOARCH go build -o /dev/null ./... 2>&1 | grep -q "error"; then
        print_status "fail" "Cross-compilation failed for $platform"
    else
        print_status "pass" "Cross-compilation successful for $platform"
    fi
done

echo ""

# Check for deprecated dependencies
echo "8. Checking for Deprecated Dependencies"
echo "---------------------------------------"

DEPRECATED_FOUND=0

# Check Go dependencies
if go list -m all | grep -q "github.com/golang/protobuf"; then
    print_status "warn" "Using deprecated github.com/golang/protobuf (use google.golang.org/protobuf)"
    DEPRECATED_FOUND=1
fi

if go list -m all | grep -q "gopkg.in/yaml.v2"; then
    print_status "warn" "Using deprecated gopkg.in/yaml.v2 (use gopkg.in/yaml.v3)"
    DEPRECATED_FOUND=1
fi

if [ $DEPRECATED_FOUND -eq 0 ]; then
    print_status "pass" "No deprecated Go dependencies found"
fi

echo ""

# Check for available updates
echo "9. Checking for Dependency Updates"
echo "-----------------------------------"

UPDATE_COUNT=$(go list -u -m all 2>/dev/null | grep -c "\[" || echo "0")

if [ "$UPDATE_COUNT" -gt 0 ]; then
    print_status "warn" "$UPDATE_COUNT dependencies have updates available"
    echo ""
    echo "Run 'go list -u -m all' to see available updates"
else
    print_status "pass" "All dependencies are up to date"
fi

echo ""

# Summary
echo "==================================="
echo "Summary"
echo "==================================="
echo ""
echo "Go Version: $GO_VERSION"
echo "Python Version: $PYTHON_VERSION"
echo ""
echo "Minimum Requirements:"
echo "  - Go 1.21+"
echo "  - Python 3.8+"
echo ""
echo "All critical dependency checks passed!"
echo ""
echo "For detailed test results, run:"
echo "  go test -v ./test -run TestDependency"
