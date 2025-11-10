package health

import (
	"fmt"
	"testing"
	"time"

	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/mcp"
	"github.com/yourusername/asc/internal/process"
)

// Mock MCP Client
type mockMCPClient struct {
	statuses []mcp.AgentStatus
	err      error
}

func (m *mockMCPClient) GetMessages(since time.Time) ([]mcp.Message, error) {
	return nil, nil
}

func (m *mockMCPClient) SendMessage(msg mcp.Message) error {
	return nil
}

func (m *mockMCPClient) GetAgentStatus(agentName string) (mcp.AgentStatus, error) {
	return mcp.AgentStatus{}, nil
}

func (m *mockMCPClient) GetAllAgentStatuses(offlineThreshold time.Duration) ([]mcp.AgentStatus, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.statuses, nil
}

// Mock Process Manager
type mockProcessManager struct {
	processes map[string]*process.ProcessInfo
	running   map[int]bool
}

func (m *mockProcessManager) Start(name string, command string, args []string, env []string) (int, error) {
	return 0, nil
}

func (m *mockProcessManager) Stop(pid int) error {
	return nil
}

func (m *mockProcessManager) StopAll() error {
	return nil
}

func (m *mockProcessManager) IsRunning(pid int) bool {
	return m.running[pid]
}

func (m *mockProcessManager) GetStatus(pid int) process.ProcessStatus {
	if m.running[pid] {
		return process.StatusRunning
	}
	return process.StatusStopped
}

func (m *mockProcessManager) GetProcessInfo(name string) (*process.ProcessInfo, error) {
	if info, ok := m.processes[name]; ok {
		return info, nil
	}
	return nil, fmt.Errorf("process not found: %s", name)
}

func (m *mockProcessManager) ListProcesses() ([]*process.ProcessInfo, error) {
	var list []*process.ProcessInfo
	for _, info := range m.processes {
		list = append(list, info)
	}
	return list, nil
}

