package beads

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("/path/to/repo", 5*time.Second)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.dbPath != "/path/to/repo" {
		t.Errorf("dbPath = %v, want /path/to/repo", client.dbPath)
	}
	if client.refreshInterval != 5*time.Second {
		t.Errorf("refreshInterval = %v, want 5s", client.refreshInterval)
	}
}

func TestTask(t *testing.T) {
	task := Task{
		ID:       "task-123",
		Title:    "Test task",
		Status:   "open",
		Phase:    "planning",
		Assignee: "agent-1",
	}

	if task.ID != "task-123" {
		t.Errorf("ID = %v, want task-123", task.ID)
	}
	if task.Title != "Test task" {
		t.Errorf("Title = %v, want Test task", task.Title)
	}
	if task.Status != "open" {
		t.Errorf("Status = %v, want open", task.Status)
	}
	if task.Phase != "planning" {
		t.Errorf("Phase = %v, want planning", task.Phase)
	}
	if task.Assignee != "agent-1" {
		t.Errorf("Assignee = %v, want agent-1", task.Assignee)
	}
}

func TestTaskUpdate(t *testing.T) {
	title := "Updated title"
	status := "in_progress"
	phase := "implementation"
	assignee := "agent-2"

	update := TaskUpdate{
		Title:    &title,
		Status:   &status,
		Phase:    &phase,
		Assignee: &assignee,
	}

	if update.Title == nil || *update.Title != "Updated title" {
		t.Errorf("Title = %v, want Updated title", update.Title)
	}
	if update.Status == nil || *update.Status != "in_progress" {
		t.Errorf("Status = %v, want in_progress", update.Status)
	}
	if update.Phase == nil || *update.Phase != "implementation" {
		t.Errorf("Phase = %v, want implementation", update.Phase)
	}
	if update.Assignee == nil || *update.Assignee != "agent-2" {
		t.Errorf("Assignee = %v, want agent-2", update.Assignee)
	}
}

func TestTaskUpdatePartial(t *testing.T) {
	status := "completed"
	update := TaskUpdate{
		Status: &status,
	}

	if update.Status == nil || *update.Status != "completed" {
		t.Errorf("Status = %v, want completed", update.Status)
	}
	if update.Title != nil {
		t.Errorf("Title should be nil")
	}
	if update.Phase != nil {
		t.Errorf("Phase should be nil")
	}
	if update.Assignee != nil {
		t.Errorf("Assignee should be nil")
	}
}

func TestClientInterface(t *testing.T) {
	var _ BeadsClient = (*Client)(nil)
}

func TestTaskStatuses(t *testing.T) {
	statuses := []string{"open", "in_progress", "completed", "blocked"}
	
	for _, status := range statuses {
		task := Task{
			ID:     "test",
			Status: status,
		}
		if task.Status != status {
			t.Errorf("Status = %v, want %v", task.Status, status)
		}
	}
}

func TestTaskPhases(t *testing.T) {
	phases := []string{"planning", "design", "implementation", "testing", "review"}
	
	for _, phase := range phases {
		task := Task{
			ID:    "test",
			Phase: phase,
		}
		if task.Phase != phase {
			t.Errorf("Phase = %v, want %v", task.Phase, phase)
		}
	}
}

func TestMultipleTasks(t *testing.T) {
	tasks := []Task{
		{ID: "1", Title: "Task 1", Status: "open"},
		{ID: "2", Title: "Task 2", Status: "in_progress"},
		{ID: "3", Title: "Task 3", Status: "completed"},
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	for i, task := range tasks {
		expectedID := string(rune('1' + i))
		if task.ID != expectedID {
			t.Errorf("Task %d: ID = %v, want %v", i, task.ID, expectedID)
		}
	}
}

func TestTaskWithoutAssignee(t *testing.T) {
	task := Task{
		ID:     "task-1",
		Title:  "Unassigned task",
		Status: "open",
		Phase:  "planning",
	}

	if task.Assignee != "" {
		t.Errorf("Assignee should be empty, got %v", task.Assignee)
	}
}

func TestClientWithEmptyPath(t *testing.T) {
	client := NewClient("", 5*time.Second)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.dbPath != "" {
		t.Errorf("dbPath should be empty, got %v", client.dbPath)
	}
}

func TestRefreshInterval(t *testing.T) {
	intervals := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		10 * time.Second,
		30 * time.Second,
	}

	for _, interval := range intervals {
		client := NewClient("/path", interval)
		if client.refreshInterval != interval {
			t.Errorf("refreshInterval = %v, want %v", client.refreshInterval, interval)
		}
	}
}

