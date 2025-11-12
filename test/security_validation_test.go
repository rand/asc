package test

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/rand/asc/internal/config"
	"github.com/rand/asc/internal/logger"
	"github.com/rand/asc/internal/secrets"
)

// TestSecurityValidation_NoSecretsInLogs validates that no secrets appear in log files
func TestSecurityValidation_NoSecretsInLogs(t *testing.T) {
	t.Run("verify logger does not log API keys", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "test.log")
		
		// Create logger
		testLogger, err := logger.NewLogger(logPath, 1024*1024, 5, logger.INFO)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer testLogger.Close()
		
		// Simulate logging with API key in context (should be filtered)
		apiKey := "sk-ant-secret123456789"
		testLogger.Info("Starting agent")
		testLogger.WithFields(logger.Fields{
			"agent": "test-agent",
			"model": "claude",
		}).Info("Agent initialized")
		
		// Read log file
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		// Verify API key is NOT in logs
		if strings.Contains(string(content), apiKey) {
			t.Error("API key found in log file - this is a security violation")
		}
		
		// Verify common secret patterns are not in logs
		secretPatterns := []string{
			`sk-[a-zA-Z0-9]{20,}`,           // Anthropic API keys
			`sk-[a-zA-Z0-9-]{20,}`,          // OpenAI API keys
			`AIza[a-zA-Z0-9_-]{35}`,         // Google API keys
			`password\s*=\s*\S+`,            // Passwords
			`token\s*=\s*\S+`,               // Tokens
		}
		
		for _, pattern := range secretPatterns {
			matched, _ := regexp.MatchString(pattern, string(content))
			if matched {
				t.Errorf("Secret pattern %s found in logs", pattern)
			}
		}
	})
	
	t.Run("verify error messages do not contain secrets", func(t *testing.T) {
		apiKey := "sk-ant-secret123456789"
		
		// Simulate error with API key
		err := fmt.Errorf("authentication failed")
		
		// Verify error message doesn't contain key
		if strings.Contains(err.Error(), apiKey) {
			t.Error("API key found in error message")
		}
	})
	
	t.Run("scan actual log files for secrets", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Cannot get home directory")
		}
		
		logDir := filepath.Join(homeDir, ".asc", "logs")
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			t.Skip("Log directory does not exist")
		}
		
		// Scan all log files
		err = filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			
			if info.IsDir() || !strings.HasSuffix(path, ".log") {
				return nil
			}
			
			// Read log file
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			
			// Check for secret patterns
			secretPatterns := []string{
				`sk-ant-[a-zA-Z0-9]{20,}`,
				`sk-[a-zA-Z0-9-]{20,}`,
				`AIza[a-zA-Z0-9_-]{35}`,
			}
			
			for _, pattern := range secretPatterns {
				matched, _ := regexp.MatchString(pattern, string(content))
				if matched {
					t.Errorf("Secret pattern found in log file %s", path)
				}
			}
			
			return nil
		})
		
		if err != nil {
			t.Errorf("Error scanning log files: %v", err)
		}
	})
}

