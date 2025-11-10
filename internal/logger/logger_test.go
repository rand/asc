package logger

import (
	"encoding/json"
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
