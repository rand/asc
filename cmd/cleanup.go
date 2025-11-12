package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/rand/asc/internal/logger"
)

var (
	cleanupDays   int
	cleanupDryRun bool
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up old log files",
	Long: `Remove log files older than the specified number of days.
This helps manage disk space by removing old logs that are no longer needed.

By default, logs older than 30 days are removed.`,
	Run: runCleanup,
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().IntVar(&cleanupDays, "days", 30, "Remove logs older than this many days")
	cleanupCmd.Flags().BoolVar(&cleanupDryRun, "dry-run", false, "Show what would be deleted without actually deleting")
}

func runCleanup(cmd *cobra.Command, args []string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	logsDir := filepath.Join(homeDir, ".asc", "logs")

	if cleanupDryRun {
		fmt.Printf("Dry run: would remove logs older than %d days from %s\n", cleanupDays, logsDir)
		// TODO: Implement dry run listing
		return
	}

	fmt.Printf("Cleaning up logs older than %d days from %s...\n", cleanupDays, logsDir)

	maxAge := time.Duration(cleanupDays) * 24 * time.Hour
	if err := logger.CleanupOldLogs(logsDir, maxAge); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to cleanup logs: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Log cleanup completed")
}
