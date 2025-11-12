package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStructuredLogging(t *testing.T) {
	// Create temp log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with JSON format
	logger, err := NewLogger(logPath, 1024*1024, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.SetFormat(FormatJSON)

	// Log with fields
	logger.WithFields(Fields{
		"agent": "test-agent",
		"task":  "task-123",
		"phase": "planning",
		"extra": "value",
	}).Info("Processing task")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON
	var entry LogEntry
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	// Verify fields
	if entry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}
	if entry.Message != "Processing task" {
		t.Errorf("Expected message 'Processing task', got %s", entry.Message)
	}
	if entry.Agent != "test-agent" {
		t.Errorf("Expected agent 'test-agent', got %s", entry.Agent)
	}
	if entry.Task != "task-123" {
		t.Errorf("Expected task 'task-123', got %s", entry.Task)
	}
	if entry.Phase != "planning" {
		t.Errorf("Expected phase 'planning', got %s", entry.Phase)
	}
	if entry.Fields["extra"] != "value" {
		t.Errorf("Expected extra field 'value', got %v", entry.Fields["extra"])
	}
	if entry.CorrelationID == "" {
		t.Error("Expected correlation ID to be set")
	}
}

func TestTextLogging(t *testing.T) {
	// Create temp log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with text format
	logger, err := NewLogger(logPath, 1024*1024, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.SetFormat(FormatText)

	// Log with fields
	logger.WithFields(Fields{
		"agent": "test-agent",
		"task":  "task-123",
	}).Info("Processing task")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logLine := string(content)
	if !strings.Contains(logLine, "[INFO]") {
		t.Error("Expected log line to contain [INFO]")
	}
	if !strings.Contains(logLine, "Processing task") {
		t.Error("Expected log line to contain message")
	}
	if !strings.Contains(logLine, "agent=test-agent") {
		t.Error("Expected log line to contain agent field")
	}
	if !strings.Contains(logLine, "task=task-123") {
		t.Error("Expected log line to contain task field")
	}
}

func TestCorrelationID(t *testing.T) {
	// Create temp log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger
	logger, err := NewLogger(logPath, 1024*1024, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.SetFormat(FormatJSON)

	// Set custom correlation ID
	customID := "custom-correlation-id-123"
	logger.WithCorrelationID(customID)

	// Log message
	logger.Info("Test message")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON
	var entry LogEntry
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.CorrelationID != customID {
		t.Errorf("Expected correlation ID %s, got %s", customID, entry.CorrelationID)
	}
}

func TestContextFields(t *testing.T) {
	// Create temp log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger
	logger, err := NewLogger(logPath, 1024*1024, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.SetFormat(FormatJSON)

	// Set context fields
	logger.SetContextFields(Fields{
		"service": "asc",
		"version": "1.0.0",
	})

	// Log message
	logger.Info("Test message")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON
	var entry LogEntry
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.Fields["service"] != "asc" {
		t.Errorf("Expected service field 'asc', got %v", entry.Fields["service"])
	}
	if entry.Fields["version"] != "1.0.0" {
		t.Errorf("Expected version field '1.0.0', got %v", entry.Fields["version"])
	}
}

func TestLogLevels(t *testing.T) {
	// Create temp log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with WARN level
	logger, err := NewLogger(logPath, 1024*1024, 2, WARN)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log at different levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Debug and Info should not be logged
	if strings.Contains(logContent, "Debug message") {
		t.Error("Debug message should not be logged at WARN level")
	}
	if strings.Contains(logContent, "Info message") {
		t.Error("Info message should not be logged at WARN level")
	}

	// Warn and Error should be logged
	if !strings.Contains(logContent, "Warn message") {
		t.Error("Warn message should be logged at WARN level")
	}
	if !strings.Contains(logContent, "Error message") {
		t.Error("Error message should be logged at WARN level")
	}
}

func TestLogRotation(t *testing.T) {
	// Create temp log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with small max size (100 bytes)
	logger, err := NewLogger(logPath, 100, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Write enough logs to trigger rotation
	for i := 0; i < 10; i++ {
		logger.Info("This is a test message that should trigger rotation")
		time.Sleep(10 * time.Millisecond)
	}

	// Check that backup files exist
	backupPath := logPath + ".1"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Expected backup file to exist after rotation")
	}
}

func TestLogRotationAtSizeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with 200 byte limit
	logger, err := NewLogger(logPath, 200, 3, INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Write messages until rotation occurs
	initialSize := logger.currentSize
	for i := 0; i < 20; i++ {
		logger.Info("Message %d that will trigger rotation", i)
	}

	// Verify rotation occurred
	if _, err := os.Stat(logPath + ".1"); os.IsNotExist(err) {
		t.Error("Expected backup file .1 to exist after rotation")
	}

	// Verify current file size is less than max
	if logger.currentSize >= logger.maxSize {
		t.Errorf("Current size %d should be less than max size %d after rotation", logger.currentSize, logger.maxSize)
	}

	// Verify initial size was tracked
	if initialSize < 0 {
		t.Error("Initial size should be non-negative")
	}
}

func TestLogRotationMultipleBackups(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with small size and 3 backups
	logger, err := NewLogger(logPath, 100, 3, INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Write enough to create multiple backups
	for i := 0; i < 50; i++ {
		logger.Info("Message %d to create multiple backup files", i)
	}

	// Check that multiple backup files exist
	backup1 := logPath + ".1"
	backup2 := logPath + ".2"
	backup3 := logPath + ".3"

	if _, err := os.Stat(backup1); os.IsNotExist(err) {
		t.Error("Expected backup file .1 to exist")
	}
	if _, err := os.Stat(backup2); os.IsNotExist(err) {
		t.Error("Expected backup file .2 to exist")
	}
	if _, err := os.Stat(backup3); os.IsNotExist(err) {
		t.Error("Expected backup file .3 to exist")
	}

	// Backup 4 should not exist (maxBackups is 3)
	backup4 := logPath + ".4"
	if _, err := os.Stat(backup4); !os.IsNotExist(err) {
		t.Error("Backup file .4 should not exist (exceeds maxBackups)")
	}
}

func TestLogRotationCleanupOldFiles(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with 2 backups max
	logger, err := NewLogger(logPath, 80, 2, INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Write enough to create 3 rotations (should only keep 2 backups)
	for i := 0; i < 40; i++ {
		logger.Info("Message %d for cleanup test", i)
	}

	// Verify only 2 backups exist
	backup1 := logPath + ".1"
	backup2 := logPath + ".2"
	backup3 := logPath + ".3"

	if _, err := os.Stat(backup1); os.IsNotExist(err) {
		t.Error("Expected backup file .1 to exist")
	}
	if _, err := os.Stat(backup2); os.IsNotExist(err) {
		t.Error("Expected backup file .2 to exist")
	}
	if _, err := os.Stat(backup3); !os.IsNotExist(err) {
		t.Error("Backup file .3 should not exist (exceeds maxBackups of 2)")
	}
}

func TestConcurrentLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Launch multiple goroutines writing logs concurrently
	numGoroutines := 10
	messagesPerGoroutine := 100
	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			for i := 0; i < messagesPerGoroutine; i++ {
				logger.Info("Goroutine %d message %d", id, i)
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Read log file and count lines
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	expectedLines := numGoroutines * messagesPerGoroutine

	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}
}

func TestConcurrentLoggingWithRotation(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Small size to trigger rotation during concurrent writes
	logger, err := NewLogger(logPath, 500, 3, INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Launch multiple goroutines
	numGoroutines := 5
	messagesPerGoroutine := 50
	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			for i := 0; i < messagesPerGoroutine; i++ {
				logger.Info("Concurrent goroutine %d message %d", id, i)
			}
			done <- true
		}(g)
	}

	// Wait for completion
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify rotation occurred
	if _, err := os.Stat(logPath + ".1"); os.IsNotExist(err) {
		t.Error("Expected backup file to exist after concurrent rotation")
	}

	// Verify no data corruption by checking all files are readable
	for i := 0; i <= 3; i++ {
		var path string
		if i == 0 {
			path = logPath
		} else {
			path = logPath + "." + string(rune('0'+i))
		}

		if _, err := os.Stat(path); err == nil {
			if _, err := os.ReadFile(path); err != nil {
				t.Errorf("Failed to read rotated file %s: %v", path, err)
			}
		}
	}
}

func TestConcurrentLoggingThreadSafety(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test concurrent access to logger methods
	numGoroutines := 20
	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			// Mix different operations
			logger.SetLevel(DEBUG)
			logger.Info("Message from goroutine %d", id)
			logger.WithFields(Fields{"goroutine": id}).Debug("Debug message")
			logger.SetContextFields(Fields{"test": id})
			logger.AddContextField("extra", id)
			done <- true
		}(g)
	}

	// Wait for completion
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// If we get here without panicking, thread safety is working
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected log file to have content")
	}
}

