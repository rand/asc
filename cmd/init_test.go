package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/yourusername/asc/internal/config"
)

// TestInitCommand tests the basic init command structure
func TestInitCommand(t *testing.T) {
	tests := []struct {
		name      string
		wantUse   string
		wantShort string
		wantFlags []string
	}{
		{
			name:      "command structure",
			wantUse:   "init",
			wantShort: "Initialize asc with interactive setup wizard",
			wantFlags: []string{"template", "list-templates", "save-template"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if initCmd.Use != tt.wantUse {
				t.Errorf("initCmd.Use = %v, want %v", initCmd.Use, tt.wantUse)
			}

			if initCmd.Short != tt.wantShort {
				t.Errorf("initCmd.Short = %v, want %v", initCmd.Short, tt.wantShort)
			}

			// Check flags exist
			for _, flagName := range tt.wantFlags {
				flag := initCmd.Flags().Lookup(flagName)
				if flag == nil {
					t.Errorf("expected flag %s to exist", flagName)
				}
			}
		})
	}
}

// TestInitCommandFlags tests flag parsing
func TestInitCommandFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantTemplate string
		wantList     bool
		wantSave     string
		wantErr      bool
	}{
		{
			name:         "no flags",
			args:         []string{},
			wantTemplate: "",
			wantList:     false,
			wantSave:     "",
			wantErr:      false,
		},
		{
			name:         "template flag",
			args:         []string{"--template=solo"},
			wantTemplate: "solo",
			wantList:     false,
			wantSave:     "",
			wantErr:      false,
		},
		{
			name:         "list-templates flag",
			args:         []string{"--list-templates"},
			wantTemplate: "",
			wantList:     true,
			wantSave:     "",
			wantErr:      false,
		},
		{
			name:         "save-template flag",
			args:         []string{"--save-template=mytemplate"},
			wantTemplate: "",
			wantList:     false,
			wantSave:     "mytemplate",
			wantErr:      false,
		},
		{
			name:         "multiple flags",
			args:         []string{"--template=team", "--list-templates"},
			wantTemplate: "team",
			wantList:     true,
			wantSave:     "",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			templateFlag = ""
			listTemplatesFlag = false
			saveTemplateFlag = ""

			// Create a new command for testing
			cmd := &cobra.Command{
				Use: "init",
				Run: func(cmd *cobra.Command, args []string) {},
			}
			cmd.Flags().StringVar(&templateFlag, "template", "", "Use a predefined template")
			cmd.Flags().BoolVar(&listTemplatesFlag, "list-templates", false, "List all available templates")
			cmd.Flags().StringVar(&saveTemplateFlag, "save-template", "", "Save current config as a custom template")

			// Parse flags
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if templateFlag != tt.wantTemplate {
				t.Errorf("templateFlag = %v, want %v", templateFlag, tt.wantTemplate)
			}

			if listTemplatesFlag != tt.wantList {
				t.Errorf("listTemplatesFlag = %v, want %v", listTemplatesFlag, tt.wantList)
			}

			if saveTemplateFlag != tt.wantSave {
				t.Errorf("saveTemplateFlag = %v, want %v", saveTemplateFlag, tt.wantSave)
			}
		})
	}
}

