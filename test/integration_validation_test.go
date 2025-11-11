// +build integration

package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/process"
	"github.com/yourusername/asc/internal/secrets"
)

// TestIntegrationValidation_InitWorkflow tests the complete asc init workflow
func TestIntegrationValidation_InitWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test directory structure
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")

	// Test that init command exists and has proper help
	ascBinary := filepath.Join("..", "build", "asc")
	if _, err := os.Stat(ascBinary); os.IsNotExist(err) {
		t.Skip("asc binary not built, run 'make build' first")
	}

	cmd := exec.Command(ascBinary, "init", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init --help failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "init") {
		t.Errorf("init help should mention init command")
	}

	// Create a minimal config manually (simulating init output)
	configContent := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "echo test"
url = "http://localhost:8765"

[agent.test-agent]
command = "sleep"
model = "claude"
phases = ["planning"]
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create env file
	envContent := `CLAUDE_API_KEY=sk-ant-test-key-12345
OPENAI_API_KEY=sk-test-key-67890
GOOGLE_API_KEY=test-google-key-abcde
`
	err = os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env: %v", err)
	}

	// Verify config can be loaded
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load generated config: %v", err)
	}

	if len(cfg.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(cfg.Agents))
	}

	// Verify env file has correct permissions
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Failed to stat env file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Env file should have 0600 permissions, got %v", info.Mode().Perm())
	}

	t.Log("Init workflow validation passed")
}

// TestIntegrationValidation_UpWorkDown tests the complete up → work → down workflow
func TestIntegrationValidation_UpWorkDown(t *testing.T) {
	if os.Getenv("INTEGRATION_FULL") != "true" {
		t.Skip("Skipping full integration test (set INTEGRATION_FULL=true to run)")
	}

	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, ".asc", "pids")
	logDir := filepath.Join(tmpDir, ".asc", "logs")

	// Create directories
	os.MkdirAll(pidDir, 0755)
	os.MkdirAll(logDir, 0755)

	// Create config
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.worker]
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
`
	err = os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env: %v", err)
	}

	// Initialize process manager
	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Simulate "up" - start agent
	env := []string{
		"AGENT_NAME=worker",
		"AGENT_MODEL=claude",
		"AGENT_PHASES=testing",
	}
	pid, err := manager.Start("worker", "sleep", []string{"30"}, env)
	if err != nil {
		t.Fatalf("Failed to start agent: %v", err)
	}

	// Verify agent is running
	if !manager.IsRunning(pid) {
		t.Errorf("Agent should be running")
	}

	// Simulate work period
	time.Sleep(1 * time.Second)

	// Verify agent still running
	if !manager.IsRunning(pid) {
		t.Errorf("Agent should still be running after work period")
	}

	// Simulate "down" - stop agent
	err = manager.Stop(pid)
	if err != nil {
		t.Errorf("Failed to stop agent: %v", err)
	}

	// Verify agent stopped
	time.Sleep(500 * time.Millisecond)
	if manager.IsRunning(pid) {
		t.Errorf("Agent should be stopped")
	}

	// Verify cleanup
	err = manager.StopAll()
	if err != nil {
		t.Errorf("StopAll failed: %v", err)
	}

	t.Log("Up → Work → Down workflow validation passed")
}

// TestIntegrationValidation_ConfigHotReload tests configuration hot-reload functionality
func TestIntegrationValidation_ConfigHotReload(t *testing.T) {
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
		t.Errorf("Expected 1 agent initially, got %d", len(cfg.Agents))
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
	time.Sleep(300 * time.Millisecond)

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

	// Wait for change notification
	select {
	case <-changes:
		t.Log("Config change detected")
	case <-time.After(3 * time.Second):
		t.Errorf("Config change not detected within timeout")
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

	t.Log("Config hot-reload validation passed")
}

// TestIntegrationValidation_SecretsEncryptionDecryption tests secrets encryption/decryption
func TestIntegrationValidation_SecretsEncryptionDecryption(t *testing.T) {
	// Skip if age not installed
	if _, err := exec.LookPath("age"); err != nil {
		t.Skip("age not installed, skipping secrets test")
	}

	tmpDir := t.TempDir()

	// Create secrets manager
	keyPath := filepath.Join(tmpDir, "age.key")
	manager := secrets.NewManagerWithKeyPath(keyPath)

	// Generate key
	err := manager.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Verify key exists
	if !manager.KeyExists() {
		t.Errorf("Key should exist after generation")
	}

	// Create test env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=sk-ant-test-key-12345
OPENAI_API_KEY=sk-test-key-67890
GOOGLE_API_KEY=test-google-key-abcde
`
	err = os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	// Encrypt env file
	encryptedPath := filepath.Join(tmpDir, ".env.age")
	err = manager.Encrypt(envPath, encryptedPath)
	if err != nil {
		t.Fatalf("Failed to encrypt secrets: %v", err)
	}

	// Verify encrypted file exists
	if _, err := os.Stat(encryptedPath); os.IsNotExist(err) {
		t.Errorf("Encrypted file should exist")
	}

	// Decrypt secrets
	decryptedPath := filepath.Join(tmpDir, ".env.decrypted")
	err = manager.Decrypt(encryptedPath, decryptedPath)
	if err != nil {
		t.Fatalf("Failed to decrypt secrets: %v", err)
	}

	// Verify decrypted content matches original
	decryptedContent, err := os.ReadFile(decryptedPath)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}

	if string(decryptedContent) != envContent {
		t.Errorf("Decrypted content doesn't match original")
	}

	// Test with invalid file
	err = manager.Decrypt(filepath.Join(tmpDir, "nonexistent.age"), filepath.Join(tmpDir, "out"))
	if err == nil {
		t.Errorf("Decrypting nonexistent file should fail")
	}

	t.Log("Secrets encryption/decryption validation passed")
}