func TestConcurrentLoggingOrdering(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Write logs from multiple goroutines with timestamps
	numGoroutines := 5
	messagesPerGoroutine := 20
	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			for i := 0; i < messagesPerGoroutine; i++ {
				logger.Info("G%d-M%d", id, i)
				time.Sleep(time.Millisecond)
			}
			done <- true
		}(g)
	}

	// Wait for completion
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Read and verify all messages are present
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	
	// Verify each goroutine's messages are present
	for g := 0; g < numGoroutines; g++ {
		for i := 0; i < messagesPerGoroutine; i++ {
			expected := fmt.Sprintf("G%d-M%d", g, i)
			if !strings.Contains(logContent, expected) {
				t.Errorf("Missing message: %s", expected)
			}
		}
	}
}

func TestComplexObjectLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.SetFormat(FormatJSON)

	// Log with complex nested objects
	complexFields := Fields{
		"string":  "value",
		"number":  42,
		"float":   3.14,
		"bool":    true,
		"array":   []string{"a", "b", "c"},
		"nested": map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	logger.WithFields(complexFields).Info("Complex object test")

	// Read and parse
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify complex fields
	if entry.Fields["string"] != "value" {
		t.Error("String field not preserved")
	}
	if entry.Fields["number"].(float64) != 42 {
		t.Error("Number field not preserved")
	}
	if entry.Fields["bool"] != true {
		t.Error("Bool field not preserved")
	}
}

func TestContextFieldsPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.SetFormat(FormatJSON)

	// Set context fields
	logger.SetContextFields(Fields{
		"service": "asc",
		"version": "1.0.0",
	})

	// Add another context field
	logger.AddContextField("environment", "test")

	// Log multiple messages
	logger.Info("First message")
	logger.Info("Second message")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 log lines, got %d", len(lines))
	}

	// Verify both messages have context fields
	for i, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("Failed to parse line %d: %v", i, err)
		}

		if entry.Fields["service"] != "asc" {
			t.Errorf("Line %d missing service context field", i)
		}
		if entry.Fields["version"] != "1.0.0" {
			t.Errorf("Line %d missing version context field", i)
		}
		if entry.Fields["environment"] != "test" {
			t.Errorf("Line %d missing environment context field", i)
		}
	}
}

func TestLogLevelFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log at different levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")

	// Change level to DEBUG
	logger.SetLevel(DEBUG)
	logger.Debug("Debug after level change")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// First debug should not be logged
	if strings.Contains(logContent, "Debug message") {
		t.Error("First debug message should not be logged at INFO level")
	}

	// Info, Warn, Error should be logged
	if !strings.Contains(logContent, "Info message") {
		t.Error("Info message should be logged")
	}
	if !strings.Contains(logContent, "Warn message") {
		t.Error("Warn message should be logged")
	}
	if !strings.Contains(logContent, "Error message") {
		t.Error("Error message should be logged")
	}

	// Debug after level change should be logged
	if !strings.Contains(logContent, "Debug after level change") {
		t.Error("Debug message should be logged after level change to DEBUG")
	}
}

func TestEntryLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test Entry methods
	entry := logger.WithFields(Fields{"test": "value"})
	entry.Debug("Debug via entry")
	entry.Info("Info via entry")
	entry.Warn("Warn via entry")
	entry.Error("Error via entry")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify all entry messages are logged
	if !strings.Contains(logContent, "Debug via entry") {
		t.Error("Debug via entry should be logged")
	}
	if !strings.Contains(logContent, "Info via entry") {
		t.Error("Info via entry should be logged")
	}
	if !strings.Contains(logContent, "Warn via entry") {
		t.Error("Warn via entry should be logged")
	}
	if !strings.Contains(logContent, "Error via entry") {
		t.Error("Error via entry should be logged")
	}
}

