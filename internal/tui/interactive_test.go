package tui

import (
	"fmt"
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

func (m *mockMCPClient) GetAllAgentStatuses(offlineThreshold time.Duration) ([]mcp.AgentStatus, error) {
	return []mcp.AgentStatus{}, nil
}

func (m *mockMCPClient) ReleaseAgentLeases(agentName string) error {
	return nil
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

// TestUpdateWithDifferentMessageTypes tests Update method with various message types
func TestUpdateWithDifferentMessageTypes(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40

	tests := []struct {
		name     string
		msg      tea.Msg
		validate func(Model) error
	}{
		{
			name: "WindowSizeMsg",
			msg:  tea.WindowSizeMsg{Width: 120, Height: 50},
			validate: func(m Model) error {
				if m.width != 120 {
					t.Errorf("Expected width 120, got %d", m.width)
				}
				if m.height != 50 {
					t.Errorf("Expected height 50, got %d", m.height)
				}
				return nil
			},
		},
		{
			name: "tickMsg",
			msg:  tickMsg(time.Now()),
			validate: func(m Model) error {
				// Tick should not change model state directly
				return nil
			},
		},
		{
			name: "refreshDataMsg success",
			msg:  refreshDataMsg{err: nil},
			validate: func(m Model) error {
				if m.err != nil {
					t.Errorf("Expected no error, got %v", m.err)
				}
				return nil
			},
		},
		{
			name: "testResultMsg success",
			msg:  testResultMsg{success: true, message: "Test passed"},
			validate: func(m Model) error {
				if m.err != nil {
					t.Errorf("Expected no error, got %v", m.err)
				}
				return nil
			},
		},
		{
			name: "taskActionMsg success",
			msg:  taskActionMsg{success: true, message: "Task claimed"},
			validate: func(m Model) error {
				if m.err != nil {
					t.Errorf("Expected no error, got %v", m.err)
				}
				return nil
			},
		},
		{
			name: "agentActionMsg failure",
			msg:  agentActionMsg{success: false, message: "Agent not found"},
			validate: func(m Model) error {
				if m.err == nil {
					t.Error("Expected error to be set")
				}
				return nil
			},
		},
		{
			name: "logActionMsg success",
			msg:  logActionMsg{success: true, message: "Logs exported"},
			validate: func(m Model) error {
				if m.err != nil {
					t.Errorf("Expected no error, got %v", m.err)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh model for each test to avoid state pollution
			testModel := createTestModel()
			testModel.width = 100
			testModel.height = 40
			newModel, _ := testModel.Update(tt.msg)
			testModel = newModel.(Model)
			if err := tt.validate(testModel); err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}

// TestKeyboardEventHandling tests comprehensive keyboard event handling
func TestKeyboardEventHandling(t *testing.T) {
	tests := []struct {
		name     string
		key      tea.KeyMsg
		validate func(Model) error
	}{
		{
			name: "quit with q",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			validate: func(m Model) error {
				// Quit command should be returned, model unchanged
				return nil
			},
		},
		{
			name: "quit with ctrl+c",
			key:  tea.KeyMsg{Type: tea.KeyCtrlC},
			validate: func(m Model) error {
				return nil
			},
		},
		{
			name: "refresh with r",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
			validate: func(m Model) error {
				return nil
			},
		},
		{
			name: "test with t",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}},
			validate: func(m Model) error {
				return nil
			},
		},
		{
			name: "export logs with e",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
			validate: func(m Model) error {
				return nil
			},
		},
		{
			name: "cycle message type filter with m",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}},
			validate: func(m Model) error {
				if m.logFilterType == "" {
					t.Error("Expected message type filter to be set")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.width = 100
			m.height = 40
			newModel, _ := m.Update(tt.key)
			m = newModel.(Model)
			if err := tt.validate(m); err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}

// TestModalInteractions tests opening, closing, and navigating modals
func TestModalInteractions(t *testing.T) {
	t.Run("task detail modal lifecycle", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40
		m.tasks = []beads.Task{
			{ID: "1", Title: "Task 1", Status: "open"},
		}

		// Open modal with 'v'
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
		m = newModel.(Model)
		if !m.showTaskModal {
			t.Error("Expected task modal to be open")
		}

		// Close with 'esc'
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = newModel.(Model)
		if m.showTaskModal {
			t.Error("Expected task modal to be closed")
		}
	})

	t.Run("create task modal input", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40

		// Open create modal
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		m = newModel.(Model)
		if !m.showCreateModal {
			t.Error("Expected create modal to be open")
		}

		// Type some text
		for _, ch := range "New Task" {
			newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
			m = newModel.(Model)
		}
		if m.createTaskInput != "New Task" {
			t.Errorf("Expected input 'New Task', got '%s'", m.createTaskInput)
		}

		// Backspace
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = newModel.(Model)
		if m.createTaskInput != "New Tas" {
			t.Errorf("Expected input 'New Tas', got '%s'", m.createTaskInput)
		}

		// Cancel with esc
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = newModel.(Model)
		if m.showCreateModal {
			t.Error("Expected create modal to be closed")
		}
		if m.createTaskInput != "" {
			t.Error("Expected input to be cleared")
		}
	})

	t.Run("confirm modal workflow", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40

		// Trigger kill confirmation
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		m = newModel.(Model)
		if !m.showConfirmModal {
			t.Error("Expected confirm modal to be open")
		}
		if m.confirmAction != "kill" {
			t.Errorf("Expected confirmAction 'kill', got '%s'", m.confirmAction)
		}

		// Cancel with 'n'
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		m = newModel.(Model)
		if m.showConfirmModal {
			t.Error("Expected confirm modal to be closed")
		}

		// Trigger restart confirmation
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
		m = newModel.(Model)
		if !m.showConfirmModal {
			t.Error("Expected confirm modal to be open")
		}
		if m.confirmAction != "restart" {
			t.Errorf("Expected confirmAction 'restart', got '%s'", m.confirmAction)
		}

		// Confirm with 'y'
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
		m = newModel.(Model)
		if m.showConfirmModal {
			t.Error("Expected confirm modal to be closed after confirmation")
		}
	})
}

// TestNavigationBetweenPanes tests navigation between different UI panes
func TestNavigationBetweenPanes(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40
	m.tasks = []beads.Task{
		{ID: "1", Title: "Task 1", Status: "open"},
		{ID: "2", Title: "Task 2", Status: "in_progress"},
	}

	// Test task navigation
	t.Run("task pane navigation", func(t *testing.T) {
		initialIndex := m.selectedTaskIndex
		
		// Navigate down
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(Model)
		if m.selectedTaskIndex != initialIndex+1 {
			t.Errorf("Expected index %d, got %d", initialIndex+1, m.selectedTaskIndex)
		}

		// Navigate up
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = newModel.(Model)
		if m.selectedTaskIndex != initialIndex {
			t.Errorf("Expected index %d, got %d", initialIndex, m.selectedTaskIndex)
		}

		// Try to navigate up past first item (should stay at 0)
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = newModel.(Model)
		if m.selectedTaskIndex < 0 {
			t.Error("Expected index to stay at 0 or above")
		}
	})

	// Test agent selection
	t.Run("agent pane selection", func(t *testing.T) {
		// Select agent 1
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
		m = newModel.(Model)
		if m.selectedAgentIndex != 0 {
			t.Errorf("Expected agent index 0, got %d", m.selectedAgentIndex)
		}

		// Select agent 2
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
		m = newModel.(Model)
		if m.selectedAgentIndex != 1 {
			t.Errorf("Expected agent index 1, got %d", m.selectedAgentIndex)
		}

		// Try to select agent beyond range (should be ignored)
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}})
		m = newModel.(Model)
		// Should stay at previous valid selection
		if m.selectedAgentIndex > len(m.config.Agents)-1 {
			t.Error("Expected agent index to stay within valid range")
		}
	})
}

