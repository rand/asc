package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigStructure(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantCore string
		wantURL  string
	}{
		{
			name: "valid config with all fields",
			config: Config{
				Core: CoreConfig{
					BeadsDBPath: "/path/to/beads",
				},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{
						StartCommand: "python -m mcp_agent_mail.server",
						URL:          "http://localhost:8765",
					},
				},
				Agents: map[string]AgentConfig{
					"test-agent": {
						Command: "python agent.py",
						Model:   "claude",
						Phases:  []string{"planning", "implementation"},
					},
				},
			},
			wantCore: "/path/to/beads",
			wantURL:  "http://localhost:8765",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Core.BeadsDBPath != tt.wantCore {
				t.Errorf("BeadsDBPath = %v, want %v", tt.config.Core.BeadsDBPath, tt.wantCore)
			}
			if tt.config.Services.MCPAgentMail.URL != tt.wantURL {
				t.Errorf("MCP URL = %v, want %v", tt.config.Services.MCPAgentMail.URL, tt.wantURL)
			}
		})
	}
}

func TestAgentConfig(t *testing.T) {
	agent := AgentConfig{
		Command: "python agent.py",
		Model:   "gemini",
		Phases:  []string{"testing", "validation"},
	}

	if agent.Command != "python agent.py" {
		t.Errorf("Command = %v, want %v", agent.Command, "python agent.py")
	}
	if agent.Model != "gemini" {
		t.Errorf("Model = %v, want %v", agent.Model, "gemini")
	}
	if len(agent.Phases) != 2 {
		t.Errorf("Phases length = %v, want 2", len(agent.Phases))
	}
}

func TestConfigWithMultipleAgents(t *testing.T) {
	config := Config{
		Agents: map[string]AgentConfig{
			"planner": {
				Command: "python agent.py",
				Model:   "gemini",
				Phases:  []string{"planning"},
			},
			"coder": {
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"implementation"},
			},
			"tester": {
				Command: "python agent.py",
				Model:   "gpt-4",
				Phases:  []string{"testing"},
			},
		},
	}

	if len(config.Agents) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(config.Agents))
	}

	if config.Agents["planner"].Model != "gemini" {
		t.Errorf("Planner model = %v, want gemini", config.Agents["planner"].Model)
	}
	if config.Agents["coder"].Model != "claude" {
		t.Errorf("Coder model = %v, want claude", config.Agents["coder"].Model)
	}
	if config.Agents["tester"].Model != "gpt-4" {
		t.Errorf("Tester model = %v, want gpt-4", config.Agents["tester"].Model)
	}
}

func TestConfigDefaults(t *testing.T) {
	config := Config{}

	if config.Core.BeadsDBPath != "" {
		t.Errorf("Expected empty BeadsDBPath, got %v", config.Core.BeadsDBPath)
	}
	if config.Agents == nil {
		config.Agents = make(map[string]AgentConfig)
	}
	if len(config.Agents) != 0 {
		t.Errorf("Expected 0 agents, got %d", len(config.Agents))
	}
}

func TestMCPConfig(t *testing.T) {
	mcp := MCPConfig{
		StartCommand: "uvx mcp-agent-mail",
		URL:          "http://127.0.0.1:9000",
	}

	if mcp.StartCommand != "uvx mcp-agent-mail" {
		t.Errorf("StartCommand = %v, want uvx mcp-agent-mail", mcp.StartCommand)
	}
	if mcp.URL != "http://127.0.0.1:9000" {
		t.Errorf("URL = %v, want http://127.0.0.1:9000", mcp.URL)
	}
}

func TestConfigFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-asc.toml")

	// Create a test config file
	content := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "python agent.py"
