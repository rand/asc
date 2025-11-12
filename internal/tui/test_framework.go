package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rand/asc/internal/beads"
	"github.com/rand/asc/internal/config"
	"github.com/rand/asc/internal/mcp"
	"github.com/rand/asc/internal/process"
)

// TestFramework provides utilities for testing TUI components
type TestFramework struct {
	model          Model
	beadsClient    *MockBeadsClient
	mcpClient      *MockMCPClient
	procManager    *MockProcessManager
	terminalWidth  int
	terminalHeight int
}

// NewTestFramework creates a new test framework with default mocks
func NewTestFramework() *TestFramework {
	beadsClient := NewMockBeadsClient()
	mcpClient := NewMockMCPClient()
	procManager := NewMockProcessManager()

	cfg := config.Config{
		Core: config.CoreConfig{
			BeadsDBPath: "./test-beads",
		},
		Services: config.ServicesConfig{
			MCPAgentMail: config.MCPConfig{
				StartCommand: "python -m mcp_agent_mail.server",
				URL:          "http://localhost:8765",
			},
		},
		Agents: map[string]config.AgentConfig{
			"test-agent-1": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
			"test-agent-2": {
				Command: "python",
				Model:   "gemini",
				Phases:  []string{"implementation"},
			},
		},
	}

	model := NewModel(cfg, beadsClient, mcpClient, procManager)
	model.width = 120
	model.height = 40

	return &TestFramework{
		model:          model,
		beadsClient:    beadsClient,
		mcpClient:      mcpClient,
		procManager:    procManager,
		terminalWidth:  120,
		terminalHeight: 40,
	}
}

// GetModel returns the current model
func (tf *TestFramework) GetModel() *Model {
	return &tf.model
}

// SendKey simulates a key press
func (tf *TestFramework) SendKey(key tea.KeyType) Model {
	msg := tea.KeyMsg{Type: key}
	newModel, _ := tf.model.Update(msg)
	tf.model = newModel.(Model)
	return tf.model
}

// SendKeyRune simulates a rune key press
func (tf *TestFramework) SendKeyRune(r rune) Model {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
	newModel, _ := tf.model.Update(msg)
	tf.model = newModel.(Model)
	return tf.model
}

// SendKeyString simulates typing a string
func (tf *TestFramework) SendKeyString(s string) Model {
	for _, r := range s {
		tf.SendKeyRune(r)
	}
	return tf.model
}

// Resize simulates a terminal resize
func (tf *TestFramework) Resize(width, height int) Model {
	tf.terminalWidth = width
	tf.terminalHeight = height
	msg := tea.WindowSizeMsg{Width: width, Height: height}
	newModel, _ := tf.model.Update(msg)
	tf.model = newModel.(Model)
	return tf.model
}

// Tick simulates a tick event
func (tf *TestFramework) Tick() Model {
	msg := tickMsg(time.Now())
	newModel, _ := tf.model.Update(msg)
	tf.model = newModel.(Model)
	return tf.model
}

// Render returns the current view
func (tf *TestFramework) Render() string {
	return tf.model.View()
}

// AddTask adds a task to the mock beads client
func (tf *TestFramework) AddTask(task beads.Task) {
	tf.beadsClient.AddTask(task)
}

// AddMessage adds a message to the mock MCP client
func (tf *TestFramework) AddMessage(msg mcp.Message) {
	tf.mcpClient.AddMessage(msg)
}

// SetAgentStatus sets an agent's status in the mock MCP client
func (tf *TestFramework) SetAgentStatus(status mcp.AgentStatus) {
	tf.mcpClient.SetAgentStatus(status)
}

// RefreshData forces a data refresh
func (tf *TestFramework) RefreshData() Model {
	// Manually trigger refresh
	tf.model.agents, _ = tf.mcpClient.GetAllAgentStatuses(2 * time.Minute)
	tf.model.tasks, _ = tf.beadsClient.GetTasks([]string{"open", "in_progress"})
	tf.model.messages, _ = tf.mcpClient.GetMessages(tf.model.lastRefresh)
	tf.model.lastRefresh = time.Now()
	return tf.model
}

// MockBeadsClient is a mock implementation of BeadsClient for testing
type MockBeadsClient struct {
	tasks []beads.Task
}

// NewMockBeadsClient creates a new mock beads client
func NewMockBeadsClient() *MockBeadsClient {
	return &MockBeadsClient{
		tasks: []beads.Task{},
	}
}