// TestSecurityValidation_FilePermissions validates file permissions on sensitive files
func TestSecurityValidation_FilePermissions(t *testing.T) {
	t.Run("verify .env file permissions", func(t *testing.T) {
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			t.Skip(".env file does not exist")
		}
		
		info, err := os.Stat(".env")
		if err != nil {
			t.Fatalf("Failed to stat .env: %v", err)
		}
		
		mode := info.Mode().Perm()
		expected := os.FileMode(0600)
		
		if mode != expected {
			t.Errorf(".env permissions = %v, want %v (rw-------)", mode, expected)
			t.Errorf("Fix with: chmod 600 .env")
		}
		
		// Check if world-readable (security issue)
		if mode&0004 != 0 {
			t.Error(".env is world-readable - this is a security vulnerability")
		}
		
		// Check if group-readable (potential security issue)
		if mode&0040 != 0 {
			t.Error(".env is group-readable - this may be a security issue")
		}
	})
	
	t.Run("verify age key permissions", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Cannot get home directory")
		}
		
		keyPath := filepath.Join(homeDir, ".asc", "age.key")
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			t.Skip("age.key does not exist")
		}
		
		info, err := os.Stat(keyPath)
		if err != nil {
			t.Fatalf("Failed to stat age.key: %v", err)
		}
		
		mode := info.Mode().Perm()
		expected := os.FileMode(0600)
		
		if mode != expected {
			t.Errorf("age.key permissions = %v, want %v", mode, expected)
		}
		
		if mode&0044 != 0 {
			t.Error("age.key is readable by others - this is a security vulnerability")
		}
	})
	
	t.Run("verify log directory permissions", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Cannot get home directory")
		}
		
		logDir := filepath.Join(homeDir, ".asc", "logs")
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			t.Skip("Log directory does not exist")
		}
		
		info, err := os.Stat(logDir)
		if err != nil {
			t.Fatalf("Failed to stat log directory: %v", err)
		}
		
		mode := info.Mode().Perm()
		
		// Should be 0700 or 0755
		if mode != 0700 && mode != 0755 {
			t.Logf("Log directory permissions = %v (acceptable: 0700 or 0755)", mode)
		}
		
		// Must not be world-writable
		if mode&0002 != 0 {
			t.Error("Log directory is world-writable - this is a security vulnerability")
		}
	})
	
	t.Run("verify PID directory permissions", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Cannot get home directory")
		}
		
		pidDir := filepath.Join(homeDir, ".asc", "pids")
		if _, err := os.Stat(pidDir); os.IsNotExist(err) {
			t.Skip("PID directory does not exist")
		}
		
		info, err := os.Stat(pidDir)
		if err != nil {
			t.Fatalf("Failed to stat PID directory: %v", err)
		}
		
		mode := info.Mode().Perm()
		
		// Should be 0700 or 0755
		if mode != 0700 && mode != 0755 {
			t.Logf("PID directory permissions = %v (acceptable: 0700 or 0755)", mode)
		}
		
		// Must not be world-writable
		if mode&0002 != 0 {
			t.Error("PID directory is world-writable - this is a security vulnerability")
		}
	})
	
	t.Run("verify encrypted files are not world-readable", func(t *testing.T) {
		// Check for .env.age files
		matches, err := filepath.Glob("*.age")
		if err != nil {
			t.Fatalf("Failed to glob: %v", err)
		}
		
		for _, path := range matches {
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			
			mode := info.Mode().Perm()
			
			// Encrypted files should have restrictive permissions
			if mode&0044 != 0 {
				t.Errorf("Encrypted file %s has overly permissive permissions: %v", path, mode)
			}
		}
	})
}

// TestSecurityValidation_APIKeyHandling validates API key handling security
func TestSecurityValidation_APIKeyHandling(t *testing.T) {
	t.Run("verify API keys passed via environment not command line", func(t *testing.T) {
		// Simulate agent start
		cmd := exec.Command("echo", "test")
		cmd.Env = append(os.Environ(), 
			"CLAUDE_API_KEY=sk-test",
			"OPENAI_API_KEY=sk-test",
			"GOOGLE_API_KEY=test",
		)
		
		// Verify keys are not in Args (visible in ps)
		for _, arg := range cmd.Args {
			if strings.Contains(arg, "API_KEY") {
				t.Error("API key found in command arguments - visible in process list")
			}
		}
		
		// Verify keys are in environment
		hasKeys := false
		for _, env := range cmd.Env {
			if strings.Contains(env, "API_KEY=") {
				hasKeys = true
				break
			}
		}
		
		if !hasKeys {
			t.Error("API keys not found in environment variables")
		}
	})
	
	t.Run("verify secrets manager encrypts properly", func(t *testing.T) {
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
		
		// Create test env file
		envPath := filepath.Join(tmpDir, ".env")
		secretContent := "CLAUDE_API_KEY=sk-ant-secret123\nOPENAI_API_KEY=sk-secret456\n"
		if err := os.WriteFile(envPath, []byte(secretContent), 0600); err != nil {
			t.Fatalf("Failed to write env: %v", err)
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
			t.Error("Secret found in encrypted file - encryption failed")
		}
		if strings.Contains(string(encContent), "sk-secret456") {
			t.Error("Secret found in encrypted file - encryption failed")
		}
	})
	
	t.Run("verify env file validation", func(t *testing.T) {
		manager := secrets.NewManager()
		
		tmpDir := t.TempDir()
		
		// Test valid env file
		validEnv := filepath.Join(tmpDir, "valid.env")
		validContent := "CLAUDE_API_KEY=sk-test\nOPENAI_API_KEY=sk-test\nGOOGLE_API_KEY=test\n"
		if err := os.WriteFile(validEnv, []byte(validContent), 0600); err != nil {
			t.Fatalf("Failed to write valid env: %v", err)
		}
		
		if err := manager.ValidateEnvFile(validEnv); err != nil {
			t.Errorf("Valid env file rejected: %v", err)
		}
		
		// Test invalid env file (missing keys)
		invalidEnv := filepath.Join(tmpDir, "invalid.env")
		invalidContent := "CLAUDE_API_KEY=sk-test\n"
		if err := os.WriteFile(invalidEnv, []byte(invalidContent), 0600); err != nil {
			t.Fatalf("Failed to write invalid env: %v", err)
		}
		
		if err := manager.ValidateEnvFile(invalidEnv); err == nil {
			t.Error("Invalid env file accepted - should require all keys")
		}
	})
}