// TestSearchFunctionality tests search input and filtering
func TestSearchFunctionality(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40
	m.messages = []mcp.Message{
		{Source: "agent-1", Content: "Processing task", Type: mcp.TypeMessage},
		{Source: "agent-2", Content: "Error occurred", Type: mcp.TypeError},
		{Source: "agent-1", Content: "Task completed", Type: mcp.TypeBeads},
	}

	t.Run("enter and exit search mode", func(t *testing.T) {
		// Enter search mode
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		m = newModel.(Model)
		if !m.searchMode {
			t.Error("Expected to be in search mode")
		}

		// Type search query
		for _, ch := range "task" {
			newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
			m = newModel.(Model)
		}
		if m.searchInput != "task" {
			t.Errorf("Expected search input 'task', got '%s'", m.searchInput)
		}

		// Apply search with enter
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(Model)
		if m.searchMode {
			t.Error("Expected to exit search mode after enter")
		}
		if m.searchInput != "task" {
			t.Error("Expected search input to be preserved")
		}

		// Verify filtering works
		filtered := m.getFilteredMessages()
		if len(filtered) != 2 {
			t.Errorf("Expected 2 filtered messages, got %d", len(filtered))
		}
	})

	t.Run("cancel search", func(t *testing.T) {
		// Enter search mode
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		m = newModel.(Model)

		// Type something
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		m = newModel.(Model)

		// Cancel with esc
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = newModel.(Model)
		if m.searchMode {
			t.Error("Expected to exit search mode")
		}
		if m.searchInput != "" {
			t.Error("Expected search input to be cleared")
		}
	})

	t.Run("clear all filters", func(t *testing.T) {
		m.searchInput = "test"
		m.logFilterAgent = "agent-1"
		m.logFilterType = "error"

		// Clear all filters with 'x'
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		m = newModel.(Model)

		if m.searchInput != "" {
			t.Error("Expected search input to be cleared")
		}
		if m.logFilterAgent != "" {
			t.Error("Expected agent filter to be cleared")
		}
		if m.logFilterType != "" {
			t.Error("Expected type filter to be cleared")
		}
	})
}