// AddTask adds a task to the mock client
func (m *MockBeadsClient) AddTask(task beads.Task) {
	m.tasks = append(m.tasks, task)
}

// GetTasks returns tasks matching the given statuses
func (m *MockBeadsClient) GetTasks(statuses []string) ([]beads.Task, error) {
	if len(statuses) == 0 {
		return m.tasks, nil
	}

	filtered := []beads.Task{}
	for _, task := range m.tasks {
		for _, status := range statuses {
			if task.Status == status {
				filtered = append(filtered, task)
				break
			}
		}
	}
	return filtered, nil
}

// CreateTask creates a new task
func (m *MockBeadsClient) CreateTask(title string) (beads.Task, error) {
	task := beads.Task{
		ID:     "mock-task-id",
		Title:  title,
		Status: "open",
	}
	m.tasks = append(m.tasks, task)
	return task, nil
}

// UpdateTask updates an existing task
func (m *MockBeadsClient) UpdateTask(id string, updates beads.TaskUpdate) error {
	for i, task := range m.tasks {
		if task.ID == id {
			if updates.Assignee != nil {
				m.tasks[i].Assignee = *updates.Assignee
			}
			if updates.Status != nil {
				m.tasks[i].Status = *updates.Status
			}
		}
	}
	return nil
}

// DeleteTask deletes a task
func (m *MockBeadsClient) DeleteTask(id string) error {
	for i, task := range m.tasks {
		if task.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			break
		}
	}
	return nil
}

// Refresh refreshes the task list
func (m *MockBeadsClient) Refresh() error {
	return nil
}

// MockMCPClient is a mock implementation of MCPClient for testing
type MockMCPClient struct {
	messages []mcp.Message
	statuses map[string]mcp.AgentStatus
}

// NewMockMCPClient creates a new mock MCP client
func NewMockMCPClient() *MockMCPClient {
	return &MockMCPClient{
		messages: []mcp.Message{},
		statuses: make(map[string]mcp.AgentStatus),
	}
}

// AddMessage adds a message to the mock client
func (m *MockMCPClient) AddMessage(msg mcp.Message) {
	m.messages = append(m.messages, msg)
}

// SetAgentStatus sets an agent's status
func (m *MockMCPClient) SetAgentStatus(status mcp.AgentStatus) {
	m.statuses[status.Name] = status
}

// GetMessages returns messages since the given time
func (m *MockMCPClient) GetMessages(since time.Time) ([]mcp.Message, error) {
	filtered := []mcp.Message{}
	for _, msg := range m.messages {
		if msg.Timestamp.After(since) || msg.Timestamp.Equal(since) {
			filtered = append(filtered, msg)
		}
	}
	return filtered, nil
}

// SendMessage sends a message
func (m *MockMCPClient) SendMessage(msg mcp.Message) error {
	m.messages = append(m.messages, msg)
	return nil
}

// GetAgentStatus returns an agent's status
func (m *MockMCPClient) GetAgentStatus(agentName string) (mcp.AgentStatus, error) {
	if status, ok := m.statuses[agentName]; ok {
		return status, nil
	}
	return mcp.AgentStatus{
		Name:  agentName,
		State: mcp.StateOffline,
	}, nil
}

// GetAllAgentStatuses returns all agent statuses
func (m *MockMCPClient) GetAllAgentStatuses(offlineThreshold time.Duration) ([]mcp.AgentStatus, error) {
	statuses := []mcp.AgentStatus{}
	for _, status := range m.statuses {
		statuses = append(statuses, status)
	}
	return statuses, nil
}

// ReleaseAgentLeases releases all leases for an agent
func (m *MockMCPClient) ReleaseAgentLeases(agentName string) error {
	return nil
}

// MockProcessManager is a mock implementation of ProcessManager for testing
type MockProcessManager struct {
	processes map[string]*process.ProcessInfo
}

// NewMockProcessManager creates a new mock process manager
func NewMockProcessManager() *MockProcessManager {
	return &MockProcessManager{
		processes: make(map[string]*process.ProcessInfo),
	}
}

// Start starts a process
func (m *MockProcessManager) Start(name string, command string, args []string, env []string) (int, error) {
	pid := 10000 + len(m.processes)
	m.processes[name] = &process.ProcessInfo{
		Name:    name,
		PID:     pid,
		Command: command,
		Args:    args,
		Env:     make(map[string]string),
		LogFile: "/tmp/test.log",
	}
	return pid, nil
}

