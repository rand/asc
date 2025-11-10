#!/bin/bash

# Performance profiling script for asc
# This script runs various performance tests and generates profiles

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROFILE_DIR="$PROJECT_ROOT/profiles"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create profiles directory
mkdir -p "$PROFILE_DIR"

echo -e "${BLUE}=== ASC Performance Profiling ===${NC}\n"

# Function to run benchmarks
run_benchmarks() {
    echo -e "${YELLOW}Running benchmarks...${NC}"
    cd "$PROJECT_ROOT"
    
    # Run all benchmarks with memory stats
    go test -bench=. -benchmem -benchtime=3s ./... | tee "$PROFILE_DIR/benchmark-results.txt"
    
    echo -e "${GREEN}✓ Benchmarks complete${NC}\n"
}

# Function to generate CPU profile
generate_cpu_profile() {
    echo -e "${YELLOW}Generating CPU profile...${NC}"
    cd "$PROJECT_ROOT"
    
    # Profile TUI rendering
    go test -bench=BenchmarkTUIRendering -cpuprofile="$PROFILE_DIR/cpu-tui.prof" -benchtime=10s ./test
    
    # Profile config loading
    go test -bench=BenchmarkLoad -cpuprofile="$PROFILE_DIR/cpu-config.prof" -benchtime=10s ./internal/config
    
    # Profile process management
    go test -bench=BenchmarkStart -cpuprofile="$PROFILE_DIR/cpu-process.prof" -benchtime=5s ./internal/process
    
    echo -e "${GREEN}✓ CPU profiles generated${NC}"
    echo -e "  - $PROFILE_DIR/cpu-tui.prof"
    echo -e "  - $PROFILE_DIR/cpu-config.prof"
    echo -e "  - $PROFILE_DIR/cpu-process.prof\n"
}

# Function to generate memory profile
generate_memory_profile() {
    echo -e "${YELLOW}Generating memory profile...${NC}"
    cd "$PROJECT_ROOT"
    
    # Profile TUI rendering
    go test -bench=BenchmarkTUIRenderingLargeDataset -memprofile="$PROFILE_DIR/mem-tui.prof" -benchtime=10s ./test
    
    # Profile log aggregation
    go test -bench=BenchmarkAggregatorMemoryUsage -memprofile="$PROFILE_DIR/mem-logger.prof" -benchtime=5s ./internal/logger
    
    echo -e "${GREEN}✓ Memory profiles generated${NC}"
    echo -e "  - $PROFILE_DIR/mem-tui.prof"
    echo -e "  - $PROFILE_DIR/mem-logger.prof\n"
}

# Function to run performance tests
run_performance_tests() {
    echo -e "${YELLOW}Running performance tests...${NC}"
    cd "$PROJECT_ROOT"
    
    # Run performance regression tests
    go test -v -run=TestPerformanceRegression ./test | tee "$PROFILE_DIR/performance-tests.txt"
    
    # Run memory usage tests
    go test -v -run=TestMemoryUsageUnderLoad ./test | tee -a "$PROFILE_DIR/performance-tests.txt"
    
    # Run startup/shutdown tests
    go test -v -run=TestStartupTime ./test | tee -a "$PROFILE_DIR/performance-tests.txt"
    go test -v -run=TestShutdownTime ./test | tee -a "$PROFILE_DIR/performance-tests.txt"
    
    # Run large dataset tests
    go test -v -run=TestLargeDatasetPerformance ./test | tee -a "$PROFILE_DIR/performance-tests.txt"
    
    echo -e "${GREEN}✓ Performance tests complete${NC}\n"
}

# Function to analyze profiles
analyze_profiles() {
    echo -e "${YELLOW}Analyzing profiles...${NC}"
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: go command not found${NC}"
        return 1
    fi
    
    # Generate top CPU consumers
    if [ -f "$PROFILE_DIR/cpu-tui.prof" ]; then
        echo -e "\n${BLUE}Top CPU consumers (TUI):${NC}"
        go tool pprof -top -nodecount=10 "$PROFILE_DIR/cpu-tui.prof" 2>/dev/null || true
    fi
    
    # Generate top memory allocators
    if [ -f "$PROFILE_DIR/mem-tui.prof" ]; then
        echo -e "\n${BLUE}Top memory allocators (TUI):${NC}"
        go tool pprof -top -nodecount=10 -alloc_space "$PROFILE_DIR/mem-tui.prof" 2>/dev/null || true
    fi
    
    echo -e "${GREEN}✓ Profile analysis complete${NC}\n"
}

