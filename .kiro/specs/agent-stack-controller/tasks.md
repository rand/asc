# Implementation Plan

- [x] 1. Initialize Go project and set up core dependencies
  - Create Go module with `go mod init`
  - Add dependencies: cobra, viper, bubbletea, lipgloss, bubbles
  - Create project directory structure (cmd/, internal/)
  - Set up main.go entry point
  - _Requirements: 1.1, 2.1, 4.1_

- [x] 2. Implement configuration system
  - [x] 2.1 Create configuration data structures
    - Define Config, CoreConfig, ServicesConfig, AgentConfig structs
    - Add TOML struct tags for viper parsing
    - _Requirements: 11.1, 11.2_
  
  - [x] 2.2 Implement TOML configuration parser
    - Create config loader using viper
    - Implement validation for required fields
    - Add default values for optional fields
    - Handle file not found and parse errors
    - _Requirements: 4.2, 11.1_
  
  - [x] 2.3 Implement environment file handling
    - Create .env file parser
    - Load API keys into environment
    - Validate required keys are present
    - _Requirements: 1.5, 4.3_

- [x] 3. Implement dependency checker
  - [x] 3.1 Create checker interface and implementation
    - Define Checker interface with CheckBinary, CheckFile, CheckConfig, CheckEnv methods
    - Implement binary existence checks using exec.LookPath
    - Implement file existence and readability checks
    - _Requirements: 4.1, 4.2_
  
  - [x] 3.2 Implement check result formatting
    - Create CheckResult struct with status and message
    - Format results as styled table using lipgloss
    - Add color coding for pass/fail/warn statuses
    - _Requirements: 4.4_

- [x] 4. Implement process management system
  - [x] 4.1 Create process manager interface
    - Define ProcessManager interface with Start, Stop, StopAll, IsRunning methods
    - Create ProcessStatus type and constants
    - _Requirements: 2.2, 2.4, 3.2_
  
  - [x] 4.2 Implement process lifecycle management
    - Implement Start method using exec.Command
    - Set up process groups with SysProcAttr
    - Capture stdout/stderr to log files
    - _Requirements: 2.4, 2.5_
  
  - [x] 4.3 Implement PID tracking
    - Create PID file storage in ~/.asc/pids/
    - Save process metadata as JSON
    - Load PIDs on startup for status checks
    - _Requirements: 3.1, 3.2_
  
  - [x] 4.4 Implement graceful shutdown
    - Send SIGTERM to processes
    - Wait for graceful shutdown with timeout
    - Send SIGKILL if timeout exceeded
    - Clean up PID files
    - _Requirements: 3.2, 3.3_

- [x] 5. Implement beads client
  - [x] 5.1 Create beads client interface and data structures
    - Define BeadsClient interface with GetTasks, CreateTask, UpdateTask, DeleteTask methods
    - Create Task struct with ID, Title, Status, Phase fields
    - _Requirements: 5.1, 8.1_
  
  - [x] 5.2 Implement beads CLI integration
    - Execute bd commands using exec.Command
    - Parse JSON output from bd --json
    - Implement error handling for command failures
    - _Requirements: 5.1, 8.2_
  
  - [x] 5.3 Implement git refresh mechanism
    - Execute git pull on beads repository
    - Handle merge conflicts gracefully
    - Implement periodic refresh with configurable interval
    - _Requirements: 8.5_

- [x] 6. Implement MCP client
  - [x] 6.1 Create MCP client interface and data structures
    - Define MCPClient interface with GetMessages, SendMessage, GetAgentStatus methods
    - Create Message struct with Timestamp, Type, Source, Content fields
    - Create AgentStatus struct with Name, State, CurrentTask fields
    - _Requirements: 7.7, 9.1_
  
  - [x] 6.2 Implement HTTP client for MCP server
    - Create HTTP client using net/http
    - Implement GetMessages endpoint polling
    - Implement SendMessage endpoint
    - Add connection error handling and retries
    - _Requirements: 5.2, 9.6_
  
  - [x] 6.3 Implement agent status tracking
    - Poll MCP server for agent heartbeats
    - Map heartbeat data to AgentStatus
    - Detect offline agents based on last seen time
    - _Requirements: 7.7_

- [x] 7. Implement CLI command: asc check
  - [x] 7.1 Create check command structure
    - Set up cobra command for "check"
    - Wire up checker implementation
    - _Requirements: 4.1_
  
  - [x] 7.2 Implement check execution and output
    - Run all dependency checks
    - Format results with lipgloss table
    - Exit with appropriate status code
    - _Requirements: 4.4, 4.5, 4.6_

