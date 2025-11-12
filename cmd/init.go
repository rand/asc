package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/tui"
)

var (
	templateFlag      string
	listTemplatesFlag bool
	saveTemplateFlag  string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize asc with interactive setup wizard",
	Long: `Launch an interactive setup wizard that guides you through:
- Checking for required dependencies
- Installing missing components
- Configuring API keys
- Generating default configuration files
- Validating the setup

Templates:
  --template=solo   Single agent for individual development
  --template=team   Planner, coder, and tester agents
  --template=swarm  Multiple agents per phase for parallel work`,
	Run: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&templateFlag, "template", "", "Use a predefined template (solo, team, swarm)")
	initCmd.Flags().BoolVar(&listTemplatesFlag, "list-templates", false, "List all available templates")
	initCmd.Flags().StringVar(&saveTemplateFlag, "save-template", "", "Save current config as a custom template")
}

func runInit(cmd *cobra.Command, args []string) {
	// Handle --list-templates flag
	if listTemplatesFlag {
		listTemplates(cmd)
		return
	}

	// Handle --save-template flag
	if saveTemplateFlag != "" {
		saveTemplate(cmd)
		return
	}

	// Launch the interactive setup wizard with optional template
	wizard := tui.NewWizard()
	if templateFlag != "" {
		wizard.SetTemplate(templateFlag)
	}
	if err := wizard.Run(); err != nil {
		cmd.PrintErrf("Error running setup wizard: %v\n", err)
		return
	}
}

func listTemplates(cmd *cobra.Command) {
	cmd.Println("Available templates:")
	cmd.Println()

	// List built-in templates
	cmd.Println("Built-in templates:")
	for _, tmpl := range config.ListTemplates() {
		cmd.Printf("  %s - %s\n", tmpl.Name, tmpl.Description)
	}

	// List custom templates
	customTemplates, err := config.ListCustomTemplates()
	if err != nil {
		cmd.PrintErrf("Warning: Failed to load custom templates: %v\n", err)
		return
	}

	if len(customTemplates) > 0 {
		cmd.Println()
		cmd.Println("Custom templates:")
		for _, tmpl := range customTemplates {
			cmd.Printf("  %s - %s\n", tmpl.Name, tmpl.Description)
		}
	}
}

func saveTemplate(cmd *cobra.Command) {
	configPath := config.DefaultConfigPath()
	if err := config.SaveCustomTemplate(configPath, saveTemplateFlag); err != nil {
		cmd.PrintErrf("Error saving template: %v\n", err)
		return
	}
	cmd.Printf("Template '%s' saved successfully\n", saveTemplateFlag)
}