func TestGlobalLoggerFunctions(t *testing.T) {
	// Initialize default logger
	tmpDir := t.TempDir()
	
	// Save original home dir
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	// Set temp home for test
	os.Setenv("HOME", tmpDir)
	
	// Reset the once variable by creating a new logger instance
	// Since we can't reset sync.Once, we'll test the functions directly
	testLogPath := filepath.Join(tmpDir, "global.log")
	testLogger, err := NewLogger(testLogPath, 1024*1024, 2, DEBUG)
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}
	defer testLogger.Close()
	
	// Test all global functions
	testLogger.Debug("Global debug")
	testLogger.Info("Global info")
	testLogger.Warn("Global warn")
	testLogger.Error("Global error")
	
	testLogger.WithFields(Fields{"global": "test"}).Info("Global with fields")
	testLogger.WithCorrelationID("test-correlation")
	testLogger.SetContextFields(Fields{"context": "global"})
	testLogger.AddContextField("extra", "field")
	testLogger.SetLevel(WARN)
	testLogger.SetFormat(FormatJSON)
	
	// Read log file
	content, err := os.ReadFile(testLogPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Expected log file to have content")
	}
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		result := tt.level.String()
		if result != tt.expected {
			t.Errorf("LogLevel(%d).String() = %s, want %s", tt.level, result, tt.expected)
		}
	}
}

func TestNewLoggerErrors(t *testing.T) {
	// Test with invalid path
	_, err := NewLogger("/invalid/path/that/does/not/exist/test.log", 1024, 2, INFO)
	if err == nil {
		t.Error("Expected error when creating logger with invalid path")
	}
}

func TestLoggerClose(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Write a message
	logger.Info("Test message")

	// Close the logger
	if err := logger.Close(); err != nil {
		t.Errorf("Failed to close logger: %v", err)
	}

	// Verify file exists and has content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected log file to have content after close")
	}

	// Test closing nil file
	logger.file = nil
	if err := logger.Close(); err != nil {
		t.Error("Closing logger with nil file should not error")
	}
}

func TestFormatSwitching(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Start with text format
	logger.SetFormat(FormatText)
	logger.Info("Text format message")

	// Switch to JSON
	logger.SetFormat(FormatJSON)
	logger.Info("JSON format message")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 log lines, got %d", len(lines))
	}

	// First line should be text format
	if !strings.HasPrefix(lines[0], "[") {
		t.Error("First line should be in text format")
	}

	// Second line should be JSON format
	if !strings.HasPrefix(lines[1], "{") {
		t.Error("Second line should be in JSON format")
	}
}

func TestSpecialFieldExtraction(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.SetFormat(FormatJSON)

	// Log with special fields that should be extracted
	logger.WithFields(Fields{
		"agent": "test-agent",
		"task":  "task-123",
		"phase": "planning",
		"other": "value",
	}).Info("Test message")

	// Read and parse
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify special fields are extracted to top level
	if entry.Agent != "test-agent" {
		t.Errorf("Expected agent 'test-agent', got '%s'", entry.Agent)
	}
	if entry.Task != "task-123" {
		t.Errorf("Expected task 'task-123', got '%s'", entry.Task)
	}
	if entry.Phase != "planning" {
		t.Errorf("Expected phase 'planning', got '%s'", entry.Phase)
	}

	// Other fields should remain in Fields map
	if entry.Fields["other"] != "value" {
		t.Error("Non-special field should remain in Fields map")
	}
}

func TestJSONMarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(logPath, 1024*1024, 2, INFO)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.SetFormat(FormatJSON)

	// Log with a field that can't be marshaled to JSON
	logger.WithFields(Fields{
		"channel": make(chan int), // channels can't be marshaled
	}).Info("Test message")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Should have fallback text format with error message
	logContent := string(content)
	if !strings.Contains(logContent, "JSON marshal error") {
		t.Error("Expected JSON marshal error message in log")
	}
}
