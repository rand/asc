package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/asc/internal/check"
	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/secrets"
)

// TestWizardModel_ViewWelcome tests the welcome screen rendering
func TestWizardModel_ViewWelcome(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepWelcome

	view := m.viewWelcome()

	// Check for key elements
	if !strings.Contains(view, "Agent Stack Controller Setup") {
		t.Error("Welcome view should contain title")
	}
	if !strings.Contains(view, "Checking for required dependencies") {
		t.Error("Welcome view should mention dependency checking")
	}
	if !strings.Contains(view, "Press Enter to begin") {
		t.Error("Welcome view should show prompt")
	}
}

// TestWizardModel_ViewChecking tests the checking screen rendering
func TestWizardModel_ViewChecking(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepChecking

	view := m.viewChecking()

	if !strings.Contains(view, "Running dependency checks") {
		t.Error("Checking view should show progress message")
	}
}


// TestWizardModel_ViewCheckResults tests the check results display
func TestWizardModel_ViewCheckResults(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepCheckResults
	m.checkResults = []check.CheckResult{
		{Name: "git", Status: check.CheckPass, Message: "Found"},
		{Name: "python3", Status: check.CheckPass, Message: "Found"},
	}

	view := m.viewCheckResults()

	if !strings.Contains(view, "Dependency Check Results") {
		t.Error("Check results view should contain title")
	}
	if !strings.Contains(view, "Press Enter to continue") {
		t.Error("Check results view should show prompt")
	}
}

// TestWizardModel_ViewAPIKeys tests the API key input screen
func TestWizardModel_ViewAPIKeys(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepAPIKeys

	view := m.viewAPIKeys()

	if !strings.Contains(view, "API Key Configuration") {
		t.Error("API keys view should contain title")
	}
	if !strings.Contains(view, "Claude API Key") {
		t.Error("API keys view should show Claude input")
	}
	if !strings.Contains(view, "OpenAI API Key") {
		t.Error("API keys view should show OpenAI input")
	}
	if !strings.Contains(view, "Google API Key") {
		t.Error("API keys view should show Google input")
	}
}

// TestWizardModel_ViewGenerating tests the generating screen
func TestWizardModel_ViewGenerating(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepGenerating

	view := m.viewGenerating()

	if !strings.Contains(view, "Generating configuration files") {
		t.Error("Generating view should show progress message")
	}
}

// TestWizardModel_ViewValidating tests the validating screen
func TestWizardModel_ViewValidating(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepValidating

	view := m.viewValidating()

	if !strings.Contains(view, "Validating setup") {
		t.Error("Validating view should show progress message")
	}
}


// TestWizardModel_ViewComplete tests the completion screen
func TestWizardModel_ViewComplete(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		shouldContain []string
	}{
		{
			name: "success",
			err:  nil,
			shouldContain: []string{
				"Setup Complete",
				"asc up",
				"Press 'q' to exit",
			},
		},
		{
			name: "with error",
			err:  os.ErrNotExist,
			shouldContain: []string{
				"Setup Failed",
				"Error:",
				"Press 'q' to exit",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWizard()
			m := w.initialModel()
			m.step = stepComplete
			m.err = tt.err

			view := m.viewComplete()

			for _, expected := range tt.shouldContain {
				if !strings.Contains(view, expected) {
					t.Errorf("Complete view should contain %q", expected)
				}
			}
		})
	}
}

// TestWizardModel_ViewAgeSetup tests the age setup screen
func TestWizardModel_ViewAgeSetup(t *testing.T) {
	tests := []struct {
		name          string
		ageInstalled  bool
		needsAgeSetup bool
		shouldContain []string
	}{
		{
			name:          "age not installed",
			ageInstalled:  false,
			needsAgeSetup: false,
			shouldContain: []string{
				"Secure Secrets Management",
				"age encryption is not installed",
				"Install age now?",
			},
		},
		{
			name:          "age installed needs setup",
			ageInstalled:  true,
			needsAgeSetup: true,
			shouldContain: []string{
				"Secure Secrets Management",
				"age is installed",
				"Set up encryption?",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWizard()
			m := w.initialModel()
			m.step = stepAgeSetup
			m.ageInstalled = tt.ageInstalled
			m.needsAgeSetup = tt.needsAgeSetup

			view := m.viewAgeSetup()

			for _, expected := range tt.shouldContain {
				if !strings.Contains(view, expected) {
					t.Errorf("Age setup view should contain %q", expected)
				}
			}
		})
	}
}


