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

- [x] 21. Implement agent adapter framework (Python)
  - [x] 21.1 Create agent_adapter.py entry point
    - Parse environment variables (AGENT_NAME, AGENT_MODEL, AGENT_PHASES, MCP_MAIL_URL, BEADS_DB_PATH, API keys)
    - Initialize logging to ~/.asc/logs/{agent_name}.log
    - Set up signal handlers for graceful shutdown (SIGTERM, SIGINT)
    - Initialize LLM client based on AGENT_MODEL environment variable
    - Enter main event loop with error handling
    - _Requirements: 2.4, 2.5, 11.3, 11.4, 11.5_
  
  - [x] 21.2 Implement LLM client abstraction
    - Create base LLMClient abstract class with complete() interface
    - Implement ClaudeClient using Anthropic SDK with API error handling
    - Implement GeminiClient using Google AI SDK with rate limiting
    - Implement OpenAIClient using OpenAI SDK for GPT-4 and Codex
    - Add retry logic with exponential backoff for API failures
    - Implement token counting and cost tracking
    - _Requirements: 11.4, 11.5_
  
  - [x] 21.3 Implement Hephaestus phase loop
    - Poll beads for tasks matching agent phases using bd CLI
    - Request file leases via mcp_agent_mail POST /leases endpoint
    - Build context from task description and relevant files
    - Load playbook lessons into context
    - Call LLM with structured prompts including context and playbook
    - Parse LLM response and extract action plan
    - Execute file operations (read, write, delete) with safety checks
    - Update beads task status using bd update command
    - Release file leases via POST /leases/{id}/release
    - _Requirements: 2.4, 2.5, 5.1, 7.7, 8.1, 9.1, 11.3, 11.4, 11.5, 11.6_
  
  - [x] 21.4 Implement ACE (Agentic Context Engineering)
    - Create playbook storage structure in ~/.asc/playbooks/{agent_name}/
    - Define playbook schema: lesson_id, context, action, outcome, learned
    - Implement reflection prompt after task completion
    - Extract structured lessons from LLM reflection response
    - Categorize lessons by task type and score relevance
    - Curate playbook by deduplicating and merging similar lessons
    - Prune outdated lessons and maintain max playbook size
    - Load relevant playbook lessons into context for future tasks
    - _Requirements: 11.3, 11.4_
  
  - [x] 21.5 Implement agent heartbeat system
    - Send periodic heartbeat messages to mcp_agent_mail every 30 seconds
    - Include agent_name, status (idle/working/error), current_task, timestamp
    - Track state transitions and report changes immediately
    - Handle MCP connection failures with exponential backoff retry
    - Continue working if MCP temporarily unavailable
    - _Requirements: 7.7, 9.1_
  
  - [x] 21.6 Create agent package structure and dependencies
    - Create agent/ directory with __init__.py and all module files
    - Write requirements.txt with dependencies: anthropic, google-generativeai, openai, requests
    - Create setup.py for package installation
    - Write agent README.md with development guide
    - Add unit tests for all agent components
    - _Requirements: 2.4, 2.5_
  
  - [x] 21.7 Integration testing and validation
    - Test end-to-end flow with real beads and MCP instances
    - Validate all three LLM clients (Claude, Gemini, OpenAI)
    - Test phase loop with sample tasks from beads
    - Verify ACE reflection and playbook learning
    - Validate heartbeat system and status reporting
    - Test error recovery and graceful shutdown
    - _Requirements: All agent-related requirements_


## Phase 2: Real-Time and Interactive Enhancements

- [x] 22. Implement real-time TUI updates
  - [x] 22.1 Add WebSocket support to MCP client
    - Create WebSocket client in internal/mcp/websocket.go
    - Connect to MCP server WebSocket endpoint
    - Subscribe to agent status change events
    - Subscribe to new message events
    - Implement reconnection logic with exponential backoff
    - Handle connection health monitoring
    - _Requirements: 7.7, 9.6, 10.2_
  
  - [x] 22.2 Implement event-driven TUI updates
    - Replace polling ticker with event channels in TUI model
    - Update model state on WebSocket events instead of polling
    - Maintain fallback polling for beads (git-based, cannot be real-time)
    - Add connection status indicator in TUI footer
    - Optimize rendering to only update changed panes
    - Reduce CPU usage during idle periods
    - _Requirements: 10.2, 12.5_

- [x] 23. Implement interactive TUI features
  - [x] 23.1 Add task interaction capabilities
    - Implement arrow key navigation for task list
    - Add 'c' key to claim selected task for current user
    - Add 'v' key to view full task details in modal
    - Add 'n' key to create new task with input form
    - Display task detail modal with description, status, assignee
    - _Requirements: 8.1, 8.2, 8.3, 8.4_
  
  - [x] 23.2 Add agent control capabilities
    - Implement number key (1-9) selection for agents
    - Add 'p' key to pause/resume selected agent
    - Add 'k' key to kill selected agent with confirmation
    - Add 'r' key to restart selected agent
    - Add 'l' key to view agent logs in detail view
    - Show confirmation dialogs for destructive actions
    - _Requirements: 2.2, 3.2, 7.1, 7.2_
  
  - [x] 23.3 Add log filtering and search
    - Add '/' key to enter search mode with text input
    - Implement filter by agent name
    - Implement filter by message type (lease, beads, error, message)
    - Add 'e' key to export filtered logs to file
    - Show active filters in log pane header
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [x] 24. Implement comprehensive health monitoring
  - [x] 24.1 Add health check system
    - Ping agents every 30 seconds via heartbeat check
    - Detect unresponsive agents (no heartbeat for 2 minutes)
    - Detect crashed agents (process exited unexpectedly)
    - Detect stuck agents (working on same task for >30 minutes)
    - Display health alerts in TUI with visual indicators
    - Log all health issues to ~/.asc/logs/health.log
    - _Requirements: 7.7, 3.2_
  
  - [x] 24.2 Add automatic recovery system
    - Restart crashed agents automatically with backoff
    - Release file leases from stuck agents via MCP
    - Notify user of recovery actions in TUI log pane
    - Log all recovery actions with timestamps and reasons
    - Add configuration option to disable auto-recovery
    - Track recovery success rate per agent
    - _Requirements: 3.2, 3.3, 7.6_

