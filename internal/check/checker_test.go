package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckBinary(t *testing.T) {
	checker := NewChecker("", "")

	tests := []struct {
		name       string
		binary     string
		wantStatus CheckStatus
	}{
		{"existing binary", "ls", CheckPass},
		{"non-existent binary", "nonexistent-binary-xyz", CheckFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.CheckBinary(tt.binary)
			if result.Status != tt.wantStatus {
				t.Errorf("CheckBinary(%s) status = %v, want %v", tt.binary, result.Status, tt.wantStatus)
			}
		})
	}
}

func TestCheckFile(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a test directory
	testDir := filepath.Join(tmpDir, "testdir")
	err = os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	checker := NewChecker("", "")

	tests := []struct {
		name       string
		path       string
		wantStatus CheckStatus
	}{
		{"existing file", testFile, CheckPass},
		{"non-existent file", filepath.Join(tmpDir, "nonexistent.txt"), CheckFail},
		{"directory instead of file", testDir, CheckFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.CheckFile(tt.path)
			if result.Status != tt.wantStatus {
				t.Errorf("CheckFile(%s) status = %v, want %v", tt.path, result.Status, tt.wantStatus)
			}
		})
	}
}

func TestCheckConfig(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		content    string
		wantStatus CheckStatus
	}{
		{
			name: "valid config",
			content: `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test]
command = "python agent.py"
model = "claude"
phases = ["planning"]
`,
			wantStatus: CheckPass,
		},
		{
			name: "missing beads_db_path",
			content: `[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"
`,
			wantStatus: CheckFail,
		},
		{
			name: "invalid TOML",
			content: `[core
beads_db_path = 
`,
			wantStatus: CheckFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, "test-"+tt.name+".toml")
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			checker := NewChecker(configPath, "")
			result := checker.CheckConfig()
			if result.Status != tt.wantStatus {
				t.Errorf("CheckConfig() status = %v, want %v (message: %s)", result.Status, tt.wantStatus, result.Message)
			}
		})
	}
}

func TestCheckConfigMissingFile(t *testing.T) {
	checker := NewChecker("/nonexistent/path/asc.toml", "")
	result := checker.CheckConfig()
	if result.Status != CheckFail {
		t.Errorf("CheckConfig() with missing file should fail, got %v", result.Status)
	}
}

func TestCheckEnv(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		content    string
		keys       []string
		wantStatus CheckStatus
	}{
		{
			name: "all keys present",
			content: `CLAUDE_API_KEY=sk-ant-123
OPENAI_API_KEY=sk-456
GOOGLE_API_KEY=789
`,
			keys:       []string{"CLAUDE_API_KEY", "OPENAI_API_KEY", "GOOGLE_API_KEY"},
			wantStatus: CheckPass,
		},
		{
			name: "missing keys",
			content: `CLAUDE_API_KEY=sk-ant-123
`,
			keys:       []string{"CLAUDE_API_KEY", "OPENAI_API_KEY"},
			wantStatus: CheckWarn,
		},
		{
			name:       "empty file",
			content:    "",
			keys:       []string{"CLAUDE_API_KEY"},
			wantStatus: CheckWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envPath := filepath.Join(tmpDir, "test-"+tt.name+".env")
			err := os.WriteFile(envPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write env file: %v", err)
			}

			checker := NewChecker("", envPath)
			result := checker.CheckEnv(tt.keys)
			if result.Status != tt.wantStatus {
				t.Errorf("CheckEnv() status = %v, want %v (message: %s)", result.Status, tt.wantStatus, result.Message)
			}
		})
	}
}

func TestCheckEnvMissingFile(t *testing.T) {
	checker := NewChecker("", "/nonexistent/.env")
	result := checker.CheckEnv([]string{"CLAUDE_API_KEY"})
	if result.Status != CheckFail {
		t.Errorf("CheckEnv() with missing file should fail, got %v", result.Status)
	}
}

func TestRunAll(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid config
	configPath := filepath.Join(tmpDir, "asc.toml")
	configContent := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create valid env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=test
OPENAI_API_KEY=test
GOOGLE_API_KEY=test
`
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write env: %v", err)
	}

	checker := NewChecker(configPath, envPath)
	results := checker.RunAll()

	if len(results) == 0 {
		t.Errorf("RunAll() returned no results")
	}

	// Verify we have checks for binaries, config, and env
	hasConfigCheck := false
	hasEnvCheck := false
	for _, result := range results {
		if result.Name == "asc.toml" {
			hasConfigCheck = true
		}
		if result.Name == ".env" {
			hasEnvCheck = true
		}
	}

	if !hasConfigCheck {
		t.Errorf("RunAll() missing config check")
	}
	if !hasEnvCheck {
		t.Errorf("RunAll() missing env check")
	}
}

func TestFormatResults(t *testing.T) {
	results := []CheckResult{
		{Name: "git", Status: CheckPass, Message: "Binary found"},
		{Name: "python3", Status: CheckPass, Message: "Binary found"},
		{Name: "bd", Status: CheckFail, Message: "Binary not found"},
		{Name: "docker", Status: CheckWarn, Message: "Docker not found (optional)"},
	}

	output := FormatResults(results)

	if output == "" {
		t.Errorf("FormatResults() returned empty string")
	}

	// Verify output contains key elements
	if !strings.Contains(output, "git") {
		t.Errorf("Output missing 'git'")
	}
	if !strings.Contains(output, "PASS") {
		t.Errorf("Output missing 'PASS'")
	}
	if !strings.Contains(output, "FAIL") {
		t.Errorf("Output missing 'FAIL'")
	}
	if !strings.Contains(output, "WARN") {
		t.Errorf("Output missing 'WARN'")
	}
}

func TestHasFailures(t *testing.T) {
	tests := []struct {
		name    string
		results []CheckResult
		want    bool
	}{
		{
			name: "no failures",
			results: []CheckResult{
				{Status: CheckPass},
				{Status: CheckPass},
			},
			want: false,
		},
		{
			name: "with failures",
			results: []CheckResult{
				{Status: CheckPass},
				{Status: CheckFail},
			},
			want: true,
		},
		{
			name: "only warnings",
			results: []CheckResult{
				{Status: CheckPass},
				{Status: CheckWarn},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasFailures(tt.results)
			if got != tt.want {
				t.Errorf("HasFailures() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckStatus(t *testing.T) {
	tests := []struct {
		name   string
		status CheckStatus
		want   string
	}{
		{"pass", CheckPass, "pass"},
		{"fail", CheckFail, "fail"},
		{"warn", CheckWarn, "warn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("Status = %v, want %v", tt.status, tt.want)
			}
		})
	}
}

func TestCheckResult(t *testing.T) {
	result := CheckResult{
		Name:    "test-component",
		Status:  CheckPass,
		Message: "Test message",
	}

	if result.Name != "test-component" {
		t.Errorf("Name = %v, want test-component", result.Name)
	}
	if result.Status != CheckPass {
		t.Errorf("Status = %v, want %v", result.Status, CheckPass)
	}
	if result.Message != "Test message" {
		t.Errorf("Message = %v, want Test message", result.Message)
	}
}
