package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yourusername/asc/internal/beads"
	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/mcp"
	"github.com/yourusername/asc/internal/process"
)

// mockBeadsClient is a mock implementation of BeadsClient for testing
type mockBeadsClient struct {
	tasks []beads.Task
}

func (m *mockBeadsClient) GetTasks(statuses []string) ([]beads.Task, error) {
	return m.tasks, nil
}

func (m *mockBeadsClient) CreateTask(title string) (beads.Task, error) {
	task := beads.Task{
		ID:     "test-123",
		Title:  title,
		Status: "open",
	}
	m.tasks = append(m.tasks, task)
	return task, nil
}

func (m *mockBeadsClient) UpdateTask(id string, updates beads.TaskUpdate) error {
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

func (m *mockBeadsClient) DeleteTask(id string) error {
	for i, task := range m.tasks {
		if task.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockBeadsClient) Refresh() error {
	return nil
}

// mockMCPClient is a mock implementation of MCPClient for testing
type mockMCPClient struct {
	messages []mcp.Message
}

func (m *mockMCPClient) GetMessages(since time.Time) ([]mcp.Message, error) {
	return m.messages, nil
}

func (m *mockMCPClient) SendMessage(msg mcp.Message) error {
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockMCPClient) GetAgentStatus(agentName string) (mcp.AgentStatus, error) {
	return mcp.AgentStatus{
		Name:  agentName,
		State: mcp.StateIdle,
	}, nil
}

// mockProcessManager is a mock implementation of ProcessManager for testing
type mockProcessManager struct{}

func (m *mockProcessManager) Start(name string, command string, args []string, env []string) (int, error) {
	return 12345, nil
}

func (m *mockProcessManager) Stop(pid int) error {
	return nil
}

func (m *mockProcessManager) StopAll() error {
	return nil
}

func (m *mockProcessManager) IsRunning(pid int) bool {
	return true
}

func (m *mockProcessManager) GetStatus(pid int) process.ProcessStatus {
	return process.StatusRunning
}

func (m *mockProcessManager) GetProcessInfo(name string) (*process.ProcessInfo, error) {
	return &process.ProcessInfo{
		Name:    name,
		PID:     12345,
		LogFile: "/tmp/test.log",
		Command: "python",
		Args:    []string{"agent.py"},
		Env:     map[string]string{},
	}, nil
}

func (m *mockProcessManager) ListProcesses() ([]*process.ProcessInfo, error) {
	return []*process.ProcessInfo{}, nil
}

// createTestModel creates a model for testing
func createTestModel() Model {
	cfg := config.Config{
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

	beadsClient := &mockBeadsClient{
		tasks: []beads.Task{
			{ID: "1", Title: "Task 1", Status: "open"},
			{ID: "2", Title: "Task 2", Status: "in_progress"},
			{ID: "3", Title: "Task 3", Status: "open"},
		},
	}

	mcpClient := &mockMCPClient{
		messages: []mcp.Message{
			{
				Timestamp: time.Now(),
				Type:      mcp.TypeMessage,
				Source:    "test-agent",
				Content:   "Test message",
			},
		},
	}

	procManager := &mockProcessManager{}

	return NewModel(cfg, beadsClient, mcpClient, procManager)
}

// TestTaskNavigation tests arrow key navigation in task list
func TestTaskNavigation(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40
	
	// Load tasks into model
	m.tasks = []beads.Task{
		{ID: "1", Title: "Task 1", Status: "open"},
		{ID: "2", Title: "Task 2", Status: "in_progress"},
		{ID: "3", Title: "Task 3", Status: "open"},
	}

	// Initial selection should be 0
	if m.selectedTaskIndex != 0 {
		t.Errorf("Expected initial selectedTaskIndex to be 0, got %d", m.selectedTaskIndex)
	}

	// Get filtered tasks to verify we have tasks to navigate
	filteredTasks := m.filterTasksByStatus([]string{"open", "in_progress"})
	if len(filteredTasks) < 2 {
		t.Fatalf("Expected at least 2 filtered tasks, got %d", len(filteredTasks))
	}

	// Press down arrow
	downKey := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(downKey)
	m = newModel.(Model)

	if m.selectedTaskIndex != 1 {
		t.Errorf("Expected selectedTaskIndex to be 1 after down, got %d", m.selectedTaskIndex)
	}

	// Press up arrow
	upKey := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(upKey)
	m = newModel.(Model)

	if m.selectedTaskIndex != 0 {
		t.Errorf("Expected selectedTaskIndex to be 0 after up, got %d", m.selectedTaskIndex)
	}
}

// TestTaskModalToggle tests opening and closing task detail modal
func TestTaskModalToggle(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40

	// Initially modal should be closed
	if m.showTaskModal {
		t.Error("Expected showTaskModal to be false initially")
	}

	// Press 'v' to open modal
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = newModel.(Model)

	if !m.showTaskModal {
		t.Error("Expected showTaskModal to be true after pressing 'v'")
	}

	// Press 'v' again to close modal
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = newModel.(Model)

	if m.showTaskModal {
		t.Error("Expected showTaskModal to be false after pressing 'v' again")
	}
}

// TestCreateTaskModal tests opening create task modal
func TestCreateTaskModal(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40

	// Initially modal should be closed
	if m.showCreateModal {
		t.Error("Expected showCreateModal to be false initially")
	}

	// Press 'n' to open create modal
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newModel.(Model)

	if !m.showCreateModal {
		t.Error("Expected showCreateModal to be true after pressing 'n'")
	}

	// Input should be empty
	if m.createTaskInput != "" {
		t.Errorf("Expected createTaskInput to be empty, got %s", m.createTaskInput)
	}
}

// TestAgentSelection tests number key selection for agents
func TestAgentSelection(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40

	// Initial selection should be 0
	if m.selectedAgentIndex != 0 {
		t.Errorf("Expected initial selectedAgentIndex to be 0, got %d", m.selectedAgentIndex)
	}

	// Press '1' to select first agent
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = newModel.(Model)

	if m.selectedAgentIndex != 0 {
		t.Errorf("Expected selectedAgentIndex to be 0 after pressing '1', got %d", m.selectedAgentIndex)
	}

	// Press '2' to select second agent
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = newModel.(Model)

	if m.selectedAgentIndex != 1 {
		t.Errorf("Expected selectedAgentIndex to be 1 after pressing '2', got %d", m.selectedAgentIndex)
	}
}

// TestConfirmModal tests confirmation dialog for destructive actions
func TestConfirmModal(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40

	// Initially modal should be closed
	if m.showConfirmModal {
		t.Error("Expected showConfirmModal to be false initially")
	}

	// Press 'k' to trigger kill confirmation
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)

	if !m.showConfirmModal {
		t.Error("Expected showConfirmModal to be true after pressing 'k'")
	}

	if m.confirmAction != "kill" {
		t.Errorf("Expected confirmAction to be 'kill', got %s", m.confirmAction)
	}

	// Press 'n' to cancel
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newModel.(Model)

	if m.showConfirmModal {
		t.Error("Expected showConfirmModal to be false after pressing 'n'")
	}
}

// TestSearchMode tests entering and exiting search mode
func TestSearchMode(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40

	// Initially search mode should be off
	if m.searchMode {
		t.Error("Expected searchMode to be false initially")
	}

	// Press '/' to enter search mode
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newModel.(Model)

	if !m.searchMode {
		t.Error("Expected searchMode to be true after pressing '/'")
	}

	// Input should be empty
	if m.searchInput != "" {
		t.Errorf("Expected searchInput to be empty, got %s", m.searchInput)
	}

	// Press 'esc' to exit search mode
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(Model)

	if m.searchMode {
		t.Error("Expected searchMode to be false after pressing 'esc'")
	}
}

// TestFilterCycling tests cycling through filters
func TestFilterCycling(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40

	// Initially no filters should be set
	if m.logFilterAgent != "" {
		t.Errorf("Expected logFilterAgent to be empty, got %s", m.logFilterAgent)
	}

	// Press 'a' to cycle agent filter
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = newModel.(Model)

	// Should now have an agent selected (any agent is fine since map order is not guaranteed)
	agentNames := m.getAgentNames()
	if len(agentNames) > 0 {
		found := false
		for _, name := range agentNames {
			if m.logFilterAgent == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected logFilterAgent to be one of %v, got %s", agentNames, m.logFilterAgent)
		}
	}

	// Press 'x' to clear filters
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = newModel.(Model)

	if m.logFilterAgent != "" {
		t.Errorf("Expected logFilterAgent to be empty after clearing, got %s", m.logFilterAgent)
	}
}
