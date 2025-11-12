package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yourusername/asc/internal/process"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Gracefully shut down all agents and services",
	Long: `Stop the agent stack by:
- Stopping all agent processes
- Stopping the mcp_agent_mail service
- Cleaning up PID files
- Reporting shutdown status`,
	Run: runDown,
}

func init() {
	rootCmd.AddCommand(downCmd)
}

func runDown(cmd *cobra.Command, args []string) {
	// Initialize process manager with ~/.asc/pids and ~/.asc/logs
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get home directory: %v\n", err)
		osExit(1)
	}

	pidsDir := filepath.Join(homeDir, ".asc", "pids")
	logsDir := filepath.Join(homeDir, ".asc", "logs")

	procManager, err := process.NewManager(pidsDir, logsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize process manager: %v\n", err)
		osExit(1)
	}

	// List all managed processes
	processes, err := procManager.ListProcesses()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to list processes: %v\n", err)
		osExit(1)
	}

	if len(processes) == 0 {
		fmt.Println("No running processes found")
		fmt.Println("Agent stack is offline")
		return
	}

	fmt.Printf("Shutting down %d process(es)...\n", len(processes))

	// Stop all processes using process manager
	// This will handle both agents and mcp_agent_mail service
	if err := procManager.StopAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Some processes failed to stop cleanly: %v\n", err)
		// Continue anyway to print confirmation
	}

	// Print confirmation message
	fmt.Println("Agent stack is offline")
}
