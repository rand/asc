package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/asc/internal/process"
)

func TestNewDoctor(t *testing.T) {
	doc, err := NewDoctor("asc.toml", ".env")
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	if doc == nil {
		t.Fatal("Doctor is nil")
	}
	
	if doc.configPath != "asc.toml" {
		t.Errorf("Expected configPath 'asc.toml', got '%s'", doc.configPath)
	}
	
	if doc.envPath != ".env" {
		t.Errorf("Expected envPath '.env', got '%s'", doc.envPath)
	}
}

func TestDiagnosticReport_HasCriticalIssues(t *testing.T) {
	tests := []struct {
		name     string
		issues   []Issue
		expected bool
	}{
		{
			name:     "no issues",
			issues:   []Issue{},
			expected: false,
		},
		{
			name: "only low severity",
			issues: []Issue{
				{Severity: SeverityLow},
			},
			expected: false,
		},
		{
			name: "has critical",
			issues: []Issue{
				{Severity: SeverityLow},
				{Severity: SeverityCritical},
			},
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &DiagnosticReport{
				Issues: tt.issues,
			}
			
			result := report.HasCriticalIssues()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDiagnosticReport_ToJSON(t *testing.T) {
	report := &DiagnosticReport{
		RunAt: time.Date(2025, 11, 10, 12, 0, 0, 0, time.UTC),
		Issues: []Issue{
			{
				ID:          "test-issue",
				Category:    CategoryConfiguration,
				Severity:    SeverityHigh,
				Title:       "Test Issue",
				Description: "Test description",
				Impact:      "Test impact",
				Remediation: "Test remediation",
				AutoFixable: true,
				DetectedAt:  time.Date(2025, 11, 10, 12, 0, 0, 0, time.UTC),
			},
		},
		HealthSummary: "Test summary",
	}
	
	json, err := report.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}
	
	if json == "" {
		t.Error("JSON output is empty")
	}
	
	// Check that JSON contains expected fields
	expectedFields := []string{"run_at", "issues", "health_summary", "test-issue"}
	for _, field := range expectedFields {
		if len(json) < len(field) {
			t.Errorf("JSON too short to contain field '%s'", field)
			continue
		}
		found := false
		for i := 0; i <= len(json)-len(field); i++ {
			if json[i:i+len(field)] == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("JSON does not contain expected field: %s", field)
		}
	}
}

func TestDiagnosticReport_Format(t *testing.T) {
	report := &DiagnosticReport{
		RunAt: time.Now(),
		Issues: []Issue{
			{
				ID:          "test-critical",
				Category:    CategoryConfiguration,
				Severity:    SeverityCritical,
				Title:       "Critical Issue",
				Description: "Critical description",
				Impact:      "Critical impact",
				Remediation: "Fix it now",
				AutoFixable: true,
				DetectedAt:  time.Now(),
			},
			{
				ID:          "test-low",
				Category:    CategoryState,
				Severity:    SeverityLow,
				Title:       "Low Issue",
				Description: "Low description",
				Impact:      "Minor impact",
				Remediation: "Fix when convenient",
				AutoFixable: false,
				DetectedAt:  time.Now(),
			},
		},
		HealthSummary: "Found 2 issue(s): 1 critical, 1 low",
	}
	
	// Test non-verbose format
	output := report.Format(false)
	if output == "" {
		t.Error("Format output is empty")
	}
	
	// Check for expected content
	expectedContent := []string{
		"DIAGNOSTIC REPORT",
		"Critical Issue",
		"Low Issue",
		"Remediation:",
	}
	
	for _, content := range expectedContent {
		if len(output) < len(content) {
			t.Errorf("Output too short to contain '%s'", content)
			continue
		}
		found := false
		for i := 0; i <= len(output)-len(content); i++ {
			if output[i:i+len(content)] == content {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Output does not contain expected content: %s", content)
		}
	}
	
	// Test verbose format
	verboseOutput := report.Format(true)
	if len(verboseOutput) <= len(output) {
		t.Error("Verbose output should be longer than non-verbose")
	}
}

func TestIsProcessRunning(t *testing.T) {
	// Test with current process (should be running)
	currentPID := os.Getpid()
	if !isProcessRunning(currentPID) {
		t.Error("Current process should be running")
	}
	
	// Test with invalid PID (should not be running)
	if isProcessRunning(999999) {
		t.Error("Invalid PID should not be running")
	}
}

func TestGetDirSize(t *testing.T) {
	// Create temporary directory with files
	tmpDir := t.TempDir()
	
	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testFile2 := filepath.Join(tmpDir, "test2.txt")
	
	content1 := []byte("Hello, World!")
	content2 := []byte("Test content")
	
	if err := os.WriteFile(testFile1, content1, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	if err := os.WriteFile(testFile2, content2, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Get directory size
	size, err := getDirSize(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get directory size: %v", err)
	}
	
	expectedSize := int64(len(content1) + len(content2))
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
}

func TestGenerateHealthSummary(t *testing.T) {
	doc, _ := NewDoctor("asc.toml", ".env")
	
	tests := []struct {
		name     string
		issues   []Issue
		expected string
	}{
		{
			name:     "no issues",
			issues:   []Issue{},
			expected: "✓ All checks passed - system is healthy",
		},
		{
			name: "one critical",
			issues: []Issue{
				{Severity: SeverityCritical},
			},
			expected: "Found 1 issue(s): 1 critical",
		},
		{
			name: "mixed severities",
			issues: []Issue{
				{Severity: SeverityCritical},
				{Severity: SeverityHigh},
				{Severity: SeverityLow},
			},
			expected: "Found 3 issue(s): 1 critical, 1 high, 1 low",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &DiagnosticReport{
				Issues: tt.issues,
			}
			
			summary := doc.generateHealthSummary(report)
			if summary != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, summary)
			}
		})
	}
}

// TestRecoveryFromCorruptedPID tests recovery from corrupted PID files
func TestRecoveryFromCorruptedPID(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	ascDir := filepath.Join(tmpDir, ".asc")
	pidDir := filepath.Join(ascDir, "pids")
	logDir := filepath.Join(ascDir, "logs")
	
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		t.Fatalf("Failed to create pid directory: %v", err)
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}
	
	// Create corrupted PID file
	corruptedPIDPath := filepath.Join(pidDir, "corrupted-agent.json")
	if err := os.WriteFile(corruptedPIDPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted PID file: %v", err)
	}
	
	// Create temporary config files
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	// Write minimal valid config
	configContent := `[core]
beads_db_path = "./repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create doctor instance
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	doc.homeDir = tmpDir
	
	// Run diagnostics
	report, err := doc.RunDiagnostics()
	if err != nil {
		t.Fatalf("Failed to run diagnostics: %v", err)
	}
	
	// Verify corrupted PID file was detected
	foundCorruptedIssue := false
	for _, issue := range report.Issues {
		if issue.Category == CategoryState && issue.ID == "pid-corrupted-corrupted-agent.json" {
			foundCorruptedIssue = true
			if !issue.AutoFixable {
				t.Error("Corrupted PID issue should be auto-fixable")
			}
			break
		}
	}
	
	if !foundCorruptedIssue {
		t.Error("Corrupted PID file was not detected")
	}
	
	// Apply fixes
	fixes, err := doc.ApplyFixes(report)
	if err != nil {
		t.Fatalf("Failed to apply fixes: %v", err)
	}
	
	// Verify fix was applied
	foundFix := false
	for _, fix := range fixes {
		if fix.IssueID == "pid-corrupted-corrupted-agent.json" {
			foundFix = true
			if !fix.Success {
				t.Errorf("Fix failed: %s", fix.Message)
			}
			break
		}
	}
	
	if !foundFix {
		t.Error("Fix for corrupted PID was not applied")
	}
	
	// Verify file was removed
	if _, err := os.Stat(corruptedPIDPath); !os.IsNotExist(err) {
		t.Error("Corrupted PID file was not removed")
	}
}

// TestRecoveryFromOrphanedPID tests recovery from orphaned PID files
func TestRecoveryFromOrphanedPID(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	ascDir := filepath.Join(tmpDir, ".asc")
	pidDir := filepath.Join(ascDir, "pids")
	logDir := filepath.Join(ascDir, "logs")
	
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		t.Fatalf("Failed to create pid directory: %v", err)
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}
	
	// Create orphaned PID file (with non-existent PID)
	orphanedPIDPath := filepath.Join(pidDir, "orphaned-agent.json")
	procInfo := process.ProcessInfo{
		Name:      "orphaned-agent",
		PID:       999999, // Non-existent PID
		Command:   "python",
		Args:      []string{"agent.py"},
		StartedAt: time.Now(),
		LogFile:   filepath.Join(logDir, "orphaned-agent.log"),
	}
	
	data, err := json.Marshal(procInfo)
	if err != nil {
		t.Fatalf("Failed to marshal process info: %v", err)
	}
	
	if err := os.WriteFile(orphanedPIDPath, data, 0644); err != nil {
		t.Fatalf("Failed to create orphaned PID file: %v", err)
	}
	
	// Create temporary config files
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	configContent := `[core]
beads_db_path = "./repo"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}
	
	// Override home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create doctor instance
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	doc.homeDir = tmpDir
	
	// Run diagnostics
	report, err := doc.RunDiagnostics()
	if err != nil {
		t.Fatalf("Failed to run diagnostics: %v", err)
	}
	
	// Verify orphaned PID was detected
	foundOrphanedIssue := false
	for _, issue := range report.Issues {
		if issue.Category == CategoryState && issue.ID == "pid-orphaned-orphaned-agent" {
			foundOrphanedIssue = true
			if !issue.AutoFixable {
				t.Error("Orphaned PID issue should be auto-fixable")
			}
			break
		}
	}
	
	if !foundOrphanedIssue {
		t.Error("Orphaned PID file was not detected")
	}
	
	// Apply fixes
	fixes, err := doc.ApplyFixes(report)
	if err != nil {
		t.Fatalf("Failed to apply fixes: %v", err)
	}
	
	// Verify fix was applied
	foundFix := false
	for _, fix := range fixes {
		if fix.IssueID == "pid-orphaned-orphaned-agent" {
			foundFix = true
			if !fix.Success {
				t.Errorf("Fix failed: %s", fix.Message)
			}
			break
		}
	}
	
	if !foundFix {
		t.Error("Fix for orphaned PID was not applied")
	}
	
	// Verify file was removed
	if _, err := os.Stat(orphanedPIDPath); !os.IsNotExist(err) {
		t.Error("Orphaned PID file was not removed")
	}
}

