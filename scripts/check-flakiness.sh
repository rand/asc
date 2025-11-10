#!/bin/bash

# Test Flakiness Checker
# Runs tests multiple times to detect flaky tests

set -e

# Configuration
RUNS=${1:-10}  # Number of times to run tests (default: 10)
OUTPUT_DIR="test-flakiness-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "🔍 Test Flakiness Checker"
echo "========================="
echo "Runs: $RUNS"
echo "Output: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Run tests multiple times
echo "Running tests $RUNS times..."
declare -A test_results
declare -A test_failures
total_runs=0
failed_runs=0

for i in $(seq 1 $RUNS); do
    echo -n "Run $i/$RUNS... "
    
    # Run tests and capture output
    if go test -v -json ./... > "$OUTPUT_DIR/run-$i.json" 2>&1; then
        echo "✅ PASS"
    else
        echo "❌ FAIL"
        failed_runs=$((failed_runs + 1))
    fi
    
    total_runs=$((total_runs + 1))
    
    # Parse results
    while IFS= read -r line; do
        test_name=$(echo "$line" | jq -r 'select(.Test != null) | .Test' 2>/dev/null || echo "")
        action=$(echo "$line" | jq -r 'select(.Action != null) | .Action' 2>/dev/null || echo "")
        
        if [ -n "$test_name" ] && [ "$action" = "fail" ]; then
            test_failures["$test_name"]=$((${test_failures["$test_name"]:-0} + 1))
        fi
        
        if [ -n "$test_name" ] && ([ "$action" = "pass" ] || [ "$action" = "fail" ]); then
            test_results["$test_name"]=$((${test_results["$test_name"]:-0} + 1))
        fi
    done < "$OUTPUT_DIR/run-$i.json"
done

echo ""
echo "Analysis Results"
echo "================"
echo ""

# Generate report
REPORT_FILE="$OUTPUT_DIR/flakiness-report-$TIMESTAMP.md"

cat > "$REPORT_FILE" << EOF
# Test Flakiness Report

**Generated**: $(date)  
**Total Runs**: $RUNS  
**Failed Runs**: $failed_runs  

## Summary

- Total test runs: $total_runs
- Runs with failures: $failed_runs
- Success rate: $(( (total_runs - failed_runs) * 100 / total_runs ))%

EOF

# Analyze flaky tests
echo "## Flaky Tests" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

flaky_count=0
for test in "${!test_failures[@]}"; do
    failures=${test_failures[$test]}
    total=${test_results[$test]:-0}
    
    # A test is flaky if it fails sometimes but not always
    if [ $failures -gt 0 ] && [ $failures -lt $total ]; then
        flakiness_rate=$(( failures * 100 / total ))
        echo "⚠️  **$test**" >> "$REPORT_FILE"
        echo "   - Failed: $failures/$total runs" >> "$REPORT_FILE"
        echo "   - Flakiness rate: $flakiness_rate%" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        flaky_count=$((flaky_count + 1))
        
        echo "⚠️  Flaky: $test ($failures/$total = $flakiness_rate%)"
    fi
done

if [ $flaky_count -eq 0 ]; then
    echo "✅ No flaky tests detected" >> "$REPORT_FILE"
    echo "✅ No flaky tests detected"
else
    echo "⚠️  Found $flaky_count flaky test(s)"
fi

echo "" >> "$REPORT_FILE"

# Analyze consistently failing tests
echo "## Consistently Failing Tests" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

consistent_count=0
for test in "${!test_failures[@]}"; do
    failures=${test_failures[$test]}
    total=${test_results[$test]:-0}
    
    # A test consistently fails if it fails every time
    if [ $failures -eq $total ] && [ $total -gt 0 ]; then
        echo "❌ **$test**" >> "$REPORT_FILE"
        echo "   - Failed: $failures/$total runs (100%)" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        consistent_count=$((consistent_count + 1))
        
        echo "❌ Consistent failure: $test ($failures/$total)"
    fi
done

if [ $consistent_count -eq 0 ]; then
    echo "✅ No consistently failing tests" >> "$REPORT_FILE"
    echo "✅ No consistently failing tests"
else
    echo "❌ Found $consistent_count consistently failing test(s)"
fi

echo "" >> "$REPORT_FILE"

# Add recommendations
cat >> "$REPORT_FILE" << EOF
## Recommendations

### For Flaky Tests

1. **Investigate race conditions**: Run with \`go test -race\`
2. **Check timing assumptions**: Look for hardcoded timeouts or sleeps
3. **Review external dependencies**: Mock or stub external services
4. **Add retries**: Consider adding retry logic for inherently flaky operations
5. **Increase timeouts**: If tests are timing out inconsistently

### For Consistently Failing Tests

1. **Fix immediately**: These tests are blocking development
2. **Check recent changes**: Review commits that may have broken tests
3. **Verify test environment**: Ensure all dependencies are available
4. **Update test expectations**: Tests may need updating for new behavior

### General Best Practices

- Keep tests deterministic and isolated
- Avoid shared state between tests
- Use proper synchronization primitives
- Mock time-dependent operations
- Clean up resources after tests

## Test Execution Details

EOF

# Add execution statistics
echo "### Execution Statistics" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "| Metric | Value |" >> "$REPORT_FILE"
echo "|--------|-------|" >> "$REPORT_FILE"
echo "| Total runs | $RUNS |" >> "$REPORT_FILE"
echo "| Successful runs | $((total_runs - failed_runs)) |" >> "$REPORT_FILE"
echo "| Failed runs | $failed_runs |" >> "$REPORT_FILE"
echo "| Unique tests | ${#test_results[@]} |" >> "$REPORT_FILE"
echo "| Tests with failures | ${#test_failures[@]} |" >> "$REPORT_FILE"
echo "| Flaky tests | $flaky_count |" >> "$REPORT_FILE"
echo "| Consistently failing | $consistent_count |" >> "$REPORT_FILE"

echo ""
echo "📊 Report saved to: $REPORT_FILE"
echo ""

# Display summary
cat "$REPORT_FILE"

# Exit with error if flaky or failing tests found
if [ $flaky_count -gt 0 ] || [ $consistent_count -gt 0 ]; then
    echo ""
    echo "⚠️  Flakiness or failures detected!"
    exit 1
else
    echo ""
    echo "✅ All tests are stable!"
    exit 0
fi
