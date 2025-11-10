package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/asc/internal/check"
	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/secrets"
)

// wizardStep represents the current step in the wizard
type wizardStep int

const (
	stepWelcome wizardStep = iota
	stepTemplateSelection
	stepChecking
	stepCheckResults
	stepInstallPrompt
	stepAgeSetup
	stepBackupPrompt
	stepAPIKeys
	stepGenerating
	stepEncrypting
	stepValidating
	stepComplete
)

// Wizard manages the interactive setup process
type Wizard struct {
	checker        check.Checker
	checkResults   []check.CheckResult
	secretsManager *secrets.Manager
	templateName   string
}

// wizardModel is the bubbletea model for the wizard
type wizardModel struct {
	step           wizardStep
	checker        check.Checker
	checkResults   []check.CheckResult
	secretsManager *secrets.Manager
	width          int
	height         int
	templateName   string
	
	// Template selection
	templates       []config.Template
	customTemplates []config.Template
	selectedTemplate int
	
	// API key inputs
	claudeInput  textinput.Model
	openaiInput  textinput.Model
	googleInput  textinput.Model
	activeInput  int
	apiKeys      map[string]string
	
	// State flags
	needsInstall   bool
	needsAgeSetup  bool
	ageInstalled   bool
	needsBackup    bool
	confirmed      bool
	encryptSecrets bool
	err            error
}

// NewWizard creates a new setup wizard
func NewWizard() *Wizard {
	return &Wizard{
		checker:        check.NewChecker("asc.toml", ".env"),
		secretsManager: secrets.NewManager(),
	}
}

// SetTemplate sets the template to use for configuration generation
func (w *Wizard) SetTemplate(templateName string) {
	w.templateName = templateName
}

// Run starts the wizard
func (w *Wizard) Run() error {
	// Initialize the model
	m := w.initialModel()
	
	// Create the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	
	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	
	// Check for errors in final model
	if fm, ok := finalModel.(wizardModel); ok {
		return fm.err
	}
	
	return nil
}

func (w *Wizard) initialModel() wizardModel {
	// Initialize API key inputs
	claudeInput := textinput.New()
	claudeInput.Placeholder = "sk-ant-..."
	claudeInput.Focus()
	claudeInput.CharLimit = 200
	claudeInput.Width = 50
	claudeInput.EchoMode = textinput.EchoPassword
	claudeInput.EchoCharacter = '•'
	
	openaiInput := textinput.New()
	openaiInput.Placeholder = "sk-..."
	openaiInput.CharLimit = 200
	openaiInput.Width = 50
	openaiInput.EchoMode = textinput.EchoPassword
	openaiInput.EchoCharacter = '•'
	
	googleInput := textinput.New()
	googleInput.Placeholder = "AIza..."
	googleInput.CharLimit = 200
	googleInput.Width = 50
	googleInput.EchoMode = textinput.EchoPassword
	googleInput.EchoCharacter = '•'
	
	// Load templates
	templates := config.ListTemplates()
	customTemplates, _ := config.ListCustomTemplates()
	
	return wizardModel{
		step:            stepWelcome,
		checker:         w.checker,
		templateName:    w.templateName,
		templates:       templates,
		customTemplates: customTemplates,
		selectedTemplate: 0,
		claudeInput:     claudeInput,
		openaiInput:     openaiInput,
		googleInput:     googleInput,
		activeInput:     0,
		apiKeys:         make(map[string]string),
	}
}

func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step == stepWelcome || m.step == stepComplete {
				return m, tea.Quit
			}
		case "enter":
			return m.handleEnter()
		case "tab", "shift+tab":
			if m.step == stepAPIKeys {
				return m.handleTab(msg.String() == "shift+tab")
			}
		case "up", "k":
			if m.step == stepTemplateSelection {
				return m.handleTemplateNav(-1)
			}
		case "down", "j":
			if m.step == stepTemplateSelection {
				return m.handleTemplateNav(1)
			}
		case "y", "Y":
			if m.step == stepInstallPrompt || m.step == stepBackupPrompt {
				m.confirmed = true
				return m.handleEnter()
			}
		case "n", "N":
			if m.step == stepInstallPrompt || m.step == stepBackupPrompt {
				m.confirmed = false
				return m.handleEnter()
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.step == stepTemplateSelection {
				return m.handleTemplateNumber(msg.String())
			}
		}
		
		// Handle text input
		if m.step == stepAPIKeys {
			return m.updateInputs(msg)
		}
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case checkCompleteMsg:
		m.checkResults = msg.results
		m.step = stepCheckResults
		return m, nil
		
	case generateCompleteMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		// After generating, check if we should encrypt
		if m.encryptSecrets {
			m.step = stepEncrypting
			return m, encryptSecrets(m.secretsManager)
		}
		m.step = stepValidating
		return m, runValidation()
		
	case encryptCompleteMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.step = stepValidating
		return m, runValidation()
		
	case validateCompleteMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.step = stepComplete
		return m, nil
	}
	
	return m, nil
}

