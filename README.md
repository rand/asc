# Agent Stack Controller (asc)

A command-line orchestration tool for managing a local colony of AI coding agents. Think of it as mission control for your headless AI developers.

```
     ┌──────────────────────────────────┐
     │  Agent Stack Controller (asc)    │
     │  - CLI Commands                  │
     │  - TUI Dashboard                 │
     │  - Process Management            │
     └──────────────────────────────────┘
                    │
     ┌──────────────┼──────────────┐
     ▼              ▼              ▼
┌────────┐    ┌────────┐    ┌────────┐
│ Agent  │    │ Agent  │    │ Agent  │
│Planner │    │ Coder  │    │ Tester │
└────────┘    └────────┘    └────────┘
     │              │              │
     └──────────────┼──────────────┘
                    │
     ┌──────────────┴──────────────┐
     ▼                             ▼
┌─────────────┐            ┌──────────────┐
│ mcp_agent_  │            │    beads     │
│    mail     │            │ (Task DB)    │
└─────────────┘            └──────────────┘
```

## Features

- **Single-command orchestration**: Start, stop, and monitor multiple AI agents with simple CLI commands
- **Real-time TUI dashboard**: Beautiful terminal interface showing agent status, task streams, and communication logs
- **Flexible agent matrix**: Mix and match any LLM (Claude, Gemini, OpenAI) with any role (planner, coder, tester)
- **Git-backed task management**: Integration with beads for persistent, version-controlled task state
- **Asynchronous coordination**: Agents communicate through mcp_agent_mail for file leasing and task coordination
- **Zero-config startup**: Interactive setup wizard handles all dependencies and configuration

## Quick Start

### Installation

#### Option 1: Install via go install (requires Go 1.21+)

The simplest way to install if you have Go installed:

```bash
go install github.com/yourusername/asc@latest
```

This will install the `asc` binary to `$GOPATH/bin` (usually `~/go/bin`). Make sure this directory is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Option 2: Build from source (requires Go 1.21+)

Clone the repository and build using the provided Makefile:

```bash
git clone https://github.com/yourusername/asc.git
cd asc

# Build for current platform
make build

# Install to $GOPATH/bin
make install

# Or build for all platforms
make build-all
```

The Makefile provides several useful targets:
- `make build` - Build for current platform
- `make build-all` - Build for all platforms (Linux, macOS Intel, macOS ARM)
- `make install` - Install to $GOPATH/bin
- `make test` - Run all tests
- `make clean` - Remove build artifacts
- `make help` - Show all available targets

#### Option 3: Download pre-built binary

Download the appropriate binary for your platform from the releases page:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/yourusername/asc/releases/latest/download/asc-darwin-arm64 -o asc
chmod +x asc
sudo mv asc /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/yourusername/asc/releases/latest/download/asc-darwin-amd64 -o asc
chmod +x asc
sudo mv asc /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/yourusername/asc/releases/latest/download/asc-linux-amd64 -o asc
chmod +x asc
sudo mv asc /usr/local/bin/
```

Verify the installation:

```bash
asc --version
```

### Initial Setup

Run the interactive setup wizard:

```bash
asc init
```

The wizard will:
1. Check for required dependencies (git, python3, uv, bd, docker)
2. Offer to install missing components
3. Backup any existing configuration files
4. Prompt for API keys (Claude, OpenAI, Google)
5. Generate default configuration files
6. Run a health check to verify everything works

## Usage

### Starting the Agent Stack

Launch all agents and the TUI dashboard:

```bash
asc up
```

This will:
- Start the mcp_agent_mail communication server
- Launch all configured agents as background processes
- Open the interactive TUI dashboard

### TUI Dashboard

Once running, you'll see a three-pane interface:

```
┌─────────────────┬─────────────────────────────────┐
│ Agent Status    │ Task Stream                     │
│                 │                                 │
│ ● main-planner  │ #42 [open] Implement auth       │
│   Idle          │ #43 [in_progress] Fix bug       │
│                 │ #44 [open] Add tests            │
│ ⟳ claude-coder  │                                 │
│   Working #43   │                                 │
│                 │                                 │
│ ○ test-agent    │                                 │
│   Offline       │                                 │
├─────────────────┴─────────────────────────────────┤
│ MCP Interaction Log                               │
│                                                   │
│ [10:30:15] [lease] claude-coder → src/auth.go    │
│ [10:30:18] [beads] claude-coder → claimed #43    │
│ [10:30:22] [message] test-agent → ready          │
└───────────────────────────────────────────────────┘
(q)uit | (r)efresh | (t)est
```

#### Keyboard Commands

- `q` - Quit and shut down all agents
- `r` - Force refresh all panes
- `t` - Run health check test

### Stopping the Agent Stack

Gracefully shut down all agents:

```bash
asc down
```

### Health Check

Verify your environment is properly configured:

```bash
asc check
```

Example output:

```
Dependency Check Results:
✓ git          Found at /usr/bin/git
✓ python3      Found at /usr/local/bin/python3
✓ uv           Found at /usr/local/bin/uv
✓ bd           Found at /usr/local/bin/bd
✓ asc.toml     Valid configuration
✓ .env         API keys present
```

### End-to-End Test

Test the full stack communication:

```bash
asc test
```

This creates a test task in beads, sends a test message through mcp_agent_mail, verifies both systems respond, and cleans up.

### Service Management

Control the mcp_agent_mail server independently:

```bash
# Start the communication server
asc services start

