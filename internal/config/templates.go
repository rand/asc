package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Template represents a configuration template
type Template struct {
	Name        string
	Description string
	Content     string
}

// TemplateType represents the type of template
type TemplateType string

const (
	TemplateSolo  TemplateType = "solo"
	TemplateTeam  TemplateType = "team"
	TemplateSwarm TemplateType = "swarm"
)

// GetTemplate returns a predefined template by name
func GetTemplate(templateType TemplateType) (*Template, error) {
	switch templateType {
	case TemplateSolo:
		return getSoloTemplate(), nil
	case TemplateTeam:
		return getTeamTemplate(), nil
	case TemplateSwarm:
		return getSwarmTemplate(), nil
	default:
		return nil, fmt.Errorf("unknown template type: %s", templateType)
	}
}

// getSoloTemplate returns a single agent configuration
func getSoloTemplate() *Template {
	return &Template{
		Name:        "solo",
		Description: "Single agent setup for individual development",
		Content: `[core]
beads_db_path = "./project-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.solo-agent]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning", "implementation", "testing", "review", "refactor"]
`,
	}
}

// getTeamTemplate returns a team configuration with specialized agents
func getTeamTemplate() *Template {
	return &Template{
		Name:        "team",
		Description: "Team setup with planner, coder, and tester agents",
		Content: `[core]
beads_db_path = "./project-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.planner]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning", "design"]

[agent.coder]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation", "coding"]

[agent.tester]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing", "review"]
`,
	}
}

// getSwarmTemplate returns a swarm configuration with multiple agents per phase
func getSwarmTemplate() *Template {
	return &Template{
		Name:        "swarm",
		Description: "Swarm setup with multiple agents per phase for parallel work",
		Content: `[core]
beads_db_path = "./project-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.planner-1]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning", "design"]

[agent.planner-2]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning", "design"]

[agent.coder-1]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation", "coding"]

[agent.coder-2]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["implementation", "coding"]

[agent.coder-3]
command = "python agent_adapter.py"
model = "codex"
phases = ["implementation", "coding"]

[agent.tester-1]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing", "review"]

[agent.tester-2]
command = "python agent_adapter.py"
model = "gemini"
phases = ["testing", "review"]

[agent.refactor]
command = "python agent_adapter.py"
model = "claude"
phases = ["refactor", "optimization"]
`,
	}
}

// ListTemplates returns all available templates
func ListTemplates() []Template {
	return []Template{
		*getSoloTemplate(),
		*getTeamTemplate(),
		*getSwarmTemplate(),
	}
}

// SaveTemplate saves a template to a file
func SaveTemplate(template *Template, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write template content to file
	if err := os.WriteFile(path, []byte(template.Content), 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	return nil
}

// LoadCustomTemplate loads a custom template from a file
func LoadCustomTemplate(path string) (*Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	name := filepath.Base(path)
	return &Template{
		Name:        name,
		Description: "Custom template",
		Content:     string(content),
	}, nil
}

// SaveCustomTemplate saves the current configuration as a custom template
func SaveCustomTemplate(configPath, templateName string) error {
	// Read current config
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Get templates directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	templatesDir := filepath.Join(home, ".asc", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Save template
	templatePath := filepath.Join(templatesDir, templateName+".toml")
	if err := os.WriteFile(templatePath, content, 0644); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}

// ListCustomTemplates returns all custom templates from ~/.asc/templates
func ListCustomTemplates() ([]Template, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	templatesDir := filepath.Join(home, ".asc", "templates")
	
	// Check if directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return []Template{}, nil
	}

	// Read directory
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	templates := []Template{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		templatePath := filepath.Join(templatesDir, entry.Name())
		content, err := os.ReadFile(templatePath)
		if err != nil {
			continue // Skip files we can't read
		}

		name := entry.Name()[:len(entry.Name())-5] // Remove .toml extension
		templates = append(templates, Template{
			Name:        name,
			Description: "Custom template",
			Content:     string(content),
		})
	}

	return templates, nil
}

// LoadCustomTemplateByName loads a custom template by name from ~/.asc/templates
func LoadCustomTemplateByName(name string) (*Template, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	templatesDir := filepath.Join(home, ".asc", "templates")
	templatePath := filepath.Join(templatesDir, name+".toml")
	
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	return &Template{
		Name:        name,
		Description: "Custom template",
		Content:     string(content),
	}, nil
}
