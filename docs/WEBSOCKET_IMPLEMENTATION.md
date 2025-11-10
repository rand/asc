# WebSocket Real-Time Updates Implementation

## Overview

This document describes the implementation of real-time TUI updates using WebSocket connections to the MCP server, replacing the previous polling-based approach for agent status and message updates.

## Architecture

### Components

1. **WebSocket Client** (`internal/mcp/websocket.go`)
   - Manages WebSocket connection to MCP server
   - Handles automatic reconnection with exponential backoff
   - Distributes events via buffered channel
   - Monitors connection health with periodic pings

2. **Event-Driven TUI** (`internal/tui/`)
   - Receives real-time events from WebSocket
   - Updates UI immediately on agent status changes
   - Maintains fallback polling for beads (git-based)
   - Shows connection status in footer

### Event Flow

```
MCP Server (WebSocket)
    │
    ├─> agent_status events ──> Update agent list in TUI
    ├─> new_message events ───> Append to message log
    ├─> connected event ──────> Set wsConnected = true
    └─> disconnected event ───> Set wsConnected = false, trigger reconnect

Beads (Git-based)
    │
    └─> Periodic polling (5s) ─> Update task list
```

## Key Features

### 1. WebSocket Client

**Connection Management:**
- Automatic connection on TUI startup
- Exponential backoff reconnection (1s → 30s max)
- Health monitoring with periodic pings (10s interval)
- Graceful shutdown on TUI exit

**Event Types:**
- `EventConnected`: WebSocket connection established
- `EventDisconnected`: Connection lost (triggers auto-reconnect)
- `EventAgentStatus`: Agent status changed (name, state, current task)
- `EventNewMessage`: New message received (timestamp, type, source, content)
- `EventError`: WebSocket error occurred

**Subscription:**
- Automatically subscribes to `agent_status` events
- Automatically subscribes to `new_message` events

### 2. Event-Driven TUI Updates

**Real-Time Updates (via WebSocket):**
- Agent status changes (idle → working → error → offline)
- New messages in MCP interaction log
- Connection status indicators

**Polling Updates (fallback for beads):**
- Task list refresh every 5 seconds
- Git-based, cannot be real-time

**Connection Status Display:**
- `● ws` - WebSocket connected (green)
- `● http` - HTTP polling fallback (yellow)
- `○` - Disconnected (red)

### 3. Performance Improvements

**Reduced CPU Usage:**
- No polling for MCP data when WebSocket is connected
- Events only trigger UI updates when data changes
- Idle periods consume minimal resources

**Reduced Network Traffic:**
- Single persistent WebSocket connection vs. repeated HTTP polls
- Server pushes updates only when changes occur
- No unnecessary polling requests

**Improved Responsiveness:**
- Agent status updates appear instantly
- Messages appear in real-time
- No 2-3 second polling delay

## Implementation Details

### WebSocket URL Construction

HTTP URLs are automatically converted to WebSocket URLs:
- `http://localhost:8765` → `ws://localhost:8765/ws`
- `https://example.com` → `wss://example.com/ws`

### Reconnection Strategy

1. Initial connection attempt on TUI startup
2. On disconnect: wait 1 second, retry
3. On failure: double delay (1s → 2s → 4s → 8s → 16s → 30s max)
4. Continue retrying until connection succeeds or TUI exits

### Fallback Behavior

If WebSocket connection fails:
- TUI continues to function normally
- Falls back to HTTP polling for MCP data
- Connection status shows `● http` (yellow)
- User can still interact with all features

### Event Buffering

- Event channel has 100-event buffer
- Prevents blocking on slow UI updates
- Messages are limited to last 100 in memory

## Testing

### Unit Tests

**WebSocket Client Tests** (`internal/mcp/websocket_test.go`):
- ✓ Connection establishment
- ✓ Agent status event reception
- ✓ New message event reception
- ✓ Automatic reconnection
- ✓ Graceful shutdown
- ✓ Event buffering

All tests pass successfully.

### Integration Testing

To test the WebSocket implementation:

1. Start MCP server with WebSocket support
2. Run `asc up`
3. Verify footer shows `● ws` (green) for MCP connection
4. Trigger agent status change on server
5. Verify TUI updates immediately without polling delay
6. Stop MCP server
7. Verify footer shows `○` (red) and reconnection attempts
8. Restart MCP server
9. Verify automatic reconnection and `● ws` status

## Configuration

No additional configuration required. The WebSocket client automatically:
- Derives WebSocket URL from `services.mcp_agent_mail.url` in `asc.toml`
- Connects on TUI startup
- Falls back to HTTP polling if WebSocket unavailable

## Future Enhancements

Potential improvements for Phase 3:
- Configurable reconnection parameters
- WebSocket compression for large message volumes
- Binary protocol for improved performance
- Multiplexed subscriptions for selective updates
- WebSocket authentication/authorization

## References

- WebSocket RFC: https://tools.ietf.org/html/rfc6455
- Gorilla WebSocket: https://github.com/gorilla/websocket
- Bubbletea Framework: https://github.com/charmbracelet/bubbletea