func (m wizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepWelcome:
		// If template was specified via flag, skip selection
		if m.templateName != "" {
			m.step = stepChecking
			return m, runChecks(m.checker)
		}
		m.step = stepTemplateSelection
		return m, nil
		
	case stepTemplateSelection:
		// Store selected template name
		allTemplates := append(m.templates, m.customTemplates...)
		if m.selectedTemplate >= 0 && m.selectedTemplate < len(allTemplates) {
			m.templateName = allTemplates[m.selectedTemplate].Name
		}
		m.step = stepChecking
		return m, runChecks(m.checker)
		
	case stepCheckResults:
		// Check if we need to install anything
		hasFailures := check.HasFailures(m.checkResults)
		if hasFailures {
			m.needsInstall = true
			m.step = stepInstallPrompt
		} else {
			// Check age installation and setup
			m.ageInstalled = m.secretsManager.IsAgeInstalled()
			m.needsAgeSetup = !m.secretsManager.KeyExists()
			
			if !m.ageInstalled || m.needsAgeSetup {
				m.step = stepAgeSetup
			} else {
				// Check if config files exist
				m.needsBackup = fileExists("asc.toml") || fileExists(".env")
				if m.needsBackup {
					m.step = stepBackupPrompt
				} else {
					m.step = stepAPIKeys
					m.claudeInput.Focus()
				}
			}
		}
		return m, nil
		
	case stepInstallPrompt:
		if m.confirmed {
			// User wants to install - show message and exit
			fmt.Println("\nPlease install the missing dependencies and run 'asc init' again.")
			return m, tea.Quit
		}
		// User doesn't want to install - check age setup
		m.ageInstalled = m.secretsManager.IsAgeInstalled()
		m.needsAgeSetup = !m.secretsManager.KeyExists()
		
		if !m.ageInstalled || m.needsAgeSetup {
			m.step = stepAgeSetup
		} else {
			// Check for backup
			m.needsBackup = fileExists("asc.toml") || fileExists(".env")
			if m.needsBackup {
				m.step = stepBackupPrompt
			} else {
				m.step = stepAPIKeys
				m.claudeInput.Focus()
			}
		}
		return m, nil
		
	case stepAgeSetup:
		if !m.ageInstalled {
			if m.confirmed {
				// Show installation instructions and exit
				fmt.Println("\nInstall age encryption:")
				fmt.Println("  macOS:   brew install age")
				fmt.Println("  Linux:   apt install age")
				fmt.Println("  Windows: scoop install age")
				fmt.Println("\nThen run 'asc init' again.")
				return m, tea.Quit
			}
			// Skip age setup
			m.encryptSecrets = false
		} else if m.needsAgeSetup {
			if m.confirmed {
				// Generate age key
				if err := m.secretsManager.GenerateKey(); err != nil {
					m.err = fmt.Errorf("failed to generate age key: %w", err)
					return m, tea.Quit
				}
				m.encryptSecrets = true
			} else {
				// Skip encryption
				m.encryptSecrets = false
			}
		}
		
		// Continue to backup check
		m.needsBackup = fileExists("asc.toml") || fileExists(".env")
		if m.needsBackup {
			m.step = stepBackupPrompt
		} else {
			m.step = stepAPIKeys
			m.claudeInput.Focus()
		}
		return m, nil
		
	case stepBackupPrompt:
		if m.confirmed {
			// Backup existing files
			if err := backupConfigFiles(); err != nil {
				m.err = fmt.Errorf("failed to backup files: %w", err)
				return m, tea.Quit
			}
		}
		m.step = stepAPIKeys
		m.claudeInput.Focus()
		return m, nil
		
	case stepAPIKeys:
		// Validate and save API keys
		m.apiKeys["CLAUDE_API_KEY"] = m.claudeInput.Value()
		m.apiKeys["OPENAI_API_KEY"] = m.openaiInput.Value()
		m.apiKeys["GOOGLE_API_KEY"] = m.googleInput.Value()
		
		// Basic validation
		if !validateAPIKey("CLAUDE_API_KEY", m.apiKeys["CLAUDE_API_KEY"]) ||
			!validateAPIKey("OPENAI_API_KEY", m.apiKeys["OPENAI_API_KEY"]) ||
			!validateAPIKey("GOOGLE_API_KEY", m.apiKeys["GOOGLE_API_KEY"]) {
			// Stay on this step if validation fails
			return m, nil
		}
		
		m.step = stepGenerating
		return m, generateConfigFiles(m.apiKeys, m.templateName)
		
	case stepGenerating:
		// After generating config, encrypt if enabled
		if m.encryptSecrets {
			m.step = stepEncrypting
			return m, encryptSecrets(m.secretsManager)
		}
		m.step = stepValidating
		return m, runValidation()
		
	case stepEncrypting:
		m.step = stepValidating
		return m, runValidation()
	}
	
	return m, nil
}