# Check server status
asc services status

# Stop the server
asc services stop
```

## Configuration

### Configuration File (asc.toml)

The `asc.toml` file defines your agent colony structure:

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

[agent.claude-coder]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation", "refactor"]

[agent.test-specialist]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing", "validation"]
```

#### Configuration Options

**[core]**
- `beads_db_path` - Path to your beads task database repository

**[services.mcp_agent_mail]**
- `start_command` - Command to start the MCP server
- `url` - HTTP endpoint for the MCP server

**[agent.{name}]**
- `command` - Command to execute the agent process
- `model` - LLM provider: `claude`, `gemini`, `gpt-4`, `codex`
- `phases` - Array of workflow phases this agent handles

### Environment Variables (.env)

Store your API keys in a `.env` file in the project root:

```bash
# Required for Claude models
CLAUDE_API_KEY=sk-ant-...

# Required for OpenAI models (GPT-4, Codex)
OPENAI_API_KEY=sk-...

# Required for Gemini models
GOOGLE_API_KEY=...
```

The `.env` file should have restricted permissions:

```bash
chmod 600 .env
```

### Agent Environment Variables

When asc launches agents, it automatically sets these environment variables:

- `AGENT_NAME` - The agent's name from asc.toml
- `AGENT_MODEL` - The LLM model to use
- `AGENT_PHASES` - Comma-separated list of phases
- `MCP_MAIL_URL` - URL of the mcp_agent_mail server
- `BEADS_DB_PATH` - Path to the beads database
- `CLAUDE_API_KEY` - Claude API key (if set)
- `OPENAI_API_KEY` - OpenAI API key (if set)
- `GOOGLE_API_KEY` - Google API key (if set)

## Architecture

### System Layers

asc operates as the orchestration layer (L4) in a multi-layered architecture:

- **L4 (Orchestration)**: asc manages process lifecycle and provides the TUI
- **L3 (Agents)**: Headless Python processes that execute development tasks
- **L1 (Communication)**: mcp_agent_mail handles async agent coordination
- **L0 (State)**: beads provides git-backed task persistence

### Agent Matrix Design

The system supports a matrix architecture where any LLM can fulfill any role:

```
           │ Planning │ Implementation │ Testing │ Refactor │
───────────┼──────────┼────────────────┼─────────┼──────────┤
Claude     │    ✓     │       ✓        │    ✓    │    ✓     │
Gemini     │    ✓     │       ✓        │    ✓    │    ✓     │
GPT-4      │    ✓     │       ✓        │    ✓    │    ✓     │
Codex      │    ✓     │       ✓        │    ✓    │    ✓     │
```

Multiple agents can handle the same phase, competing for tasks based on availability.

## Troubleshooting

### "Command not found: bd"

The beads CLI is not installed. Install it with:

```bash
pip install beads-cli
# or
uv pip install beads-cli
```

### "Failed to start mcp_agent_mail"

Ensure the mcp_agent_mail package is installed:

```bash
pip install mcp-agent-mail
# or
uv pip install mcp-agent-mail
```

Verify the start command in asc.toml matches your installation.

