# Frequently Asked Questions (FAQ)

Common questions and answers about the Agent Stack Controller.

## Table of Contents

- [General](#general)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Agents](#agents)
- [Security](#security)
- [Troubleshooting](#troubleshooting)
- [Performance](#performance)
- [Development](#development)

---

## General

### What is the Agent Stack Controller?

The Agent Stack Controller (asc) is a command-line orchestration tool that manages a local colony of AI coding agents. It provides a mission control interface for starting, monitoring, and coordinating headless background agents that work collaboratively on software development tasks.

### What are the main components?

- **asc (CLI/TUI)**: The orchestration layer that manages everything
- **Agents**: Headless Python processes that execute development tasks using LLMs
- **mcp_agent_mail**: Communication server for agent coordination
- **beads**: Git-backed task database for persistent task state

### What LLMs are supported?

- Anthropic Claude (Sonnet, Opus)
- Google Gemini
- OpenAI GPT-4
- OpenAI Codex

### Is this free to use?

The asc tool itself is open source (MIT license), but you'll need API keys for the LLM providers, which have their own pricing.

### Can I use this for commercial projects?

Yes, the MIT license allows commercial use. Check the license terms of your chosen LLM providers.

---

## Installation

### What are the system requirements?

- **OS**: Linux or macOS (Windows via WSL)
- **Go**: 1.21+ (for building from source)
- **Python**: 3.8+
- **Git**: Any recent version
- **Disk**: ~100MB for binaries and dependencies
- **Memory**: ~500MB per agent

### Do I need to install Go to use asc?

No, if you download a pre-built binary. Yes, if you want to build from source.

### How do I install on macOS?

```bash
# Using Homebrew (if available)
brew install asc

# Or download binary
curl -L https://github.com/yourusername/asc/releases/latest/download/asc-darwin-arm64 -o asc
chmod +x asc
sudo mv asc /usr/local/bin/
```

### How do I install on Linux?

```bash
# Download binary
curl -L https://github.com/yourusername/asc/releases/latest/download/asc-linux-amd64 -o asc
chmod +x asc
sudo mv asc /usr/local/bin/

# Or use go install
go install github.com/yourusername/asc@latest
```

### Can I install without sudo?

Yes, place the binary in `~/bin` or any directory in your PATH:

```bash
mkdir -p ~/bin
mv asc ~/bin/
export PATH=$PATH:~/bin
```

---

## Configuration

### Where is the configuration file?

By default, `asc.toml` in your project root. You can specify a different location with `--config`.

### How do I configure multiple agents?

Add multiple `[agent.name]` sections to `asc.toml`:

```toml
[agent.planner]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning"]

[agent.coder]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation"]
```

### Can agents use different LLMs?

Yes! That's the matrix architecture. Each agent can use any LLM for any role.

### How do I add API keys?

Run `asc init` for interactive setup, or manually create `.env`:

```bash
CLAUDE_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=AIza...
```

### Should I commit my API keys?

**Never commit `.env`!** Use `asc secrets encrypt` to create `.env.age`, which is safe to commit.

### Can I use environment-specific configs?

Yes:

```bash
asc up --config=asc.dev.toml
asc up --config=asc.prod.toml
```

---

## Usage

### How do I start the agent stack?

```bash
asc up
```

This starts all agents and launches the TUI dashboard.

### How do I stop everything?

Press `q` in the TUI, or run:

```bash
asc down
```

### Can I run without the TUI?

Yes:

```bash
asc up --no-tui
```

### How do I check if everything is working?

```bash
asc check  # Check dependencies
asc test   # Run health check
```

### What keyboard shortcuts are available in the TUI?

- `q` - Quit and shut down
- `r` - Force refresh
- `t` - Run health check
- `1-9` - Select agent
- `p` - Pause/resume agent
- `k` - Kill agent
- `/` - Search logs
- `e` - Export logs

### How do I view logs?

Logs are in `~/.asc/logs/`:

```bash
# View agent log
tail -f ~/.asc/logs/agent-name.log

# View asc log
tail -f ~/.asc/logs/asc.log
```

---

## Agents

### How many agents should I run?

Start with 1-3 agents. Add more as needed:
- **1 agent**: Simple projects, learning
- **3 agents**: Balanced (planner, coder, tester)
- **5+ agents**: Large projects, high throughput

### Can multiple agents work on the same task?

No, agents claim tasks exclusively. Multiple agents can work on different tasks in the same phase.

### What happens if an agent crashes?

The health monitor detects crashes and can auto-restart agents (if enabled).

### How do agents coordinate?

Through mcp_agent_mail:
- File leases prevent conflicts
- Messages coordinate work
- Heartbeats track status

### Can I pause an agent?

Yes, in the TUI:
1. Select agent with number key
2. Press `p` to pause/resume

Or manually:

```bash
kill -STOP <pid>  # Pause
kill -CONT <pid>  # Resume
```

### How do I add a new agent?

1. Add to `asc.toml`:
   ```toml
   [agent.new-agent]
   command = "python agent_adapter.py"
   model = "claude"
   phases = ["implementation"]
   ```
2. Configuration hot-reloads automatically
3. Or restart: `asc down && asc up`

---

## Security

### Are my API keys secure?

Yes, if you follow best practices:
- Use `asc secrets encrypt` to encrypt `.env`
- Never commit `.env` (automatically gitignored)
- File permissions set to 0600
- Keys never logged

### What is age encryption?

age is a simple, modern encryption tool. asc uses it to encrypt your `.env` file to `.env.age`, which is safe to commit to git.

### How do I rotate my encryption key?

```bash
asc secrets rotate
```

This generates a new key and re-encrypts your secrets.

### Can I share encrypted secrets with my team?

Yes:

```bash
# Get your public key
asc secrets status

# Team member encrypts for you
age -r <your-public-key> -o .env.age .env
```

### What if I lose my age key?

You'll need to:
1. Regenerate API keys from providers
2. Create new age key
3. Re-encrypt secrets

**Always backup your age key!**

---

## Troubleshooting

### "Command not found: asc"

The binary isn't in your PATH. Either:
- Move it to `/usr/local/bin/`
- Add its location to PATH
- Use full path: `/path/to/asc`

### "Command not found: bd"

Install beads:

```bash
pip install beads-cli
# or
uv pip install beads-cli
```

### "Failed to start mcp_agent_mail"

Install mcp_agent_mail:

```bash
pip install mcp-agent-mail
# or
uv pip install mcp-agent-mail
```

### "API key not found"

Check your `.env` file:

```bash
cat .env
```

If it doesn't exist, decrypt it:

```bash
asc secrets decrypt
```

Or run setup:

```bash
asc init
```

### "Agent stuck in 'Working' state"

The agent may have crashed. Check logs:

```bash
cat ~/.asc/logs/agent-name.log
```

Restart:

```bash
asc down
asc up
```

### "Port already in use"

Another instance is running. Stop it:

```bash
asc down
# or
pkill -f mcp_agent_mail
```

### TUI looks broken

Ensure your terminal:
- Supports 256 colors
- Is at least 80x24 characters
- Has proper TERM variable:
  ```bash
  echo $TERM  # Should be xterm-256color or similar
  ```

### How do I get more detailed error messages?

Run in debug mode:

```bash
asc up --debug
```

Or check logs:

```bash
tail -f ~/.asc/logs/asc.log
```

---

## Performance

### How much memory does asc use?

- asc itself: ~50MB
- Each agent: ~200-500MB
- MCP server: ~50MB
- Total for 3 agents: ~1-2GB

### How much does it cost to run?

Depends on LLM usage:
- Claude: ~$0.01-0.10 per task
- GPT-4: ~$0.02-0.20 per task
- Gemini: ~$0.001-0.01 per task

Actual costs vary by task complexity.

### Can I limit API usage?

Not directly in asc, but you can:
- Set rate limits in agent code
- Use cheaper models (Gemini)
- Limit number of agents
- Monitor usage in provider dashboards

### How fast are agents?

Typical task times:
- Planning: 30-60 seconds
- Implementation: 2-5 minutes
- Testing: 1-3 minutes

Depends on task complexity and LLM speed.

### Can I run this on a server?

Yes, but you'll need:
- Headless mode: `asc up --no-tui`
- Process manager (systemd, supervisor)
- Remote access (SSH, tmux)

---

## Development

### How do I contribute?

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

### How do I build from source?

```bash
git clone https://github.com/yourusername/asc.git
cd asc
make build
```

### How do I run tests?

```bash
make test
```

### How do I add a new LLM provider?

1. Implement `LLMClient` interface in `agent/llm_client.py`
2. Add to model validation in `internal/config/config.go`
3. Update documentation
4. Submit PR

### Can I customize the TUI?

Yes! The TUI is built with bubbletea and lipgloss. See:
- `internal/tui/` for TUI code
- `internal/tui/theme.go` for styling
- [Vaporwave Design](VAPORWAVE_DESIGN.md) for aesthetic

### How do I add a new CLI command?

1. Create `cmd/mycommand.go`
2. Implement cobra command
3. Register in `cmd/root.go`
4. Add tests
5. Update documentation

---

## Still Have Questions?

- Check the [Documentation](README.md)
- Read the [Troubleshooting Guide](../TROUBLESHOOTING.md)
- Search [GitHub Issues](https://github.com/yourusername/asc/issues)
- Ask in [Discussions](https://github.com/yourusername/asc/discussions)
- Read the [API Reference](API_REFERENCE.md)

---

## Contributing to FAQ

Found a question that should be here? Please:
1. Open an issue
2. Submit a PR adding it
3. Ask in discussions

We update this FAQ based on common user questions.
