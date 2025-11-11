package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestEnvironment represents a test environment with temporary directories and files
type TestEnvironment struct {
	TempDir    string
	ConfigPath string
	EnvPath    string
	PIDDir     string
	LogDir     string
	BackupDir  string
	t          *testing.T
}

// NewTestEnvironment creates a new test environment with temporary directories
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	tempDir := t.TempDir()

	env := &TestEnvironment{
		TempDir:    tempDir,
		ConfigPath: filepath.Join(tempDir, "asc.toml"),
		EnvPath:    filepath.Join(tempDir, ".env"),
		PIDDir:     filepath.Join(tempDir, ".asc", "pids"),
		LogDir:     filepath.Join(tempDir, ".asc", "logs"),
		BackupDir:  filepath.Join(tempDir, ".asc_backup"),
		t:          t,
	}

	// Create necessary directories
	if err := os.MkdirAll(env.PIDDir, 0755); err != nil {
		t.Fatalf("Failed to create PID directory: %v", err)
	}
	if err := os.MkdirAll(env.LogDir, 0755); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}
	if err := os.MkdirAll(env.BackupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	return env
}

// WriteConfig writes a configuration file to the test environment
func (e *TestEnvironment) WriteConfig(content string) {
	e.t.Helper()
	if err := os.WriteFile(e.ConfigPath, []byte(content), 0644); err != nil {
		e.t.Fatalf("Failed to write config file: %v", err)
	}
}

// WriteEnv writes an environment file to the test environment
func (e *TestEnvironment) WriteEnv(content string) {
	e.t.Helper()
	if err := os.WriteFile(e.EnvPath, []byte(content), 0600); err != nil {
		e.t.Fatalf("Failed to write env file: %v", err)
	}
}

// WritePIDFile writes a PID file for a process
func (e *TestEnvironment) WritePIDFile(name string, content string) {
	e.t.Helper()
	pidFile := filepath.Join(e.PIDDir, name+".json")
	if err := os.WriteFile(pidFile, []byte(content), 0644); err != nil {
		e.t.Fatalf("Failed to write PID file: %v", err)
	}
}

// FileExists checks if a file exists in the test environment
func (e *TestEnvironment) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile reads a file from the test environment
func (e *TestEnvironment) ReadFile(path string) string {
	e.t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		e.t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(content)
}

// Cleanup cleans up the test environment (automatically called by t.TempDir())
func (e *TestEnvironment) Cleanup() {
	// No-op since t.TempDir() handles cleanup automatically
}

// CaptureOutput captures stdout and stderr during command execution
type CaptureOutput struct {
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
	oldOut   *os.File
	oldErr   *os.File
	rOut     *os.File
	wOut     *os.File
	rErr     *os.File
	wErr     *os.File
	doneChan chan bool
}

// NewCaptureOutput creates a new output capture
func NewCaptureOutput() *CaptureOutput {
	return &CaptureOutput{
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
		doneChan: make(chan bool, 2),
	}
}

// Start begins capturing output
func (c *CaptureOutput) Start() {
	c.oldOut = os.Stdout
	c.oldErr = os.Stderr

	c.rOut, c.wOut, _ = os.Pipe()
	c.rErr, c.wErr, _ = os.Pipe()

	os.Stdout = c.wOut
	os.Stderr = c.wErr

	go func() {
		io.Copy(c.Stdout, c.rOut)
		c.doneChan <- true
	}()
	go func() {
		io.Copy(c.Stderr, c.rErr)
		c.doneChan <- true
	}()
}

// Stop stops capturing output and restores original stdout/stderr
func (c *CaptureOutput) Stop() {
	if c.wOut != nil {
		c.wOut.Close()
	}
	if c.wErr != nil {
		c.wErr.Close()
	}
	
	// Wait for copy to complete
	<-c.doneChan
	<-c.doneChan
	
	if c.rOut != nil {
		c.rOut.Close()
	}
	if c.rErr != nil {
		c.rErr.Close()
	}
	
	if c.oldOut != nil {
		os.Stdout = c.oldOut
	}
	if c.oldErr != nil {
		os.Stderr = c.oldErr
	}
}

// GetStdout returns captured stdout as a string
func (c *CaptureOutput) GetStdout() string {
	return c.Stdout.String()
}

// GetStderr returns captured stderr as a string
func (c *CaptureOutput) GetStderr() string {
	return c.Stderr.String()
}

