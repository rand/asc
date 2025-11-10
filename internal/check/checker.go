// Package check provides dependency verification for the Agent Stack Controller.
// It checks for required binaries, configuration files, and environment variables,
// with support for formatted output using lipgloss.
//
// Example usage:
//
//	checker := check.NewChecker("asc.toml", ".env")
//	results := checker.RunAll()
//	fmt.Println(check.FormatResults(results))
//	if check.HasFailures(results) {
//	    os.Exit(1)
//	}
package check

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

// CheckStatus represents the status of a check.
type CheckStatus string

const (
	CheckPass CheckStatus = "pass"
	CheckFail CheckStatus = "fail"
	CheckWarn CheckStatus = "warn"
)

// CheckResult represents the result of a single check including
// the component name, status, and a descriptive message.
type CheckResult struct {
	Name    string
	Status  CheckStatus
	Message string
}

// Checker defines the interface for dependency checking
type Checker interface {
	CheckBinary(name string) CheckResult
	CheckFile(path string) CheckResult
	CheckConfig() CheckResult
	CheckEnv(keys []string) CheckResult
	RunAll() []CheckResult
}

// DefaultChecker implements the Checker interface.
// It performs checks against the file system and system PATH.
type DefaultChecker struct {
	configPath string // Path to asc.toml configuration file
	envPath    string // Path to .env environment file
}

// NewChecker creates a new DefaultChecker instance with the specified
// configuration and environment file paths.
//
// Example:
//
//	checker := check.NewChecker("asc.toml", ".env")
func NewChecker(configPath, envPath string) Checker {
	return &DefaultChecker{
		configPath: configPath,
		envPath:    envPath,
	}
}

// CheckBinary checks if a binary exists in the system PATH.
// Returns CheckPass if found, CheckFail otherwise.
func (c *DefaultChecker) CheckBinary(name string) CheckResult {
	_, err := exec.LookPath(name)
	if err != nil {
		return CheckResult{
			Name:    name,
			Status:  CheckFail,
			Message: fmt.Sprintf("Binary '%s' not found in PATH", name),
		}
	}
	return CheckResult{
		Name:    name,
		Status:  CheckPass,
		Message: fmt.Sprintf("Binary '%s' found", name),
	}
}

// CheckFile checks if a file exists and is readable.
// It expands ~ to the home directory and verifies the file is not a directory.
func (c *DefaultChecker) CheckFile(path string) CheckResult {
	// Expand home directory if present
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return CheckResult{
				Name:    filepath.Base(path),
				Status:  CheckFail,
				Message: fmt.Sprintf("File '%s' does not exist", path),
			}
		}
		return CheckResult{
			Name:    filepath.Base(path),
			Status:  CheckFail,
			Message: fmt.Sprintf("Cannot access file '%s': %v", path, err),
		}
	}

	if info.IsDir() {
		return CheckResult{
			Name:    filepath.Base(path),
			Status:  CheckFail,
			Message: fmt.Sprintf("'%s' is a directory, not a file", path),
		}
	}

	// Check if file is readable
	file, err := os.Open(path)
	if err != nil {
		return CheckResult{
			Name:    filepath.Base(path),
			Status:  CheckFail,
			Message: fmt.Sprintf("File '%s' is not readable: %v", path, err),
		}
	}
	file.Close()

	return CheckResult{
		Name:    filepath.Base(path),
		Status:  CheckPass,
		Message: fmt.Sprintf("File '%s' exists and is readable", path),
	}
}

// CheckConfig checks if the configuration file exists and is valid TOML.
// It verifies the file can be parsed and contains required fields.
func (c *DefaultChecker) CheckConfig() CheckResult {
	// First check if file exists
	fileResult := c.CheckFile(c.configPath)
	if fileResult.Status == CheckFail {
		return CheckResult{
			Name:    "asc.toml",
			Status:  CheckFail,
			Message: fmt.Sprintf("Config file not found at '%s'", c.configPath),
		}
	}

	// Try to parse the TOML file
	v := viper.New()
	v.SetConfigFile(c.configPath)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return CheckResult{
			Name:    "asc.toml",
			Status:  CheckFail,
			Message: fmt.Sprintf("Invalid TOML syntax: %v", err),
		}
	}

	// Validate required fields
	if !v.IsSet("core.beads_db_path") {
		return CheckResult{
			Name:    "asc.toml",
			Status:  CheckFail,
			Message: "Missing required field: core.beads_db_path",
		}
	}

	if !v.IsSet("services.mcp_agent_mail") {
		return CheckResult{
			Name:    "asc.toml",
			Status:  CheckWarn,
			Message: "Missing services.mcp_agent_mail configuration",
		}
	}

	return CheckResult{
		Name:    "asc.toml",
		Status:  CheckPass,
		Message: "Configuration file is valid",
	}
}

