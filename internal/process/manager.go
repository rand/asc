package process

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// ProcessStatus represents the current state of a process
type ProcessStatus string

const (
	StatusRunning ProcessStatus = "running"
	StatusStopped ProcessStatus = "stopped"
	StatusError   ProcessStatus = "error"
)

// ProcessInfo contains metadata about a managed process
type ProcessInfo struct {
	Name      string            `json:"name"`
	PID       int               `json:"pid"`
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env"`
	StartedAt time.Time         `json:"started_at"`
	LogFile   string            `json:"log_file"`
}

// ProcessManager defines the interface for managing background processes
type ProcessManager interface {
	// Start launches a new process with the given name, command, and environment
	Start(name string, command string, args []string, env []string) (int, error)

	// Stop terminates a process by PID
	Stop(pid int) error

	// StopAll terminates all managed processes
	StopAll() error

	// IsRunning checks if a process with the given PID is running
	IsRunning(pid int) bool

	// GetStatus returns the status of a process by PID
	GetStatus(pid int) ProcessStatus

	// GetProcessInfo returns metadata about a managed process by name
	GetProcessInfo(name string) (*ProcessInfo, error)

	// ListProcesses returns all managed processes
	ListProcesses() ([]*ProcessInfo, error)
}

// Manager implements the ProcessManager interface
type Manager struct {
	pidDir string
	logDir string
}

// NewManager creates a new process manager with the specified directories
func NewManager(pidDir, logDir string) (*Manager, error) {
	// Create directories if they don't exist
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create PID directory: %w", err)
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &Manager{
		pidDir: pidDir,
		logDir: logDir,
	}, nil
}

// Start launches a new process
func (m *Manager) Start(name string, command string, args []string, env []string) (int, error) {
	// Create log file
	logPath := filepath.Join(m.logDir, fmt.Sprintf("%s.log", name))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFile.Close()

	// Create command
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Set process group for proper cleanup
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start process: %w", err)
	}

	pid := cmd.Process.Pid

	// Convert env slice to map for storage
	envMap := make(map[string]string)
	for _, e := range env {
		// Parse KEY=VALUE format
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				envMap[e[:i]] = e[i+1:]
				break
			}
		}
	}

	// Save process info
	info := &ProcessInfo{
		Name:      name,
		PID:       pid,
		Command:   command,
		Args:      args,
		Env:       envMap,
		StartedAt: time.Now(),
		LogFile:   logPath,
	}

	if err := m.saveProcessInfo(info); err != nil {
		// Try to kill the process if we can't save its info
		_ = cmd.Process.Kill()
		return 0, fmt.Errorf("failed to save process info: %w", err)
	}

	return pid, nil
}

// Stop terminates a process by PID
func (m *Manager) Stop(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Send SIGTERM for graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Wait for graceful shutdown with timeout
	done := make(chan error, 1)
	go func() {
		_, err := process.Wait()
		done <- err
	}()

	select {
	case <-time.After(5 * time.Second):
		// Timeout - send SIGKILL
		if err := process.Signal(syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to send SIGKILL: %w", err)
		}
		// Wait for SIGKILL to complete
		<-done
	case err := <-done:
		if err != nil && err.Error() != "signal: terminated" && err.Error() != "signal: killed" {
			return fmt.Errorf("process wait error: %w", err)
		}
	}

	return nil
}

// StopAll terminates all managed processes
func (m *Manager) StopAll() error {
	processes, err := m.ListProcesses()
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}

	var errors []error
	for _, info := range processes {
		if m.IsRunning(info.PID) {
			if err := m.Stop(info.PID); err != nil {
				errors = append(errors, fmt.Errorf("failed to stop %s (PID %d): %w", info.Name, info.PID, err))
			}
		}
		// Clean up PID file
		if err := m.deleteProcessInfo(info.Name); err != nil {
			errors = append(errors, fmt.Errorf("failed to delete PID file for %s: %w", info.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping processes: %v", errors)
	}

	return nil
}

// IsRunning checks if a process with the given PID is running
func (m *Manager) IsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// GetStatus returns the status of a process by PID
func (m *Manager) GetStatus(pid int) ProcessStatus {
	if m.IsRunning(pid) {
		return StatusRunning
	}
	return StatusStopped
}

// GetProcessInfo returns metadata about a managed process by name
func (m *Manager) GetProcessInfo(name string) (*ProcessInfo, error) {
	pidFile := filepath.Join(m.pidDir, fmt.Sprintf("%s.json", name))
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("process %s not found", name)
		}
		return nil, fmt.Errorf("failed to read PID file: %w", err)
	}

	var info ProcessInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse PID file: %w", err)
	}

	return &info, nil
}

// ListProcesses returns all managed processes
func (m *Manager) ListProcesses() ([]*ProcessInfo, error) {
	entries, err := os.ReadDir(m.pidDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read PID directory: %w", err)
	}

	var processes []*ProcessInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		name := entry.Name()[:len(entry.Name())-5] // Remove .json extension
		info, err := m.GetProcessInfo(name)
		if err != nil {
			continue // Skip invalid entries
		}
		processes = append(processes, info)
	}

	return processes, nil
}

// saveProcessInfo saves process metadata to a JSON file
func (m *Manager) saveProcessInfo(info *ProcessInfo) error {
	pidFile := filepath.Join(m.pidDir, fmt.Sprintf("%s.json", info.Name))
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal process info: %w", err)
	}

	if err := os.WriteFile(pidFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// deleteProcessInfo removes the PID file for a process
func (m *Manager) deleteProcessInfo(name string) error {
	pidFile := filepath.Join(m.pidDir, fmt.Sprintf("%s.json", name))
	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete PID file: %w", err)
	}
	return nil
}
