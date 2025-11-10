package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
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

// Logger provides structured logging with rotation
type Logger struct {
	mu           sync.Mutex
	file         *os.File
	logPath      string
	maxSize      int64
	maxBackups   int
	currentSize  int64
	minLevel     LogLevel
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the default logger
func Init() error {
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
	})
	return err
}

// NewLogger creates a new logger instance
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
		file:        file,
		logPath:     logPath,
		maxSize:     maxSize,
		maxBackups:  maxBackups,
		currentSize: stat.Size(),
		minLevel:    minLevel,
	}

	return logger, nil
}

// Close closes the log file
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

// log writes a log message with the given level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level.String(), message)

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

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
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

// SetLevel sets the minimum log level for the default logger
func SetLevel(level LogLevel) {
	if defaultLogger != nil {
		defaultLogger.SetLevel(level)
	}
}

// Close closes the default logger
func Close() error {
	if defaultLogger != nil {
		return defaultLogger.Close()
	}
	return nil
}