// CheckEnv checks if required environment variables or .env file keys exist.
// It reads the .env file and verifies that all specified keys are present.
// Returns CheckWarn if keys are missing, CheckPass if all are present.
func (c *DefaultChecker) CheckEnv(keys []string) CheckResult {
	// First check if .env file exists
	fileResult := c.CheckFile(c.envPath)
	if fileResult.Status == CheckFail {
		return CheckResult{
			Name:    ".env",
			Status:  CheckFail,
			Message: fmt.Sprintf(".env file not found at '%s'", c.envPath),
		}
	}

	// Read .env file and check for required keys
	content, err := os.ReadFile(c.envPath)
	if err != nil {
		return CheckResult{
			Name:    ".env",
			Status:  CheckFail,
			Message: fmt.Sprintf("Cannot read .env file: %v", err),
		}
	}

	// Simple check for key presence (not a full .env parser)
	missingKeys := []string{}
	for _, key := range keys {
		found := false
		// Check if key exists in file content
		lines := string(content)
		if len(lines) > 0 {
			// Simple substring check for key=
			searchStr := key + "="
			if len(lines) >= len(searchStr) {
				for i := 0; i <= len(lines)-len(searchStr); i++ {
					if lines[i:i+len(searchStr)] == searchStr {
						found = true
						break
					}
				}
			}
		}
		if !found {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return CheckResult{
			Name:    ".env",
			Status:  CheckWarn,
			Message: fmt.Sprintf("Missing API keys: %v", missingKeys),
		}
	}

	return CheckResult{
		Name:    ".env",
		Status:  CheckPass,
		Message: "All required API keys present",
	}
}

// RunAll runs all dependency checks including binaries, configuration,
// and environment variables. Returns a slice of CheckResult for each check.
func (c *DefaultChecker) RunAll() []CheckResult {
	results := []CheckResult{}

	// Check required binaries
	binaries := []string{"git", "python3", "uv", "bd"}
	for _, binary := range binaries {
		results = append(results, c.CheckBinary(binary))
	}

	// Check optional binaries
	dockerResult := c.CheckBinary("docker")
	if dockerResult.Status == CheckFail {
		dockerResult.Status = CheckWarn
		dockerResult.Message = "Docker not found (optional)"
	}
	results = append(results, dockerResult)

	// Check age for secrets management
	ageResult := c.CheckBinary("age")
	if ageResult.Status == CheckFail {
		ageResult.Status = CheckWarn
		ageResult.Message = "age not found (recommended for secrets management)"
	}
	results = append(results, ageResult)

	// Check configuration file
	results = append(results, c.CheckConfig())

	// Check environment file
	requiredKeys := []string{"CLAUDE_API_KEY", "OPENAI_API_KEY", "GOOGLE_API_KEY"}
	results = append(results, c.CheckEnv(requiredKeys))

	return results
}

// FormatResults formats check results as a styled table using lipgloss.
// It color-codes results (green for pass, red for fail, yellow for warn)
// and returns a formatted string suitable for terminal output.
func FormatResults(results []CheckResult) string {
	// Define styles
	passStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)  // Green
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)   // Red
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)  // Yellow
	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	
	// Build table
	var output string
	output += headerStyle.Render("Dependency Checks") + "\n\n"
	
	// Column widths
	nameWidth := 20
	statusWidth := 10
	
	// Header row
	output += fmt.Sprintf("%-*s %-*s %s\n", nameWidth, "Component", statusWidth, "Status", "Message")
	output += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
		fmt.Sprintf("%s %s %s\n", 
			lipgloss.NewStyle().Width(nameWidth).Render("─────────────────────"),
			lipgloss.NewStyle().Width(statusWidth).Render("──────────"),
			"────────────────────────────────────────────────────────────────"),
	)
	
	// Data rows
	for _, result := range results {
		var statusStr string
		switch result.Status {
		case CheckPass:
			statusStr = passStyle.Render("✓ PASS")
		case CheckFail:
			statusStr = failStyle.Render("✗ FAIL")
		case CheckWarn:
			statusStr = warnStyle.Render("⚠ WARN")
		}
		
		output += fmt.Sprintf("%-*s %-*s %s\n", 
			nameWidth, result.Name, 
			statusWidth+10, statusStr, // +10 for ANSI color codes
			result.Message)
	}
	
	return output
}

// HasFailures returns true if any check failed (CheckFail status).
// Warnings (CheckWarn) are not considered failures.
func HasFailures(results []CheckResult) bool {
	for _, result := range results {
		if result.Status == CheckFail {
			return true
		}
	}
	return false
}