// TestWizardModel_ViewEncrypting tests the encrypting screen
func TestWizardModel_ViewEncrypting(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepEncrypting

	view := m.viewEncrypting()

	if !strings.Contains(view, "Encrypting secrets") {
		t.Error("Encrypting view should show progress message")
	}
}

// TestWizardModel_ViewBackupPrompt tests the backup prompt screen
func TestWizardModel_ViewBackupPrompt(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepBackupPrompt

	view := m.viewBackupPrompt()

	if !strings.Contains(view, "Existing configuration files detected") {
		t.Error("Backup prompt should mention existing files")
	}
	if !strings.Contains(view, "Create backup and continue?") {
		t.Error("Backup prompt should ask for confirmation")
	}
}

// TestWizardModel_ViewInstallPrompt tests the install prompt screen
func TestWizardModel_ViewInstallPrompt(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepInstallPrompt

	view := m.viewInstallPrompt()

	if !strings.Contains(view, "Some dependencies are missing") {
		t.Error("Install prompt should mention missing dependencies")
	}
	if !strings.Contains(view, "Do you want to exit and install dependencies?") {
		t.Error("Install prompt should ask for confirmation")
	}
}

// TestWizardModel_ViewTemplateSelection tests the template selection screen
func TestWizardModel_ViewTemplateSelection(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepTemplateSelection
	m.templates = []config.Template{
		{Name: "solo", Description: "Single agent setup"},
		{Name: "team", Description: "Team of agents"},
	}
	m.selectedTemplate = 0

	view := m.viewTemplateSelection()

	if !strings.Contains(view, "Select Configuration Template") {
		t.Error("Template selection should contain title")
	}
	if !strings.Contains(view, "solo") {
		t.Error("Template selection should show solo template")
	}
	if !strings.Contains(view, "team") {
		t.Error("Template selection should show team template")
	}
}


// TestRunChecks tests the runChecks function
func TestRunChecks(t *testing.T) {
	checker := check.NewChecker("asc.toml", ".env")
	cmd := runChecks(checker)

	if cmd == nil {
		t.Fatal("runChecks should return a command")
	}

	// Execute the command
	msg := cmd()

	// Check that we got a checkCompleteMsg
	if _, ok := msg.(checkCompleteMsg); !ok {
		t.Errorf("Expected checkCompleteMsg, got %T", msg)
	}
}

// TestGenerateConfigFiles tests the generateConfigFiles function
func TestGenerateConfigFiles(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	apiKeys := map[string]string{
		"CLAUDE_API_KEY": "sk-ant-test123",
		"OPENAI_API_KEY": "sk-test456",
		"GOOGLE_API_KEY": "AIzatest789",
	}

	cmd := generateConfigFiles(apiKeys, "team")
	if cmd == nil {
		t.Fatal("generateConfigFiles should return a command")
	}

	// Execute the command
	msg := cmd()

	// Check that we got a generateCompleteMsg
	genMsg, ok := msg.(generateCompleteMsg)
	if !ok {
		t.Fatalf("Expected generateCompleteMsg, got %T", msg)
	}

	// Check for errors
	if genMsg.err != nil {
		t.Errorf("generateConfigFiles should not return error: %v", genMsg.err)
	}

	// Verify .env file was created
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		t.Error(".env file should be created")
	}

	// Verify .env contains API keys
	envContent, _ := os.ReadFile(".env")
	envStr := string(envContent)
	if !strings.Contains(envStr, "CLAUDE_API_KEY=sk-ant-test123") {
		t.Error(".env should contain Claude API key")
	}
	if !strings.Contains(envStr, "OPENAI_API_KEY=sk-test456") {
		t.Error(".env should contain OpenAI API key")
	}
	if !strings.Contains(envStr, "GOOGLE_API_KEY=AIzatest789") {
		t.Error(".env should contain Google API key")
	}
}


// TestRunValidation tests the runValidation function
func TestRunValidation(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Create minimal config files
	os.WriteFile("asc.toml", []byte(`
[core]
beads_db_path = "./test"

[services.mcp_agent_mail]
start_command = "test"
url = "http://localhost:8765"
`), 0644)

	os.WriteFile(".env", []byte("CLAUDE_API_KEY=sk-ant-test\n"), 0644)

	cmd := runValidation()
	if cmd == nil {
		t.Fatal("runValidation should return a command")
	}

	// Execute the command
	msg := cmd()

	// Check that we got a validateCompleteMsg
	if _, ok := msg.(validateCompleteMsg); !ok {
		t.Errorf("Expected validateCompleteMsg, got %T", msg)
	}
}

