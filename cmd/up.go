package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/yourusername/asc/internal/beads"
	"github.com/yourusername/asc/internal/check"
	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/mcp"
	"github.com/yourusername/asc/internal/process"
	"github.com/yourusername/asc/internal/tui"
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
}

func runUp(cmd *cobra.Command, args []string) {
	// Default paths
	configPath := "asc.toml"
	envPath := ".env"

	// Step 1: Run silent dependency check
	checker := check.NewChecker(configPath, envPath)
	results := checker.RunAll()

	if check.HasFailures(results) {
		fmt.Fprintln(os.Stderr, "Dependency check failed. Run 'asc check' for details.")
		os.Exit(1)
	}

	// Step 2: Load configuration from asc.toml
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Load environment variables from .env
	if err := config.LoadAndValidateEnv(envPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load environment: %v\n", err)
		os.Exit(1)
	}

	// Step 4: Initialize process manager with ~/.asc/pids and ~/.asc/logs
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	pidsDir := filepath.Join(homeDir, ".asc", "pids")
	logsDir := filepath.Join(homeDir, ".asc", "logs")

	procManager, err := process.NewManager(pidsDir, logsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize process manager: %v\n", err)
		os.Exit(1)
	}

	// Step 5: Start mcp_agent_mail service
	fmt.Println("Starting mcp_agent_mail service...")
	mcpEnv := buildMCPEnv()
	mcpCmd, mcpArgs := parseCommand(cfg.Services.MCPAgentMail.StartCommand)
	_, err = procManager.Start("mcp_agent_mail", mcpCmd, mcpArgs, mcpEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start mcp_agent_mail: %v\n", err)
		os.Exit(1)
	}

	// Step 6: Launch agent processes (handled in subtask 16.2)
	if err := launchAgents(cfg, procManager); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to launch agents: %v\n", err)
		// Clean up: stop mcp_agent_mail
		_ = procManager.StopAll()
		os.Exit(1)
	}

	// Step 7: Initialize and run TUI (handled in subtask 16.3)
	if err := runTUI(cfg, procManager); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		// Clean up: stop all processes
		_ = procManager.StopAll()
		os.Exit(1)
	}

	// Clean up on exit
	fmt.Println("\nShutting down agent stack...")
	if err := procManager.StopAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
	}
	fmt.Println("Agent stack is offline")
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

	// Iterate through agents in config
	for agentName, agentCfg := range cfg.Agents {
		fmt.Printf("  Starting agent: %s (model: %s)...\n", agentName, agentCfg.Model)

		// Build environment variables for this agent
		agentEnv := buildAgentEnv(agentName, agentCfg, cfg)

		// Parse command into command and args
		cmd, args := parseCommand(agentCfg.Command)

		// Start the agent using process manager
		_, err := procManager.Start(agentName, cmd, args, agentEnv)
		if err != nil {
			return fmt.Errorf("failed to start agent '%s': %w", agentName, err)
		}

		fmt.Printf("  ✓ Agent %s started\n", agentName)
	}

	fmt.Println("All agents started successfully")
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
func runTUI(cfg *config.Config, procManager process.ProcessManager) error {
	// Clear terminal screen
	fmt.Print("\033[H\033[2J")

	// Initialize beads client with 5 second refresh interval
	beadsClient := beads.NewClient(cfg.Core.BeadsDBPath, 5*time.Second)

	// Initialize MCP client
	mcpClient := mcp.NewHTTPClient(cfg.Services.MCPAgentMail.URL)

	// Create bubbletea Model with config and clients
	model := tui.NewModel(*cfg, beadsClient, mcpClient, procManager)

	// Start TUI event loop with tea.NewProgram
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program and handle exit
	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if there was an error in the final model state
	if m, ok := finalModel.(tui.Model); ok {
		if m.GetError() != nil {
			return fmt.Errorf("TUI exited with error: %w", m.GetError())
		}
	}

	return nil
}