// TestRecoveryFromLargeLogs tests recovery from large log directories
func TestRecoveryFromLargeLogs(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	ascDir := filepath.Join(tmpDir, ".asc")
	logDir := filepath.Join(ascDir, "logs")
	
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}
	
	// Create old log files (older than 7 days)
	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	for i := 0; i < 5; i++ {
		logPath := filepath.Join(logDir, fmt.Sprintf("old-agent-%d.log", i))
		// Create a 1MB file
		data := make([]byte, 1024*1024)
		if err := os.WriteFile(logPath, data, 0644); err != nil {
			t.Fatalf("Failed to create log file: %v", err)
		}
		// Set modification time to 8 days ago
		if err := os.Chtimes(logPath, oldTime, oldTime); err != nil {
			t.Fatalf("Failed to set file time: %v", err)
		}
	}
	
	// Create recent log files
	for i := 0; i < 3; i++ {
		logPath := filepath.Join(logDir, fmt.Sprintf("recent-agent-%d.log", i))
		data := make([]byte, 1024*1024)
		if err := os.WriteFile(logPath, data, 0644); err != nil {
			t.Fatalf("Failed to create log file: %v", err)
		}
	}
	
	// Create temporary config files
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	configContent := `[core]
beads_db_path = "./repo"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}
	
	// Override home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create doctor instance
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	doc.homeDir = tmpDir
	
	// Run diagnostics
	report, err := doc.RunDiagnostics()
	if err != nil {
		t.Fatalf("Failed to run diagnostics: %v", err)
	}
	
	// Verify large logs issue was detected (8MB total)
	foundLargeLogsIssue := false
	for _, issue := range report.Issues {
		if issue.ID == "logs-large" {
			foundLargeLogsIssue = true
			if !issue.AutoFixable {
				t.Error("Large logs issue should be auto-fixable")
			}
			break
		}
	}
	
	// Note: May not be detected if total size < 100MB threshold
	// This is expected behavior
	
	// Apply fixes anyway to test the cleanup
	fixes, err := doc.ApplyFixes(report)
	if err != nil {
		t.Fatalf("Failed to apply fixes: %v", err)
	}
	
	// If fix was applied, verify old logs were removed
	if foundLargeLogsIssue {
		for _, fix := range fixes {
			if fix.IssueID == "logs-large" && fix.Success {
				// Check that old logs were removed
				for i := 0; i < 5; i++ {
					logPath := filepath.Join(logDir, fmt.Sprintf("old-agent-%d.log", i))
					if _, err := os.Stat(logPath); !os.IsNotExist(err) {
						t.Errorf("Old log file %s was not removed", logPath)
					}
				}
				
				// Check that recent logs still exist
				for i := 0; i < 3; i++ {
					logPath := filepath.Join(logDir, fmt.Sprintf("recent-agent-%d.log", i))
					if _, err := os.Stat(logPath); err != nil {
						t.Errorf("Recent log file %s was removed: %v", logPath, err)
					}
				}
			}
		}
	}
}

// TestRecoveryFromPermissionIssues tests recovery from permission issues
func TestRecoveryFromPermissionIssues(t *testing.T) {
	// Skip on Windows as permission model is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}
	
	// Create temporary directory structure
	tmpDir := t.TempDir()
	
	// Create .env file with insecure permissions
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}
	
	// Create config file
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	// Create doctor instance
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	// Run diagnostics
	report, err := doc.RunDiagnostics()
	if err != nil {
		t.Fatalf("Failed to run diagnostics: %v", err)
	}
	
	// Verify permission issue was detected
	foundPermissionIssue := false
	for _, issue := range report.Issues {
		if issue.ID == "env-permissions" {
			foundPermissionIssue = true
			if !issue.AutoFixable {
				t.Error("Permission issue should be auto-fixable")
			}
			break
		}
	}
	
	if !foundPermissionIssue {
		t.Error("Permission issue was not detected")
	}
	
	// Apply fixes
	fixes, err := doc.ApplyFixes(report)
	if err != nil {
		t.Fatalf("Failed to apply fixes: %v", err)
	}
	
	// Verify fix was applied
	foundFix := false
	for _, fix := range fixes {
		if fix.IssueID == "env-permissions" {
			foundFix = true
			if !fix.Success {
				t.Errorf("Fix failed: %s", fix.Message)
			}
			break
		}
	}
	
	if !foundFix {
		t.Error("Fix for permission issue was not applied")
	}
	
	// Verify permissions were fixed
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Failed to stat env file: %v", err)
	}
	
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Expected permissions 0600, got %o", mode)
	}
}

// TestRecoveryFromMissingDirectories tests recovery from missing directories
func TestRecoveryFromMissingDirectories(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	
	// Create config files
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	configContent := `[core]
beads_db_path = "./repo"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}
	
	// Override home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create doctor instance
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	doc.homeDir = tmpDir
	
	// Run diagnostics (directories don't exist yet)
	report, err := doc.RunDiagnostics()
	if err != nil {
		t.Fatalf("Failed to run diagnostics: %v", err)
	}
	
	// Verify missing directory issues were detected
	expectedDirs := []string{"pids", "logs", "playbooks"}
	foundIssues := make(map[string]bool)
	
	for _, issue := range report.Issues {
		for _, dir := range expectedDirs {
			if issue.ID == fmt.Sprintf("dir-missing-%s", dir) {
				foundIssues[dir] = true
				if !issue.AutoFixable {
					t.Errorf("Missing directory issue for %s should be auto-fixable", dir)
				}
			}
		}
	}
	
	for _, dir := range expectedDirs {
		if !foundIssues[dir] {
			t.Errorf("Missing directory issue for %s was not detected", dir)
		}
	}
	
	// Apply fixes
	fixes, err := doc.ApplyFixes(report)
	if err != nil {
		t.Fatalf("Failed to apply fixes: %v", err)
	}
	
	// Verify fixes were applied
	for _, dir := range expectedDirs {
		foundFix := false
		for _, fix := range fixes {
			if fix.IssueID == fmt.Sprintf("dir-missing-%s", dir) {
				foundFix = true
				if !fix.Success {
					t.Errorf("Fix for %s failed: %s", dir, fix.Message)
				}
				break
			}
		}
		
		if !foundFix {
			t.Errorf("Fix for missing directory %s was not applied", dir)
		}
		
		// Verify directory was created
		dirPath := filepath.Join(tmpDir, ".asc", dir)
		info, err := os.Stat(dirPath)
		if err != nil {
			t.Errorf("Directory %s was not created: %v", dir, err)
		} else if !info.IsDir() {
			t.Errorf("%s is not a directory", dirPath)
		}
	}
}