// TestBackupConfigFiles tests the backupConfigFiles function
func TestBackupConfigFiles(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Create test files
	os.WriteFile("asc.toml", []byte("test config"), 0644)
	os.WriteFile(".env", []byte("test env"), 0644)

	// Override home directory for test
	home := filepath.Join(tmpDir, "home")
	os.Setenv("HOME", home)
	defer os.Unsetenv("HOME")

	err := backupConfigFiles()
	if err != nil {
		t.Fatalf("backupConfigFiles failed: %v", err)
	}

	// Check backup directory was created
	backupDir := filepath.Join(home, ".asc_backup")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		t.Error("Backup directory should be created")
	}

	// Check backup files exist
	files, _ := os.ReadDir(backupDir)
	if len(files) != 2 {
		t.Errorf("Expected 2 backup files, got %d", len(files))
	}
}


// TestValidateAPIKey tests the validateAPIKey function
func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		keyName string
		value   string
		want    bool
	}{
		{
			name:    "valid Claude key",
			keyName: "CLAUDE_API_KEY",
			value:   "sk-ant-test123",
			want:    true,
		},
		{
			name:    "invalid Claude key",
			keyName: "CLAUDE_API_KEY",
			value:   "invalid",
			want:    false,
		},
		{
			name:    "valid OpenAI key",
			keyName: "OPENAI_API_KEY",
			value:   "sk-test456",
			want:    true,
		},
		{
			name:    "invalid OpenAI key",
			keyName: "OPENAI_API_KEY",
			value:   "invalid",
			want:    false,
		},
		{
			name:    "valid Google key",
			keyName: "GOOGLE_API_KEY",
			value:   "AIzatest789",
			want:    true,
		},
		{
			name:    "invalid Google key",
			keyName: "GOOGLE_API_KEY",
			value:   "invalid",
			want:    false,
		},
		{
			name:    "empty key",
			keyName: "CLAUDE_API_KEY",
			value:   "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateAPIKey(tt.keyName, tt.value)
			if got != tt.want {
				t.Errorf("validateAPIKey(%q, %q) = %v, want %v", tt.keyName, tt.value, got, tt.want)
			}
		})
	}
}

// TestGenerateConfigFromTemplate tests the generateConfigFromTemplate function
func TestGenerateConfigFromTemplate(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	tests := []struct {
		name         string
		templateName string
		wantErr      bool
	}{
		{
			name:         "solo template",
			templateName: "solo",
			wantErr:      false,
		},
		{
			name:         "team template",
			templateName: "team",
			wantErr:      false,
		},
		{
			name:         "swarm template",
			templateName: "swarm",
			wantErr:      false,
		},
		{
			name:         "default template",
			templateName: "",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generateConfigFromTemplate(tt.templateName)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateConfigFromTemplate(%q) error = %v, wantErr %v", tt.templateName, err, tt.wantErr)
			}

			// Check that asc.toml was created
			if !tt.wantErr {
				if _, err := os.Stat("asc.toml"); os.IsNotExist(err) {
					t.Error("asc.toml should be created")
				}
				// Clean up for next test
				os.Remove("asc.toml")
			}
		})
	}
}


// TestWizardModel_HandleEnter tests the handleEnter function for different steps
func TestWizardModel_HandleEnter(t *testing.T) {
	tests := []struct {
		name         string
		initialStep  wizardStep
		templateName string
		confirmed    bool
		expectedStep wizardStep
	}{
		{
			name:         "welcome to template selection",
			initialStep:  stepWelcome,
			templateName: "",
			expectedStep: stepTemplateSelection,
		},
		{
			name:         "welcome to checking with template",
			initialStep:  stepWelcome,
			templateName: "team",
			expectedStep: stepChecking,
		},
		{
			name:         "template selection to checking",
			initialStep:  stepTemplateSelection,
			templateName: "",
			expectedStep: stepChecking,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWizard()
			m := w.initialModel()
			m.step = tt.initialStep
			m.templateName = tt.templateName
			m.confirmed = tt.confirmed

			newModel, _ := m.handleEnter()
			newM := newModel.(wizardModel)

			if newM.step != tt.expectedStep {
				t.Errorf("Expected step %v, got %v", tt.expectedStep, newM.step)
			}
		})
	}
}

