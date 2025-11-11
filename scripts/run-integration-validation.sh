#!/bin/bash
# Integration Validation Script for Agent Stack Controller
# This script runs comprehensive integration validation tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "ASC Integration Validation Suite"
echo "========================================="
echo ""

# Check if build exists
if [ ! -f "build/asc" ]; then
    echo -e "${YELLOW}Building asc binary...${NC}"
    make build
    echo ""
fi

# Function to run a test and report results
run_test() {
    local test_name=$1
    local test_pattern=$2
    local skip_reason=$3
    
    echo -e "${YELLOW}Running: ${test_name}${NC}"
    
    if [ -n "$skip_reason" ]; then
        echo -e "${YELLOW}  SKIPPED: ${skip_reason}${NC}"
        echo ""
        return 0
    fi
    
    if go test -tags=integration ./test/integration_validation_test.go -v -run "$test_pattern" 2>&1 | tee /tmp/test_output.txt | grep -q "PASS"; then
        echo -e "${GREEN}  ✓ PASSED${NC}"
    else
        echo -e "${RED}  ✗ FAILED${NC}"
        cat /tmp/test_output.txt | grep -A 5 "FAIL:"
        FAILED_TESTS+=("$test_name")
    fi
    echo ""
}

# Track failed tests
FAILED_TESTS=()

echo "Phase 1: Basic Workflow Tests"
echo "------------------------------"
run_test "Init Workflow" "TestIntegrationValidation_InitWorkflow"
run_test "Config Hot-Reload" "TestIntegrationValidation_ConfigHotReload"
run_test "Config Templates" "TestIntegrationValidation_ConfigTemplates"
echo ""

echo "Phase 2: Process Management Tests"
echo "----------------------------------"
run_test "Up → Work → Down Workflow" "TestIntegrationValidation_UpWorkDown" "Requires INTEGRATION_FULL=true"
run_test "Multi-Agent Coordination" "TestIntegrationValidation_MultiAgentCoordination"
run_test "Error Recovery" "TestIntegrationValidation_ErrorRecovery"
echo ""

echo "Phase 3: Security Tests"
echo "-----------------------"
if command -v age &> /dev/null; then
    run_test "Secrets Encryption/Decryption" "TestIntegrationValidation_SecretsEncryptionDecryption"
else
    run_test "Secrets Encryption/Decryption" "" "age not installed"
fi
echo ""

echo "Phase 4: External Integration Tests"
echo "------------------------------------"
if command -v bd &> /dev/null; then
    run_test "Real Beads Repository" "TestIntegrationValidation_RealBeadsRepository"
else
    run_test "Real Beads Repository" "" "bd (beads) not installed"
fi

run_test "Real MCP Server" "TestIntegrationValidation_RealMCPServer" "Requires INTEGRATION_MCP=true and running MCP server"
run_test "Health Monitoring" "TestIntegrationValidation_HealthMonitoringAndRecovery" "Requires INTEGRATION_FULL=true"
echo ""

echo "Phase 5: Complete Workflow Tests"
echo "---------------------------------"
run_test "Complete Workflow" "TestIntegrationValidation_CompleteWorkflow" "Requires INTEGRATION_FULL=true"
echo ""

echo "Phase 6: Stress Tests"
echo "---------------------"
run_test "Stress Test" "TestIntegrationValidation_StressTest" "Requires INTEGRATION_STRESS=true"
echo ""

# Summary
echo "========================================="
echo "Integration Validation Summary"
echo "========================================="
echo ""

if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo ""
    echo "Integration validation completed successfully."
    exit 0
else
    echo -e "${RED}✗ ${#FAILED_TESTS[@]} test(s) failed:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "${RED}  - $test${NC}"
    done
    echo ""
    echo "Please review the failures above and fix any issues."
    exit 1
fi
