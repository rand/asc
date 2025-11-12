package config

import (
	"fmt"
	"testing"
)

// Mock ProcessManager for testing
type mockProcessManager struct {
	processes map[string]*mockProcessInfo
	startErr  error
	stopErr   error
}

type mockProcessInfo struct {
	name    string
	pid     int
	command string
	args    []string
	env     map[string]string
	running bool
}

func (m *mockProcessInfo) GetPID() int                     { return m.pid }
func (m *mockProcessInfo) GetName() string                { return m.name }
func (m *mockProcessInfo) GetCommand() string             { return m.command }
func (m *mockProcessInfo) GetArgs() []string              { return m.args }
func (m *mockProcessInfo) GetEnv() map[string]string      { return m.env }

func newMockProcessManager() *mockProcessManager {
	return &mockProcessManager{
		processes: make(map[string]*mockProcessInfo),
	}
}

func (m *mockProcessManager) Start(name, command string, args []string, env []string) (int, error) {
	if m.startErr != nil {
		return 0, m.startErr
	}
	
	pid := len(m.processes) + 1000
	envMap := make(map[string]string)
	for _, e := range env {
		// Simple parsing for testing
		envMap[e] = e
	}
	
	m.processes[name] = &mockProcessInfo{
		name:    name,
		pid:     pid,
		command: command,
		args:    args,
		env:     envMap,
		running: true,
	}
	
	return pid, nil
}

func (m *mockProcessManager) Stop(pid int) error {
	if m.stopErr != nil {
		return m.stopErr
	}
	
	for _, proc := range m.processes {
		if proc.pid == pid {
			proc.running = false
			return nil
		}
	}
	
	return fmt.Errorf("process not found: %d", pid)
}

func (m *mockProcessManager) IsRunning(pid int) bool {
	for _, proc := range m.processes {
		if proc.pid == pid {
			return proc.running
		}
	}
	return false
}

func (m *mockProcessManager) GetProcessInfo(name string) (ProcessInfoGetter, error) {
	proc, exists := m.processes[name]
	if !exists {
		return nil, fmt.Errorf("process not found: %s", name)
	}
	return proc, nil
}

func TestNewReloadManager(t *testing.T) {
	config := &Config{
		Core: CoreConfig{
			BeadsDBPath: "./test-repo",
		},
	}
	
	procManager := newMockProcessManager()
	envVars := map[string]string{
		"CLAUDE_API_KEY": "test-key",
	}
	
	rm := NewReloadManager(config, procManager, envVars)
	
	if rm == nil {
		t.Fatal("NewReloadManager() returned nil")
	}
	
	if rm.currentConfig != config {
		t.Errorf("NewReloadManager() config mismatch")
	}
	
	if rm.processManager != procManager {
		t.Errorf("NewReloadManager() processManager mismatch")
	}
}

func TestReloadManager_GetCurrentConfig(t *testing.T) {
	config := &Config{
		Core: CoreConfig{
			BeadsDBPath: "./test-repo",
		},
	}
	
	rm := NewReloadManager(config, newMockProcessManager(), nil)
	
	currentConfig := rm.GetCurrentConfig()
	if currentConfig != config {
		t.Errorf("GetCurrentConfig() returned wrong config")
	}
}

