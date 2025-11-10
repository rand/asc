#!/bin/bash

# Test Timing Analyzer
# Analyzes test execution times and identifies slow tests

set -e

OUTPUT_FILE="test-timing-analysis.md"
JSON_OUTPUT="test-output.json"

echo "⏱️  Test Timing Analyzer"
echo "========================"
echo ""

# Run tests with JSON output
echo "Running tests with timing information..."
go test -v -json ./... 2>&1 | tee "$JSON_OUTPUT"

echo ""
echo "Analyzing test timing..."

# Create Go program to analyze timing
cat > /tmp/analyze_timing.go << 'EOF'
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "sort"
    "time"
)

type TestEvent struct {
    Time    time.Time
    Action  string
    Package string
    Test    string
    Elapsed float64
}

type TestTiming struct {
    Name    string
    Package string
    Elapsed float64
}

func main() {
    file, err := os.Open(os.Args[1])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
        os.Exit(1)
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file)
    timings := []TestTiming{}
    packageTimes := make(map[string]float64)
    
    for scanner.Scan() {
        var event TestEvent
        if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
            continue
        }
        
        if event.Action == "pass" || event.Action == "fail" {
            if event.Test != "" && event.Elapsed > 0 {
                timings = append(timings, TestTiming{
                    Name:    event.Test,
                    Package: event.Package,
                    Elapsed: event.Elapsed,
                })
            }
            
            if event.Test == "" && event.Elapsed > 0 {
                packageTimes[event.Package] = event.Elapsed
            }
        }
    }
    
    // Sort by elapsed time (slowest first)
    sort.Slice(timings, func(i, j int) bool {
        return timings[i].Elapsed > timings[j].Elapsed
    })
    
    fmt.Println("# Test Timing Analysis")
    fmt.Println()
    fmt.Printf("**Generated**: %s\n", time.Now().Format("2006-01-02 15:04:05"))
    fmt.Println()
    
    // Overall statistics
    var totalTime float64
    for _, t := range timings {
        totalTime += t.Elapsed
    }
    
    fmt.Println("## Summary")
    fmt.Println()
    fmt.Printf("- **Total tests**: %d\n", len(timings))
    fmt.Printf("- **Total time**: %.2fs\n", totalTime)
    if len(timings) > 0 {
        fmt.Printf("- **Average time**: %.3fs\n", totalTime/float64(len(timings)))
        fmt.Printf("- **Slowest test**: %.3fs (%s)\n", timings[0].Elapsed, timings[0].Name)
        fmt.Printf("- **Fastest test**: %.3fs (%s)\n", timings[len(timings)-1].Elapsed, timings[len(timings)-1].Name)
    }
    fmt.Println()
    
    // Slowest tests
    fmt.Println("## Slowest Tests (Top 20)")
    fmt.Println()
    fmt.Println("| Rank | Test | Package | Duration |")
    fmt.Println("|------|------|---------|----------|")
    
    count := 20
    if len(timings) < count {
        count = len(timings)
    }
    
    for i := 0; i < count; i++ {
        fmt.Printf("| %d | %s | %s | %.3fs |\n",
            i+1,
            timings[i].Name,
            timings[i].Package,
            timings[i].Elapsed)
    }
    fmt.Println()
    
    // Package timing
    fmt.Println("## Package Timing")
    fmt.Println()
    
    type PackageTiming struct {
        Package string
        Time    float64
    }
    
    var pkgTimings []PackageTiming
    for pkg, time := range packageTimes {
        pkgTimings = append(pkgTimings, PackageTiming{pkg, time})
    }
    
    sort.Slice(pkgTimings, func(i, j int) bool {
        return pkgTimings[i].Time > pkgTimings[j].Time
    })
    
    fmt.Println("| Package | Duration |")
    fmt.Println("|---------|----------|")
    for _, pt := range pkgTimings {
        fmt.Printf("| %s | %.2fs |\n", pt.Package, pt.Time)
    }
    fmt.Println()
    
    // Warnings for slow tests
    fmt.Println("## Performance Warnings")
    fmt.Println()
    
    slowCount := 0
    verySlowCount := 0
    
    for _, t := range timings {
        if t.Elapsed > 10.0 {
            fmt.Printf("🔴 **Very slow test**: %s (%.2fs)\n", t.Name, t.Elapsed)
            verySlowCount++
        } else if t.Elapsed > 5.0 {
            fmt.Printf("🟡 **Slow test**: %s (%.2fs)\n", t.Name, t.Elapsed)
            slowCount++
        }
    }
    
    if verySlowCount == 0 && slowCount == 0 {
        fmt.Println("✅ No slow tests detected (all tests < 5s)")
    } else {
        fmt.Println()
        fmt.Printf("- Very slow tests (>10s): %d\n", verySlowCount)
        fmt.Printf("- Slow tests (5-10s): %d\n", slowCount)
    }
    fmt.Println()
    
    // Recommendations
    fmt.Println("## Recommendations")
    fmt.Println()
    
    if verySlowCount > 0 || slowCount > 0 {
        fmt.Println("### For Slow Tests")
        fmt.Println()
        fmt.Println("1. **Profile the test**: Use `-cpuprofile` to identify bottlenecks")
        fmt.Println("2. **Reduce test scope**: Break large tests into smaller units")
        fmt.Println("3. **Mock expensive operations**: Replace real I/O with mocks")
        fmt.Println("4. **Parallelize**: Use `t.Parallel()` for independent tests")
        fmt.Println("5. **Consider integration tests**: Move slow tests to integration suite")
        fmt.Println()
    }
    
    fmt.Println("### General Best Practices")
    fmt.Println()
    fmt.Println("- Unit tests should complete in < 1s")
    fmt.Println("- Integration tests should complete in < 10s")
    fmt.Println("- Use table-driven tests to reduce setup overhead")
    fmt.Println("- Cache expensive setup in `TestMain`")
    fmt.Println("- Run slow tests in parallel when possible")
    fmt.Println()
    
    // Time distribution
    fmt.Println("## Time Distribution")
    fmt.Println()
    
    var under1s, under5s, under10s, over10s int
    for _, t := range timings {
        if t.Elapsed < 1.0 {
            under1s++
        } else if t.Elapsed < 5.0 {
            under5s++
        } else if t.Elapsed < 10.0 {
            under10s++
        } else {
            over10s++
        }
    }
    
    fmt.Println("| Duration | Count | Percentage |")
    fmt.Println("|----------|-------|------------|")
    fmt.Printf("| < 1s | %d | %.1f%% |\n", under1s, float64(under1s)*100/float64(len(timings)))
    fmt.Printf("| 1-5s | %d | %.1f%% |\n", under5s, float64(under5s)*100/float64(len(timings)))
    fmt.Printf("| 5-10s | %d | %.1f%% |\n", under10s, float64(under10s)*100/float64(len(timings)))
    fmt.Printf("| > 10s | %d | %.1f%% |\n", over10s, float64(over10s)*100/float64(len(timings)))
    
    // Exit with warning if slow tests found
    if verySlowCount > 0 || slowCount > 5 {
        os.Exit(1)
    }
}
EOF

# Run the analyzer
go run /tmp/analyze_timing.go "$JSON_OUTPUT" > "$OUTPUT_FILE"

# Display the report
cat "$OUTPUT_FILE"

echo ""
echo "📊 Full report saved to: $OUTPUT_FILE"
echo ""

# Check exit code
if [ $? -eq 0 ]; then
    echo "✅ Test timing is acceptable"
    exit 0
else
    echo "⚠️  Slow tests detected - consider optimization"
    exit 1
fi
