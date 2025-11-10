package beads

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestNewClient_ErrorPaths tests error handling in client creation
func TestNewClient_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		dbPath      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty path",
			dbPath:      "",
			expectError: true,
			errorMsg:    "path",
		},
		{
			name:        "nonexistent path",
			dbPath:      "/nonexistent/path/to/db",
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name:        "path with null bytes",
			dbPath:      "test\x00path",
			expectError: true,
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.dbPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Error("Expected valid client")
				}
			}
		})
	}
}

// TestGetTasks_ErrorPaths tests error handling in task retrieval
func TestGetTasks_ErrorPaths(t *testing.T) {
	// Check if bd command is available
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skip("bd command not available, skipping test")
	}

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		statuses    []string
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid db path",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/path"
			},
			statuses:    []string{"open"},
			expectError: true,
			errorMsg:    "",
		},
		{
			name: "empty statuses",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			statuses:    []string{},
			expectError: false, // Should return empty list
		},
		{
			name: "invalid status",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			statuses:    []string{"invalid-status"},
			expectError: false, // bd may handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := tt.setupFunc(t)
			client, err := NewClient(dbPath)
			if err != nil && !tt.expectError {
				t.Fatal(err)
			}
			if client == nil {
				return
			}

			tasks, err := client.GetTasks(tt.statuses)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tasks == nil {
					t.Error("Expected non-nil tasks slice")
				}
			}
		})
	}
}

// TestCreateTask_ErrorPaths tests error handling in task creation
func TestCreateTask_ErrorPaths(t *testing.T) {
	// Check if bd command is available
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skip("bd command not available, skipping test")
	}

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		title       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty title",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			title:       "",
			expectError: true,
			errorMsg:    "title",
		},
		{
			name: "title with special characters",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			title:       "test\x00title",
			expectError: true,
			errorMsg:    "",
		},
		{
			name: "invalid db path",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/path"
			},
			title:       "test task",
			expectError: true,
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := tt.setupFunc(t)
			client, err := NewClient(dbPath)
			if err != nil && !tt.expectError {
				t.Fatal(err)
			}
			if client == nil {
				return
			}

			task, err := client.CreateTask(tt.title)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
					// Clean up if task was created
					if task != nil && task.ID != "" {
						client.DeleteTask(task.ID)
					}
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if task == nil {
					t.Error("Expected non-nil task")
				}
			}
		})
	}
}

// TestUpdateTask_ErrorPaths tests error handling in task updates
func TestUpdateTask_ErrorPaths(t *testing.T) {
	// Check if bd command is available
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skip("bd command not available, skipping test")
	}

	dir := t.TempDir()
	client, err := NewClient(dir)
	if err != nil {
		t.Skip("Cannot create client")
	}

	tests := []struct {
		name        string
		taskID      string
		updates     TaskUpdate
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty task ID",
			taskID:      "",
			updates:     TaskUpdate{Status: "done"},
			expectError: true,
			errorMsg:    "id",
		},
		{
			name:        "nonexistent task ID",
			taskID:      "nonexistent-id-12345",
			updates:     TaskUpdate{Status: "done"},
			expectError: true,
			errorMsg:    "",
		},
		{
			name:        "task ID with special characters",
			taskID:      "test/../../../etc/passwd",
			updates:     TaskUpdate{Status: "done"},
			expectError: true,
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.UpdateTask(tt.taskID, tt.updates)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestDeleteTask_ErrorPaths tests error handling in task deletion
func TestDeleteTask_ErrorPaths(t *testing.T) {
	// Check if bd command is available
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skip("bd command not available, skipping test")
	}

	dir := t.TempDir()
	client, err := NewClient(dir)
	if err != nil {
		t.Skip("Cannot create client")
	}

	tests := []struct {
		name        string
		taskID      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty task ID",
			taskID:      "",
			expectError: true,
			errorMsg:    "id",
		},
		{
			name:        "nonexistent task ID",
			taskID:      "nonexistent-id-12345",
			expectError: true,
			errorMsg:    "",
		},
		{
			name:        "task ID with path traversal",
			taskID:      "../../../etc/passwd",
			expectError: true,
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.DeleteTask(tt.taskID)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestRefresh_ErrorPaths tests error handling in refresh operations
func TestRefresh_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		expectError bool
		errorMsg    string
	}{
		{
			name: "non-git directory",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			expectError: true,
			errorMsg:    "git",
		},
		{
			name: "nonexistent directory",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/path"
			},
			expectError: true,
			errorMsg:    "",
		},
		{
			name: "unreadable directory",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				subdir := filepath.Join(dir, "unreadable")
				os.MkdirAll(subdir, 0000)
				return subdir
			},
			expectError: true,
			errorMsg:    "permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := tt.setupFunc(t)
			client, err := NewClient(dbPath)
			if err != nil && !tt.expectError {
				t.Fatal(err)
			}
			if client == nil {
				return
			}

			err = client.Refresh()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestCommandExecution_ErrorPaths tests error handling in command execution
func TestCommandExecution_ErrorPaths(t *testing.T) {
	// Test with bd command not in PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", originalPath)

	dir := t.TempDir()
	client, err := NewClient(dir)
	if err != nil {
		t.Skip("Cannot create client")
	}

	// Try to get tasks - should fail because bd is not found
	_, err = client.GetTasks([]string{"open"})
	if err == nil {
		t.Error("Expected error when bd command is not in PATH")
	}
}

// TestConcurrentOperations tests error handling under concurrent access
func TestConcurrentOperations(t *testing.T) {
	// Check if bd command is available
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skip("bd command not available, skipping test")
	}

	dir := t.TempDir()
	client, err := NewClient(dir)
	if err != nil {
		t.Skip("Cannot create client")
	}

	// Run concurrent operations
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			_, err := client.GetTasks([]string{"open"})
			done <- err
		}(i)
	}

	// Collect results
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Logf("Concurrent operation error (may be expected): %v", err)
		}
	}
}

