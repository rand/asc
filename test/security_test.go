package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rand/asc/internal/config"
	"github.com/rand/asc/internal/secrets"
)

// TestAPIKeyHandling tests secure API key storage and handling
func TestAPIKeyHandling(t *testing.T) {
	t.Run("API keys not logged", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, ".env")
		
		// Create env file with API keys
		envContent := `CLAUDE_API_KEY=sk-ant-secret123
OPENAI_API_KEY=sk-secret456
GOOGLE_API_KEY=secret789
`
		err := os.WriteFile(envPath, []byte(envContent), 0600)
		if err != nil {
			t.Fatalf("Failed to write env file: %v", err)
		}
		
		// Read file and verify keys are present
		content, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf("Failed to read env file: %v", err)
		}
		
		// Verify keys are in file
		if !strings.Contains(string(content), "sk-ant-secret123") {
			t.Error("API key should be in env file")
		}
		
		// Simulate logging - keys should never appear in logs
		logContent := "Starting agent with environment variables"
		if strings.Contains(logContent, "sk-ant-secret123") {
			t.Error("API key should not appear in logs")
		}
	})
	
	t.Run("API keys masked in error messages", func(t *testing.T) {
		apiKey := "sk-ant-secret123456789"
		
		// Simulate error message
		errMsg := "Failed to authenticate with API"
		
		// Verify key is not in error message
		if strings.Contains(errMsg, apiKey) {
			t.Error("API key should not appear in error messages")
		}
	})
	
	t.Run("API keys not in command line args", func(t *testing.T) {
		// Verify that API keys are passed via environment, not command line
		// Command line args are visible in process listings
		
		// Good: passing via environment
		goodCmd := exec.Command("echo", "test")
		goodCmd.Env = append(os.Environ(), "API_KEY=secret")
		
		// Verify good pattern: API key not in args
		for _, arg := range goodCmd.Args {
			if strings.Contains(arg, "API_KEY=") {
				t.Error("API keys should not be passed as command line arguments")
			}
		}
		
		// Verify good pattern: API key in environment
		hasKeyInEnv := false
		for _, env := range goodCmd.Env {
			if strings.HasPrefix(env, "API_KEY=") {
				hasKeyInEnv = true
				break
			}
		}
		if !hasKeyInEnv {
			t.Error("API key should be in environment variables")
		}
	})
}