model = "claude"
phases = ["planning"]
`

	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}
}

func TestValidateAgent(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		agent     AgentConfig
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid agent",
			agentName: "test-agent",
			agent: AgentConfig{
				Command: "echo",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			wantError: false,
		},
		{
			name:      "missing command",
			agentName: "test-agent",
			agent: AgentConfig{
				Model:  "claude",
				Phases: []string{"planning"},
			},
			wantError: true,
			errorMsg:  "command is required",
		},
		{
			name:      "command not in PATH",
			agentName: "test-agent",
			agent: AgentConfig{
				Command: "nonexistent-command-xyz",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			wantError: true,
			errorMsg:  "not found in PATH",
		},
		{
			name:      "unsupported model",
			agentName: "test-agent",
			agent: AgentConfig{
				Command: "echo",
				Model:   "unsupported-model",
				Phases:  []string{"planning"},
			},
			wantError: true,
			errorMsg:  "unsupported model",
		},
		{
			name:      "invalid phase",
			agentName: "test-agent",
			agent: AgentConfig{
				Command: "echo",
				Model:   "claude",
				Phases:  []string{"invalid-phase"},
			},
			wantError: true,
			errorMsg:  "invalid phase",
		},
		{
			name:      "no phases",
			agentName: "test-agent",
			agent: AgentConfig{
				Command: "echo",
				Model:   "claude",
				Phases:  []string{},
			},
			wantError: true,
			errorMsg:  "at least one phase is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAgent(tt.agentName, tt.agent)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestIsValidModel(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"claude", true},
		{"gemini", true},
		{"gpt-4", true},
		{"codex", true},
		{"openai", true},
		{"Claude", true}, // case insensitive
		{"GEMINI", true},
		{"invalid", false},
		{"gpt-3", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := isValidModel(tt.model)
			if got != tt.want {
				t.Errorf("isValidModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestIsValidPhase(t *testing.T) {
	tests := []struct {
		phase string
		want  bool
	}{
		{"planning", true},
		{"implementation", true},
		{"testing", true},
		{"review", true},
		{"Planning", true}, // case insensitive
		{"TESTING", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			got := isValidPhase(tt.phase)
			if got != tt.want {
				t.Errorf("isValidPhase(%q) = %v, want %v", tt.phase, got, tt.want)
			}
		})
	}
}

func TestFindClosestPhase(t *testing.T) {
	validPhases := []string{"planning", "implementation", "testing", "documentation"}
	
	tests := []struct {
		input string
		want  string
	}{
		{"plan", "planning"},
		{"impl", "implementation"},
		{"test", "testing"},
		{"doc", "documentation"},
		{"docs", "documentation"},
		{"plann", "planning"},
		{"implement", "implementation"},
		{"xyz", ""}, // no match
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := findClosestPhase(tt.input, validPhases)
			if got != tt.want {
				t.Errorf("findClosestPhase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateWithWarnings(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		wantWarnings int
		wantError    bool
	}{
		{
			name: "all agents same model",
			config: Config{
				Core: CoreConfig{
					BeadsDBPath: "./test",
				},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{
						StartCommand: "test",
						URL:          "http://localhost:8765",
					},
				},
				Agents: map[string]AgentConfig{
					"agent1": {
						Command: "echo",
						Model:   "claude",
						Phases:  []string{"planning"},
					},
					"agent2": {
						Command: "echo",
						Model:   "claude",
						Phases:  []string{"implementation", "testing"},
					},
				},
			},
			wantWarnings: 1, // Only warning about same model
			wantError:    false,
		},
		{
			name: "missing critical phase",
			config: Config{
				Core: CoreConfig{
					BeadsDBPath: "./test",
				},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{
						StartCommand: "test",
						URL:          "http://localhost:8765",
					},
				},
				Agents: map[string]AgentConfig{
					"agent1": {
						Command: "echo",
						Model:   "claude",
						Phases:  []string{"documentation"},
					},
				},
			},
			wantWarnings: 3, // missing planning, implementation, testing
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := ValidateWithWarnings(&tt.config)
			if tt.wantError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if len(warnings) != tt.wantWarnings {
				t.Errorf("Expected %d warnings, got %d: %v", tt.wantWarnings, len(warnings), warnings)
			}
		})
	}
}

func TestDuplicateAgentNames(t *testing.T) {
	// Note: Viper is case-insensitive and will silently deduplicate keys in TOML
	// So [agent.test-agent] and [agent.Test-Agent] will result in only one agent
	// This test verifies that behavior rather than expecting an error
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-asc.toml")

	content := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "echo"
model = "claude"
phases = ["planning"]

[agent.Test-Agent]
command = "echo"
model = "gemini"
phases = ["implementation"]
`

	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}

	// Viper deduplicates case-insensitive keys, so we should only have 1 agent
	if len(cfg.Agents) != 1 {
		t.Errorf("Expected 1 agent after deduplication, got %d", len(cfg.Agents))
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