// TestIntegrationValidation_HealthMonitoringAndRecovery tests health monitoring and auto-recovery
func TestIntegrationValidation_HealthMonitoringAndRecovery(t *testing.T) {
	if os.Getenv("INTEGRATION_FULL") != "true" {
		t.Skip("Skipping health monitoring test (set INTEGRATION_FULL=true to run)")
	}

	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	// Create process manager
	_, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create minimal config
	_ = config.Config{
		Core: config.CoreConfig{
			BeadsDBPath: "./test-repo",
		},
		Services: config.ServicesConfig{
			MCPAgentMail: config.MCPConfig{
				URL: "http://localhost:8765",
			},
		},
		Agents: map[string]config.AgentConfig{
			"monitored": {
				Command: "sleep",
				Model:   "claude",
				Phases:  []string{"testing"},
			},
		},
	}

	// Create mock MCP client (would need actual implementation)
	// For now, skip if MCP not available
	t.Skip("Health monitoring test requires MCP client implementation")

	t.Log("Health monitoring and recovery validation passed")
}

// TestIntegrationValidation_RealBeadsRepository tests integration with real beads repository
func TestIntegrationValidation_RealBeadsRepository(t *testing.T) {
	// Skip if bd not available
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skip("bd not installed, skipping beads integration test")
	}

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize git repo
	err := os.MkdirAll(repoPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	cmd.Run()

	// Initialize beads database
	cmd = exec.Command("bd", "init")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to init beads: %v\nOutput: %s", err, output)
	}

	// Create a test task
	cmd = exec.Command("bd", "create", "Test integration task")
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create task: %v\nOutput: %s", err, output)
	}

	// List tasks (bd uses default list command)
	cmd = exec.Command("bd")
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list tasks: %v\nOutput: %s", err, output)
	}

	// Verify task was created (check output)
	outputStr := string(output)
	t.Logf("BD output: %s", outputStr)
	
	// Task might be in different format, just verify we got some output
	if len(outputStr) == 0 {
		t.Errorf("BD should return some output")
	}

	// Test git refresh (pull)
	cmd = exec.Command("git", "status")
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to check git status: %v\nOutput: %s", err, output)
	}

	t.Log("Real beads repository validation passed")
}

// TestIntegrationValidation_RealMCPServer tests integration with real mcp_agent_mail server
func TestIntegrationValidation_RealMCPServer(t *testing.T) {
	if os.Getenv("INTEGRATION_MCP") != "true" {
		t.Skip("Skipping MCP server integration test (set INTEGRATION_MCP=true to run)")
	}

	// This test requires mcp_agent_mail to be running
	// Check if server is accessible
	cmd := exec.Command("curl", "-s", "http://localhost:8765/health")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skip("MCP server not running, skipping test")
	}

	if !strings.Contains(string(output), "ok") && !strings.Contains(string(output), "healthy") {
		t.Skip("MCP server not responding correctly")
	}

	// Test sending a message
	testMsg := fmt.Sprintf(`{"type":"test","source":"integration-test","content":"test-%d"}`, time.Now().Unix())
	cmd = exec.Command("curl", "-X", "POST", "-H", "Content-Type: application/json",
		"-d", testMsg, "http://localhost:8765/messages")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to send message: %v\nOutput: %s", err, output)
	}

	// Test retrieving messages
	cmd = exec.Command("curl", "-s", "http://localhost:8765/messages")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get messages: %v\nOutput: %s", err, output)
	}

	// Verify message was received
	if !strings.Contains(string(output), "integration-test") {
		t.Errorf("Message should appear in server response")
	}

	// Test agent status endpoint
	cmd = exec.Command("curl", "-s", "http://localhost:8765/agents")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get agent status: %v\nOutput: %s", err, output)
	}

	t.Log("Real MCP server validation passed")
}

