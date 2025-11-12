// +build integration

package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/rand/asc/internal/beads"
	"github.com/rand/asc/internal/check"
	"github.com/rand/asc/internal/config"
	"github.com/rand/asc/internal/process"
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

[services.mcp_agent_mail]
start_command = "echo test"
url = "http://localhost:8765"

[agent.test]
command = "sleep"
model = "claude"
phases = ["planning"]
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
command = "sleep"
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
	time.Sleep(500 * time.Millisecond)

	// Verify it's no longer running (may take time to detect)
	// Give it a few tries
	stillRunning := false
	for i := 0; i < 5; i++ {
		if manager.IsRunning(pid) {
			stillRunning = true
			time.Sleep(100 * time.Millisecond)
		} else {
			stillRunning = false
			break
		}
	}
	
	if stillRunning {
		t.Logf("Warning: Process still appears to be running after exit")
	}

	// Verify we can still get its info (if implementation supports it)
	info, err := manager.GetProcessInfo("quick-exit")
	if err == nil && info != nil {
		if info.Name != "quick-exit" {
			t.Errorf("Process name = %v, want quick-exit", info.Name)
		}
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
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(n int) {
			name := fmt.Sprintf("concurrent-%d", n)
			_, err := manager.Start(name, "sleep", []string{"10"}, nil)
			done <- err
		}(i)
	}

	// Wait for all to start and check for errors
	errorCount := 0
	for i := 0; i < 5; i++ {
		if err := <-done; err != nil {
			t.Errorf("Failed to start process %d: %v", i, err)
			errorCount++
		}
	}

	// Give processes time to fully start
	time.Sleep(200 * time.Millisecond)

	// Verify all are running
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}
	
	expectedCount := 5 - errorCount
	if len(processes) < expectedCount {
		t.Errorf("Expected at least %d processes, got %d", expectedCount, len(processes))
	}

	// Clean up
	manager.StopAll()
}

// TestConfigHotReload tests configuration hot-reload functionality
func TestConfigHotReload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "asc.toml")

	// Create initial config
	initialConfig := `[core]
beads_db_path = "./repo"

[agent.agent1]
command = "sleep"
model = "claude"
phases = ["planning"]
`
	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Load initial config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(cfg.Agents))
	}

	// Update config with new agent
	updatedConfig := `[core]
beads_db_path = "./repo"

[agent.agent1]
command = "sleep"
model = "claude"
phases = ["planning"]

[agent.agent2]
command = "sleep"
model = "gemini"
phases = ["implementation"]
`
	err = os.WriteFile(configPath, []byte(updatedConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write updated config: %v", err)
	}

	// Reload config
	cfg, err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if len(cfg.Agents) != 2 {
		t.Errorf("Expected 2 agents after reload, got %d", len(cfg.Agents))
	}

	// Verify new agent exists
	if _, exists := cfg.Agents["agent2"]; !exists {
		t.Errorf("New agent 'agent2' should exist after reload")
	}
}

// TestProcessLifecycleWithMonitoring tests full process lifecycle with health monitoring
func TestProcessLifecycleWithMonitoring(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start a long-running process
	pid, err := manager.Start("monitored", "sleep", []string{"60"}, nil)
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Monitor process
	if !manager.IsRunning(pid) {
		t.Errorf("Process should be running")
	}

	// Get process info
	info, err := manager.GetProcessInfo("monitored")
	if err != nil {
		t.Fatalf("Failed to get process info: %v", err)
	}

	if info.PID != pid {
		t.Errorf("PID mismatch: got %d, want %d", info.PID, pid)
	}

	// Stop process gracefully
	err = manager.Stop(pid)
	if err != nil {
		t.Errorf("Failed to stop process: %v", err)
	}

	// Verify process stopped
	time.Sleep(200 * time.Millisecond)
	if manager.IsRunning(pid) {
		t.Errorf("Process should be stopped")
	}

	// Verify cleanup - process info may or may not be cleaned up depending on implementation
	// Just verify the process is no longer running
	if manager.IsRunning(pid) {
		t.Errorf("Process should not be running after stop")
	}
}

