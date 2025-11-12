package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/rand/asc/internal/beads"
	"github.com/rand/asc/internal/check"
	"github.com/rand/asc/internal/config"
	"github.com/rand/asc/internal/logger"
	"github.com/rand/asc/internal/mcp"
	"github.com/rand/asc/internal/process"
	"github.com/rand/asc/internal/secrets"
	"github.com/rand/asc/internal/tui"
)

var (
	debugMode bool
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start all agents and services with TUI dashboard",
	Long: `Start the agent stack by:
- Running dependency checks
- Starting the mcp_agent_mail service
- Launching all configured agents
- Opening the TUI dashboard for monitoring`,
	Run: runUp,
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().BoolVar(&debugMode, "debug", false, "Enable debug mode with verbose output")
}

func runUp(cmd *cobra.Command, args []string) {
	// Initialize logger
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		osExit(1)
	}
	defer logger.Close()

	// Enable debug mode if flag is set
	if debugMode {
		logger.SetLevel(logger.DEBUG)
		logger.SetFormat(logger.FormatJSON)
		logger.Info("Debug mode enabled")
	}

	// Default paths
	configPath := "asc.toml"
	envPath := ".env"

	logger.Debug("Starting asc up command with config=%s, env=%s", configPath, envPath)

	// Step 0: Auto-decrypt secrets if needed
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		// Check if encrypted version exists
		if _, err := os.Stat(envPath + ".age"); err == nil {
			fmt.Println("🔐 Decrypting secrets...")
			logger.Debug("Decrypting secrets from %s.age", envPath)
			secretsManager := secrets.NewManager()
			if err := secretsManager.DecryptEnv(envPath); err != nil {
				logger.Error("Failed to decrypt secrets: %v", err)
				fmt.Fprintf(os.Stderr, "Failed to decrypt secrets: %v\n", err)
				fmt.Fprintln(os.Stderr, "Run 'asc secrets decrypt' manually or 'asc init' to set up encryption.")
				osExit(1)
			}
			fmt.Println("✓ Secrets decrypted")
			logger.Debug("Secrets decrypted successfully")
		}
	}

	// Step 1: Run silent dependency check
	logger.Debug("Running dependency checks")
	checker := check.NewChecker(configPath, envPath)
	results := checker.RunAll()

	if debugMode {
		for _, result := range results {
			logger.WithFields(logger.Fields{
				"check":  result.Name,
				"status": result.Status,
			}).Debug("Dependency check result: %s", result.Message)
		}
	}

	if check.HasFailures(results) {
		logger.Error("Dependency check failed")
		fmt.Fprintln(os.Stderr, "Dependency check failed. Run 'asc check' for details.")
		osExit(1)
	}
	logger.Debug("All dependency checks passed")

	// Step 2: Load configuration from asc.toml
	logger.Debug("Loading configuration from %s", configPath)
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		osExit(1)
	}
	if debugMode {
		logger.WithFields(logger.Fields{
			"agents":   len(cfg.Agents),
			"mcp_url":  cfg.Services.MCPAgentMail.URL,
			"beads_db": cfg.Core.BeadsDBPath,
		}).Debug("Configuration loaded successfully")
	}

	// Step 3: Load environment variables from .env
	logger.Debug("Loading environment variables from %s", envPath)
	if err := config.LoadAndValidateEnv(envPath); err != nil {
		logger.Error("Failed to load environment: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to load environment: %v\n", err)
		osExit(1)
	}
	logger.Debug("Environment variables loaded successfully")

	// Step 4: Initialize process manager with ~/.asc/pids and ~/.asc/logs
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get home directory: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to get home directory: %v\n", err)
		osExit(1)
	}

	pidsDir := filepath.Join(homeDir, ".asc", "pids")
	logsDir := filepath.Join(homeDir, ".asc", "logs")

	logger.Debug("Initializing process manager with pids=%s, logs=%s", pidsDir, logsDir)
	procManager, err := process.NewManager(pidsDir, logsDir)
	if err != nil {
		logger.Error("Failed to initialize process manager: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to initialize process manager: %v\n", err)
		osExit(1)
	}

	// Step 5: Start mcp_agent_mail service
	fmt.Println("Starting mcp_agent_mail service...")
	mcpEnv := buildMCPEnv()
	mcpCmd, mcpArgs := parseCommand(cfg.Services.MCPAgentMail.StartCommand)
	logger.WithFields(logger.Fields{
		"command": mcpCmd,
		"args":    mcpArgs,
	}).Debug("Starting mcp_agent_mail service")
	_, err = procManager.Start("mcp_agent_mail", mcpCmd, mcpArgs, mcpEnv)
	if err != nil {
		logger.Error("Failed to start mcp_agent_mail: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to start mcp_agent_mail: %v\n", err)
		osExit(1)
	}
	logger.Info("mcp_agent_mail service started successfully")

	// Step 6: Launch agent processes (handled in subtask 16.2)
	logger.Debug("Launching agent processes")
	if err := launchAgents(cfg, procManager); err != nil {
		logger.Error("Failed to launch agents: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to launch agents: %v\n", err)
		// Clean up: stop mcp_agent_mail
		_ = procManager.StopAll()
		osExit(1)
	}

	// Step 7: Initialize and run TUI (handled in subtask 16.3)
	logger.Debug("Initializing TUI dashboard")
	if err := runTUI(cfg, procManager, debugMode); err != nil {
		logger.Error("TUI error: %v", err)
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		// Clean up: stop all processes
		_ = procManager.StopAll()
		osExit(1)
	}

	// Clean up on exit
	fmt.Println("\nShutting down agent stack...")
	logger.Info("Shutting down agent stack")
	if err := procManager.StopAll(); err != nil {
		logger.Error("Error during shutdown: %v", err)
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
	}
	fmt.Println("Agent stack is offline")
	logger.Info("Agent stack is offline")
}

