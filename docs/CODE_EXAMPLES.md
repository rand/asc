# Code Examples

Practical code examples for working with the Agent Stack Controller.

## Table of Contents

- [CLI Usage Examples](#cli-usage-examples)
- [Configuration Examples](#configuration-examples)
- [Go API Examples](#go-api-examples)
- [Python Agent Examples](#python-agent-examples)
- [Integration Examples](#integration-examples)
- [Automation Examples](#automation-examples)

---

## CLI Usage Examples

### Basic Workflow

```bash
# Initialize new project
cd my-project
asc init

# Start the stack
asc up

# In another terminal, check status
asc check

# Run health test
asc test

# Stop the stack
asc down
```

### Using Templates

```bash
# List available templates
asc init --list-templates

# Initialize with team template
asc init --template=team

# Save custom template
asc init --save-template my-custom-setup

# Use custom template later
cd new-project
asc init --template=my-custom-setup
```

### Secrets Management

```bash
# Encrypt secrets
asc secrets encrypt

# Check encryption status
asc secrets status

# Decrypt for editing
asc secrets decrypt
vim .env
asc secrets encrypt

# Rotate encryption key
asc secrets rotate
```

### Service Management

```bash
# Start MCP server only
asc services start

# Check server status
asc services status

# Restart server
asc services restart

# Stop server
asc services stop
```

### Diagnostics

```bash
# Run diagnostics
asc doctor

# Verbose output
asc doctor --verbose

# Auto-fix issues
asc doctor --fix

# JSON output
asc doctor --json
```

---

## Configuration Examples

### Minimal Setup

```toml
# asc.toml - Simplest configuration
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

### Team Setup

```toml
# asc.toml - Specialized team
[core]
beads_db_path = "./project"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

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

[agent.documenter]
command = "python agent_adapter.py"
model = "gemini"
phases = ["documentation"]
```

### High-Throughput Setup

```toml
# asc.toml - Multiple agents per phase
[core]
beads_db_path = "./project"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

# Planning team
[agent.planner-1]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning"]

[agent.planner-2]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning"]

# Implementation team
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

# Testing team
[agent.tester-1]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing"]

[agent.tester-2]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing"]
```

### Environment Variables

```bash
# .env - API keys and secrets
CLAUDE_API_KEY=sk-ant-api03-...
OPENAI_API_KEY=sk-proj-...
GOOGLE_API_KEY=AIzaSy...

# Optional: Custom settings
ASC_LOG_LEVEL=debug
ASC_REFRESH_INTERVAL=5
```

---

## Go API Examples

### Using the Config Package

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/yourusername/asc/internal/config"
)

func main() {
    // Load configuration
    cfg, err := config.Load("asc.toml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Validate configuration
    if err := cfg.Validate(); err != nil {
        log.Fatal(err)
    }
    
    // Access configuration
    fmt.Println("Beads DB:", cfg.Core.BeadsDBPath)
    fmt.Println("MCP URL:", cfg.Services.MCPAgentMail.URL)
    
    // Iterate agents
    for name, agent := range cfg.Agents {
        fmt.Printf("Agent %s: model=%s, phases=%v\n",
            name, agent.Model, agent.Phases)
    }
}
```

### Using the Process Manager

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/yourusername/asc/internal/process"
)

func main() {
    // Create process manager
    mgr := process.NewManager("~/.asc/pids", "~/.asc/logs")
    
    // Start a process
    env := []string{
        "AGENT_NAME=my-agent",
        "AGENT_MODEL=claude",
        "AGENT_PHASES=planning,implementation",
    }
    
    pid, err := mgr.Start("my-agent", "python agent_adapter.py", env)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Started agent with PID %d\n", pid)
    
    // Check if running
    if mgr.IsRunning(pid) {
        fmt.Println("Agent is running")
    }
    
    // Get status
    status := mgr.GetStatus(pid)
    fmt.Printf("Status: %s\n", status)
    
    // Wait a bit
    time.Sleep(5 * time.Second)
    
    // Stop the process
    if err := mgr.Stop(pid); err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Agent stopped")
}
```

### Using the Beads Client

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/yourusername/asc/internal/beads"
)

func main() {
    // Create beads client
    client := beads.NewClient("./project-repo")
    
    // Get open tasks
    tasks, err := client.GetTasks([]string{"open", "in_progress"})
    if err != nil {
        log.Fatal(err)
    }
    
    // Display tasks
    for _, task := range tasks {
        fmt.Printf("#%s: %s [%s]\n", task.ID, task.Title, task.Status)
    }
    
    // Create a new task
    newTask, err := client.CreateTask("Implement new feature")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created task #%s\n", newTask.ID)
    
    // Update task
    err = client.UpdateTask(newTask.ID, beads.TaskUpdate{
        Status: "in_progress",
        Assignee: "my-agent",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Refresh from git
    if err := client.Refresh(); err != nil {
        log.Fatal(err)
    }
}
```

### Using the MCP Client

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/yourusername/asc/internal/mcp"
)

func main() {
    // Create MCP client
    client := mcp.NewClient("http://localhost:8765")
    
    // Get agent statuses
    statuses, err := client.GetAllAgentStatuses()
    if err != nil {
        log.Fatal(err)
    }
    
    // Display statuses
    for _, status := range statuses {
        fmt.Printf("%s: %s", status.Name, status.State)
        if status.CurrentTask != "" {
            fmt.Printf(" (working on %s)", status.CurrentTask)
        }
        fmt.Println()
    }
    
    // Send a message
    msg := mcp.Message{
        Type:    mcp.TypeMessage,
        Source:  "my-app",
        Content: "Hello from Go!",
    }
    
    if err := client.SendMessage(msg); err != nil {
        log.Fatal(err)
    }
    
    // Get recent messages
    since := time.Now().Add(-1 * time.Hour)
    messages, err := client.GetMessages(since)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, msg := range messages {
        fmt.Printf("[%s] %s: %s\n",
            msg.Timestamp.Format("15:04:05"),
            msg.Source,
            msg.Content)
    }
}
```

---

## Python Agent Examples

### Basic Agent

```python
# simple_agent.py
import os
from agent_adapter import AgentAdapter
from llm_client import ClaudeClient

def main():
    # Get configuration from environment
    agent_name = os.getenv("AGENT_NAME")
    model = os.getenv("AGENT_MODEL")
    phases = os.getenv("AGENT_PHASES").split(",")
    
    # Create LLM client
    api_key = os.getenv("CLAUDE_API_KEY")
    llm_client = ClaudeClient(api_key=api_key)
    
    # Create and run agent
    agent = AgentAdapter(
        name=agent_name,
        model=model,
        phases=phases,
        llm_client=llm_client
    )
    
    agent.run()

if __name__ == "__main__":
    main()
```

### Custom LLM Client

```python
# custom_llm.py
from llm_client import LLMClient
import requests

class CustomLLMClient(LLMClient):
    def __init__(self, api_key: str, endpoint: str):
        self.api_key = api_key
        self.endpoint = endpoint
    
    def complete(self, prompt: str, context: dict) -> str:
        """Generate completion from custom LLM"""
        response = requests.post(
            self.endpoint,
            headers={"Authorization": f"Bearer {self.api_key}"},
            json={
                "prompt": prompt,
                "context": context,
                "max_tokens": 2000
            }
        )
        
        response.raise_for_status()
        return response.json()["completion"]

# Usage
client = CustomLLMClient(
    api_key=os.getenv("CUSTOM_API_KEY"),
    endpoint="https://api.custom-llm.com/v1/complete"
)

response = client.complete(
    prompt="Write a function to sort a list",
    context={"language": "python"}
)
```

### Custom Phase Loop

```python
# custom_phase_loop.py
from phase_loop import PhaseLoop
import logging

class CustomPhaseLoop(PhaseLoop):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.logger = logging.getLogger(__name__)
    
    def process_task(self, task):
        """Custom task processing logic"""
        self.logger.info(f"Processing task #{task.id}: {task.title}")
        
        # Custom pre-processing
        if task.phase == "planning":
            self.logger.info("Running planning-specific logic")
            # Add custom planning logic
        
        # Call parent implementation
        result = super().process_task(task)
        
        # Custom post-processing
        self.logger.info(f"Task #{task.id} completed")
        
        return result

# Usage
loop = CustomPhaseLoop(
    agent_name="my-agent",
    phases=["planning", "implementation"],
    llm_client=llm_client
)

loop.run()
```

### ACE Playbook Usage

```python
# ace_example.py
from ace import ACEPlaybook

# Create playbook
playbook = ACEPlaybook(agent_name="my-agent")

# Add a lesson after completing a task
playbook.add_lesson({
    "context": "implementing REST API",
    "action": "used FastAPI framework",
    "outcome": "successful, clean code",
    "learned": "FastAPI is great for quick API development",
    "task_type": "implementation",
    "relevance_score": 0.9
})

# Get relevant lessons for a new task
context = {
    "task": "implement GraphQL API",
    "task_type": "implementation"
}

lessons = playbook.get_relevant_lessons(context)

for lesson in lessons:
    print(f"Learned: {lesson['learned']}")
    print(f"Context: {lesson['context']}")
    print(f"Relevance: {lesson['relevance_score']}")
    print()
```

---

## Integration Examples

### Integrating with CI/CD

```yaml
# .github/workflows/asc.yml
name: ASC Integration

on:
  push:
    branches: [main]

jobs:
  test-with-asc:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Install asc
        run: |
          go install github.com/yourusername/asc@latest
      
      - name: Setup configuration
        run: |
          cat > asc.toml <<EOF
          [core]
          beads_db_path = "./project"
          
          [services.mcp_agent_mail]
          start_command = "python -m mcp_agent_mail.server"
          url = "http://localhost:8765"
          
          [agent.tester]
          command = "python agent_adapter.py"
          model = "claude"
          phases = ["testing"]
          EOF
      
      - name: Setup secrets
        env:
          CLAUDE_API_KEY: ${{ secrets.CLAUDE_API_KEY }}
        run: |
          echo "CLAUDE_API_KEY=$CLAUDE_API_KEY" > .env
      
      - name: Run tests with asc
        run: |
          asc up --no-tui &
          sleep 30
          asc test
          asc down
```

### Docker Integration

```dockerfile
# Dockerfile
FROM golang:1.21 AS builder

WORKDIR /app
COPY . .
RUN go build -o asc main.go

FROM python:3.11-slim

# Install dependencies
RUN apt-get update && apt-get install -y git
RUN pip install mcp-agent-mail beads-cli

# Copy asc binary
COPY --from=builder /app/asc /usr/local/bin/

# Copy agent code
COPY agent/ /app/agent/

WORKDIR /app

# Run asc
CMD ["asc", "up", "--no-tui"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  asc:
    build: .
    volumes:
      - ./asc.toml:/app/asc.toml
      - ./project:/app/project
      - ~/.asc:/root/.asc
    environment:
      - CLAUDE_API_KEY=${CLAUDE_API_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - GOOGLE_API_KEY=${GOOGLE_API_KEY}
    ports:
      - "8765:8765"
```

### Monitoring Integration

```python
# monitoring.py
import requests
import time
from prometheus_client import start_http_server, Gauge

# Prometheus metrics
agent_status = Gauge('asc_agent_status', 'Agent status', ['agent'])
task_count = Gauge('asc_task_count', 'Number of tasks', ['status'])

def collect_metrics():
    """Collect metrics from asc"""
    # Get agent statuses
    response = requests.get('http://localhost:8765/agents')
    agents = response.json()['agents']
    
    for agent in agents:
        status_value = {
            'idle': 0,
            'working': 1,
            'error': 2,
            'offline': 3
        }.get(agent['state'], 3)
        
        agent_status.labels(agent=agent['name']).set(status_value)
    
    # Get task counts (would need beads API)
    # task_count.labels(status='open').set(open_count)
    # task_count.labels(status='in_progress').set(in_progress_count)

if __name__ == '__main__':
    # Start Prometheus server
    start_http_server(9090)
    
    # Collect metrics every 10 seconds
    while True:
        collect_metrics()
        time.sleep(10)
```

---

## Automation Examples

### Automated Startup Script

```bash
#!/bin/bash
# start-asc.sh - Automated startup with checks

set -e

echo "Starting Agent Stack Controller..."

# Check dependencies
echo "Checking dependencies..."
if ! asc check; then
    echo "Dependency check failed!"
    exit 1
fi

# Decrypt secrets if needed
if [ ! -f .env ] && [ -f .env.age ]; then
    echo "Decrypting secrets..."
    asc secrets decrypt
fi

# Start the stack
echo "Starting agents..."
asc up --no-tui &
ASC_PID=$!

# Wait for startup
echo "Waiting for startup..."
sleep 10

# Run health check
echo "Running health check..."
if ! asc test; then
    echo "Health check failed!"
    kill $ASC_PID
    exit 1
fi

echo "Agent stack is running (PID: $ASC_PID)"
echo "Press Ctrl+C to stop"

# Wait for interrupt
trap "asc down; exit 0" INT TERM
wait $ASC_PID
```

### Monitoring Script

```bash
#!/bin/bash
# monitor-asc.sh - Monitor agent health

while true; do
    # Check if agents are running
    if ! pgrep -f agent_adapter > /dev/null; then
        echo "$(date): ALERT - No agents running!"
        # Send notification
        # curl -X POST https://hooks.slack.com/... -d '{"text":"ASC agents down!"}'
    fi
    
    # Check error rate
    ERROR_COUNT=$(grep ERROR ~/.asc/logs/asc.log | grep "$(date +%Y-%m-%d)" | wc -l)
    if [ $ERROR_COUNT -gt 10 ]; then
        echo "$(date): ALERT - High error rate: $ERROR_COUNT errors"
    fi
    
    # Check disk space
    DISK_FREE=$(df -h ~/.asc | tail -1 | awk '{print $4}' | sed 's/G//')
    if [ $(echo "$DISK_FREE < 1" | bc) -eq 1 ]; then
        echo "$(date): ALERT - Low disk space: ${DISK_FREE}GB"
    fi
    
    sleep 60
done
```

### Backup Script

```bash
#!/bin/bash
# backup-asc.sh - Automated backup

BACKUP_DIR=~/backups/asc
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p $BACKUP_DIR

echo "Backing up ASC configuration and state..."

# Backup configuration
tar -czf $BACKUP_DIR/config-$DATE.tar.gz \
    asc.toml .env.age

# Backup state
tar -czf $BACKUP_DIR/state-$DATE.tar.gz \
    ~/.asc/playbooks/ \
    ~/.asc/templates/ \
    ~/.asc/age.key

# Backup logs (last 7 days)
find ~/.asc/logs -name "*.log" -mtime -7 | \
    tar -czf $BACKUP_DIR/logs-$DATE.tar.gz -T -

# Keep only last 30 days of backups
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "Backup complete: $BACKUP_DIR"
```

---

## See Also

- [API Reference](API_REFERENCE.md)
- [Configuration Reference](CONFIGURATION.md)
- [Operator's Handbook](OPERATORS_HANDBOOK.md)
- [Agent Documentation](../agent/README.md)
