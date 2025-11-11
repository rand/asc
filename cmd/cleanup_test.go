package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCleanupCommand tests the cleanup command workflow
func TestCleanupCommand(t *testing.T) {
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create old log file
	oldLogPath := filepath.Join(env.LogDir, "old.log")
	if err := os.WriteFile(oldLogPath, []byte("old log content"), 0644); err != nil {
		t.Fatalf("Failed to create old log file: %v", err)
	}
	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	if err := os.Chtimes(oldLogPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old log file time: %v", err)
	}

	// Create recent log file
	recentLogPath := filepath.Join(env.LogDir, "recent.log")
	if err := os.WriteFile(recentLogPath, []byte("recent log content"), 0644); err != nil {
		t.Fatalf("Failed to create recent log file: %v", err)
	}

	capture := NewCaptureOutput()
	capture.Start()

	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	if exitCalled && exitCode != 0 {
		t.Errorf("Expected successful completion, got exit code %d", exitCode)
	}

	if _, err := os.Stat(oldLogPath); !os.IsNotExist(err) {
		t.Error("Expected old log file to be removed")
	}

	if _, err := os.Stat(recentLogPath); err != nil {
		t.Error("Expected recent log file to be kept")
	}

	stdout := capture.GetStdout()
	if !strings.Contains(stdout, "Cleaning up logs") {
		t.Error("Output should mention cleaning up logs")
	}
	if !strings.Contains(stdout, "completed") {
		t.Error("Output should mention completion")
	}
}

// TestCleanupCommand_DryRun tests the cleanup command with --dry-run flag
func TestCleanupCommand_DryRun(t *testing.T) {
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	oldLogPath := filepath.Join(env.LogDir, "old.log")
	if err := os.WriteFile(oldLogPath, []byte("old log content"), 0644); err != nil {
		t.Fatalf("Failed to create old log file: %v", err)
	}
	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	if err := os.Chtimes(oldLogPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old log file time: %v", err)
	}

	cleanupDryRun = true
	defer func() { cleanupDryRun = false }()

	capture := NewCaptureOutput()
	capture.Start()

	cleanupCmd.Run(cleanupCmd, []string{})

	capture.Stop()

	if _, err := os.Stat(oldLogPath); err != nil {
		t.Error("Expected old log file to still exist in dry-run mode")
	}

	stdout := capture.GetStdout()
	if !strings.Contains(stdout, "Dry run") || !strings.Contains(stdout, "would remove") {
		t.Error("Output should mention dry-run mode")
	}
}

// TestCleanupCommand_MissingLogDir tests cleanup when log directory doesn't exist
func TestCleanupCommand_MissingLogDir(t *testing.T) {
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	tempHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	capture := NewCaptureOutput()
	capture.Start()

	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	if !exitCalled {
		t.Error("Expected os.Exit to be called for missing directory")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for missing log directory")
	}

	stderr := capture.GetStderr()
	if !strings.Contains(stderr, "Failed to cleanup logs") {
		t.Error("Error output should mention cleanup failure")
	}
}

// TestCleanupCommand_MultipleOldLogs tests cleanup with multiple old log files
func TestCleanupCommand_MultipleOldLogs(t *testing.T) {
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	for i := 1; i <= 5; i++ {
		logPath := filepath.Join(env.LogDir, "old"+string(rune('0'+i))+".log")
		if err := os.WriteFile(logPath, []byte("old log"), 0644); err != nil {
			t.Fatalf("Failed to create old log file: %v", err)
		}
		if err := os.Chtimes(logPath, oldTime, oldTime); err != nil {
			t.Fatalf("Failed to set log file time: %v", err)
		}
	}

	for i := 1; i <= 3; i++ {
		logPath := filepath.Join(env.LogDir, "recent"+string(rune('0'+i))+".log")
		if err := os.WriteFile(logPath, []byte("recent log"), 0644); err != nil {
			t.Fatalf("Failed to create recent log file: %v", err)
		}
	}

	capture := NewCaptureOutput()
	capture.Start()

	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	if exitCalled && exitCode != 0 {
		t.Errorf("Expected successful completion, got exit code %d", exitCode)
	}

	for i := 1; i <= 5; i++ {
		logPath := filepath.Join(env.LogDir, "old"+string(rune('0'+i))+".log")
		if _, err := os.Stat(logPath); !os.IsNotExist(err) {
			t.Errorf("Expected old log file %s to be removed", logPath)
		}
	}

	for i := 1; i <= 3; i++ {
		logPath := filepath.Join(env.LogDir, "recent"+string(rune('0'+i))+".log")
		if _, err := os.Stat(logPath); err != nil {
			t.Errorf("Expected recent log file %s to be kept", logPath)
		}
	}
}

// TestCleanupCommand_NonLogFiles tests that non-.log files are not removed
func TestCleanupCommand_NonLogFiles(t *testing.T) {
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	
	txtPath := filepath.Join(env.LogDir, "old.txt")
	if err := os.WriteFile(txtPath, []byte("text file"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}
	if err := os.Chtimes(txtPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set txt file time: %v", err)
	}

	logPath := filepath.Join(env.LogDir, "old.log")
	if err := os.WriteFile(logPath, []byte("log file"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	if err := os.Chtimes(logPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set log file time: %v", err)
	}

	capture := NewCaptureOutput()
	capture.Start()

	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	if exitCalled && exitCode != 0 {
		t.Errorf("Expected successful completion, got exit code %d", exitCode)
	}

	if _, err := os.Stat(txtPath); err != nil {
		t.Error("Expected .txt file to be kept")
	}

	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("Expected .log file to be removed")
	}
}
