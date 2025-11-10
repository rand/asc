# Configuration Templates

The Agent Stack Controller (asc) provides a template system for quickly setting up common agent configurations. Templates allow you to start with pre-configured agent setups and customize them to your needs.

## Built-in Templates

### Solo Template (`--template=solo`)

A single agent configuration for individual development. This template creates one versatile agent that handles all phases of development.

**Use case:** Best for solo developers or small projects where a single agent can handle all tasks.

**Configuration:**
- 1 agent (solo-agent)
- Model: Claude
- Phases: planning, implementation, testing, review, refactor

### Team Template (`--template=team`)

A team configuration with specialized agents for different roles. This is the default template.

**Use case:** Ideal for most projects where you want specialized agents for different development phases.

**Configuration:**
- 3 agents:
  - planner (Gemini) - planning, design
  - coder (Claude) - implementation, coding
  - tester (GPT-4) - testing, review

### Swarm Template (`--template=swarm`)

A swarm configuration with multiple agents per phase for parallel work.

**Use case:** Best for large projects or teams that need high throughput and parallel task execution.

**Configuration:**
- 8 agents:
  - 2 planners (Gemini, Claude) - planning, design
  - 3 coders (Claude, GPT-4, Codex) - implementation, coding
  - 2 testers (GPT-4, Gemini) - testing, review
  - 1 refactor agent (Claude) - refactor, optimization

## Using Templates

### List Available Templates

To see all available templates (built-in and custom):

```bash
asc init --list-templates
```

### Initialize with a Template

To initialize asc with a specific template:

```bash
# Use solo template
asc init --template=solo

# Use team template (default)
asc init --template=team

# Use swarm template
asc init --template=swarm
```

When you run `asc init` without the `--template` flag, you'll be presented with an interactive template selection menu where you can choose from all available templates.

### Interactive Template Selection

When running `asc init` without specifying a template, you'll see an interactive menu:

```
📋 Select Configuration Template

Choose a template for your agent setup:

Built-in Templates:
▶ 1. solo
     Single agent setup for individual development
  2. team
     Team setup with planner, coder, and tester agents
  3. swarm
     Swarm setup with multiple agents per phase for parallel work

↑/↓ or j/k: Navigate | 1-9: Quick select | Enter: Confirm | q: Quit
```

Navigation:
- Use arrow keys (↑/↓) or vim keys (j/k) to navigate
- Press number keys (1-9) for quick selection
- Press Enter to confirm selection
- Press q to quit

## Custom Templates

### Creating Custom Templates

You can save your current configuration as a custom template for reuse:

```bash
# First, create and configure your asc.toml
# Then save it as a template
asc init --save-template my-custom-setup
```

This saves your current `asc.toml` to `~/.asc/templates/my-custom-setup.toml`.

### Using Custom Templates

Custom templates appear in the template list and can be used just like built-in templates:

```bash
# List templates (includes custom templates)
asc init --list-templates

# Use custom template
asc init --template=my-custom-setup
```

### Managing Custom Templates

Custom templates are stored in `~/.asc/templates/` as TOML files. You can:

- **View templates:** `ls ~/.asc/templates/`
- **Edit templates:** Edit the `.toml` files directly
- **Delete templates:** `rm ~/.asc/templates/template-name.toml`
- **Share templates:** Copy `.toml` files to other machines

### Custom Template Example

Here's an example of creating a custom template for a specialized workflow:

```bash
# Create a custom configuration
cat > asc.toml << 'EOF'
[core]
beads_db_path = "./my-project"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.architect]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["planning", "design"]

[agent.frontend-dev]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation"]

[agent.backend-dev]
command = "python agent_adapter.py"
model = "codex"
phases = ["implementation"]

[agent.qa-engineer]
command = "python agent_adapter.py"
model = "gemini"
phases = ["testing", "review"]
EOF

# Save as custom template
asc init --save-template fullstack-team

# Now you can use it
asc init --template=fullstack-team
```

## Template Structure

All templates follow the standard asc.toml format:

```toml
[core]
beads_db_path = "./project-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.agent-name]
command = "python agent_adapter.py"
model = "claude"  # claude, gemini, gpt-4, codex, openai
phases = ["planning", "implementation", "testing"]
```

### Supported Models

- `claude` - Anthropic Claude (best for implementation and refactoring)
- `gemini` - Google Gemini (good for planning and design)
- `gpt-4` - OpenAI GPT-4 (excellent for testing and review)
- `codex` - OpenAI Codex (specialized for code generation)
- `openai` - Generic OpenAI model

### Supported Phases

- `planning` - High-level planning and task breakdown
- `design` - Architecture and design decisions
- `implementation` - Writing code
- `coding` - Code generation and modification
- `testing` - Writing and running tests
- `review` - Code review and quality checks
- `refactor` - Code refactoring and optimization
- `documentation` - Writing documentation
- `debugging` - Debugging and troubleshooting
- `optimization` - Performance optimization
- `deployment` - Deployment tasks

## Best Practices

### Choosing a Template

1. **Solo Template** - Use when:
   - Working on small projects
   - Learning the system
   - Limited API budget
   - Simple workflows

2. **Team Template** - Use when:
   - Working on medium-sized projects
   - Want specialized agents
   - Need clear separation of concerns
   - This is the recommended default

3. **Swarm Template** - Use when:
   - Working on large projects
   - Need high throughput
   - Have multiple parallel tasks
   - Want redundancy and competition

### Customizing Templates

After initializing with a template, you can customize the generated `asc.toml`:

1. Add or remove agents
2. Change agent models
3. Adjust phase assignments
4. Modify service configurations

### Template Naming

When creating custom templates, use descriptive names:

- ✅ Good: `fullstack-team`, `ml-pipeline`, `microservices-dev`
- ❌ Bad: `template1`, `test`, `my-config`

### Sharing Templates

To share templates with your team:

1. Save your template: `asc init --save-template team-standard`
2. Copy the file: `~/.asc/templates/team-standard.toml`
3. Share with team members
4. Team members place it in their `~/.asc/templates/` directory

## Troubleshooting

### Template Not Found

If you get a "template not found" error:

```bash
# List available templates
asc init --list-templates

# Check custom templates directory
ls ~/.asc/templates/
```

### Template Validation Errors

If a template fails validation:

1. Check that all required fields are present
2. Verify agent commands exist in PATH
3. Ensure models are supported
4. Validate phase names

### Custom Template Not Appearing

If your custom template doesn't appear in the list:

1. Verify the file is in `~/.asc/templates/`
2. Check the file has `.toml` extension
3. Ensure the file is readable
4. Verify the TOML syntax is valid

## Examples

### Example 1: Quick Start with Solo Template

```bash
# Initialize with solo template
asc init --template=solo

# Follow the wizard to configure API keys
# Start the agent stack
asc up
```

### Example 2: Team Setup with Custom Template

```bash
# Create custom configuration
vim asc.toml

# Save as template
asc init --save-template my-team

# Use on another project
cd ../new-project
asc init --template=my-team
```

### Example 3: Swarm for High Throughput

```bash
# Initialize with swarm template
asc init --template=swarm

# Configure API keys
# Start the swarm
asc up

# Monitor all agents in TUI
# Multiple agents will compete for tasks
```

## See Also

- [Configuration Guide](../README.md#configuration)
- [Agent Setup](../README.md#agent-configuration)
- [Getting Started](../README.md#getting-started)
