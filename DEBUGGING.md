# Debugging Guide

This guide covers debugging techniques and tools for developing and troubleshooting asc.

## Table of Contents

- [Quick Debugging Tips](#quick-debugging-tips)
- [Debugging Go Code](#debugging-go-code)
- [Debugging TUI Applications](#debugging-tui-applications)
- [Debugging Agent Processes](#debugging-agent-processes)
- [Analyzing Logs](#analyzing-logs)
- [Performance Profiling](#performance-profiling)
- [Common Issues](#common-issues)

## Quick Debugging Tips

### Enable Debug Mode

Run asc with verbose logging:

```bash
# Set log level to debug
export ASC_LOG_LEVEL=debug
asc up

# Or use the debug flag (if implemented)
asc up --debug
```

### Check Logs

```bash
# View main asc logs
tail -f ~/.asc/logs/asc.log

# View specific agent logs
tail -f ~/.asc/logs/agent-name.log

# View all logs
tail -f ~/.asc/logs/*.log

# Search logs for errors
grep -i error ~/.asc/logs/*.log

# View logs with timestamps
tail -f ~/.asc/logs/asc.log | while read line; do echo "$(date '+%Y-%m-%d %H:%M:%S') $line"; done
```

### Check Process Status

```bash
# List all asc-related processes
ps aux | grep -E '(asc|agent_adapter|mcp_agent_mail)'

# Check PID files
ls -la ~/.asc/pids/
cat ~/.asc/pids/agents.json

# Check if ports are in use
lsof -i :8765  # MCP server port
```

## Debugging Go Code

### Using Delve Debugger

Install Delve:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

#### Debug a Command

```bash
# Debug the 'up' command
dlv debug . -- up

# Set breakpoints
(dlv) break main.main
(dlv) break cmd/up.go:45

# Run the program
(dlv) continue

# Inspect variables
(dlv) print variableName
(dlv) locals
(dlv) args

# Step through code
(dlv) next      # Next line
(dlv) step      # Step into function
(dlv) stepout   # Step out of function

# View stack trace
(dlv) stack

# List source code
(dlv) list
```

#### Debug Tests

```bash
# Debug a specific test
dlv test ./internal/config -- -test.run TestParseConfig

# Set breakpoint in test
(dlv) break TestParseConfig
(dlv) continue
```

#### Remote Debugging

```bash
# Start debug server
dlv debug --headless --listen=:2345 --api-version=2 . -- up

# Connect from another terminal
dlv connect :2345
```

### Using Print Debugging

Sometimes simple print statements are fastest:

```go
import "fmt"

func debugFunction() {
    fmt.Printf("DEBUG: variable = %+v\n", variable)
    fmt.Printf("DEBUG: entering function at %s\n", time.Now())
}
```

For production code, use the logger:

```go
import "github.com/yourusername/asc/internal/logger"

func debugFunction() {
    logger.Debug("variable value", "var", variable)
    logger.Debug("entering function")
}
```

### Using Go's Built-in Tools

#### Race Detector

Detect race conditions:

```bash
# Run with race detector
go run -race . up

# Test with race detector
go test -race ./...

# Build with race detector
go build -race -o asc-race .
./asc-race up
```

#### Memory Sanitizer

Detect memory issues:

```bash
# Run with memory sanitizer (requires CGO)
go run -msan . up
```

## Debugging TUI Applications

TUI applications are tricky to debug because they take over the terminal.

### Method 1: Log to File

The easiest approach - log everything to a file:

```go
// In your TUI code
logFile, _ := os.OpenFile("/tmp/asc-tui-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
defer logFile.Close()

fmt.Fprintf(logFile, "DEBUG: %s - %+v\n", time.Now(), data)
```

Watch the log in another terminal:

```bash
tail -f /tmp/asc-tui-debug.log
```

### Method 2: Use Two Terminals

Run the TUI in one terminal, watch logs in another:

Terminal 1:
```bash
asc up
```

Terminal 2:
```bash
tail -f ~/.asc/logs/asc.log
```

### Method 3: Delve with TTY

Use Delve with a separate TTY:

```bash
# Terminal 1: Start debug server
dlv debug --headless --listen=:2345 . -- up

# Terminal 2: Connect and debug
dlv connect :2345
```

### Method 4: Debug TUI Rendering

Create a test that renders to a buffer:

```go
func TestTUIRendering(t *testing.T) {
    m := NewModel()
    m.width = 80
    m.height = 24
    
    view := m.View()
    
    // Print to see what's rendered
    fmt.Println(view)
    
    // Or write to file
    os.WriteFile("/tmp/tui-render.txt", []byte(view), 0644)
}
```

### Common TUI Issues

**TUI not rendering:**
- Check terminal size: `echo $COLUMNS $LINES`
- Check TERM variable: `echo $TERM`
- Try different terminal emulator

**Colors not showing:**
- Ensure terminal supports 256 colors
- Check TERM=xterm-256color
- Test with: `tput colors`

**Input not working:**
- Check if another process is reading stdin
- Verify terminal is in raw mode
- Check for conflicting key bindings

## Debugging Agent Processes

### View Agent Logs

```bash
# View specific agent log
tail -f ~/.asc/logs/agent-name.log

# View with context
tail -n 100 ~/.asc/logs/agent-name.log

# Search for errors
grep -A 5 -B 5 "error" ~/.asc/logs/agent-name.log
```

### Attach to Running Agent

```bash
# Find agent PID
ps aux | grep agent_adapter

# Attach debugger (Python)
python -m pdb -p <PID>

# Or use py-spy for profiling
py-spy top --pid <PID>
py-spy record --pid <PID> -o profile.svg
```

### Test Agent in Isolation

```bash
# Run agent manually with debug output
export AGENT_NAME=test-agent
export AGENT_MODEL=claude
export AGENT_PHASES=planning
export MCP_MAIL_URL=http://localhost:8765
export BEADS_DB_PATH=./test-repo
export CLAUDE_API_KEY=your-key
export ASC_LOG_LEVEL=debug

python agent/agent_adapter.py
```

### Monitor Agent Communication

```bash
# Watch MCP traffic
curl http://localhost:8765/messages | jq

# Watch beads updates
watch -n 1 'bd list --json | jq'

# Monitor file leases
curl http://localhost:8765/leases | jq
```

## Analyzing Logs

### Log Structure

asc logs are structured JSON (when using the logger package):

```json
{
  "timestamp": "2025-11-10T10:30:15Z",
  "level": "info",
  "message": "agent started",
  "agent": "main-planner",
  "pid": 12345
}
```

### Useful Log Queries

```bash
# Extract all errors
jq 'select(.level == "error")' ~/.asc/logs/asc.log

# Filter by agent
jq 'select(.agent == "main-planner")' ~/.asc/logs/asc.log

# Count errors by type
jq -r 'select(.level == "error") | .message' ~/.asc/logs/asc.log | sort | uniq -c

# Timeline of events
jq -r '[.timestamp, .level, .message] | @tsv' ~/.asc/logs/asc.log

# Find slow operations
jq 'select(.duration > 1000)' ~/.asc/logs/asc.log
```

### Log Analysis Tools

```bash
# Install lnav (log file navigator)
brew install lnav  # macOS
apt install lnav   # Linux

# View logs with lnav
lnav ~/.asc/logs/*.log

# Search in lnav: press '/'
# Filter in lnav: press 'i'
# View errors only: press 'e'
```

## Performance Profiling

### CPU Profiling

```bash
# Profile a test
go test -cpuprofile=cpu.prof -bench=. ./internal/tui

# Analyze profile
go tool pprof cpu.prof

# In pprof:
(pprof) top10        # Top 10 functions by CPU time
(pprof) list FuncName # Show source code with timing
(pprof) web          # Open in browser (requires graphviz)
```

### Memory Profiling

```bash
# Profile memory allocation
go test -memprofile=mem.prof -bench=. ./internal/tui

# Analyze profile
go tool pprof mem.prof

# In pprof:
(pprof) top10
(pprof) list FuncName
```

### Live Profiling

```bash
# Add pprof endpoint to your code
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

# Then profile the running application
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Trace Analysis

```bash
# Generate trace
go test -trace=trace.out ./internal/tui

# View trace
go tool trace trace.out
```

### Benchmark Comparison

```bash
# Run benchmarks and save results
go test -bench=. -benchmem ./internal/config > old.txt

# Make changes, then run again
go test -bench=. -benchmem ./internal/config > new.txt

# Compare results
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

## Common Issues

### Issue: "Cannot connect to MCP server"

**Debug steps:**

1. Check if server is running:
```bash
ps aux | grep mcp_agent_mail
lsof -i :8765
```

2. Check server logs:
```bash
tail -f ~/.asc/logs/mcp_agent_mail.log
```

3. Test connection manually:
```bash
curl http://localhost:8765/health
```

4. Check firewall:
```bash
# macOS
sudo pfctl -s rules | grep 8765

# Linux
sudo iptables -L | grep 8765
```

### Issue: "Agent stuck in working state"

**Debug steps:**

1. Check agent logs:
```bash
tail -f ~/.asc/logs/agent-name.log
```

2. Check if process is alive:
```bash
ps aux | grep agent-name
```

3. Check what the agent is doing:
```bash
# macOS
sample <PID> 10

# Linux
perf record -p <PID> -g -- sleep 10
perf report
```

4. Check file leases:
```bash
curl http://localhost:8765/leases | jq
```

### Issue: "High CPU usage"

**Debug steps:**

1. Profile CPU:
```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

2. Check for busy loops:
```bash
# Look for tight loops in code
grep -r "for {" .
```

3. Check goroutine count:
```bash
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

### Issue: "Memory leak"

**Debug steps:**

1. Profile memory:
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

2. Check goroutine leaks:
```bash
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

3. Look for unclosed resources:
```bash
# Check file descriptors
lsof -p <PID> | wc -l

# Monitor over time
watch -n 1 'lsof -p <PID> | wc -l'
```

### Issue: "Tests failing randomly"

**Debug steps:**

1. Run with race detector:
```bash
go test -race -count=100 ./...
```

2. Run in parallel:
```bash
go test -parallel=10 -count=100 ./...
```

3. Check for shared state:
```bash
# Look for global variables
grep -r "var.*=" . | grep -v "_test.go"
```

4. Add test logging:
```go
func TestFlaky(t *testing.T) {
    t.Logf("Starting test at %s", time.Now())
    // ... test code ...
    t.Logf("Finished test at %s", time.Now())
}
```

## Debugging Checklist

When debugging an issue:

- [ ] Can you reproduce it consistently?
- [ ] What changed recently?
- [ ] Check the logs
- [ ] Check process status
- [ ] Verify configuration
- [ ] Test in isolation
- [ ] Add debug logging
- [ ] Use a debugger
- [ ] Profile if performance-related
- [ ] Check for race conditions
- [ ] Verify dependencies are correct versions
- [ ] Test with minimal configuration
- [ ] Check for resource leaks

## Getting Help

If you're stuck:

1. **Search existing issues** - Someone may have had the same problem
2. **Check documentation** - Review relevant docs
3. **Ask in discussions** - Post in GitHub Discussions
4. **Create an issue** - If it's a bug, open an issue with:
   - Steps to reproduce
   - Expected vs actual behavior
   - Logs and error messages
   - Environment details (OS, Go version, etc.)

## Resources

- [Delve Documentation](https://github.com/go-delve/delve/tree/master/Documentation)
- [Go Diagnostics](https://golang.org/doc/diagnostics)
- [pprof Tutorial](https://blog.golang.org/pprof)
- [Debugging Go Programs](https://golang.org/doc/gdb)
- [Effective Go](https://golang.org/doc/effective_go)

---

Happy debugging! Remember: bugs are just features waiting to be understood. 🐛
