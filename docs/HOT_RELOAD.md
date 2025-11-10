# Configuration Hot-Reload

The Agent Stack Controller supports hot-reloading of the `asc.toml` configuration file without requiring a full restart. This allows you to add, remove, or update agent configurations on the fly while the TUI is running.

## How It Works

The hot-reload system consists of three main components:

1. **File Watcher** (`internal/config/watcher.go`): Monitors the `asc.toml` file for changes using `fsnotify`
2. **Reload Manager** (`internal/config/reload.go`): Compares old and new configurations and manages agent lifecycle
3. **TUI Integration** (`internal/tui/model.go`): Displays reload notifications and updates the UI

### File Watching

The file watcher uses `fsnotify` to monitor the configuration file for write and create events. It includes:

- **Debouncing**: Multiple rapid changes are debounced to a single reload event (500ms delay)
- **Validation**: Only valid configurations trigger reload events
- **Error Handling**: Invalid configurations are logged but don't stop the watcher

### Configuration Comparison

When a configuration change is detected, the reload manager:

1. Compares the new configuration with the current one
2. Identifies agents that were:
   - **Added**: New agents in the configuration
   - **Removed**: Agents no longer in the configuration
   - **Updated**: Agents with changed command, model, or phases

### Agent Lifecycle Management

Based on the comparison results:

- **Removed agents**: Stopped gracefully (SIGTERM)
- **Updated agents**: Stopped and restarted with new configuration
- **Added agents**: Started with the new configuration

## Usage

### Automatic Reload

When running `asc up`, hot-reload is automatically enabled. Simply edit your `asc.toml` file and save it. The changes will be detected and applied automatically.

### Reload Notifications

The TUI displays reload notifications in the footer for 5 seconds:

- ✓ **Success**: `Config reloaded: Added: agent-name | Removed: agent-name | Updated: agent-name`
- ❌ **Error**: `Config reload failed: error message`

### Example Workflow

1. Start the agent stack:
   ```bash
   asc up
   ```

2. Edit `asc.toml` to add a new agent:
   ```toml
   [agent.new-agent]
   command = "python agent_adapter.py"
   model = "gemini"
   phases = ["testing"]
   ```

3. Save the file - the new agent will start automatically

4. Edit `asc.toml` to change an agent's model:
   ```toml
   [agent.existing-agent]
   command = "python agent_adapter.py"
   model = "claude"  # Changed from "gemini"
   phases = ["planning"]
   ```

5. Save the file - the agent will restart with the new model

6. Remove an agent from `asc.toml` and save - the agent will stop

## Configuration Changes Detected

The reload system detects the following changes:

### Agent Addition
- A new `[agent.name]` section appears in the configuration

### Agent Removal
- An existing `[agent.name]` section is removed from the configuration

### Agent Updates
Any of the following changes trigger an agent restart:
- `command` field changes
- `model` field changes
- `phases` array changes (added, removed, or reordered phases)

### Non-Reloadable Changes

The following configuration changes require a full restart (`asc down` then `asc up`):

- `core.beads_db_path` changes
- `services.mcp_agent_mail.url` changes
- `services.mcp_agent_mail.start_command` changes

These changes affect core services that cannot be reloaded without restarting the entire stack.

## Error Handling

### Invalid Configuration

If you save an invalid configuration (e.g., unsupported model, invalid phase), the reload will fail and an error notification will be displayed. The previous configuration remains active.

Common validation errors:
- Command not found in PATH
- Unsupported model name
- Invalid phase name
- Duplicate agent names

### Partial Failures

If some agents fail to start or stop during reload, the reload continues with other agents. Errors are displayed in the notification:

```
✓ Config reloaded: Added: agent1 | Updated: agent2 (Errors: failed to start agent1: command not found)
```

### Recovery

If a reload fails:
1. The previous configuration remains active
2. Check the error message in the notification
3. Fix the configuration issue
4. Save the file again to retry

## Implementation Details

### Debouncing

The file watcher includes a 500ms debounce timer to handle:
- Text editors that save files by creating a temporary file and renaming it
- Multiple rapid saves during editing
- File system events that fire multiple times for a single save

### Process Management

Agent processes are managed using the existing process manager:
- PIDs are tracked in `~/.asc/pids/`
- Logs are written to `~/.asc/logs/`
- Graceful shutdown with SIGTERM (5s timeout before SIGKILL)

### Environment Variables

When restarting agents, the reload manager preserves:
- API keys from the `.env` file
- Agent-specific environment variables (AGENT_NAME, AGENT_MODEL, AGENT_PHASES)
- MCP and beads configuration

## Testing

The hot-reload system includes comprehensive tests:

### Watcher Tests
```bash
go test ./internal/config -run TestWatcher -v
```

Tests cover:
- Basic file change detection
- Invalid configuration handling
- Multiple rapid changes (debouncing)

### Reload Manager Tests
```bash
go test ./internal/config -run TestReloadManager -v
```

Tests cover:
- Adding new agents
- Removing existing agents
- Updating agent configurations

## Troubleshooting

### Reload Not Triggering

If configuration changes aren't being detected:

1. Check that you're editing the correct `asc.toml` file (in the current directory)
2. Ensure the file is being saved (not just modified in the editor)
3. Check for validation errors in the TUI notification
4. Look for watcher errors in `~/.asc/logs/asc.log`

### Agent Not Starting After Reload

If an agent fails to start after reload:

1. Check the error message in the reload notification
2. Verify the command exists in PATH
3. Check agent logs in `~/.asc/logs/{agent-name}.log`
4. Ensure the model name is valid (claude, gemini, gpt-4, codex, openai)
5. Verify phases are valid

### Performance Impact

The file watcher has minimal performance impact:
- Uses efficient OS-level file system notifications (inotify on Linux, FSEvents on macOS)
- Debouncing reduces unnecessary reloads
- Only monitors a single file (`asc.toml`)

## Future Enhancements

Potential improvements for future versions:

- Hot-reload of `.env` file for API key updates
- Reload of core service configurations
- Configurable debounce duration
- Reload history and rollback capability
- Dry-run mode to preview changes before applying