func (m wizardModel) handleTab(reverse bool) (tea.Model, tea.Cmd) {
	inputs := []*textinput.Model{&m.claudeInput, &m.openaiInput, &m.googleInput}
	
	if reverse {
		m.activeInput--
		if m.activeInput < 0 {
			m.activeInput = len(inputs) - 1
		}
	} else {
		m.activeInput++
		if m.activeInput >= len(inputs) {
			m.activeInput = 0
		}
	}
	
	// Update focus
	for i, input := range inputs {
		if i == m.activeInput {
			input.Focus()
		} else {
			input.Blur()
		}
	}
	
	return m, nil
}

func (m wizardModel) handleTemplateNav(delta int) (tea.Model, tea.Cmd) {
	allTemplates := append(m.templates, m.customTemplates...)
	m.selectedTemplate += delta
	
	if m.selectedTemplate < 0 {
		m.selectedTemplate = len(allTemplates) - 1
	} else if m.selectedTemplate >= len(allTemplates) {
		m.selectedTemplate = 0
	}
	
	return m, nil
}

func (m wizardModel) handleTemplateNumber(key string) (tea.Model, tea.Cmd) {
	allTemplates := append(m.templates, m.customTemplates...)
	num := int(key[0] - '1') // Convert '1'-'9' to 0-8
	
	if num >= 0 && num < len(allTemplates) {
		m.selectedTemplate = num
		return m.handleEnter()
	}
	
	return m, nil
}

func (m wizardModel) updateInputs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch m.activeInput {
	case 0:
		m.claudeInput, cmd = m.claudeInput.Update(msg)
	case 1:
		m.openaiInput, cmd = m.openaiInput.Update(msg)
	case 2:
		m.googleInput, cmd = m.googleInput.Update(msg)
	}
	
	return m, cmd
}

func (m wizardModel) View() string {
	switch m.step {
	case stepWelcome:
		return m.viewWelcome()
	case stepTemplateSelection:
		return m.viewTemplateSelection()
	case stepChecking:
		return m.viewChecking()
	case stepCheckResults:
		return m.viewCheckResults()
	case stepInstallPrompt:
		return m.viewInstallPrompt()
	case stepAgeSetup:
		return m.viewAgeSetup()
	case stepBackupPrompt:
		return m.viewBackupPrompt()
	case stepAPIKeys:
		return m.viewAPIKeys()
	case stepGenerating:
		return m.viewGenerating()
	case stepEncrypting:
		return m.viewEncrypting()
	case stepValidating:
		return m.viewValidating()
	case stepComplete:
		return m.viewComplete()
	}
	return ""
}

func (m wizardModel) viewWelcome() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginTop(2).
		MarginBottom(1)
	
	textStyle := lipgloss.NewStyle().
		MarginBottom(1)
	
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true).
		MarginTop(2)
	
	var s strings.Builder
	s.WriteString(titleStyle.Render("🚀 Agent Stack Controller Setup"))
	s.WriteString("\n\n")
	s.WriteString(textStyle.Render("Welcome to the asc initialization wizard!"))
	s.WriteString("\n")
	s.WriteString(textStyle.Render("This wizard will guide you through:"))
	s.WriteString("\n\n")
	s.WriteString("  • Checking for required dependencies\n")
	s.WriteString("  • Setting up secure secrets encryption\n")
	s.WriteString("  • Configuring API keys securely\n")
	s.WriteString("  • Generating default configuration files\n")
	s.WriteString("  • Validating your setup\n")
	s.WriteString("\n")
	s.WriteString(promptStyle.Render("Press Enter to begin, or 'q' to quit"))
	
	return s.String()
}

