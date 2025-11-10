package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogAggregator(t *testing.T) {
	// Create temp directory for logs
	tmpDir := t.TempDir()

	// Create some test log files
	createTestLogFile(t, tmpDir, "asc.log", []string{
		`{"timestamp":"2025-11-10 10:00:00.000","level":"INFO","message":"Starting asc","correlation_id":"test-123"}`,
		`{"timestamp":"2025-11-10 10:00:01.000","level":"DEBUG","message":"Loading config","correlation_id":"test-123"}`,
	})

	createTestLogFile(t, tmpDir, "agent1.log", []string{
		`[2025-11-10 10:00:02.000] [INFO] Processing task`,
		`[2025-11-10 10:00:03.000] [ERROR] Task failed`,
	})

	// Create aggregator
	aggregator := NewLogAggregator(tmpDir, 100)

	// Collect logs
	if err := aggregator.CollectLogs(); err != nil {
		t.Fatalf("Failed to collect logs: %v", err)
	}

	// Verify entries were collected
	filters := LogFilters{}
	entries := aggregator.GetFilteredLogs(filters)

	if len(entries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(entries))
	}

	// Verify entries are sorted by timestamp (newest first)
	for i := 0; i < len(entries)-1; i++ {
		if entries[i].Timestamp.Before(entries[i+1].Timestamp) {
			t.Error("Entries are not sorted correctly (newest first)")
		}
	}
}

func TestLogFiltering(t *testing.T) {
	// Create temp directory for logs
	tmpDir := t.TempDir()

	// Create test log files
	createTestLogFile(t, tmpDir, "agent1.log", []string{
		`{"timestamp":"2025-11-10 10:00:00.000","level":"INFO","message":"Agent1 message","agent":"agent1"}`,
		`{"timestamp":"2025-11-10 10:00:01.000","level":"DEBUG","message":"Agent1 debug","agent":"agent1"}`,
	})

	createTestLogFile(t, tmpDir, "agent2.log", []string{
		`{"timestamp":"2025-11-10 10:00:02.000","level":"INFO","message":"Agent2 message","agent":"agent2"}`,
		`{"timestamp":"2025-11-10 10:00:03.000","level":"ERROR","message":"Agent2 error","agent":"agent2"}`,
	})

	// Create aggregator
	aggregator := NewLogAggregator(tmpDir, 100)
	if err := aggregator.CollectLogs(); err != nil {
		t.Fatalf("Failed to collect logs: %v", err)
	}

	// Test agent name filter
	filters := LogFilters{AgentName: "agent1"}
	entries := aggregator.GetFilteredLogs(filters)
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries for agent1, got %d", len(entries))
	}

	// Test level filter
	filters = LogFilters{Level: WARN}
	entries = aggregator.GetFilteredLogs(filters)
	if len(entries) != 1 { // Only ERROR should pass
		t.Errorf("Expected 1 entry with level >= WARN, got %d", len(entries))
	}

	// Test search text filter
	filters = LogFilters{SearchText: "error"}
	entries = aggregator.GetFilteredLogs(filters)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry matching 'error', got %d", len(entries))
	}
}

func TestLogExport(t *testing.T) {
	// Create temp directory for logs
	tmpDir := t.TempDir()

	// Create test log files
	createTestLogFile(t, tmpDir, "test.log", []string{
		`{"timestamp":"2025-11-10 10:00:00.000","level":"INFO","message":"Test message 1"}`,
		`{"timestamp":"2025-11-10 10:00:01.000","level":"INFO","message":"Test message 2"}`,
	})

	// Create aggregator
	aggregator := NewLogAggregator(tmpDir, 100)
	if err := aggregator.CollectLogs(); err != nil {
		t.Fatalf("Failed to collect logs: %v", err)
	}

	// Export logs
	exportPath := filepath.Join(tmpDir, "export.txt")
	filters := LogFilters{}
	if err := aggregator.ExportToFile(exportPath, filters); err != nil {
		t.Fatalf("Failed to export logs: %v", err)
	}

	// Verify export file exists and has content
	content, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Export file is empty")
	}
}

func TestLogStats(t *testing.T) {
	// Create temp directory for logs
	tmpDir := t.TempDir()

	// Create test log files
	createTestLogFile(t, tmpDir, "test.log", []string{
		`{"timestamp":"2025-11-10 10:00:00.000","level":"INFO","message":"Info message"}`,
		`{"timestamp":"2025-11-10 10:00:01.000","level":"ERROR","message":"Error message"}`,
		`{"timestamp":"2025-11-10 10:00:02.000","level":"DEBUG","message":"Debug message"}`,
	})

	// Create aggregator
	aggregator := NewLogAggregator(tmpDir, 100)
	if err := aggregator.CollectLogs(); err != nil {
		t.Fatalf("Failed to collect logs: %v", err)
	}

	// Get stats
	stats := aggregator.GetStats()

	if stats.TotalEntries != 3 {
		t.Errorf("Expected 3 total entries, got %d", stats.TotalEntries)
	}

	if stats.ByLevel["INFO"] != 1 {
		t.Errorf("Expected 1 INFO entry, got %d", stats.ByLevel["INFO"])
	}

	if stats.ByLevel["ERROR"] != 1 {
		t.Errorf("Expected 1 ERROR entry, got %d", stats.ByLevel["ERROR"])
	}

	if stats.ByLevel["DEBUG"] != 1 {
		t.Errorf("Expected 1 DEBUG entry, got %d", stats.ByLevel["DEBUG"])
	}
}

func TestCleanupOldLogs(t *testing.T) {
	// Create temp directory for logs
	tmpDir := t.TempDir()

	// Create an old log file
	oldLogPath := filepath.Join(tmpDir, "old.log")
	if err := os.WriteFile(oldLogPath, []byte("old log"), 0644); err != nil {
		t.Fatalf("Failed to create old log file: %v", err)
	}

	// Set modification time to 31 days ago
	oldTime := time.Now().Add(-31 * 24 * time.Hour)
	if err := os.Chtimes(oldLogPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old time: %v", err)
	}

	// Create a recent log file
	recentLogPath := filepath.Join(tmpDir, "recent.log")
	if err := os.WriteFile(recentLogPath, []byte("recent log"), 0644); err != nil {
		t.Fatalf("Failed to create recent log file: %v", err)
	}

	// Cleanup logs older than 30 days
	if err := CleanupOldLogs(tmpDir, 30*24*time.Hour); err != nil {
		t.Fatalf("Failed to cleanup logs: %v", err)
	}

	// Verify old log was removed
	if _, err := os.Stat(oldLogPath); !os.IsNotExist(err) {
		t.Error("Old log file should have been removed")
	}

	// Verify recent log still exists
	if _, err := os.Stat(recentLogPath); err != nil {
		t.Error("Recent log file should still exist")
	}
}

// Helper function to create test log files
func createTestLogFile(t *testing.T, dir string, filename string, lines []string) {
	path := filepath.Join(dir, filename)
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	defer file.Close()

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			t.Fatalf("Failed to write to test log file: %v", err)
		}
	}
}