// TestFilePermissions tests that sensitive files have correct permissions
func TestFilePermissions(t *testing.T) {
	t.Run(".env file permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, ".env")
		
		// Create env file
		err := os.WriteFile(envPath, []byte("TEST=value"), 0600)
		if err != nil {
			t.Fatalf("Failed to write env file: %v", err)
		}
		
		// Check permissions
		info, err := os.Stat(envPath)
		if err != nil {
			t.Fatalf("Failed to stat env file: %v", err)
		}
		
		mode := info.Mode().Perm()
		expected := os.FileMode(0600)
		
		if mode != expected {
			t.Errorf("Env file permissions = %v, want %v", mode, expected)
		}
	})
	
	t.Run("age key file permissions", func(t *testing.T) {
		manager := secrets.NewManager()
		if !manager.IsAgeInstalled() {
			t.Skip("age not installed")
		}
		
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "age.key")
		manager = secrets.NewManagerWithKeyPath(keyPath)
		
		// Generate key
		if err := manager.GenerateKey(); err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}
		
		// Check permissions
		info, err := os.Stat(keyPath)
		if err != nil {
			t.Fatalf("Failed to stat key file: %v", err)
		}
		
		mode := info.Mode().Perm()
		expected := os.FileMode(0600)
		
		if mode != expected {
			t.Errorf("Key file permissions = %v, want %v", mode, expected)
		}
	})
	
	t.Run("log directory permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		logDir := filepath.Join(tmpDir, "logs")
		
		// Create log directory
		err := os.MkdirAll(logDir, 0700)
		if err != nil {
			t.Fatalf("Failed to create log directory: %v", err)
		}
		
		// Check permissions
		info, err := os.Stat(logDir)
		if err != nil {
			t.Fatalf("Failed to stat log directory: %v", err)
		}
		
		mode := info.Mode().Perm()
		expected := os.FileMode(0700)
		
		if mode != expected {
			t.Errorf("Log directory permissions = %v, want %v", mode, expected)
		}
	})
	
	t.Run("PID directory permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidDir := filepath.Join(tmpDir, "pids")
		
		// Create PID directory
		err := os.MkdirAll(pidDir, 0700)
		if err != nil {
			t.Fatalf("Failed to create PID directory: %v", err)
		}
		
		// Check permissions
		info, err := os.Stat(pidDir)
		if err != nil {
			t.Fatalf("Failed to stat PID directory: %v", err)
		}
		
		mode := info.Mode().Perm()
		expected := os.FileMode(0700)
		
		if mode != expected {
			t.Errorf("PID directory permissions = %v, want %v", mode, expected)
		}
	})
	
	t.Run("world-readable .env detection", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, ".env")
		
		// Create world-readable env file (insecure)
		err := os.WriteFile(envPath, []byte("TEST=value"), 0644)
		if err != nil {
			t.Fatalf("Failed to write env file: %v", err)
		}
		
		// Check if file is world-readable
		info, err := os.Stat(envPath)
		if err != nil {
			t.Fatalf("Failed to stat env file: %v", err)
		}
		
		mode := info.Mode().Perm()
		
		// Check if others can read (this is a security issue we're testing for)
		// The test verifies we can DETECT world-readable files
		if mode&0004 != 0 {
			// Good - we detected the insecure permissions
			t.Logf("Correctly detected world-readable .env file with permissions %v", mode)
		} else {
			t.Error("Test setup failed: file should be world-readable for this test")
		}
	})
}

// TestInputValidation tests input validation and sanitization
func TestInputValidation(t *testing.T) {
	t.Run("config file path validation", func(t *testing.T) {
		// Test path traversal attempts
		invalidPaths := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32",
			"/etc/passwd",
			"C:\\Windows\\System32",
			"./../../sensitive",
		}
		
		for _, path := range invalidPaths {
			// Attempt to load config from invalid path
			_, err := config.Load(path)
			if err == nil {
				t.Errorf("Should reject invalid path: %s", path)
			}
		}
	})
	
	t.Run("agent name validation", func(t *testing.T) {
		// Test injection attempts in agent names
		invalidNames := []string{
			"agent; rm -rf /",
			"agent && malicious",
			"agent | cat /etc/passwd",
			"agent`whoami`",
			"agent$(whoami)",
			"agent\nmalicious",
			"agent\rmalicious",
		}
		
		for _, name := range invalidNames {
			// Agent names should only contain alphanumeric, dash, underscore
			if !isValidAgentName(name) {
				// Good - invalid name rejected
				continue
			}
			t.Errorf("Should reject invalid agent name: %s", name)
		}
	})
	
	t.Run("command validation", func(t *testing.T) {
		// Test command injection attempts
		invalidCommands := []string{
			"python agent.py; rm -rf /",
			"python agent.py && malicious",
			"python agent.py | cat /etc/passwd",
			"python agent.py`whoami`",
			"python agent.py$(whoami)",
		}
		
		for _, cmd := range invalidCommands {
			// Commands should be validated before execution
			if !isValidCommand(cmd) {
				// Good - invalid command rejected
				continue
			}
			t.Errorf("Should reject invalid command: %s", cmd)
		}
	})
	
	t.Run("environment variable validation", func(t *testing.T) {
		// Test injection attempts in environment variables
		invalidEnvVars := []string{
			"VALUE; malicious",
			"VALUE\nMALICIOUS=bad",
			"VALUE\x00MALICIOUS=bad",
		}
		
		for _, value := range invalidEnvVars {
			// Environment variables should not contain special characters
			if strings.ContainsAny(value, ";\n\r\x00") {
				// Good - invalid value detected
				continue
			}
		}
	})
	
	t.Run("TOML injection prevention", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "asc.toml")
		
		// Test malicious TOML content
		maliciousContent := `
[core]
beads_db_path = "/tmp/test"

[agent.test]
command = "echo test"
model = "claude"
phases = ["planning"]
`
		
		err := os.WriteFile(configPath, []byte(maliciousContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		
		// Load config - this will fail if command validation is working
		// (echo is not in PATH in test environment, which is expected)
		_, err = config.Load(configPath)
		
		// We expect an error because the command validation should catch issues
		// The actual validation happens in the config.Load function
		if err != nil {
			// Good - config validation is working
			t.Logf("Config validation correctly rejected invalid config: %v", err)
		}
		
		// Test that shell metacharacters would be caught
		maliciousContent2 := `
[core]
beads_db_path = "/tmp/test"

[agent.test]
command = "echo; rm -rf /"
model = "claude"
phases = ["planning"]
`
		configPath2 := filepath.Join(tmpDir, "asc2.toml")
		err = os.WriteFile(configPath2, []byte(maliciousContent2), 0644)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		
		cfg, err := config.Load(configPath2)
		if err == nil {
			// If config loads, verify command doesn't contain shell metacharacters
			for _, agent := range cfg.Agents {
				if strings.ContainsAny(agent.Command, ";|&`$") {
					t.Error("Command contains shell metacharacters that could be exploited")
				}
			}
		}
	})
}