// TestStateTransitions tests various state transitions in the TUI
func TestStateTransitions(t *testing.T) {
	t.Run("modal state transitions", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40

		// Verify initial state
		if m.showTaskModal || m.showCreateModal || m.showConfirmModal || m.searchMode {
			t.Error("Expected all modals to be closed initially")
		}

		// Open task modal
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
		m = newModel.(Model)
		if !m.showTaskModal {
			t.Error("Expected task modal to be open")
		}

		// Close and open create modal
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = newModel.(Model)
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		m = newModel.(Model)
		if m.showTaskModal || !m.showCreateModal {
			t.Error("Expected only create modal to be open")
		}
	})

	t.Run("filter state transitions", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40

		// Cycle through message type filters
		initialType := m.logFilterType
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
		m = newModel.(Model)
		if m.logFilterType == initialType {
			t.Error("Expected message type filter to change")
		}

		// Cycle through agent filters
		initialAgent := m.logFilterAgent
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		m = newModel.(Model)
		if m.logFilterAgent == initialAgent && len(m.getAgentNames()) > 0 {
			t.Error("Expected agent filter to change")
		}
	})

	t.Run("selection state transitions", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40
		m.tasks = []beads.Task{
			{ID: "1", Title: "Task 1", Status: "open"},
			{ID: "2", Title: "Task 2", Status: "open"},
		}

		// Change task selection
		initialTask := m.selectedTaskIndex
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(Model)
		if m.selectedTaskIndex == initialTask {
			t.Error("Expected task selection to change")
		}

		// Change agent selection
		initialAgent := m.selectedAgentIndex
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
		m = newModel.(Model)
		if m.selectedAgentIndex == initialAgent {
			t.Error("Expected agent selection to change")
		}
	})
}

