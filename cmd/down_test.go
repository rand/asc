package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDownCommand_Success tests successful shutdown with stale processes
func TestDownCommand_Success(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)

	// Change to temp directory
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override home directory for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create PID files for multiple processes
	// Use non-existent PIDs to avoid signal issues in tests
	stalePID := 999999

	// Create agent PID file
	agentPIDContent := fmt.Sprintf(`{
  "pid": %d,
  "name": "test-agent",
  "command": "python",
  "args": ["agent_adapter.py"],
  "started_at": "%s",
  "log_file": "%s"
}`, stalePID, time.Now().Format(time.RFC3339), filepath.Join(env.LogDir, "test-agent.log"))
	env.WritePIDFile("test-agent", agentPIDContent)

	// Create MCP service PID file
	mcpPIDContent := fmt.Sprintf(`{
  "pid": %d,
  "name": "mcp_agent_mail",
  "command": "python",
  "args": ["-m", "mcp_agent_mail.server"],
  "started_at": "%s",
  "log_file": "%s"
}`, stalePID+1, time.Now().Format(time.RFC3339), filepath.Join(env.LogDir, "mcp_agent_mail.log"))
	env.WritePIDFile("mcp_agent_mail", mcpPIDContent)

	// Run down command
	// Note: Down command doesn't call osExit on success
	runDown(downCmd, []string{})

	// Verify PID files were cleaned up
	agentPIDFile := filepath.Join(env.PIDDir, "test-agent.json")
	mcpPIDFile := filepath.Join(env.PIDDir, "mcp_agent_mail.json")

	if env.FileExists(agentPIDFile) {
		t.Error("Expected agent PID file to be cleaned up")
	}
	if env.FileExists(mcpPIDFile) {
		t.Error("Expected MCP PID file to be cleaned up")
	}
}

// TestDownCommand_NoProcesses tests down command when no processes are running
func TestDownCommand_NoProcesses(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)

	// Change to temp directory
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override home directory for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Don't create any PID files

	// Run down command
	// Should succeed and print "No running processes found"
	runDown(downCmd, []string{})

	// Test passes if no panic/exit occurs
}

// TestDownCommand_StalePIDFiles tests down command with stale PID files
func TestDownCommand_StalePIDFiles(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)

	// Change to temp directory
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override home directory for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create PID files with non-existent PIDs
	stalePID := 999999

	stalePIDContent := fmt.Sprintf(`{
  "pid": %d,
  "name": "stale-agent",
  "command": "python",
  "args": ["agent_adapter.py"],
  "started_at": "%s",
  "log_file": "%s"
}`, stalePID, time.Now().Format(time.RFC3339), filepath.Join(env.LogDir, "stale-agent.log"))
	env.WritePIDFile("stale-agent", stalePIDContent)

	// Run down command
	runDown(downCmd, []string{})

	// Verify stale PID file was cleaned up
	stalePIDFile := filepath.Join(env.PIDDir, "stale-agent.json")
	if env.FileExists(stalePIDFile) {
		t.Error("Expected stale PID file to be cleaned up")
	}
}

// TestDownCommand_MixedProcesses tests down with multiple stale processes
func TestDownCommand_MixedProcesses(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)

	// Change to temp directory
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override home directory for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create multiple PID files with non-existent PIDs
	// This simulates processes that have already stopped
	stalePID1 := 999998
	stalePIDContent1 := fmt.Sprintf(`{
  "pid": %d,
  "name": "agent-1",
  "command": "python",
  "args": ["agent_adapter.py"],
  "started_at": "%s",
  "log_file": "%s"
}`, stalePID1, time.Now().Format(time.RFC3339), filepath.Join(env.LogDir, "agent-1.log"))
	env.WritePIDFile("agent-1", stalePIDContent1)

	stalePID2 := 999999
	stalePIDContent2 := fmt.Sprintf(`{
  "pid": %d,
  "name": "agent-2",
  "command": "python",
  "args": ["agent_adapter.py"],
  "started_at": "%s",
  "log_file": "%s"
}`, stalePID2, time.Now().Format(time.RFC3339), filepath.Join(env.LogDir, "agent-2.log"))
	env.WritePIDFile("agent-2", stalePIDContent2)

	// Run down command
	runDown(downCmd, []string{})

	// Verify both PID files were cleaned up
	pidFile1 := filepath.Join(env.PIDDir, "agent-1.json")
	pidFile2 := filepath.Join(env.PIDDir, "agent-2.json")

	if env.FileExists(pidFile1) {
		t.Error("Expected agent-1 PID file to be cleaned up")
	}
	if env.FileExists(pidFile2) {
		t.Error("Expected agent-2 PID file to be cleaned up")
	}
}

