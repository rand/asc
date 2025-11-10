# Logging and Debugging Guide

This document describes the enhanced logging and debugging features in the Agent Stack Controller (asc).

## Overview

The asc tool now includes comprehensive logging and debugging capabilities:

1. **Structured Logging** - JSON and text format logs with context fields
2. **Debug Mode** - Verbose output for troubleshooting
3. **Log Aggregation** - Unified view of logs from all agents
4. **Log Export** - Export filtered logs to files
5. **Log Cleanup** - Automatic removal of old log files

## Structured Logging

### Log Formats

The logger supports two output formats:

- **Text Format** (default): Human-readable logs with context fields
- **JSON Format**: Machine-parseable logs for analysis tools

Example text format:
```
[2025-11-10 10:00:00.000] [INFO] Starting agent {agent=planner, model=gemini}
```

Example JSON format:
```json
{
  "timestamp": "2025-11-10 10:00:00.000",
  "level": "INFO",
  "message": "Starting agent",
  "correlation_id": "abc-123",
  "agent": "planner",
  "task": "task-456",
  "phase": "planning",
  "fields": {
    "model": "gemini"
  }
}
```

### Context Fields

Logs can include structured context fields:

- `agent`: Agent name
- `task`: Task ID
- `phase`: Workflow phase
- `correlation_id`: Request tracing ID
- Custom fields: Any additional metadata

### Usage in Code

```go
import "github.com/yourusername/asc/internal/logger"

// Initialize logger
logger.Init()
defer logger.Close()

// Simple logging
logger.Info("Starting process")
logger.Error("Failed to connect: %v", err)

// Structured logging with fields
logger.WithFields(logger.Fields{
    "agent": "planner",
    "task": "task-123",
    "phase": "planning",
}).Info("Processing task")

// Set correlation ID for request tracing
logger.WithCorrelationID("request-abc-123")

// Set persistent context fields
logger.SetContextFields(logger.Fields{
    "service": "asc",
    "version": "1.0.0",
})
```

### Log Levels

- `DEBUG`: Detailed diagnostic information
- `INFO`: General informational messages
- `WARN`: Warning messages for potential issues
- `ERROR`: Error messages for failures

## Debug Mode

### Enabling Debug Mode

Start asc with the `--debug` flag to enable verbose output:

```bash
asc up --debug
```

Debug mode enables:

- JSON-formatted logs for machine parsing
- DEBUG level logging (shows all log messages)
- Detailed logging of:
  - LLM prompts and responses (in agent logs)
  - File lease operations and conflicts
  - Beads database queries and git operations
  - MCP server communication
  - Process management operations

### Debug Indicator

When debug mode is enabled, the TUI footer displays a `[DEBUG]` indicator.

### Debug Logging Examples

With debug mode enabled, you'll see detailed logs like:

```json
{
  "timestamp": "2025-11-10 10:00:00.000",
  "level": "DEBUG",
  "message": "Executing beads query",
  "fields": {
    "command": "bd",
    "args": ["--json", "list", "--status", "open,in_progress"],
    "db_path": "./project-repo"
  }
}
```

```json
{
  "timestamp": "2025-11-10 10:00:01.000",
  "level": "DEBUG",
  "message": "Beads query completed successfully",
  "fields": {
    "task_count": 5,
    "statuses": ["open", "in_progress"]
  }
}
```

## Log Aggregation

### Overview

The log aggregator collects logs from all sources into a unified view:

- Main asc logs (`~/.asc/logs/asc.log`)
- Agent logs (`~/.asc/logs/<agent-name>.log`)
- Service logs (`~/.asc/logs/mcp_agent_mail.log`)

### Features

- **Unified View**: See logs from all agents in one place
- **Filtering**: Filter by agent name, log level, time range, or search text
- **Sorting**: Logs are sorted by timestamp (newest first)
- **Statistics**: View log counts by level and source

### Usage in TUI

The TUI automatically aggregates logs and displays them in the log pane. Use the following keys:

- `/`: Enter search mode to filter logs
- `e`: Export filtered logs to a file

### Programmatic Usage

```go
import "github.com/yourusername/asc/internal/logger"

// Create aggregator
aggregator := logger.NewLogAggregator("~/.asc/logs", 1000)

// Collect logs from all files
aggregator.CollectLogs()

// Get filtered logs
filters := logger.LogFilters{
    AgentName:  "planner",
    Level:      logger.INFO,
    SearchText: "error",
}
logs := aggregator.GetFilteredLogs(filters)

// Get statistics
stats := aggregator.GetStats()
fmt.Printf("Total entries: %d\n", stats.TotalEntries)
fmt.Printf("INFO: %d, ERROR: %d\n", stats.ByLevel["INFO"], stats.ByLevel["ERROR"])
```