// TestWizardModel_HandleTab tests the tab navigation in API keys
func TestWizardModel_HandleTab(t *testing.T) {
	tests := []struct {
		name          string
		initialActive int
		reverse       bool
		expectedActive int
	}{
		{
			name:          "forward from first",
			initialActive: 0,
			reverse:       false,
			expectedActive: 1,
		},
		{
			name:          "forward from second",
			initialActive: 1,
			reverse:       false,
			expectedActive: 2,
		},
		{
			name:          "forward from last wraps",
			initialActive: 2,
			reverse:       false,
			expectedActive: 0,
		},
		{
			name:          "backward from last",
			initialActive: 2,
			reverse:       true,
			expectedActive: 1,
		},
		{
			name:          "backward from first wraps",
			initialActive: 0,
			reverse:       true,
			expectedActive: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWizard()
			m := w.initialModel()
			m.activeInput = tt.initialActive

			newModel, _ := m.handleTab(tt.reverse)
			newM := newModel.(wizardModel)

			if newM.activeInput != tt.expectedActive {
				t.Errorf("Expected activeInput %d, got %d", tt.expectedActive, newM.activeInput)
			}
		})
	}
}


// TestWizardModel_HandleTemplateNav tests template navigation
func TestWizardModel_HandleTemplateNav(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.templates = []config.Template{
		{Name: "solo"},
		{Name: "team"},
		{Name: "swarm"},
	}
	m.customTemplates = []config.Template{} // Clear custom templates for predictable test
	m.selectedTemplate = 1

	// Navigate down
	newModel, _ := m.handleTemplateNav(1)
	m = newModel.(wizardModel)
	if m.selectedTemplate != 2 {
		t.Errorf("Expected selectedTemplate 2, got %d", m.selectedTemplate)
	}

	// Navigate down again (should wrap to 0)
	newModel, _ = m.handleTemplateNav(1)
	m = newModel.(wizardModel)
	if m.selectedTemplate != 0 {
		t.Errorf("Expected selectedTemplate 0 (wrapped), got %d", m.selectedTemplate)
	}

	// Navigate up from 0 (should wrap to 2)
	newModel, _ = m.handleTemplateNav(-1)
	m = newModel.(wizardModel)
	if m.selectedTemplate != 2 {
		t.Errorf("Expected selectedTemplate 2 (wrapped), got %d", m.selectedTemplate)
	}
}

// TestWizardModel_HandleTemplateNumber tests quick template selection by number
func TestWizardModel_HandleTemplateNumber(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepTemplateSelection
	m.templates = []config.Template{
		{Name: "solo"},
		{Name: "team"},
		{Name: "swarm"},
	}

	// Select template 2 (index 1)
	newModel, _ := m.handleTemplateNumber("2")
	newM := newModel.(wizardModel)

	if newM.selectedTemplate != 1 {
		t.Errorf("Expected selectedTemplate 1, got %d", newM.selectedTemplate)
	}

	// Should advance to checking step
	if newM.step != stepChecking {
		t.Errorf("Expected step checking, got %v", newM.step)
	}
}

// TestWizardModel_Update tests the Update function with various messages
func TestWizardModel_Update(t *testing.T) {
	tests := []struct {
		name         string
		initialStep  wizardStep
		msg          tea.Msg
		expectedStep wizardStep
	}{
		{
			name:        "window resize",
			initialStep: stepWelcome,
			msg:         tea.WindowSizeMsg{Width: 100, Height: 50},
			expectedStep: stepWelcome,
		},
		{
			name:        "check complete",
			initialStep: stepChecking,
			msg: checkCompleteMsg{
				results: []check.CheckResult{
					{Name: "git", Status: check.CheckPass},
				},
			},
			expectedStep: stepCheckResults,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWizard()
			m := w.initialModel()
			m.step = tt.initialStep

			newModel, _ := m.Update(tt.msg)
			newM := newModel.(wizardModel)

			if newM.step != tt.expectedStep {
				t.Errorf("Expected step %v, got %v", tt.expectedStep, newM.step)
			}
		})
	}
}


// TestWizardModel_UpdateInputs tests text input handling
func TestWizardModel_UpdateInputs(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()
	m.step = stepAPIKeys
	m.activeInput = 0

	// Simulate typing
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	newModel, _ := m.updateInputs(msg)
	newM := newModel.(wizardModel)

	// Check that input was updated
	if newM.claudeInput.Value() != "t" {
		t.Errorf("Expected claude input to be 't', got %q", newM.claudeInput.Value())
	}
}