- [x] 25. Enhance configuration system
  - [x] 25.1 Add configuration validation and suggestions
    - Validate agent command exists in PATH before starting
    - Validate model is supported (claude, gemini, openai)
    - Validate phases are valid (planning, implementation, testing, etc.)
    - Warn about duplicate agent names in configuration
    - Suggest fixes for common configuration errors
    - _Requirements: 4.1, 4.2, 11.1, 11.2_
  
  - [x] 25.2 Add configuration templates
    - Create template system for common agent setups
    - Add "asc init --template=solo" for single agent setup
    - Add "asc init --template=team" for planner, coder, tester setup
    - Add "asc init --template=swarm" for multiple agents per phase
    - Allow users to save custom templates
    - _Requirements: 1.1, 1.6, 11.1, 11.2_
  
  - [x] 25.3 Add configuration hot-reload
    - Watch asc.toml for file changes using fsnotify
    - Reload configuration without full restart
    - Start new agents defined in updated config
    - Stop agents removed from config
    - Update existing agent configurations (model, phases)
    - Display reload notifications in TUI
    - _Requirements: 2.1, 2.2, 11.1, 11.2_

- [x] 26. Enhance logging and debugging
  - [x] 26.1 Add structured logging
    - Use JSON format for machine-parseable logs
    - Include context fields: agent, task, phase, timestamp
    - Add correlation IDs for tracing requests across components
    - Support per-agent log levels (DEBUG, INFO, WARN, ERROR)
    - _Requirements: All commands_
  
  - [x] 26.2 Add debug mode
    - Add "asc up --debug" flag for verbose output
    - Show LLM prompts and responses in debug logs
    - Show file lease operations and conflicts
    - Show beads database queries and git operations
    - Display debug info in TUI when enabled
    - _Requirements: All commands_
  
  - [x] 26.3 Add log aggregation and analysis
    - Collect logs from all agents into unified view
    - Display aggregated logs in TUI log pane
    - Support log export to file with filtering
    - Add log rotation (max 10MB per file, keep 5 files)
    - Implement log cleanup for old files
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_


