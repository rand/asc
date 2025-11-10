# Next Phase Implementation Tasks

> **⚠️ ARCHIVED DOCUMENT**  
> This document outlined planned tasks that have now been completed or superseded.  
> For current tasks, see:
> - [Task List](../../.kiro/specs/agent-stack-controller/tasks.md) - Active implementation tasks
> - [Agent README](../../agent/README.md) - Agent adapter documentation
>
> Archived: 2024-11-09  
> Reason: Task 21 (Agent Adapter) completed. Remaining tasks tracked in Kiro spec.

## Overview

This document outlines the critical tasks needed to complete the asc (Agent Stack Controller) vision. The current implementation has excellent infrastructure but is missing the agent execution layer.

## 🚨 Phase 1: Make It Work (CRITICAL - 2-3 weeks)

### Task 21: Implement Agent Adapter Framework

**Goal**: Create the Python agent adapter that asc launches to run actual AI agents.

#### 21.1 Create agent_adapter.py entry point
**Priority**: CRITICAL
**Estimated Time**: 2-3 days

```python
# agent_adapter.py - Main entry point for headless agents

Requirements:
- Parse environment variables (AGENT_NAME, AGENT_MODEL, AGENT_PHASES, MCP_MAIL_URL, BEADS_DB_PATH)
- Initialize logging to ~/.asc/logs/{agent_name}.log
- Set up signal handlers for graceful shutdown
- Initialize LLM client based on AGENT_MODEL
- Enter main event loop
- Handle errors and restart logic

Deliverables:
- agent_adapter.py with CLI argument parsing
- Configuration validation
- Logging setup
- Error handling framework
- Unit tests for initialization
```

#### 21.2 Implement LLM client abstraction
**Priority**: CRITICAL
**Estimated Time**: 3-4 days

```python
# llm_client.py - Abstract LLM client interface

Requirements:
- Create base LLMClient abstract class
- Define interface: complete(prompt, context, max_tokens, temperature)
- Implement ClaudeClient using Anthropic SDK
- Implement GeminiClient using Google AI SDK
- Implement OpenAIClient using OpenAI SDK
- Handle API errors, rate limiting, retries
- Add streaming support for long responses
- Add token counting and cost tracking

Deliverables:
- llm_client.py with base class
- claude_client.py implementation
- gemini_client.py implementation
- openai_client.py implementation
- Unit tests with mocked API responses
- Integration tests with real APIs (optional)
```

#### 21.3 Implement Hephaestus phase loop
**Priority**: CRITICAL
**Estimated Time**: 4-5 days

```python
# phase_loop.py - Main agent execution loop

Requirements:
- Poll beads for tasks matching agent phases
  - Use: bd ready --json --phase {phase}
  - Parse JSON response
  - Select highest priority task
  
- Request file leases via mcp_agent_mail
  - POST /leases with file paths
  - Wait for lease confirmation
  - Handle lease conflicts
  
- Build context from task and files
  - Read task description from beads
  - Read relevant files from filesystem
  - Load playbook lessons
  - Format as structured prompt
  
- Call LLM with structured prompts
  - Send context to LLM client
  - Parse LLM response
  - Extract action plan
  - Validate plan structure
  
- Execute LLM-generated plans
  - Parse plan into discrete steps
  - Execute file operations (read, write, delete)
  - Run shell commands (with safety checks)
  - Capture execution results
  
- Update beads task status
  - Use: bd update {task_id} --status {status}
  - Add execution notes
  - Link to changed files
  
- Release file leases
  - POST /leases/{id}/release
  - Handle release errors

Deliverables:
- phase_loop.py with main loop
- Task polling logic
- Lease management
- Context building
- Plan execution engine
- Status update logic
- Unit tests for each component
- Integration test with mock beads/MCP
```

#### 21.4 Implement ACE (Agentic Context Engineering)
**Priority**: HIGH
**Estimated Time**: 3-4 days

```python
# ace.py - Reflection and curation system

Requirements:
- Create playbook storage structure
  - JSON/YAML files in ~/.asc/playbooks/{agent_name}/
  - Schema: {lesson_id, context, action, outcome, learned}
  - Version control playbooks
  
- Implement reflection prompt after task completion
  - Ask LLM: "What went well? What could improve?"
  - Extract structured lessons
  - Categorize by task type
  
- Extract lessons learned from LLM reflection
  - Parse reflection response
  - Identify actionable insights
  - Score lesson relevance
  
- Curate and store lessons in playbook
  - Deduplicate similar lessons
  - Merge related lessons
  - Prune outdated lessons
  - Maintain max playbook size
  
- Load playbook into context for future tasks
  - Select relevant lessons for current task
  - Format as "learned patterns"
  - Include in LLM prompt

Deliverables:
- ace.py with reflection/curation logic
- playbook.py for storage management
- Lesson extraction and formatting
- Playbook pruning algorithms
- Unit tests for ACE components
- Example playbooks for testing
```