func (m wizardModel) viewTemplateSelection() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginTop(2).
		MarginBottom(1)
	
	textStyle := lipgloss.NewStyle().
		MarginBottom(1)
	
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true).
		Background(lipgloss.Color("236"))
	
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))
	
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(2)
	
	var s strings.Builder
	s.WriteString(titleStyle.Render("📋 Select Configuration Template"))
	s.WriteString("\n\n")
	s.WriteString(textStyle.Render("Choose a template for your agent setup:"))
	s.WriteString("\n\n")
	
	// Built-in templates
	if len(m.templates) > 0 {
		s.WriteString(lipgloss.NewStyle().Bold(true).Render("Built-in Templates:"))
		s.WriteString("\n")
		for i, tmpl := range m.templates {
			prefix := "  "
			if i == m.selectedTemplate {
				prefix = "▶ "
				s.WriteString(selectedStyle.Render(fmt.Sprintf("%s%d. %s", prefix, i+1, tmpl.Name)))
			} else {
				s.WriteString(normalStyle.Render(fmt.Sprintf("%s%d. %s", prefix, i+1, tmpl.Name)))
			}
			s.WriteString("\n")
			s.WriteString(descStyle.Render(fmt.Sprintf("     %s", tmpl.Description)))
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}
	
	// Custom templates
	if len(m.customTemplates) > 0 {
		s.WriteString(lipgloss.NewStyle().Bold(true).Render("Custom Templates:"))
		s.WriteString("\n")
		offset := len(m.templates)
		for i, tmpl := range m.customTemplates {
			idx := offset + i
			prefix := "  "
			if idx == m.selectedTemplate {
				prefix = "▶ "
				s.WriteString(selectedStyle.Render(fmt.Sprintf("%s%d. %s", prefix, idx+1, tmpl.Name)))
			} else {
				s.WriteString(normalStyle.Render(fmt.Sprintf("%s%d. %s", prefix, idx+1, tmpl.Name)))
			}
			s.WriteString("\n")
			s.WriteString(descStyle.Render(fmt.Sprintf("     %s", tmpl.Description)))
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}
	
	s.WriteString(helpStyle.Render("↑/↓ or j/k: Navigate | 1-9: Quick select | Enter: Confirm | q: Quit"))
	
	return s.String()
}

func (m wizardModel) viewChecking() string {
	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		MarginTop(2)
	
	return spinnerStyle.Render("⟳ Running dependency checks...")
}

func (m wizardModel) viewCheckResults() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		MarginTop(2).
		MarginBottom(1)
	
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true).
		MarginTop(2)
	
	var s strings.Builder
	s.WriteString(titleStyle.Render("Dependency Check Results"))
	s.WriteString("\n\n")
	s.WriteString(check.FormatResults(m.checkResults))
	s.WriteString("\n")
	s.WriteString(promptStyle.Render("Press Enter to continue"))
	
	return s.String()
}

func (m wizardModel) viewInstallPrompt() string {
	promptStyle := lipgloss.NewStyle().
		MarginTop(2).
		MarginBottom(1)
	
	questionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true)
	
	var s strings.Builder
	s.WriteString(promptStyle.Render("Some dependencies are missing."))
	s.WriteString("\n\n")
	s.WriteString("Please install the missing components before continuing.\n")
	s.WriteString("You can run 'asc init' again after installation.\n")
	s.WriteString("\n")
	s.WriteString(questionStyle.Render("Do you want to exit and install dependencies? (y/n)"))
	
	return s.String()
}

func (m wizardModel) viewBackupPrompt() string {
	promptStyle := lipgloss.NewStyle().
		MarginTop(2).
		MarginBottom(1)
	
	questionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true)
	
	var s strings.Builder
	s.WriteString(promptStyle.Render("Existing configuration files detected."))
	s.WriteString("\n\n")
	s.WriteString("The following files will be backed up to ~/.asc_backup:\n")
	if fileExists("asc.toml") {
		s.WriteString("  • asc.toml\n")
	}
	if fileExists(".env") {
		s.WriteString("  • .env\n")
	}
	s.WriteString("\n")
	s.WriteString(questionStyle.Render("Create backup and continue? (y/n)"))
	
	return s.String()
}