- [x] 8. Implement CLI command: asc services
  - [x] 8.1 Create services command with subcommands
    - Create cmd/services.go with cobra command structure
    - Add start, stop, and status subcommands
    - Wire up process manager for service lifecycle control
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [x] 9. Implement CLI command: asc test
  - [x] 9.1 Create test command and implement end-to-end test flow
    - Create cmd/test.go with cobra command structure
    - Load configuration and initialize beads and MCP clients
    - Create test beads task with title "asc test task"
    - Send test message to MCP server
    - Poll both beads and MCP for confirmation with 30s timeout
    - Clean up test artifacts (delete task and message)
    - Report success or failure with detailed error messages
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

- [x] 10. Implement CLI command: asc init
  - [x] 10.1 Create init command structure
    - Create cmd/init.go with cobra command structure
    - _Requirements: 1.1_
  
  - [x] 10.2 Implement interactive setup wizard
    - Create internal/tui/wizard.go with bubbletea application
    - Display welcome screen with project overview
    - Run dependency checks and display results in wizard
    - _Requirements: 1.1, 1.2_
  
  - [x] 10.3 Implement configuration prompts
    - Prompt for missing component installation with confirmation
    - Backup existing asc.toml and .env files with timestamps to ~/.asc_backup
    - Collect API keys with masked input using bubbles/textinput
    - Validate API key format before proceeding
    - _Requirements: 1.3, 1.4, 1.5_
  
  - [x] 10.4 Generate default configuration files
    - Create default asc.toml with sample agent definitions
    - Create .env file with collected API keys
    - Set file permissions to 0600 for .env
    - _Requirements: 1.6_
  
  - [x] 10.5 Run validation and report success
    - Execute asc test command to verify stack
    - Display success message or detailed error information
    - _Requirements: 1.7_

- [x] 11. Implement TUI application core
  - [x] 11.1 Create bubbletea model and initialization
    - Create internal/tui/model.go with Model struct
    - Add fields: config, beadsClient, mcpClient, procManager, agents, tasks, messages
    - Add UI state fields: width, height, lastRefresh
    - Implement Init method to set up initial state and start ticker
    - _Requirements: 2.6, 2.7_
  
  - [x] 11.2 Implement update loop and event handling
    - Create internal/tui/update.go with Update method
    - Handle tea.KeyMsg for keyboard events (q for quit, r for refresh, t for test)
    - Handle tea.WindowSizeMsg for terminal resize
    - Handle tickMsg for periodic refresh (every 2-5 seconds)
    - Route messages to appropriate handlers
    - _Requirements: 10.1, 10.2, 10.3_
  
  - [x] 11.3 Implement data refresh logic
    - Create refreshData method in model
    - Fetch agent statuses from MCP client using GetAllAgentStatuses
    - Fetch tasks from beads client with statuses ["open", "in_progress"]
    - Fetch messages from MCP client since last refresh
    - Update model state with new data
    - Handle errors gracefully without crashing TUI
    - _Requirements: 7.7, 8.5, 9.6_

- [x] 12. Implement TUI pane: Agent Status
  - [x] 12.1 Create agent status pane rendering
    - Create internal/tui/agents.go with renderAgentPane method
    - Iterate through agents from config
    - Display agent name and current status
    - Map agent state to icon: ● (idle), ⟳ (working), ! (error), ○ (offline)
    - Apply color styling: green (idle), blue (working), red (error), gray (offline)
    - Display current task ID if agent is working
    - Add border with title "Agent Status" using lipgloss
    - Apply consistent spacing and alignment
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 12.1, 12.3_

- [x] 13. Implement TUI pane: Beads Task Stream
  - [x] 13.1 Create task stream pane rendering
    - Create internal/tui/tasks.go with renderTaskPane method
    - Filter tasks by status (open, in_progress)
    - Display task ID, status icon, and title
    - Highlight in-progress tasks with distinct styling (bold or different color)
    - Implement scrolling for overflow content using viewport or manual truncation
    - Add border with title "Task Stream" using lipgloss
    - Apply consistent spacing and alignment
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 12.1, 12.3_