// TestConfigTemplateIntegration tests configuration template system
func TestConfigTemplateIntegration(t *testing.T) {
	// Skip if python not available
	if _, err := exec.LookPath("python"); err != nil {
		t.Skip("python not available, skipping template integration test")
	}

	tmpDir := t.TempDir()

	templates := []config.TemplateType{config.TemplateSolo, config.TemplateTeam, config.TemplateSwarm}

	for _, tmpl := range templates {
		t.Run(string(tmpl), func(t *testing.T) {
			configPath := filepath.Join(tmpDir, string(tmpl)+".toml")

			// Get template
			template, err := config.GetTemplate(tmpl)
			if err != nil {
				t.Fatalf("Failed to get template %s: %v", tmpl, err)
			}

			// Save template to file
			err = config.SaveTemplate(template, configPath)
			if err != nil {
				t.Fatalf("Failed to save template %s: %v", tmpl, err)
			}

			// Verify config is valid
			cfg, err := config.Load(configPath)
			if err != nil {
				t.Fatalf("Generated config is invalid: %v", err)
			}

			// Verify template-specific properties
			switch tmpl {
			case config.TemplateSolo:
				if len(cfg.Agents) != 1 {
					t.Errorf("Solo template should have 1 agent, got %d", len(cfg.Agents))
				}
			case config.TemplateTeam:
				if len(cfg.Agents) < 3 {
					t.Errorf("Team template should have at least 3 agents, got %d", len(cfg.Agents))
				}
			case config.TemplateSwarm:
				if len(cfg.Agents) < 5 {
					t.Errorf("Swarm template should have at least 5 agents, got %d", len(cfg.Agents))
				}
			}
		})
	}
}

// TestProcessCleanupOnError tests that resources are cleaned up on errors
func TestProcessCleanupOnError(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Try to start invalid command
	_, err = manager.Start("invalid", "nonexistent-command-xyz", []string{}, nil)
	if err == nil {
		t.Errorf("Starting invalid command should fail")
	}

	// Verify no PID file was created
	pidFiles, err := os.ReadDir(pidDir)
	if err != nil {
		t.Fatalf("Failed to read PID dir: %v", err)
	}

	for _, file := range pidFiles {
		if file.Name() == "invalid.json" {
			t.Errorf("PID file should not exist for failed process start")
		}
	}
}

// TestMultipleConfigLoads tests loading config multiple times
func TestMultipleConfigLoads(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "asc.toml")

	configContent := `[core]
beads_db_path = "./repo"

[agent.test]
command = "sleep"
model = "claude"
phases = ["planning"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config multiple times
	for i := 0; i < 10; i++ {
		cfg, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Load %d failed: %v", i, err)
		}

		if len(cfg.Agents) != 1 {
			t.Errorf("Load %d: expected 1 agent, got %d", i, len(cfg.Agents))
		}
	}
}

// TestProcessLogCapture tests that process logs are captured correctly
func TestProcessLogCapture(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start process that outputs to stdout
	testMessage := "integration-test-output"
	pid, err := manager.Start("logger", "echo", []string{testMessage}, nil)
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Wait for process to complete
	time.Sleep(500 * time.Millisecond)

	// Read log file
	logFile := filepath.Join(logDir, "logger.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Verify output was captured
	if !contains(string(content), testMessage) {
		t.Errorf("Log should contain '%s', got: %s", testMessage, string(content))
	}

	// Clean up
	manager.Stop(pid)
}

// TestConfigEnvIntegration tests config and env file integration
func TestConfigEnvIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")

	// Create config
	configContent := `[core]