// Stop stops a process
func (m *MockProcessManager) Stop(pid int) error {
	for name, info := range m.processes {
		if info.PID == pid {
			delete(m.processes, name)
			break
		}
	}
	return nil
}

// StopAll stops all processes
func (m *MockProcessManager) StopAll() error {
	m.processes = make(map[string]*process.ProcessInfo)
	return nil
}

// IsRunning checks if a process is running
func (m *MockProcessManager) IsRunning(pid int) bool {
	for _, info := range m.processes {
		if info.PID == pid {
			return true
		}
	}
	return false
}

// GetStatus returns a process status
func (m *MockProcessManager) GetStatus(pid int) process.ProcessStatus {
	if m.IsRunning(pid) {
		return process.StatusRunning
	}
	return process.StatusStopped
}

// GetProcessInfo returns process information
func (m *MockProcessManager) GetProcessInfo(name string) (*process.ProcessInfo, error) {
	if info, ok := m.processes[name]; ok {
		return info, nil
	}
	return nil, nil
}

// ListProcesses returns all processes
func (m *MockProcessManager) ListProcesses() ([]*process.ProcessInfo, error) {
	list := []*process.ProcessInfo{}
	for _, info := range m.processes {
		list = append(list, info)
	}
	return list, nil
}

// MockTerminal simulates a terminal for testing rendering
type MockTerminal struct {
	width  int
	height int
	output string
}

// NewMockTerminal creates a new mock terminal
func NewMockTerminal(width, height int) *MockTerminal {
	return &MockTerminal{
		width:  width,
		height: height,
	}
}

// Render renders the model to the mock terminal
func (mt *MockTerminal) Render(m Model) string {
	mt.output = m.View()
	return mt.output
}

// GetOutput returns the last rendered output
func (mt *MockTerminal) GetOutput() string {
	return mt.output
}

// GetWidth returns the terminal width
func (mt *MockTerminal) GetWidth() int {
	return mt.width
}

// GetHeight returns the terminal height
func (mt *MockTerminal) GetHeight() int {
	return mt.height
}

// TestHelper provides common test helper functions
type TestHelper struct{}

// NewTestHelper creates a new test helper
func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

// CreateTestTask creates a test task with default values
func (th *TestHelper) CreateTestTask(id, title, status string) beads.Task {
	return beads.Task{
		ID:     id,
		Title:  title,
		Status: status,
		Phase:  "planning",
	}
}

// CreateTestMessage creates a test message with default values
func (th *TestHelper) CreateTestMessage(msgType mcp.MessageType, source, content string) mcp.Message {
	return mcp.Message{
		Timestamp: time.Now(),
		Type:      msgType,
		Source:    source,
		Content:   content,
	}
}

// CreateTestAgentStatus creates a test agent status with default values
func (th *TestHelper) CreateTestAgentStatus(name string, state mcp.AgentState) mcp.AgentStatus {
	return mcp.AgentStatus{
		Name:     name,
		State:    state,
		LastSeen: time.Now(),
	}
}

// CreateTestConfig creates a test configuration
func (th *TestHelper) CreateTestConfig() config.Config {
	return config.Config{
		Core: config.CoreConfig{
			BeadsDBPath: "./test-beads",
		},
		Services: config.ServicesConfig{
			MCPAgentMail: config.MCPConfig{
				StartCommand: "python -m mcp_agent_mail.server",
				URL:          "http://localhost:8765",
			},
		},
		Agents: map[string]config.AgentConfig{
			"test-agent": {
				Command: "python",
				Model:   "claude",
				Phases:  []string{"planning"},
			},
		},
	}
}

// AssertContains checks if a string contains a substring
func (th *TestHelper) AssertContains(t interface{ Errorf(format string, args ...interface{}) }, haystack, needle string) bool {
	if !contains(haystack, needle) {
		t.Errorf("Expected output to contain %q, but it didn't", needle)
		return false
	}
	return true
}

// AssertNotContains checks if a string does not contain a substring
func (th *TestHelper) AssertNotContains(t interface{ Errorf(format string, args ...interface{}) }, haystack, needle string) bool {
	if contains(haystack, needle) {
		t.Errorf("Expected output to not contain %q, but it did", needle)
		return false
	}
	return true
}

// contains checks if a string contains a substring
func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || len(needle) == 0 || indexOfSubstring(haystack, needle) >= 0)
}

// indexOfSubstring finds the index of a substring
func indexOfSubstring(haystack, needle string) int {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