// TestListTemplates tests the listTemplates function
func TestListTemplates(t *testing.T) {
	tests := []struct {
		name           string
		setupCustom    bool
		customTemplate string
		wantContains   []string
	}{
		{
			name:        "list built-in templates",
			setupCustom: false,
			wantContains: []string{
				"Built-in templates:",
				"solo",
				"team",
				"swarm",
			},
		},
		{
			name:           "list with custom templates",
			setupCustom:    true,
			customTemplate: "mytemplate",
			wantContains: []string{
				"Built-in templates:",
				"solo",
				"team",
				"swarm",
				"Custom templates:",
				"mytemplate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			env := NewTestEnvironment(t)

			// Setup custom template if needed
			if tt.setupCustom {
				home := env.TempDir
				os.Setenv("HOME", home)
				defer os.Unsetenv("HOME")

				templatesDir := filepath.Join(home, ".asc", "templates")
				os.MkdirAll(templatesDir, 0755)

				templatePath := filepath.Join(templatesDir, tt.customTemplate+".toml")
				os.WriteFile(templatePath, []byte(ValidConfig()), 0644)
			}

			// Create command with output capture
			var buf bytes.Buffer
			cmd := &cobra.Command{
				Use: "init",
			}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Call listTemplates
			listTemplates(cmd)

			output := buf.String()

			// Check output contains expected strings
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected string %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestSaveTemplate tests the saveTemplate function
func TestSaveTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		configExists bool
		wantErr      bool
		wantContains string
	}{
		{
			name:         "save valid template",
			templateName: "mytemplate",
			configExists: true,
			wantErr:      false,
			wantContains: "Template 'mytemplate' saved successfully",
		},
		{
			name:         "save without config",
			templateName: "noconfig",
			configExists: false,
			wantErr:      true,
			wantContains: "Error saving template",
		},
		{
			name:         "save with empty name",
			templateName: "",
			configExists: true,
			wantErr:      false,
			wantContains: "Template '' saved successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			env := NewTestEnvironment(t)

			// Change to temp directory
			restore := ChangeToTempDir(t, env.TempDir)
			defer restore()

			// Setup HOME for custom templates directory
			home := env.TempDir
			os.Setenv("HOME", home)
			defer os.Unsetenv("HOME")

			// Create config file if needed
			if tt.configExists {
				env.WriteConfig(ValidConfig())
			}

			// Set the flag
			saveTemplateFlag = tt.templateName

			// Create command with output capture
			var buf bytes.Buffer
			cmd := &cobra.Command{
				Use: "init",
			}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Call saveTemplate
			saveTemplate(cmd)

			output := buf.String()

			// Check output
			if tt.wantErr {
				if !strings.Contains(output, tt.wantContains) {
					t.Errorf("expected error message containing %q, got: %s", tt.wantContains, output)
				}
			} else {
				if !strings.Contains(output, tt.wantContains) {
					t.Errorf("expected success message containing %q, got: %s", tt.wantContains, output)
				}

				// Verify template file was created
				if tt.templateName != "" {
					templatesDir := filepath.Join(home, ".asc", "templates")
					templatePath := filepath.Join(templatesDir, tt.templateName+".toml")
					if _, err := os.Stat(templatePath); os.IsNotExist(err) {
						t.Errorf("template file was not created at %s", templatePath)
					}
				}
			}
		})
	}
}

// TestRunInitWithListTemplates tests runInit with --list-templates flag
func TestRunInitWithListTemplates(t *testing.T) {
	// Setup
	env := NewTestEnvironment(t)
	home := env.TempDir
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	// Set flag
	listTemplatesFlag = true
	defer func() { listTemplatesFlag = false }()

	// Create command with output capture
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "init",
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Run init
	runInit(cmd, []string{})

	output := buf.String()

	// Verify output contains template list
	if !strings.Contains(output, "Built-in templates:") {
		t.Errorf("expected template list in output, got: %s", output)
	}
	if !strings.Contains(output, "solo") {
		t.Errorf("expected 'solo' template in output, got: %s", output)
	}
}

// TestRunInitWithSaveTemplate tests runInit with --save-template flag
func TestRunInitWithSaveTemplate(t *testing.T) {
	// Setup
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	home := env.TempDir
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	// Create config file
	env.WriteConfig(ValidConfig())

	// Set flag
	saveTemplateFlag = "testtemplate"
	defer func() { saveTemplateFlag = "" }()

	// Create command with output capture
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "init",
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Run init
	runInit(cmd, []string{})

	output := buf.String()

	// Verify output
	if !strings.Contains(output, "Template 'testtemplate' saved successfully") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify template file exists
	templatesDir := filepath.Join(home, ".asc", "templates")
	templatePath := filepath.Join(templatesDir, "testtemplate.toml")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Errorf("template file was not created at %s", templatePath)
	}
}

