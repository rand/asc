# Requirements Document

## Introduction

The Agent Stack Controller (asc) is a command-line orchestration tool that manages a local colony of AI coding agents. It provides developers with a mission control interface for starting, monitoring, and coordinating headless background agents that work collaboratively on software development tasks. The system integrates with beads (task management), mcp_agent_mail (agent communication), and supports multiple LLM providers (Claude, Gemini, Codex) in a matrix architecture where any model can fulfill any role.

## Glossary

- **asc**: The Agent Stack Controller CLI tool and TUI dashboard
- **Agent**: A headless background process that executes development tasks using an LLM
- **beads**: A Git-backed task database that serves as the single source of truth for tasks
- **mcp_agent_mail**: An asynchronous communication server for agent coordination and file leasing
- **TUI**: Terminal User Interface built with Go and bubbletea
- **Agent Role**: The functional responsibility of an agent (e.g., planner, tester, refactor)
- **LLM Model**: The underlying language model provider (e.g., Claude, Gemini, Codex)
- **Phase**: A workflow stage in the Hephaestus pattern (e.g., Plan, Implement, Test)
- **Lease**: A file lock mechanism to prevent concurrent agent modifications
- **Playbook**: A collection of learned lessons stored by agents using the ACE pattern

## Requirements

### Requirement 1

**User Story:** As a developer, I want to initialize the asc tool with a single command, so that all dependencies and configurations are set up automatically

#### Acceptance Criteria

1. WHEN the developer executes "asc init", THE asc SHALL launch an interactive setup wizard using bubbletea
2. WHEN the setup wizard starts, THE asc SHALL execute dependency checks to identify existing tools and missing components
3. WHEN missing dependencies are detected, THE asc SHALL offer to install the missing components with user confirmation
4. WHEN existing configuration files are found, THE asc SHALL create timestamped backups in the ~/.asc_backup directory
5. WHEN API key configuration is required, THE asc SHALL prompt for Claude, OpenAI, and Google API keys and store them securely in a project-local .env file
6. WHEN configuration is complete, THE asc SHALL create a default asc.toml configuration file with agent definitions
7. WHEN all setup steps succeed, THE asc SHALL execute "asc test" to verify the stack and report success status

### Requirement 2

**User Story:** As a developer, I want to start all agents and services with a single command, so that I can quickly begin orchestrating my agent colony

#### Acceptance Criteria

1. WHEN the developer executes "asc up", THE asc SHALL perform silent dependency checks and exit with a clear error message if checks fail
2. WHEN dependency checks pass, THE asc SHALL parse the asc.toml file to identify all defined agents
3. WHEN agents are identified, THE asc SHALL start the mcp_agent_mail server before launching agents
4. WHEN launching each agent, THE asc SHALL execute the agent's defined command as a durable background process
5. WHEN launching each agent, THE asc SHALL pass required environment variables including API keys, MCP_MAIL_URL, BEADS_DB_PATH, AGENT_NAME, AGENT_MODEL, and AGENT_PHASES
6. WHEN all processes are launched successfully, THE asc SHALL clear the terminal and launch the bubbletea TUI dashboard
7. WHEN the TUI launches, THE asc SHALL connect to the beads database and mcp_agent_mail server to populate the dashboard state

### Requirement 3

**User Story:** As a developer, I want to gracefully shut down all agents and services, so that I can cleanly stop the agent colony without orphaned processes

#### Acceptance Criteria

1. WHEN the developer executes "asc down", THE asc SHALL read the asc.toml file to identify all managed agent processes
2. WHEN agent processes are identified, THE asc SHALL send SIGTERM signals to all child agent processes
3. WHEN agent processes are terminated, THE asc SHALL stop the mcp_agent_mail service
4. WHEN all services are stopped, THE asc SHALL print a confirmation message indicating the agent stack is offline

### Requirement 4

**User Story:** As a developer, I want to verify my environment setup, so that I can identify missing dependencies before attempting to start the agent stack

#### Acceptance Criteria

1. WHEN the developer executes "asc check", THE asc SHALL verify that git, python3, uv, bd, and docker binaries exist in the system PATH
2. WHEN binary verification completes, THE asc SHALL verify that the asc.toml configuration file exists and contains valid TOML syntax
3. WHEN configuration verification completes, THE asc SHALL verify that the .env file exists and contains all required API keys
4. WHEN all checks complete, THE asc SHALL print a list of checks with pass or fail status indicators
5. WHEN all checks pass, THE asc SHALL exit with status code 0
6. WHEN any check fails, THE asc SHALL exit with status code 1

### Requirement 5

**User Story:** As a developer, I want to run an end-to-end test of the agent stack, so that I can verify all components are communicating correctly

#### Acceptance Criteria

1. WHEN the developer executes "asc test", THE asc SHALL create a test beads task with the title "asc test task"
2. WHEN the test task is created, THE asc SHALL send a test message to the mcp_agent_mail server
3. WHEN the test message is sent, THE asc SHALL poll both beads and mcp_agent_mail to confirm message receipt within a timeout period
4. WHEN message receipt is confirmed, THE asc SHALL delete the test task and test message to clean up
5. WHEN all test steps succeed, THE asc SHALL report "Stack is healthy" with exit code 0
6. WHEN any test step fails, THE asc SHALL report which step failed and exit with a non-zero code