// parseCommand parses a command string into command and args
// For example: "python -m mcp_agent_mail.server" -> ("python", ["-m", "mcp_agent_mail.server"])
func parseCommand(cmdStr string) (string, []string) {
	// Simple space-based parsing
	// For more complex parsing with quotes, we'd need a proper shell parser
	parts := []string{}
	current := ""
	inQuote := false
	
	for _, char := range cmdStr {
		if char == '"' || char == '\'' {
			inQuote = !inQuote
		} else if char == ' ' && !inQuote {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	if len(parts) == 0 {
		return "", []string{}
	}
	
	return parts[0], parts[1:]
}

// buildMCPEnv builds environment variables for the MCP service
func buildMCPEnv() []string {
	// Pass through all current environment variables
	env := os.Environ()
	return env
}

// launchAgents starts all configured agent processes
func launchAgents(cfg *config.Config, procManager process.ProcessManager) error {
	fmt.Printf("Launching %d agent(s)...\n", len(cfg.Agents))
	logger.Info("Launching %d agent(s)", len(cfg.Agents))

	// Iterate through agents in config
	for agentName, agentCfg := range cfg.Agents {
		fmt.Printf("  Starting agent: %s (model: %s)...\n", agentName, agentCfg.Model)
		logger.WithFields(logger.Fields{
			"agent": agentName,
			"model": agentCfg.Model,
			"phases": agentCfg.Phases,
		}).Info("Starting agent")

		// Build environment variables for this agent
		agentEnv := buildAgentEnv(agentName, agentCfg, cfg)

		if debugMode {
			logger.WithFields(logger.Fields{
				"agent": agentName,
				"env_count": len(agentEnv),
			}).Debug("Built agent environment variables")
		}

		// Parse command into command and args
		cmd, args := parseCommand(agentCfg.Command)

		logger.WithFields(logger.Fields{
			"agent": agentName,
			"command": cmd,
			"args": args,
		}).Debug("Parsed agent command")

		// Start the agent using process manager
		pid, err := procManager.Start(agentName, cmd, args, agentEnv)
		if err != nil {
			logger.WithFields(logger.Fields{
				"agent": agentName,
			}).Error("Failed to start agent: %v", err)
			return fmt.Errorf("failed to start agent '%s': %w", agentName, err)
		}

		fmt.Printf("  ✓ Agent %s started\n", agentName)
		logger.WithFields(logger.Fields{
			"agent": agentName,
			"pid": pid,
		}).Info("Agent started successfully")
	}

	fmt.Println("All agents started successfully")
	logger.Info("All agents started successfully")
	return nil
}

// buildAgentEnv builds environment variables for an agent process
func buildAgentEnv(agentName string, agentCfg config.AgentConfig, cfg *config.Config) []string {
	// Start with all current environment variables (includes API keys from .env)
	env := os.Environ()

	// Add agent-specific environment variables
	env = append(env, fmt.Sprintf("AGENT_NAME=%s", agentName))
	env = append(env, fmt.Sprintf("AGENT_MODEL=%s", agentCfg.Model))
	
	// Convert phases array to comma-separated string
	phases := ""
	for i, phase := range agentCfg.Phases {
		if i > 0 {
			phases += ","
		}
		phases += phase
	}
	env = append(env, fmt.Sprintf("AGENT_PHASES=%s", phases))

	// Add MCP and beads configuration
	env = append(env, fmt.Sprintf("MCP_MAIL_URL=%s", cfg.Services.MCPAgentMail.URL))
	env = append(env, fmt.Sprintf("BEADS_DB_PATH=%s", cfg.Core.BeadsDBPath))

	return env
}

// runTUI initializes and runs the TUI dashboard
func runTUI(cfg *config.Config, procManager process.ProcessManager, debug bool) error {
	// Clear terminal screen
	fmt.Print("\033[H\033[2J")

	logger.Debug("Initializing beads client with path=%s", cfg.Core.BeadsDBPath)
	// Initialize beads client with 5 second refresh interval
	beadsClient := beads.NewClient(cfg.Core.BeadsDBPath, 5*time.Second)

	logger.Debug("Initializing MCP client with url=%s", cfg.Services.MCPAgentMail.URL)
	// Initialize MCP client
	mcpClient := mcp.NewHTTPClient(cfg.Services.MCPAgentMail.URL)

	// Create bubbletea Model with config and clients
	model := tui.NewModel(*cfg, beadsClient, mcpClient, procManager)
	model.SetDebugMode(debug)

	logger.Info("Starting TUI dashboard")
	// Start TUI event loop with tea.NewProgram
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program and handle exit
	finalModel, err := program.Run()
	if err != nil {
		logger.Error("TUI error: %v", err)
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if there was an error in the final model state
	if m, ok := finalModel.(tui.Model); ok {
		// Cleanup WebSocket and other resources
		m.Cleanup()
		
		if m.GetError() != nil {
			logger.Error("TUI exited with error: %v", m.GetError())
			return fmt.Errorf("TUI exited with error: %w", m.GetError())
		}
	}

	logger.Info("TUI exited normally")
	return nil
}
