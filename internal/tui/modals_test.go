package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yourusername/asc/internal/beads"
)

// TestRenderTaskDetailModal tests task detail modal rendering
func TestRenderTaskDetailModal(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40
	m.tasks = []beads.Task{
		{
			ID:       "task-123",
			Title:    "Implement feature X",
			Status:   "in_progress",
			Phase:    "implementation",
			Assignee: "developer-1",
		},
	}
	m.selectedTaskIndex = 0
	m.showTaskModal = true

	// Render the modal
	output := m.renderTaskDetailModal()

	// Verify modal contains expected content
	if !strings.Contains(output, "task-123") {
		t.Error("Expected modal to contain task ID")
	}
	if !strings.Contains(output, "Implement feature X") {
		t.Error("Expected modal to contain task title")
	}
	if !strings.Contains(output, "in_progress") {
		t.Error("Expected modal to contain task status")
	}
	if !strings.Contains(output, "implementation") {
		t.Error("Expected modal to contain task phase")
	}
	if !strings.Contains(output, "developer-1") {
		t.Error("Expected modal to contain assignee")
	}
}

// TestRenderTaskDetailModalNoSelection tests modal with no task selected
func TestRenderTaskDetailModalNoSelection(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40
	m.selectedTaskIndex = -1
	m.showTaskModal = true

	// Render the modal
	output := m.renderTaskDetailModal()

	// Should return empty string when no task is selected
	if output != "" {
		t.Error("Expected empty output when no task is selected")
	}
}

// TestRenderTaskDetailModalOutOfBounds tests modal with out of bounds index
func TestRenderTaskDetailModalOutOfBounds(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40
	m.tasks = []beads.Task{
		{ID: "1", Title: "Task 1", Status: "open"},
	}
	m.selectedTaskIndex = 10 // Out of bounds
	m.showTaskModal = true

	// Render the modal
	output := m.renderTaskDetailModal()

	// Should return empty string when index is out of bounds
	if output != "" {
		t.Error("Expected empty output when index is out of bounds")
	}
}

// TestRenderCreateTaskModal tests create task modal rendering
func TestRenderCreateTaskModal(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40
	m.showCreateModal = true
	m.createTaskInput = "New task title"

	// Render the modal
	output := m.renderCreateTaskModal()

	// Verify modal contains expected content
	if !strings.Contains(output, "Create New Task") {
		t.Error("Expected modal to contain title")
	}
	if !strings.Contains(output, "New task title") {
		t.Error("Expected modal to contain input text")
	}
	if !strings.Contains(output, "enter") {
		t.Error("Expected modal to contain instructions")
	}
}

// TestRenderCreateTaskModalEmpty tests create modal with empty input
func TestRenderCreateTaskModalEmpty(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40
	m.showCreateModal = true
	m.createTaskInput = ""

	// Render the modal
	output := m.renderCreateTaskModal()

	// Should still render with cursor
	if !strings.Contains(output, "█") {
		t.Error("Expected modal to contain cursor")
	}
}

// TestRenderConfirmModal tests confirmation modal rendering
func TestRenderConfirmModal(t *testing.T) {
	tests := []struct {
		name          string
		confirmAction string
		expectedText  string
	}{
		{
			name:          "kill action",
			confirmAction: "kill",
			expectedText:  "Kill agent",
		},
		{
			name:          "restart action",
			confirmAction: "restart",
			expectedText:  "Restart agent",
		},
		{
			name:          "unknown action",
			confirmAction: "unknown",
			expectedText:  "Confirm this action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.width = 100
			m.height = 40
			m.showConfirmModal = true
			m.confirmAction = tt.confirmAction
			m.selectedAgentIndex = 0

			// Render the modal
			output := m.renderConfirmModal()

			// Verify modal contains expected content
			if !strings.Contains(output, tt.expectedText) {
				t.Errorf("Expected modal to contain '%s'", tt.expectedText)
			}
			if !strings.Contains(output, "Confirm Action") {
				t.Error("Expected modal to contain title")
			}
		})
	}
}

// TestRenderSearchInput tests search input bar rendering
func TestRenderSearchInput(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40
	m.searchMode = true
	m.searchInput = "test query"

	// Render the search input
	output := m.renderSearchInput()

	// Verify search bar contains expected content
	if !strings.Contains(output, "Search:") {
		t.Error("Expected search bar to contain label")
	}
	if !strings.Contains(output, "test query") {
		t.Error("Expected search bar to contain input text")
	}
	if !strings.Contains(output, "█") {
		t.Error("Expected search bar to contain cursor")
	}
}

