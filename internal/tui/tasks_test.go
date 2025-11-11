package tui

import (
	"strings"
	"testing"

	"github.com/yourusername/asc/internal/beads"
)

// TestRenderTaskPane tests the renderTaskPane method
func TestRenderTaskPane(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add test task
	tf.AddTask(beads.Task{
		ID:     "task-1",
		Title:  "Test Task",
		Status: "open",
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderTaskPane(80, 20)

	if pane == "" {
		t.Error("Task pane should not be empty")
	}

	// Should contain title
	if !strings.Contains(pane, "Task Stream") {
		t.Error("Task pane should contain title")
	}

	// Should contain task
	if !strings.Contains(pane, "task-1") {
		t.Error("Task pane should contain task ID")
	}
	if !strings.Contains(pane, "Test Task") {
		t.Error("Task pane should contain task title")
	}
}

// TestRenderTaskPane_MultipleTasks tests rendering with multiple tasks
func TestRenderTaskPane_MultipleTasks(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add multiple tasks
	tf.AddTask(beads.Task{ID: "task-1", Title: "Task 1", Status: "open"})
	tf.AddTask(beads.Task{ID: "task-2", Title: "Task 2", Status: "in_progress"})
	tf.AddTask(beads.Task{ID: "task-3", Title: "Task 3", Status: "open"})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderTaskPane(80, 20)

	// Should contain all tasks
	if !strings.Contains(pane, "task-1") {
		t.Error("Task pane should contain first task")
	}
	if !strings.Contains(pane, "task-2") {
		t.Error("Task pane should contain second task")
	}
	if !strings.Contains(pane, "task-3") {
		t.Error("Task pane should contain third task")
	}
}

// TestRenderTaskPane_DifferentStatuses tests rendering tasks with different statuses
func TestRenderTaskPane_DifferentStatuses(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add tasks with different statuses
	tf.AddTask(beads.Task{ID: "task-1", Title: "Open Task", Status: "open"})
	tf.AddTask(beads.Task{ID: "task-2", Title: "In Progress Task", Status: "in_progress"})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderTaskPane(80, 20)

	// Should contain both tasks
	if !strings.Contains(pane, "task-1") {
		t.Error("Task pane should contain open task")
	}
	if !strings.Contains(pane, "task-2") {
		t.Error("Task pane should contain in-progress task")
	}
}

// TestRenderTaskPane_FilteredStatuses tests that only open/in_progress tasks are shown
func TestRenderTaskPane_FilteredStatuses(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add tasks with various statuses
	tf.AddTask(beads.Task{ID: "task-1", Title: "Open Task", Status: "open"})
	tf.AddTask(beads.Task{ID: "task-2", Title: "In Progress Task", Status: "in_progress"})
	tf.AddTask(beads.Task{ID: "task-3", Title: "Closed Task", Status: "closed"})
	tf.AddTask(beads.Task{ID: "task-4", Title: "Done Task", Status: "done"})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderTaskPane(80, 20)

	// Should contain open and in_progress tasks
	if !strings.Contains(pane, "task-1") {
		t.Error("Task pane should contain open task")
	}
	if !strings.Contains(pane, "task-2") {
		t.Error("Task pane should contain in-progress task")
	}

	// Should NOT contain closed or done tasks
	if strings.Contains(pane, "task-3") {
		t.Error("Task pane should not contain closed task")
	}
	if strings.Contains(pane, "task-4") {
		t.Error("Task pane should not contain done task")
	}
}

// TestRenderTaskPane_NoTasks tests rendering with no tasks
func TestRenderTaskPane_NoTasks(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	pane := model.renderTaskPane(80, 20)

	if !strings.Contains(pane, "No open or in-progress tasks") {
		t.Error("Task pane should show 'No open or in-progress tasks' message")
	}
}

// TestRenderTaskPane_WithSelection tests rendering with task selection
func TestRenderTaskPane_WithSelection(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	tf.AddTask(beads.Task{ID: "task-1", Title: "Task 1", Status: "open"})
	tf.AddTask(beads.Task{ID: "task-2", Title: "Task 2", Status: "open"})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	// Select first task
	model.selectedTaskIndex = 0

	pane := model.renderTaskPane(80, 20)

	// Should contain selection indicator
	if !strings.Contains(pane, "▶") {
		t.Error("Task pane should contain selection indicator")
	}
}

// TestFormatTaskLine tests the formatTaskLine method
func TestFormatTaskLine(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	task := beads.Task{
		ID:     "task-123",
		Title:  "Test Task",
		Status: "open",
	}

	line := model.formatTaskLine(task, 80, false)

	if line == "" {
		t.Error("Formatted task line should not be empty")
	}

	// Should contain task ID and title
	if !strings.Contains(line, "task-123") {
		t.Error("Formatted line should contain task ID")
	}
	if !strings.Contains(line, "Test Task") {
		t.Error("Formatted line should contain task title")
	}
}

// TestFormatTaskLine_InProgress tests formatting in-progress task
func TestFormatTaskLine_InProgress(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	task := beads.Task{
		ID:     "task-123",
		Title:  "In Progress Task",
		Status: "in_progress",
	}

	line := model.formatTaskLine(task, 80, false)

	// Should contain in-progress icon
	if !strings.Contains(line, iconInProgress) {
		t.Error("In-progress task should contain in-progress icon")
	}
}

// TestFormatTaskLine_Selected tests formatting selected task
func TestFormatTaskLine_Selected(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	task := beads.Task{
		ID:     "task-123",
		Title:  "Test Task",
		Status: "open",
	}

	line := model.formatTaskLine(task, 80, true)

	// Should contain selection indicator
	if !strings.Contains(line, "▶") {
		t.Error("Selected task should contain selection indicator")
	}
}

// TestFormatTaskLine_Truncation tests line truncation
func TestFormatTaskLine_Truncation(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	task := beads.Task{
		ID:     "task-123",
		Title:  "Very long task title that should be truncated to fit the available width",
		Status: "open",
	}

	// Use small width to force truncation
	line := model.formatTaskLine(task, 30, false)

	// Line should be truncated (accounting for ANSI codes)
	if len(line) > 150 {
		t.Error("Line should be truncated to fit width")
	}
}

// TestGetTaskIconAndStyle tests icon and style selection
func TestGetTaskIconAndStyle(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	tests := []struct {
		status       string
		expectedIcon string
	}{
		{"open", iconOpen},
		{"in_progress", iconInProgress},
		{"unknown", iconOpen}, // Default to open icon
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			icon, style := model.getTaskIconAndStyle(tt.status)

			if icon != tt.expectedIcon {
				t.Errorf("Expected icon %q, got %q", tt.expectedIcon, icon)
			}

			// Style should be valid (we can't easily check color value)
			_ = style
		})
	}
}

