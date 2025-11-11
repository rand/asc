package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/asc/internal/doctor"
	"github.com/yourusername/asc/internal/logger"
)

var (
	doctorFix    bool
	doctorVerbose bool
	doctorJSON   bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose and fix common issues with the agent stack",
	Long: `Run comprehensive diagnostics on the agent stack to detect and fix common issues.

The doctor command checks for:
- Configuration problems
- Corrupted state (PIDs, logs)
- Permission issues
- Resource problems
- Network connectivity
- Agent health issues

Use --fix to automatically remediate detected issues where possible.`,
	Run: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Automatically fix issues where possible")
	doctorCmd.Flags().BoolVar(&doctorVerbose, "verbose", false, "Show detailed diagnostic information")
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output results in JSON format")
}

func runDoctor(cmd *cobra.Command, args []string) {
	logger.Info("Running asc doctor diagnostics...")
	
	// Default paths
	configPath := "asc.toml"
	envPath := ".env"
	
	// Create doctor instance
	doc, err := doctor.NewDoctor(configPath, envPath)
	if err != nil {
		logger.Error("Failed to initialize doctor: %v", err)
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize doctor: %v\n", err)
		osExit(1)
	}
	
	// Run diagnostics
	report, err := doc.RunDiagnostics()
	if err != nil {
		logger.Error("Failed to run diagnostics: %v", err)
		fmt.Fprintf(os.Stderr, "Error: Failed to run diagnostics: %v\n", err)
		osExit(1)
	}
	
	// Apply fixes if requested
	if doctorFix {
		logger.Info("Applying automatic fixes...")
		fixReport, err := doc.ApplyFixes(report)
		if err != nil {
			logger.Error("Failed to apply fixes: %v", err)
			fmt.Fprintf(os.Stderr, "Error: Failed to apply fixes: %v\n", err)
			osExit(1)
		}
		report.FixesApplied = fixReport
	}
	
	// Output results
	if doctorJSON {
		output, err := report.ToJSON()
		if err != nil {
			logger.Error("Failed to format JSON output: %v", err)
			fmt.Fprintf(os.Stderr, "Error: Failed to format JSON output: %v\n", err)
			osExit(1)
		}
		fmt.Println(output)
	} else {
		output := report.Format(doctorVerbose)
		fmt.Println(output)
	}
	
	// Exit with appropriate code
	if report.HasCriticalIssues() {
		osExit(1)
	}
	osExit(0)
}