// TestDownCommand_HomeDirectoryError tests error handling when home directory cannot be determined
func TestDownCommand_HomeDirectoryError(t *testing.T) {
	// This test is challenging because os.UserHomeDir() is hard to mock
	// We'll skip it as it's an edge case that's difficult to trigger
	t.Skip("Skipping test that requires mocking os.UserHomeDir()")
}

// TestDownCommand_ProcessManagerInitError tests error handling when process manager fails to initialize
func TestDownCommand_ProcessManagerInitError(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)

	// Change to temp directory
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override home directory to a path that will cause permission errors
	// Create a directory with no write permissions
	restrictedDir := filepath.Join(env.TempDir, "restricted")
	if err := os.MkdirAll(restrictedDir, 0000); err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}
	defer os.Chmod(restrictedDir, 0755) // Restore permissions for cleanup

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", restrictedDir)
	defer os.Setenv("HOME", oldHome)

	// Run down command
	var exitCode int
	var exitCalled bool
	exitCode, exitCalled = RunWithExitCapture(func() {
		runDown(downCmd, []string{})
	})

	// Verify exit code is 1 (failure)
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for process manager init error, got %d", exitCode)
	}
}

// TestDownCommand_ListProcessesError tests error handling when listing processes fails
func TestDownCommand_ListProcessesError(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)

	// Change to temp directory
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override home directory for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create an invalid PID file (not JSON)
	invalidPIDFile := filepath.Join(env.PIDDir, "invalid.json")
	if err := os.WriteFile(invalidPIDFile, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid PID file: %v", err)
	}

	// Run down command
	// Note: ListProcesses skips invalid entries, so this should succeed
	runDown(downCmd, []string{})

	// Test passes if no panic/exit occurs
}

// TestDownCommand_GracefulShutdown tests that processes are shut down gracefully
func TestDownCommand_GracefulShutdown(t *testing.T) {
	t.Skip("Skipping test that requires starting real processes - tested in integration test")

	// This test would require starting actual processes and verifying they receive
	// SIGTERM before SIGKILL. This is complex to set up reliably in a unit test
	// and is better tested in integration tests.
}

// TestDownCommand_CleanupOperations tests that cleanup operations are performed
func TestDownCommand_CleanupOperations(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)

	// Change to temp directory
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override home directory for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create multiple PID files
	for i := 1; i <= 3; i++ {
		pidContent := fmt.Sprintf(`{
  "pid": 999%d,
  "name": "agent-%d",
  "command": "python",
  "args": ["agent_adapter.py"],
  "started_at": "%s",
  "log_file": "%s"
}`, i, i, time.Now().Format(time.RFC3339), filepath.Join(env.LogDir, fmt.Sprintf("agent-%d.log", i)))
		env.WritePIDFile(fmt.Sprintf("agent-%d", i), pidContent)
	}

	// Verify PID files exist before cleanup
	for i := 1; i <= 3; i++ {
		pidFile := filepath.Join(env.PIDDir, fmt.Sprintf("agent-%d.json", i))
		if !env.FileExists(pidFile) {
			t.Fatalf("Expected PID file agent-%d.json to exist before cleanup", i)
		}
	}

	// Run down command
	runDown(downCmd, []string{})

	// Verify all PID files were cleaned up
	for i := 1; i <= 3; i++ {
		pidFile := filepath.Join(env.PIDDir, fmt.Sprintf("agent-%d.json", i))
		if env.FileExists(pidFile) {
			t.Errorf("Expected PID file agent-%d.json to be cleaned up", i)
		}
	}
}