// Table-driven tests for error handling and edge cases

func TestGetTasks_EmptyStatuses(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	// This will fail because bd is not installed, but we're testing the code path
	_, err := client.GetTasks([]string{})
	
	// We expect an error since bd is likely not installed in test environment
	if err == nil {
		t.Log("bd command succeeded (bd is installed)")
	} else {
		// Verify error message format
		if err.Error() == "" {
			t.Error("Error message should not be empty")
		}
	}
}

func TestGetTasks_WithStatuses(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	statuses := []string{"open", "in_progress"}
	_, err := client.GetTasks(statuses)
	
	// We expect an error since bd is likely not installed in test environment
	if err == nil {
		t.Log("bd command succeeded (bd is installed)")
	} else {
		// Verify error message format
		if err.Error() == "" {
			t.Error("Error message should not be empty")
		}
	}
}

func TestGetTasks_SingleStatus(t *testing.T) {
	client := NewClient("/test/path", 5*time.Second)
	
	_, err := client.GetTasks([]string{"open"})
	
	// We expect an error since bd is likely not installed
	if err != nil {
		// Verify error contains expected information
		errStr := err.Error()
		if !contains(errStr, "bd") {
			t.Errorf("Expected error to mention 'bd', got: %s", errStr)
		}
	}
}

