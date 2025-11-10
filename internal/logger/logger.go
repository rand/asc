// Package logger provides structured logging with automatic file rotation
// for the Agent Stack Controller. It supports multiple log levels and
// thread-safe concurrent logging with JSON formatting for machine-parseable logs.
//
// Example usage:
//
//	if err := logger.Init(); err != nil {
//	    log.Fatal(err)
//	}
//	defer logger.Close()
//
//	logger.Info("Starting agent stack")
//	logger.WithFields(Fields{"agent": "planner", "task": "123"}).Info("Processing task")
//	logger.Error("Failed to start agent: %v", err)
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents the severity of a log message.
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Fields represents structured context fields for logging
type Fields map[string]interface{}

// LogFormat represents the output format for logs
type LogFormat int

const (
	FormatText LogFormat = iota
	FormatJSON
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp     string                 `json:"timestamp"`
	Level         string                 `json:"level"`
	Message       string                 `json:"message"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Agent         string                 `json:"agent,omitempty"`
	Task          string                 `json:"task,omitempty"`
	Phase         string                 `json:"phase,omitempty"`
	Fields        map[string]interface{} `json:"fields,omitempty"`
}

// Logger provides structured logging with automatic file rotation.
// It is thread-safe and supports configurable log levels and rotation policies.
type Logger struct {
	mu            sync.Mutex
	file          *os.File
	logPath       string
	maxSize       int64
	maxBackups    int
	currentSize   int64
	minLevel      LogLevel
	format        LogFormat
	correlationID string
	contextFields Fields
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Entry creates a new log entry with context fields
type Entry struct {
	logger *Logger
	fields Fields
}

// Init initializes the default logger with standard settings.
// The log file is created at ~/.asc/logs/asc.log with a 10MB max size,
// 5 backup files, and INFO level. This should be called once at startup.
func Init() error {
	return InitWithFormat(FormatText)
}

// InitWithFormat initializes the default logger with a specific format.
// Use FormatJSON for machine-parseable logs or FormatText for human-readable logs.
func InitWithFormat(format LogFormat) error {
	var err error
	once.Do(func() {
		homeDir, e := os.UserHomeDir()
		if e != nil {
			err = fmt.Errorf("failed to get home directory: %w", e)
			return
		}

		logDir := filepath.Join(homeDir, ".asc", "logs")
		if e := os.MkdirAll(logDir, 0755); e != nil {
			err = fmt.Errorf("failed to create log directory: %w", e)
			return
		}

		logPath := filepath.Join(logDir, "asc.log")
		defaultLogger, err = NewLogger(logPath, 10*1024*1024, 5, INFO)
		if err == nil {
			defaultLogger.SetFormat(format)
		}
	})
	return err
}

// NewLogger creates a new logger instance with custom settings.
// maxSize is in bytes, maxBackups is the number of old log files to keep,
// and minLevel is the minimum severity level to log.
//
// Example:
//
//	logger, err := logger.NewLogger("/var/log/app.log", 10*1024*1024, 5, logger.INFO)
func NewLogger(logPath string, maxSize int64, maxBackups int, minLevel LogLevel) (*Logger, error) {
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat log file: %w", err)
	}

	logger := &Logger{
		file:          file,
		logPath:       logPath,
		maxSize:       maxSize,
		maxBackups:    maxBackups,
		currentSize:   stat.Size(),
		minLevel:      minLevel,
		format:        FormatText,
		correlationID: uuid.New().String(),
		contextFields: make(Fields),
	}

	return logger, nil
}

// Close closes the log file. This should be called when shutting down
// to ensure all buffered data is written.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// rotate rotates the log file if it exceeds maxSize
func (l *Logger) rotate() error {
	if l.currentSize < l.maxSize {
		return nil
	}

	// Close current file
	if err := l.file.Close(); err != nil {
		return err
	}

	// Rotate existing backups
	for i := l.maxBackups - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.logPath, i)
		newPath := fmt.Sprintf("%s.%d", l.logPath, i+1)
		os.Rename(oldPath, newPath) // Ignore errors if file doesn't exist
	}

	// Move current log to .1
	backupPath := fmt.Sprintf("%s.1", l.logPath)
	if err := os.Rename(l.logPath, backupPath); err != nil {
		return err
	}

	// Open new log file
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.file = file
	l.currentSize = 0
	return nil
}

// log writes a log message with the given level and optional fields
func (l *Logger) log(level LogLevel, fields Fields, format string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)

	var logLine string
	if l.format == FormatJSON {
		entry := LogEntry{
			Timestamp:     timestamp,
			Level:         level.String(),
			Message:       message,
			CorrelationID: l.correlationID,
			Fields:        make(map[string]interface{}),
		}

		// Merge context fields and provided fields
		for k, v := range l.contextFields {
			entry.Fields[k] = v
		}
		for k, v := range fields {
			// Extract special fields
			switch k {
			case "agent":
				if s, ok := v.(string); ok {
					entry.Agent = s
				}
			case "task":
				if s, ok := v.(string); ok {
					entry.Task = s
				}
			case "phase":
				if s, ok := v.(string); ok {
					entry.Phase = s
				}
			default:
				entry.Fields[k] = v
			}
		}

		jsonBytes, err := json.Marshal(entry)
		if err != nil {
			logLine = fmt.Sprintf("[%s] [%s] %s (JSON marshal error: %v)\n", timestamp, level.String(), message, err)
		} else {
			logLine = string(jsonBytes) + "\n"
		}
	} else {
		// Text format
		logLine = fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), message)
		if len(fields) > 0 || len(l.contextFields) > 0 {
			logLine += " {"
			first := true
			for k, v := range l.contextFields {
				if !first {
					logLine += ", "
				}
				logLine += fmt.Sprintf("%s=%v", k, v)
				first = false
			}
			for k, v := range fields {
				if !first {
					logLine += ", "
				}
				logLine += fmt.Sprintf("%s=%v", k, v)
				first = false
			}
			logLine += "}"
		}
		logLine += "\n"
	}

	// Check if rotation is needed
	if err := l.rotate(); err != nil {
		// If rotation fails, try to write anyway
		fmt.Fprintf(os.Stderr, "Failed to rotate log: %v\n", err)
	}

	n, err := io.WriteString(l.file, logLine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
		return
	}

	l.currentSize += int64(n)
}

// Debug logs a debug message with printf-style formatting.
// Only logged if the logger's level is DEBUG or lower.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, nil, format, args...)
}

// Info logs an info message with printf-style formatting.
// Only logged if the logger's level is INFO or lower.
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, nil, format, args...)
}

// Warn logs a warning message with printf-style formatting.
// Only logged if the logger's level is WARN or lower.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, nil, format, args...)
}

// Error logs an error message with printf-style formatting.
// Always logged regardless of the logger's level.
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, nil, format, args...)
}

// WithFields creates a new Entry with the given fields for structured logging
func (l *Logger) WithFields(fields Fields) *Entry {
	return &Entry{
		logger: l,
		fields: fields,
	}
}

// WithCorrelationID sets a correlation ID for tracing requests across components
func (l *Logger) WithCorrelationID(id string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.correlationID = id
	return l
}

// SetContextFields sets persistent context fields that will be included in all log entries
func (l *Logger) SetContextFields(fields Fields) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.contextFields = fields
}

// AddContextField adds a single context field
func (l *Logger) AddContextField(key string, value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.contextFields[key] = value
}

// SetLevel sets the minimum log level. Messages below this level
// will not be logged. Thread-safe.
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// SetFormat sets the log output format (text or JSON)
func (l *Logger) SetFormat(format LogFormat) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.format = format
}

// Entry methods for structured logging

// Debug logs a debug message with the entry's fields
func (e *Entry) Debug(format string, args ...interface{}) {
	if e.logger != nil {
		e.logger.log(DEBUG, e.fields, format, args...)
	}
}

// Info logs an info message with the entry's fields
func (e *Entry) Info(format string, args ...interface{}) {
	if e.logger != nil {
		e.logger.log(INFO, e.fields, format, args...)
	}
}

// Warn logs a warning message with the entry's fields
func (e *Entry) Warn(format string, args ...interface{}) {
	if e.logger != nil {
		e.logger.log(WARN, e.fields, format, args...)
	}
}

// Error logs an error message with the entry's fields
func (e *Entry) Error(format string, args ...interface{}) {
	if e.logger != nil {
		e.logger.log(ERROR, e.fields, format, args...)
	}
}

// Global logging functions using the default logger

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	}
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	}
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(format, args...)
	}
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, args...)
	}
}

// WithFields creates a new Entry with the given fields using the default logger
func WithFields(fields Fields) *Entry {
	if defaultLogger != nil {
		return defaultLogger.WithFields(fields)
	}
	// Return a no-op entry if logger is not initialized
	return &Entry{logger: &Logger{minLevel: ERROR + 1}, fields: fields}
}

// WithCorrelationID sets a correlation ID for the default logger
func WithCorrelationID(id string) *Logger {
	if defaultLogger != nil {
		return defaultLogger.WithCorrelationID(id)
	}
	return nil
}

// SetContextFields sets persistent context fields for the default logger
func SetContextFields(fields Fields) {
	if defaultLogger != nil {
		defaultLogger.SetContextFields(fields)
	}
}

// AddContextField adds a single context field to the default logger
func AddContextField(key string, value interface{}) {
	if defaultLogger != nil {
		defaultLogger.AddContextField(key, value)
	}
}

// SetLevel sets the minimum log level for the default logger
func SetLevel(level LogLevel) {
	if defaultLogger != nil {
		defaultLogger.SetLevel(level)
	}
}

// SetFormat sets the log output format for the default logger
func SetFormat(format LogFormat) {
	if defaultLogger != nil {
		defaultLogger.SetFormat(format)
	}
}

// Close closes the default logger
func Close() error {
	if defaultLogger != nil {
		return defaultLogger.Close()
	}
	return nil
}