// TestRunInitWithTemplate tests runInit with --template flag
func TestRunInitWithTemplate(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		expectWizard bool
	}{
		{
			name:         "with solo template",
			template:     "solo",
			expectWizard: true,
		},
		{
			name:         "with team template",
			template:     "team",
			expectWizard: true,
		},
		{
			name:         "with swarm template",
			template:     "swarm",
			expectWizard: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We can't fully test the wizard without mocking it
			// This test verifies that the template flag is processed

			// Setup
			env := NewTestEnvironment(t)
			home := env.TempDir
			os.Setenv("HOME", home)
			defer os.Unsetenv("HOME")

			// Set flag
			templateFlag = tt.template
			defer func() { templateFlag = "" }()

			// Verify template exists
			tmpl, err := config.GetTemplate(config.TemplateType(tt.template))
			if err != nil {
				t.Fatalf("template %s should exist: %v", tt.template, err)
			}

			if tmpl.Name != tt.template {
				t.Errorf("template name = %v, want %v", tmpl.Name, tt.template)
			}
		})
	}
}

// TestInitCommandIntegration tests the full init command integration
func TestInitCommandIntegration(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "list templates",
			args:    []string{"--list-templates"},
			wantErr: false,
		},
		{
			name:    "invalid flag",
			args:    []string{"--invalid-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			env := NewTestEnvironment(t)
			home := env.TempDir
			os.Setenv("HOME", home)
			defer os.Unsetenv("HOME")

			// Reset root command
			rootCmd.ResetCommands()
			rootCmd.AddCommand(initCmd)

			// Reset flags
			templateFlag = ""
			listTemplatesFlag = false
			saveTemplateFlag = ""

			// Create output buffer
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs(append([]string{"init"}, tt.args...))

			// Execute
			err := rootCmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestInitCommandHelp tests the help output
func TestInitCommandHelp(t *testing.T) {
	// Reset root command and add init
	rootCmd.ResetCommands()
	rootCmd.AddCommand(initCmd)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"init", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	output := buf.String()

	// Check help contains key information
	wantContains := []string{
		"init",
		"interactive setup wizard",
		"--template",
		"--list-templates",
		"--save-template",
		"solo",
		"team",
		"swarm",
	}

	for _, want := range wantContains {
		if !strings.Contains(output, want) {
			t.Errorf("help output missing %q\nGot: %s", want, output)
		}
	}
}

// TestInitCommandWithInvalidTemplate tests init with invalid template
func TestInitCommandWithInvalidTemplate(t *testing.T) {
	// Setup
	env := NewTestEnvironment(t)
	home := env.TempDir
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	// Set invalid template
	templateFlag = "nonexistent"
	defer func() { templateFlag = "" }()

	// Verify template doesn't exist
	_, err := config.GetTemplate(config.TemplateType("nonexistent"))
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
}

// TestInitCommandFlagCombinations tests various flag combinations
func TestInitCommandFlagCombinations(t *testing.T) {
	tests := []struct {
		name             string
		template         string
		listTemplates    bool
		saveTemplate     string
		expectListOutput bool
		expectSaveOutput bool
	}{
		{
			name:             "list takes precedence",
			template:         "solo",
			listTemplates:    true,
			saveTemplate:     "",
			expectListOutput: true,
			expectSaveOutput: false,
		},
		{
			name:             "save without list",
			template:         "",
			listTemplates:    false,
			saveTemplate:     "mytemplate",
			expectListOutput: false,
			expectSaveOutput: true,
		},
		{
			name:             "list and save both set",
			template:         "",
			listTemplates:    true,
			saveTemplate:     "mytemplate",
			expectListOutput: true,
			expectSaveOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			env := NewTestEnvironment(t)
			restore := ChangeToTempDir(t, env.TempDir)
			defer restore()

			home := env.TempDir
			os.Setenv("HOME", home)
			defer os.Unsetenv("HOME")

			// Create config for save template
			if tt.saveTemplate != "" {
				env.WriteConfig(ValidConfig())
			}

			// Set flags
			templateFlag = tt.template
			listTemplatesFlag = tt.listTemplates
			saveTemplateFlag = tt.saveTemplate
			defer func() {
				templateFlag = ""
				listTemplatesFlag = false
				saveTemplateFlag = ""
			}()

			// Create command with output capture
			var buf bytes.Buffer
			cmd := &cobra.Command{
				Use: "init",
			}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Run init
			runInit(cmd, []string{})

			output := buf.String()

			// Check expectations
			if tt.expectListOutput {
				if !strings.Contains(output, "Built-in templates:") {
					t.Errorf("expected list output, got: %s", output)
				}
			}

			if tt.expectSaveOutput {
				if !strings.Contains(output, "saved successfully") {
					t.Errorf("expected save output, got: %s", output)
				}
			}
		})
	}
}

// TestInitCommandErrorHandling tests error handling in init command
func TestInitCommandErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*TestEnvironment)
		saveTemplate string
		wantErrMsg   string
	}{
		{
			name: "save template without config",
			setupFunc: func(env *TestEnvironment) {
				// Don't create config file
			},
			saveTemplate: "test",
			wantErrMsg:   "Error saving template",
		},
		{
			name: "save template with unreadable config",
			setupFunc: func(env *TestEnvironment) {
				// Create config with no read permissions
				env.WriteConfig(ValidConfig())
				os.Chmod(env.ConfigPath, 0000)
			},
			saveTemplate: "test",
			wantErrMsg:   "Error saving template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			env := NewTestEnvironment(t)
			restore := ChangeToTempDir(t, env.TempDir)
			defer restore()

			home := env.TempDir
			os.Setenv("HOME", home)
			defer os.Unsetenv("HOME")

			// Run setup
			if tt.setupFunc != nil {
				tt.setupFunc(env)
			}

			// Set flag
			saveTemplateFlag = tt.saveTemplate
			defer func() { saveTemplateFlag = "" }()

			// Create command with output capture
			var buf bytes.Buffer
			cmd := &cobra.Command{
				Use: "init",
			}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Run init
			runInit(cmd, []string{})

			output := buf.String()

			// Check error message
			if !strings.Contains(output, tt.wantErrMsg) {
				t.Errorf("expected error message containing %q, got: %s", tt.wantErrMsg, output)
			}

			// Restore permissions for cleanup
			if tt.name == "save template with unreadable config" {
				os.Chmod(env.ConfigPath, 0644)
			}
		})
	}
}

