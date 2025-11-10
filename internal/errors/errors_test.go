package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		category ErrorCategory
		message  string
	}{
		{"config error", ConfigError, "invalid configuration"},
		{"dependency error", DependencyError, "missing binary"},
		{"process error", ProcessError, "failed to start"},
		{"network error", NetworkError, "connection refused"},
		{"user error", UserError, "invalid command"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.category, tt.message)
			if err.Category != tt.category {
				t.Errorf("Expected category %s, got %s", tt.category, err.Category)
			}
			if err.Message != tt.message {
				t.Errorf("Expected message %s, got %s", tt.message, err.Message)
			}
			if err.Err != nil {
				t.Errorf("Expected nil wrapped error, got %v", err.Err)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrap(originalErr, ConfigError, "wrapped message")

	if wrapped.Category != ConfigError {
		t.Errorf("Expected category %s, got %s", ConfigError, wrapped.Category)
	}
	if wrapped.Message != "wrapped message" {
		t.Errorf("Expected message 'wrapped message', got %s", wrapped.Message)
	}
	if wrapped.Err != originalErr {
		t.Errorf("Expected wrapped error to be original error")
	}
}

func TestASCError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ASCError
		expected string
	}{
		{
			name:     "simple error",
			err:      New(ConfigError, "test message"),
			expected: "Configuration Error: test message",
		},
		{
			name:     "wrapped error",
			err:      Wrap(errors.New("underlying"), ProcessError, "process failed"),
			expected: "Process Error: process failed (underlying)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestASCError_Unwrap(t *testing.T) {
	originalErr := errors.New("original")
	wrapped := Wrap(originalErr, ConfigError, "wrapped")

	unwrapped := wrapped.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Expected unwrapped error to be original error")
	}

	simple := New(ConfigError, "simple")
	if simple.Unwrap() != nil {
		t.Errorf("Expected nil unwrap for simple error")
	}
}

func TestASCError_WithReason(t *testing.T) {
	err := New(ConfigError, "test").WithReason("test reason")
	if err.Reason != "test reason" {
		t.Errorf("Expected reason 'test reason', got %s", err.Reason)
	}
}

func TestASCError_WithSolution(t *testing.T) {
	err := New(ConfigError, "test").WithSolution("test solution")
	if err.Solution != "test solution" {
		t.Errorf("Expected solution 'test solution', got %s", err.Solution)
	}
}

func TestASCError_Chaining(t *testing.T) {
	err := New(ConfigError, "test").
		WithReason("reason").
		WithSolution("solution")

	if err.Reason != "reason" {
		t.Errorf("Expected reason 'reason', got %s", err.Reason)
	}
	if err.Solution != "solution" {
		t.Errorf("Expected solution 'solution', got %s", err.Solution)
	}
}

func TestASCError_FormatCLI(t *testing.T) {
	tests := []struct {
		name     string
		err      *ASCError
		contains []string
	}{
		{
			name:     "simple error",
			err:      New(ConfigError, "test message"),
			contains: []string{"Error:", "test message"},
		},
		{
			name:     "error with reason",
			err:      New(ConfigError, "test").WithReason("test reason"),
			contains: []string{"Error:", "test", "Reason:", "test reason"},
		},
		{
			name:     "error with solution",
			err:      New(ConfigError, "test").WithSolution("test solution"),
			contains: []string{"Error:", "test", "Solution:", "test solution"},
		},
		{
			name:     "error with all fields",
			err:      Wrap(errors.New("underlying"), ConfigError, "test").WithReason("reason").WithSolution("solution"),
			contains: []string{"Error:", "test", "Reason:", "reason", "Solution:", "solution", "Details:", "underlying"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.FormatCLI()
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, result)
				}
			}
		})
	}
}

func TestASCError_FormatTUI(t *testing.T) {
	tests := []struct {
		name     string
		err      *ASCError
		contains []string
	}{
		{
			name:     "simple error",
			err:      New(ConfigError, "test message"),
			contains: []string{"[ERROR]", "test message"},
		},
		{
			name:     "error with reason",
			err:      New(ConfigError, "test").WithReason("test reason"),
			contains: []string{"[ERROR]", "test", "→", "test reason"},
		},
		{
			name:     "error with solution",
			err:      New(ConfigError, "test").WithSolution("test solution"),
			contains: []string{"[ERROR]", "test", "✓", "test solution"},
		},
		{
			name:     "error with all fields",
			err:      New(ConfigError, "test").WithReason("reason").WithSolution("solution"),
			contains: []string{"[ERROR]", "test", "→", "reason", "✓", "solution"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.FormatTUI()
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, result)
				}
			}
		})
	}
}