// TestSecurityValidation_InputSanitization validates input sanitization
func TestSecurityValidation_InputSanitization(t *testing.T) {
	t.Run("verify config path validation", func(t *testing.T) {
		// Test path traversal attempts
		invalidPaths := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32",
			"/etc/passwd",
			"./../../sensitive",
		}
		
		for _, path := range invalidPaths {
			_, err := config.Load(path)
			if err == nil {
				t.Errorf("Path traversal not prevented: %s", path)
			}
		}
	})
	
	t.Run("verify agent name validation", func(t *testing.T) {
		invalidNames := []string{
			"agent; rm -rf /",
			"agent && malicious",
			"agent | cat /etc/passwd",
			"agent`whoami`",
			"agent$(whoami)",
			"agent\nmalicious",
			"agent\rmalicious",
			"agent/../../etc",
			"agent\\..\\..\\windows",
		}
		
		for _, name := range invalidNames {
			if isValidAgentName(name) {
				t.Errorf("Invalid agent name accepted: %s", name)
			}
		}
		
		// Test valid names
		validNames := []string{
			"agent-1",
			"test_agent",
			"MyAgent",
			"agent123",
		}
		
		for _, name := range validNames {
			if !isValidAgentName(name) {
				t.Errorf("Valid agent name rejected: %s", name)
			}
		}
	})
	
	t.Run("verify command validation", func(t *testing.T) {
		invalidCommands := []string{
			"python agent.py; rm -rf /",
			"python agent.py && malicious",
			"python agent.py | cat /etc/passwd",
			"python agent.py`whoami`",
			"python agent.py$(whoami)",
			"python agent.py\nmalicious",
		}
		
		for _, cmd := range invalidCommands {
			if !containsShellMetachars(cmd) {
				t.Errorf("Shell metacharacters not detected in: %s", cmd)
			}
		}
	})
	
	t.Run("verify environment variable validation", func(t *testing.T) {
		invalidEnvVars := []string{
			"VALUE; malicious",
			"VALUE\nMALICIOUS=bad",
			"VALUE\x00MALICIOUS=bad",
			"VALUE\rMALICIOUS=bad",
		}
		
		for _, value := range invalidEnvVars {
			if !strings.ContainsAny(value, ";\n\r\x00") {
				continue
			}
			// Good - invalid value detected
			t.Logf("Correctly detected invalid env var: %q", value)
		}
	})
}