// TestFilterTasksByStatus tests task filtering
func TestFilterTasksByStatus(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add tasks with various statuses
	model.tasks = []beads.Task{
		{ID: "task-1", Status: "open"},
		{ID: "task-2", Status: "in_progress"},
		{ID: "task-3", Status: "closed"},
		{ID: "task-4", Status: "done"},
	}

	// Filter for open and in_progress
	filtered := model.filterTasksByStatus([]string{"open", "in_progress"})

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered tasks, got %d", len(filtered))
	}

	// Verify correct tasks are included
	foundOpen := false
	foundInProgress := false
	for _, task := range filtered {
		if task.ID == "task-1" {
			foundOpen = true
		}
		if task.ID == "task-2" {
			foundInProgress = true
		}
	}

	if !foundOpen {
		t.Error("Filtered tasks should include open task")
	}
	if !foundInProgress {
		t.Error("Filtered tasks should include in-progress task")
	}
}

// TestRenderTaskPane_DifferentSizes tests rendering with different pane sizes
func TestRenderTaskPane_DifferentSizes(t *testing.T) {
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"Small", 60, 10},
		{"Medium", 80, 20},
		{"Large", 120, 40},
	}

	for _, tt := range sizes {
		t.Run(tt.name, func(t *testing.T) {
			tf := NewTestFramework()
			model := tf.GetModel()

			tf.AddTask(beads.Task{
				ID:     "task-1",
				Title:  "Test Task",
				Status: "open",
			})

			err := model.refreshData()
			if err != nil {
				t.Fatalf("refreshData failed: %v", err)
			}

			pane := model.renderTaskPane(tt.width, tt.height)

			if pane == "" {
				t.Errorf("Task pane should not be empty for size %dx%d", tt.width, tt.height)
			}
		})
	}
}

// TestRenderTaskPane_WithKeybindingHints tests that keybinding hints are shown
func TestRenderTaskPane_WithKeybindingHints(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	pane := model.renderTaskPane(80, 20)

	// Should contain keybinding hints
	if !strings.Contains(pane, "↑↓:select") {
		t.Error("Task pane should contain navigation keybinding hint")
	}
	if !strings.Contains(pane, "c:claim") {
		t.Error("Task pane should contain claim keybinding hint")
	}
	if !strings.Contains(pane, "v:view") {
		t.Error("Task pane should contain view keybinding hint")
	}
	if !strings.Contains(pane, "n:new") {
		t.Error("Task pane should contain new keybinding hint")
	}
}

// TestRenderTaskPane_ManyTasks tests rendering with many tasks
func TestRenderTaskPane_ManyTasks(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	// Add many tasks
	for i := 1; i <= 50; i++ {
		tf.AddTask(beads.Task{
			ID:     string(rune(i)),
			Title:  "Task",
			Status: "open",
		})
	}

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderTaskPane(80, 20)

	// Should render without panic
	if pane == "" {
		t.Error("Task pane should not be empty with many tasks")
	}
}

// TestRenderTaskPane_LongTaskTitles tests rendering with long task titles
func TestRenderTaskPane_LongTaskTitles(t *testing.T) {
	tf := NewTestFramework()
	model := tf.GetModel()

	tf.AddTask(beads.Task{
		ID:     "task-1",
		Title:  "This is a very long task title that should be truncated to fit within the available pane width without breaking the layout",
		Status: "open",
	})

	err := model.refreshData()
	if err != nil {
		t.Fatalf("refreshData failed: %v", err)
	}

	pane := model.renderTaskPane(80, 20)

	// Should render without panic
	if pane == "" {
		t.Error("Task pane should not be empty with long task titles")
	}
}
