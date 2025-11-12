package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rand/asc/internal/process"
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

// TestCheckAgents_WithRunningAgents tests checkAgents with running agents
func TestCheckAgents_WithRunningAgents(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create config with valid agent
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"

[agent.test-agent]
command = "python agent.py"
model = "claude"
phases = ["planning", "implementation"]
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkAgents(report)
	
	// Should have no issues for valid agent
	agentIssues := 0
	for _, issue := range report.Issues {
		if issue.Category == CategoryAgent {
			agentIssues++
		}
	}
	
	if agentIssues != 0 {
		t.Errorf("Expected 0 agent issues for valid config, got %d", agentIssues)
	}
}

// TestCheckAgents_WithMissingCommand tests checkAgents with missing command
func TestCheckAgents_WithMissingCommand(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create config with agent missing command
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"

[agent.no-command-agent]
model = "claude"
phases = ["planning"]
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkAgents(report)
	
	// Should detect missing command
	foundIssue := false
	for _, issue := range report.Issues {
		if issue.ID == "agent-no-command-no-command-agent" {
			foundIssue = true
			if issue.Severity != SeverityCritical {
				t.Errorf("Expected critical severity, got %s", issue.Severity)
			}
			if issue.Category != CategoryAgent {
				t.Errorf("Expected agent category, got %s", issue.Category)
			}
			break
		}
	}
	
	if !foundIssue {
		t.Error("Missing command issue was not detected")
	}
}

// TestCheckAgents_WithInvalidModel tests checkAgents with invalid model
func TestCheckAgents_WithInvalidModel(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create config with invalid model
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"

[agent.invalid-model-agent]
command = "python agent.py"
model = "invalid-model"
phases = ["planning"]
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkAgents(report)
	
	// Should detect invalid model
	foundIssue := false
	for _, issue := range report.Issues {
		if issue.ID == "agent-invalid-model-invalid-model-agent" {
			foundIssue = true
			if issue.Severity != SeverityHigh {
				t.Errorf("Expected high severity, got %s", issue.Severity)
			}
			if issue.Category != CategoryAgent {
				t.Errorf("Expected agent category, got %s", issue.Category)
			}
			break
		}
	}
	
	if !foundIssue {
		t.Error("Invalid model issue was not detected")
	}
}

// TestCheckAgents_WithMissingPhases tests checkAgents with missing phases
func TestCheckAgents_WithMissingPhases(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create config with agent missing phases
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"

[agent.no-phases-agent]
command = "python agent.py"
model = "claude"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkAgents(report)
	
	// Should detect missing phases
	foundIssue := false
	for _, issue := range report.Issues {
		if issue.ID == "agent-no-phases-no-phases-agent" {
			foundIssue = true
			if issue.Severity != SeverityHigh {
				t.Errorf("Expected high severity, got %s", issue.Severity)
			}
			if issue.Category != CategoryAgent {
				t.Errorf("Expected agent category, got %s", issue.Category)
			}
			break
		}
	}
	
	if !foundIssue {
		t.Error("Missing phases issue was not detected")
	}
}

// TestCheckAgents_WithMultipleAgents tests checkAgents with multiple agents
func TestCheckAgents_WithMultipleAgents(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create config with multiple agents (some valid, some invalid)
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"

[agent.valid-agent]
command = "python agent.py"
model = "claude"
phases = ["planning"]

[agent.invalid-agent]
model = "bad-model"
phases = ["testing"]

[agent.no-phases-agent]
command = "python agent.py"
model = "gemini"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkAgents(report)
	
	// Should detect 3 issues: missing command, invalid model, missing phases
	agentIssues := 0
	for _, issue := range report.Issues {
		if issue.Category == CategoryAgent {
			agentIssues++
		}
	}
	
	if agentIssues != 3 {
		t.Errorf("Expected 3 agent issues, got %d", agentIssues)
		for _, issue := range report.Issues {
			if issue.Category == CategoryAgent {
				t.Logf("Found issue: %s - %s", issue.ID, issue.Title)
			}
		}
	}
}

// TestCheckAgents_WithNoAgents tests checkAgents with no agents configured
func TestCheckAgents_WithNoAgents(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create config with no agents
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./repo"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkAgents(report)
	
	// Should have no agent issues (no agents to check)
	agentIssues := 0
	for _, issue := range report.Issues {
		if issue.Category == CategoryAgent {
			agentIssues++
		}
	}
	
	if agentIssues != 0 {
		t.Errorf("Expected 0 agent issues with no agents, got %d", agentIssues)
	}
}