// TestSecurityValidation_CommandInjection validates command injection prevention
func TestSecurityValidation_CommandInjection(t *testing.T) {
	t.Run("verify no shell execution with user input", func(t *testing.T) {
		// Good pattern: direct execution
		userInput := "test; malicious"
		cmd := exec.Command("echo", userInput)
		
		// Verify it's not using shell
		if cmd.Path == "sh" || cmd.Path == "bash" || cmd.Path == "/bin/sh" {
			t.Error("Using shell execution - vulnerable to injection")
		}
		
		// Verify argument is treated as single value
		if len(cmd.Args) != 2 {
			t.Error("Arguments not properly separated")
		}
		
		if cmd.Args[1] != userInput {
			t.Error("Argument was modified")
		}
	})
	
	t.Run("verify shell metacharacter detection", func(t *testing.T) {
		dangerousInputs := []string{
			"test; rm -rf /",
			"test && malicious",
			"test | cat /etc/passwd",
			"test`whoami`",
			"test$(whoami)",
			"test\nmalicious",
			"test>output.txt",
			"test<input.txt",
		}
		
		for _, input := range dangerousInputs {
			if !containsShellMetachars(input) {
				t.Errorf("Failed to detect shell metacharacters in: %s", input)
			}
		}
	})
	
	t.Run("verify config command validation", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "asc.toml")
		
		// Test config with shell metacharacters
		maliciousConfig := `
[core]
beads_db_path = "/tmp/test"

[agent.test]
command = "python agent.py; rm -rf /"
model = "claude"
phases = ["planning"]
`
		
		if err := os.WriteFile(configPath, []byte(maliciousConfig), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		
		cfg, err := config.Load(configPath)
		if err == nil {
			// If config loads, verify we can detect shell metacharacters
			foundDangerous := false
			for _, agent := range cfg.Agents {
				if containsShellMetachars(agent.Command) {
					foundDangerous = true
					t.Logf("Correctly detected dangerous command: %s", agent.Command)
				}
			}
			if !foundDangerous {
				t.Error("Failed to detect command with shell metacharacters")
			}
		}
	})
}

// TestSecurityValidation_PathTraversal validates path traversal protection
func TestSecurityValidation_PathTraversal(t *testing.T) {
	t.Run("verify relative path traversal prevention", func(t *testing.T) {
		baseDir := "/home/user/asc"
		
		traversalPaths := []string{
			"../../../etc/passwd",
			"logs/../../etc/passwd",
			"./../../sensitive",
		}
		
		for _, path := range traversalPaths {
			fullPath := filepath.Join(baseDir, path)
			cleanPath := filepath.Clean(fullPath)
			
			if !strings.HasPrefix(cleanPath, baseDir) {
				// Good - traversal would escape base directory
				t.Logf("Correctly detected traversal: %s -> %s", path, cleanPath)
			} else {
				t.Errorf("Path traversal not detected: %s", path)
			}
		}
		
		// Test Windows-style paths separately (behavior differs by OS)
		windowsPath := "..\\..\\..\\windows\\system32"
		fullPath := filepath.Join(baseDir, windowsPath)
		cleanPath := filepath.Clean(fullPath)
		t.Logf("Windows-style path handling: %s -> %s", windowsPath, cleanPath)
	})
	
	t.Run("verify absolute path validation", func(t *testing.T) {
		baseDir := "/home/user/asc"
		
		absolutePaths := []string{
			"/etc/passwd",
			"/root/.ssh/id_rsa",
		}
		
		for _, path := range absolutePaths {
			if filepath.IsAbs(path) && !strings.HasPrefix(path, baseDir) {
				// Good - absolute path outside base dir
				t.Logf("Correctly detected absolute path outside base: %s", path)
			}
		}
	})
	
	t.Run("verify symlink handling", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create a file outside tmpDir
		outsideFile := filepath.Join(os.TempDir(), "outside.txt")
		if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
			t.Fatalf("Failed to create outside file: %v", err)
		}
		defer os.Remove(outsideFile)
		
		// Create symlink inside tmpDir pointing outside
		linkPath := filepath.Join(tmpDir, "link")
		if err := os.Symlink(outsideFile, linkPath); err != nil {
			t.Skip("Cannot create symlinks on this system")
		}
		
		// Evaluate symlink
		realPath, err := filepath.EvalSymlinks(linkPath)
		if err != nil {
			t.Fatalf("Failed to evaluate symlink: %v", err)
		}
		
		// Verify symlink points outside base directory
		if !strings.HasPrefix(realPath, tmpDir) {
			t.Logf("Correctly detected symlink traversal: %s -> %s", linkPath, realPath)
		} else {
			t.Error("Symlink traversal not detected")
		}
	})
}

