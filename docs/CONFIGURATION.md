# Configuration Reference

Complete guide to configuring the Agent Stack Controller.

## Table of Contents

- [Configuration Files](#configuration-files)
- [Core Configuration](#core-configuration)
- [Service Configuration](#service-configuration)
- [Agent Configuration](#agent-configuration)
- [Environment Variables](#environment-variables)
- [Templates](#templates)
- [Advanced Configuration](#advanced-configuration)
- [Examples](#examples)

---

## Configuration Files

### asc.toml

Main configuration file defining agents and services.

**Location:** Project root (default: `./asc.toml`)

**Format:** TOML (Tom's Obvious, Minimal Language)

**Example:**
```toml
[core]
beads_db_path = "./project-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.my-agent]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning", "implementation"]
```

### .env

Environment variables and API keys.

**Location:** Project root (default: `./.env`)

**Format:** KEY=VALUE pairs

**Security:** 
- Automatically gitignored
- File permissions set to 0600
- Should be encrypted to `.env.age`

**Example:**
```bash
CLAUDE_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=AIza...
```

### .env.age

Encrypted secrets file (safe to commit).

**Location:** Project root (default: `./.env.age`)

**Format:** age-encrypted binary

**Usage:**
```bash
# Encrypt
asc secrets encrypt

# Decrypt
asc secrets decrypt
```

---

## Core Configuration

### [core] Section

Core system settings.

#### beads_db_path

Path to the beads task database repository.

**Type:** String (path)  
**Required:** Yes  
**Default:** None

**Example:**
```toml
[core]
beads_db_path = "./project-repo"
```

**Notes:**
- Must be a valid git repository
- Must have beads initialized (`bd init`)
- Can be relative or absolute path
- Tilde (`~`) expansion supported

---

## Service Configuration

### [services.mcp_agent_mail] Section

Configuration for the MCP agent mail communication server.

#### start_command

Command to start the MCP server.

**Type:** String (command)  
**Required:** Yes  
**Default:** None

**Example:**
```toml
[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
```

**Notes:**
- Command is executed in a shell
- Should start a long-running process
- Process is managed by asc

#### url

HTTP endpoint for the MCP server.

**Type:** String (URL)  
**Required:** Yes  
**Default:** None

**Example:**
```toml
[services.mcp_agent_mail]
url = "http://localhost:8765"
```

**Notes:**
- Must be a valid HTTP URL
- Used by agents and TUI to connect
- Should match server configuration

---

## Agent Configuration

### [agent.{name}] Sections

Each agent is defined in its own section.

**Section Name Format:** `[agent.{name}]`  
**Name Requirements:**
- Alphanumeric and hyphens only
- Must be unique
- Used as agent identifier

#### command

Command to execute the agent process.

**Type:** String (command)  
**Required:** Yes  
**Default:** None

**Example:**
```toml
[agent.my-planner]
command = "python agent_adapter.py"
```

**Notes:**
- Command is executed in a shell
- Should start a long-running process
- Receives environment variables from asc

#### model

LLM model provider for the agent.

**Type:** String (enum)  
**Required:** Yes  
**Default:** None  
**Valid Values:** `claude`, `gemini`, `gpt-4`, `codex`

**Example:**
```toml
[agent.my-planner]
model = "claude"
```

**Model Details:**

| Model | Provider | API Key Required | Notes |
|-------|----------|------------------|-------|
| `claude` | Anthropic | `CLAUDE_API_KEY` | Claude Sonnet/Opus |
| `gemini` | Google | `GOOGLE_API_KEY` | Gemini Pro |
| `gpt-4` | OpenAI | `OPENAI_API_KEY` | GPT-4 Turbo |
| `codex` | OpenAI | `OPENAI_API_KEY` | Codex (deprecated) |

#### phases

Workflow phases the agent handles.

**Type:** Array of strings  
**Required:** Yes  
**Default:** None  
**Valid Values:** See phase list below

**Example:**
```toml
[agent.my-planner]
phases = ["planning", "design"]
```

**Valid Phases:**

| Phase | Description | Typical Tasks |
|-------|-------------|---------------|
| `planning` | Task planning and breakdown | Create task plans, estimate effort |
| `design` | Architecture and design | Design systems, create diagrams |
| `implementation` | Code implementation | Write code, implement features |
| `refactor` | Code refactoring | Improve code quality, optimize |
| `testing` | Test writing | Write unit/integration tests |
| `validation` | Test validation | Run tests, verify functionality |
| `documentation` | Documentation writing | Write docs, comments, guides |

**Notes:**
- Multiple agents can handle the same phase
- Agents compete for tasks in their phases
- Order doesn't matter

---

## Environment Variables

### System Variables

Set automatically by asc when launching agents.

#### AGENT_NAME

Agent identifier from configuration.

**Type:** String  
**Set by:** asc  
**Example:** `my-planner`

#### AGENT_MODEL

LLM model for the agent.

**Type:** String  
**Set by:** asc  
**Example:** `claude`

#### AGENT_PHASES

Comma-separated list of phases.

**Type:** String (comma-separated)  
**Set by:** asc  
**Example:** `planning,design`

#### MCP_MAIL_URL

URL of the MCP server.

**Type:** String (URL)  
**Set by:** asc  
**Example:** `http://localhost:8765`

#### BEADS_DB_PATH

Path to the beads repository.

**Type:** String (path)  
**Set by:** asc  
**Example:** `./project-repo`

### User Variables

Set by user in `.env` file.

#### CLAUDE_API_KEY

Anthropic Claude API key.

**Type:** String  
**Required for:** `model = "claude"`  
**Format:** `sk-ant-...`

**Example:**
```bash
CLAUDE_API_KEY=sk-ant-api03-...
```

#### OPENAI_API_KEY

OpenAI API key.

**Type:** String  
**Required for:** `model = "gpt-4"` or `model = "codex"`  
**Format:** `sk-...`

**Example:**
```bash
OPENAI_API_KEY=sk-proj-...
```

#### GOOGLE_API_KEY

Google AI API key.

**Type:** String  
**Required for:** `model = "gemini"`  
**Format:** `AIza...`

**Example:**
```bash
GOOGLE_API_KEY=AIzaSy...
```

---

## Templates

Pre-configured agent setups for common scenarios.

### Using Templates

```bash
# List available templates
asc init --list-templates

# Use a template
asc init --template=team

# Save custom template
asc init --save-template my-setup
```

### Built-in Templates

#### solo

Single agent handling all phases.

**Use case:** Personal projects, learning, experimentation

**Configuration:**
```toml
[agent.solo-agent]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning", "design", "implementation", "testing", "documentation"]
```

#### team (default)

Three specialized agents: planner, coder, tester.

**Use case:** Balanced workflow, most projects

**Configuration:**
```toml
[agent.planner]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning", "design"]

[agent.coder]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation", "refactor"]

[agent.tester]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing", "validation"]
```

#### swarm

Multiple agents per phase for parallel work.

**Use case:** Large projects, high throughput

**Configuration:**
```toml
[agent.planner-1]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning"]

[agent.planner-2]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning"]

[agent.coder-1]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation"]

[agent.coder-2]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["implementation"]

[agent.tester-1]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing"]
```

### Custom Templates

Create your own templates:

1. Configure asc.toml as desired
2. Save as template:
   ```bash
   asc init --save-template my-custom-setup
   ```
3. Use later:
   ```bash
   asc init --template=my-custom-setup
   ```

**Template Location:** `~/.asc/templates/`

---

## Advanced Configuration

### Multiple Configuration Files

Use different configs for different environments:

```bash
# Development
asc up --config=asc.dev.toml

# Production
asc up --config=asc.prod.toml

# Testing
asc up --config=asc.test.toml
```

### Configuration Validation

Validate configuration without starting:

```bash
asc check
```

### Hot-Reload

Configuration changes are detected automatically:

- New agents are started
- Removed agents are stopped
- Modified agents are restarted

**Watched files:**
- `asc.toml`

**Not watched:**
- `.env` (requires restart)

### Configuration Overrides

Override config values via environment variables:

```bash
# Override beads path
export ASC_BEADS_DB_PATH=/custom/path
asc up

# Override MCP URL
export ASC_MCP_URL=http://custom:8765
asc up
```

**Environment Variable Format:**
- Prefix: `ASC_`
- Uppercase
- Underscores for nesting
- Example: `ASC_SERVICES_MCP_AGENT_MAIL_URL`

---

## Examples

### Minimal Configuration

Simplest possible setup:

```toml
[core]
beads_db_path = "./project"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.main]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning", "implementation", "testing"]
```

### Multi-Model Setup

Different models for different roles:

```toml
[core]
beads_db_path = "./project"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

# Gemini for planning (fast, creative)
[agent.planner]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning", "design"]

# Claude for coding (accurate, detailed)
[agent.coder]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation", "refactor"]

# GPT-4 for testing (thorough, edge cases)
[agent.tester]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing", "validation"]

# Gemini for docs (clear, concise)
[agent.documenter]
command = "python agent_adapter.py"
model = "gemini"
phases = ["documentation"]
```

### Specialized Agents

Agents with specific focus areas:

```toml
[core]
beads_db_path = "./project"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

# Frontend specialist
[agent.frontend]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation"]
# Note: Phase filtering happens in agent logic

# Backend specialist
[agent.backend]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["implementation"]

# Test specialist
[agent.qa]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing", "validation"]

# Refactoring specialist
[agent.refactorer]
command = "python agent_adapter.py"
model = "claude"
phases = ["refactor"]
```

### High-Throughput Setup

Maximum parallelism:

```toml
[core]
beads_db_path = "./project"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

# 2 planners
[agent.planner-1]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning"]

[agent.planner-2]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning"]

# 4 coders
[agent.coder-1]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation"]

[agent.coder-2]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation"]

[agent.coder-3]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["implementation"]

[agent.coder-4]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["implementation"]

# 2 testers
[agent.tester-1]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing"]

[agent.tester-2]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing"]
```

---

## Troubleshooting

### Configuration Errors

**Error:** `Invalid configuration: missing required field 'beads_db_path'`

**Solution:** Add `beads_db_path` to `[core]` section

---

**Error:** `Unknown model: gpt-3.5`

**Solution:** Use valid model: `claude`, `gemini`, `gpt-4`, or `codex`

---

**Error:** `Invalid phase: coding`

**Solution:** Use valid phase: `planning`, `design`, `implementation`, `refactor`, `testing`, `validation`, or `documentation`

---

**Error:** `Duplicate agent name: my-agent`

**Solution:** Ensure all agent names are unique

---

### Validation

Run validation before starting:

```bash
asc check
```

### Schema Validation

The configuration schema is validated on load. Common issues:

- Missing required fields
- Invalid types (string vs array)
- Unknown fields (typos)
- Invalid values (unknown model/phase)

---

## See Also

- [API Reference](API_REFERENCE.md)
- [Templates Documentation](TEMPLATES.md)
- [Operator's Handbook](OPERATORS_HANDBOOK.md)
- [Troubleshooting](../TROUBLESHOOTING.md)
