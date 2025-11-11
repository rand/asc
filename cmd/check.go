package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/asc/internal/check"
)

// osExit is a variable that can be mocked in tests
var osExit = os.Exit

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify environment setup and dependencies",
	Long: `Verify that all required dependencies are installed and properly configured.
This includes checking for required binaries (git, python3, uv, bd), 
validating the asc.toml configuration file, and verifying API keys in .env file.`,
	Run: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) {
	// Default paths
	configPath := "asc.toml"
	envPath := ".env"

	// Create checker instance
	checker := check.NewChecker(configPath, envPath)

	// Run all checks
	results := checker.RunAll()

	// Format and print results
	output := check.FormatResults(results)
	fmt.Println(output)

	// Exit with appropriate status code
	if check.HasFailures(results) {
		osExit(1)
	}
	osExit(0)
}
