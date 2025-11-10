package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// LogAggregator collects and aggregates logs from multiple sources
type LogAggregator struct {
	mu          sync.RWMutex
	logsDir     string
	entries     []AggregatedEntry
	maxEntries  int
	filters     LogFilters
}

// AggregatedEntry represents a log entry from any source with metadata
type AggregatedEntry struct {
	Timestamp     time.Time
	Level         string
	Message       string
	Source        string // Agent name or "asc" for main logs
	CorrelationID string
	Agent         string
	Task          string
	Phase         string
	Fields        map[string]interface{}
}

// LogFilters defines filters for log aggregation
type LogFilters struct {
	AgentName   string      // Filter by agent name
	MessageType string      // Filter by message type
	Level       LogLevel    // Minimum log level
	Since       time.Time   // Only logs after this time
	SearchText  string      // Search in message text
}

// NewLogAggregator creates a new log aggregator
func NewLogAggregator(logsDir string, maxEntries int) *LogAggregator {
	return &LogAggregator{
		logsDir:    logsDir,
		entries:    make([]AggregatedEntry, 0, maxEntries),
		maxEntries: maxEntries,
	}
}

// CollectLogs reads logs from all log files in the logs directory
func (a *LogAggregator) CollectLogs() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Clear existing entries
	a.entries = make([]AggregatedEntry, 0, a.maxEntries)

	// Read all log files in the directory
	files, err := os.ReadDir(a.logsDir)
	if err != nil {
		return fmt.Errorf("failed to read logs directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".log") {
			continue
		}

		logPath := filepath.Join(a.logsDir, file.Name())
		source := strings.TrimSuffix(file.Name(), ".log")

		if err := a.readLogFile(logPath, source); err != nil {
			// Log error but continue with other files
			Error("Failed to read log file %s: %v", logPath, err)
		}
	}

	// Sort entries by timestamp (newest first)
	sort.Slice(a.entries, func(i, j int) bool {
		return a.entries[i].Timestamp.After(a.entries[j].Timestamp)
	})

	// Trim to max entries
	if len(a.entries) > a.maxEntries {
		a.entries = a.entries[:a.maxEntries]
	}

	return nil
}

// readLogFile reads a single log file and parses entries
func (a *LogAggregator) readLogFile(path string, source string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry, err := a.parseLogLine(line, source)
		if err != nil {
			// Skip unparseable lines
			continue
		}

		a.entries = append(a.entries, entry)
	}

	return scanner.Err()
}

// parseLogLine parses a log line (JSON or text format)
func (a *LogAggregator) parseLogLine(line string, source string) (AggregatedEntry, error) {
	// Try JSON format first
	if strings.HasPrefix(line, "{") {
		var logEntry LogEntry
		if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
			timestamp, _ := time.Parse("2006-01-02 15:04:05.000", logEntry.Timestamp)
			return AggregatedEntry{
				Timestamp:     timestamp,
				Level:         logEntry.Level,
				Message:       logEntry.Message,
				Source:        source,
				CorrelationID: logEntry.CorrelationID,
				Agent:         logEntry.Agent,
				Task:          logEntry.Task,
				Phase:         logEntry.Phase,
				Fields:        logEntry.Fields,
			}, nil
		}
	}

	// Try text format: [timestamp] [level] message
	parts := strings.SplitN(line, "]", 3)
	if len(parts) < 3 {
		return AggregatedEntry{}, fmt.Errorf("invalid log format")
	}

	timestampStr := strings.TrimPrefix(parts[0], "[")
	level := strings.TrimSpace(strings.TrimPrefix(parts[1], "["))
	message := strings.TrimSpace(parts[2])

	timestamp, err := time.Parse("2006-01-02 15:04:05.000", timestampStr)
	if err != nil {
		return AggregatedEntry{}, err
	}

	return AggregatedEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   message,
		Source:    source,
	}, nil
}

// GetFilteredLogs returns logs matching the current filters
func (a *LogAggregator) GetFilteredLogs(filters LogFilters) []AggregatedEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	filtered := make([]AggregatedEntry, 0)

	for _, entry := range a.entries {
		if !a.matchesFilters(entry, filters) {
			continue
		}
		filtered = append(filtered, entry)
	}

	return filtered
}

// matchesFilters checks if an entry matches the given filters
func (a *LogAggregator) matchesFilters(entry AggregatedEntry, filters LogFilters) bool {
	// Filter by agent name
	if filters.AgentName != "" && entry.Source != filters.AgentName && entry.Agent != filters.AgentName {
		return false
	}

	// Filter by level
	entryLevel := parseLogLevel(entry.Level)
	if entryLevel < filters.Level {
		return false
	}

	// Filter by time
	if !filters.Since.IsZero() && entry.Timestamp.Before(filters.Since) {
		return false
	}

	// Filter by search text
	if filters.SearchText != "" {
		searchLower := strings.ToLower(filters.SearchText)
		if !strings.Contains(strings.ToLower(entry.Message), searchLower) &&
			!strings.Contains(strings.ToLower(entry.Source), searchLower) {
			return false
		}
	}

	return true
}

// parseLogLevel converts a string level to LogLevel
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// ExportToFile exports filtered logs to a file
func (a *LogAggregator) ExportToFile(path string, filters LogFilters) error {
	logs := a.GetFilteredLogs(filters)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, entry := range logs {
		line := fmt.Sprintf("[%s] [%s] [%s] %s\n",
			entry.Timestamp.Format("2006-01-02 15:04:05.000"),
			entry.Level,
			entry.Source,
			entry.Message,
		)
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("failed to write log entry: %w", err)
		}
	}

	return nil
}

// GetStats returns statistics about the aggregated logs
func (a *LogAggregator) GetStats() LogStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := LogStats{
		TotalEntries: len(a.entries),
		ByLevel:      make(map[string]int),
		BySource:     make(map[string]int),
	}

	for _, entry := range a.entries {
		stats.ByLevel[entry.Level]++
		stats.BySource[entry.Source]++
	}

	if len(a.entries) > 0 {
		stats.OldestEntry = a.entries[len(a.entries)-1].Timestamp
		stats.NewestEntry = a.entries[0].Timestamp
	}

	return stats
}

// LogStats contains statistics about aggregated logs
type LogStats struct {
	TotalEntries int
	ByLevel      map[string]int
	BySource     map[string]int
	OldestEntry  time.Time
	NewestEntry  time.Time
}

// CleanupOldLogs removes log files older than the specified duration
func CleanupOldLogs(logsDir string, maxAge time.Duration) error {
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return fmt.Errorf("failed to read logs directory: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".log") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			logPath := filepath.Join(logsDir, file.Name())
			if err := os.Remove(logPath); err != nil {
				Error("Failed to remove old log file %s: %v", logPath, err)
			} else {
				Info("Removed old log file: %s", file.Name())
			}
		}
	}

	return nil
}