#### 21.5 Implement agent heartbeat system
**Priority**: HIGH
**Estimated Time**: 1-2 days

```python
# heartbeat.py - Agent health reporting

Requirements:
- Send periodic heartbeat messages to mcp_agent_mail
  - POST /heartbeats every 30 seconds
  - Include: agent_name, status, current_task, timestamp
  
- Include current status (idle, working, error)
  - Track state transitions
  - Report state changes immediately
  
- Include current task ID if working
  - Link to beads task
  - Include progress percentage
  
- Handle connection failures gracefully
  - Retry with exponential backoff
  - Continue working if MCP unavailable
  - Log connection issues

Deliverables:
- heartbeat.py with background thread
- Status tracking
- Connection retry logic
- Unit tests for heartbeat
```

#### 21.6 Create agent package structure
**Priority**: MEDIUM
**Estimated Time**: 1 day

```
agent/
├── __init__.py
├── agent_adapter.py      # Main entry point
├── llm_client.py         # Base LLM client
├── claude_client.py      # Claude implementation
├── gemini_client.py      # Gemini implementation
├── openai_client.py      # OpenAI implementation
├── phase_loop.py         # Main execution loop
├── ace.py                # Reflection/curation
├── playbook.py           # Playbook storage
├── heartbeat.py          # Health reporting
├── config.py             # Configuration management
├── utils.py              # Utility functions
├── requirements.txt      # Python dependencies
└── tests/
    ├── test_llm_client.py
    ├── test_phase_loop.py
    ├── test_ace.py
    └── test_heartbeat.py

Deliverables:
- Complete package structure
- requirements.txt with dependencies
- setup.py for installation
- README.md for agent development
```

#### 21.7 Integration and testing
**Priority**: HIGH
**Estimated Time**: 2-3 days

```
Requirements:
- End-to-end test with real beads and MCP
- Test all three LLM clients
- Test phase loop with sample tasks
- Test ACE reflection and learning
- Test heartbeat system
- Test error recovery
- Performance testing (task throughput)
- Load testing (multiple agents)

Deliverables:
- Integration test suite
- Performance benchmarks
- Load test results
- Bug fixes from testing
```

## ⚡ Phase 2: Make It Real-Time (HIGH - 1 week)

### Task 22: Implement Real-Time TUI Updates

#### 22.1 Add WebSocket support to MCP client
**Priority**: HIGH
**Estimated Time**: 2-3 days

```go
Requirements:
- Add WebSocket connection to MCP server
- Subscribe to agent status events
- Subscribe to message events
- Handle reconnection on disconnect
- Maintain connection health

Deliverables:
- WebSocket client in internal/mcp/websocket.go
- Event subscription system
- Reconnection logic
- Unit tests
```

#### 22.2 Implement event-driven TUI updates
**Priority**: HIGH
**Estimated Time**: 2-3 days

```go
Requirements:
- Replace polling ticker with event channels
- Update model on WebSocket events
- Maintain fallback polling for beads
- Add connection status indicator
- Optimize rendering

Deliverables:
- Event-driven update loop
- Channel-based architecture
- Connection status UI
- Performance improvements
```

## 🎮 Phase 3: Make It Interactive (MEDIUM - 1-2 weeks)

### Task 23: Implement Interactive TUI Features

#### 23.1 Add task interaction
**Priority**: MEDIUM
**Estimated Time**: 2-3 days

```go
Requirements:
- Navigate task list with arrow keys
- Press 'c' to claim selected task
- Press 'v' to view task details
- Press 'n' to create new task
- Show task detail modal

Deliverables:
- Task navigation system
- Task claiming logic
- Task detail view
- Task creation form
```

#### 23.2 Add agent controls
**Priority**: MEDIUM
**Estimated Time**: 2-3 days

```go
Requirements:
- Select agent with number keys
- Press 'p' to pause/resume agent
- Press 'k' to kill agent
- Press 'r' to restart agent
- Press 'l' to view agent logs

Deliverables:
- Agent selection system
- Agent control commands
- Agent log viewer
- Confirmation dialogs
```

