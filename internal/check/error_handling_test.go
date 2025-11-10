package check

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCheckBinary_ErrorPaths tests error handling in binary checks
func TestCheckBinary_ErrorPaths(t *testing.T) {
	checker := NewChecker("asc.toml", ".env")

	tests := []struct {
		name         string
		binary       string
		expectPass   bool
		expectStatus CheckStatus
	}{
		{
			name:         "nonexistent binary",
			binary:       "nonexistent-binary-12345",
			expectPass:   false,
			expectStatus: CheckFail,
		},
		{
			name:         "empty binary name",
			binary:       "",
			expectPass:   false,
			expectStatus: CheckFail,
		},
		{
			name:         "binary with path traversal",
			binary:       "../../../bin/sh",
			expectPass:   false,
			expectStatus: CheckFail,
		},
		{
			name:         "binary with special characters",
			binary:       "test; rm -rf /",
			expectPass:   false,
			expectStatus: CheckFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.CheckBinary(tt.binary)

			if result.Status != tt.expectStatus {
				t.Errorf("Expected status %s, got %s", tt.expectStatus, result.Status)
			}

			if tt.expectPass && result.Status != CheckPass {
				t.Errorf("Expected check to pass, got status: %s, message: %s", result.Status, result.Message)
			}

			if !tt.expectPass && result.Status == CheckPass {
				t.Error("Expected check to fail, but it passed")
			}
		})
	}
}

// TestCheckFile_ErrorPaths tests error handling in file checks
func TestCheckFile_ErrorPaths(t *testing.T) {
	checker := NewChecker("asc.toml", ".env")

	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) string
		expectStatus CheckStatus
		errorMsg     string
	}{
		{
			name: "nonexistent file",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.txt")
			},
			expectStatus: CheckFail,
			errorMsg:     "not found",
		},
		{
			name: "unreadable file",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "unreadable.txt")
				if err := os.WriteFile(path, []byte("test"), 0000); err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectStatus: CheckFail,
			errorMsg:     "permission",
		},
		{
			name: "directory instead of file",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				subdir := filepath.Join(dir, "subdir")
				if err := os.MkdirAll(subdir, 0755); err != nil {
					t.Fatal(err)
				}
				return subdir
			},
			expectStatus: CheckWarn,
			errorMsg:     "directory",
		},
		{
			name: "empty path",
			setupFunc: func(t *testing.T) string {
				return ""
			},
			expectStatus: CheckFail,
			errorMsg:     "empty",
		},
		{
			name: "path with null bytes",
			setupFunc: func(t *testing.T) string {
				return "test\x00file"
			},
			expectStatus: CheckFail,
			errorMsg:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFunc(t)
			result := checker.CheckFile(path)

			if result.Status != tt.expectStatus {
				t.Errorf("Expected status %s, got %s", tt.expectStatus, result.Status)
			}

			if tt.errorMsg != "" && !contains(result.Message, tt.errorMsg) {
				t.Errorf("Expected message to contain %q, got: %s", tt.errorMsg, result.Message)
			}
		})
	}
}

// TestCheckConfig_ErrorPaths tests error handling in config checks
func TestCheckConfig_ErrorPaths(t *testing.T) {
	// Note: checker is created per test case with specific config path
	_ = NewChecker("asc.toml", ".env") // Placeholder to show API

	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) string
		expectStatus CheckStatus
		errorMsg     string
	}{
		{
			name: "missing config file",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.toml")
			},
			expectStatus: CheckFail,
			errorMsg:     "not found",
		},
		{
			name: "invalid TOML syntax",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "invalid.toml")
				content := `[core
beads_db_path = "invalid`
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectStatus: CheckFail,
			errorMsg:     "parse",
		},
		{
			name: "empty config file",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "empty.toml")
				if err := os.WriteFile(path, []byte(""), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectStatus: CheckFail,
			errorMsg:     "",
		},
		{
			name: "config with missing required fields",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "incomplete.toml")
				content := `[services.mcp_agent_mail]
url = "http://localhost:8765"`
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectStatus: CheckFail,
			errorMsg:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupFunc(t)
			// Create checker with the test config path
			testChecker := NewChecker(configPath, ".env")
			result := testChecker.CheckConfig()

			if result.Status != tt.expectStatus {
				t.Errorf("Expected status %s, got %s (message: %s)", tt.expectStatus, result.Status, result.Message)
			}

			if tt.errorMsg != "" && !contains(result.Message, tt.errorMsg) {
				t.Errorf("Expected message to contain %q, got: %s", tt.errorMsg, result.Message)
			}
		})
	}
}