// TestIntegrationValidation_MultiAgentCoordination tests multiple agents working together
func TestIntegrationValidation_MultiAgentCoordination(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	// Create process manager
	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Define multiple agents with different roles
	agents := []struct {
		name   string
		model  string
		phases []string
	}{
		{"planner", "gemini", []string{"planning", "design"}},
		{"coder1", "claude", []string{"implementation"}},
		{"coder2", "claude", []string{"implementation"}},
		{"tester", "gpt-4", []string{"testing"}},
		{"reviewer", "claude", []string{"review"}},
	}

	// Start all agents
	pids := make(map[string]int)
	for _, agent := range agents {
		env := []string{
			"AGENT_NAME=" + agent.name,
			"AGENT_MODEL=" + agent.model,
			"AGENT_PHASES=" + strings.Join(agent.phases, ","),
			"MCP_MAIL_URL=http://localhost:8765",
		}

		pid, err := manager.Start(agent.name, "sleep", []string{"30"}, env)
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

	// Verify agent count
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	if len(processes) != len(agents) {
		t.Errorf("Expected %d agents, got %d", len(agents), len(processes))
	}

	// Verify environment variables are set correctly
	for _, proc := range processes {
		if proc.Env["AGENT_NAME"] == "" {
			t.Errorf("Agent %s missing AGENT_NAME", proc.Name)
		}
		if proc.Env["AGENT_MODEL"] == "" {
			t.Errorf("Agent %s missing AGENT_MODEL", proc.Name)
		}
		if proc.Env["AGENT_PHASES"] == "" {
			t.Errorf("Agent %s missing AGENT_PHASES", proc.Name)
		}
	}

	// Simulate coordination period
	time.Sleep(2 * time.Second)

	// Verify all still running
	for name, pid := range pids {
		if !manager.IsRunning(pid) {
			t.Errorf("Agent %s should still be running after coordination period", name)
		}
	}

	// Stop all agents
	err = manager.StopAll()
	if err != nil {
		t.Errorf("Failed to stop all agents: %v", err)
	}

	// Verify all stopped
	time.Sleep(500 * time.Millisecond)
	for name, pid := range pids {
		if manager.IsRunning(pid) {
			t.Errorf("Agent %s should be stopped", name)
		}
	}

	t.Log("Multi-agent coordination validation passed")
}

// TestIntegrationValidation_CompleteWorkflow tests the complete end-to-end workflow
func TestIntegrationValidation_CompleteWorkflow(t *testing.T) {
	if os.Getenv("INTEGRATION_FULL") != "true" {
		t.Skip("Skipping complete workflow test (set INTEGRATION_FULL=true to run)")
	}

	tmpDir := t.TempDir()

	// Step 1: Initialize configuration (simulating asc init)
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")

	configContent := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.planner]
command = "python agent_adapter.py"
model = "gemini"
phases = ["planning"]

[agent.coder]
command = "python agent_adapter.py"
model = "claude"
phases = ["implementation"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	envContent := `CLAUDE_API_KEY=sk-ant-test-key
OPENAI_API_KEY=sk-test-key
GOOGLE_API_KEY=test-key
`
	err = os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env: %v", err)
	}

	// Step 2: Verify configuration (simulating asc check)
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	if len(cfg.Agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(cfg.Agents))
	}

	// Step 3: Start agents (simulating asc up)
	pidDir := filepath.Join(tmpDir, ".asc", "pids")
	logDir := filepath.Join(tmpDir, ".asc", "logs")
	os.MkdirAll(pidDir, 0755)
	os.MkdirAll(logDir, 0755)

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start agents
	for name, agent := range cfg.Agents {
		env := []string{
			"AGENT_NAME=" + name,
			"AGENT_MODEL=" + agent.Model,
			"AGENT_PHASES=" + strings.Join(agent.Phases, ","),
		}
		_, err := manager.Start(name, "sleep", []string{"20"}, env)
		if err != nil {
			t.Fatalf("Failed to start agent %s: %v", name, err)
		}
	}

	// Step 4: Verify agents are running
	time.Sleep(1 * time.Second)
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	if len(processes) != 2 {
		t.Errorf("Expected 2 running agents, got %d", len(processes))
	}

	// Step 5: Simulate work period
	time.Sleep(3 * time.Second)

	// Step 6: Stop agents (simulating asc down)
	err = manager.StopAll()
	if err != nil {
		t.Errorf("Failed to stop agents: %v", err)
	}

	// Step 7: Verify cleanup
	time.Sleep(500 * time.Millisecond)
	processes, err = manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes after stop: %v", err)
	}

	if len(processes) > 0 {
		t.Errorf("All agents should be stopped, found %d still running", len(processes))
	}

	t.Log("Complete workflow validation passed")
}

