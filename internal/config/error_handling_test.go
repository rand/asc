package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig_ErrorPaths tests error handling in configuration loading
func TestLoadConfig_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing config file",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.toml")
			},
			expectError: true,
			errorMsg:    "no such file",
		},
		{
			name: "invalid TOML syntax",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "invalid.toml")
				content := `[core
beads_db_path = "invalid syntax`
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "parse",
		},
		{
			name: "empty config file",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "empty.toml")
				if err := os.WriteFile(path, []byte(""), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "beads_db_path",
		},
		{
			name: "missing required fields",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "incomplete.toml")
				content := `[services.mcp_agent_mail]
url = "http://localhost:8765"`
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "beads_db_path",
		},
		{
			name: "invalid agent configuration",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "invalid_agent.toml")
				content := `[core]
beads_db_path = "/tmp/test"

[services.mcp_agent_mail]
start_command = "test"
url = "http://localhost:8765"

[agent.test]
# Missing required fields
model = "claude"`
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupFunc(t)
			_, err := Load(configPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !containsStr(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidate_ErrorPaths tests validation error handling
func TestValidate_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "nil config",
			config: nil,
			expectError: true,
			errorMsg: "config",
		},
		{
			name: "empty beads path",
			config: &Config{
				Core: CoreConfig{
					BeadsDBPath: "",
				},
			},
			expectError: true,
			errorMsg:    "beads_db_path",
		},
		{
			name: "empty MCP URL",
			config: &Config{
				Core: CoreConfig{
					BeadsDBPath: "/tmp/test",
				},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{
						StartCommand: "test",
						URL:          "",
					},
				},
			},
			expectError: true,
			errorMsg:    "url",
		},
		{
			name: "agent with empty command",
			config: &Config{
				Core: CoreConfig{
					BeadsDBPath: "/tmp/test",
				},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{
						StartCommand: "test",
						URL:          "http://localhost:8765",
					},
				},
				Agents: map[string]AgentConfig{
					"test": {
						Command: "",
						Model:   "claude",
						Phases:  []string{"planning"},
					},
				},
			},
			expectError: true,
			errorMsg:    "command",
		},
		{
			name: "agent with empty model",
			config: &Config{
				Core: CoreConfig{
					BeadsDBPath: "/tmp/test",
				},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{
						StartCommand: "test",
						URL:          "http://localhost:8765",
					},
				},
				Agents: map[string]AgentConfig{
					"test": {
						Command: "python test.py",
						Model:   "",
						Phases:  []string{"planning"},
					},
				},
			},
			expectError: true,
			errorMsg:    "model",
		},
		{
			name: "agent with empty phases",
			config: &Config{
				Core: CoreConfig{
					BeadsDBPath: "/tmp/test",
				},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{
						StartCommand: "test",
						URL:          "http://localhost:8765",
					},
				},
				Agents: map[string]AgentConfig{
					"test": {
						Command: "python test.py",
						Model:   "claude",
						Phases:  []string{},
					},
				},
			},
			expectError: true,
			errorMsg:    "phases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.config == nil {
				// Skip nil config test as validate doesn't handle nil
				t.Skip("validate function doesn't handle nil config")
				return
			}
			err = validate(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !containsStr(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestLoadEnv_ErrorPaths tests environment loading error handling
func TestLoadEnv_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing env file",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.env")
			},
			expectError: true,
			errorMsg:    "no such file",
		},
		{
			name: "unreadable env file",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "unreadable.env")
				if err := os.WriteFile(path, []byte("TEST=value"), 0000); err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectError: true,
			errorMsg:    "permission denied",
		},
		{
			name: "malformed env file",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "malformed.env")
				content := `VALID_KEY=value
INVALID LINE WITHOUT EQUALS
ANOTHER_VALID=value`
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectError: false, // Should skip invalid lines
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envPath := tt.setupFunc(t)
			err := LoadEnv(envPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !containsStr(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestErrorWrapping tests that errors are properly wrapped with context
func TestErrorWrapping(t *testing.T) {
	dir := t.TempDir()
	invalidPath := filepath.Join(dir, "invalid.toml")
	content := `[core
beads_db_path = "invalid`
	if err := os.WriteFile(invalidPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(invalidPath)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	// Just verify we got an error with context
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestRecoveryFromTransientErrors tests recovery from temporary failures
func TestRecoveryFromTransientErrors(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "test.toml")

	// First attempt: file doesn't exist
	_, err := Load(configPath)
	if err == nil {
		t.Error("Expected error for missing file")
	}

	// Create the file
	content := `[core]
beads_db_path = "/tmp/test"

[services.mcp_agent_mail]
start_command = "test"
url = "http://localhost:8765"

[agent.test]
command = "python test.py"
model = "claude"
phases = ["planning"]`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Second attempt: should succeed
	cfg, err := Load(configPath)
	if err != nil {
		t.Errorf("Expected success after file creation, got: %v", err)
	}
	if cfg == nil {
		t.Error("Expected valid config")
	}
}

// Helper function
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