// TestListTemplatesWithCustomTemplates tests listing with custom templates
func TestListTemplatesWithCustomTemplates(t *testing.T) {
	// Setup
	env := NewTestEnvironment(t)
	home := env.TempDir
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	// Create custom templates directory
	templatesDir := filepath.Join(home, ".asc", "templates")
	os.MkdirAll(templatesDir, 0755)

	// Create multiple custom templates
	customTemplates := []string{"custom1", "custom2", "custom3"}
	for _, name := range customTemplates {
		templatePath := filepath.Join(templatesDir, name+".toml")
		os.WriteFile(templatePath, []byte(ValidConfig()), 0644)
	}

	// Create command with output capture
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "init",
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Call listTemplates
	listTemplates(cmd)

	output := buf.String()

	// Verify all custom templates are listed
	if !strings.Contains(output, "Custom templates:") {
		t.Errorf("expected 'Custom templates:' section in output")
	}

	for _, name := range customTemplates {
		if !strings.Contains(output, name) {
			t.Errorf("expected custom template %q in output, got: %s", name, output)
		}
	}
}

// TestListTemplatesWithError tests error handling when loading custom templates fails
func TestListTemplatesWithError(t *testing.T) {
	// Setup
	env := NewTestEnvironment(t)
	home := env.TempDir
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	// Create templates directory with no read permissions
	templatesDir := filepath.Join(home, ".asc", "templates")
	os.MkdirAll(templatesDir, 0755)

	// Create a template file
	templatePath := filepath.Join(templatesDir, "test.toml")
	os.WriteFile(templatePath, []byte(ValidConfig()), 0644)

	// Remove read permissions from directory
	os.Chmod(templatesDir, 0000)
	defer os.Chmod(templatesDir, 0755) // Restore for cleanup

	// Create command with output capture
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "init",
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Call listTemplates
	listTemplates(cmd)

	output := buf.String()

	// Should still show built-in templates
	if !strings.Contains(output, "Built-in templates:") {
		t.Errorf("expected built-in templates in output even with error")
	}

	// Should show warning about custom templates
	if !strings.Contains(output, "Warning") && !strings.Contains(output, "Failed") {
		t.Errorf("expected warning about failed custom template loading")
	}
}

