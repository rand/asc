package config

import (
	"fmt"
	"testing"
)

// mockProcessManager is a mock implementation of ProcessManager for testing
type mockProcessManager struct {
	processes map[string]*ProcessInfoAdapter
	started   []string
	stopped   []int
}

func newMockProcessManager() *mockProcessManager {
	return &mockProcessManager{
		processes: make(map[string]*ProcessInfoAdapter),
		started:   []string{},
		stopped:   []int{},
	}
}

func (m *mockProcessManager) Start(name, command string, args []string, env []string) (int, error) {
	pid := len(m.processes) + 1000
	envMap := make(map[string]string)
	for _, e := range env {
		// Simple parsing of KEY=VALUE
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				envMap[e[:i]] = e[i+1:]
				break
			}
		}
	}
	
	m.processes[name] = &ProcessInfoAdapter{
		Name:    name,
		PID:     pid,
		Command: command,
		Args:    args,
		Env:     envMap,
	}
	m.started = append(m.started, name)
	return pid, nil
}

func (m *mockProcessManager) Stop(pid int) error {
	m.stopped = append(m.stopped, pid)
	// Find and remove process
	for name, info := range m.processes {
		if info.PID == pid {
			delete(m.processes, name)
			break
		}
	}
	return nil
}

func (m *mockProcessManager) StopAll() error {
	for _, info := range m.processes {
		m.stopped = append(m.stopped, info.PID)
	}
	m.processes = make(map[string]*ProcessInfoAdapter)
	return nil
}

func (m *mockProcessManager) IsRunning(pid int) bool {
	for _, info := range m.processes {
		if info.PID == pid {
			return true
		}
	}
	return false
}

func (m *mockProcessManager) GetStatus(pid int) string {
	if m.IsRunning(pid) {
		return "running"
	}
	return "stopped"
}

func (m *mockProcessManager) GetProcessInfo(name string) (ProcessInfoGetter, error) {
	if info, exists := m.processes[name]; exists {
		return info, nil
	}
	return nil, fmt.Errorf("process not found: %s", name)
}

func (m *mockProcessManager) ListProcesses() ([]*ProcessInfoAdapter, error) {
	list := make([]*ProcessInfoAdapter, 0, len(m.processes))
	for _, info := range m.processes {
		list = append(list, info)
	}
	return list, nil
}

func TestReloadManager_AddAgent(t *testing.T) {
	// Create initial config with one agent
	oldConfig := &Config{
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
				Command: "echo test",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}
	
	// Create new config with two agents
	newConfig := &Config{
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
				Command: "echo test",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			"agent2": {
				Command: "echo test2",
				Model:   "gemini",
				Phases:  []string{"implementation"},
			},
		},
	}
	
	// Create mock process manager
	procManager := newMockProcessManager()
	
	// Create reload manager
	envVars := map[string]string{
		"CLAUDE_API_KEY": "test-key",
	}
	reloadManager := NewReloadManager(oldConfig, procManager, envVars)
	
	// Perform reload
	result, err := reloadManager.Reload(newConfig)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	
	// Verify results
	if len(result.AgentsAdded) != 1 {
		t.Errorf("Expected 1 agent added, got %d", len(result.AgentsAdded))
	}
	
	if len(result.AgentsAdded) > 0 && result.AgentsAdded[0] != "agent2" {
		t.Errorf("Expected agent2 to be added, got %s", result.AgentsAdded[0])
	}
	
	if len(result.AgentsRemoved) != 0 {
		t.Errorf("Expected 0 agents removed, got %d", len(result.AgentsRemoved))
	}
	
	if len(result.AgentsUpdated) != 0 {
		t.Errorf("Expected 0 agents updated, got %d", len(result.AgentsUpdated))
	}
	
	// Verify agent2 was started
	if len(procManager.started) != 1 {
		t.Errorf("Expected 1 agent started, got %d", len(procManager.started))
	}
	
	if len(procManager.started) > 0 && procManager.started[0] != "agent2" {
		t.Errorf("Expected agent2 to be started, got %s", procManager.started[0])
	}
}

