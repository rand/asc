package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	expected := "asc.toml"
	
	if path != expected {
		t.Errorf("DefaultConfigPath() = %q, want %q", path, expected)
	}
}

func TestDefaultEnvPath(t *testing.T) {
	path := DefaultEnvPath()
	expected := ".env"
	
	if path != expected {
		t.Errorf("DefaultEnvPath() = %q, want %q", path, expected)
	}
}

func TestGetDefaultPIDDir(t *testing.T) {
	dir, err := GetDefaultPIDDir()
	if err != nil {
		t.Fatalf("GetDefaultPIDDir() error = %v", err)
	}
	
	// Should contain .asc/pids
	if !contains(dir, ".asc/pids") {
		t.Errorf("GetDefaultPIDDir() = %q, should contain '.asc/pids'", dir)
	}
	
	// Should be an absolute path
	if !filepath.IsAbs(dir) {
		t.Errorf("GetDefaultPIDDir() = %q, should be absolute path", dir)
	}
}

func TestGetDefaultLogDir(t *testing.T) {
	dir, err := GetDefaultLogDir()
	if err != nil {
		t.Fatalf("GetDefaultLogDir() error = %v", err)
	}
	
	// Should contain .asc/logs
	if !contains(dir, ".asc/logs") {
		t.Errorf("GetDefaultLogDir() = %q, should contain '.asc/logs'", dir)
	}
	
	// Should be an absolute path
	if !filepath.IsAbs(dir) {
		t.Errorf("GetDefaultLogDir() = %q, should be absolute path", dir)
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		check   func(string) bool
	}{
		{
			name:    "relative path",
			path:    "./test",
			wantErr: false,
			check: func(result string) bool {
				return filepath.IsAbs(result) && contains(result, "test")
			},
		},
		{
			name:    "home directory expansion",
			path:    "~/test",
			wantErr: false,
			check: func(result string) bool {
				return filepath.IsAbs(result) && !contains(result, "~")
			},
		},
		{
			name:    "environment variable expansion",
			path:    "$HOME/test",
			wantErr: false,
			check: func(result string) bool {
				return filepath.IsAbs(result) && !contains(result, "$HOME")
			},
		},
		{
			name:    "absolute path",
			path:    "/tmp/test",
			wantErr: false,
			check: func(result string) bool {
				return result == "/tmp/test"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(result) {
				t.Errorf("expandPath(%q) = %q, check failed", tt.path, result)
			}
		})
	}
}

func TestValidateEnv(t *testing.T) {
	// Save original environment
	originalClaude := os.Getenv("CLAUDE_API_KEY")
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalGoogle := os.Getenv("GOOGLE_API_KEY")
	
	defer func() {
		// Restore original environment
		os.Setenv("CLAUDE_API_KEY", originalClaude)
		os.Setenv("OPENAI_API_KEY", originalOpenAI)
		os.Setenv("GOOGLE_API_KEY", originalGoogle)
	}()

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "all keys present",
			setup: func() {
				os.Setenv("CLAUDE_API_KEY", "test-key-1")
				os.Setenv("OPENAI_API_KEY", "test-key-2")
				os.Setenv("GOOGLE_API_KEY", "test-key-3")
			},
			wantErr: false,
		},
		{
			name: "missing all keys",
			setup: func() {
				os.Unsetenv("CLAUDE_API_KEY")
				os.Unsetenv("OPENAI_API_KEY")
				os.Unsetenv("GOOGLE_API_KEY")
			},
			wantErr: true,
		},
		{
			name: "missing one key",
			setup: func() {
				os.Setenv("CLAUDE_API_KEY", "test-key-1")
				os.Setenv("OPENAI_API_KEY", "test-key-2")
				os.Unsetenv("GOOGLE_API_KEY")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := ValidateEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadAndValidateEnv(t *testing.T) {
	// Save original environment
	originalClaude := os.Getenv("CLAUDE_API_KEY")
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalGoogle := os.Getenv("GOOGLE_API_KEY")
	
	defer func() {
		// Restore original environment
		os.Setenv("CLAUDE_API_KEY", originalClaude)
		os.Setenv("OPENAI_API_KEY", originalOpenAI)
		os.Setenv("GOOGLE_API_KEY", originalGoogle)
	}()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid env file with all keys",
			content: `CLAUDE_API_KEY=test-key-1
OPENAI_API_KEY=test-key-2
GOOGLE_API_KEY=test-key-3`,
			wantErr: false,
		},
		{
			name: "missing required keys",
			content: `CLAUDE_API_KEY=test-key-1
OTHER_KEY=test-value`,
			wantErr: true,
		},
		{
			name:    "invalid format",
			content: `INVALID LINE WITHOUT EQUALS`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("CLAUDE_API_KEY")
			os.Unsetenv("OPENAI_API_KEY")
			os.Unsetenv("GOOGLE_API_KEY")
			
			// Create temp env file
			tmpDir := t.TempDir()
			envPath := filepath.Join(tmpDir, ".env")
			err := os.WriteFile(envPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test env file: %v", err)
			}
			
			err = LoadAndValidateEnv(envPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAndValidateEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadEnv_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func() bool
	}{
		{
			name: "quoted values",
			content: `TEST_KEY="quoted value"
TEST_KEY2='single quoted'`,
			wantErr: false,
			check: func() bool {
				return os.Getenv("TEST_KEY") == "quoted value" &&
					os.Getenv("TEST_KEY2") == "single quoted"
			},
		},
		{
			name: "comments and empty lines",
			content: `# This is a comment
TEST_KEY=value1

# Another comment
TEST_KEY2=value2`,
			wantErr: false,
			check: func() bool {
				return os.Getenv("TEST_KEY") == "value1" &&
					os.Getenv("TEST_KEY2") == "value2"
			},
		},
		{
			name: "whitespace handling",
			content: `  TEST_KEY  =  value with spaces  
TEST_KEY2=value2`,
			wantErr: false,
			check: func() bool {
				return os.Getenv("TEST_KEY") == "value with spaces" &&
					os.Getenv("TEST_KEY2") == "value2"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp env file
			tmpDir := t.TempDir()
			envPath := filepath.Join(tmpDir, ".env")
			err := os.WriteFile(envPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test env file: %v", err)
			}
			
			err = LoadEnv(envPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && !tt.check() {
				t.Errorf("LoadEnv() check failed")
			}
		})
	}
}
