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
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create some old log files
	oldLogPath := filepath.Join(env.LogDir, "old.log")
	if err := os.WriteFile(oldLogPath, []byte("old log content"), 0644); err != nil {
		t.Fatalf("Failed to create old log file: %v", err)
	}
	
	// Set modification time to 40 days ago
	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	if err := os.Chtimes(oldLogPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old log file time: %v", err)
	}

	// Create a recent log file
	recentLogPath := filepath.Join(env.LogDir, "recent.log")
	if err := os.WriteFile(recentLogPath, []byte("recent log content"), 0644); err != nil {
		t.Fatalf("Failed to create recent log file: %v", err)
	}

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command with default 30 days
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify command completed successfully (no exit call means success)
	if exitCalled && exitCode != 0 {
		t.Errorf("Expected successful completion, got exit code %d", exitCode)
	}

	// Verify old log was removed
	if _, err := os.Stat(oldLogPath); !os.IsNotExist(err) {
		t.Error("Expected old log file to be removed")
	}

	// Verify recent log was kept
	if _, err := os.Stat(recentLogPath); err != nil {
		t.Error("Expected recent log file to be kept")
	}

	// Verify output
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
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create an old log file
	oldLogPath := filepath.Join(env.LogDir, "old.log")
	if err := os.WriteFile(oldLogPath, []byte("old log content"), 0644); err != nil {
		t.Fatalf("Failed to create old log file: %v", err)
	}
	
	// Set modification time to 40 days ago
	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	if err := os.Chtimes(oldLogPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old log file time: %v", err)
	}

	// Set dry-run flag
	cleanupDryRun = true
	defer func() { cleanupDryRun = false }()

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code
	if exitCalled {
		t.Error("Expected os.Exit not to be called in dry-run mode")
	}
	_ = exitCode

	// Verify old log was NOT removed (dry-run)
	if _, err := os.Stat(oldLogPath); err != nil {
		t.Error("Expected old log file to still exist in dry-run mode")
	}

	// Verify output mentions dry-run
	stdout := capture.GetStdout()
	if !strings.Contains(stdout, "Dry run") || !strings.Contains(stdout, "would remove") {
		t.Error("Output should mention dry-run mode")
	}
}

// TestCleanupCommand_CustomDays tests the cleanup command with custom --days flag
func TestCleanupCommand_CustomDays(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create a log file that's 20 days old
	logPath := filepath.Join(env.LogDir, "medium.log")
	if err := os.WriteFile(logPath, []byte("medium age log"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	
	// Set modification time to 20 days ago
	oldTime := time.Now().Add(-20 * 24 * time.Hour)
	if err := os.Chtimes(logPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set log file time: %v", err)
	}

	// Set custom days to 10 (should remove 20-day-old file)
	cleanupDays = 10
	defer func() { cleanupDays = 30 }()

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify log was removed (older than 10 days)
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("Expected log file to be removed with custom days setting")
	}

	// Verify output mentions custom days
	stdout := capture.GetStdout()
	if !strings.Contains(stdout, "10 days") {
		t.Error("Output should mention custom days setting")
	}
}

// TestCleanupCommand_NoOldLogs tests cleanup when there are no old logs
func TestCleanupCommand_NoOldLogs(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create only recent log files
	recentLogPath := filepath.Join(env.LogDir, "recent.log")
	if err := os.WriteFile(recentLogPath, []byte("recent log"), 0644); err != nil {
		t.Fatalf("Failed to create recent log file: %v", err)
	}

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify recent log still exists
	if _, err := os.Stat(recentLogPath); err != nil {
		t.Error("Expected recent log file to still exist")
	}

	// Verify output
	stdout := capture.GetStdout()
	if !strings.Contains(stdout, "completed") {
		t.Error("Output should mention completion")
	}
}

// TestCleanupCommand_EmptyLogDir tests cleanup with empty log directory
func TestCleanupCommand_EmptyLogDir(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Log directory exists but is empty (created by NewTestEnvironment)

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify output
	stdout := capture.GetStdout()
	if !strings.Contains(stdout, "completed") {
		t.Error("Output should mention completion")
	}
}

// TestCleanupCommand_MissingLogDir tests cleanup when log directory doesn't exist
func TestCleanupCommand_MissingLogDir(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME to a directory without .asc/logs
	tempHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code (should fail)
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for missing log directory")
	}

	// Verify error output
	stderr := capture.GetStderr()
	if !strings.Contains(stderr, "Failed to cleanup logs") {
		t.Error("Error output should mention cleanup failure")
	}
}

// TestCleanupCommand_InvalidHomeDir tests cleanup when HOME is invalid
func TestCleanupCommand_InvalidHomeDir(t *testing.T) {
	// Save original HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Set HOME to empty (will cause UserHomeDir to fail on some systems)
	os.Unsetenv("HOME")

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit was called
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}

	// Note: On some systems, UserHomeDir may still succeed even without HOME
	// The test verifies the command completes without panicking
	_ = exitCode
}