func TestReloadManager_RemoveAgent(t *testing.T) {
	// Create initial config with two agents
	oldConfig := &Config{
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
				Command: "echo test",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			"agent2": {
				Command: "echo test2",
				Model:   "gemini",
				Phases:  []string{"implementation"},
			},
		},
	}
	
	// Create new config with one agent
	newConfig := &Config{
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
				Command: "echo test",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}
	
	// Create mock process manager with agent2 running
	procManager := newMockProcessManager()
	procManager.processes["agent2"] = &ProcessInfoAdapter{
		Name:    "agent2",
		PID:     1001,
		Command: "echo",
		Args:    []string{"test2"},
		Env:     map[string]string{},
	}
	
	// Create reload manager
	envVars := map[string]string{
		"CLAUDE_API_KEY": "test-key",
	}
	reloadManager := NewReloadManager(oldConfig, procManager, envVars)
	
	// Perform reload
	result, err := reloadManager.Reload(newConfig)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	
	// Verify results
	if len(result.AgentsRemoved) != 1 {
		t.Errorf("Expected 1 agent removed, got %d", len(result.AgentsRemoved))
	}
	
	if len(result.AgentsRemoved) > 0 && result.AgentsRemoved[0] != "agent2" {
		t.Errorf("Expected agent2 to be removed, got %s", result.AgentsRemoved[0])
	}
	
	// Verify agent2 was stopped
	if len(procManager.stopped) != 1 {
		t.Errorf("Expected 1 agent stopped, got %d", len(procManager.stopped))
	}
	
	if len(procManager.stopped) > 0 && procManager.stopped[0] != 1001 {
		t.Errorf("Expected PID 1001 to be stopped, got %d", procManager.stopped[0])
	}
}

func TestReloadManager_UpdateAgent(t *testing.T) {
	// Create initial config
	oldConfig := &Config{
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
				Command: "echo test",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}
	
	// Create new config with updated agent
	newConfig := &Config{
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
				Command: "echo test",
				Model:   "gemini", // Changed model
				Phases:  []string{"planning", "implementation"}, // Added phase
			},
		},
	}
	
	// Create mock process manager with agent1 running
	procManager := newMockProcessManager()
	procManager.processes["agent1"] = &ProcessInfoAdapter{
		Name:    "agent1",
		PID:     1000,
		Command: "echo",
		Args:    []string{"test"},
		Env:     map[string]string{},
	}
	
	// Create reload manager
	envVars := map[string]string{
		"CLAUDE_API_KEY": "test-key",
	}
	reloadManager := NewReloadManager(oldConfig, procManager, envVars)
	
	// Perform reload
	result, err := reloadManager.Reload(newConfig)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	
	// Verify results
	if len(result.AgentsUpdated) != 1 {
		t.Errorf("Expected 1 agent updated, got %d", len(result.AgentsUpdated))
	}
	
	if len(result.AgentsUpdated) > 0 && result.AgentsUpdated[0] != "agent1" {
		t.Errorf("Expected agent1 to be updated, got %s", result.AgentsUpdated[0])
	}
	
	// Verify agent1 was stopped and restarted
	if len(procManager.stopped) != 1 {
		t.Errorf("Expected 1 agent stopped, got %d", len(procManager.stopped))
	}
	
	if len(procManager.started) != 1 {
		t.Errorf("Expected 1 agent started, got %d", len(procManager.started))
	}
	
	// Verify new process has updated environment
	info, err := procManager.GetProcessInfo("agent1")
	if err != nil {
		t.Fatalf("Failed to get process info: %v", err)
	}
	
	env := info.GetEnv()
	if env["AGENT_MODEL"] != "gemini" {
		t.Errorf("Expected model 'gemini', got '%s'", env["AGENT_MODEL"])
	}
	
	if env["AGENT_PHASES"] != "planning,implementation" {
		t.Errorf("Expected phases 'planning,implementation', got '%s'", env["AGENT_PHASES"])
	}
}
