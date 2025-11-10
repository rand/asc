# API Reference

Complete API documentation for the Agent Stack Controller (asc).

## Table of Contents

- [CLI Commands](#cli-commands)
- [Go Packages](#go-packages)
- [Python Agent API](#python-agent-api)
- [Configuration API](#configuration-api)
- [MCP Protocol](#mcp-protocol)
- [Beads Integration](#beads-integration)

---

## CLI Commands

### asc init

Initialize the agent stack with interactive setup wizard.

**Usage:**
```bash
asc init [flags]
```

**Flags:**
- `--template=<name>` - Use a configuration template (solo, team, swarm)
- `--list-templates` - List available templates
- `--save-template=<name>` - Save current config as a template
- `--skip-checks` - Skip dependency checks
- `--non-interactive` - Run without prompts (use defaults)

**Examples:**
```bash
# Interactive setup
asc init

# Use team template
asc init --template=team

# List available templates
asc init --list-templates

# Save custom template
asc init --save-template my-setup
```

**Exit Codes:**
- `0` - Success
- `1` - Dependency check failed
- `2` - Configuration error
- `3` - User cancelled

---

### asc up

Start all agents and launch the TUI dashboard.

**Usage:**
```bash
asc up [flags]
```

**Flags:**
- `--debug` - Enable debug logging
- `--no-tui` - Start agents without TUI
- `--config=<path>` - Use alternate config file (default: asc.toml)

**Examples:**
```bash
# Start with TUI
asc up

# Start in debug mode
asc up --debug

# Start without TUI (headless)
asc up --no-tui

# Use custom config
asc up --config=asc.production.toml
```

**Exit Codes:**
- `0` - Clean shutdown
- `1` - Startup failed
- `2` - Configuration error

---

### asc down

Stop all agents and services gracefully.

**Usage:**
```bash
asc down [flags]
```

**Flags:**
- `--force` - Force kill processes (SIGKILL)
- `--timeout=<seconds>` - Graceful shutdown timeout (default: 5)

**Examples:**
```bash
# Graceful shutdown
asc down

# Force shutdown
asc down --force

# Custom timeout
asc down --timeout=10
```

**Exit Codes:**
- `0` - All processes stopped
- `1` - Some processes failed to stop

---

### asc check

Verify environment dependencies and configuration.

**Usage:**
```bash
asc check [flags]
```

**Flags:**
- `--json` - Output results as JSON
- `--verbose` - Show detailed check information

**Examples:**
```bash
# Run checks
asc check

# JSON output
asc check --json

# Verbose output
asc check --verbose
```

**Exit Codes:**
- `0` - All checks passed
- `1` - One or more checks failed

**JSON Output Format:**
```json
{
  "checks": [
    {
      "name": "git",
      "status": "pass",
      "message": "Found at /usr/bin/git",
      "version": "2.39.0"
    }
  ],
  "summary": {
    "total": 6,
    "passed": 6,
    "failed": 0,
    "warnings": 0
  }
}
```

---

### asc test

Run end-to-end health check of the agent stack.

**Usage:**
```bash
asc test [flags]
```

**Flags:**
- `--timeout=<seconds>` - Test timeout (default: 30)
- `--verbose` - Show detailed test output

**Examples:**
```bash
# Run health check
asc test

# With custom timeout
asc test --timeout=60

# Verbose output
asc test --verbose
```

**Exit Codes:**
- `0` - Stack is healthy
- `1` - Health check failed

---

### asc services

Manage long-running services (mcp_agent_mail).

**Usage:**
```bash
asc services <command> [flags]
```

**Commands:**
- `start` - Start the MCP server
- `stop` - Stop the MCP server
- `status` - Check server status
- `restart` - Restart the server

**Examples:**
```bash
# Start MCP server
asc services start

# Check status
asc services status

# Restart server
asc services restart
```

**Exit Codes:**
- `0` - Command succeeded
- `1` - Command failed

---

### asc doctor

Diagnose and fix common issues.

**Usage:**
```bash
asc doctor [flags]
```

**Flags:**
- `--fix` - Automatically fix detected issues
- `--verbose` - Show detailed diagnostics
- `--json` - Output as JSON

**Examples:**
```bash
# Run diagnostics
asc doctor

# Auto-fix issues
asc doctor --fix

# JSON output
asc doctor --json
```

**Exit Codes:**
- `0` - No issues found
- `1` - Issues detected
- `2` - Fix failed

---

### asc secrets

Manage encrypted secrets.

**Usage:**
```bash
asc secrets <command> [flags]
```

**Commands:**
- `encrypt` - Encrypt .env to .env.age
- `decrypt` - Decrypt .env.age to .env
- `status` - Show encryption status
- `rotate` - Rotate encryption key

**Examples:**
```bash
# Encrypt secrets
asc secrets encrypt

# Decrypt secrets
asc secrets decrypt

# Check status
asc secrets status

# Rotate key
asc secrets rotate
```

**Exit Codes:**
- `0` - Command succeeded
- `1` - Command failed

---

## Go Packages

### internal/config

Configuration parsing and management.

#### type Config

```go
type Config struct {
    Core     CoreConfig
    Services ServicesConfig
    Agents   map[string]AgentConfig
}
```

**Methods:**

```go
// Load loads configuration from a TOML file
func Load(path string) (*Config, error)

// Validate validates the configuration
func (c *Config) Validate() error

// Save saves configuration to a TOML file
func (c *Config) Save(path string) error
```

**Example:**

```go
import "github.com/yourusername/asc/internal/config"

cfg, err := config.Load("asc.toml")
if err != nil {
    log.Fatal(err)
}

if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

#### type AgentConfig

```go
type AgentConfig struct {
    Command string   `mapstructure:"command"`
    Model   string   `mapstructure:"model"`
    Phases  []string `mapstructure:"phases"`
}
```

---

### internal/process

Process lifecycle management.

#### type ProcessManager

```go
type ProcessManager interface {
    Start(name string, cmd string, env []string) (pid int, err error)
    Stop(pid int) error
    StopAll() error
    IsRunning(pid int) bool
    GetStatus(pid int) ProcessStatus
}
```

**Methods:**

```go
// NewManager creates a new process manager
func NewManager(pidDir, logDir string) ProcessManager

// Start starts a new process
func (m *Manager) Start(name string, cmd string, env []string) (int, error)

// Stop stops a process by PID
func (m *Manager) Stop(pid int) error

// StopAll stops all managed processes
func (m *Manager) StopAll() error

// IsRunning checks if a process is running
func (m *Manager) IsRunning(pid int) bool

// GetStatus returns the status of a process
func (m *Manager) GetStatus(pid int) ProcessStatus
```

**Example:**

```go
import "github.com/yourusername/asc/internal/process"

mgr := process.NewManager("~/.asc/pids", "~/.asc/logs")

pid, err := mgr.Start("agent-1", "python agent_adapter.py", []string{
    "AGENT_NAME=agent-1",
    "AGENT_MODEL=claude",
})
if err != nil {
    log.Fatal(err)
}

// Later...
if err := mgr.Stop(pid); err != nil {
    log.Fatal(err)
}
```

---

### internal/check

Dependency checking and validation.

#### type Checker

```go
type Checker interface {
    CheckBinary(name string) CheckResult
    CheckFile(path string) CheckResult
    CheckConfig() CheckResult
    CheckEnv(keys []string) CheckResult
    RunAll() []CheckResult
}
```

**Example:**

```go
import "github.com/yourusername/asc/internal/check"

checker := check.NewChecker()
results := checker.RunAll()

for _, result := range results {
    fmt.Printf("%s: %s\n", result.Name, result.Status)
}
```

---

### internal/beads

Beads task database integration.

#### type BeadsClient

```go
type BeadsClient interface {
    GetTasks(statuses []string) ([]Task, error)
    CreateTask(title string) (Task, error)
    UpdateTask(id string, updates TaskUpdate) error
    DeleteTask(id string) error
    Refresh() error
}
```

**Example:**

```go
import "github.com/yourusername/asc/internal/beads"

client := beads.NewClient("./project-repo")

tasks, err := client.GetTasks([]string{"open", "in_progress"})
if err != nil {
    log.Fatal(err)
}

for _, task := range tasks {
    fmt.Printf("#%s: %s\n", task.ID, task.Title)
}
```

---

### internal/mcp

MCP agent mail integration.

#### type MCPClient

```go
type MCPClient interface {
    GetMessages(since time.Time) ([]Message, error)
    SendMessage(msg Message) error
    GetAgentStatus(agentName string) (AgentStatus, error)
    GetAllAgentStatuses() ([]AgentStatus, error)
}
```

**Example:**

```go
import "github.com/yourusername/asc/internal/mcp"

client := mcp.NewClient("http://localhost:8765")

statuses, err := client.GetAllAgentStatuses()
if err != nil {
    log.Fatal(err)
}

for _, status := range statuses {
    fmt.Printf("%s: %s\n", status.Name, status.State)
}
```

---

## Python Agent API

### agent_adapter.py

Main entry point for agent processes.

**Environment Variables:**

- `AGENT_NAME` - Agent identifier
- `AGENT_MODEL` - LLM model (claude, gemini, gpt-4, codex)
- `AGENT_PHASES` - Comma-separated phases
- `MCP_MAIL_URL` - MCP server URL
- `BEADS_DB_PATH` - Path to beads repository
- `CLAUDE_API_KEY` - Claude API key
- `OPENAI_API_KEY` - OpenAI API key
- `GOOGLE_API_KEY` - Google API key

**Usage:**

```bash
export AGENT_NAME=my-agent
export AGENT_MODEL=claude
export AGENT_PHASES=planning,implementation
export MCP_MAIL_URL=http://localhost:8765
export BEADS_DB_PATH=./project-repo
export CLAUDE_API_KEY=sk-...

python agent_adapter.py
```

---

### llm_client.py

LLM provider abstraction.

#### class LLMClient

```python
class LLMClient(ABC):
    @abstractmethod
    def complete(self, prompt: str, context: dict) -> str:
        """Generate completion from LLM"""
        pass
```

**Implementations:**

- `ClaudeClient` - Anthropic Claude
- `GeminiClient` - Google Gemini
- `OpenAIClient` - OpenAI GPT-4/Codex

**Example:**

```python
from llm_client import ClaudeClient

client = ClaudeClient(api_key=os.getenv("CLAUDE_API_KEY"))

response = client.complete(
    prompt="Implement a function to sort a list",
    context={"language": "python"}
)

print(response)
```

---

### phase_loop.py

Hephaestus task execution loop.

#### class PhaseLoop

```python
class PhaseLoop:
    def __init__(self, agent_name: str, phases: list, llm_client: LLMClient):
        """Initialize phase loop"""
        pass
    
    def run(self):
        """Run the main event loop"""
        pass
```

**Example:**

```python
from phase_loop import PhaseLoop
from llm_client import ClaudeClient

client = ClaudeClient(api_key=os.getenv("CLAUDE_API_KEY"))
loop = PhaseLoop(
    agent_name="my-agent",
    phases=["planning", "implementation"],
    llm_client=client
)

loop.run()
```

---

### ace.py

ACE (Agentic Context Engineering) playbook system.

#### class ACEPlaybook

```python
class ACEPlaybook:
    def __init__(self, agent_name: str):
        """Initialize playbook"""
        pass
    
    def add_lesson(self, lesson: dict):
        """Add a new lesson to the playbook"""
        pass
    
    def get_relevant_lessons(self, context: dict) -> list:
        """Retrieve relevant lessons for context"""
        pass
```

**Example:**

```python
from ace import ACEPlaybook

playbook = ACEPlaybook(agent_name="my-agent")

playbook.add_lesson({
    "context": "implementing authentication",
    "action": "used JWT tokens",
    "outcome": "successful",
    "learned": "JWT is better than sessions for stateless APIs"
})

lessons = playbook.get_relevant_lessons({"task": "implement auth"})
```

---

### heartbeat.py

Agent status reporting.

#### class HeartbeatManager

```python
class HeartbeatManager:
    def __init__(self, agent_name: str, mcp_url: str):
        """Initialize heartbeat manager"""
        pass
    
    def send_heartbeat(self, status: str, current_task: str = None):
        """Send heartbeat to MCP server"""
        pass
    
    def start(self):
        """Start periodic heartbeat"""
        pass
```

**Example:**

```python
from heartbeat import HeartbeatManager

hb = HeartbeatManager(
    agent_name="my-agent",
    mcp_url="http://localhost:8765"
)

hb.start()
hb.send_heartbeat(status="working", current_task="#42")
```

---

## Configuration API

### asc.toml Format

Complete configuration file reference.

```toml
# Core settings
[core]
beads_db_path = "./project-repo"  # Path to beads repository

# Service configuration
[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

# Agent definitions
[agent.agent-name]
command = "python agent_adapter.py"  # Command to start agent
model = "claude"                      # LLM model: claude, gemini, gpt-4, codex
phases = ["planning", "implementation"]  # Workflow phases
```

**Valid Models:**
- `claude` - Anthropic Claude (Sonnet, Opus)
- `gemini` - Google Gemini
- `gpt-4` - OpenAI GPT-4
- `codex` - OpenAI Codex

**Valid Phases:**
- `planning` - Task planning and design
- `design` - Architecture design
- `implementation` - Code implementation
- `refactor` - Code refactoring
- `testing` - Test writing
- `validation` - Test validation
- `documentation` - Documentation writing

---

## MCP Protocol

### HTTP Endpoints

#### GET /messages

Retrieve messages since a timestamp.

**Query Parameters:**
- `since` - ISO 8601 timestamp

**Response:**
```json
{
  "messages": [
    {
      "timestamp": "2024-11-09T10:30:15Z",
      "type": "lease",
      "source": "agent-1",
      "content": "Requested lease for src/auth.go"
    }
  ]
}
```

#### POST /messages

Send a message.

**Request Body:**
```json
{
  "type": "message",
  "source": "agent-1",
  "content": "Task completed"
}
```

**Response:**
```json
{
  "id": "msg-123",
  "status": "sent"
}
```

#### GET /agents

Get all agent statuses.

**Response:**
```json
{
  "agents": [
    {
      "name": "agent-1",
      "state": "working",
      "current_task": "#42",
      "last_seen": "2024-11-09T10:30:15Z"
    }
  ]
}
```

#### POST /leases

Request a file lease.

**Request Body:**
```json
{
  "agent": "agent-1",
  "file": "src/auth.go"
}
```

**Response:**
```json
{
  "lease_id": "lease-123",
  "granted": true,
  "expires_at": "2024-11-09T11:30:15Z"
}
```

#### POST /leases/{id}/release

Release a file lease.

**Response:**
```json
{
  "status": "released"
}
```

---

## Beads Integration

### bd CLI Commands

Commands used by asc to interact with beads.

#### List Tasks

```bash
bd list --status=open,in_progress --json
```

**Output:**
```json
[
  {
    "id": "42",
    "title": "Implement authentication",
    "status": "open",
    "phase": "implementation"
  }
]
```

#### Create Task

```bash
bd create "Task title" --phase=planning --json
```

#### Update Task

```bash
bd update 42 --status=in_progress --assignee=agent-1 --json
```

#### Delete Task

```bash
bd delete 42 --json
```

---

## Error Codes

### Common Error Codes

- `ERR_CONFIG_INVALID` - Configuration validation failed
- `ERR_DEPENDENCY_MISSING` - Required dependency not found
- `ERR_PROCESS_START_FAILED` - Failed to start process
- `ERR_PROCESS_STOP_FAILED` - Failed to stop process
- `ERR_MCP_CONNECTION` - Cannot connect to MCP server
- `ERR_BEADS_CONNECTION` - Cannot access beads database
- `ERR_API_KEY_MISSING` - Required API key not found
- `ERR_FILE_PERMISSION` - File permission error

### Error Response Format

```json
{
  "error": {
    "code": "ERR_CONFIG_INVALID",
    "message": "Invalid configuration: missing required field 'beads_db_path'",
    "details": {
      "field": "core.beads_db_path",
      "expected": "string",
      "got": "null"
    },
    "suggestion": "Add 'beads_db_path' to the [core] section in asc.toml"
  }
}
```

---

## See Also

- [Configuration Guide](CONFIGURATION.md)
- [Operator's Handbook](OPERATORS_HANDBOOK.md)
- [FAQ](FAQ.md)
- [Troubleshooting](../TROUBLESHOOTING.md)
