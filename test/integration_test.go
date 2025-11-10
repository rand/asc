// +build integration

package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/asc/internal/beads"
	"github.com/yourusername/asc/internal/check"
	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/process"
)

// TestProcessManagerIntegration tests the full lifecycle of process management
func TestProcessManagerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start multiple processes
	processes := []struct {
		name string
		cmd  string
		args []string
	}{
		{"proc1", "sleep", []string{"30"}},
		{"proc2", "sleep", []string{"30"}},
		{"proc3", "sleep", []string{"30"}},
	}

	pids := make([]int, len(processes))
	for i, proc := range processes {
		pid, err := manager.Start(proc.name, proc.cmd, proc.args, []string{"TEST=value"})
		if err != nil {
			t.Fatalf("Failed to start %s: %v", proc.name, err)
		}
		pids[i] = pid
	}

	// Verify all are running
	for i, pid := range pids {
		if !manager.IsRunning(pid) {
			t.Errorf("Process %s (PID %d) should be running", processes[i].name, pid)
		}
	}

	// List processes
	procList, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}
	if len(procList) != 3 {
		t.Errorf("Expected 3 processes, got %d", len(procList))
	}

	// Stop all
	if err := manager.StopAll(); err != nil {
		t.Errorf("StopAll failed: %v", err)
	}

	// Verify all stopped
	time.Sleep(200 * time.Millisecond)
	for i, pid := range pids {
		if manager.IsRunning(pid) {
			t.Errorf("Process %s (PID %d) should be stopped", processes[i].name, pid)
		}
	}
}

// TestConfigAndCheckIntegration tests config loading and dependency checking together
func TestConfigAndCheckIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config file
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./test-repo"

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
phases = ["implementation"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=sk-ant-test123
OPENAI_API_KEY=sk-test456
GOOGLE_API_KEY=test789
`
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write env: %v", err)
	}

	// Run checks
	checker := check.NewChecker(configPath, envPath)
	results := checker.RunAll()

	// Verify config check passed
	configPassed := false
	envPassed := false
	for _, result := range results {
		if result.Name == "asc.toml" && result.Status == check.CheckPass {
			configPassed = true
		}
		if result.Name == ".env" && result.Status == check.CheckPass {
			envPassed = true
		}
	}

	if !configPassed {
		t.Errorf("Config check should pass")
	}
	if !envPassed {
		t.Errorf("Env check should pass")
	}
}

// TestBeadsClientIntegration tests beads client with mock bd commands
func TestBeadsClientIntegration(t *testing.T) {
	// This test requires bd to be installed
	// Skip if bd is not available
	if _, err := os.Stat("/usr/local/bin/bd"); os.IsNotExist(err) {
		t.Skip("bd not installed, skipping integration test")
	}

	tmpDir := t.TempDir()
	client := beads.NewClient(tmpDir, 5*time.Second)

	// Note: This would require a real beads repo to be initialized
	// For now, we just test the client creation
	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

// TestMultiAgentScenario simulates multiple agents starting and stopping
func TestMultiAgentScenario(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Simulate agent configuration
	agents := []struct {
		name  string
		model string
		phase string
	}{
		{"planner-gemini", "gemini", "planning"},
		{"coder-claude", "claude", "implementation"},
		{"tester-gpt4", "gpt-4", "testing"},
	}

	// Start all agents
	for _, agent := range agents {
		env := []string{
			"AGENT_NAME=" + agent.name,
			"AGENT_MODEL=" + agent.model,
			"AGENT_PHASES=" + agent.phase,
		}
		_, err := manager.Start(agent.name, "sleep", []string{"20"}, env)
		if err != nil {
			t.Fatalf("Failed to start agent %s: %v", agent.name, err)
		}
	}

	// Verify all agents are running
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}
	if len(processes) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(processes))
	}

	// Verify environment variables were set
	for _, proc := range processes {
		if proc.Env["AGENT_NAME"] == "" {
			t.Errorf("Agent %s missing AGENT_NAME env var", proc.Name)
		}
		if proc.Env["AGENT_MODEL"] == "" {
			t.Errorf("Agent %s missing AGENT_MODEL env var", proc.Name)
		}
	}

	// Stop all agents
	if err := manager.StopAll(); err != nil {
		t.Errorf("Failed to stop all agents: %v", err)
	}
}

// TestConfigValidation tests various config validation scenarios
func TestConfigValidation(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		content    string
		shouldPass bool
	}{
		{
			name: "valid minimal config",
			content: `[core]
beads_db_path = "./repo"
`,
			shouldPass: true,
		},
		{
			name: "valid full config",
			content: `[core]
beads_db_path = "./repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test]
command = "python agent.py"
model = "claude"
phases = ["planning"]
`,
			shouldPass: true,
		},
		{
			name: "missing beads_db_path",
			content: `[services.mcp_agent_mail]
url = "http://localhost:8765"
`,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, tt.name+".toml")
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			checker := check.NewChecker(configPath, "")
			result := checker.CheckConfig()

			passed := result.Status == check.CheckPass
			if passed != tt.shouldPass {
				t.Errorf("Config validation = %v, want %v (status: %v, message: %s)",
					passed, tt.shouldPass, result.Status, result.Message)
			}
		})
	}
}

// TestProcessRecovery tests process recovery after crashes
func TestProcessRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start a process that will exit quickly
	pid, err := manager.Start("quick-exit", "echo", []string{"test"}, nil)
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Wait for it to exit
	time.Sleep(200 * time.Millisecond)

	// Verify it's no longer running
	if manager.IsRunning(pid) {
		t.Errorf("Process should have exited")
	}

	// Verify we can still get its info
	info, err := manager.GetProcessInfo("quick-exit")
	if err != nil {
		t.Errorf("Should be able to get info for exited process: %v", err)
	}
	if info.Name != "quick-exit" {
		t.Errorf("Process name = %v, want quick-exit", info.Name)
	}
}

// TestConcurrentProcessManagement tests concurrent process operations
func TestConcurrentProcessManagement(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start processes concurrently
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(n int) {
			name := filepath.Join("concurrent", string(rune('0'+n)))
			_, err := manager.Start(name, "sleep", []string{"10"}, nil)
			if err != nil {
				t.Errorf("Failed to start process %d: %v", n, err)
			}
			done <- true
		}(i)
	}

	// Wait for all to start
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all are running
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}
	if len(processes) < 5 {
		t.Errorf("Expected at least 5 processes, got %d", len(processes))
	}

	// Clean up
	manager.StopAll()
}