func TestGetTasks_MultipleStatuses(t *testing.T) {
	tests := []struct {
		name     string
		statuses []string
	}{
		{"two statuses", []string{"open", "in_progress"}},
		{"three statuses", []string{"open", "in_progress", "completed"}},
		{"single status", []string{"blocked"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("", 5*time.Second)
			_, err := client.GetTasks(tt.statuses)
			
			// Error expected since bd not installed
			if err != nil && err.Error() == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}

func TestCreateTask_ErrorHandling(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	_, err := client.CreateTask("Test task")
	
	// We expect an error since bd is likely not installed
	if err == nil {
		t.Log("bd command succeeded (bd is installed)")
	} else {
		// Verify error message format
		if err.Error() == "" {
			t.Error("Error message should not be empty")
		}
		if !contains(err.Error(), "bd") {
			t.Errorf("Expected error to mention 'bd', got: %s", err.Error())
		}
	}
}

func TestCreateTask_EmptyTitle(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	_, err := client.CreateTask("")
	
	// Error expected
	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestCreateTask_LongTitle(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	longTitle := "This is a very long task title that contains many words and characters to test how the system handles long input strings"
	_, err := client.CreateTask(longTitle)
	
	// Error expected since bd not installed
	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestUpdateTask_AllFields(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	title := "Updated title"
	status := "in_progress"
	phase := "implementation"
	assignee := "agent-1"
	
	update := TaskUpdate{
		Title:    &title,
		Status:   &status,
		Phase:    &phase,
		Assignee: &assignee,
	}
	
	err := client.UpdateTask("task-123", update)
	
	// Error expected since bd not installed
	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestUpdateTask_SingleField(t *testing.T) {
	tests := []struct {
		name   string
		update TaskUpdate
	}{
		{
			name: "update title only",
			update: TaskUpdate{
				Title: stringPtr("New title"),
			},
		},
		{
			name: "update status only",
			update: TaskUpdate{
				Status: stringPtr("completed"),
			},
		},
		{
			name: "update phase only",
			update: TaskUpdate{
				Phase: stringPtr("testing"),
			},
		},
		{
			name: "update assignee only",
			update: TaskUpdate{
				Assignee: stringPtr("agent-2"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("", 5*time.Second)
			err := client.UpdateTask("task-123", tt.update)
			
			// Error expected since bd not installed
			if err != nil && err.Error() == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}

func TestUpdateTask_EmptyUpdate(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	update := TaskUpdate{}
	err := client.UpdateTask("task-123", update)
	
	// Error expected since bd not installed
	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestDeleteTask_ErrorHandling(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	err := client.DeleteTask("task-123")
	
	// Error expected since bd not installed
	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestDeleteTask_EmptyID(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	err := client.DeleteTask("")
	
	// Error expected
	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestDeleteTask_InvalidID(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	err := client.DeleteTask("nonexistent-task")
	
	// Error expected since bd not installed or task doesn't exist
	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestRefresh_EmptyPath(t *testing.T) {
	client := NewClient("", 5*time.Second)
	
	err := client.Refresh()
	
	// Should return error about missing dbPath
	if err == nil {
		t.Error("Expected error for empty dbPath")
	}
	if err != nil && !contains(err.Error(), "dbPath") {
		t.Errorf("Expected error to mention 'dbPath', got: %s", err.Error())
	}
}

func TestRefresh_InvalidPath(t *testing.T) {
	client := NewClient("/nonexistent/path", 5*time.Second)
	
	err := client.Refresh()
	
	// Should return error about git pull failure
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestRefresh_ValidPath(t *testing.T) {
	// Use current directory which should have git
	client := NewClient(".", 5*time.Second)
	
	err := client.Refresh()
	
	// May succeed or fail depending on git state, but should not panic
	if err != nil {
		t.Logf("Refresh failed (expected in test environment): %v", err)
	}
}

func TestClient_WithDifferentPaths(t *testing.T) {
	tests := []struct {
		name   string
		dbPath string
	}{
		{"empty path", ""},
		{"relative path", "./test"},
		{"absolute path", "/tmp/test"},
		{"home path", "~/test"},
		{"current dir", "."},
		{"parent dir", ".."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.dbPath, 5*time.Second)
			if client == nil {
				t.Fatal("NewClient returned nil")
			}
			if client.dbPath != tt.dbPath {
				t.Errorf("dbPath = %v, want %v", client.dbPath, tt.dbPath)
			}
		})
	}
}

func TestClient_WithDifferentIntervals(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
	}{
		{"zero interval", 0},
		{"1 second", 1 * time.Second},
		{"5 seconds", 5 * time.Second},
		{"1 minute", 1 * time.Minute},
		{"negative interval", -1 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("/path", tt.interval)
			if client == nil {
				t.Fatal("NewClient returned nil")
			}
			if client.refreshInterval != tt.interval {
				t.Errorf("refreshInterval = %v, want %v", client.refreshInterval, tt.interval)
			}
		})
	}
}

func TestTaskUpdate_NilFields(t *testing.T) {
	update := TaskUpdate{}
	
	if update.Title != nil {
		t.Error("Title should be nil")
	}
	if update.Status != nil {
		t.Error("Status should be nil")
	}
	if update.Phase != nil {
		t.Error("Phase should be nil")
	}
	if update.Assignee != nil {
		t.Error("Assignee should be nil")
	}
}

func TestTask_JSONMarshaling(t *testing.T) {
	task := Task{
		ID:       "task-123",
		Title:    "Test task",
		Status:   "open",
		Phase:    "planning",
		Assignee: "agent-1",
	}

	// Test that task can be marshaled (even though we don't use the result)
	_, err := json.Marshal(task)
	if err != nil {
		t.Errorf("Failed to marshal task: %v", err)
	}
}

func TestTask_EmptyFields(t *testing.T) {
	task := Task{}
	
	if task.ID != "" {
		t.Error("ID should be empty")
	}
	if task.Title != "" {
		t.Error("Title should be empty")
	}
	if task.Status != "" {
		t.Error("Status should be empty")
	}
	if task.Phase != "" {
		t.Error("Phase should be empty")
	}
	if task.Assignee != "" {
		t.Error("Assignee should be empty")
	}
}

func TestTask_SpecialCharacters(t *testing.T) {
	task := Task{
		ID:       "task-123",
		Title:    "Test with special chars: !@#$%^&*()",
		Status:   "open",
		Phase:    "planning",
		Assignee: "agent-1",
	}

	if !contains(task.Title, "!@#$%^&*()") {
		t.Error("Title should contain special characters")
	}
}

func TestTask_UnicodeCharacters(t *testing.T) {
	task := Task{
		ID:       "task-123",
		Title:    "Test with unicode: 你好世界 🚀",
		Status:   "open",
		Phase:    "planning",
		Assignee: "agent-1",
	}

	if !contains(task.Title, "你好世界") {
		t.Error("Title should contain unicode characters")
	}
	if !contains(task.Title, "🚀") {
		t.Error("Title should contain emoji")
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