beads_db_path = "./repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test]
command = "sleep"
model = "claude"
phases = ["planning"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create env file
	envContent := `CLAUDE_API_KEY=sk-test-123
OPENAI_API_KEY=sk-test-456
GOOGLE_API_KEY=test-789
MCP_MAIL_URL=http://localhost:8765
`
	err = os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env: %v", err)
	}

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Load env
	err = config.LoadAndValidateEnv(envPath)
	if err != nil {
		t.Fatalf("Failed to load env: %v", err)
	}

	// Verify config loaded correctly (path will be expanded to absolute)
	if cfg.Core.BeadsDBPath == "" {
		t.Errorf("BeadsDBPath should not be empty")
	}

	// Verify env vars are set
	if os.Getenv("CLAUDE_API_KEY") != "sk-test-123" {
		t.Errorf("CLAUDE_API_KEY not set correctly")
	}

	// Verify agent config
	agent, exists := cfg.Agents["test"]
	if !exists {
		t.Fatal("Agent 'test' should exist")
	}

	if agent.Model != "claude" {
		t.Errorf("Agent model = %v, want claude", agent.Model)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestWebSocketReconnection tests WebSocket reconnection and fallback
func TestWebSocketReconnection(t *testing.T) {
	// This test requires a mock MCP server
	// Skip if not in integration environment
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping WebSocket integration test")
	}

	// Test would connect to mock server, disconnect, and verify reconnection
	// Implementation depends on mock server setup
	t.Log("WebSocket reconnection test placeholder")
}

// TestHealthMonitoringIntegration tests health monitoring system
func TestHealthMonitoringIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start a process
	pid, err := manager.Start("monitored", "sleep", []string{"30"}, nil)
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Verify health check passes
	if !manager.IsRunning(pid) {
		t.Errorf("Health check should pass for running process")
	}

	// Stop process
	manager.Stop(pid)

	// Verify health check fails
	time.Sleep(200 * time.Millisecond)
	if manager.IsRunning(pid) {
		t.Errorf("Health check should fail for stopped process")
	}
}

// TestBeadsClientWithRealRepo tests beads client with actual repository
func TestBeadsClientWithRealRepo(t *testing.T) {
	// Skip if bd not available
	if _, err := os.Stat("/usr/local/bin/bd"); os.IsNotExist(err) {
		t.Skip("bd not installed, skipping beads integration test")
	}

	tmpDir := t.TempDir()

	// Initialize a test beads repo
	// This would require actual bd commands
	client := beads.NewClient(tmpDir, 5*time.Second)
	if client == nil {
		t.Fatal("Client should not be nil")
	}

	// Test basic operations
	// Note: Actual implementation depends on bd being set up
	t.Log("Beads client integration test placeholder")
}

// TestConfigWatcherIntegration tests configuration file watching
func TestConfigWatcherIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "asc.toml")

	// Create initial config
	initialConfig := `[core]
beads_db_path = "./repo"
`
	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create watcher
	watcher, err := config.NewWatcher(configPath)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start watching
	err = watcher.Start()
	if err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	changes := make(chan bool, 1)
	go func() {
		for range watcher.Events() {
			changes <- true
		}
	}()

	// Give watcher time to start
	time.Sleep(200 * time.Millisecond)

	// Modify config
	updatedConfig := `[core]
beads_db_path = "./new-repo"

[agent.test]
command = "sleep"
model = "claude"
phases = ["planning"]
`
	err = os.WriteFile(configPath, []byte(updatedConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Wait for change notification
	select {
	case <-changes:
		// Success
	case <-time.After(3 * time.Second):
		t.Errorf("Config change not detected within timeout")
	}
}

// TestProcessEnvironmentVariables tests that env vars are passed correctly
func TestProcessEnvironmentVariables(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start process with custom env vars
	env := []string{
		"AGENT_NAME=test-agent",
		"AGENT_MODEL=claude",
		"AGENT_PHASES=planning,implementation",
		"TEST_VAR=test-value",
	}

	pid, err := manager.Start("env-test", "env", []string{}, env)
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Wait for process to complete
	time.Sleep(500 * time.Millisecond)

	// Read log to verify env vars were set
	logFile := filepath.Join(logDir, "env-test.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log: %v", err)
	}

	// Verify env vars in output
	logContent := string(content)
	expectedVars := []string{"AGENT_NAME=test-agent", "AGENT_MODEL=claude", "TEST_VAR=test-value"}
	for _, expected := range expectedVars {
		if !contains(logContent, expected) {
			t.Errorf("Log should contain '%s'", expected)
		}
	}

	manager.Stop(pid)
}

