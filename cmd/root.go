package cmd

import (
	"github.com/spf13/cobra"
	_ "github.com/spf13/viper"
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/lipgloss"
	_ "github.com/charmbracelet/bubbles"
	"github.com/yourusername/asc/internal/logger"
)

var (
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "asc",
	Short: "Agent Stack Controller - Orchestrate your AI coding agent colony",
	Long: `asc is a command-line orchestration tool that manages a local colony of AI coding agents.
It provides developers with a mission control interface for starting, monitoring, and 
coordinating headless background agents that work collaboratively on software development tasks.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logger.SetLevel(logger.DEBUG)
		}
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging (debug level)")
}