- [x] 14. Implement TUI pane: MCP Interaction Log
  - [x] 14.1 Create log pane rendering
    - Create internal/tui/logs.go with renderLogPane method
    - Display messages in chronological order
    - Format each message: [HH:MM:SS] [Type] [Source] → [Content]
    - Apply color coding: blue (lease), green (beads), red (error), default (message)
    - Auto-scroll to bottom on new messages
    - Limit display to last 100 messages
    - Add border with title "MCP Interaction Log" using lipgloss
    - Apply consistent spacing and alignment
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 12.1, 12.3_

- [x] 15. Implement TUI layout and composition
  - [x] 15.1 Create main View method and layout
    - Create internal/tui/view.go with View method
    - Calculate pane dimensions based on terminal size (width, height)
    - Left pane (Agent Status): 1/3 width, full height
    - Right top pane (Task Stream): 2/3 width, 1/2 height
    - Right bottom pane (MCP Log): 2/3 width, 1/2 height
    - Compose layout using lipgloss.JoinHorizontal and JoinVertical
    - _Requirements: 12.2_
  
  - [x] 15.2 Implement footer rendering
    - Create renderFooter method
    - Display keybindings: (q)uit | (r)efresh | (t)est
    - Add connection status indicators for beads and MCP
    - Style with lipgloss
    - _Requirements: 10.4_
  
  - [x] 15.3 Implement responsive layout
    - Handle terminal resize in Update method
    - Recalculate pane dimensions on tea.WindowSizeMsg
    - Adjust content to fit new dimensions
    - _Requirements: 12.5_

- [x] 16. Implement CLI command: asc up
  - [x] 16.1 Create up command and implement startup sequence
    - Create cmd/up.go with cobra command structure
    - Run silent dependency check using checker (exit on failure)
    - Load configuration from asc.toml
    - Load environment variables from .env
    - Initialize process manager with ~/.asc/pids and ~/.asc/logs
    - Start mcp_agent_mail service using process manager
    - _Requirements: 2.1, 2.2, 2.3_
  
  - [x] 16.2 Launch agent processes
    - Iterate through agents in config
    - Build environment variables for each agent: AGENT_NAME, AGENT_MODEL, AGENT_PHASES (comma-separated), MCP_MAIL_URL, BEADS_DB_PATH, API keys
    - Start each agent using process manager
    - Handle startup failures gracefully with error messages
    - _Requirements: 2.4, 2.5, 11.3, 11.4, 11.5, 11.6_
  
  - [x] 16.3 Initialize and run TUI
    - Clear terminal screen
    - Initialize beads and MCP clients
    - Create bubbletea Model with config and clients
    - Start TUI event loop with tea.NewProgram
    - Handle TUI exit and cleanup
    - _Requirements: 2.6, 2.7_

- [x] 17. Implement CLI command: asc down
  - [x] 17.1 Create down command and implement shutdown sequence
    - Create cmd/down.go with cobra command structure
    - Initialize process manager with ~/.asc/pids and ~/.asc/logs
    - List all managed processes
    - Stop all agent processes using process manager
    - Stop mcp_agent_mail service
    - Clean up PID files
    - Print confirmation message: "Agent stack is offline"
    - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 18. Implement error handling and logging
  - [x] 18.1 Create centralized logging system
    - Create internal/logger package
    - Set up log file at ~/.asc/logs/asc.log
    - Implement log levels: DEBUG, INFO, WARN, ERROR
    - Implement log rotation (max size 10MB, keep 5 files)
    - Add structured logging with timestamps and context
    - _Requirements: All commands_
  
  - [x] 18.2 Implement user-friendly error messages
    - Create error formatting utilities
    - Format errors with actionable solutions
    - Display errors appropriately in TUI (in log pane) vs CLI (stderr)
    - Add error recovery suggestions
    - _Requirements: All commands_

- [x] 19. Write documentation
  - [x] 19.1 Create README.md
    - Add project overview and architecture diagram
    - Document installation instructions (go install, binary download)
    - Provide usage examples for all commands (init, up, down, check, test, services)
    - Include asc.toml configuration guide with examples
    - Add troubleshooting section for common issues
    - Document environment variables and API key setup
    - _Requirements: All_
  
  - [x] 19.2 Add inline code documentation
    - Document all exported functions and types with godoc comments
    - Add package-level documentation for each package
    - Include usage examples in godoc comments
    - _Requirements: All_

- [x] 20. Create build and distribution setup
  - [x] 20.1 Set up build scripts and installation guide
    - Create Makefile with build, test, install, and clean targets
    - Add multi-platform build targets (linux/amd64, darwin/amd64, darwin/arm64)
    - Create installation guide in README
    - Document go install method
    - _Requirements: All, 1.1_
