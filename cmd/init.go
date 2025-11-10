package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yourusername/asc/internal/tui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize asc with interactive setup wizard",
	Long: `Launch an interactive setup wizard that guides you through:
- Checking for required dependencies
- Installing missing components
- Configuring API keys
- Generating default configuration files
- Validating the setup`,
	Run: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	// Launch the interactive setup wizard
	wizard := tui.NewWizard()
	if err := wizard.Run(); err != nil {
		cmd.PrintErrf("Error running setup wizard: %v\n", err)
		return
	}
}
