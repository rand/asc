# Design Document: Agent Stack Controller (asc)

## Overview

The Agent Stack Controller (asc) is a Go-based CLI tool and TUI dashboard that orchestrates a colony of AI coding agents. The system follows a layered architecture where asc serves as the orchestration layer (L4) managing the communication layer (mcp_agent_mail), task state (beads), and multiple headless agent processes. The design emphasizes developer experience, elegant information density, and zero-framework cognition principles.

## Architecture

### System Layers

```
┌─────────────────────────────────────────────────────────┐
│ L4: Orchestration (asc)                                 │
│ - CLI Commands (cobra)                                  │
│ - Process Management                                    │
│ - TUI Dashboard (bubbletea)                             │
└─────────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Agent        │  │ Agent        │  │ Agent        │
│ Processes    │  │ Processes    │  │ Processes    │
│ (Python)     │  │ (Python)     │  │ (Python)     │
└──────────────┘  └──────────────┘  └──────────────┘
        │                 │                 │
        └─────────────────┼─────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
┌──────────────────┐  ┌──────────────────┐
│ L1: mcp_agent_   │  │ L0: beads        │
│     mail         │  │     (Git DB)     │
│ (Communication)  │  │ (Task State)     │
└──────────────────┘  └──────────────────┘
```

### Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: [cobra](https://github.com/spf13/cobra) for command parsing
- **TUI Framework**: [bubbletea](https://github.com/charmbracelet/bubbletea) for interactive UI
- **TUI Styling**: [lipgloss](https://github.com/charmbracelet/lipgloss) for layout and colors
- **TUI Components**: [bubbles](https://github.com/charmbracelet/bubbles) for reusable widgets
- **Config Parsing**: [viper](https://github.com/spf13/viper) for TOML configuration
- **Process Management**: Go standard library `os/exec` with custom process tracking

### Project Structure

```
asc/
├── cmd/
│   ├── root.go           # Root command and global flags
│   ├── init.go           # asc init command
│   ├── up.go             # asc up command
│   ├── down.go           # asc down command
│   ├── check.go          # asc check command
│   ├── test.go           # asc test command
│   └── services.go       # asc services command
├── internal/
│   ├── config/
│   │   ├── config.go     # Configuration structures
│   │   └── parser.go     # TOML parsing logic
│   ├── process/
│   │   ├── manager.go    # Process lifecycle management
│   │   └── tracker.go    # PID tracking and monitoring
│   ├── check/
│   │   └── checker.go    # Dependency verification
│   ├── tui/
│   │   ├── app.go        # Main bubbletea application
│   │   ├── model.go      # TUI state model
│   │   ├── update.go     # Event handlers
│   │   ├── view.go       # Rendering logic
│   │   ├── agents.go     # Agent status pane
│   │   ├── tasks.go      # Beads task pane
│   │   └── logs.go       # MCP log pane
│   ├── beads/
│   │   ├── client.go     # Beads database interface
│   │   └── parser.go     # JSONL parsing
│   └── mcp/
│       └── client.go     # MCP agent mail HTTP client
├── go.mod
├── go.sum
├── main.go
└── README.md
```

## Components and Interfaces

### 1. Configuration System

#### Config Structure

```go
type Config struct {
    Core     CoreConfig
    Services ServicesConfig
    Agents   map[string]AgentConfig
}

type CoreConfig struct {
    BeadsDBPath string `mapstructure:"beads_db_path"`
}

type ServicesConfig struct {
    MCPAgentMail MCPConfig `mapstructure:"mcp_agent_mail"`
}

type MCPConfig struct {
    StartCommand string `mapstructure:"start_command"`
    URL          string `mapstructure:"url"`
}

type AgentConfig struct {
    Command string   `mapstructure:"command"`
    Model   string   `mapstructure:"model"`
    Phases  []string `mapstructure:"phases"`
}
```

#### Configuration Loading

- Use viper to load and parse `asc.toml`
- Validate required fields on load
- Provide sensible defaults for optional fields
- Support environment variable overrides for sensitive data

### 2. Process Management

#### Process Manager Interface

```go
type ProcessManager interface {
    Start(name string, cmd string, env []string) (pid int, err error)
    Stop(pid int) error
    StopAll() error
    IsRunning(pid int) bool
    GetStatus(pid int) ProcessStatus
}

type ProcessStatus string

const (
    StatusRunning ProcessStatus = "running"
    StatusStopped ProcessStatus = "stopped"
    StatusError   ProcessStatus = "error"
)
```

#### Implementation Details

- Use `exec.Command` to spawn agent processes
- Set `Cmd.SysProcAttr` for process group management
- Store PIDs in `~/.asc/pids/` directory as JSON files
- Implement graceful shutdown with SIGTERM followed by SIGKILL timeout
- Capture stdout/stderr to log files in `~/.asc/logs/`
- Monitor process health via periodic checks

### 3. Dependency Checker

#### Checker Interface

```go
type Checker interface {
    CheckBinary(name string) CheckResult
    CheckFile(path string) CheckResult
    CheckConfig() CheckResult
    CheckEnv(keys []string) CheckResult
    RunAll() []CheckResult
}

type CheckResult struct {
    Name    string
    Status  CheckStatus
    Message string
}

type CheckStatus string

const (
    CheckPass CheckStatus = "pass"
    CheckFail CheckStatus = "fail"
    CheckWarn CheckStatus = "warn"
)
```

#### Checks to Implement

1. **Binary Checks**: git, python3, uv, bd, docker (if needed)
2. **File Checks**: asc.toml exists and is valid TOML
3. **Environment Checks**: .env file exists with required API keys
4. **Path Checks**: beads_db_path is accessible
5. **Network Checks**: mcp_agent_mail URL is reachable (for status command)

### 4. Beads Client

#### Client Interface

```go
type BeadsClient interface {
    GetTasks(statuses []string) ([]Task, error)
    CreateTask(title string) (Task, error)
    UpdateTask(id string, updates TaskUpdate) error
    DeleteTask(id string) error
    Refresh() error
}

type Task struct {
    ID       string
    Title    string
    Status   string
    Phase    string
    Assignee string
}
```

#### Implementation Strategy

- Execute `bd` CLI commands via `exec.Command`
- Parse JSON output from `bd --json` commands
- Implement git pull for refresh operations
- Cache task list and refresh on interval (configurable, default 5s)
- Handle git conflicts gracefully

### 5. MCP Client

#### Client Interface

```go
type MCPClient interface {
    GetMessages(since time.Time) ([]Message, error)
    SendMessage(msg Message) error
    GetAgentStatus(agentName string) (AgentStatus, error)
}

type Message struct {
    Timestamp time.Time
    Type      MessageType
    Source    string
    Content   string
}

type MessageType string

const (
    TypeLease   MessageType = "lease"
    TypeBeads   MessageType = "beads"
    TypeError   MessageType = "error"
    TypeMessage MessageType = "message"
)

type AgentStatus struct {
    Name      string
    State     AgentState
    CurrentTask string
    LastSeen  time.Time
}

type AgentState string

const (
    StateIdle    AgentState = "idle"
    StateWorking AgentState = "working"
    StateError   AgentState = "error"
    StateOffline AgentState = "offline"
)
```

#### Implementation Strategy

- Use Go's `net/http` client for REST API calls
- Implement polling with configurable interval (default 2s)
- Parse JSON responses into structured types
- Handle connection errors gracefully
- Implement exponential backoff for retries

### 6. TUI Application

#### Bubbletea Model

```go
type Model struct {
    // Configuration
    config Config
    
    // Data sources
    beadsClient BeadsClient
    mcpClient   MCPClient
    procManager ProcessManager
    
    // State
    agents      []AgentStatus
    tasks       []Task
    messages    []Message
    
    // UI state
    width       int
    height      int
    activePane  int
    
    // Refresh control
    lastRefresh time.Time
    ticker      *time.Ticker
}
```

#### Update Loop

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyPress(msg)
    case tea.WindowSizeMsg:
        return m.handleResize(msg)
    case tickMsg:
        return m.handleTick()
    case refreshMsg:
        return m.handleRefresh(msg)
    }
    return m, nil
}
```

#### View Rendering

Use lipgloss to create a three-pane layout:

```go
func (m Model) View() string {
    // Calculate dimensions
    paneHeight := m.height - 4 // Reserve space for footer
    leftWidth := m.width / 3
    rightWidth := m.width - leftWidth
    topHeight := paneHeight / 2
    bottomHeight := paneHeight - topHeight
    
    // Render panes
    agentPane := m.renderAgentPane(leftWidth, paneHeight)
    taskPane := m.renderTaskPane(rightWidth, topHeight)
    logPane := m.renderLogPane(m.width, bottomHeight)
    footer := m.renderFooter(m.width)
    
    // Compose layout
    topRow := lipgloss.JoinHorizontal(lipgloss.Top, agentPane, taskPane)
    mainView := lipgloss.JoinVertical(lipgloss.Left, topRow, logPane)
    
    return lipgloss.JoinVertical(lipgloss.Left, mainView, footer)
}
```

#### Pane Implementations

**Agent Status Pane:**
- List all agents from config
- Show status icon (●/⟳/!/○) with color
- Display current task if working
- Update from MCP heartbeat messages

**Task Stream Pane:**
- Display tasks with status "open" or "in_progress"
- Show task ID, status icon, and title
- Highlight in-progress tasks
- Scroll if content exceeds height

**MCP Log Pane:**
- Tail-style log display
- Color-code by message type
- Auto-scroll to bottom
- Limit to last N messages (configurable)

**Footer:**
- Show keybindings: (q)uit | (r)efresh | (t)est
- Display connection status indicators

### 7. CLI Commands

#### asc init

**Flow:**
1. Display welcome screen
2. Run dependency checks
3. Prompt for missing installations
4. Backup existing configs
5. Collect API keys (with masked input)
6. Generate default asc.toml
7. Create .env file
8. Run asc test
9. Display success message

**Implementation:**
- Use bubbletea for interactive prompts
- Use bubbles/textinput for API key entry
- Validate inputs before proceeding
- Provide clear error messages

#### asc up

**Flow:**
1. Run silent check (exit on failure)
2. Load config
3. Start mcp_agent_mail service
4. Launch each agent process
5. Initialize TUI
6. Enter event loop

**Implementation:**
- Use process manager to track all PIDs
- Pass environment variables to agents
- Handle startup failures gracefully
- Provide startup progress feedback

#### asc down

**Flow:**
1. Load config
2. Read PID files
3. Send SIGTERM to all processes
4. Wait for graceful shutdown (5s timeout)
5. Send SIGKILL to remaining processes
6. Stop mcp_agent_mail
7. Clean up PID files

**Implementation:**
- Use process manager for shutdown
- Log shutdown progress
- Handle missing processes gracefully

#### asc check

**Flow:**
1. Run all dependency checks
2. Format results as table
3. Print to stdout
4. Exit with appropriate code

**Implementation:**
- Use checker interface
- Format with lipgloss table
- Provide actionable error messages

#### asc test

**Flow:**
1. Create test beads task
2. Send test MCP message
3. Poll for confirmation (30s timeout)
4. Clean up test artifacts
5. Report results

**Implementation:**
- Use beads and MCP clients
- Implement timeout handling
- Provide detailed failure diagnostics

#### asc services [start|stop|status]

**Flow:**
- **start**: Launch mcp_agent_mail, save PID
- **stop**: Kill mcp_agent_mail process
- **status**: Check if process is running

**Implementation:**
- Use process manager
- Store service PID separately from agents
- Provide clear status output

## Data Models

### Configuration File (asc.toml)

```toml
[core]
beads_db_path = "./project-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.main-planner]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning", "design"]

[agent.claude-refactor]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation", "refactor"]
```

### Environment File (.env)

```
CLAUDE_API_KEY=sk-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=...
```

### PID Tracking File (~/.asc/pids/agents.json)

```json
{
  "main-planner": {
    "pid": 12345,
    "started_at": "2025-11-09T10:30:00Z",
    "command": "python agent_adapter.py",
    "env": {
      "AGENT_NAME": "main-planner",
      "AGENT_MODEL": "gemini",
      "AGENT_PHASES": "planning,design"
    }
  }
}
```

## Error Handling

### Error Categories

1. **Configuration Errors**: Invalid TOML, missing required fields
2. **Dependency Errors**: Missing binaries, inaccessible paths
3. **Process Errors**: Failed to start, crashed agents
4. **Network Errors**: MCP server unreachable, beads sync failures
5. **User Errors**: Invalid commands, missing arguments

### Error Handling Strategy

- Use Go's error wrapping for context
- Provide actionable error messages
- Log detailed errors to `~/.asc/logs/asc.log`
- Display user-friendly errors in TUI
- Implement graceful degradation (e.g., show cached data if refresh fails)

### Error Message Format

```
Error: Failed to start agent 'main-planner'
Reason: Command 'python agent_adapter.py' not found
Solution: Ensure Python 3 is installed and agent_adapter.py exists
```

## Testing Strategy

### Unit Tests

- **Config parsing**: Test TOML loading and validation
- **Process management**: Test start/stop/status operations
- **Checkers**: Test each dependency check
- **Clients**: Test beads and MCP client parsing

### Integration Tests

- **End-to-end flow**: Test init → up → test → down
- **TUI rendering**: Test layout calculations and pane rendering
- **Process lifecycle**: Test agent startup and shutdown

### Manual Testing

- **Visual testing**: Verify TUI appearance and responsiveness
- **Error scenarios**: Test behavior with missing dependencies
- **Performance**: Test with multiple agents and high message volume

### Test Data

- Create mock asc.toml configurations
- Mock beads database responses
- Mock MCP server responses
- Use test fixtures for various scenarios

## Performance Considerations

### Polling Intervals

- **Beads refresh**: 5 seconds (configurable)
- **MCP messages**: 2 seconds (configurable)
- **Agent health**: 10 seconds (configurable)

### Resource Management

- Limit log message buffer to 1000 entries
- Implement log rotation for agent stdout/stderr
- Use efficient JSON parsing for beads database
- Minimize TUI re-renders with smart diffing

### Scalability

- Support up to 20 concurrent agents
- Handle beads databases with 10,000+ tasks
- Process 100+ MCP messages per second

## Security Considerations

### API Key Management

- Store API keys in .env file (not in asc.toml)
- Never log API keys
- Use file permissions (0600) for .env
- Warn if .env is world-readable

### Process Isolation

- Run agents with same user permissions as asc
- Do not elevate privileges
- Isolate agent stdout/stderr to separate log files

### Input Validation

- Validate all TOML configuration values
- Sanitize command arguments before execution
- Prevent command injection in agent commands

## Deployment and Distribution

### Build Process

```bash
# Build for current platform
go build -o asc main.go

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o asc-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o asc-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o asc-darwin-arm64
```

### Installation

```bash
# Install to /usr/local/bin
sudo cp asc /usr/local/bin/
sudo chmod +x /usr/local/bin/asc

# Or use go install
go install github.com/yourusername/asc@latest
```

### Dependencies

**Required:**
- Go 1.21+ (for building)
- git
- python3
- beads CLI (bd)
- mcp_agent_mail server

**Optional:**
- docker (if agents require containers)
- uv (Python package manager)

## Future Enhancements

### Phase 2 Features

- **Agent templates**: Pre-configured agent setups for common workflows
- **Remote agents**: Support for agents running on different machines
- **Web dashboard**: Browser-based alternative to TUI
- **Metrics**: Collect and display agent performance metrics
- **Notifications**: Desktop notifications for important events

### Phase 3 Features

- **Agent marketplace**: Share and discover agent configurations
- **Distributed coordination**: Multi-machine agent colonies
- **Advanced scheduling**: Priority queues and resource allocation
- **Replay mode**: Replay agent interactions for debugging

## References

- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Bubbletea TUI Framework](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)
- [Viper Configuration](https://github.com/spf13/viper)
- [beads Task Database](https://github.com/steveyegge/beads)
- [mcp_agent_mail](https://github.com/Dicklesworthstone/mcp_agent_mail)