// TestWizard_SetTemplate tests the SetTemplate method
func TestWizard_SetTemplate(t *testing.T) {
	w := NewWizard()
	w.SetTemplate("solo")

	if w.templateName != "solo" {
		t.Errorf("Expected templateName 'solo', got %q", w.templateName)
	}
}

// TestWizard_InitialModel tests the initialModel method
func TestWizard_InitialModel(t *testing.T) {
	w := NewWizard()
	w.SetTemplate("team")
	m := w.initialModel()

	if m.step != stepWelcome {
		t.Errorf("Expected initial step to be welcome, got %v", m.step)
	}

	if m.templateName != "team" {
		t.Errorf("Expected templateName 'team', got %q", m.templateName)
	}

	if m.activeInput != 0 {
		t.Error("Expected activeInput to be 0")
	}

	if m.apiKeys == nil {
		t.Error("Expected apiKeys map to be initialized")
	}
}

// TestFileExists tests the fileExists helper function
func TestFileExists(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Test non-existent file
	if fileExists("nonexistent.txt") {
		t.Error("fileExists should return false for non-existent file")
	}

	// Create a file
	os.WriteFile("test.txt", []byte("test"), 0644)

	// Test existing file
	if !fileExists("test.txt") {
		t.Error("fileExists should return true for existing file")
	}
}

// TestCopyFile tests the copyFile helper function
func TestCopyFile(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Create source file
	content := []byte("test content")
	os.WriteFile("source.txt", content, 0644)

	// Copy file
	err := copyFile("source.txt", "dest.txt")
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify destination file
	destContent, err := os.ReadFile("dest.txt")
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(destContent) != string(content) {
		t.Errorf("Expected content %q, got %q", content, destContent)
	}
}


// TestGenerateEnvFile tests the generateEnvFile function
func TestGenerateEnvFile(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	apiKeys := map[string]string{
		"CLAUDE_API_KEY": "sk-ant-test123",
		"OPENAI_API_KEY": "sk-test456",
		"GOOGLE_API_KEY": "AIzatest789",
	}

	err := generateEnvFile(apiKeys)
	if err != nil {
		t.Fatalf("generateEnvFile failed: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		t.Fatal(".env file should be created")
	}

	// Check file permissions
	info, _ := os.Stat(".env")
	mode := info.Mode()
	if mode.Perm() != 0600 {
		t.Errorf("Expected .env permissions 0600, got %o", mode.Perm())
	}

	// Check content
	content, _ := os.ReadFile(".env")
	contentStr := string(content)

	for key, value := range apiKeys {
		expected := key + "=" + value
		if !strings.Contains(contentStr, expected) {
			t.Errorf(".env should contain %q", expected)
		}
	}
}

// TestGenerateDefaultConfig tests the generateDefaultConfig function
func TestGenerateDefaultConfig(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	err := generateDefaultConfig()
	if err != nil {
		t.Fatalf("generateDefaultConfig failed: %v", err)
	}

	// Check that asc.toml was created
	if _, err := os.Stat("asc.toml"); os.IsNotExist(err) {
		t.Error("asc.toml should be created")
	}
}

// TestEncryptSecrets tests the encryptSecrets command
func TestEncryptSecrets(t *testing.T) {
	// Skip if age is not installed
	manager := secrets.NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping encryption test")
	}

	// Create temp directory for test
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create .env file
	os.WriteFile(".env", []byte("TEST_KEY=test_value\n"), 0600)

	// Test encrypt command
	cmd := encryptSecrets(manager)
	if cmd == nil {
		t.Fatal("encryptSecrets should return a command")
	}

	msg := cmd()
	encMsg, ok := msg.(encryptCompleteMsg)
	if !ok {
		t.Fatalf("Expected encryptCompleteMsg, got %T", msg)
	}

	if encMsg.err != nil {
		t.Errorf("Encryption should not fail: %v", encMsg.err)
	}

	// Check that .env.age was created
	if _, err := os.Stat(".env.age"); os.IsNotExist(err) {
		t.Error(".env.age should be created")
	}
}

// TestWizardModel_Init tests the Init method
func TestWizardModel_Init(t *testing.T) {
	w := NewWizard()
	m := w.initialModel()

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