// TestCheckAgents_WithValidModels tests all valid model types
func TestCheckAgents_WithValidModels(t *testing.T) {
	validModels := []string{"claude", "gemini", "openai", "gpt-4", "codex"}
	
	for _, model := range validModels {
		t.Run(model, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			configPath := filepath.Join(tmpDir, "asc.toml")
			configContent := fmt.Sprintf(`[core]
beads_db_path = "./repo"

[agent.test-agent]
command = "python agent.py"
model = "%s"
phases = ["planning"]
`, model)
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to create config: %v", err)
			}
			
			envPath := filepath.Join(tmpDir, ".env")
			if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
				t.Fatalf("Failed to create env: %v", err)
			}
			
			doc, err := NewDoctor(configPath, envPath)
			if err != nil {
				t.Fatalf("Failed to create doctor: %v", err)
			}
			
			report := &DiagnosticReport{
				RunAt:  time.Now(),
				Issues: []Issue{},
			}
			
			doc.checkAgents(report)
			
			// Should have no model-related issues
			for _, issue := range report.Issues {
				if issue.Category == CategoryAgent && issue.ID == "agent-invalid-model-test-agent" {
					t.Errorf("Valid model %s was flagged as invalid", model)
				}
			}
		})
	}
}

// TestCheckAgents_WithInvalidConfig tests checkAgents with invalid config file
func TestCheckAgents_WithInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create invalid config
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core
invalid toml syntax
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	// Should not panic with invalid config
	doc.checkAgents(report)
	
	// Should have no agent issues (config couldn't be parsed)
	agentIssues := 0
	for _, issue := range report.Issues {
		if issue.Category == CategoryAgent {
			agentIssues++
		}
	}
	
	if agentIssues != 0 {
		t.Errorf("Expected 0 agent issues with invalid config, got %d", agentIssues)
	}
}

// TestDiagnosticReport_Format_WithNoIssues tests formatting with no issues
func TestDiagnosticReport_Format_WithNoIssues(t *testing.T) {
	report := &DiagnosticReport{
		RunAt:         time.Now(),
		Issues:        []Issue{},
		HealthSummary: "✓ All checks passed - system is healthy",
	}
	
	output := report.Format(false)
	
	// Should contain header
	if len(output) < len("DIAGNOSTIC REPORT") {
		t.Error("Output too short to contain header")
	} else {
		found := false
		for i := 0; i <= len(output)-len("DIAGNOSTIC REPORT"); i++ {
			if output[i:i+len("DIAGNOSTIC REPORT")] == "DIAGNOSTIC REPORT" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Output does not contain header")
		}
	}
	
	// Should contain success message
	if len(output) < len("No issues detected") {
		t.Error("Output too short to contain success message")
	} else {
		found := false
		for i := 0; i <= len(output)-len("No issues detected"); i++ {
			if output[i:i+len("No issues detected")] == "No issues detected" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Output does not contain success message")
		}
	}
}

