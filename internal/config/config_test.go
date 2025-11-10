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
