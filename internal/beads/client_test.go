package beads

import (
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