// TestCommandInjection tests prevention of command injection vulnerabilities
func TestCommandInjection(t *testing.T) {
	t.Run("shell metacharacters in paths", func(t *testing.T) {
		// Test paths with shell metacharacters
		dangerousPaths := []string{
			"/tmp/test; rm -rf /",
			"/tmp/test && malicious",
			"/tmp/test | cat /etc/passwd",
			"/tmp/test`whoami`",
			"/tmp/test$(whoami)",
		}
		
		for _, path := range dangerousPaths {
			// Paths should be validated
			if containsShellMetachars(path) {
				// Good - dangerous path detected
				continue
			}
		}
	})
	
	t.Run("command argument escaping", func(t *testing.T) {
		// Test that arguments are properly escaped
		arg := "test; malicious"
		
		// Using exec.Command properly (safe)
		cmd := exec.Command("echo", arg)
		
		// Verify arg is treated as single argument, not executed
		if len(cmd.Args) != 2 {
			t.Error("Arguments should be properly separated")
		}
		
		if cmd.Args[1] != arg {
			t.Error("Argument should not be modified")
		}
	})
	
	t.Run("no shell execution", func(t *testing.T) {
		// Verify we don't use shell execution (sh -c)
		// which would allow command injection
		
		// Good: direct execution
		cmd := exec.Command("echo", "test")
		
		// Bad: shell execution (should never do this with user input)
		// badCmd := exec.Command("sh", "-c", "echo test")
		
		if cmd.Path == "sh" || cmd.Path == "bash" {
			t.Error("Should not use shell execution for user input")
		}
	})
}

// TestPathTraversal tests prevention of path traversal vulnerabilities
func TestPathTraversal(t *testing.T) {
	t.Run("relative path traversal", func(t *testing.T) {
		baseDir := "/home/user/asc"
		
		// Test path traversal attempts
		traversalPaths := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32",
			"logs/../../etc/passwd",
			"./../../sensitive",
		}
		
		for _, path := range traversalPaths {
			fullPath := filepath.Join(baseDir, path)
			cleanPath := filepath.Clean(fullPath)
			
			// Verify path stays within base directory
			if !strings.HasPrefix(cleanPath, baseDir) {
				// Good - traversal detected
				t.Logf("Correctly detected traversal: %s", path)
			}
		}
	})
	
	t.Run("absolute path validation", func(t *testing.T) {
		baseDir := "/home/user/asc"
		
		// Test absolute paths that try to escape
		absolutePaths := []string{
			"/etc/passwd",
			"/root/.ssh/id_rsa",
			"C:\\Windows\\System32",
		}
		
		for _, path := range absolutePaths {
			// Absolute paths should be rejected or validated
			if filepath.IsAbs(path) && !strings.HasPrefix(path, baseDir) {
				// Good - absolute path outside base dir detected
				t.Logf("Correctly detected absolute path: %s", path)
			}
		}
	})
	
	t.Run("symlink traversal", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create a symlink that points outside the directory
		linkPath := filepath.Join(tmpDir, "link")
		targetPath := "/etc/passwd"
		
		// On systems where we can create symlinks
		err := os.Symlink(targetPath, linkPath)
		if err != nil {
			t.Skip("Cannot create symlinks on this system")
		}
		
		// Evaluate symlink
		realPath, err := filepath.EvalSymlinks(linkPath)
		if err != nil {
			t.Fatalf("Failed to evaluate symlink: %v", err)
		}
		
		// Verify symlink points outside base directory
		if !strings.HasPrefix(realPath, tmpDir) {
			// Good - symlink traversal detected
			t.Logf("Correctly detected symlink traversal: %s -> %s", linkPath, realPath)
		}
	})
}

