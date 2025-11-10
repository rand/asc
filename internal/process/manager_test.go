package process

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	// Verify directories were created
	if _, err := os.Stat(pidDir); os.IsNotExist(err) {
		t.Errorf("PID directory was not created")
	}
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Errorf("Log directory was not created")
	}
}

func TestStartAndStopProcess(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Start a simple process (sleep)
	pid, err := manager.Start("test-process", "sleep", []string{"10"}, []string{"TEST_ENV=value"})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if pid <= 0 {
		t.Errorf("Expected positive PID, got %d", pid)
	}

	// Verify process is running
	if !manager.IsRunning(pid) {
		t.Errorf("Process should be running")
	}

	// Verify status
	status := manager.GetStatus(pid)
	if status != StatusRunning {
		t.Errorf("Status = %v, want %v", status, StatusRunning)
	}

	// Stop the process
	err = manager.Stop(pid)
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)

	// Verify process is stopped
	if manager.IsRunning(pid) {
		t.Errorf("Process should be stopped")
	}
}

func TestProcessInfo(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Start a process
	pid, err := manager.Start("info-test", "sleep", []string{"5"}, []string{"KEY=value"})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop(pid)

	// Get process info
	info, err := manager.GetProcessInfo("info-test")
	if err != nil {
		t.Fatalf("GetProcessInfo failed: %v", err)
	}

	if info.Name != "info-test" {
		t.Errorf("Name = %v, want info-test", info.Name)
	}
	if info.PID != pid {
		t.Errorf("PID = %v, want %v", info.PID, pid)
	}
	if info.Command != "sleep" {
		t.Errorf("Command = %v, want sleep", info.Command)
	}
	if len(info.Args) != 1 || info.Args[0] != "5" {
		t.Errorf("Args = %v, want [5]", info.Args)
	}
	if info.Env["KEY"] != "value" {
		t.Errorf("Env[KEY] = %v, want value", info.Env["KEY"])
	}
}

func TestListProcesses(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Start multiple processes
	pid1, err := manager.Start("proc1", "sleep", []string{"10"}, nil)
	if err != nil {
		t.Fatalf("Start proc1 failed: %v", err)
	}
	defer manager.Stop(pid1)

	pid2, err := manager.Start("proc2", "sleep", []string{"10"}, nil)
	if err != nil {
		t.Fatalf("Start proc2 failed: %v", err)
	}
	defer manager.Stop(pid2)

	// List processes
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("ListProcesses failed: %v", err)
	}

	if len(processes) != 2 {
		t.Errorf("Expected 2 processes, got %d", len(processes))
	}

	// Verify process names
	names := make(map[string]bool)
	for _, proc := range processes {
		names[proc.Name] = true
	}
	if !names["proc1"] || !names["proc2"] {
		t.Errorf("Expected proc1 and proc2 in list, got %v", names)
	}
}

func TestStopAll(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Start multiple processes
	pid1, _ := manager.Start("stop-all-1", "sleep", []string{"30"}, nil)
	pid2, _ := manager.Start("stop-all-2", "sleep", []string{"30"}, nil)

	// Stop all
	err = manager.StopAll()
	if err != nil {
		t.Errorf("StopAll failed: %v", err)
	}

	// Give processes time to stop
	time.Sleep(200 * time.Millisecond)

	// Verify all stopped
	if manager.IsRunning(pid1) {
		t.Errorf("Process 1 should be stopped")
	}
	if manager.IsRunning(pid2) {
		t.Errorf("Process 2 should be stopped")
	}

	// Verify PID files cleaned up
	processes, _ := manager.ListProcesses()
	if len(processes) != 0 {
		t.Errorf("Expected 0 processes after StopAll, got %d", len(processes))
	}
}

func TestLogFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Start a process that outputs something
	pid, err := manager.Start("log-test", "echo", []string{"test output"}, nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for process to complete
	time.Sleep(200 * time.Millisecond)

	// Verify log file exists
	logPath := filepath.Join(logDir, "log-test.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created")
	}

	// Read log content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Verify output was captured
	if len(content) == 0 {
		t.Errorf("Log file is empty")
	}

	manager.Stop(pid)
}

func TestIsRunningNonExistentProcess(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Check a PID that doesn't exist
	if manager.IsRunning(999999) {
		t.Errorf("Non-existent process should not be running")
	}
}

func TestGetProcessInfoNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	pidDir := filepath.Join(tmpDir, "pids")
	logDir := filepath.Join(tmpDir, "logs")

	manager, err := NewManager(pidDir, logDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Try to get info for non-existent process
	_, err = manager.GetProcessInfo("non-existent")
	if err == nil {
		t.Errorf("Expected error for non-existent process")
	}
}

func TestProcessStatus(t *testing.T) {
	tests := []struct {
		name   string
		status ProcessStatus
		want   string
	}{
		{"running", StatusRunning, "running"},
		{"stopped", StatusStopped, "stopped"},
		{"error", StatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("Status = %v, want %v", tt.status, tt.want)
			}
		})
	}
}