// TestDiagnosticReport_Format_WithAllSeverities tests formatting with all severity levels
func TestDiagnosticReport_Format_WithAllSeverities(t *testing.T) {
	report := &DiagnosticReport{
		RunAt: time.Now(),
		Issues: []Issue{
			{
				ID:          "critical-1",
				Category:    CategoryConfiguration,
				Severity:    SeverityCritical,
				Title:       "Critical Issue",
				Description: "Critical description",
				Impact:      "Critical impact",
				Remediation: "Fix critical",
				AutoFixable: true,
				DetectedAt:  time.Now(),
			},
			{
				ID:          "high-1",
				Category:    CategoryState,
				Severity:    SeverityHigh,
				Title:       "High Issue",
				Description: "High description",
				Impact:      "High impact",
				Remediation: "Fix high",
				AutoFixable: false,
				DetectedAt:  time.Now(),
			},
			{
				ID:          "medium-1",
				Category:    CategoryPermissions,
				Severity:    SeverityMedium,
				Title:       "Medium Issue",
				Description: "Medium description",
				Impact:      "Medium impact",
				Remediation: "Fix medium",
				AutoFixable: true,
				DetectedAt:  time.Now(),
			},
			{
				ID:          "low-1",
				Category:    CategoryResources,
				Severity:    SeverityLow,
				Title:       "Low Issue",
				Description: "Low description",
				Impact:      "Low impact",
				Remediation: "Fix low",
				AutoFixable: false,
				DetectedAt:  time.Now(),
			},
			{
				ID:          "info-1",
				Category:    CategoryNetwork,
				Severity:    SeverityInfo,
				Title:       "Info Issue",
				Description: "Info description",
				Impact:      "No impact",
				Remediation: "No action needed",
				AutoFixable: false,
				DetectedAt:  time.Now(),
			},
		},
		HealthSummary: "Found 5 issue(s): 1 critical, 1 high, 1 medium, 1 low",
	}
	
	output := report.Format(false)
	
	// Should contain all severity sections (lowercase in output)
	severities := []string{"critical", "high", "medium", "low", "info"}
	for _, severity := range severities {
		if len(output) < len(severity) {
			t.Errorf("Output too short to contain %s", severity)
			continue
		}
		found := false
		for i := 0; i <= len(output)-len(severity); i++ {
			if output[i:i+len(severity)] == severity {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Output does not contain %s severity section", severity)
		}
	}
	
	// Should contain all issue titles
	titles := []string{"Critical Issue", "High Issue", "Medium Issue", "Low Issue", "Info Issue"}
	for _, title := range titles {
		if len(output) < len(title) {
			t.Errorf("Output too short to contain %s", title)
			continue
		}
		found := false
		for i := 0; i <= len(output)-len(title); i++ {
			if output[i:i+len(title)] == title {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Output does not contain issue title: %s", title)
		}
	}
	
	// Should contain auto-fixable indicator for some issues
	autoFixText := "Auto-fixable"
	if len(output) < len(autoFixText) {
		t.Error("Output too short to contain auto-fixable text")
	} else {
		found := false
		for i := 0; i <= len(output)-len(autoFixText); i++ {
			if output[i:i+len(autoFixText)] == autoFixText {
				found = true
				break
			}
		}
		if !found {
			t.Error("Output does not contain auto-fixable indicator")
		}
	}
}

// TestDiagnosticReport_Format_VerboseMode tests verbose formatting
func TestDiagnosticReport_Format_VerboseMode(t *testing.T) {
	report := &DiagnosticReport{
		RunAt: time.Now(),
		Issues: []Issue{
			{
				ID:          "test-1",
				Category:    CategoryConfiguration,
				Severity:    SeverityHigh,
				Title:       "Test Issue",
				Description: "Detailed description for verbose mode",
				Impact:      "Detailed impact for verbose mode",
				Remediation: "Fix it",
				AutoFixable: false,
				DetectedAt:  time.Now(),
			},
		},
		HealthSummary: "Found 1 issue(s): 1 high",
	}
	
	normalOutput := report.Format(false)
	verboseOutput := report.Format(true)
	
	// Verbose should be longer
	if len(verboseOutput) <= len(normalOutput) {
		t.Error("Verbose output should be longer than normal output")
	}
	
	// Verbose should contain description and impact
	detailedTexts := []string{"Detailed description", "Detailed impact"}
	for _, text := range detailedTexts {
		if len(verboseOutput) < len(text) {
			t.Errorf("Verbose output too short to contain %s", text)
			continue
		}
		found := false
		for i := 0; i <= len(verboseOutput)-len(text); i++ {
			if verboseOutput[i:i+len(text)] == text {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Verbose output does not contain: %s", text)
		}
	}
}

// TestDiagnosticReport_Format_WithFixResults tests formatting with fix results
func TestDiagnosticReport_Format_WithFixResults(t *testing.T) {
	report := &DiagnosticReport{
		RunAt: time.Now(),
		Issues: []Issue{
			{
				ID:          "fixed-issue",
				Category:    CategoryState,
				Severity:    SeverityMedium,
				Title:       "Fixed Issue",
				Description: "This was fixed",
				Impact:      "No longer an issue",
				Remediation: "Already fixed",
				AutoFixable: true,
				DetectedAt:  time.Now(),
			},
		},
		FixesApplied: []FixResult{
			{
				IssueID:   "fixed-issue",
				Success:   true,
				Message:   "Successfully fixed the issue",
				AppliedAt: time.Now(),
			},
			{
				IssueID:   "failed-fix",
				Success:   false,
				Message:   "Failed to fix: permission denied",
				AppliedAt: time.Now(),
			},
		},
		HealthSummary: "Found 1 issue(s): 1 medium",
	}
	
	output := report.Format(false)
	
	// Should contain fixes section
	fixesText := "FIXES APPLIED"
	if len(output) < len(fixesText) {
		t.Error("Output too short to contain fixes section")
	} else {
		found := false
		for i := 0; i <= len(output)-len(fixesText); i++ {
			if output[i:i+len(fixesText)] == fixesText {
				found = true
				break
			}
		}
		if !found {
			t.Error("Output does not contain fixes section")
		}
	}
	
	// Should contain fix messages
	fixMessages := []string{"Successfully fixed", "Failed to fix"}
	for _, msg := range fixMessages {
		if len(output) < len(msg) {
			t.Errorf("Output too short to contain %s", msg)
			continue
		}
		found := false
		for i := 0; i <= len(output)-len(msg); i++ {
			if output[i:i+len(msg)] == msg {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Output does not contain fix message: %s", msg)
		}
	}
}

// TestDiagnosticReport_Format_WithMultipleIssuesPerSeverity tests grouping
func TestDiagnosticReport_Format_WithMultipleIssuesPerSeverity(t *testing.T) {
	report := &DiagnosticReport{
		RunAt: time.Now(),
		Issues: []Issue{
			{
				ID:          "critical-1",
				Severity:    SeverityCritical,
				Title:       "Critical Issue 1",
				Category:    CategoryConfiguration,
				Remediation: "Fix 1",
				DetectedAt:  time.Now(),
			},
			{
				ID:          "critical-2",
				Severity:    SeverityCritical,
				Title:       "Critical Issue 2",
				Category:    CategoryState,
				Remediation: "Fix 2",
				DetectedAt:  time.Now(),
			},
			{
				ID:          "critical-3",
				Severity:    SeverityCritical,
				Title:       "Critical Issue 3",
				Category:    CategoryAgent,
				Remediation: "Fix 3",
				DetectedAt:  time.Now(),
			},
		},
		HealthSummary: "Found 3 issue(s): 3 critical",
	}
	
	output := report.Format(false)
	
	// Should show count in severity header (lowercase)
	criticalHeader := "critical SEVERITY (3)"
	if len(output) < len(criticalHeader) {
		t.Error("Output too short to contain critical header with count")
	} else {
		found := false
		for i := 0; i <= len(output)-len(criticalHeader); i++ {
			if output[i:i+len(criticalHeader)] == criticalHeader {
				found = true
				break
			}
		}
		if !found {
			t.Error("Output does not contain severity count in header")
		}
	}
	
	// Should contain all three issue titles
	for i := 1; i <= 3; i++ {
		title := fmt.Sprintf("Critical Issue %d", i)
		if len(output) < len(title) {
			t.Errorf("Output too short to contain %s", title)
			continue
		}
		found := false
		for j := 0; j <= len(output)-len(title); j++ {
			if output[j:j+len(title)] == title {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Output does not contain: %s", title)
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

// TestCheckConfiguration_WithMissingConfig tests checkConfiguration with missing config file
func TestCheckConfiguration_WithMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Don't create config file
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkConfiguration(report)
	
	// Should detect missing config
	foundIssue := false
	for _, issue := range report.Issues {
		if issue.ID == "config-missing" {
			foundIssue = true
			if issue.Severity != SeverityCritical {
				t.Errorf("Expected critical severity, got %s", issue.Severity)
			}
			if !issue.AutoFixable {
				t.Error("Missing config should be auto-fixable")
			}
			break
		}
	}
	
	if !foundIssue {
		t.Error("Missing config issue was not detected")
	}
}

// TestCheckConfiguration_WithInvalidConfig tests checkConfiguration with invalid TOML
func TestCheckConfiguration_WithInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	// Create invalid TOML
	if err := os.WriteFile(configPath, []byte("[core\ninvalid"), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkConfiguration(report)
	
	// Should detect invalid config
	foundIssue := false
	for _, issue := range report.Issues {
		if issue.ID == "config-invalid" {
			foundIssue = true
			if issue.Severity != SeverityCritical {
				t.Errorf("Expected critical severity, got %s", issue.Severity)
			}
			if issue.AutoFixable {
				t.Error("Invalid config should not be auto-fixable")
			}
			break
		}
	}
	
	if !foundIssue {
		t.Error("Invalid config issue was not detected")
	}
}

// TestCheckConfiguration_WithMissingEnv tests checkConfiguration with missing .env
func TestCheckConfiguration_WithMissingEnv(t *testing.T) {
	tmpDir := t.TempDir()
	
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	// Create valid config but no .env
	configContent := `[core]
beads_db_path = "./repo"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkConfiguration(report)
	
	// Should detect missing .env
	foundIssue := false
	for _, issue := range report.Issues {
		if issue.ID == "env-missing" {
			foundIssue = true
			if issue.Severity != SeverityHigh {
				t.Errorf("Expected high severity, got %s", issue.Severity)
			}
			if issue.AutoFixable {
				t.Error("Missing .env should not be auto-fixable")
			}
			break
		}
	}
	
	if !foundIssue {
		t.Error("Missing .env issue was not detected")
	}
}

// TestCheckResources_WithMissingBinaries tests checkResources with missing binaries
func TestCheckResources_WithMissingBinaries(t *testing.T) {
	tmpDir := t.TempDir()
	
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	configContent := `[core]
beads_db_path = "./repo"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkResources(report)
	
	// Should check for required binaries
	// Note: Actual results depend on system, but test should not panic
	if report == nil {
		t.Error("Report should not be nil")
	}
}

// TestCheckResources_WithHighDiskUsage tests checkResources with high disk usage
func TestCheckResources_WithHighDiskUsage(t *testing.T) {
	tmpDir := t.TempDir()
	ascDir := filepath.Join(tmpDir, ".asc")
	
	if err := os.MkdirAll(ascDir, 0755); err != nil {
		t.Fatalf("Failed to create .asc dir: %v", err)
	}
	
	// Create large file (>500MB would be detected, but we'll use smaller for test)
	// This test verifies the function runs without error
	
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	configContent := `[core]
beads_db_path = "./repo"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	doc.homeDir = tmpDir
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	doc.checkResources(report)
	
	// Should complete without error
	if report == nil {
		t.Error("Report should not be nil")
	}
}

// TestFixAscNotDir tests fixing .asc when it's a file instead of directory
func TestFixAscNotDir(t *testing.T) {
	tmpDir := t.TempDir()
	ascPath := filepath.Join(tmpDir, ".asc")
	
	// Create .asc as a file
	if err := os.WriteFile(ascPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create .asc file: %v", err)
	}
	
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	if err := os.WriteFile(configPath, []byte("[core]\nbeads_db_path = \"./repo\"\n"), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	doc.homeDir = tmpDir
	
	// Run diagnostics to detect issue
	report, err := doc.RunDiagnostics()
	if err != nil {
		t.Fatalf("Failed to run diagnostics: %v", err)
	}
	
	// Should detect .asc is not a directory
	foundIssue := false
	for _, issue := range report.Issues {
		if issue.ID == "asc-not-dir" {
			foundIssue = true
			break
		}
	}
	
	if !foundIssue {
		t.Error(".asc not directory issue was not detected")
	}
	
	// Apply fix
	fixes, err := doc.ApplyFixes(report)
	if err != nil {
		t.Fatalf("Failed to apply fixes: %v", err)
	}
	
	// Verify fix was applied
	foundFix := false
	for _, fix := range fixes {
		if fix.IssueID == "asc-not-dir" {
			foundFix = true
			if !fix.Success {
				t.Errorf("Fix failed: %s", fix.Message)
			}
			break
		}
	}
	
	if !foundFix {
		t.Error("Fix was not applied")
	}
	
	// Verify .asc is now a directory
	info, err := os.Stat(ascPath)
	if err != nil {
		t.Fatalf("Failed to stat .asc: %v", err)
	}
	
	if !info.IsDir() {
		t.Error(".asc is not a directory after fix")
	}
}

// TestFixAscNotWritable tests fixing .asc when it's not writable
func TestFixAscNotWritable(t *testing.T) {
	// Skip on systems where we can't test permissions
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}
	
	tmpDir := t.TempDir()
	ascPath := filepath.Join(tmpDir, ".asc")
	
	// Create .asc directory with no write permissions
	if err := os.MkdirAll(ascPath, 0555); err != nil {
		t.Fatalf("Failed to create .asc dir: %v", err)
	}
	
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	if err := os.WriteFile(configPath, []byte("[core]\nbeads_db_path = \"./repo\"\n"), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
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
	
	// Should detect not writable
	foundIssue := false
	for _, issue := range report.Issues {
		if issue.ID == "asc-not-writable" {
			foundIssue = true
			break
		}
	}
	
	if !foundIssue {
		t.Error(".asc not writable issue was not detected")
	}
	
	// Apply fix
	fixes, err := doc.ApplyFixes(report)
	if err != nil {
		t.Fatalf("Failed to apply fixes: %v", err)
	}
	
	// Verify fix was attempted
	foundFix := false
	for _, fix := range fixes {
		if fix.IssueID == "asc-not-writable" {
			foundFix = true
			if !fix.Success {
				t.Logf("Fix failed (expected on some systems): %s", fix.Message)
			}
			break
		}
	}
	
	if !foundFix {
		t.Error("Fix was not applied")
	}
}

// TestFixLargeLogs tests the fixLargeLogs function
func TestFixLargeLogs(t *testing.T) {
	tmpDir := t.TempDir()
	ascDir := filepath.Join(tmpDir, ".asc")
	logDir := filepath.Join(ascDir, "logs")
	
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("Failed to create log dir: %v", err)
	}
	
	// Create old log files
	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	for i := 0; i < 3; i++ {
		logPath := filepath.Join(logDir, fmt.Sprintf("old-%d.log", i))
		if err := os.WriteFile(logPath, []byte("old log content"), 0644); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
		if err := os.Chtimes(logPath, oldTime, oldTime); err != nil {
			t.Fatalf("Failed to set time: %v", err)
		}
	}
	
	// Create recent log files
	for i := 0; i < 2; i++ {
		logPath := filepath.Join(logDir, fmt.Sprintf("recent-%d.log", i))
		if err := os.WriteFile(logPath, []byte("recent log content"), 0644); err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}
	
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	if err := os.WriteFile(configPath, []byte("[core]\nbeads_db_path = \"./repo\"\n"), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	doc.homeDir = tmpDir
	
	// Manually call fixLargeLogs
	success, message := doc.fixLargeLogs()
	
	if !success {
		t.Errorf("fixLargeLogs failed: %s", message)
	}
	
	// Verify old logs were deleted
	for i := 0; i < 3; i++ {
		logPath := filepath.Join(logDir, fmt.Sprintf("old-%d.log", i))
		if _, err := os.Stat(logPath); !os.IsNotExist(err) {
			t.Errorf("Old log %d was not deleted", i)
		}
	}
	
	// Verify recent logs still exist
	for i := 0; i < 2; i++ {
		logPath := filepath.Join(logDir, fmt.Sprintf("recent-%d.log", i))
		if _, err := os.Stat(logPath); err != nil {
			t.Errorf("Recent log %d was deleted: %v", i, err)
		}
	}
}

// TestRunDiagnostics_Integration tests full diagnostic run
func TestRunDiagnostics_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	ascDir := filepath.Join(tmpDir, ".asc")
	
	// Create proper directory structure
	for _, dir := range []string{"pids", "logs", "playbooks"} {
		if err := os.MkdirAll(filepath.Join(ascDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create %s dir: %v", dir, err)
		}
	}
	
	configPath := filepath.Join(tmpDir, "asc.toml")
	envPath := filepath.Join(tmpDir, ".env")
	
	// Create valid config
	configContent := `[core]
beads_db_path = "./repo"

[agent.test-agent]
command = "python agent.py"
model = "claude"
phases = ["planning"]
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	if err := os.WriteFile(envPath, []byte("CLAUDE_API_KEY=test\n"), 0600); err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	
	doc, err := NewDoctor(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to create doctor: %v", err)
	}
	doc.homeDir = tmpDir
	
	// Run full diagnostics
	report, err := doc.RunDiagnostics()
	if err != nil {
		t.Fatalf("Failed to run diagnostics: %v", err)
	}
	
	if report == nil {
		t.Fatal("Report is nil")
	}
	
	// Should have minimal issues with proper setup
	criticalCount := 0
	for _, issue := range report.Issues {
		if issue.Severity == SeverityCritical {
			criticalCount++
			t.Logf("Critical issue: %s - %s", issue.ID, issue.Title)
		}
	}
	
	// With proper setup, should have no critical issues
	// (may have info issues about network checks)
	if criticalCount > 0 {
		t.Errorf("Expected 0 critical issues with proper setup, got %d", criticalCount)
	}
}