// TestCheckEnv_ErrorPaths tests error handling in environment checks
func TestCheckEnv_ErrorPaths(t *testing.T) {
	// Note: checker is created per test case with specific env path
	_ = NewChecker("asc.toml", ".env") // Placeholder to show API

	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) string
		requiredKeys []string
		expectStatus CheckStatus
		errorMsg     string
	}{
		{
			name: "missing env file",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.env")
			},
			requiredKeys: []string{"TEST_KEY"},
			expectStatus: CheckFail,
			errorMsg:     "not found",
		},
		{
			name: "missing required keys",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "test.env")
				content := `PRESENT_KEY=value`
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			requiredKeys: []string{"MISSING_KEY"},
			expectStatus: CheckFail,
			errorMsg:     "MISSING_KEY",
		},
		{
			name: "empty required keys list",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "test.env")
				if err := os.WriteFile(path, []byte("KEY=value"), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			requiredKeys: []string{},
			expectStatus: CheckPass,
			errorMsg:     "",
		},
		{
			name: "unreadable env file",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "unreadable.env")
				if err := os.WriteFile(path, []byte("KEY=value"), 0000); err != nil {
					t.Fatal(err)
				}
				return path
			},
			requiredKeys: []string{"KEY"},
			expectStatus: CheckFail,
			errorMsg:     "permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envPath := tt.setupFunc(t)
			// Create checker with the test env path
			testChecker := NewChecker("asc.toml", envPath)
			result := testChecker.CheckEnv(tt.requiredKeys)

			if result.Status != tt.expectStatus {
				t.Errorf("Expected status %s, got %s (message: %s)", tt.expectStatus, result.Status, result.Message)
			}

			if tt.errorMsg != "" && !contains(result.Message, tt.errorMsg) {
				t.Errorf("Expected message to contain %q, got: %s", tt.errorMsg, result.Message)
			}
		})
	}
}

// TestRunAll_ErrorPaths tests error handling when running all checks
func TestRunAll_ErrorPaths(t *testing.T) {
	// Create a temporary directory with invalid config
	dir := t.TempDir()
	configPath := filepath.Join(dir, "invalid.toml")
	envPath := filepath.Join(dir, ".env")

	// Write invalid config
	if err := os.WriteFile(configPath, []byte("[invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	// Write valid env
	if err := os.WriteFile(envPath, []byte("KEY=value"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create checker with test paths
	checker := NewChecker(configPath, envPath)
	results := checker.RunAll()

	// Should have multiple results
	if len(results) == 0 {
		t.Error("Expected multiple check results")
	}

	// At least one should fail
	hasFailure := false
	for _, result := range results {
		if result.Status == CheckFail {
			hasFailure = true
			break
		}
	}

	if !hasFailure {
		t.Error("Expected at least one check to fail")
	}
}

// TestConcurrentChecks tests error handling under concurrent access
func TestConcurrentChecks(t *testing.T) {
	checker := NewChecker("asc.toml", ".env")

	// Create test file
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run checks concurrently
	done := make(chan CheckResult, 10)
	for i := 0; i < 10; i++ {
		go func() {
			done <- checker.CheckFile(testFile)
		}()
	}

	// Collect results
	for i := 0; i < 10; i++ {
		result := <-done
		if result.Status != CheckPass {
			t.Errorf("Concurrent check failed: %s", result.Message)
		}
	}
}

// TestInvalidInput tests handling of invalid input
func TestInvalidInput(t *testing.T) {
	checker := NewChecker("asc.toml", ".env")

	tests := []struct {
		name string
		fn   func() CheckResult
	}{
		{
			name: "binary with null bytes",
			fn: func() CheckResult {
				return checker.CheckBinary("test\x00binary")
			},
		},
		{
			name: "file with null bytes",
			fn: func() CheckResult {
				return checker.CheckFile("test\x00file")
			},
		},
		{
			name: "config with null bytes",
			fn: func() CheckResult {
				testChecker := NewChecker("test\x00config", ".env")
				return testChecker.CheckConfig()
			},
		},
		{
			name: "extremely long path",
			fn: func() CheckResult {
				longPath := string(make([]byte, 10000))
				return checker.CheckFile(longPath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic occurred: %v", r)
				}
			}()

			result := tt.fn()
			// Should not panic, should return fail status
			if result.Status == CheckPass {
				t.Error("Expected check to fail for invalid input")
			}
		})
	}
}

// TestErrorMessageClarity tests that error messages are clear and actionable
func TestErrorMessageClarity(t *testing.T) {
	checker := NewChecker("asc.toml", ".env")

	tests := []struct {
		name          string
		checkFunc     func() CheckResult
		expectKeyword string
	}{
		{
			name: "missing binary",
			checkFunc: func() CheckResult {
				return checker.CheckBinary("nonexistent-binary-xyz")
			},
			expectKeyword: "not found",
		},
		{
			name: "missing file",
			checkFunc: func() CheckResult {
				return checker.CheckFile("/nonexistent/path/file.txt")
			},
			expectKeyword: "not found",
		},
		{
			name: "invalid config",
			checkFunc: func() CheckResult {
				dir := t.TempDir()
				path := filepath.Join(dir, "bad.toml")
				os.WriteFile(path, []byte("[invalid"), 0644)
				testChecker := NewChecker(path, ".env")
				return testChecker.CheckConfig()
			},
			expectKeyword: "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checkFunc()

			if result.Status == CheckPass {
				t.Error("Expected check to fail")
			}

			if result.Message == "" {
				t.Error("Expected error message to be non-empty")
			}

			if !contains(result.Message, tt.expectKeyword) {
				t.Errorf("Expected message to contain %q, got: %s", tt.expectKeyword, result.Message)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