func TestNewConfigError(t *testing.T) {
	originalErr := errors.New("parse error")
	err := NewConfigError("failed to parse", originalErr)

	if err.Category != ConfigError {
		t.Errorf("Expected category %s, got %s", ConfigError, err.Category)
	}
	if err.Message != "failed to parse" {
		t.Errorf("Expected message 'failed to parse', got %s", err.Message)
	}
	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be original error")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewDependencyError(t *testing.T) {
	err := NewDependencyError("python3")

	if err.Category != DependencyError {
		t.Errorf("Expected category %s, got %s", DependencyError, err.Category)
	}
	if !strings.Contains(err.Message, "python3") {
		t.Errorf("Expected message to contain 'python3', got %s", err.Message)
	}
	if err.Reason == "" {
		t.Error("Expected reason to be set")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewProcessError(t *testing.T) {
	originalErr := errors.New("exit code 1")
	err := NewProcessError("test-agent", originalErr)

	if err.Category != ProcessError {
		t.Errorf("Expected category %s, got %s", ProcessError, err.Category)
	}
	if !strings.Contains(err.Message, "test-agent") {
		t.Errorf("Expected message to contain 'test-agent', got %s", err.Message)
	}
	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be original error")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewNetworkError(t *testing.T) {
	originalErr := errors.New("connection refused")
	err := NewNetworkError("mcp_agent_mail", originalErr)

	if err.Category != NetworkError {
		t.Errorf("Expected category %s, got %s", NetworkError, err.Category)
	}
	if !strings.Contains(err.Message, "mcp_agent_mail") {
		t.Errorf("Expected message to contain 'mcp_agent_mail', got %s", err.Message)
	}
	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be original error")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewUserError(t *testing.T) {
	err := NewUserError("invalid command")

	if err.Category != UserError {
		t.Errorf("Expected category %s, got %s", UserError, err.Category)
	}
	if err.Message != "invalid command" {
		t.Errorf("Expected message 'invalid command', got %s", err.Message)
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewFileNotFoundError(t *testing.T) {
	err := NewFileNotFoundError("asc.toml")

	if err.Category != ConfigError {
		t.Errorf("Expected category %s, got %s", ConfigError, err.Category)
	}
	if !strings.Contains(err.Message, "asc.toml") {
		t.Errorf("Expected message to contain 'asc.toml', got %s", err.Message)
	}
	if err.Reason == "" {
		t.Error("Expected reason to be set")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewInvalidConfigError(t *testing.T) {
	err := NewInvalidConfigError("beads_db_path", "path does not exist")

	if err.Category != ConfigError {
		t.Errorf("Expected category %s, got %s", ConfigError, err.Category)
	}
	if !strings.Contains(err.Message, "beads_db_path") {
		t.Errorf("Expected message to contain 'beads_db_path', got %s", err.Message)
	}
	if err.Reason != "path does not exist" {
		t.Errorf("Expected reason 'path does not exist', got %s", err.Reason)
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewProcessStartError(t *testing.T) {
	originalErr := errors.New("exec failed")
	err := NewProcessStartError("test-agent", originalErr)

	if err.Category != ProcessError {
		t.Errorf("Expected category %s, got %s", ProcessError, err.Category)
	}
	if !strings.Contains(err.Message, "test-agent") {
		t.Errorf("Expected message to contain 'test-agent', got %s", err.Message)
	}
	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be original error")
	}
	if err.Reason == "" {
		t.Error("Expected reason to be set")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewProcessStopError(t *testing.T) {
	originalErr := errors.New("kill failed")
	err := NewProcessStopError("test-agent", originalErr)

	if err.Category != ProcessError {
		t.Errorf("Expected category %s, got %s", ProcessError, err.Category)
	}
	if !strings.Contains(err.Message, "test-agent") {
		t.Errorf("Expected message to contain 'test-agent', got %s", err.Message)
	}
	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be original error")
	}
	if err.Reason == "" {
		t.Error("Expected reason to be set")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewBeadsError(t *testing.T) {
	originalErr := errors.New("git pull failed")
	err := NewBeadsError("refresh", originalErr)

	if err.Category != NetworkError {
		t.Errorf("Expected category %s, got %s", NetworkError, err.Category)
	}
	if !strings.Contains(err.Message, "refresh") {
		t.Errorf("Expected message to contain 'refresh', got %s", err.Message)
	}
	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be original error")
	}
	if err.Reason == "" {
		t.Error("Expected reason to be set")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewMCPError(t *testing.T) {
	originalErr := errors.New("connection timeout")
	err := NewMCPError("get messages", originalErr)

	if err.Category != NetworkError {
		t.Errorf("Expected category %s, got %s", NetworkError, err.Category)
	}
	if !strings.Contains(err.Message, "get messages") {
		t.Errorf("Expected message to contain 'get messages', got %s", err.Message)
	}
	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be original error")
	}
	if err.Reason == "" {
		t.Error("Expected reason to be set")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewEnvError(t *testing.T) {
	err := NewEnvError("CLAUDE_API_KEY")

	if err.Category != ConfigError {
		t.Errorf("Expected category %s, got %s", ConfigError, err.Category)
	}
	if !strings.Contains(err.Message, "CLAUDE_API_KEY") {
		t.Errorf("Expected message to contain 'CLAUDE_API_KEY', got %s", err.Message)
	}
	if err.Reason == "" {
		t.Error("Expected reason to be set")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestNewAPIKeyError(t *testing.T) {
	err := NewAPIKeyError("Claude")

	if err.Category != ConfigError {
		t.Errorf("Expected category %s, got %s", ConfigError, err.Category)
	}
	if !strings.Contains(err.Message, "Claude") {
		t.Errorf("Expected message to contain 'Claude', got %s", err.Message)
	}
	if err.Reason == "" {
		t.Error("Expected reason to be set")
	}
	if err.Solution == "" {
		t.Error("Expected solution to be set")
	}
}

func TestErrorCategories(t *testing.T) {
	categories := []ErrorCategory{
		ConfigError,
		DependencyError,
		ProcessError,
		NetworkError,
		UserError,
	}

	for _, cat := range categories {
		t.Run(string(cat), func(t *testing.T) {
			err := New(cat, "test")
			if err.Category != cat {
				t.Errorf("Expected category %s, got %s", cat, err.Category)
			}
		})
	}
}

// Test error wrapping with errors.Is and errors.As
func TestErrorWrapping(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrapped := Wrap(originalErr, ConfigError, "wrapped")

	if !errors.Is(wrapped, originalErr) {
		t.Error("Expected errors.Is to work with wrapped error")
	}

	var ascErr *ASCError
	if !errors.As(wrapped, &ascErr) {
		t.Error("Expected errors.As to work with ASCError")
	}
}