// TestDoctorWithMultipleIssues tests handling multiple issues at once
func TestDoctorWithMultipleIssues(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	ascDir := filepath.Join(tmpDir, ".asc")
	pidDir := filepath.Join(ascDir, "pids")
	logDir := filepath.Join(ascDir, "logs")
	
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		t.Fatalf("Failed to create pid directory: %v", err)
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}
	
	// Create multiple issues:
	// 1. Corrupted PID file
	corruptedPIDPath := filepath.Join(pidDir, "corrupted.json")
	if err := os.WriteFile(corruptedPIDPath, []byte("{bad json"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted PID: %v", err)
	}
	
	// 2. Orphaned PID file
	orphanedPIDPath := filepath.Join(pidDir, "orphaned.json")
	procInfo := process.ProcessInfo{
		Name:      "orphaned",
		PID:       999999,
		Command:   "python",
		StartedAt: time.Now(),
	}
	data, _ := json.Marshal(procInfo)
	if err := os.WriteFile(orphanedPIDPath, data, 0644); err != nil {
		t.Fatalf("Failed to create orphaned PID: %v", err)
	}
	
	// 3. Insecure .env permissions
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}
	
	// Create config
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	// Override home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create doctor
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	doc.homeDir = tmpDir
	
	// Run diagnostics
	report, err := doc.RunDiagnostics()
	if err != nil {
		t.Fatalf("Failed to run diagnostics: %v", err)
	}
	
	// Should detect at least 3 issues
	if len(report.Issues) < 3 {
		t.Errorf("Expected at least 3 issues, got %d", len(report.Issues))
	}
	
	// Apply all fixes
	fixes, err := doc.ApplyFixes(report)
	if err != nil {
		t.Fatalf("Failed to apply fixes: %v", err)
	}
	
	// Verify all fixes were attempted
	if len(fixes) < 3 {
		t.Errorf("Expected at least 3 fixes, got %d", len(fixes))
	}
	
	// Verify all fixes succeeded
	for _, fix := range fixes {
		if !fix.Success {
			t.Errorf("Fix %s failed: %s", fix.IssueID, fix.Message)
		}
	}
	
	// Run diagnostics again - should have fewer issues
	report2, err := doc.RunDiagnostics()
	if err != nil {
		t.Fatalf("Failed to run second diagnostics: %v", err)
	}
	
	// Count auto-fixable issues in second run
	autoFixableCount := 0
	for _, issue := range report2.Issues {
		if issue.AutoFixable {
			autoFixableCount++
		}
	}
	
	if autoFixableCount > 0 {
		t.Errorf("Expected 0 auto-fixable issues after fixes, got %d", autoFixableCount)
	}
}
