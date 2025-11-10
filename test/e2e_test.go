// +build e2e

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestE2ECheckCommand tests the asc check command end-to-end
func TestE2ECheckCommand(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid config
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=test
OPENAI_API_KEY=test
GOOGLE_API_KEY=test
`
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write env: %v", err)
	}

	// Run asc check
	cmd := exec.Command("./build/asc", "check")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	// Check command should run (may fail on missing binaries, but should execute)
	if err != nil && !strings.Contains(string(output), "Dependency") {
		t.Errorf("asc check failed unexpectedly: %v\nOutput: %s", err, output)
	}

	// Verify output contains expected elements
	outputStr := string(output)
	if !strings.Contains(outputStr, "git") {
		t.Errorf("Output should mention git")
	}
}

// TestE2EInitWorkflow tests the initialization workflow
func TestE2EInitWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	// Note: This test would require interactive input handling
	// For now, we test that the command exists and can be invoked
	cmd := exec.Command("./build/asc", "init", "--help")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("asc init --help failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "init") {
		t.Errorf("Help output should mention init command")
	}
}

// TestE2EServicesCommand tests service management commands
func TestE2EServicesCommand(t *testing.T) {
	tmpDir := t.TempDir()

	// Test services status (should work even without config)
	cmd := exec.Command("./build/asc", "services", "status")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	// Command should execute (may report service not running)
	outputStr := string(output)
	if err != nil && !strings.Contains(outputStr, "not running") && !strings.Contains(outputStr, "status") {
		t.Errorf("asc services status failed: %v\nOutput: %s", err, output)
	}
}

// TestE2EFullStackLifecycle tests the complete stack lifecycle
func TestE2EFullStackLifecycle(t *testing.T) {
	// This test requires all dependencies to be installed
	// Skip if not in a full environment
	if os.Getenv("E2E_FULL") != "true" {
		t.Skip("Skipping full stack test (set E2E_FULL=true to run)")
	}

	tmpDir := t.TempDir()

	// Create complete configuration
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "sleep"
model = "claude"
phases = ["testing"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=test-key
OPENAI_API_KEY=test-key
GOOGLE_API_KEY=test-key
`
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write env: %v", err)
	}

	// Initialize beads repo
	repoPath := filepath.Join(tmpDir, "test-repo")
	err = os.MkdirAll(repoPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Run check
	cmd = exec.Command("./build/asc", "check")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	t.Logf("Check output: %s", output)

	// Note: Full up/down cycle would require mcp_agent_mail to be running
	// This is tested in manual/integration environments
}

// TestE2EErrorHandling tests error scenarios
func TestE2EErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		command []string
		wantErr bool
	}{
		{
			name:    "check without config",
			command: []string{"check"},
			wantErr: true,
		},
		{
			name:    "up without config",
			command: []string{"up"},
			wantErr: true,
		},
		{
			name:    "invalid command",
			command: []string{"invalid-command"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./build/asc", tt.command...)
			cmd.Dir = tmpDir
			err := cmd.Run()

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestE2EConfigValidation tests config validation scenarios
func TestE2EConfigValidation(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		config     string
		shouldFail bool
	}{
		{
			name: "valid config",
			config: `[core]
beads_db_path = "./repo"
`,
			shouldFail: false,
		},
		{
			name: "invalid TOML",
			config: `[core
beads_db_path = 
`,
			shouldFail: true,
		},
		{
			name: "missing required field",
			config: `[services.mcp_agent_mail]
url = "http://localhost:8765"
`,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}

			configPath := filepath.Join(testDir, "asc.toml")
			err = os.WriteFile(configPath, []byte(tt.config), 0644)
			if err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			cmd := exec.Command("./build/asc", "check")
			cmd.Dir = testDir
			err = cmd.Run()

			if tt.shouldFail && err == nil {
				t.Errorf("Expected check to fail but it passed")
			}
		})
	}
}

// TestE2EMultiAgentConfiguration tests multiple agent configurations
func TestE2EMultiAgentConfiguration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with multiple agents
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.planner]
command = "python agent.py"
model = "gemini"
phases = ["planning", "design"]