func TestReloadManager_AgentConfigChanged(t *testing.T) {
	rm := NewReloadManager(&Config{}, newMockProcessManager(), nil)
	
	tests := []struct {
		name string
		old  AgentConfig
		new  AgentConfig
		want bool
	}{
		{
			name: "no change",
			old: AgentConfig{
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			new: AgentConfig{
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			want: false,
		},
		{
			name: "command changed",
			old: AgentConfig{
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			new: AgentConfig{
				Command: "python new_agent.py",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			want: true,
		},
		{
			name: "model changed",
			old: AgentConfig{
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			new: AgentConfig{
				Command: "python agent.py",
				Model:   "gemini",
				Phases:  []string{"planning"},
			},
			want: true,
		},
		{
			name: "phases changed - different count",
			old: AgentConfig{
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			new: AgentConfig{
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"planning", "implementation"},
			},
			want: true,
		},
		{
			name: "phases changed - different phases",
			old: AgentConfig{
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			new: AgentConfig{
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"implementation"},
			},
			want: true,
		},
		{
			name: "phases same but different order",
			old: AgentConfig{
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"planning", "implementation"},
			},
			new: AgentConfig{
				Command: "python agent.py",
				Model:   "claude",
				Phases:  []string{"implementation", "planning"},
			},
			want: false, // Order doesn't matter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rm.agentConfigChanged(tt.old, tt.new)
			if got != tt.want {
				t.Errorf("agentConfigChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReloadManager_StopAgent(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*mockProcessManager)
		agentName string
		wantErr   bool
	}{
		{
			name: "stop running agent",
			setup: func(pm *mockProcessManager) {
				pm.Start("test-agent", "echo", []string{}, []string{})
			},
			agentName: "test-agent",
			wantErr:   false,
		},
		{
			name: "stop non-existent agent",
			setup: func(pm *mockProcessManager) {
				// Don't start any agents
			},
			agentName: "non-existent",
			wantErr:   false, // Should not error for non-existent agent
		},
		{
			name: "stop already stopped agent",
			setup: func(pm *mockProcessManager) {
				pid, _ := pm.Start("test-agent", "echo", []string{}, []string{})
				pm.Stop(pid)
			},
			agentName: "test-agent",
			wantErr:   false, // Should not error for already stopped agent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := newMockProcessManager()
			tt.setup(pm)
			
			rm := NewReloadManager(&Config{}, pm, nil)
			
			err := rm.stopAgent(tt.agentName)
			if (err != nil) != tt.wantErr {
				t.Errorf("stopAgent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReloadManager_StartAgent(t *testing.T) {
	tests := []struct {
		name        string
		agentName   string
		agentConfig AgentConfig
		config      *Config
		wantErr     bool
	}{
		{
			name:      "start valid agent",
			agentName: "test-agent",
			agentConfig: AgentConfig{
				Command: "echo test",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			config: &Config{
				Core: CoreConfig{
					BeadsDBPath: "./test-repo",
				},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{
						URL: "http://localhost:8765",
					},
				},
			},
			wantErr: false,
		},
		{
			name:      "start agent with empty command",
			agentName: "test-agent",
			agentConfig: AgentConfig{
				Command: "",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			config: &Config{
				Core: CoreConfig{
					BeadsDBPath: "./test-repo",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := newMockProcessManager()
			rm := NewReloadManager(&Config{}, pm, map[string]string{
				"CLAUDE_API_KEY": "test-key",
			})
			
			err := rm.startAgent(tt.agentName, tt.agentConfig, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("startAgent() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if !tt.wantErr {
				// Verify agent was started
				_, err := pm.GetProcessInfo(tt.agentName)
				if err != nil {
					t.Errorf("Agent was not started: %v", err)
				}
			}
		})
	}
}

func TestReloadManager_Reload(t *testing.T) {
	tests := []struct {
		name           string
		currentConfig  *Config
		newConfig      *Config
		wantAdded      int
		wantRemoved    int
		wantUpdated    int
		wantErrors     int
	}{
		{
			name: "add new agent",
			currentConfig: &Config{
				Core: CoreConfig{BeadsDBPath: "./test"},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{URL: "http://localhost:8765"},
				},
				Agents: map[string]AgentConfig{
					"agent1": {
						Command: "echo test",
						Model:   "claude",
						Phases:  []string{"planning"},
					},
				},
			},
			newConfig: &Config{
				Core: CoreConfig{BeadsDBPath: "./test"},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{URL: "http://localhost:8765"},
				},
				Agents: map[string]AgentConfig{
					"agent1": {
						Command: "echo test",
						Model:   "claude",
						Phases:  []string{"planning"},
					},
					"agent2": {
						Command: "echo test",
						Model:   "gemini",
						Phases:  []string{"implementation"},
					},
				},
			},
			wantAdded:   1,
			wantRemoved: 0,
			wantUpdated: 0,
			wantErrors:  0,
		},
		{
			name: "remove agent",
			currentConfig: &Config{
				Core: CoreConfig{BeadsDBPath: "./test"},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{URL: "http://localhost:8765"},
				},
				Agents: map[string]AgentConfig{
					"agent1": {
						Command: "echo test",
						Model:   "claude",
						Phases:  []string{"planning"},
					},
					"agent2": {
						Command: "echo test",
						Model:   "gemini",
						Phases:  []string{"implementation"},
					},
				},
			},
			newConfig: &Config{
				Core: CoreConfig{BeadsDBPath: "./test"},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{URL: "http://localhost:8765"},
				},
				Agents: map[string]AgentConfig{
					"agent1": {
						Command: "echo test",
						Model:   "claude",
						Phases:  []string{"planning"},
					},
				},
			},
			wantAdded:   0,
			wantRemoved: 1,
			wantUpdated: 0,
			wantErrors:  0,
		},
		{
			name: "update agent",
			currentConfig: &Config{
				Core: CoreConfig{BeadsDBPath: "./test"},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{URL: "http://localhost:8765"},
				},
				Agents: map[string]AgentConfig{
					"agent1": {
						Command: "echo test",
						Model:   "claude",
						Phases:  []string{"planning"},
					},
				},
			},
			newConfig: &Config{
				Core: CoreConfig{BeadsDBPath: "./test"},
				Services: ServicesConfig{
					MCPAgentMail: MCPConfig{URL: "http://localhost:8765"},
				},
				Agents: map[string]AgentConfig{
					"agent1": {
						Command: "echo test",
						Model:   "gemini", // Changed model
						Phases:  []string{"planning"},
					},
				},
			},
			wantAdded:   0,
			wantRemoved: 0,
			wantUpdated: 1,
			wantErrors:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := newMockProcessManager()
			
			// Start agents from current config
			for name, agent := range tt.currentConfig.Agents {
				pm.Start(name, agent.Command, []string{}, []string{})
			}
			
			rm := NewReloadManager(tt.currentConfig, pm, map[string]string{
				"CLAUDE_API_KEY": "test-key",
			})
			
			result, err := rm.Reload(tt.newConfig)
			if err != nil {
				t.Fatalf("Reload() error = %v", err)
			}
			
			if len(result.AgentsAdded) != tt.wantAdded {
				t.Errorf("Reload() added %d agents, want %d", len(result.AgentsAdded), tt.wantAdded)
			}
			
			if len(result.AgentsRemoved) != tt.wantRemoved {
				t.Errorf("Reload() removed %d agents, want %d", len(result.AgentsRemoved), tt.wantRemoved)
			}
			
			if len(result.AgentsUpdated) != tt.wantUpdated {
				t.Errorf("Reload() updated %d agents, want %d", len(result.AgentsUpdated), tt.wantUpdated)
			}
			
			if len(result.Errors) != tt.wantErrors {
				t.Errorf("Reload() had %d errors, want %d: %v", len(result.Errors), tt.wantErrors, result.Errors)
			}
			
			// Verify current config was updated
			if rm.GetCurrentConfig() != tt.newConfig {
				t.Errorf("Reload() did not update current config")
			}
		})
	}
}

func TestReloadManager_BuildAgentEnv(t *testing.T) {
	config := &Config{
		Core: CoreConfig{
			BeadsDBPath: "./test-repo",
		},
		Services: ServicesConfig{
			MCPAgentMail: MCPConfig{
				URL: "http://localhost:8765",
			},
		},
	}
	
	agentConfig := AgentConfig{
		Command: "python agent.py",
		Model:   "claude",
		Phases:  []string{"planning", "implementation"},
	}
	
	envVars := map[string]string{
		"CLAUDE_API_KEY": "test-key-123",
		"OPENAI_API_KEY": "test-key-456",
	}
	
	rm := NewReloadManager(&Config{}, newMockProcessManager(), envVars)
	
	env := rm.buildAgentEnv("test-agent", agentConfig, config)
	
	// Check that all required environment variables are present
	expectedVars := map[string]bool{
		"AGENT_NAME=test-agent":                      false,
		"AGENT_MODEL=claude":                         false,
		"AGENT_PHASES=planning,implementation":       false,
		"MCP_MAIL_URL=http://localhost:8765":         false,
		"BEADS_DB_PATH=./test-repo":                  false,
		"CLAUDE_API_KEY=test-key-123":                false,
		"OPENAI_API_KEY=test-key-456":                false,
	}
	
	for _, envVar := range env {
		if _, exists := expectedVars[envVar]; exists {
			expectedVars[envVar] = true
		}
	}
	
	for varName, found := range expectedVars {
		if !found {
			t.Errorf("buildAgentEnv() missing environment variable: %s", varName)
		}
	}
}
