package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/process"
)

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage long-running services",
	Long: `Manage long-running services like mcp_agent_mail independently from agents.
This allows you to control the communication server without affecting agent processes.`,
}

var servicesStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the mcp_agent_mail service",
	Long:  `Start the mcp_agent_mail server as a background process.`,
	Run:   runServicesStart,
}

var servicesStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the mcp_agent_mail service",
	Long:  `Terminate the mcp_agent_mail server process.`,
	Run:   runServicesStop,
}

var servicesStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if the mcp_agent_mail service is running",
	Long:  `Report whether the mcp_agent_mail server is currently running.`,
	Run:   runServicesStatus,
}

func init() {
	rootCmd.AddCommand(servicesCmd)
	servicesCmd.AddCommand(servicesStartCmd)
	servicesCmd.AddCommand(servicesStopCmd)
	servicesCmd.AddCommand(servicesStatusCmd)
}

// getProcessManager creates a process manager instance with default directories
func getProcessManager() (*process.Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	pidDir := filepath.Join(homeDir, ".asc", "pids")
	logDir := filepath.Join(homeDir, ".asc", "logs")

	return process.NewManager(pidDir, logDir)
}

// runServicesStart starts the mcp_agent_mail service
func runServicesStart(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load(config.DefaultConfigPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create process manager
	pm, err := getProcessManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create process manager: %v\n", err)
		os.Exit(1)
	}

	// Check if service is already running
	info, err := pm.GetProcessInfo("mcp_agent_mail")
	if err == nil && pm.IsRunning(info.PID) {
		fmt.Printf("mcp_agent_mail is already running (PID %d)\n", info.PID)
		os.Exit(0)
	}

	// Parse the start command
	cmdParts := strings.Fields(cfg.Services.MCPAgentMail.StartCommand)
	if len(cmdParts) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Invalid start command in configuration\n")
		os.Exit(1)
	}

	command := cmdParts[0]
	cmdArgs := cmdParts[1:]

	// Start the service
	pid, err := pm.Start("mcp_agent_mail", command, cmdArgs, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to start mcp_agent_mail: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ mcp_agent_mail started (PID %d)\n", pid)
	fmt.Printf("  URL: %s\n", cfg.Services.MCPAgentMail.URL)
	
	homeDir, _ := os.UserHomeDir()
	logPath := filepath.Join(homeDir, ".asc", "logs", "mcp_agent_mail.log")
	fmt.Printf("  Log: %s\n", logPath)
}

// runServicesStop stops the mcp_agent_mail service
func runServicesStop(cmd *cobra.Command, args []string) {
	// Create process manager
	pm, err := getProcessManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create process manager: %v\n", err)
		os.Exit(1)
	}

	// Get process info
	info, err := pm.GetProcessInfo("mcp_agent_mail")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: mcp_agent_mail is not running\n")
		os.Exit(1)
	}

	// Check if process is actually running
	if !pm.IsRunning(info.PID) {
		fmt.Fprintf(os.Stderr, "Error: mcp_agent_mail is not running (stale PID file)\n")
		// Clean up stale PID file
		homeDir, _ := os.UserHomeDir()
		pidFile := filepath.Join(homeDir, ".asc", "pids", "mcp_agent_mail.json")
		os.Remove(pidFile)
		os.Exit(1)
	}

	// Stop the service
	fmt.Printf("Stopping mcp_agent_mail (PID %d)...\n", info.PID)
	if err := pm.Stop(info.PID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to stop mcp_agent_mail: %v\n", err)
		os.Exit(1)
	}

	// Clean up PID file
	homeDir, _ := os.UserHomeDir()
	pidFile := filepath.Join(homeDir, ".asc", "pids", "mcp_agent_mail.json")
	os.Remove(pidFile)

	fmt.Println("✓ mcp_agent_mail stopped")
}

// runServicesStatus checks if the mcp_agent_mail service is running
func runServicesStatus(cmd *cobra.Command, args []string) {
	// Create process manager
	pm, err := getProcessManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create process manager: %v\n", err)
		os.Exit(1)
	}

	// Get process info
	info, err := pm.GetProcessInfo("mcp_agent_mail")
	if err != nil {
		fmt.Println("mcp_agent_mail: ○ stopped")
		os.Exit(0)
	}

	// Check if process is running
	if pm.IsRunning(info.PID) {
		fmt.Printf("mcp_agent_mail: ● running (PID %d)\n", info.PID)
		fmt.Printf("  Started: %s\n", info.StartedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Log: %s\n", info.LogFile)
	} else {
		fmt.Println("mcp_agent_mail: ○ stopped (stale PID file)")
		// Clean up stale PID file
		homeDir, _ := os.UserHomeDir()
		pidFile := filepath.Join(homeDir, ".asc", "pids", "mcp_agent_mail.json")
		os.Remove(pidFile)
	}
}