// TestCleanupCommand_MultipleOldLogs tests cleanup with multiple old log files
func TestCleanupCommand_MultipleOldLogs(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create multiple old log files
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

	// Create some recent log files
	for i := 1; i <= 3; i++ {
		logPath := filepath.Join(env.LogDir, "recent"+string(rune('0'+i))+".log")
		if err := os.WriteFile(logPath, []byte("recent log"), 0644); err != nil {
			t.Fatalf("Failed to create recent log file: %v", err)
		}
	}

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify old logs were removed
	for i := 1; i <= 5; i++ {
		logPath := filepath.Join(env.LogDir, "old"+string(rune('0'+i))+".log")
		if _, err := os.Stat(logPath); !os.IsNotExist(err) {
			t.Errorf("Expected old log file %s to be removed", logPath)
		}
	}

	// Verify recent logs were kept
	for i := 1; i <= 3; i++ {
		logPath := filepath.Join(env.LogDir, "recent"+string(rune('0'+i))+".log")
		if _, err := os.Stat(logPath); err != nil {
			t.Errorf("Expected recent log file %s to be kept", logPath)
		}
	}
}

// TestCleanupCommand_NonLogFiles tests that non-.log files are not removed
func TestCleanupCommand_NonLogFiles(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create old non-.log files
	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	
	txtPath := filepath.Join(env.LogDir, "old.txt")
	if err := os.WriteFile(txtPath, []byte("text file"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}
	if err := os.Chtimes(txtPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set txt file time: %v", err)
	}

	jsonPath := filepath.Join(env.LogDir, "old.json")
	if err := os.WriteFile(jsonPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create json file: %v", err)
	}
	if err := os.Chtimes(jsonPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set json file time: %v", err)
	}

	// Create old .log file
	logPath := filepath.Join(env.LogDir, "old.log")
	if err := os.WriteFile(logPath, []byte("log file"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	if err := os.Chtimes(logPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set log file time: %v", err)
	}

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify non-.log files were NOT removed
	if _, err := os.Stat(txtPath); err != nil {
		t.Error("Expected .txt file to be kept")
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Error("Expected .json file to be kept")
	}

	// Verify .log file was removed
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("Expected .log file to be removed")
	}
}

// TestCleanupCommand_Subdirectories tests that subdirectories are ignored
func TestCleanupCommand_Subdirectories(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create a subdirectory with old log files
	subDir := filepath.Join(env.LogDir, "archive")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	subLogPath := filepath.Join(subDir, "old.log")
	if err := os.WriteFile(subLogPath, []byte("archived log"), 0644); err != nil {
		t.Fatalf("Failed to create log in subdirectory: %v", err)
	}
	if err := os.Chtimes(subLogPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set log file time: %v", err)
	}

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify subdirectory and its contents were NOT removed
	if _, err := os.Stat(subDir); err != nil {
		t.Error("Expected subdirectory to be kept")
	}
	if _, err := os.Stat(subLogPath); err != nil {
		t.Error("Expected log file in subdirectory to be kept")
	}
}

// TestCleanupCommand_PermissionError tests cleanup with permission errors
func TestCleanupCommand_PermissionError(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create an old log file
	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	logPath := filepath.Join(env.LogDir, "old.log")
	if err := os.WriteFile(logPath, []byte("old log"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	if err := os.Chtimes(logPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set log file time: %v", err)
	}

	// Make log directory read-only to cause permission error
	if err := os.Chmod(env.LogDir, 0555); err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}
	defer os.Chmod(env.LogDir, 0755) // Restore permissions for cleanup

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code (should fail due to permission error)
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for permission error")
	}

	// Verify error output
	stderr := capture.GetStderr()
	if !strings.Contains(stderr, "Failed to cleanup logs") {
		t.Error("Error output should mention cleanup failure")
	}
}

// TestCleanupCommand_ZeroDays tests cleanup with zero days (should remove all logs)
func TestCleanupCommand_ZeroDays(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create log files with various ages
	for i, days := range []int{1, 5, 10, 30, 60} {
		logPath := filepath.Join(env.LogDir, "log"+string(rune('0'+i))+".log")
		if err := os.WriteFile(logPath, []byte("log"), 0644); err != nil {
			t.Fatalf("Failed to create log file: %v", err)
		}
		oldTime := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
		if err := os.Chtimes(logPath, oldTime, oldTime); err != nil {
			t.Fatalf("Failed to set log file time: %v", err)
		}
	}

	// Set days to 0 (remove all logs)
	cleanupDays = 0
	defer func() { cleanupDays = 30 }()

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify all logs were removed
	entries, err := os.ReadDir(env.LogDir)
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}
	
	logCount := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".log" {
			logCount++
		}
	}
	
	if logCount > 0 {
		t.Errorf("Expected all log files to be removed, found %d", logCount)
	}
}

// TestCleanupCommand_LargeDays tests cleanup with very large days value
func TestCleanupCommand_LargeDays(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override HOME for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create old log files
	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	logPath := filepath.Join(env.LogDir, "old.log")
	if err := os.WriteFile(logPath, []byte("old log"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	if err := os.Chtimes(logPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set log file time: %v", err)
	}

	// Set days to very large value (should keep all logs)
	cleanupDays = 365
	defer func() { cleanupDays = 30 }()

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	// Run cleanup command
	exitCode, exitCalled := RunWithExitCapture(func() {
		cleanupCmd.Run(cleanupCmd, []string{})
	})

	capture.Stop()

	// Verify exit code
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify log was kept (not old enough)
	if _, err := os.Stat(logPath); err != nil {
		t.Error("Expected log file to be kept with large days value")
	}
}