// TestCenterModal tests modal centering logic
func TestCenterModal(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 40

	// Create a simple modal content
	modalContent := "Test Modal Content"

	// Center the modal
	centered := m.centerModal(modalContent)

	// Verify output is not empty
	if centered == "" {
		t.Error("Expected centered modal to have content")
	}

	// Verify content is present
	if !strings.Contains(centered, modalContent) {
		t.Error("Expected centered modal to contain original content")
	}
}

// TestCenterModalSmallTerminal tests modal centering with small terminal
func TestCenterModalSmallTerminal(t *testing.T) {
	m := createTestModel()
	m.width = 20
	m.height = 10

	// Create a modal content that's larger than terminal
	modalContent := strings.Repeat("X", 50)

	// Center the modal
	centered := m.centerModal(modalContent)

	// Should still work without crashing
	if centered == "" {
		t.Error("Expected centered modal to have content even with small terminal")
	}
}

// TestModalInputHandling tests input handling in different modal states
func TestModalInputHandling(t *testing.T) {
	t.Run("create modal accepts alphanumeric", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40
		m.showCreateModal = true

		// Type various characters
		chars := []rune{'a', 'b', 'c', '1', '2', '3', ' ', '-'}
		for _, ch := range chars {
			newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
			m = newModel.(Model)
		}

		expected := "abc123 -"
		if m.createTaskInput != expected {
			t.Errorf("Expected input '%s', got '%s'", expected, m.createTaskInput)
		}
	})

	t.Run("create modal backspace handling", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40
		m.showCreateModal = true
		m.createTaskInput = "test"

		// Backspace twice
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = newModel.(Model)
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = newModel.(Model)

		if m.createTaskInput != "te" {
			t.Errorf("Expected input 'te', got '%s'", m.createTaskInput)
		}

		// Backspace on empty should not crash
		m.createTaskInput = ""
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = newModel.(Model)
		if m.createTaskInput != "" {
			t.Error("Expected input to remain empty")
		}
	})

	t.Run("search input accepts text", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40
		m.searchMode = true

		// Type search query
		for _, ch := range "error" {
			newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
			m = newModel.(Model)
		}

		if m.searchInput != "error" {
			t.Errorf("Expected search input 'error', got '%s'", m.searchInput)
		}
	})

	t.Run("confirm modal accepts y/n", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40
		m.showConfirmModal = true
		m.confirmAction = "kill"

		// Press 'y' to confirm
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
		m = newModel.(Model)

		if m.showConfirmModal {
			t.Error("Expected modal to close after confirmation")
		}

		// Open again and press 'N' (uppercase)
		m.showConfirmModal = true
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
		m = newModel.(Model)

		if m.showConfirmModal {
			t.Error("Expected modal to close after cancellation")
		}
	})
}

// TestModalPriority tests that modals take priority over normal key handling
func TestModalPriority(t *testing.T) {
	t.Run("create modal blocks normal keys", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40
		m.showCreateModal = true
		initialTaskIndex := m.selectedTaskIndex

		// Try to use arrow keys (should not navigate tasks)
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(Model)

		if m.selectedTaskIndex != initialTaskIndex {
			t.Error("Expected task navigation to be blocked by modal")
		}
	})

	t.Run("search mode blocks normal keys", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40
		m.searchMode = true
		initialAgentIndex := m.selectedAgentIndex

		// Try to select agent (should not work in search mode)
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
		m = newModel.(Model)

		// The '1' should be added to search input, not select agent
		if m.searchInput != "1" {
			t.Error("Expected '1' to be added to search input")
		}
		if m.selectedAgentIndex != initialAgentIndex {
			t.Error("Expected agent selection to be blocked by search mode")
		}
	})

	t.Run("task modal blocks task navigation", func(t *testing.T) {
		m := createTestModel()
		m.width = 100
		m.height = 40
		m.showTaskModal = true
		m.tasks = []beads.Task{
			{ID: "1", Title: "Task 1", Status: "open"},
			{ID: "2", Title: "Task 2", Status: "open"},
		}
		initialIndex := m.selectedTaskIndex

		// Try to navigate (should not work)
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(Model)

		if m.selectedTaskIndex != initialIndex {
			t.Error("Expected navigation to be blocked by task modal")
		}
	})
}