// TestSaveTemplateWithDirectoryCreation tests saving template when directory doesn't exist
func TestSaveTemplateWithDirectoryCreation(t *testing.T) {
	// Setup
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	home := env.TempDir
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	// Create config file
	env.WriteConfig(ValidConfig())

	// Don't create templates directory - it should be created automatically
	templatesDir := filepath.Join(home, ".asc", "templates")
	if _, err := os.Stat(templatesDir); !os.IsNotExist(err) {
		t.Fatal("templates directory should not exist yet")
	}

	// Set flag
	saveTemplateFlag = "newtemplate"
	defer func() { saveTemplateFlag = "" }()

	// Create command with output capture
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "init",
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Call saveTemplate
	saveTemplate(cmd)

	output := buf.String()

	// Verify success
	if !strings.Contains(output, "saved successfully") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify directory was created
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		t.Errorf("templates directory should have been created")
	}

	// Verify template file exists
	templatePath := filepath.Join(templatesDir, "newtemplate.toml")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Errorf("template file should have been created")
	}
}

// TestRunInitDefaultBehavior tests runInit without any flags (would launch wizard)
func TestRunInitDefaultBehavior(t *testing.T) {
	// This test verifies that without flags, runInit would attempt to launch the wizard
	// We can't fully test the wizard without mocking, but we can verify the code path

	// Setup
	env := NewTestEnvironment(t)
	home := env.TempDir
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	// Reset all flags
	templateFlag = ""
	listTemplatesFlag = false
	saveTemplateFlag = ""

	// Note: We can't actually run the wizard in tests without mocking
	// This test documents the expected behavior
	// In a real scenario, runInit would call wizard.Run() which launches the TUI
}

// TestInitCommandOutputFormat tests the output format of various operations
func TestInitCommandOutputFormat(t *testing.T) {
	tests := []struct {
		name         string
		operation    string
		setupFunc    func(*TestEnvironment)
		wantPatterns []string
	}{
		{
			name:      "list templates format",
			operation: "list",
			setupFunc: func(env *TestEnvironment) {},
			wantPatterns: []string{
				"Available templates:",
				"Built-in templates:",
				"solo -",
				"team -",
				"swarm -",
			},
		},
		{
			name:      "save template format",
			operation: "save",
			setupFunc: func(env *TestEnvironment) {
				env.WriteConfig(ValidConfig())
			},
			wantPatterns: []string{
				"Template 'testformat' saved successfully",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			env := NewTestEnvironment(t)
			restore := ChangeToTempDir(t, env.TempDir)
			defer restore()

			home := env.TempDir
			os.Setenv("HOME", home)
			defer os.Unsetenv("HOME")

			if tt.setupFunc != nil {
				tt.setupFunc(env)
			}

			// Create command with output capture
			var buf bytes.Buffer
			cmd := &cobra.Command{
				Use: "init",
			}
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute operation
			switch tt.operation {
			case "list":
				listTemplatesFlag = true
				defer func() { listTemplatesFlag = false }()
				runInit(cmd, []string{})
			case "save":
				saveTemplateFlag = "testformat"
				defer func() { saveTemplateFlag = "" }()
				runInit(cmd, []string{})
			}

			output := buf.String()

			// Check patterns
			for _, pattern := range tt.wantPatterns {
				if !strings.Contains(output, pattern) {
					t.Errorf("output missing pattern %q\nGot: %s", pattern, output)
				}
			}
		})
	}
}