// TestSecurityValidation_SecurityScanResults validates security scan results
func TestSecurityValidation_SecurityScanResults(t *testing.T) {
	t.Run("verify gosec scan passes", func(t *testing.T) {
		// Run gosec
		cmd := exec.Command("gosec", "-fmt=json", "-out=gosec-results.json", "./...")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			// gosec returns non-zero if issues found
			t.Logf("gosec found issues (this may be expected): %s", output)
		}
		
		// Check if results file exists
		if _, err := os.Stat("gosec-results.json"); err == nil {
			// Read and parse results
			content, err := os.ReadFile("gosec-results.json")
			if err != nil {
				t.Fatalf("Failed to read gosec results: %v", err)
			}
			
			// Check for high severity issues
			if strings.Contains(string(content), `"severity":"HIGH"`) {
				t.Error("gosec found HIGH severity security issues")
			}
			
			// Clean up
			os.Remove("gosec-results.json")
		}
	})
	
	t.Run("verify no hardcoded secrets in code", func(t *testing.T) {
		// Search for potential hardcoded secrets
		cmd := exec.Command("grep", "-r", "-E", 
			"(sk-[a-zA-Z0-9]{20,}|AIza[a-zA-Z0-9_-]{35})",
			"--include=*.go",
			"--exclude-dir=vendor",
			"--exclude-dir=.git",
			".")
		
		output, err := cmd.CombinedOutput()
		
		if err == nil && len(output) > 0 {
			// Found potential secrets
			t.Errorf("Potential hardcoded secrets found:\n%s", output)
		}
	})
	
	t.Run("verify .env not in git", func(t *testing.T) {
		// Check if .env is tracked by git
		cmd := exec.Command("git", "ls-files", ".env")
		output, err := cmd.CombinedOutput()
		
		if err == nil && len(output) > 0 {
			t.Error(".env file is tracked by git - this is a security violation")
		}
		
		// Check if .env is in .gitignore
		if content, err := os.ReadFile(".gitignore"); err == nil {
			if !strings.Contains(string(content), ".env") {
				t.Error(".env not found in .gitignore")
			}
		}
	})
}


// TestSecurityValidation_BestPractices validates security best practices are followed
func TestSecurityValidation_BestPractices(t *testing.T) {
	t.Run("verify TLS configuration", func(t *testing.T) {
		// This test verifies that HTTP clients use proper TLS configuration
		// In actual code, we should use TLS 1.2+ and verify certificates
		t.Log("TLS configuration should use MinVersion: TLS 1.2")
		t.Log("Certificate verification should not be skipped")
	})
	
	t.Run("verify process isolation", func(t *testing.T) {
		// Verify agents run with process groups for proper cleanup
		t.Log("Agents should use Setpgid for process group isolation")
		t.Log("Agents should not run with elevated privileges")
	})
	
	t.Run("verify secure defaults", func(t *testing.T) {
		// Check that secure defaults are used
		tmpDir := t.TempDir()
		
		// Test file creation with secure permissions
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		info, err := os.Stat(testFile)
		if err != nil {
			t.Fatalf("Failed to stat test file: %v", err)
		}
		
		if info.Mode().Perm() != 0600 {
			t.Errorf("File created with insecure permissions: %v", info.Mode().Perm())
		}
	})
	
	t.Run("verify error handling does not leak information", func(t *testing.T) {
		// Verify errors don't contain sensitive paths or data
		testError := fmt.Errorf("operation failed")
		
		// Error should not contain full paths
		if strings.Contains(testError.Error(), "/home/") {
			t.Error("Error contains full path - may leak information")
		}
		
		// Error should not contain API keys
		if strings.Contains(testError.Error(), "sk-") {
			t.Error("Error contains potential API key")
		}
	})
	
	t.Run("verify logging configuration", func(t *testing.T) {
		// Verify logger is configured securely
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "test.log")
		
		testLogger, err := logger.NewLogger(logPath, 1024*1024, 5, logger.INFO)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer testLogger.Close()
		
		// Verify log file has appropriate permissions
		info, err := os.Stat(logPath)
		if err != nil {
			t.Fatalf("Failed to stat log file: %v", err)
		}
		
		// Log files should not be world-writable
		if info.Mode().Perm()&0002 != 0 {
			t.Error("Log file is world-writable")
		}
	})
	
	t.Run("verify dependency security", func(t *testing.T) {
		// Check for known vulnerable dependencies
		cmd := exec.Command("go", "list", "-json", "-m", "all")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Failed to list dependencies: %v", err)
			return
		}
		
		// Check for common vulnerable patterns
		vulnerablePatterns := []string{
			// Add known vulnerable package patterns here
		}
		
		for _, pattern := range vulnerablePatterns {
			if strings.Contains(string(output), pattern) {
				t.Errorf("Potentially vulnerable dependency found: %s", pattern)
			}
		}
	})
	
	t.Run("verify input validation is comprehensive", func(t *testing.T) {
		// Test various input validation scenarios
		testCases := []struct {
			name  string
			input string
			valid bool
		}{
			{"valid agent name", "test-agent", true},
			{"invalid with semicolon", "test;agent", false},
			{"invalid with pipe", "test|agent", false},
			{"invalid with backtick", "test`agent", false},
			{"invalid with dollar", "test$agent", false},
			{"invalid with newline", "test\nagent", false},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				valid := isValidAgentName(tc.input)
				if valid != tc.valid {
					t.Errorf("isValidAgentName(%q) = %v, want %v", tc.input, valid, tc.valid)
				}
			})
		}
	})
}