func TestNewMonitor(t *testing.T) {
	cfg := config.Config{
		Agents: map[string]config.AgentConfig{
			"test-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}

	mcpClient := &mockMCPClient{}
	procManager := &mockProcessManager{
		processes: make(map[string]*process.ProcessInfo),
		running:   make(map[int]bool),
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	if len(monitor.agentStates) != 1 {
		t.Errorf("Expected 1 agent state, got %d", len(monitor.agentStates))
	}

	if _, exists := monitor.agentStates["test-agent"]; !exists {
		t.Error("Expected test-agent to be in agent states")
	}
}

func TestDetectCrashedAgent(t *testing.T) {
	cfg := config.Config{
		Agents: map[string]config.AgentConfig{
			"crashed-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}

	mcpClient := &mockMCPClient{
		statuses: []mcp.AgentStatus{},
	}

	procManager := &mockProcessManager{
		processes: make(map[string]*process.ProcessInfo),
		running:   make(map[int]bool),
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	// Perform health check
	monitor.performHealthCheck()

	issues := monitor.GetHealthIssues()
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	if issues[0].Type != IssueCrashed {
		t.Errorf("Expected crashed issue, got %s", issues[0].Type)
	}

	if issues[0].AgentName != "crashed-agent" {
		t.Errorf("Expected crashed-agent, got %s", issues[0].AgentName)
	}
}

func TestDetectUnresponsiveAgent(t *testing.T) {
	cfg := config.Config{
		Agents: map[string]config.AgentConfig{
			"unresponsive-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}

	// Agent with old heartbeat
	oldTime := time.Now().Add(-5 * time.Minute)
	mcpClient := &mockMCPClient{
		statuses: []mcp.AgentStatus{
			{
				Name:     "unresponsive-agent",
				State:    mcp.StateOffline,
				LastSeen: oldTime,
			},
		},
	}

	procManager := &mockProcessManager{
		processes: map[string]*process.ProcessInfo{
			"unresponsive-agent": {
				Name: "unresponsive-agent",
				PID:  12345,
			},
		},
		running: map[int]bool{
			12345: true,
		},
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	// Perform health check
	monitor.performHealthCheck()

	issues := monitor.GetHealthIssues()
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	if issues[0].Type != IssueUnresponsive {
		t.Errorf("Expected unresponsive issue, got %s", issues[0].Type)
	}
}

func TestDetectStuckAgent(t *testing.T) {
	cfg := config.Config{
		Agents: map[string]config.AgentConfig{
			"stuck-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}

	mcpClient := &mockMCPClient{
		statuses: []mcp.AgentStatus{
			{
				Name:        "stuck-agent",
				State:       mcp.StateWorking,
				CurrentTask: "task-123",
				LastSeen:    time.Now(),
			},
		},
	}

	procManager := &mockProcessManager{
		processes: map[string]*process.ProcessInfo{
			"stuck-agent": {
				Name: "stuck-agent",
				PID:  12345,
			},
		},
		running: map[int]bool{
			12345: true,
		},
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	// Set up agent state to simulate stuck task
	monitor.mu.Lock()
	monitor.agentStates["stuck-agent"].LastTask = "task-123"
	monitor.agentStates["stuck-agent"].TaskStartTime = time.Now().Add(-45 * time.Minute)
	monitor.mu.Unlock()

	// Perform health check
	monitor.performHealthCheck()

	issues := monitor.GetHealthIssues()
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	if issues[0].Type != IssueStuck {
		t.Errorf("Expected stuck issue, got %s", issues[0].Type)
	}
}

func TestHealthyAgents(t *testing.T) {
	cfg := config.Config{
		Agents: map[string]config.AgentConfig{
			"healthy-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}

	mcpClient := &mockMCPClient{
		statuses: []mcp.AgentStatus{
			{
				Name:     "healthy-agent",
				State:    mcp.StateIdle,
				LastSeen: time.Now(),
			},
		},
	}

	procManager := &mockProcessManager{
		processes: map[string]*process.ProcessInfo{
			"healthy-agent": {
				Name: "healthy-agent",
				PID:  12345,
			},
		},
		running: map[int]bool{
			12345: true,
		},
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	// Perform health check
	monitor.performHealthCheck()

	issues := monitor.GetHealthIssues()
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for healthy agent, got %d", len(issues))
	}

	if !monitor.IsHealthy() {
		t.Error("Expected monitor to report healthy")
	}
}

func TestGetHealthSummary(t *testing.T) {
	cfg := config.Config{
		Agents: map[string]config.AgentConfig{
			"agent1": {Command: "python"},
			"agent2": {Command: "python"},
		},
	}

	mcpClient := &mockMCPClient{}
	procManager := &mockProcessManager{
		processes: make(map[string]*process.ProcessInfo),
		running:   make(map[int]bool),
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	// Add some issues manually
	monitor.mu.Lock()
	monitor.healthIssues = []HealthIssue{
		{Type: IssueCrashed, Severity: "critical"},
		{Type: IssueStuck, Severity: "warning"},
	}
	monitor.mu.Unlock()

	summary := monitor.GetHealthSummary()
	expected := "1 critical, 1 warning"
	if summary != expected {
		t.Errorf("Expected summary '%s', got '%s'", expected, summary)
	}
}

func TestAutoRecoveryConfiguration(t *testing.T) {
	autoRecoveryFalse := false
	cfg := config.Config{
		Core: config.CoreConfig{
			AutoRecovery: &autoRecoveryFalse,
		},
		Agents: map[string]config.AgentConfig{
			"test-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}

	mcpClient := &mockMCPClient{}
	procManager := &mockProcessManager{
		processes: make(map[string]*process.ProcessInfo),
		running:   make(map[int]bool),
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	// Auto-recovery should be enabled by default
	if !monitor.IsAutoRecoveryEnabled() {
		t.Error("Expected auto-recovery to be enabled by default")
	}

	// Disable auto-recovery
	monitor.SetAutoRecovery(false)
	if monitor.IsAutoRecoveryEnabled() {
		t.Error("Expected auto-recovery to be disabled")
	}

	// Enable auto-recovery
	monitor.SetAutoRecovery(true)
	if !monitor.IsAutoRecoveryEnabled() {
		t.Error("Expected auto-recovery to be enabled")
	}
}

func TestRecoveryActions(t *testing.T) {
	cfg := config.Config{
		Agents: map[string]config.AgentConfig{
			"test-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}

	mcpClient := &mockMCPClient{}
	procManager := &mockProcessManager{
		processes: make(map[string]*process.ProcessInfo),
		running:   make(map[int]bool),
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	// Initially, there should be no recovery actions
	actions := monitor.GetRecoveryActions()
	if len(actions) != 0 {
		t.Errorf("Expected 0 recovery actions, got %d", len(actions))
	}

	// Record a recovery action
	monitor.mu.Lock()
	monitor.recordRecoveryAction("test-agent", "restart", "crashed", true, "")
	monitor.mu.Unlock()

	// Verify the action was recorded
	actions = monitor.GetRecoveryActions()
	if len(actions) != 1 {
		t.Fatalf("Expected 1 recovery action, got %d", len(actions))
	}

	action := actions[0]
	if action.AgentName != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", action.AgentName)
	}
	if action.Action != "restart" {
		t.Errorf("Expected action 'restart', got '%s'", action.Action)
	}
	if action.Reason != "crashed" {
		t.Errorf("Expected reason 'crashed', got '%s'", action.Reason)
	}
	if !action.Success {
		t.Error("Expected action to be successful")
	}
}

func TestRecoverySuccessRate(t *testing.T) {
	cfg := config.Config{
		Agents: map[string]config.AgentConfig{
			"test-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}

	mcpClient := &mockMCPClient{}
	procManager := &mockProcessManager{
		processes: make(map[string]*process.ProcessInfo),
		running:   make(map[int]bool),
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	// Initially, success rate should be 0
	rate := monitor.GetRecoverySuccessRate("test-agent")
	if rate != 0.0 {
		t.Errorf("Expected success rate 0.0, got %f", rate)
	}

	// Record some recovery attempts
	monitor.mu.Lock()
	stats := monitor.getOrCreateRecoveryStats("test-agent")
	monitor.updateRecoveryStats(stats, true)  // Success
	monitor.updateRecoveryStats(stats, true)  // Success
	monitor.updateRecoveryStats(stats, false) // Failure
	monitor.mu.Unlock()

	// Success rate should be 2/3 = 0.666...
	rate = monitor.GetRecoverySuccessRate("test-agent")
	expected := 2.0 / 3.0
	if rate < expected-0.01 || rate > expected+0.01 {
		t.Errorf("Expected success rate ~%f, got %f", expected, rate)
	}
}

func TestRecoveryBackoff(t *testing.T) {
	cfg := config.Config{
		Agents: map[string]config.AgentConfig{
			"test-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}

	mcpClient := &mockMCPClient{}
	procManager := &mockProcessManager{
		processes: make(map[string]*process.ProcessInfo),
		running:   make(map[int]bool),
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	monitor.mu.Lock()
	stats := monitor.getOrCreateRecoveryStats("test-agent")
	
	// First failure should set backoff to 1 minute
	monitor.updateRecoveryStats(stats, false)
	if stats.BackoffUntil.IsZero() {
		t.Error("Expected backoff to be set after failure")
	}
	
	// Backoff should be approximately 1 minute from now
	expectedBackoff := time.Now().Add(1 * time.Minute)
	if stats.BackoffUntil.Before(expectedBackoff.Add(-5*time.Second)) || 
	   stats.BackoffUntil.After(expectedBackoff.Add(5*time.Second)) {
		t.Errorf("Expected backoff around %v, got %v", expectedBackoff, stats.BackoffUntil)
	}
	
	// Success should reset backoff
	monitor.updateRecoveryStats(stats, true)
	if !stats.BackoffUntil.IsZero() {
		t.Error("Expected backoff to be reset after success")
	}
	monitor.mu.Unlock()
}

func TestGetRecoveryStats(t *testing.T) {
	cfg := config.Config{
		Agents: map[string]config.AgentConfig{
			"test-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}

	mcpClient := &mockMCPClient{}
	procManager := &mockProcessManager{
		processes: make(map[string]*process.ProcessInfo),
		running:   make(map[int]bool),
	}

	monitor, err := NewMonitor(mcpClient, procManager, cfg)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Stop()

	// Initially, stats should be nil
	stats := monitor.GetRecoveryStats("test-agent")
	if stats != nil {
		t.Error("Expected nil stats for agent with no recovery attempts")
	}

	// Record a recovery attempt
	monitor.mu.Lock()
	internalStats := monitor.getOrCreateRecoveryStats("test-agent")
	monitor.updateRecoveryStats(internalStats, true)
	monitor.mu.Unlock()

	// Now stats should exist
	stats = monitor.GetRecoveryStats("test-agent")
	if stats == nil {
		t.Fatal("Expected stats to exist after recovery attempt")
	}

	if stats.TotalAttempts != 1 {
		t.Errorf("Expected 1 total attempt, got %d", stats.TotalAttempts)
	}
	if stats.SuccessfulCount != 1 {
		t.Errorf("Expected 1 successful attempt, got %d", stats.SuccessfulCount)
	}
}

// Add ReleaseAgentLeases to mock MCP client
func (m *mockMCPClient) ReleaseAgentLeases(agentName string) error {
	return nil
}