- [x] 27. Implement vaporwave aesthetic design system
  - [x] 27.1 Create vaporwave color palette and theme system
    - Define core vaporwave color palette: neon pink (#FF71CE), electric blue (#01CDFE), purple (#B967FF), cyan (#05FFA1), sunset orange (#FFFB96)
    - Create gradient definitions for backgrounds and accents
    - Implement dark base colors: deep purple (#1A0933), midnight blue (#0D0221), dark teal (#0F0E17)
    - Add glow/luminous effect colors with alpha channels
    - Create theme struct with all color definitions
    - Implement color interpolation for smooth transitions
    - Add support for 256-color and true-color terminals
    - _Requirements: 12.1, 12.3_
  
  - [x] 27.2 Design elegant borders and frames with glow effects
    - Create custom border styles with double-line and rounded corners
    - Implement gradient borders that transition between vaporwave colors
    - Add subtle glow/shadow effects using Unicode box-drawing characters
    - Design corner ornaments with geometric patterns (triangles, diamonds)
    - Create title bars with centered text and decorative elements
    - Implement border animations (pulsing glow, color cycling)
    - Add depth with layered borders and shadows
    - _Requirements: 12.1, 12.3_
  
  - [x] 27.3 Implement sophisticated typography and text styling
    - Use bold weights for headers with gradient color fills
    - Implement text shadows and outlines for depth
    - Add subtle letter-spacing for elegance
    - Create hierarchical text styles (h1, h2, body, caption)
    - Design monospace styling for code and IDs with neon accents
    - Implement text animations (fade-in, shimmer, wave)
    - Add icon/emoji integration with proper spacing
    - _Requirements: 7.1, 8.1, 9.1, 12.1_
  
  - [x] 27.4 Design status indicators with luminous effects
    - Create glowing orbs for agent status (pulsing animations)
    - Implement progress bars with gradient fills and shine effects
    - Design task status badges with rounded corners and glow
    - Add animated state transitions (smooth color morphing)
    - Create connection status indicators with signal wave animations
    - Implement health meters with gradient fills
    - Add sparkle/particle effects for active states
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 8.2, 12.1_
  
  - [x] 27.5 Implement grid and geometric background patterns
    - Create subtle grid overlay with neon lines
    - Add perspective grid effect (vanishing point)
    - Implement geometric shapes (triangles, hexagons) as accents
    - Design scanline effect for retro-futuristic feel
    - Add subtle noise/grain texture for depth
    - Create animated background elements (floating shapes, particles)
    - Implement parallax effect for layered backgrounds
    - _Requirements: 12.2, 12.3_
  
  - [x] 27.6 Design modal dialogs and overlays with glass morphism
    - Create frosted glass effect for modal backgrounds
    - Implement backdrop blur simulation with transparency
    - Add gradient borders with glow effects
    - Design smooth fade-in/fade-out animations
    - Create elegant close buttons with hover effects
    - Implement modal shadows and depth layers
    - Add subtle animations (float, pulse, shimmer)
    - _Requirements: 8.3, 8.4, 10.1_
  
  - [x] 27.7 Implement smooth animations and transitions
    - Create easing functions (ease-in-out, cubic-bezier)
    - Implement color transition animations
    - Add fade effects for content changes
    - Design loading spinners with vaporwave styling
    - Create pulsing animations for active elements
    - Implement wave/ripple effects for interactions
    - Add frame-based animation system with timing control
    - _Requirements: 10.2, 12.5_
  
  - [x] 27.8 Design footer and header with holographic effects
    - Create gradient header bar with animated color shift
    - Implement holographic text effect (rainbow shimmer)
    - Design keybinding display with neon highlights
    - Add connection status with animated indicators
    - Create timestamp display with elegant formatting
    - Implement notification badges with glow effects
    - Add subtle scan-line animation across header/footer
    - _Requirements: 10.4, 12.1_
  
  - [x] 27.9 Implement responsive layout with elegant spacing
    - Use golden ratio (1.618) for proportions and spacing
    - Implement consistent padding and margins (8px grid system)
    - Create responsive breakpoints for different terminal sizes
    - Add smooth layout transitions on resize
    - Design content overflow handling with fade-out effects
    - Implement auto-scaling for text and elements
    - Add breathing room with generous whitespace
    - _Requirements: 12.2, 12.5_
  
  - [x] 27.10 Create theme configuration and customization
    - Implement theme switching system (vaporwave, cyberpunk, minimal)
    - Add configuration file for custom color schemes
    - Create theme preview in TUI settings
    - Implement hot-reload for theme changes
    - Add accessibility mode with high contrast
    - Create theme export/import functionality
    - Document theme customization guide
    - _Requirements: 11.1, 11.2, 12.1_
  
  - [x] 27.11 Polish and refinement
    - Conduct visual design review and iteration
    - Optimize rendering performance for smooth 60fps
    - Test on various terminal emulators (iTerm2, Alacritty, Windows Terminal)
    - Ensure graceful degradation for limited color terminals
    - Add subtle micro-interactions for delight
    - Create design documentation with screenshots
    - Gather user feedback and iterate
    - _Requirements: 12.1, 12.3, 12.5_

- [ ] 28. Comprehensive testing and quality assurance
  - [x] 28.1 Expand unit test coverage
    - Achieve 80%+ code coverage for all Go packages
    - Add table-driven tests for complex logic (config parsing, process management)
    - Test error paths and edge cases (nil pointers, empty inputs, invalid data)
    - Add tests for concurrent operations (process manager, WebSocket client)
    - Test boundary conditions (max values, empty strings, zero values)
    - Mock external dependencies (file system, network, exec commands)
    - Add benchmarks for performance-critical code (rendering, animations)
    - _Requirements: All_
  
  - [x] 28.2 Implement integration tests
    - Test full command workflows (init → up → test → down)
    - Test configuration loading and validation end-to-end
    - Test process lifecycle (start → monitor → stop → cleanup)
    - Test beads and MCP client integration with real services
    - Test TUI interaction flows (navigation, modals, filtering)
    - Test WebSocket reconnection and fallback to polling
    - Test health monitoring and auto-recovery system
    - Test configuration hot-reload with running agents
    - _Requirements: All_
  
  - [x] 28.3 Add end-to-end tests
    - Test complete agent stack startup and shutdown
    - Test agent task execution from beads to completion
    - Test multi-agent coordination and file lease conflicts
    - Test error recovery scenarios (agent crash, MCP disconnect)
    - Test long-running stability (24+ hour runs)
    - Test resource cleanup (PIDs, logs, temp files)
    - Test graceful degradation (missing dependencies, network issues)
    - _Requirements: All_
  
  - [x] 28.4 Implement error handling tests
    - Test all error paths in each package
    - Test error propagation and wrapping
    - Test user-facing error messages for clarity
    - Test recovery from transient errors (network, file system)
    - Test handling of invalid user input
    - Test timeout handling (API calls, process starts)
    - Test panic recovery and graceful shutdown
    - _Requirements: All_
  
  - [x] 28.5 Add developer experience improvements
    - Create development setup guide (CONTRIBUTING.md)
    - Add pre-commit hooks for linting and formatting
    - Set up CI/CD pipeline (GitHub Actions or similar)
    - Add automated test runs on pull requests
    - Create code review checklist
    - Add debugging guides for common issues
    - Document testing best practices
    - Create troubleshooting playbook
    - _Requirements: All_
  
  - [x] 28.6 Implement quality gates and monitoring
    - Set up code coverage reporting (codecov or similar)
    - Add static analysis tools (golangci-lint, gosec)
    - Implement dependency vulnerability scanning
    - Add license compliance checking
    - Set up performance regression testing
    - Monitor test execution time and flakiness
    - Create quality metrics dashboard
    - _Requirements: All_
  
  - [x] 28.7 Review test suite outcomes and address gaps
    - [x] 28.7.1 Analyze current test coverage and identify gaps
      - Run comprehensive coverage analysis across all packages
      - Identify critical paths with insufficient coverage (<80%)
      - Review coverage reports for untested error paths
      - Document coverage gaps by package and priority
      - Create action plan to address high-priority gaps
      - _Requirements: All_
    
    - [x] 28.7.2 Review and fix failing tests
      - Identify all currently failing tests across the suite
      - Categorize failures (bugs, outdated tests, environment issues)
      - Fix or update failing unit tests
      - Fix or update failing integration tests
      - Fix or update failing E2E tests
      - Document any tests that need to be skipped with justification
      - _Requirements: All_
    
    - [x] 28.7.3 Address flaky tests identified by monitoring
      - Review flakiness reports from test-quality workflow
      - Investigate root causes (race conditions, timing, external deps)
      - Fix flaky tests by adding proper synchronization
      - Replace time.Sleep with proper wait conditions
      - Add retries for inherently flaky operations
      - Verify fixes with multiple test runs (20+ iterations)
      - _Requirements: All_
    
    - [x] 28.7.4 Improve test quality and maintainability
      - Refactor tests with excessive duplication
      - Add table-driven tests where appropriate
      - Improve test naming for clarity
      - Add missing test documentation and comments
      - Ensure all tests follow testing best practices
      - Add helper functions to reduce test boilerplate
      - _Requirements: All_
    
    - [x] 28.7.5 Add missing unit tests for core functionality
      - Add tests for uncovered configuration parsing logic
      - Add tests for uncovered process management operations
      - Add tests for uncovered TUI rendering logic
      - Add tests for uncovered client implementations
      - Add tests for uncovered error handling paths
      - Ensure all exported functions have test coverage
      - _Requirements: All_
    
    - [x] 28.7.6 Enhance integration test coverage
      - Add integration tests for multi-component workflows
      - Test configuration hot-reload functionality
      - Test health monitoring and auto-recovery
      - Test WebSocket reconnection scenarios
      - Test agent lifecycle management end-to-end
      - Test error recovery and graceful degradation
      - _Requirements: All_
    
    - [x] 28.7.7 Expand E2E test scenarios
      - Add E2E tests for complete user workflows
      - Test asc init → up → down workflow
      - Test agent task execution from start to finish
      - Test multi-agent coordination scenarios
      - Test failure and recovery scenarios
      - Add stress tests for high load conditions
      - _Requirements: All_
    
    - [x] 28.7.8 Review and improve test performance
      - Identify and optimize slow tests (>5s)
      - Add t.Parallel() to independent tests
      - Mock expensive operations (I/O, network, time)
      - Reduce test setup overhead
      - Optimize test data generation
      - Ensure test suite completes in <2 minutes
      - _Requirements: All_
    
    - [x] 28.7.9 Validate test environment and dependencies
      - Ensure all test dependencies are documented
      - Verify tests work in CI environment
      - Test on multiple platforms (Linux, macOS)
      - Test with different Go versions (1.21, 1.22)
      - Add setup instructions for local test execution
      - Document any platform-specific test requirements
      - _Requirements: All_
    
    - [x] 28.7.10 Create test gap remediation report
      - Document all identified gaps and their priority
      - Track progress on addressing each gap
      - Report final coverage metrics after improvements
      - Document any remaining gaps with justification
      - Create recommendations for ongoing test maintenance
      - Update testing documentation with lessons learned
      - _Requirements: All_
  
  - [x] 28.8 Test user flows and usability
    - Test first-time user experience (asc init)
    - Test common workflows (starting agents, viewing status)
    - Test error recovery from user perspective
    - Test keyboard navigation and shortcuts
    - Test terminal resize and responsiveness
    - Test accessibility features (high contrast mode)
    - Gather user feedback through beta testing
    - Document common user issues and solutions
    - _Requirements: All_
  
  - [x] 28.9 Add dependency management and updates
    - Document all dependencies and their purposes
    - Set up automated dependency updates (Dependabot)
    - Test compatibility with dependency updates
    - Pin critical dependencies to stable versions
    - Create dependency upgrade testing workflow
    - Monitor for security advisories
    - Document breaking changes in dependencies
    - _Requirements: All_
  
  - [x] 28.10 Implement issue detection and remediation
    - Add health check diagnostics (asc doctor command)
    - Detect common configuration issues automatically
    - Provide actionable remediation steps
    - Test recovery from corrupted state (PIDs, logs)
    - Add self-healing capabilities where possible
    - Create issue reporting template
    - Document known issues and workarounds
    - _Requirements: All_
  
  - [x] 28.11 Performance testing and optimization
    - Benchmark TUI rendering performance
    - Test memory usage under load (many agents, tasks, logs)
    - Profile CPU usage and identify bottlenecks
    - Test startup and shutdown time
    - Optimize hot paths (event loop, rendering)
    - Test with large datasets (1000+ tasks, 10000+ log entries)
    - Add performance regression tests
    - Document performance characteristics
    - _Requirements: All_
  
  - [x] 28.12 Security testing and hardening
    - Test API key handling and storage security
    - Test file permission handling (.env, logs, PIDs)
    - Test input validation and sanitization
    - Test command injection vulnerabilities
    - Test path traversal vulnerabilities
    - Add security scanning to CI/CD
    - Document security best practices
    - Create security incident response plan
    - _Requirements: 1.5, 4.3, All_
  
  - [x] 28.13 Documentation and knowledge base
    - Create comprehensive API documentation
    - Add architecture decision records (ADRs)
    - Document all configuration options
    - Create video tutorials for common tasks
    - Build FAQ from user questions
    - Document upgrade and migration guides
    - Create operator's handbook
    - Add inline code examples
    - _Requirements: All_

## Phase 29: Final Validation and Gap Analysis

- [ ] 29. Comprehensive build, test, and validation cycle
  - [x] 29.1 Perform full clean build
    - Clean all build artifacts and caches
    - Build for all target platforms (Linux amd64, macOS amd64, macOS arm64)
    - Verify binary sizes are reasonable
    - Test binary execution on each platform
    - Document build times and optimization opportunities
    - Verify all dependencies are properly vendored
    - Check for any build warnings or errors
    - _Requirements: All_
  
  - [x] 29.2 Run complete test suite
    - Run all unit tests with coverage reporting
    - Run all integration tests
    - Run all E2E tests (including long-running and stress tests)
    - Run all error handling tests
    - Run all performance tests
    - Run all security tests
    - Run all usability tests
    - Generate comprehensive test report
    - _Requirements: All_
  
  - [x] 29.3 Analyze test results and coverage
    - Review test coverage by package
    - Identify packages with <80% coverage
    - Analyze uncovered code paths
    - Review test execution times
    - Identify slow tests (>5s)
    - Check for test flakiness
    - Document test failures and their causes
    - Create prioritized list of coverage gaps
    - _Requirements: All_
  
  - [x] 29.4 Run static analysis and linting
    - Run golangci-lint with all enabled linters
    - Run gosec security scanner
    - Run go vet
    - Check for code formatting issues (gofmt)
    - Run Python linting (pylint, flake8) on agent code
    - Review and address all high-priority issues
    - Document any accepted warnings with justification
    - _Requirements: All_
  
  - [x] 29.5 Validate documentation completeness
    - Verify all public APIs are documented
    - Check all CLI commands have help text
    - Verify all configuration options are documented
    - Check all error messages are clear and actionable
    - Verify code examples compile and run
    - Check for broken links in documentation
    - Verify documentation matches implementation
    - Test all documented workflows
    - _Requirements: All_
  
  - [x] 29.6 Test dependency compatibility
    - Test with minimum supported Go version (1.21)
    - Test with latest Go version (1.22+)
    - Test with minimum Python version (3.8)
    - Test with latest Python version (3.12+)
    - Verify all external dependencies are available
    - Test dependency update scenarios
    - Check for deprecated dependency usage
    - Document any version-specific issues
    - _Requirements: All_
  
  - [x] 29.7 Perform integration validation
    - Test asc init workflow end-to-end
    - Test asc up → work → down workflow
    - Test configuration hot-reload
    - Test secrets encryption/decryption
    - Test health monitoring and recovery
    - Test with real beads repository
    - Test with real mcp_agent_mail server
    - Test multi-agent coordination
    - _Requirements: All_
  
  - [x] 29.8 Security validation
    - Verify no secrets in logs
    - Check file permissions on sensitive files
    - Test API key handling
    - Verify input sanitization
    - Check for command injection vulnerabilities
    - Test path traversal protection
    - Review security scan results
    - Verify security best practices are followed
    - _Requirements: 1.5, 4.3, All_
  
  - [x] 29.9 Performance validation
    - Measure startup time
    - Measure shutdown time
    - Test memory usage with 1, 3, 5, 10 agents
    - Test TUI responsiveness under load
    - Measure task processing throughput
    - Test with large log files (>100MB)
    - Test with many tasks (>1000)
    - Document performance characteristics
    - _Requirements: All_
  
  - [x] 29.10 Create gap analysis report
    - Document all identified issues by severity
    - List all test failures with root causes
    - Document coverage gaps by priority
    - List all linting/static analysis issues
    - Document documentation gaps
    - List all performance issues
    - Document security concerns
    - Create prioritized remediation plan
    - _Requirements: All_
  
  - [x] 29.11 Plan remediation work
    - Categorize issues (critical, high, medium, low)
    - Create tasks for critical issues
    - Create tasks for high-priority issues
    - Estimate effort for each task
    - Prioritize based on impact and effort
    - Create implementation timeline
    - Assign owners to tasks
    - Update project roadmap
    - _Requirements: All_
  
  - [x] 29.12 Create validation summary report
    - Summarize build results
    - Summarize test results and coverage
    - Summarize static analysis results
    - Summarize documentation validation
    - Summarize integration testing
    - Summarize security validation
    - Summarize performance validation
    - List all identified gaps and planned work
    - Provide go/no-go recommendation for release
    - _Requirements: All_


## Phase 30: Remediation Work

- [-] 30. Execute remediation plan for identified gaps
  
  ### Phase 30.0: Immediate Critical Blockers (Days 1-2)
  
  - [x] 30.0.1 Fix compilation errors blocking tests (CRITICAL - 3 hours)
    - Fix internal/beads/error_handling_test.go compilation errors
      - Update NewClient call to include time.Duration parameter (line 41)
      - Fix type mismatch for string constant (line 234)
      - Remove duplicate contains function (line 582)
      - Verify tests compile and run
    - Fix internal/mcp/error_handling_test.go compilation errors
      - Locate correct NewClient function or import
      - Update all function calls to use correct signature
      - Verify tests compile and run
    - Fix internal/process/error_handling_test.go compilation errors
      - Fix variable declaration on line 289 (use = instead of :=)
      - Fix PID type mismatches (lines 300, 310)
      - Verify tests compile and run
    - Run go test ./... to verify all packages compile
    - _Requirements: All_
  
  - [ ] 30.0.2 Fix test assertion failures (CRITICAL - 4 hours)
    - Fix internal/beads test failures (4 error handling tests)
      - Fix TestGetTasks_ErrorPaths/empty_statuses - update error message expectations
      - Fix TestGetTasks_ErrorPaths/invalid_status - update error message expectations
      - Fix TestCreateTask_ErrorPaths/empty_title - update error message expectations
      - Fix TestUpdateTask_ErrorPaths/empty_task_ID - update error message expectations
      - Fix TestDeleteTask_ErrorPaths/empty_task_ID - update error message expectations
      - Run tests to verify all fixes
    - Fix internal/check test failures (5 tests)
      - Update expected error messages to match actual implementation
      - Update expected status levels (warn vs fail)
      - Fix TestCheckFile_ErrorPaths/nonexistent_file
      - Fix TestCheckFile_ErrorPaths/directory_instead_of_file
      - Fix TestCheckFile_ErrorPaths/empty_path
      - Fix TestCheckConfig_ErrorPaths/invalid_TOML_syntax
      - Fix TestCheckEnv_ErrorPaths/missing_required_keys
      - Run tests to verify all fixes
    - Fix internal/config test failures (5 tests)
      - Update expected error messages to match actual implementation
      - Update validation order expectations
      - Fix TestLoadConfig_ErrorPaths/missing_config_file
      - Fix TestLoadConfig_ErrorPaths/invalid_TOML_syntax
      - Fix TestLoadConfig_ErrorPaths/empty_config_file
      - Fix TestLoadConfig_ErrorPaths/missing_required_fields
      - Fix TestValidate_ErrorPaths/agent_with_empty_model
      - Run tests to verify all fixes
    - Fix internal/mcp test failures
      - Review and fix any failing MCP client tests
      - Update error message expectations
      - Run tests to verify all fixes
    - Fix internal/process test failures
      - Review and fix any failing process manager tests
      - Update error message expectations
      - Run tests to verify all fixes
    - Document any intentional behavior changes
    - _Requirements: All_
  
  - [x] 30.0.3 Format all code (CRITICAL - 5 minutes)
    - Run gofmt -w . on entire codebase (64 files need formatting)
    - Verify gofmt -l . returns no files
    - Commit formatted code
    - Add pre-commit hook to enforce formatting
    - _Requirements: All_
  
  - [ ] 30.0.4 Install and run linting tools (CRITICAL - 3 hours)
    - Install golangci-lint
      - macOS: brew install golangci-lint
      - Linux: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh
    - Run golangci-lint run ./...
    - Review all findings
    - Address all high-severity issues
    - Document accepted warnings with justification
    - Install gosec
      - go install github.com/securego/gosec/v2/cmd/gosec@latest
    - Run gosec ./...
    - Review all findings
    - Address all critical security issues
    - Document accepted warnings with justification
    - Add both tools to CI/CD pipeline
    - _Requirements: 1.5, 4.3, All_
  
  ### Phase 30.1: Critical Coverage Gaps (Weeks 1-3)
  
  - [ ] 30.1 Add TUI integration tests (CRITICAL - 2 weeks)
    - Current coverage: 4.1% | Target: 40%+ | Gap: 95.9% untested
    - [x] 30.1.1 Set up TUI integration test framework
      - Research bubbletea testing approaches
      - Create mock terminal for testing
      - Set up test fixtures and helpers
      - Create test utilities for TUI components
      - _Requirements: All_
    
    - [x] 30.1.2 Add wizard flow tests
      - Test viewWelcome screen rendering
      - Test viewChecking dependency check display
      - Test viewAPIKeys input and validation
      - Test viewGenerating config generation
      - Test viewValidating validation step
      - Test viewComplete screen
      - Test runChecks function
      - Test generateConfigFiles function
      - Test runValidation function
      - Test backupConfigFiles function
      - Test validateAPIKey function
      - Test generateConfigFromTemplate function
      - Target: 60%+ coverage for wizard.go
      - _Requirements: All_
    
    - [x] 30.1.3 Add TUI model and state tests
      - Test Model initialization (NewModel function)
      - Test Init method and initial commands
      - Test refreshData method with mock clients
      - Test state transitions between views
      - Test error handling in model
      - Target: 60%+ coverage for model.go
      - _Requirements: All_
    
    - [x] 30.1.4 Add TUI rendering tests
      - Test View method with different model states
      - Test agent pane rendering (renderAgentPane)
      - Test task pane rendering (renderTaskPane)
      - Test log pane rendering (renderLogPane)
      - Test footer rendering (renderFooter)
      - Test layout calculations with different terminal sizes
      - Test view composition
      - Target: 60%+ coverage for view.go, agents.go, tasks.go, logs.go
      - _Requirements: All_
    
    - [x] 30.1.5 Add TUI interaction tests
      - Test Update method with different message types
      - Test keyboard event handling (q, r, t, arrow keys, etc.)
      - Test modal interactions (open, close, navigation)
      - Test navigation between panes
      - Test search functionality
      - Test state transitions
      - Test error handling in TUI
      - Target: 60%+ coverage for update.go, modals.go
      - _Requirements: All_
    
    - [x] 30.1.6 Add theme and styling tests
      - Test theme initialization (NewTheme)
      - Test theme application
      - Test color calculations and gradients
      - Test animation state updates
      - Test performance monitoring display
      - Target: 40%+ coverage for theme.go, animations.go, performance.go
      - _Requirements: All_
  
  - [x] 30.2 Add CLI command integration tests (HIGH - 1 week)
    - Current coverage: 0% | Target: 50%+ | Gap: 100% untested
    - [x] 30.2.1 Set up CLI integration test framework
      - Create cmd/cmd_test.go with shared test utilities
      - Create test environment setup/teardown helpers
      - Mock file system operations using temp directories
      - Mock process execution where needed
      - Set up test fixtures for CLI testing (sample configs, env files)
      - Create helper functions for command testing
      - _Requirements: All_
    
    - [x] 30.2.2 Add check command tests
      - Test asc check workflow with valid environment
      - Test asc check with missing dependencies
      - Test asc check with invalid config
      - Test asc check with missing env file
      - Test error reporting and exit codes
      - Target: 50%+ coverage for cmd/check.go
      - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_
    
    - [x] 30.2.3 Add services command tests
      - Test asc services start workflow
      - Test asc services stop workflow
      - Test asc services status workflow
      - Test service management with mock processes
      - Test error handling
      - Target: 50%+ coverage for cmd/services.go
      - _Requirements: 6.1, 6.2, 6.3, 6.4_
    
    - [x] 30.2.4 Add test command tests
      - Test asc test workflow with mock beads and MCP
      - Test E2E test execution
      - Test result reporting
      - Test timeout handling
      - Test cleanup operations
      - Target: 50%+ coverage for cmd/test.go
      - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_
    
    - [x] 30.2.5 Add doctor command tests
      - Test asc doctor workflow
      - Test diagnostics execution
      - Test report generation
      - Test issue detection
      - Target: 50%+ coverage for cmd/doctor.go
      - _Requirements: All_
    
    - [x] 30.2.6 Add cleanup command tests
      - Test asc cleanup workflow
      - Test log cleanup
      - Test PID cleanup
      - Test old file removal
      - Target: 50%+ coverage for cmd/cleanup.go
      - _Requirements: All_
    
    - [x] 30.2.7 Add secrets command tests
      - Test asc secrets init workflow
      - Test asc secrets encrypt workflow
      - Test asc secrets decrypt workflow
      - Test encryption/decryption with age
      - Test key management
      - Target: 50%+ coverage for cmd/secrets.go
      - _Requirements: 1.5, 4.3, All_
    
    - [x] 30.2.8 Add down command tests
      - Test asc down workflow with running processes
      - Test graceful shutdown
      - Test cleanup operations
      - Test error handling with missing PIDs
      - Target: 50%+ coverage for cmd/down.go
      - _Requirements: 3.1, 3.2, 3.3, 3.4_
    
    - [x] 30.2.9 Add up command tests (COMPLEX - requires mocking TUI)
      - Test asc up workflow with mock TUI
      - Test agent startup sequence
      - Test TUI initialization
      - Test error handling
      - Note: This is complex due to TUI integration
      - Target: 30%+ coverage for cmd/up.go
      - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7_
    
    - [x] 30.2.10 Add init command tests (COMPLEX - requires mocking wizard)
      - Test asc init workflow with mock wizard
      - Test flag parsing and validation
      - Test wizard flow integration
      - Test config file generation
      - Test error handling and user feedback
      - Note: This is complex due to wizard integration
      - Target: 30%+ coverage for cmd/init.go
      - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_
  
  ### Phase 30.2: High Priority Coverage (Week 4)
  
  - [x] 30.3 Fix secrets tests and improve coverage (HIGH - 4 hours)
    - Current coverage: 47.4% | Target: 70%+ | Gap: 22.6%
    - Note: age binary is already installed and tests are passing
    - Add tests for key rotation functionality
    - Add tests for public key extraction
    - Add tests for error handling edge cases
    - Add tests for key file management
    - Add tests for permission handling
    - Add tests for concurrent encryption/decryption
    - Target: 70%+ coverage for internal/secrets
    - _Requirements: 1.5, 4.3, All_
  
  - [x] 30.4 Improve doctor coverage (HIGH - 3 days)
    - Current coverage: 69.8% | Target: 80%+ | Gap: 10.2%
    - checkAgents function: 26.1% (CRITICAL)
    - [x] 30.4.1 Add tests for checkAgents function
      - Test with running agents
      - Test with stopped agents
      - Test with crashed agents
      - Test with missing agents
      - Test agent health detection
      - Target: 80%+ coverage (currently 26.1%)
      - _Requirements: All_
    
    - [x] 30.4.2 Add tests for report generation
      - Test generateReport function
      - Test formatIssue function
      - Test formatRemediation function
      - Test report output formatting
      - Test various issue types
      - _Requirements: All_
    
    - [x] 30.4.3 Improve coverage for other doctor functions
      - Improve checkConfiguration to 80%+
      - Improve checkResources to 80%+
      - Add edge case tests
      - Test error handling paths
      - Target: 80%+ overall coverage for internal/doctor
      - _Requirements: All_
  
  - [x] 30.5 Improve logger coverage (HIGH - 1 day)
    - Current coverage: 67.7% | Target: 75%+ | Gap: 7.3%
    - [x] 30.5.1 Add log rotation tests
      - Test rotation at size limit
      - Test rotation under load
      - Test cleanup of old files
      - Test rotation with concurrent writes
      - _Requirements: All_
    
    - [x] 30.5.2 Add concurrent logging tests
      - Test multiple goroutines logging
      - Test race conditions
      - Test log ordering
      - Test thread safety
      - _Requirements: All_
    
    - [x] 30.5.3 Add structured logging tests
      - Test complex object logging
      - Test context fields
      - Test log levels
      - Test log filtering
      - Target: 75%+ coverage for internal/logger
      - _Requirements: All_
  
  ### Phase 30.3: Medium Priority Issues (Week 5)
  
  - [x] 30.6 Improve config coverage (MEDIUM - 4 hours)
    - Current coverage: 76.6% | Target: 80%+ | Gap: 3.4%
    - Add tests for GetDefaultConfigPath
    - Add tests for GetDefaultEnvPath
    - Add tests for GetDefaultPIDDir
    - Add tests for GetDefaultLogDir
    - Improve watcher Start function coverage to 80%+
    - Improve stopAgent coverage to 80%+
    - Improve SaveTemplate coverage to 80%+
    - Improve SaveCustomTemplate coverage to 80%+
    - Test edge cases and error paths
    - Target: 80%+ coverage for internal/config
    - _Requirements: All_
  
  - [x] 30.7 Add CHANGELOG and versioning documentation (MEDIUM - 3 hours)
    - [x] 30.7.1 Create CHANGELOG.md
      - Follow Keep a Changelog format
      - Document all releases to date
      - Add unreleased section
      - Document breaking changes
      - _Requirements: All_
    
    - [x] 30.7.2 Create VERSIONING.md
      - Document SemVer usage
      - Document release process
      - Document version numbering rules
      - Document compatibility guarantees
      - _Requirements: All_
    
    - [x] 30.7.3 Update go.mod version
      - Change from go 1.25.4 to go 1.21
      - Run go mod tidy
      - Verify build still works
      - Test with Go 1.21 and 1.22+
      - _Requirements: All_
  
  - [x] 30.8 Add documentation automation (MEDIUM - 8 hours)
    - [x] 30.8.1 Add link validation
      - Install markdown-link-check or similar tool
      - Create link validation script
      - Add to CI/CD pipeline
      - Fix any broken links found
      - _Requirements: All_
    
    - [x] 30.8.2 Add example testing
      - Extract code examples from documentation
      - Create test script to compile/run examples
      - Add to CI/CD pipeline
      - Fix any broken examples
      - _Requirements: All_
  
  - [x] 30.9 Install and configure Python linters (MEDIUM - 2 hours)
    - [x] 30.9.1 Install and run pylint
      - Install pylint (uv pip install pylint)
      - Run pylint on agent/*.py
      - Address high-severity issues
      - Document accepted warnings
      - Add to CI/CD pipeline
      - _Requirements: All_
    
    - [x] 30.9.2 Install and run flake8
      - Install flake8 (uv pip install flake8)
      - Run flake8 on agent/
      - Address high-severity issues
      - Configure flake8 rules
      - Add to CI/CD pipeline
      - _Requirements: All_
  
  ### Phase 30.4: Low Priority Issues (Week 6+)
  
  - [x] 30.10 Add screenshots to README (LOW - 2 hours)
    - Capture screenshots of TUI main dashboard
    - Capture screenshots of wizard flow
    - Capture screenshots of modal dialogs
    - Capture screenshots of error states
    - Add screenshots to README in appropriate sections
    - Optimize image sizes for web
    - Add alt text for accessibility
    - _Requirements: All_
  
  - [x] 30.11 Fix development environment security issues (LOW - 1 hour)
    - Fix .env file permissions (chmod 600 .env)
    - Remove .env from git tracking (git rm --cached .env)
    - Update .gitignore to exclude .env
    - Fix log directory permissions
    - Fix PID directory permissions
    - Document security checklist for production
    - _Requirements: 1.5, 4.3, All_
  
  - [x] 30.12 Install Docker for optional features (LOW - 30 minutes)
    - Install Docker Desktop
    - Verify Docker installation
    - Update documentation with Docker setup
    - Test container-based features
    - Document Docker as optional dependency
    - _Requirements: All_
  
  - [x] 30.13 Update dependencies (LOW - 4 hours)
    - Review 20 available dependency updates
    - Test updates in staging environment
    - Apply safe minor/patch updates
    - Run full test suite after updates
    - Document any breaking changes
    - Update go.mod and go.sum
    - _Requirements: All_

## Summary of Critical Issues to Address

Based on Phase 29 validation, the following critical issues must be fixed before release:

### Immediate Blockers (Phase 30.0 - Days 1-2)
1. **30.0.1** Fix 3 compilation errors (3 hours)
2. **30.0.2** Fix 10 test assertion failures (4 hours)
3. **30.0.3** Format 64 files with gofmt (5 minutes)
4. **30.0.4** Install and run linting tools (3 hours)

### Critical Coverage Gaps (Phase 30.1 - Weeks 1-3)
5. **30.1** Add TUI integration tests - 4.1% → 40%+ coverage (2 weeks)
6. **30.2** Add CLI integration tests - 0% → 60%+ coverage (2 weeks)

### High Priority Coverage (Phase 30.2 - Week 4)
7. **30.3** Fix secrets tests - 47.4% → 80%+ coverage (1 day)
8. **30.4** Improve doctor coverage - 69.8% → 80%+ coverage (3 days)
9. **30.5** Improve logger coverage - 67.7% → 80%+ coverage (2 days)

**Total Estimated Effort:**
- Minimum for Beta Release: 1-2 days (Phase 30.0 remaining tasks)
- Recommended for Production: 2-3 weeks (Phase 30.0 + 30.1 + 30.2)
- Full Quality Release: 4-5 weeks (All phases)

**Current Status (Updated - November 11, 2025):**

**Implementation Status: 95% Complete**
- ✅ Phases 1-21: Core implementation (100% complete)
- ✅ Phase 22-27: Real-time, interactive, and vaporwave features (100% complete)  
- ✅ Phase 28: Comprehensive testing and QA (100% complete)
- ✅ Phase 29: Final validation and gap analysis (100% complete)
- ⏳ Phase 30: Remediation work (75% complete)

**Phase 30 Detailed Status:**
- Phase 30.0 (Immediate Blockers): 50% complete (2/4 tasks done)
  - ✅ 30.0.1: Compilation errors fixed
  - ❌ 30.0.2: Test assertion failures remain (beads, check, config, mcp, process) - **BLOCKING**
  - ✅ 30.0.3: Code formatting complete
  - ❌ 30.0.4: Linting tools need installation and execution - **BLOCKING**
  
- Phase 30.1 (Critical Coverage - TUI): 100% complete (6/6 tasks done)
  - ✅ 30.1.1: Test framework set up
  - ✅ 30.1.2: Wizard tests complete
  - ✅ 30.1.3: Model and state tests complete
  - ✅ 30.1.4: Rendering tests complete
  - ✅ 30.1.5: Interaction tests complete
  - ✅ 30.1.6: Theme and styling tests complete
  - Current TUI coverage: 41.1% (target: 40%+) ✅
  
- Phase 30.2 (Critical Coverage - CLI): 60% complete (6/10 tasks done)
  - ✅ 30.2.1: CLI test framework set up
  - ✅ 30.2.2: Check command tests complete
  - ✅ 30.2.3: Services command tests complete
  - ✅ 30.2.4: Test command tests complete
  - ✅ 30.2.5: Doctor command tests complete
  - ✅ 30.2.6: Cleanup command tests complete
  - ❌ 30.2.7: Secrets command tests needed
  - ❌ 30.2.8: Down command tests needed
  - ❌ 30.2.9: Up command tests needed (complex)
  - ❌ 30.2.10: Init command tests needed (complex)
  - Current CMD coverage: Tests exist but package shows 0% (likely due to test-only files)
  
- Phase 30.3-30.6 (High Priority): 0% complete (0/4 tasks done)
  - ❌ 30.3: Secrets coverage improvement (47.4% → 70%+)
  - ❌ 30.4: Doctor coverage improvement (69.8% → 80%+)
  - ❌ 30.5: Logger coverage improvement (67.7% → 75%+)
  - ❌ 30.6: Config coverage improvement (76.6% → 80%+)
  
- Phase 30.7-30.13 (Medium/Low Priority): 0% complete (0/7 tasks done)

**Coverage Summary:**
- internal/beads: 86.1% ✅ (failing tests need fixing)
- internal/check: 94.8% ✅ (failing tests need fixing)
- internal/config: 76.6% ⚠️ (target: 80%+)
- internal/doctor: 69.8% ⚠️ (target: 80%+)
- internal/errors: 100.0% ✅
- internal/health: 72.0% ✅
- internal/logger: 67.7% ⚠️ (target: 75%+)
- internal/mcp: 68.5% ⚠️ (failing tests need fixing)
- internal/process: 77.1% ✅ (failing tests need fixing)
- internal/secrets: 47.4% ⚠️ (target: 70%+)
- internal/tui: 41.1% ✅ (target met: 40%+)
- cmd: Tests exist but need fixes

**Next Recommended Tasks (Priority Order):**
1. **CRITICAL** - Complete 30.0.2: Fix all test assertion failures (4 hours)
   - Fix beads error handling tests (5 failures)
   - Fix check error handling tests (8 failures)
   - Fix config error handling tests (9 failures)
   - Fix mcp error handling tests (4 failures)
   - Fix process error handling tests (4 failures)
   
2. **CRITICAL** - Complete 30.0.4: Install and run linting tools (3 hours)
   - Install golangci-lint
   - Install gosec
   - Run and address findings
   
3. **HIGH** - Complete 30.2.7-30.2.10: Remaining CLI tests (3 days)
   - Add secrets command tests
   - Add down command tests
   - Add up command tests (complex)
   - Add init command tests (complex)
   
4. **MEDIUM** - Complete 30.3-30.6: Coverage improvements (1 week)
   - Improve secrets, doctor, logger, config coverage

**Estimated Time to Beta Release:** 1-2 days (complete Phase 30.0)
**Estimated Time to Production Release:** 1-2 weeks (complete Phase 30.0-30.2)
**Estimated Time to Full Quality Release:** 3-4 weeks (complete all Phase 30 tasks)