func (m wizardModel) viewAPIKeys() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		MarginTop(2).
		MarginBottom(1)
	
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true)
	
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(2)
	
	var s strings.Builder
	s.WriteString(titleStyle.Render("API Key Configuration"))
	s.WriteString("\n\n")
	s.WriteString("Please enter your API keys (input is masked):\n\n")
	
	s.WriteString(labelStyle.Render("Claude API Key:"))
	s.WriteString("\n")
	s.WriteString(m.claudeInput.View())
	s.WriteString("\n\n")
	
	s.WriteString(labelStyle.Render("OpenAI API Key:"))
	s.WriteString("\n")
	s.WriteString(m.openaiInput.View())
	s.WriteString("\n\n")
	
	s.WriteString(labelStyle.Render("Google API Key:"))
	s.WriteString("\n")
	s.WriteString(m.googleInput.View())
	s.WriteString("\n")
	
	s.WriteString(helpStyle.Render("Tab: Next field | Shift+Tab: Previous field | Enter: Continue"))
	
	return s.String()
}

func (m wizardModel) viewGenerating() string {
	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		MarginTop(2)
	
	return spinnerStyle.Render("⟳ Generating configuration files...")
}

func (m wizardModel) viewValidating() string {
	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		MarginTop(2)
	
	return spinnerStyle.Render("⟳ Validating setup...")
}

func (m wizardModel) viewComplete() string {
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true).
		MarginTop(2).
		MarginBottom(1)
	
	textStyle := lipgloss.NewStyle().
		MarginBottom(1)
	
	var s strings.Builder
	
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)
		
		s.WriteString(errorStyle.Render("✗ Setup Failed"))
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		s.WriteString("\n")
		s.WriteString("Please check the error and try again.\n")
	} else {
		s.WriteString(successStyle.Render("✓ Setup Complete!"))
		s.WriteString("\n\n")
		s.WriteString(textStyle.Render("Your agent stack is ready to use."))
		s.WriteString("\n\n")
		s.WriteString("Next steps:\n")
		s.WriteString("  • Run 'asc up' to start your agent colony\n")
		s.WriteString("  • Edit asc.toml to customize agent configurations\n")
		s.WriteString("  • Run 'asc check' to verify your setup anytime\n")
	}
	
	s.WriteString("\n")
	s.WriteString("Press 'q' to exit")
	
	return s.String()
}

// Messages for async operations
type checkCompleteMsg struct {
	results []check.CheckResult
}

type generateCompleteMsg struct {
	err error
}

type validateCompleteMsg struct {
	err error
}

// Commands for async operations
func runChecks(checker check.Checker) tea.Cmd {
	return func() tea.Msg {
		results := checker.RunAll()
		return checkCompleteMsg{results: results}
	}
}

func generateConfigFiles(apiKeys map[string]string, templateName string) tea.Cmd {
	return func() tea.Msg {
		// Generate asc.toml
		if err := generateConfigFromTemplate(templateName); err != nil {
			return generateCompleteMsg{err: err}
		}
		
		// Generate .env
		if err := generateEnvFile(apiKeys); err != nil {
			return generateCompleteMsg{err: err}
		}
		
		return generateCompleteMsg{err: nil}
	}
}

func runValidation() tea.Cmd {
	return func() tea.Msg {
		// Run basic validation checks
		checker := check.NewChecker("asc.toml", ".env")
		results := checker.RunAll()
		
		// Check if validation passed
		if check.HasFailures(results) {
			return validateCompleteMsg{err: fmt.Errorf("validation failed: some checks did not pass")}
		}
		
		return validateCompleteMsg{err: nil}
	}
}