[agent.coder]
command = "python agent.py"
model = "claude"
phases = ["implementation", "refactor"]

[agent.tester]
command = "python agent.py"
model = "gpt-4"
phases = ["testing", "validation"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=test
OPENAI_API_KEY=test
GOOGLE_API_KEY=test
`
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write env: %v", err)
	}

	// Run check
	cmd := exec.Command("./build/asc", "check")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	t.Logf("Multi-agent check output: %s", output)

	// Config should be valid
	outputStr := string(output)
	if strings.Contains(outputStr, "Invalid") {
		t.Errorf("Config should be valid")
	}
}

// TestE2EHelpCommands tests help output for all commands
func TestE2EHelpCommands(t *testing.T) {
	commands := []string{"init", "up", "down", "check", "test", "services"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			command := exec.Command("./build/asc", cmd, "--help")
			output, err := command.CombinedOutput()

			if err != nil {
				t.Errorf("Help command failed: %v\nOutput: %s", err, output)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, cmd) {
				t.Errorf("Help output should mention %s command", cmd)
			}
		})
	}
}

// TestE2EVersionCommand tests version output
func TestE2EVersionCommand(t *testing.T) {
	cmd := exec.Command("./build/asc", "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Version command failed: %v\nOutput: %s", err, output)
	}

	// Output should contain version information
	outputStr := string(output)
	if len(outputStr) == 0 {
		t.Errorf("Version output should not be empty")
	}
}

// TestE2EProcessCleanup tests that processes are properly cleaned up
func TestE2EProcessCleanup(t *testing.T) {
	if os.Getenv("E2E_FULL") != "true" {
		t.Skip("Skipping process cleanup test (set E2E_FULL=true to run)")
	}

	tmpDir := t.TempDir()

	// Create minimal config
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"

[agent.test]
command = "sleep"
model = "claude"
phases = ["testing"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Start the stack (this would normally start agents)
	// For testing, we just verify the down command works
	cmd := exec.Command("./build/asc", "down")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	t.Logf("Down output: %s", output)

	// Down should complete (even if nothing was running)
	if err != nil && !strings.Contains(string(output), "not running") {
		t.Errorf("Down command failed: %v", err)
	}
}

// TestE2ERapidStartStop tests rapid start/stop cycles
func TestE2ERapidStartStop(t *testing.T) {
	if os.Getenv("E2E_STRESS") != "true" {
		t.Skip("Skipping stress test (set E2E_STRESS=true to run)")
	}

	tmpDir := t.TempDir()

	// Create config
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Perform rapid start/stop cycles
	for i := 0; i < 5; i++ {
		// Start
		cmd := exec.Command("./build/asc", "services", "start")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		time.Sleep(100 * time.Millisecond)

		// Stop
		cmd = exec.Command("./build/asc", "services", "stop")
		cmd.Dir = tmpDir
		_ = cmd.Run()

		time.Sleep(100 * time.Millisecond)
	}

	// Final cleanup
	cmd := exec.Command("./build/asc", "down")
	cmd.Dir = tmpDir
	_ = cmd.Run()
}

// TestE2ELargeConfiguration tests handling of large configurations
func TestE2ELargeConfiguration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with many agents
	configPath := filepath.Join(tmpDir, "asc.toml")
	var configContent strings.Builder
	configContent.WriteString(`[core]
beads_db_path = "./repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

`)

	// Add 20 agents
	for i := 0; i < 20; i++ {
		configContent.WriteString("[agent.agent-")
		configContent.WriteString(string(rune('0' + i)))
		configContent.WriteString("]\n")
		configContent.WriteString("command = \"python agent.py\"\n")
		configContent.WriteString("model = \"claude\"\n")
		configContent.WriteString("phases = [\"testing\"]\n\n")
	}

	err := os.WriteFile(configPath, []byte(configContent.String()), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=test
`
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write env: %v", err)
	}

	// Run check
	cmd := exec.Command("./build/asc", "check")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	t.Logf("Large config check output: %s", output)

	// Should handle large config without crashing
	if err != nil && strings.Contains(string(output), "panic") {
		t.Errorf("Should handle large config without panicking")
	}
}
