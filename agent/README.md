# Agent Adapter

AI coding agent adapter for the Agent Stack Controller (ASC) system.

## Overview

The agent adapter is a Python framework that enables headless AI agents to execute software development tasks autonomously. It integrates with:

- **beads**: Git-backed task database for task management
- **mcp_agent_mail**: Asynchronous communication server for agent coordination
- **LLM Providers**: Claude (Anthropic), Gemini (Google), OpenAI (GPT-4, Codex)

## Features

- **Multi-LLM Support**: Unified interface for Claude, Gemini, and OpenAI models
- **Hephaestus Phase Loop**: Task polling, execution, and status management
- **ACE (Agentic Context Engineering)**: Playbook-based learning system
- **Heartbeat System**: Real-time status reporting to MCP
- **File Leasing**: Coordinated file access to prevent conflicts
- **Graceful Shutdown**: Signal handling for clean termination

## Installation

### From Source

```bash
cd agent
pip install -e .
```

### Dependencies

```bash
pip install -r requirements.txt
```

## Usage

### Environment Variables

The agent adapter requires the following environment variables:

- `AGENT_NAME`: Unique identifier for the agent
- `AGENT_MODEL`: LLM model to use (claude, gemini, gpt-4, codex)
- `AGENT_PHASES`: Comma-separated list of phases (planning, implementation, testing, etc.)
- `MCP_MAIL_URL`: URL of the mcp_agent_mail server (default: http://localhost:8765)
- `BEADS_DB_PATH`: Path to the beads repository (default: ./project-repo)

### API Keys

Set the appropriate API key based on your model:

- `CLAUDE_API_KEY`: For Claude models
- `GOOGLE_API_KEY`: For Gemini models
- `OPENAI_API_KEY`: For OpenAI models (GPT-4, Codex)

### Running the Agent

```bash
# Set environment variables
export AGENT_NAME="my-agent"
export AGENT_MODEL="claude"
export AGENT_PHASES="implementation,testing"
export MCP_MAIL_URL="http://localhost:8765"
export BEADS_DB_PATH="./project-repo"
export CLAUDE_API_KEY="sk-..."

# Run the agent
python -m agent.agent_adapter
```

### Using with ASC

The agent adapter is designed to be launched by the ASC orchestrator:

```bash
# Start the agent stack
asc up
```

ASC will automatically:
1. Load configuration from `asc.toml`
2. Set environment variables for each agent
3. Launch agent processes
4. Monitor agent status via heartbeats

## Architecture

### Components

#### 1. Agent Adapter (`agent_adapter.py`)

Main entry point that:
- Parses environment variables
- Initializes logging
- Sets up signal handlers
- Coordinates all components
- Runs the main event loop

#### 2. LLM Client (`llm_client.py`)

Unified interface for LLM providers:
- Abstract base class for consistency
- Provider-specific implementations (Claude, Gemini, OpenAI)
- Retry logic with exponential backoff
- Token counting and cost tracking

#### 3. Phase Loop (`phase_loop.py`)

Hephaestus pattern implementation:
- Polls beads for tasks matching agent phases
- Requests file leases from MCP
- Builds context from task and files
- Calls LLM with structured prompts
- Executes file operations
- Updates task status

#### 4. ACE Playbook (`ace.py`)

Learning system that:
- Reflects on completed tasks
- Extracts structured lessons
- Categorizes by task type
- Scores relevance
- Curates and prunes playbook
- Loads relevant lessons for new tasks

#### 5. Heartbeat Manager (`heartbeat.py`)

Status reporting system:
- Sends periodic heartbeats to MCP
- Reports status changes immediately
- Handles connection failures with backoff
- Continues working if MCP unavailable

### Data Flow

```
┌─────────────────────────────────────────────────────────┐
│ Agent Adapter                                           │
│                                                         │
│  ┌──────────────┐    ┌──────────────┐                 │
│  │ Phase Loop   │───▶│ LLM Client   │                 │
│  └──────────────┘    └──────────────┘                 │
│         │                    │                         │
│         │            ┌───────▼────────┐               │
│         │            │ ACE Playbook   │               │
│         │            └────────────────┘               │
│         │                                              │
│         │            ┌────────────────┐               │
│         └───────────▶│ Heartbeat Mgr  │               │
│                      └────────────────┘               │
└─────────────────────────────────────────────────────────┘
         │                      │
         ▼                      ▼
┌──────────────┐      ┌──────────────┐
│ beads (bd)   │      │ mcp_agent_   │
│              │      │ mail         │
└──────────────┘      └──────────────┘
```

## Development

### Project Structure

```
agent/
├── __init__.py           # Package initialization
├── agent_adapter.py      # Main entry point
├── llm_client.py         # LLM abstraction
├── phase_loop.py         # Task execution loop
├── ace.py                # Learning system
├── heartbeat.py          # Status reporting
├── requirements.txt      # Dependencies
├── setup.py              # Package setup
├── README.md             # This file
└── tests/                # Unit tests
    ├── __init__.py
    ├── test_llm_client.py
    ├── test_phase_loop.py
    ├── test_ace.py
    └── test_heartbeat.py
```

### Running Tests

```bash
# Install test dependencies
pip install pytest pytest-cov

# Run tests
pytest tests/

# Run with coverage
pytest --cov=agent tests/
```

### Adding a New LLM Provider

1. Create a new class inheriting from `LLMClient`
2. Implement the `complete()` method
3. Add provider detection in `create_llm_client()`
4. Update documentation

Example:

```python
class NewProviderClient(LLMClient):
    def __init__(self, model: str = "new-model"):
        super().__init__(model)
        # Initialize provider SDK
    
    def complete(self, prompt, system_prompt=None, max_tokens=4096, temperature=0.7):
        # Implement completion logic
        pass
```

## Configuration

### Agent Configuration (asc.toml)

```toml
[agent.my-planner]
command = "python -m agent.agent_adapter"
model = "gemini"
phases = ["planning", "design"]

[agent.my-coder]
command = "python -m agent.agent_adapter"
model = "claude"
phases = ["implementation", "refactor"]

[agent.my-tester]
command = "python -m agent.agent_adapter"
model = "gpt-4"
phases = ["testing"]
```

### Logging

Logs are written to `~/.asc/logs/{agent_name}.log` with the following format:

```
2024-11-09 10:30:00,123 [INFO] agent.my-agent: Agent my-agent starting up
2024-11-09 10:30:01,456 [INFO] agent.my-agent: Model: claude
2024-11-09 10:30:01,457 [INFO] agent.my-agent: Phases: implementation, refactor
```

### Playbook Storage

ACE playbooks are stored in `~/.asc/playbooks/{agent_name}/playbook.json`:

```json
{
  "agent_name": "my-agent",
  "updated_at": "2024-11-09T10:30:00",
  "lessons": [
    {
      "lesson_id": "abc123",
      "context": "Task in implementation phase: Add user authentication",
      "action": "Implemented JWT-based auth",
      "outcome": "success",
      "learned": "JWT tokens work well for stateless auth",
      "task_type": "implementation",
      "relevance_score": 1.5,
      "created_at": "2024-11-09T10:30:00"
    }
  ]
}
```

## Troubleshooting

### Agent Not Starting

Check logs at `~/.asc/logs/{agent_name}.log` for errors.

Common issues:
- Missing environment variables
- Invalid API keys
- beads repository not accessible
- MCP server not running

### Agent Not Picking Up Tasks

Verify:
- Task phase matches agent phases
- Task status is "open"
- beads repository is up to date (`git pull`)

### LLM API Errors

Check:
- API key is valid and has credits
- Network connectivity
- Rate limits not exceeded

### Heartbeat Failures

Verify:
- MCP server is running at `MCP_MAIL_URL`
- Network connectivity
- Check backoff time in logs

## License

MIT License - see LICENSE file for details

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## Support

For issues and questions:
- GitHub Issues: [link]
- Documentation: [link]
- Discord: [link]