# Function to generate report
generate_report() {
    echo -e "${YELLOW}Generating performance report...${NC}"
    
    REPORT_FILE="$PROFILE_DIR/performance-report.md"
    
    cat > "$REPORT_FILE" << EOF
# Performance Report

Generated: $(date)

## Benchmark Results

\`\`\`
$(cat "$PROFILE_DIR/benchmark-results.txt" 2>/dev/null || echo "No benchmark results available")
\`\`\`

## Performance Tests

\`\`\`
$(cat "$PROFILE_DIR/performance-tests.txt" 2>/dev/null || echo "No performance test results available")
\`\`\`

## CPU Profile Analysis

### TUI Rendering

\`\`\`
$(go tool pprof -top -nodecount=20 "$PROFILE_DIR/cpu-tui.prof" 2>/dev/null || echo "No CPU profile available")
\`\`\`

### Configuration Loading

\`\`\`
$(go tool pprof -top -nodecount=20 "$PROFILE_DIR/cpu-config.prof" 2>/dev/null || echo "No CPU profile available")
\`\`\`

## Memory Profile Analysis

### TUI Rendering

\`\`\`
$(go tool pprof -top -nodecount=20 -alloc_space "$PROFILE_DIR/mem-tui.prof" 2>/dev/null || echo "No memory profile available")
\`\`\`

### Log Aggregation

\`\`\`
$(go tool pprof -top -nodecount=20 -alloc_space "$PROFILE_DIR/mem-logger.prof" 2>/dev/null || echo "No memory profile available")
\`\`\`

## Recommendations

1. Review top CPU consumers and optimize hot paths
2. Check memory allocations and reduce unnecessary allocations
3. Compare results with baseline metrics in docs/PERFORMANCE.md
4. Address any performance regressions

## Profile Files

- CPU Profiles: $PROFILE_DIR/cpu-*.prof
- Memory Profiles: $PROFILE_DIR/mem-*.prof

To analyze interactively:
\`\`\`bash
go tool pprof $PROFILE_DIR/cpu-tui.prof
go tool pprof -alloc_space $PROFILE_DIR/mem-tui.prof
\`\`\`

EOF
    
    echo -e "${GREEN}✓ Report generated: $REPORT_FILE${NC}\n"
}

# Function to open interactive profiler
open_interactive() {
    echo -e "${BLUE}Opening interactive profiler...${NC}"
    echo -e "Available profiles:"
    echo -e "  1. CPU - TUI Rendering"
    echo -e "  2. CPU - Config Loading"
    echo -e "  3. CPU - Process Management"
    echo -e "  4. Memory - TUI Rendering"
    echo -e "  5. Memory - Log Aggregation"
    echo -e ""
    read -p "Select profile (1-5): " choice
    
    case $choice in
        1) go tool pprof "$PROFILE_DIR/cpu-tui.prof" ;;
        2) go tool pprof "$PROFILE_DIR/cpu-config.prof" ;;
        3) go tool pprof "$PROFILE_DIR/cpu-process.prof" ;;
        4) go tool pprof -alloc_space "$PROFILE_DIR/mem-tui.prof" ;;
        5) go tool pprof -alloc_space "$PROFILE_DIR/mem-logger.prof" ;;
        *) echo -e "${RED}Invalid choice${NC}" ;;
    esac
}

# Main menu
show_menu() {
    echo -e "${BLUE}Performance Profiling Options:${NC}"
    echo -e "  1. Run all (benchmarks + profiles + tests + report)"
    echo -e "  2. Run benchmarks only"
    echo -e "  3. Generate CPU profiles"
    echo -e "  4. Generate memory profiles"
    echo -e "  5. Run performance tests"
    echo -e "  6. Analyze existing profiles"
    echo -e "  7. Generate report"
    echo -e "  8. Open interactive profiler"
    echo -e "  9. Clean profile directory"
    echo -e "  0. Exit"
    echo -e ""
}

# Parse command line arguments
if [ $# -eq 0 ]; then
    # Interactive mode
    while true; do
        show_menu
        read -p "Select option: " option
        
        case $option in
            1)
                run_benchmarks
                generate_cpu_profile
                generate_memory_profile
                run_performance_tests
                analyze_profiles
                generate_report
                ;;
            2) run_benchmarks ;;
            3) generate_cpu_profile ;;
            4) generate_memory_profile ;;
            5) run_performance_tests ;;
            6) analyze_profiles ;;
            7) generate_report ;;
            8) open_interactive ;;
            9)
                echo -e "${YELLOW}Cleaning profile directory...${NC}"
                rm -rf "$PROFILE_DIR"/*
                echo -e "${GREEN}✓ Profile directory cleaned${NC}\n"
                ;;
            0)
                echo -e "${GREEN}Goodbye!${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}Invalid option${NC}\n"
                ;;
        esac
    done
else
    # Command line mode
    case "$1" in
        all)
            run_benchmarks
            generate_cpu_profile
            generate_memory_profile
            run_performance_tests
            analyze_profiles
            generate_report
            ;;
        bench) run_benchmarks ;;
        cpu) generate_cpu_profile ;;
        mem) generate_memory_profile ;;
        test) run_performance_tests ;;
        analyze) analyze_profiles ;;
        report) generate_report ;;
        interactive) open_interactive ;;
        clean)
            rm -rf "$PROFILE_DIR"/*
            echo -e "${GREEN}✓ Profile directory cleaned${NC}"
            ;;
        *)
            echo -e "${RED}Usage: $0 [all|bench|cpu|mem|test|analyze|report|interactive|clean]${NC}"
            exit 1
            ;;
    esac
fi

echo -e "${GREEN}=== Profiling Complete ===${NC}"