### Requirement 6

**User Story:** As a developer, I want to manage long-running services independently, so that I can control the mcp_agent_mail server without affecting agents

#### Acceptance Criteria

1. WHEN the developer executes "asc services start", THE asc SHALL start the mcp_agent_mail server as a background process
2. WHEN the developer executes "asc services stop", THE asc SHALL terminate the mcp_agent_mail server process
3. WHEN the developer executes "asc services status", THE asc SHALL report whether the mcp_agent_mail server is running
4. WHEN managing services, THE asc SHALL use process ID files or a process manager to track service state

### Requirement 7

**User Story:** As a developer, I want to view real-time agent status in a dashboard, so that I can monitor which agents are idle, working, or experiencing errors

#### Acceptance Criteria

1. WHEN the TUI dashboard is active, THE asc SHALL display an Agent Status pane showing all agents defined in asc.toml
2. WHEN displaying agent status, THE asc SHALL show each agent's name, current state icon, and status description
3. WHEN an agent is idle, THE asc SHALL display a green filled circle icon with "Idle" status
4. WHEN an agent is working, THE asc SHALL display a blue rotating icon with "Working on #[task_id]" status
5. WHEN an agent has an error, THE asc SHALL display a red exclamation icon with "Error" status
6. WHEN an agent is offline, THE asc SHALL display a gray empty circle icon with "Offline" status
7. WHEN agent status changes, THE asc SHALL update the display by polling mcp_agent_mail for heartbeat messages

### Requirement 8

**User Story:** As a developer, I want to view the current task list in the dashboard, so that I can see what work is available and in progress

#### Acceptance Criteria

1. WHEN the TUI dashboard is active, THE asc SHALL display a Beads Task Stream pane showing tasks from the beads database
2. WHEN displaying tasks, THE asc SHALL show only tasks with status "open" or "in_progress"
3. WHEN displaying each task, THE asc SHALL show the task ID, status icon, and task title
4. WHEN a task is in progress, THE asc SHALL highlight the task with distinct styling
5. WHEN task data changes, THE asc SHALL refresh the display by periodically executing git pull on the beads repository and re-reading the database file

### Requirement 9

**User Story:** As a developer, I want to view agent communication logs in real-time, so that I can understand how agents are coordinating and identify issues

#### Acceptance Criteria

1. WHEN the TUI dashboard is active, THE asc SHALL display an MCP Interaction Log pane showing messages from mcp_agent_mail
2. WHEN displaying messages, THE asc SHALL show the timestamp, message type, source agent, and message content
3. WHEN displaying lease messages, THE asc SHALL use blue color styling
4. WHEN displaying beads task messages, THE asc SHALL use green color styling
5. WHEN displaying error messages, THE asc SHALL use red color styling
6. WHEN new messages arrive, THE asc SHALL append them to the log by polling the mcp_agent_mail server API

### Requirement 10

**User Story:** As a developer, I want to interact with the dashboard using keyboard commands, so that I can control the system without leaving the TUI

#### Acceptance Criteria

1. WHEN the TUI dashboard is active and the developer presses "q", THE asc SHALL execute the shutdown sequence equivalent to "asc down"
2. WHEN the TUI dashboard is active and the developer presses "r", THE asc SHALL force-refresh all dashboard panes by re-fetching data from beads and mcp_agent_mail
3. WHEN the TUI dashboard is active and the developer presses "t", THE asc SHALL execute the "asc test" command and display results in the log pane
4. WHEN the TUI dashboard is active, THE asc SHALL display available keybindings in a footer pane

### Requirement 11

**User Story:** As a developer, I want to configure agents with different models and roles, so that I can create a heterogeneous agent colony optimized for different tasks

#### Acceptance Criteria

1. WHEN the asc.toml file defines an agent section, THE asc SHALL parse the agent name from the section key
2. WHEN parsing an agent definition, THE asc SHALL extract the command, model, and phases fields
3. WHEN launching an agent, THE asc SHALL pass the agent name as the AGENT_NAME environment variable
4. WHEN launching an agent, THE asc SHALL pass the model value as the AGENT_MODEL environment variable
5. WHEN launching an agent, THE asc SHALL pass the phases array as a comma-separated AGENT_PHASES environment variable
6. WHEN multiple agents are defined with the same phase, THE asc SHALL launch all agents and allow them to compete for tasks in that phase

### Requirement 12

**User Story:** As a developer, I want the TUI to have a clean, information-dense aesthetic, so that I can quickly understand system state at a glance

#### Acceptance Criteria

1. WHEN rendering the TUI, THE asc SHALL use the lipgloss library for all styling including borders, colors, and layout
2. WHEN rendering the TUI, THE asc SHALL divide the display into three main panes with a footer for keybindings
3. WHEN rendering pane borders, THE asc SHALL use clean line styles consistent with the Charm.land aesthetic
4. WHEN rendering status information, THE asc SHALL use color coding to convey state without requiring text reading
5. WHEN the terminal is resized, THE asc SHALL adjust pane dimensions to maintain the layout proportions