## Log Export

### Exporting from TUI

Press `e` in the TUI to export filtered logs to a file. The file will be named:

```
asc-logs-YYYYMMDD-HHMMSS.txt
```

The export includes all logs matching the current filters (agent name, search text, etc.).

### Programmatic Export

```go
// Export logs to file
filters := logger.LogFilters{
    AgentName: "planner",
    Level:     logger.WARN,
}
err := aggregator.ExportToFile("export.txt", filters)
```

### Export Format

Exported logs are in text format:

```
[2025-11-10 10:00:00.000] [INFO] [asc] Starting agent stack
[2025-11-10 10:00:01.000] [INFO] [planner] Processing task
[2025-11-10 10:00:02.000] [ERROR] [planner] Task failed
```

## Log Cleanup

### Automatic Cleanup

Use the `cleanup` command to remove old log files:

```bash
# Remove logs older than 30 days (default)
asc cleanup

# Remove logs older than 7 days
asc cleanup --days 7

# Dry run (show what would be deleted)
asc cleanup --dry-run
```

### Programmatic Cleanup

```go
import "github.com/yourusername/asc/internal/logger"

// Remove logs older than 30 days
maxAge := 30 * 24 * time.Hour
err := logger.CleanupOldLogs("~/.asc/logs", maxAge)
```

## Log Rotation

Logs are automatically rotated when they exceed the maximum size:

- **Max Size**: 10MB per log file
- **Max Backups**: 5 backup files kept
- **Naming**: Backups are named `<logfile>.1`, `<logfile>.2`, etc.

Example:
```
asc.log       (current log)
asc.log.1     (most recent backup)
asc.log.2
asc.log.3
asc.log.4
asc.log.5     (oldest backup)
```

## Log Locations

All logs are stored in `~/.asc/logs/`:

- `asc.log`: Main asc process logs
- `<agent-name>.log`: Individual agent logs
- `mcp_agent_mail.log`: MCP server logs
- `health.log`: Health monitoring logs

## Best Practices

### For Development

1. Use `--debug` flag during development for detailed logs
2. Use structured logging with context fields for better traceability
3. Set correlation IDs for request tracing across components
4. Export logs when investigating issues

### For Production

1. Use INFO level logging (default) for normal operation
2. Set up log rotation to manage disk space
3. Run `asc cleanup` periodically to remove old logs
4. Monitor ERROR level logs for failures

### For Troubleshooting

1. Enable debug mode: `asc up --debug`
2. Filter logs by agent or search text in TUI
3. Export filtered logs for analysis
4. Check correlation IDs to trace requests across components
5. Review aggregated logs for patterns

## Examples

### Example 1: Debug Agent Startup Issues

```bash
# Start with debug mode
asc up --debug

# In TUI, press '/' to search for "failed"
# Press 'e' to export filtered logs
# Review exported file for detailed error information
```

### Example 2: Analyze Task Processing

```go
// In agent code, add structured logging
logger.WithFields(logger.Fields{
    "agent": agentName,
    "task": taskID,
    "phase": "implementation",
}).Info("Starting task processing")

// Log detailed steps in debug mode
logger.WithFields(logger.Fields{
    "agent": agentName,
    "task": taskID,
    "file": filename,
}).Debug("Requesting file lease")
```

### Example 3: Monitor Agent Health

```bash
# Start asc
asc up

# In TUI, view aggregated logs in the log pane
# Filter by agent name to see specific agent activity
# Look for ERROR level logs indicating issues
```

## Troubleshooting

### Logs Not Appearing

- Check that logger is initialized: `logger.Init()`
- Verify log level is appropriate (DEBUG < INFO < WARN < ERROR)
- Check file permissions on `~/.asc/logs/`

### Log Files Too Large

- Run `asc cleanup` to remove old logs
- Logs automatically rotate at 10MB
- Adjust rotation settings if needed

### Missing Debug Information

- Ensure `--debug` flag is used: `asc up --debug`
- Check that debug logging is added in the code
- Verify logger is not filtered by level

## API Reference

See the [logger package documentation](../internal/logger/) for detailed API reference.