// TestDownCommand_MultipleAgents tests down command with multiple agent types
func TestDownCommand_MultipleAgents(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)

	// Change to temp directory
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override home directory for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create PID files for different types of processes
	processes := []struct {
		name    string
		command string
		args    []string
	}{
		{"planner-agent", "python", []string{"agent_adapter.py"}},
		{"coder-agent", "python", []string{"agent_adapter.py"}},
		{"tester-agent", "python", []string{"agent_adapter.py"}},
		{"mcp_agent_mail", "python", []string{"-m", "mcp_agent_mail.server"}},
	}

	for _, proc := range processes {
		pidContent := fmt.Sprintf(`{
  "pid": 999999,
  "name": "%s",
  "command": "%s",
  "args": %s,
  "started_at": "%s",
  "log_file": "%s"
}`, proc.name, proc.command, formatArgsJSON(proc.args), time.Now().Format(time.RFC3339), filepath.Join(env.LogDir, proc.name+".log"))
		env.WritePIDFile(proc.name, pidContent)
	}

	// Run down command
	runDown(downCmd, []string{})

	// Verify all PID files were cleaned up
	for _, proc := range processes {
		pidFile := filepath.Join(env.PIDDir, proc.name+".json")
		if env.FileExists(pidFile) {
			t.Errorf("Expected PID file %s.json to be cleaned up", proc.name)
		}
	}
}

// TestDownCommand_ErrorHandling tests various error scenarios
func TestDownCommand_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*TestEnvironment)
		expectedExit int
		description  string
	}{
		{
			name: "no processes",
			setupFunc: func(env *TestEnvironment) {
				// Don't create any PID files
			},
			expectedExit: -1, // No exit expected
			description:  "Should succeed with no processes",
		},
		{
			name: "stale PID files only",
			setupFunc: func(env *TestEnvironment) {
				stalePIDContent := fmt.Sprintf(`{
  "pid": 999999,
  "name": "stale",
  "command": "python",
  "args": ["agent.py"],
  "started_at": "%s",
  "log_file": "%s"
}`, time.Now().Format(time.RFC3339), filepath.Join(env.LogDir, "stale.log"))
				env.WritePIDFile("stale", stalePIDContent)
			},
			expectedExit: -1, // No exit expected
			description:  "Should clean up stale PID files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test environment
			env := NewTestEnvironment(t)

			// Change to temp directory
			restore := ChangeToTempDir(t, env.TempDir)
			defer restore()

			// Override home directory for this test
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", env.TempDir)
			defer os.Setenv("HOME", oldHome)

			// Run setup
			if tt.setupFunc != nil {
				tt.setupFunc(env)
			}

			// Run command
			if tt.expectedExit == -1 {
				// No exit expected
				runDown(downCmd, []string{})
			} else {
				exitCode, exitCalled := RunWithExitCapture(func() {
					runDown(downCmd, []string{})
				})

				if !exitCalled {
					t.Errorf("%s: Expected os.Exit to be called", tt.description)
				}
				if exitCode != tt.expectedExit {
					t.Errorf("%s: Expected exit code %d, got %d", tt.description, tt.expectedExit, exitCode)
				}
			}
		})
	}
}

// TestDownCommand_PIDFileCleanup tests that PID files are properly cleaned up
func TestDownCommand_PIDFileCleanup(t *testing.T) {
	// Create test environment
	env := NewTestEnvironment(t)

	// Change to temp directory
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Override home directory for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", env.TempDir)
	defer os.Setenv("HOME", oldHome)

	// Create PID files
	pidFiles := []string{"agent-1", "agent-2", "mcp_agent_mail"}
	for _, name := range pidFiles {
		pidContent := fmt.Sprintf(`{
  "pid": 999999,
  "name": "%s",
  "command": "python",
  "args": ["test.py"],
  "started_at": "%s",
  "log_file": "%s"
}`, name, time.Now().Format(time.RFC3339), filepath.Join(env.LogDir, name+".log"))
		env.WritePIDFile(name, pidContent)
	}

	// Verify files exist
	for _, name := range pidFiles {
		pidFile := filepath.Join(env.PIDDir, name+".json")
		if !env.FileExists(pidFile) {
			t.Fatalf("Expected PID file %s.json to exist", name)
		}
	}

	// Run down command
	runDown(downCmd, []string{})

	// Verify all files were cleaned up
	for _, name := range pidFiles {
		pidFile := filepath.Join(env.PIDDir, name+".json")
		if env.FileExists(pidFile) {
			t.Errorf("Expected PID file %s.json to be cleaned up", name)
		}
	}
}

// Helper function to format args as JSON array
func formatArgsJSON(args []string) string {
	if len(args) == 0 {
		return "[]"
	}
	result := "["
	for i, arg := range args {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf(`"%s"`, arg)
	}
	result += "]"
	return result
}