### "API key not found"

Check that your `.env` file exists and contains the required keys:

```bash
cat .env
```

Ensure the file has proper permissions:

```bash
chmod 600 .env
```

### "Agent stuck in 'Working' state"

An agent may have crashed. Check the logs:

```bash
cat ~/.asc/logs/<agent-name>.log
```

Restart the stack:

```bash
asc down
asc up
```

### "beads database not found"

Ensure the `beads_db_path` in asc.toml points to a valid git repository with beads initialized:

```bash
cd /path/to/project
bd init
```

### "Port already in use"

Another instance of mcp_agent_mail may be running. Stop it:

```bash
asc services stop
# or
pkill -f mcp_agent_mail
```

### TUI not rendering correctly

Ensure your terminal supports 256 colors and has sufficient size:

```bash
echo $TERM  # Should show something like xterm-256color
```

Resize your terminal to at least 80x24 characters.

### Agents not picking up tasks

1. Verify agents are running: Check the Agent Status pane
2. Check task phases match agent configuration
3. Verify beads database is accessible
4. Check agent logs for errors: `~/.asc/logs/<agent-name>.log`

## Development

### Building from Source

```bash
git clone https://github.com/yourusername/asc.git
cd asc

# Build for current platform
make build

# Or use go directly
go build -o asc main.go
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Or use go directly
go test ./...
go test -v -race -coverprofile=coverage.out ./...
```

### Development Workflow

```bash
# Format code
make fmt

# Run linter
make vet

# Run all checks (format, vet, test)
make check

# Build and run
make run

# Run in development mode with race detector
make dev
```

### Building for Multiple Platforms

```bash
# Build for all platforms (Linux, macOS Intel, macOS ARM)
make build-all

# Binaries will be in build/ directory:
# - build/asc-darwin-amd64
# - build/asc-darwin-arm64
# - build/asc-linux-amd64

# Or build manually for specific platforms
GOOS=linux GOARCH=amd64 go build -o asc-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o asc-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o asc-darwin-arm64
```

### Makefile Targets

Run `make help` to see all available targets:

```
make build       - Build the binary for current platform
make build-all   - Build binaries for all platforms
make test        - Run all tests
make install     - Install to $GOPATH/bin
make clean       - Remove build artifacts
make deps        - Download dependencies
make tidy        - Tidy and verify dependencies
make fmt         - Format Go code
make vet         - Run go vet
make lint        - Run golangci-lint
make check       - Run all checks (fmt, vet, test)
make release     - Prepare a release (build all platforms, run tests)
```

## Project Structure

```
asc/
├── cmd/                    # CLI command implementations
│   ├── root.go            # Root command and global flags
│   ├── init.go            # Interactive setup wizard
│   ├── up.go              # Start agents and TUI
│   ├── down.go            # Stop all agents
│   ├── check.go           # Dependency verification
│   ├── test.go            # End-to-end health check
│   └── services.go        # Service management
├── internal/
│   ├── config/            # Configuration parsing
│   ├── process/           # Process lifecycle management
│   ├── check/             # Dependency checking
│   ├── tui/               # Bubbletea TUI components
│   ├── beads/             # Beads database client
│   ├── mcp/               # MCP agent mail client
│   ├── logger/            # Logging utilities
│   └── errors/            # Error handling
├── asc.toml               # Configuration file
├── .env                   # API keys (gitignored)
└── main.go                # Entry point
```

## Dependencies

### Required

- **Go 1.21+** (for building from source)
- **git** - Version control
- **python3** - For running agents
- **beads (bd)** - Task database CLI
- **mcp_agent_mail** - Agent communication server

### Optional

- **uv** - Fast Python package manager
- **docker** - If agents require containerization

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- TUI powered by [Bubbletea](https://github.com/charmbracelet/bubbletea)
- Styling with [Lipgloss](https://github.com/charmbracelet/lipgloss)
- Integrates with [beads](https://github.com/steveyegge/beads) task management
- Uses [mcp_agent_mail](https://github.com/Dicklesworthstone/mcp_agent_mail) for coordination

## Support

- Report issues: https://github.com/yourusername/asc/issues
- Documentation: https://github.com/yourusername/asc/wiki
- Discussions: https://github.com/yourusername/asc/discussions