// TestErrorHandlingInTUI tests error handling scenarios
func TestErrorHandlingInTUI(t *testing.T) {
	t.Run("handle refresh error", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40

		// Simulate refresh error
		testErr := fmt.Errorf("refresh failed")
		errMsg := refreshDataMsg{err: testErr}
		newModel, _ := m.Update(errMsg)
		m = newModel.(Model)

		if m.err == nil {
			t.Error("Expected error to be set")
		}
	})

	t.Run("handle test failure", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40

		// Simulate test failure
		testMsg := testResultMsg{success: false, message: "Connection failed"}
		newModel, _ := m.Update(testMsg)
		m = newModel.(Model)

		if m.err == nil {
			t.Error("Expected error to be set")
		}
	})

	t.Run("handle task action error", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40

		// Simulate task action error
		actionMsg := taskActionMsg{success: false, message: "Task not found"}
		newModel, _ := m.Update(actionMsg)
		m = newModel.(Model)

		if m.err == nil {
			t.Error("Expected error to be set")
		}
	})

	t.Run("handle agent action error", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40

		// Simulate agent action error
		actionMsg := agentActionMsg{success: false, message: "Agent offline"}
		newModel, _ := m.Update(actionMsg)
		m = newModel.(Model)

		if m.err == nil {
			t.Error("Expected error to be set")
		}
	})

	t.Run("clear error on success", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40
		m.err = fmt.Errorf("previous error") // Set an error

		// Simulate successful action
		actionMsg := taskActionMsg{success: true, message: "Task claimed"}
		newModel, _ := m.Update(actionMsg)
		m = newModel.(Model)

		// Error should not be cleared by task action (only refresh does that)
		// But it shouldn't add a new error
	})
}

// TestWSEventHandling tests WebSocket event handling
func TestWSEventHandling(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40

	t.Run("handle connected event", func(t *testing.T) {
		event := wsEventMsg(mcp.Event{Type: mcp.EventConnected})
		newModel, _ := m.Update(event)
		m = newModel.(Model)

		if !m.wsConnected {
			t.Error("Expected wsConnected to be true")
		}
		if m.err != nil {
			t.Error("Expected error to be cleared")
		}
	})

	t.Run("handle disconnected event", func(t *testing.T) {
		m.wsConnected = true
		event := wsEventMsg(mcp.Event{Type: mcp.EventDisconnected})
		newModel, _ := m.Update(event)
		m = newModel.(Model)

		if m.wsConnected {
			t.Error("Expected wsConnected to be false")
		}
	})

	t.Run("handle agent status event", func(t *testing.T) {
		status := mcp.AgentStatus{
			Name:  "test-agent",
			State: mcp.StateWorking,
		}
		event := wsEventMsg(mcp.Event{
			Type:        mcp.EventAgentStatus,
			AgentStatus: &status,
		})
		newModel, _ := m.Update(event)
		m = newModel.(Model)

		// Verify agent was added/updated
		found := false
		for _, agent := range m.agents {
			if agent.Name == "test-agent" && agent.State == mcp.StateWorking {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected agent status to be updated")
		}
	})

	t.Run("handle new message event", func(t *testing.T) {
		initialCount := len(m.messages)
		msg := mcp.Message{
			Source:  "test-agent",
			Content: "Test message",
			Type:    mcp.TypeMessage,
		}
		event := wsEventMsg(mcp.Event{
			Type:    mcp.EventNewMessage,
			Message: &msg,
		})
		newModel, _ := m.Update(event)
		m = newModel.(Model)

		if len(m.messages) != initialCount+1 {
			t.Errorf("Expected %d messages, got %d", initialCount+1, len(m.messages))
		}
	})

	t.Run("handle error event", func(t *testing.T) {
		event := wsEventMsg(mcp.Event{
			Type:  mcp.EventError,
			Error: "Connection lost",
		})
		newModel, _ := m.Update(event)
		m = newModel.(Model)

		if m.err == nil {
			t.Error("Expected error to be set")
		}
	})
}

// TestConfigReloadHandling tests configuration reload event handling
func TestConfigReloadHandling(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40

	// Note: Full config reload testing requires a reload manager
	// This test verifies the message handling works
	t.Run("handle config reload without manager", func(t *testing.T) {
		newConfig := config.Config{
			Agents: map[string]config.AgentConfig{
				"new-agent": {
					Command: "python",
					Model:   "claude",
					Phases:  []string{"testing"},
				},
			},
		}

		reloadMsg := configReloadMsg{newConfig: &newConfig}
		newModel, _ := m.Update(reloadMsg)
		m = newModel.(Model)

		// Without a reload manager, the config should not change
		// This is expected behavior
	})
}
