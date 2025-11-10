package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Benchmark configuration loading
func BenchmarkLoad(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "bench-asc.toml")

	content := `[core]
beads_db_path = "./test-repo"

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
phases = ["implementation", "refactor"]

[agent.tester]
command = "python agent_adapter.py"
model = "gpt-4"
phases = ["testing", "review"]
`

	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		b.Fatalf("Failed to write test config: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load(configPath)
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}

// Benchmark validation
func BenchmarkValidate(b *testing.B) {
	cfg := &Config{
		Core: CoreConfig{
			BeadsDBPath: "./test-repo",
		},
		Services: ServicesConfig{
			MCPAgentMail: MCPConfig{
				StartCommand: "python -m mcp_agent_mail.server",
				URL:          "http://localhost:8765",
			},
		},
		Agents: map[string]AgentConfig{
			"planner": {
				Command: "echo",
				Model:   "gemini",
				Phases:  []string{"planning", "design"},
			},
			"coder": {
				Command: "echo",
				Model:   "claude",
				Phases:  []string{"implementation"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validate(cfg)
		if err != nil {
			b.Fatalf("Validate failed: %v", err)
		}
	}
}

// Benchmark agent validation
func BenchmarkValidateAgent(b *testing.B) {
	agent := AgentConfig{
		Command: "echo",
		Model:   "claude",
		Phases:  []string{"planning", "implementation", "testing"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateAgent("test-agent", agent)
		if err != nil {
			b.Fatalf("ValidateAgent failed: %v", err)
		}
	}
}

// Benchmark model validation
func BenchmarkIsValidModel(b *testing.B) {
	models := []string{"claude", "gemini", "gpt-4", "codex", "openai", "invalid"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, model := range models {
			_ = isValidModel(model)
		}
	}
}

// Benchmark phase validation
func BenchmarkIsValidPhase(b *testing.B) {
	phases := []string{"planning", "implementation", "testing", "review", "invalid"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, phase := range phases {
			_ = isValidPhase(phase)
		}
	}
}

// Benchmark closest phase finding
func BenchmarkFindClosestPhase(b *testing.B) {
	validPhases := []string{
		"planning", "design", "implementation", "coding",
		"testing", "review", "refactor", "documentation",
		"debugging", "optimization", "deployment",
	}
	
	inputs := []string{"plan", "impl", "test", "doc", "xyz"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			_ = findClosestPhase(input, validPhases)
		}
	}
}

// Benchmark path expansion
func BenchmarkExpandPath(b *testing.B) {
	paths := []string{
		"./relative/path",
		"~/home/path",
		"/absolute/path",
		"$HOME/env/path",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_, _ = expandPath(path)
		}
	}
}

// Benchmark validation with warnings
func BenchmarkValidateWithWarnings(b *testing.B) {
	cfg := &Config{
		Core: CoreConfig{
			BeadsDBPath: "./test-repo",
		},
		Services: ServicesConfig{
			MCPAgentMail: MCPConfig{
				StartCommand: "python -m mcp_agent_mail.server",
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
				Phases:  []string{"implementation"},
			},
			"agent3": {
				Command: "echo",
				Model:   "claude",
				Phases:  []string{"testing"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateWithWarnings(cfg)
		if err != nil {
			b.Fatalf("ValidateWithWarnings failed: %v", err)
		}
	}
}

// Benchmark loading large config with many agents
func BenchmarkLoadLargeConfig(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "large-asc.toml")

	// Generate config with 20 agents
	content := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

`
	for i := 1; i <= 20; i++ {
		content += `
[agent.agent` + string(rune('0'+i%10)) + `]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning", "implementation", "testing"]
`
	}

	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		b.Fatalf("Failed to write test config: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load(configPath)
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}