// MockExitFunc is a mock function for os.Exit
var MockExitFunc func(int)

// MockExit mocks os.Exit for testing
func MockExit(code int) {
	if MockExitFunc != nil {
		MockExitFunc(code)
	}
}

// ExitRecorder records exit codes for testing
type ExitRecorder struct {
	Called   bool
	ExitCode int
}

// NewExitRecorder creates a new exit recorder
func NewExitRecorder() *ExitRecorder {
	return &ExitRecorder{}
}

// Record records an exit code
func (e *ExitRecorder) Record(code int) {
	e.Called = true
	e.ExitCode = code
}

// Reset resets the exit recorder
func (e *ExitRecorder) Reset() {
	e.Called = false
	e.ExitCode = 0
}

// Sample configuration files for testing

// ValidConfig returns a valid asc.toml configuration
func ValidConfig() string {
	return `[core]
beads_db_path = "./project-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test-agent]
command = "python agent_adapter.py"
model = "claude"
phases = ["planning", "implementation"]
`
}

// MinimalConfig returns a minimal valid configuration
func MinimalConfig() string {
	return `[core]
beads_db_path = "./project-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"
`
}

// InvalidConfig returns an invalid configuration (malformed TOML)
func InvalidConfig() string {
	return `[core
beads_db_path = "./project-repo"
invalid toml syntax
`
}

// EmptyConfig returns an empty configuration
func EmptyConfig() string {
	return ``
}

// ValidEnv returns a valid .env file
func ValidEnv() string {
	return `CLAUDE_API_KEY=sk-test-claude-key-12345
OPENAI_API_KEY=sk-test-openai-key-67890
GOOGLE_API_KEY=test-google-key-abcde
`
}

// PartialEnv returns a .env file with some missing keys
func PartialEnv() string {
	return `CLAUDE_API_KEY=sk-test-claude-key-12345
`
}

// EmptyEnv returns an empty .env file
func EmptyEnv() string {
	return ``
}

// ValidPIDFile returns a valid PID file JSON
func ValidPIDFile(pid int, name string) string {
	return `{
  "pid": ` + string(rune(pid)) + `,
  "name": "` + name + `",
  "command": "python",
  "args": ["-m", "mcp_agent_mail.server"],
  "started_at": "2025-11-11T10:00:00Z",
  "log_file": "/tmp/.asc/logs/` + name + `.log"
}`
}

// SetupMockBinaries creates mock binaries in a temporary bin directory
func SetupMockBinaries(t *testing.T, binaries []string) string {
	t.Helper()

	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}

	for _, binary := range binaries {
		binPath := filepath.Join(binDir, binary)
		// Create a simple shell script that exits successfully
		content := "#!/bin/sh\nexit 0\n"
		if err := os.WriteFile(binPath, []byte(content), 0755); err != nil {
			t.Fatalf("Failed to create mock binary %s: %v", binary, err)
		}
	}

	return binDir
}

// WithMockPath temporarily modifies PATH to ONLY include mock binaries
// This ensures tests don't accidentally find real binaries
func WithMockPath(t *testing.T, mockBinDir string, fn func()) {
	t.Helper()

	oldPath := os.Getenv("PATH")
	// Set PATH to ONLY the mock bin directory
	os.Setenv("PATH", mockBinDir)

	defer func() {
		os.Setenv("PATH", oldPath)
	}()

	fn()
}

// ChangeToTempDir changes the current directory to a temporary directory
func ChangeToTempDir(t *testing.T, dir string) func() {
	t.Helper()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	return func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("Failed to restore directory: %v", err)
		}
	}
}

// RunWithExitCapture runs a function and captures any os.Exit calls
// Returns the exit code and whether os.Exit was called
func RunWithExitCapture(fn func()) (exitCode int, exitCalled bool) {
	exitCode = -1
	exitCalled = false
	
	// Recover from os.Exit panic
	defer func() {
		if r := recover(); r != nil {
			// Check if it's our expected panic from os.Exit
			if str, ok := r.(string); ok && str == "os.Exit called" {
				// This is expected, don't re-panic
				return
			}
			// Re-panic if it's a different error
			panic(r)
		}
	}()
	
	// Wrap os.Exit to capture the exit code
	oldOsExit := osExit
	osExit = func(code int) {
		exitCode = code
		exitCalled = true
		panic("os.Exit called") // Use panic to stop execution
	}
	defer func() { osExit = oldOsExit }()
	
	fn()
	return
}