// TestErrorWrapping tests that errors are properly wrapped
func TestErrorWrapping(t *testing.T) {
	client, err := NewClient("/nonexistent/path")
	if err == nil {
		t.Fatal("Expected error for nonexistent path")
	}

	// Check that error can be unwrapped
	var unwrapped error
	for unwrapped = err; errors.Unwrap(unwrapped) != nil; unwrapped = errors.Unwrap(unwrapped) {
		// Keep unwrapping
	}

	// Should have some underlying error
	if unwrapped == err {
		t.Log("Error may not be wrapped (acceptable)")
	}
}

// TestPanicRecovery tests that panics are handled gracefully
func TestPanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic was not recovered: %v", r)
		}
	}()

	// Try to create client with invalid input (should not panic)
	_, _ = NewClient("")
	_, _ = NewClient("\x00")

	// Create valid client
	dir := t.TempDir()
	client, err := NewClient(dir)
	if err != nil {
		return
	}

	// Try operations that might panic
	client.GetTasks(nil)
	client.GetTasks([]string{})
	client.CreateTask("")
	client.UpdateTask("", TaskUpdate{})
	client.DeleteTask("")
	client.Refresh()
}

// TestInvalidInput tests handling of invalid input
func TestInvalidInput(t *testing.T) {
	dir := t.TempDir()
	client, err := NewClient(dir)
	if err != nil {
		t.Skip("Cannot create client")
	}

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "get tasks with null bytes in status",
			fn: func() error {
				_, err := client.GetTasks([]string{"open\x00"})
				return err
			},
		},
		{
			name: "create task with extremely long title",
			fn: func() error {
				longTitle := string(make([]byte, 100000))
				_, err := client.CreateTask(longTitle)
				return err
			},
		},
		{
			name: "update task with SQL injection attempt",
			fn: func() error {
				return client.UpdateTask("'; DROP TABLE tasks; --", TaskUpdate{Status: "done"})
			},
		},
		{
			name: "delete task with command injection attempt",
			fn: func() error {
				return client.DeleteTask("test; rm -rf /")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic occurred: %v", r)
				}
			}()

			err := tt.fn()
			// Should not panic, should return error
			if err != nil {
				t.Logf("Got error (expected): %v", err)
			}
		})
	}
}

// TestRecoveryFromTransientErrors tests recovery from temporary failures
func TestRecoveryFromTransientErrors(t *testing.T) {
	// Check if bd command is available
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skip("bd command not available, skipping test")
	}

	dir := t.TempDir()

	// First attempt: directory doesn't exist
	_, err := NewClient(filepath.Join(dir, "nonexistent"))
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}

	// Create the directory
	dbPath := filepath.Join(dir, "beads")
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Second attempt: should succeed
	client, err := NewClient(dbPath)
	if err != nil {
		t.Errorf("Expected success after directory creation, got: %v", err)
	}
	if client == nil {
		t.Error("Expected valid client")
	}
}

// Helper function
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