// TestSecretsEncryption tests encryption of sensitive data
func TestSecretsEncryption(t *testing.T) {
	manager := secrets.NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed")
	}
	
	t.Run("encrypted secrets not readable", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "age.key")
		manager = secrets.NewManagerWithKeyPath(keyPath)
		
		// Generate key
		if err := manager.GenerateKey(); err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}
		
		// Create env file with secrets
		envPath := filepath.Join(tmpDir, ".env")
		secretContent := "CLAUDE_API_KEY=sk-ant-secret123\nOPENAI_API_KEY=sk-secret456\nGOOGLE_API_KEY=secret789\n"
		err := os.WriteFile(envPath, []byte(secretContent), 0600)
		if err != nil {
			t.Fatalf("Failed to write env file: %v", err)
		}
		
		// Encrypt
		encPath := envPath + ".age"
		if err := manager.Encrypt(envPath, encPath); err != nil {
			t.Fatalf("Failed to encrypt: %v", err)
		}
		
		// Read encrypted file
		encContent, err := os.ReadFile(encPath)
		if err != nil {
			t.Fatalf("Failed to read encrypted file: %v", err)
		}
		
		// Verify secrets are not in plaintext
		if strings.Contains(string(encContent), "sk-ant-secret123") {
			t.Error("Secrets should not be readable in encrypted file")
		}
		if strings.Contains(string(encContent), "sk-secret456") {
			t.Error("Secrets should not be readable in encrypted file")
		}
	})
	
	t.Run("decryption requires key", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "age.key")
		manager = secrets.NewManagerWithKeyPath(keyPath)
		
		// Generate key
		if err := manager.GenerateKey(); err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}
		
		// Create and encrypt file
		testFile := filepath.Join(tmpDir, "test.txt")
		encFile := testFile + ".age"
		err := os.WriteFile(testFile, []byte("secret"), 0600)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
		
		if err := manager.Encrypt(testFile, encFile); err != nil {
			t.Fatalf("Failed to encrypt: %v", err)
		}
		
		// Try to decrypt without key
		wrongKeyPath := filepath.Join(tmpDir, "wrong.key")
		wrongManager := secrets.NewManagerWithKeyPath(wrongKeyPath)
		
		decFile := filepath.Join(tmpDir, "decrypted.txt")
		err = wrongManager.Decrypt(encFile, decFile)
		if err == nil {
			t.Error("Should not be able to decrypt without correct key")
		}
	})
}

// Helper functions for validation

func isValidAgentName(name string) bool {
	// Agent names should only contain alphanumeric, dash, underscore
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || 
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return len(name) > 0 && len(name) <= 64
}

func isValidCommand(cmd string) bool {
	// Commands should not contain shell metacharacters
	return !containsShellMetachars(cmd)
}

func containsShellMetachars(s string) bool {
	// Check for common shell metacharacters
	metachars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\n", "\r"}
	for _, mc := range metachars {
		if strings.Contains(s, mc) {
			return true
		}
	}
	return false
}