#### 23.3 Add log filtering
**Priority**: MEDIUM
**Estimated Time**: 1-2 days

```go
Requirements:
- Press '/' to enter search mode
- Filter logs by agent name
- Filter logs by message type
- Export logs to file

Deliverables:
- Search input mode
- Filter logic
- Log export functionality
```

## 🏥 Phase 4: Make It Robust (MEDIUM - 1 week)

### Task 24: Implement Health Monitoring

#### 24.1 Add health check system
**Priority**: MEDIUM
**Estimated Time**: 2-3 days

```go
Requirements:
- Ping agents every 30 seconds
- Detect unresponsive agents
- Detect crashed agents
- Detect stuck agents
- Alert user to issues

Deliverables:
- Health check system
- Agent monitoring
- Alert system
```

#### 24.2 Add automatic recovery
**Priority**: MEDIUM
**Estimated Time**: 2-3 days

```go
Requirements:
- Restart crashed agents automatically
- Release leases from stuck agents
- Notify user of recovery actions
- Log all recovery actions

Deliverables:
- Auto-restart logic
- Lease cleanup
- Recovery notifications
```

## 📋 Implementation Checklist

### Before Starting
- [ ] Review gap analysis with team
- [ ] Prioritize tasks based on business needs
- [ ] Set up development environment
- [ ] Create feature branch

### Phase 1 (Agent Adapter)
- [ ] Task 21.1: Entry point
- [ ] Task 21.2: LLM clients
- [ ] Task 21.3: Phase loop
- [ ] Task 21.4: ACE system
- [ ] Task 21.5: Heartbeat
- [ ] Task 21.6: Package structure
- [ ] Task 21.7: Integration testing

### Phase 2 (Real-Time)
- [ ] Task 22.1: WebSocket support
- [ ] Task 22.2: Event-driven updates

### Phase 3 (Interactive)
- [ ] Task 23.1: Task interaction
- [ ] Task 23.2: Agent controls
- [ ] Task 23.3: Log filtering

### Phase 4 (Robust)
- [ ] Task 24.1: Health checks
- [ ] Task 24.2: Auto-recovery

## 🎯 Success Criteria

### Phase 1 Complete When:
- [ ] Agent adapter runs successfully
- [ ] Agents can complete simple tasks
- [ ] All three LLM clients work
- [ ] ACE system learns from tasks
- [ ] Heartbeats appear in TUI
- [ ] Integration tests pass

### Phase 2 Complete When:
- [ ] TUI updates in <1 second
- [ ] No polling delays
- [ ] WebSocket connection stable
- [ ] Connection status visible

### Phase 3 Complete When:
- [ ] Can claim tasks from TUI
- [ ] Can control agents from TUI
- [ ] Can search/filter logs
- [ ] All interactions feel smooth

### Phase 4 Complete When:
- [ ] Crashed agents restart automatically
- [ ] Stuck agents detected and recovered
- [ ] Health issues visible in TUI
- [ ] System runs reliably for hours

## 📊 Estimated Timeline

| Phase | Duration | Parallel Work Possible |
|-------|----------|------------------------|
| Phase 1 | 2-3 weeks | Yes (LLM clients, ACE) |
| Phase 2 | 1 week | No (depends on Phase 1) |
| Phase 3 | 1-2 weeks | Yes (all tasks independent) |
| Phase 4 | 1 week | Yes (health + recovery) |

**Total**: 5-7 weeks for complete implementation

**Minimum Viable**: Phase 1 only (2-3 weeks) for working system

## 🚀 Quick Start Guide

To begin implementation:

```bash
# 1. Create agent package
mkdir -p agent/tests
cd agent

# 2. Set up Python environment
python -m venv venv
source venv/bin/activate
pip install anthropic google-generativeai openai requests

# 3. Create initial files
touch agent_adapter.py llm_client.py phase_loop.py
touch requirements.txt setup.py README.md

# 4. Start with entry point
# Implement agent_adapter.py first
# Then LLM clients
# Then phase loop
# Then ACE
# Then heartbeat

# 5. Test as you go
pytest tests/
```

## 📝 Notes

- **Agent adapter is Python**: Spec assumes Python, but architecture supports any language
- **LLM APIs require keys**: Need valid API keys for testing
- **Beads and MCP required**: Need running instances for integration tests
- **Iterative development**: Build incrementally, test frequently
- **User feedback**: Get feedback after Phase 1 before continuing

---

**Ready to start?** Begin with Task 21.1 (agent_adapter.py entry point)