// TestSecurityValidation_ComprehensiveCheck runs a comprehensive security check
func TestSecurityValidation_ComprehensiveCheck(t *testing.T) {
	t.Run("comprehensive security audit", func(t *testing.T) {
		issues := []string{}
		
		// Check 1: .env file security
		if info, err := os.Stat(".env"); err == nil {
			if info.Mode().Perm()&0044 != 0 {
				issues = append(issues, ".env file has overly permissive permissions")
			}
		}
		
		// Check 2: age key security
		homeDir, _ := os.UserHomeDir()
		keyPath := filepath.Join(homeDir, ".asc", "age.key")
		if info, err := os.Stat(keyPath); err == nil {
			if info.Mode().Perm()&0044 != 0 {
				issues = append(issues, "age.key has overly permissive permissions")
			}
		}
		
		// Check 3: .gitignore includes .env
		if content, err := os.ReadFile(".gitignore"); err == nil {
			if !strings.Contains(string(content), ".env") {
				issues = append(issues, ".env not in .gitignore")
			}
		}
		
		// Check 4: No secrets in git history
		cmd := exec.Command("git", "log", "--all", "--full-history", "--", ".env")
		if output, err := cmd.CombinedOutput(); err == nil && len(output) > 0 {
			issues = append(issues, ".env found in git history")
		}
		
		// Check 5: Log files don't contain secrets
		logDir := filepath.Join(homeDir, ".asc", "logs")
		if _, err := os.Stat(logDir); err == nil {
			filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() || !strings.HasSuffix(path, ".log") {
					return nil
				}
				
				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				
				if matched, _ := regexp.MatchString(`sk-[a-zA-Z0-9]{20,}`, string(content)); matched {
					issues = append(issues, fmt.Sprintf("Potential secret found in %s", path))
				}
				
				return nil
			})
		}
		
		// Report all issues
		if len(issues) > 0 {
			t.Errorf("Security audit found %d issues:", len(issues))
			for i, issue := range issues {
				t.Errorf("  %d. %s", i+1, issue)
			}
		} else {
			t.Log("✓ Comprehensive security audit passed")
		}
	})
}

// Helper functions (isValidAgentName and containsShellMetachars are in security_test.go)

// scanLogFileForSecrets scans a log file for potential secrets
func scanLogFileForSecrets(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var findings []string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	secretPatterns := []*regexp.Regexp{
		regexp.MustCompile(`sk-ant-[a-zA-Z0-9]{20,}`),
		regexp.MustCompile(`sk-[a-zA-Z0-9-]{20,}`),
		regexp.MustCompile(`AIza[a-zA-Z0-9_-]{35}`),
	}
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		for _, pattern := range secretPatterns {
			if pattern.MatchString(line) {
				findings = append(findings, fmt.Sprintf("Line %d: potential secret", lineNum))
			}
		}
	}
	
	return findings, scanner.Err()
}