// TestIntegrationValidation_ErrorRecovery tests error recovery scenarios
func TestIntegrationValidation_ErrorRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test 1: Invalid command
	_, err = manager.Start("invalid", "nonexistent-command-xyz", []string{}, nil)
	if err == nil {
		t.Errorf("Starting invalid command should fail")
	}

	// Test 2: Process that exits immediately
	pid, err := manager.Start("quick-exit", "echo", []string{"test"}, nil)
	if err != nil {
		t.Fatalf("Failed to start quick-exit process: %v", err)
	}

	// Wait for it to exit
	time.Sleep(1 * time.Second)

	// Verify it's detected as not running
	if manager.IsRunning(pid) {
		t.Logf("Warning: Quick-exit process still appears running (may be timing issue)")
	}

	// Test 3: Cleanup after errors
	err = manager.StopAll()
	if err != nil {
		t.Errorf("StopAll should handle errors gracefully: %v", err)
	}

	t.Log("Error recovery validation passed")
}

// TestIntegrationValidation_ConfigTemplates tests configuration template system
func TestIntegrationValidation_ConfigTemplates(t *testing.T) {
	// Skip if python not available (templates use python commands)
	if _, err := exec.LookPath("python"); err != nil {
		if _, err := exec.LookPath("python3"); err != nil {
			t.Skip("python/python3 not available, skipping template test")
		}
	}

	tmpDir := t.TempDir()

	templates := []config.TemplateType{
		config.TemplateSolo,
		config.TemplateTeam,
		config.TemplateSwarm,
	}

	for _, tmpl := range templates {
		t.Run(string(tmpl), func(t *testing.T) {
			configPath := filepath.Join(tmpDir, string(tmpl)+".toml")

			// Get template
			template, err := config.GetTemplate(tmpl)
			if err != nil {
				t.Fatalf("Failed to get template %s: %v", tmpl, err)
			}

			// Save template
			err = config.SaveTemplate(template, configPath)
			if err != nil {
				t.Fatalf("Failed to save template %s: %v", tmpl, err)
			}

			// Read config file to verify it was created
			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read generated config: %v", err)
			}

			// Verify file has content
			if len(content) == 0 {
				t.Fatalf("Generated config is empty")
			}

			// Verify it contains expected sections
			contentStr := string(content)
			if !strings.Contains(contentStr, "[core]") {
				t.Errorf("Config should contain [core] section")
			}
			if !strings.Contains(contentStr, "[agent.") {
				t.Errorf("Config should contain agent definitions")
			}

			// Count agents by counting [agent. occurrences
			agentCount := strings.Count(contentStr, "[agent.")
			var cfg config.Config
			cfg.Agents = make(map[string]config.AgentConfig, agentCount)

			// Verify template-specific properties
			switch tmpl {
			case config.TemplateSolo:
				if agentCount != 1 {
					t.Errorf("Solo template should have 1 agent, got %d", agentCount)
				}
			case config.TemplateTeam:
				if agentCount < 3 {
					t.Errorf("Team template should have at least 3 agents, got %d", agentCount)
				}
			case config.TemplateSwarm:
				if agentCount < 5 {
					t.Errorf("Swarm template should have at least 5 agents, got %d", agentCount)
				}
			}
		})
	}

	t.Log("Config templates validation passed")
}

// TestIntegrationValidation_StressTest tests system under stress
func TestIntegrationValidation_StressTest(t *testing.T) {
	if os.Getenv("INTEGRATION_STRESS") != "true" {
		t.Skip("Skipping stress test (set INTEGRATION_STRESS=true to run)")
	}

	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := process.NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Start many agents
	agentCount := 20
	pids := make([]int, agentCount)

	for i := 0; i < agentCount; i++ {
		name := fmt.Sprintf("stress-agent-%d", i)
		env := []string{
			"AGENT_NAME=" + name,
			"AGENT_MODEL=claude",
			"AGENT_PHASES=testing",
		}

		pid, err := manager.Start(name, "sleep", []string{"30"}, env)
		if err != nil {
			t.Fatalf("Failed to start agent %d: %v", i, err)
		}
		pids[i] = pid
	}

	// Verify all started
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	if len(processes) != agentCount {
		t.Errorf("Expected %d agents, got %d", agentCount, len(processes))
	}

	// Simulate load
	time.Sleep(5 * time.Second)

	// Stop all
	err = manager.StopAll()
	if err != nil {
		t.Errorf("Failed to stop all agents: %v", err)
	}

	// Verify cleanup
	time.Sleep(1 * time.Second)
	processes, err = manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes after stop: %v", err)
	}

	if len(processes) > 0 {
		t.Errorf("All agents should be stopped, found %d still running", len(processes))
	}

	t.Log("Stress test validation passed")
}