// Helper functions
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func backupConfigFiles() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	backupDir := filepath.Join(home, ".asc_backup")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}
	
	timestamp := time.Now().Format("20060102_150405")
	
	// Backup asc.toml
	if fileExists("asc.toml") {
		backupPath := filepath.Join(backupDir, fmt.Sprintf("asc.toml.%s", timestamp))
		if err := copyFile("asc.toml", backupPath); err != nil {
			return err
		}
	}
	
	// Backup .env
	if fileExists(".env") {
		backupPath := filepath.Join(backupDir, fmt.Sprintf(".env.%s", timestamp))
		if err := copyFile(".env", backupPath); err != nil {
			return err
		}
	}
	
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func validateAPIKey(keyName, value string) bool {
	if len(value) == 0 {
		return false
	}
	
	// Basic format validation
	switch keyName {
	case "CLAUDE_API_KEY":
		return strings.HasPrefix(value, "sk-ant-")
	case "OPENAI_API_KEY":
		return strings.HasPrefix(value, "sk-")
	case "GOOGLE_API_KEY":
		return strings.HasPrefix(value, "AIza")
	}
	
	return true
}

func generateConfigFromTemplate(templateName string) error {
	// If no template specified, use default team template
	if templateName == "" {
		templateName = "team"
	}
	
	// Try to get built-in template
	var template *config.Template
	var err error
	
	switch templateName {
	case "solo":
		template, err = config.GetTemplate(config.TemplateSolo)
	case "team":
		template, err = config.GetTemplate(config.TemplateTeam)
	case "swarm":
		template, err = config.GetTemplate(config.TemplateSwarm)
	default:
		// Try to load custom template
		template, err = config.LoadCustomTemplateByName(templateName)
	}
	
	if err != nil {
		return fmt.Errorf("failed to load template '%s': %w", templateName, err)
	}
	
	// Save template to asc.toml
	return config.SaveTemplate(template, "asc.toml")
}

func generateDefaultConfig() error {
	// Use team template as default
	return generateConfigFromTemplate("team")
}

func generateEnvFile(apiKeys map[string]string) error {
	var content strings.Builder
	
	for key, value := range apiKeys {
		content.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}
	
	// Set restrictive permissions for .env file
	return os.WriteFile(".env", []byte(content.String()), 0600)
}

// viewAgeSetup displays the age encryption setup screen
func (m wizardModel) viewAgeSetup() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		MarginTop(2).
		MarginBottom(1)
	
	textStyle := lipgloss.NewStyle().
		MarginBottom(1)
	
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true).
		MarginTop(2)
	
	var s strings.Builder
	s.WriteString(titleStyle.Render("🔐 Secure Secrets Management"))
	s.WriteString("\n\n")
	
	if !m.ageInstalled {
		s.WriteString(textStyle.Render("age encryption is not installed."))
		s.WriteString("\n\n")
		s.WriteString("age provides secure encryption for your API keys, preventing\n")
		s.WriteString("accidental exposure in git repositories.\n")
		s.WriteString("\n")
		s.WriteString("Without age, your API keys will be stored in plaintext.\n")
		s.WriteString("\n")
		s.WriteString(promptStyle.Render("Install age now? (y/N)"))
	} else if m.needsAgeSetup {
		s.WriteString(textStyle.Render("age is installed! Let's set up encryption."))
		s.WriteString("\n\n")
		s.WriteString("This will:\n")
		s.WriteString("  • Generate a secure encryption key (~/.asc/age.key)\n")
		s.WriteString("  • Encrypt your .env file automatically\n")
		s.WriteString("  • Keep your secrets safe in git\n")
		s.WriteString("\n")
		s.WriteString("Your API keys will be encrypted and only .env.age will be\n")
		s.WriteString("committed to git. The plaintext .env is automatically gitignored.\n")
		s.WriteString("\n")
		s.WriteString(promptStyle.Render("Set up encryption? (Y/n)"))
	}
	
	return s.String()
}

// viewEncrypting displays the encryption progress screen
func (m wizardModel) viewEncrypting() string {
	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		MarginTop(2)
	
	return spinnerStyle.Render("🔐 Encrypting secrets...")
}

// encryptSecrets encrypts the .env file
type encryptCompleteMsg struct {
	err error
}

func encryptSecrets(manager *secrets.Manager) tea.Cmd {
	return func() tea.Msg {
		// Encrypt .env to .env.age
		if err := manager.EncryptEnv(".env"); err != nil {
			return encryptCompleteMsg{err: err}
		}
		return encryptCompleteMsg{}
	}
}

// Update the message handling to include encrypt complete
func (m wizardModel) handleEncryptComplete(msg encryptCompleteMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		return m, tea.Quit
	}
	m.step = stepValidating
	return m, runValidation()
}