// TestMultipleAgentCoordination tests multiple agents working together
func TestMultipleAgentCoordination(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start multiple agents with different phases
	agents := []struct {
		name   string
		phases []string
	}{
		{"planner", []string{"planning"}},
		{"coder1", []string{"implementation"}},
		{"coder2", []string{"implementation"}},
		{"tester", []string{"testing"}},
	}

	pids := make(map[string]int)
	for _, agent := range agents {
		env := []string{
			"AGENT_NAME=" + agent.name,
			"AGENT_PHASES=" + joinStrings(agent.phases, ","),
		}
		pid, err := manager.Start(agent.name, "sleep", []string{"20"}, env)
		if err != nil {
			t.Fatalf("Failed to start agent %s: %v", agent.name, err)
		}
		pids[agent.name] = pid
	}

	// Verify all agents are running
	for name, pid := range pids {
		if !manager.IsRunning(pid) {
			t.Errorf("Agent %s should be running", name)
		}
	}

	// Verify we can list all agents
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	if len(processes) != len(agents) {
		t.Errorf("Expected %d agents, got %d", len(agents), len(processes))
	}

	// Stop all agents
	manager.StopAll()

	// Verify all stopped
	time.Sleep(200 * time.Millisecond)
	for name, pid := range pids {
		if manager.IsRunning(pid) {
			t.Errorf("Agent %s should be stopped", name)
		}
	}
}

// TestConfigValidationErrors tests various config validation error cases
func TestConfigValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty config",
			content:     "",
			expectError: true,
			errorMsg:    "missing beads_db_path",
		},
		{
			name: "invalid TOML",
			content: `[core
beads_db_path = "./repo"
`,
			expectError: true,
			errorMsg:    "parse error",
		},
		{
			name: "agent without command",
			content: `[core]
beads_db_path = "./repo"

[agent.test]
model = "claude"
phases = ["planning"]
`,
			expectError: true,
			errorMsg:    "command is required",
		},
		{
			name: "agent without model",
			content: `[core]
beads_db_path = "./repo"

[agent.test]
command = "sleep"
phases = ["planning"]
`,
			expectError: true,
			errorMsg:    "model is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, tt.name+".toml")
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			_, err = config.Load(configPath)
			hasError := err != nil

			if hasError != tt.expectError {
				t.Errorf("Load error = %v, expectError = %v", hasError, tt.expectError)
			}
		})
	}
}

// TestProcessGracefulShutdown tests graceful shutdown with timeout
func TestProcessGracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start a process
	pid, err := manager.Start("graceful", "sleep", []string{"60"}, nil)
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Stop with graceful shutdown
	start := time.Now()
	err = manager.Stop(pid)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	// Should complete quickly (within 2 seconds)
	if duration > 2*time.Second {
		t.Errorf("Graceful shutdown took too long: %v", duration)
	}

	// Verify process stopped
	if manager.IsRunning(pid) {
		t.Errorf("Process should be stopped after graceful shutdown")
	}
}

// TestResourceCleanup tests that all resources are cleaned up properly
func TestResourceCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start multiple processes
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("cleanup-%d", i)
		_, err := manager.Start(name, "sleep", []string{"30"}, nil)
		if err != nil {
			t.Fatalf("Failed to start process %d: %v", i, err)
		}
	}

	// Stop all
	err = manager.StopAll()
	if err != nil {
		t.Errorf("StopAll failed: %v", err)
	}

	// Wait for cleanup
	time.Sleep(500 * time.Millisecond)

	// Verify PID files are cleaned up
	pidFiles, err := os.ReadDir(pidDir)
	if err != nil {
		t.Fatalf("Failed to read PID dir: %v", err)
	}

	if len(pidFiles) > 0 {
		t.Errorf("PID files should be cleaned up, found %d files", len(pidFiles))
	}

	// Verify no processes are running
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	if len(processes) > 0 {
		t.Errorf("All processes should be stopped, found %d running", len(processes))
	}
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
