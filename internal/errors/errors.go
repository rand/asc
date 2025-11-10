// Package errors provides structured error handling for the Agent Stack Controller.
// It defines error categories, formatting utilities, and common error constructors
// with actionable solutions for users.
//
// Example usage:
//
//	err := errors.NewDependencyError("python3")
//	fmt.Fprintln(os.Stderr, err.FormatCLI())
//	os.Exit(1)
package errors

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ErrorCategory represents the type of error for classification and handling.
type ErrorCategory string

const (
	ConfigError     ErrorCategory = "Configuration Error"
	DependencyError ErrorCategory = "Dependency Error"
	ProcessError    ErrorCategory = "Process Error"
	NetworkError    ErrorCategory = "Network Error"
	UserError       ErrorCategory = "User Error"
)

// ASCError represents a structured error with context and actionable solutions.
// It includes a category, message, optional reason, solution, and wrapped error.
type ASCError struct {
	Category ErrorCategory
	Message  string
	Reason   string
	Solution string
	Err      error
}

// Error implements the error interface
func (e *ASCError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Category, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Category, e.Message)
}

// Unwrap returns the underlying error
func (e *ASCError) Unwrap() error {
	return e.Err
}

// New creates a new ASCError with the specified category and message.
// Use WithReason and WithSolution to add additional context.
func New(category ErrorCategory, message string) *ASCError {
	return &ASCError{
		Category: category,
		Message:  message,
	}
}

// Wrap wraps an existing error with ASC context
func Wrap(err error, category ErrorCategory, message string) *ASCError {
	return &ASCError{
		Category: category,
		Message:  message,
		Err:      err,
	}
}

// WithReason adds a reason to the error
func (e *ASCError) WithReason(reason string) *ASCError {
	e.Reason = reason
	return e
}

// WithSolution adds a solution to the error
func (e *ASCError) WithSolution(solution string) *ASCError {
	e.Solution = solution
	return e
}

// FormatCLI formats the error for CLI output (stderr)
func (e *ASCError) FormatCLI() string {
	var sb strings.Builder
	
	// Error header with category
	errorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9"))
	
	sb.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", e.Message)))
	sb.WriteString("\n")
	
	// Reason if provided
	if e.Reason != "" {
		reasonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))
		sb.WriteString(reasonStyle.Render(fmt.Sprintf("Reason: %s", e.Reason)))
		sb.WriteString("\n")
	}
	
	// Solution if provided
	if e.Solution != "" {
		solutionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))
		sb.WriteString(solutionStyle.Render(fmt.Sprintf("Solution: %s", e.Solution)))
		sb.WriteString("\n")
	}
	
	// Underlying error if present
	if e.Err != nil {
		detailStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
		sb.WriteString(detailStyle.Render(fmt.Sprintf("Details: %v", e.Err)))
		sb.WriteString("\n")
	}
	
	return sb.String()
}

// FormatTUI formats the error for TUI display (log pane)
func (e *ASCError) FormatTUI() string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("[ERROR] %s", e.Message))
	
	if e.Reason != "" {
		parts = append(parts, fmt.Sprintf("  → %s", e.Reason))
	}
	
	if e.Solution != "" {
		parts = append(parts, fmt.Sprintf("  ✓ %s", e.Solution))
	}
	
	return strings.Join(parts, "\n")
}

// Common error constructors with predefined solutions

// NewConfigError creates a configuration error
func NewConfigError(message string, err error) *ASCError {
	return Wrap(err, ConfigError, message).
		WithSolution("Check your asc.toml file for syntax errors and required fields")
}

// NewDependencyError creates a dependency error
func NewDependencyError(dependency string) *ASCError {
	return New(DependencyError, fmt.Sprintf("Missing required dependency: %s", dependency)).
		WithReason(fmt.Sprintf("Command '%s' not found in system PATH", dependency)).
		WithSolution(fmt.Sprintf("Install %s and ensure it's in your PATH. Run 'asc check' for details", dependency))
}

// NewProcessError creates a process error
func NewProcessError(processName string, err error) *ASCError {
	return Wrap(err, ProcessError, fmt.Sprintf("Failed to manage process '%s'", processName)).
		WithSolution("Check process logs in ~/.asc/logs/ for details")
}

// NewNetworkError creates a network error
func NewNetworkError(service string, err error) *ASCError {
	return Wrap(err, NetworkError, fmt.Sprintf("Failed to connect to %s", service)).
		WithSolution(fmt.Sprintf("Ensure %s is running and accessible. Check service status with 'asc services status'", service))
}

// NewUserError creates a user error
func NewUserError(message string) *ASCError {
	return New(UserError, message).
		WithSolution("Run 'asc --help' for usage information")
}

// NewFileNotFoundError creates a file not found error
func NewFileNotFoundError(filename string) *ASCError {
	return New(ConfigError, fmt.Sprintf("Required file not found: %s", filename)).
		WithReason(fmt.Sprintf("File '%s' does not exist or is not readable", filename)).
		WithSolution(fmt.Sprintf("Create %s or run 'asc init' to generate default configuration", filename))
}

// NewInvalidConfigError creates an invalid configuration error
func NewInvalidConfigError(field string, reason string) *ASCError {
	return New(ConfigError, fmt.Sprintf("Invalid configuration field: %s", field)).
		WithReason(reason).
		WithSolution("Update asc.toml with valid values. See documentation for examples")
}

// NewProcessStartError creates a process start error
func NewProcessStartError(processName string, err error) *ASCError {
	return Wrap(err, ProcessError, fmt.Sprintf("Failed to start process '%s'", processName)).
		WithReason("Process failed to launch or exited immediately").
		WithSolution("Check that the command exists and has correct permissions. Review logs in ~/.asc/logs/")
}

// NewProcessStopError creates a process stop error
func NewProcessStopError(processName string, err error) *ASCError {
	return Wrap(err, ProcessError, fmt.Sprintf("Failed to stop process '%s'", processName)).
		WithReason("Process did not respond to termination signal").
		WithSolution("Process may have already stopped. Use 'ps' to check if it's still running")
}

// NewBeadsError creates a beads-related error
func NewBeadsError(operation string, err error) *ASCError {
	return Wrap(err, NetworkError, fmt.Sprintf("Beads operation failed: %s", operation)).
		WithReason("Failed to communicate with beads database").
		WithSolution("Ensure beads is installed ('bd --version') and the database path in asc.toml is correct")
}

// NewMCPError creates an MCP-related error
func NewMCPError(operation string, err error) *ASCError {
	return Wrap(err, NetworkError, fmt.Sprintf("MCP operation failed: %s", operation)).
		WithReason("Failed to communicate with mcp_agent_mail server").
		WithSolution("Ensure mcp_agent_mail is running. Start it with 'asc services start'")
}

// NewEnvError creates an environment variable error
func NewEnvError(key string) *ASCError {
	return New(ConfigError, fmt.Sprintf("Missing required environment variable: %s", key)).
		WithReason(fmt.Sprintf("Environment variable '%s' is not set", key)).
		WithSolution("Add the variable to your .env file or run 'asc init' to configure API keys")
}

// NewAPIKeyError creates an API key error
func NewAPIKeyError(provider string) *ASCError {
	return New(ConfigError, fmt.Sprintf("Missing or invalid API key for %s", provider)).
		WithReason(fmt.Sprintf("%s API key is required but not found in .env file", provider)).
		WithSolution("Add your API key to .env file or run 'asc init' to configure")
}
