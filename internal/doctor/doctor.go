// Package doctor provides comprehensive diagnostics and remediation for the Agent Stack Controller.
// It detects common configuration issues, corrupted state, permission problems, and provides
// actionable remediation steps with optional automatic fixes.
package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/rand/asc/internal/check"
	"github.com/rand/asc/internal/logger"
	"github.com/rand/asc/internal/process"
)

// IssueSeverity represents the severity level of a detected issue
type IssueSeverity string

const (
	SeverityCritical IssueSeverity = "critical"
	SeverityHigh     IssueSeverity = "high"
	SeverityMedium   IssueSeverity = "medium"
	SeverityLow      IssueSeverity = "low"
	SeverityInfo     IssueSeverity = "info"
)

// IssueCategory represents the category of issue
type IssueCategory string

const (
	CategoryConfiguration IssueCategory = "configuration"
	CategoryState         IssueCategory = "state"
	CategoryPermissions   IssueCategory = "permissions"
	CategoryResources     IssueCategory = "resources"
	CategoryNetwork       IssueCategory = "network"
	CategoryAgent         IssueCategory = "agent"
)

// Issue represents a detected problem
type Issue struct {
	ID          string        `json:"id"`
	Category    IssueCategory `json:"category"`
	Severity    IssueSeverity `json:"severity"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Impact      string        `json:"impact"`
	Remediation string        `json:"remediation"`
	AutoFixable bool          `json:"auto_fixable"`
	DetectedAt  time.Time     `json:"detected_at"`
}

// FixResult represents the result of applying a fix
type FixResult struct {
	IssueID   string    `json:"issue_id"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	AppliedAt time.Time `json:"applied_at"`
}

// DiagnosticReport contains all detected issues and fix results
type DiagnosticReport struct {
	RunAt         time.Time    `json:"run_at"`
	Issues        []Issue      `json:"issues"`
	FixesApplied  []FixResult  `json:"fixes_applied,omitempty"`
	HealthSummary string       `json:"health_summary"`
}

// Doctor performs diagnostics and remediation
type Doctor struct {
	configPath string
	envPath    string
	checker    check.Checker
	homeDir    string
}

// NewDoctor creates a new Doctor instance
func NewDoctor(configPath, envPath string) (*Doctor, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	return &Doctor{
		configPath: configPath,
		envPath:    envPath,
		checker:    check.NewChecker(configPath, envPath),
		homeDir:    homeDir,
	}, nil
}

// RunDiagnostics performs all diagnostic checks
func (d *Doctor) RunDiagnostics() (*DiagnosticReport, error) {
	logger.Info("Running comprehensive diagnostics...")
	
	report := &DiagnosticReport{
		RunAt:  time.Now(),
		Issues: []Issue{},
	}
	
	// Run all diagnostic checks
	d.checkConfiguration(report)
	d.checkState(report)
	d.checkPermissions(report)
	d.checkResources(report)
	d.checkNetwork(report)
	d.checkAgents(report)
	
	// Generate health summary
	report.HealthSummary = d.generateHealthSummary(report)
	
	logger.Info("Diagnostics complete: found %d issue(s)", len(report.Issues))
	return report, nil
}

// checkConfiguration validates configuration files and settings
func (d *Doctor) checkConfiguration(report *DiagnosticReport) {
	// Check if config file exists
	if _, err := os.Stat(d.configPath); os.IsNotExist(err) {
		report.Issues = append(report.Issues, Issue{
			ID:          "config-missing",
			Category:    CategoryConfiguration,
			Severity:    SeverityCritical,
			Title:       "Configuration file not found",
			Description: fmt.Sprintf("The configuration file '%s' does not exist", d.configPath),
			Impact:      "asc cannot start without a valid configuration file",
			Remediation: "Run 'asc init' to create a default configuration file",
			AutoFixable: true,
			DetectedAt:  time.Now(),
		})
		return
	}
	
	// Check config validity
	configResult := d.checker.CheckConfig()
	if configResult.Status == check.CheckFail {
		report.Issues = append(report.Issues, Issue{
			ID:          "config-invalid",
			Category:    CategoryConfiguration,
			Severity:    SeverityCritical,
			Title:       "Invalid configuration",
			Description: configResult.Message,
			Impact:      "asc cannot parse the configuration file",
			Remediation: "Fix the TOML syntax errors or run 'asc init' to regenerate",
			AutoFixable: false,
			DetectedAt:  time.Now(),
		})
	}
	
	// Check .env file
	if _, err := os.Stat(d.envPath); os.IsNotExist(err) {
		report.Issues = append(report.Issues, Issue{
			ID:          "env-missing",
			Category:    CategoryConfiguration,
			Severity:    SeverityHigh,
			Title:       "Environment file not found",
			Description: fmt.Sprintf("The .env file '%s' does not exist", d.envPath),
			Impact:      "API keys will not be loaded, agents cannot authenticate",
			Remediation: "Create a .env file with required API keys (CLAUDE_API_KEY, OPENAI_API_KEY, GOOGLE_API_KEY)",
			AutoFixable: false,
			DetectedAt:  time.Now(),
		})
	} else {
		// Check .env permissions
		info, err := os.Stat(d.envPath)
		if err == nil {
			mode := info.Mode().Perm()
			if mode&0077 != 0 {
				report.Issues = append(report.Issues, Issue{
					ID:          "env-permissions",
					Category:    CategoryPermissions,
					Severity:    SeverityMedium,
					Title:       "Insecure .env file permissions",
					Description: fmt.Sprintf(".env file has permissions %o (should be 0600)", mode),
					Impact:      "API keys may be readable by other users",
					Remediation: "Run 'chmod 600 .env' to secure the file",
					AutoFixable: true,
					DetectedAt:  time.Now(),
				})
			}
		}
	}
}

// checkState validates PID files, logs, and other state
func (d *Doctor) checkState(report *DiagnosticReport) {
	ascDir := filepath.Join(d.homeDir, ".asc")
	pidDir := filepath.Join(ascDir, "pids")
	logDir := filepath.Join(ascDir, "logs")
	
	// Check for orphaned PID files
	if _, err := os.Stat(pidDir); err == nil {
		files, err := os.ReadDir(pidDir)
		if err == nil {
			for _, file := range files {
				if filepath.Ext(file.Name()) == ".json" {
					pidPath := filepath.Join(pidDir, file.Name())
					data, err := os.ReadFile(pidPath)
					if err != nil {
						continue
					}
					
					var procInfo process.ProcessInfo
					if err := json.Unmarshal(data, &procInfo); err != nil {
						report.Issues = append(report.Issues, Issue{
							ID:          fmt.Sprintf("pid-corrupted-%s", file.Name()),
							Category:    CategoryState,
							Severity:    SeverityMedium,
							Title:       "Corrupted PID file",
							Description: fmt.Sprintf("PID file '%s' contains invalid JSON", file.Name()),
							Impact:      "Cannot track process status",
							Remediation: fmt.Sprintf("Delete the corrupted file: rm %s", pidPath),
							AutoFixable: true,
							DetectedAt:  time.Now(),
						})
						continue
					}
					
					// Check if process is actually running
					if !isProcessRunning(procInfo.PID) {
						report.Issues = append(report.Issues, Issue{
							ID:          fmt.Sprintf("pid-orphaned-%s", procInfo.Name),
							Category:    CategoryState,
							Severity:    SeverityLow,
							Title:       "Orphaned PID file",
							Description: fmt.Sprintf("PID file exists for '%s' but process %d is not running", procInfo.Name, procInfo.PID),
							Impact:      "Stale state may cause confusion",
							Remediation: fmt.Sprintf("Delete the orphaned file: rm %s", pidPath),
							AutoFixable: true,
							DetectedAt:  time.Now(),
						})
					}
				}
			}
		}
	}
	
	// Check log directory size
	if info, err := os.Stat(logDir); err == nil && info.IsDir() {
		size, err := getDirSize(logDir)
		if err == nil && size > 100*1024*1024 { // 100MB
			report.Issues = append(report.Issues, Issue{
				ID:          "logs-large",
				Category:    CategoryResources,
				Severity:    SeverityLow,
				Title:       "Large log directory",
				Description: fmt.Sprintf("Log directory is %.2f MB", float64(size)/(1024*1024)),
				Impact:      "Consuming excessive disk space",
				Remediation: "Clean old logs: find ~/.asc/logs -name '*.log' -mtime +7 -delete",
				AutoFixable: true,
				DetectedAt:  time.Now(),
			})
		}
	}
}

// checkPermissions validates file and directory permissions
func (d *Doctor) checkPermissions(report *DiagnosticReport) {
	ascDir := filepath.Join(d.homeDir, ".asc")
	
	// Check if .asc directory exists and is writable
	if info, err := os.Stat(ascDir); err == nil {
		if !info.IsDir() {
			report.Issues = append(report.Issues, Issue{
				ID:          "asc-not-dir",
				Category:    CategoryPermissions,
				Severity:    SeverityCritical,
				Title:       "~/.asc is not a directory",
				Description: "~/.asc exists but is a file, not a directory",
				Impact:      "Cannot store state, logs, or PIDs",
				Remediation: "Remove the file and recreate: rm ~/.asc && mkdir ~/.asc",
				AutoFixable: true,
				DetectedAt:  time.Now(),
			})
		} else {
			// Check if writable
			testFile := filepath.Join(ascDir, ".write_test")
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				report.Issues = append(report.Issues, Issue{
					ID:          "asc-not-writable",
					Category:    CategoryPermissions,
					Severity:    SeverityCritical,
					Title:       "~/.asc directory not writable",
					Description: fmt.Sprintf("Cannot write to ~/.asc: %v", err),
					Impact:      "Cannot store state, logs, or PIDs",
					Remediation: "Fix permissions: chmod 755 ~/.asc",
					AutoFixable: true,
					DetectedAt:  time.Now(),
				})
			} else {
				os.Remove(testFile)
			}
		}
	}
	
	// Check subdirectories
	for _, subdir := range []string{"pids", "logs", "playbooks"} {
		dirPath := filepath.Join(ascDir, subdir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			report.Issues = append(report.Issues, Issue{
				ID:          fmt.Sprintf("dir-missing-%s", subdir),
				Category:    CategoryPermissions,
				Severity:    SeverityMedium,
				Title:       fmt.Sprintf("Missing %s directory", subdir),
				Description: fmt.Sprintf("Directory ~/.asc/%s does not exist", subdir),
				Impact:      fmt.Sprintf("Cannot store %s", subdir),
				Remediation: fmt.Sprintf("Create directory: mkdir -p ~/.asc/%s", subdir),
				AutoFixable: true,
				DetectedAt:  time.Now(),
			})
		}
	}
}

// checkResources validates system resources
func (d *Doctor) checkResources(report *DiagnosticReport) {
	// Check disk space
	ascDir := filepath.Join(d.homeDir, ".asc")
	if info, err := os.Stat(ascDir); err == nil && info.IsDir() {
		// Get available disk space (simplified check)
		// In production, use syscall.Statfs or similar
		size, _ := getDirSize(ascDir)
		if size > 500*1024*1024 { // 500MB
			report.Issues = append(report.Issues, Issue{
				ID:          "disk-space-high",
				Category:    CategoryResources,
				Severity:    SeverityMedium,
				Title:       "High disk usage",
				Description: fmt.Sprintf("~/.asc directory is using %.2f MB", float64(size)/(1024*1024)),
				Impact:      "May run out of disk space",
				Remediation: "Clean up old logs and playbooks",
				AutoFixable: false,
				DetectedAt:  time.Now(),
			})
		}
	}
	
	// Check for required binaries
	requiredBinaries := []string{"git", "python3", "uv", "bd"}
	for _, binary := range requiredBinaries {
		result := d.checker.CheckBinary(binary)
		if result.Status == check.CheckFail {
			severity := SeverityCritical
			if binary == "uv" {
				severity = SeverityMedium // uv is recommended but not critical
			}
			
			report.Issues = append(report.Issues, Issue{
				ID:          fmt.Sprintf("binary-missing-%s", binary),
				Category:    CategoryConfiguration,
				Severity:    severity,
				Title:       fmt.Sprintf("Missing required binary: %s", binary),
				Description: result.Message,
				Impact:      fmt.Sprintf("Cannot use features that require %s", binary),
				Remediation: fmt.Sprintf("Install %s and ensure it's in your PATH", binary),
				AutoFixable: false,
				DetectedAt:  time.Now(),
			})
		}
	}
}

// checkNetwork validates network connectivity
func (d *Doctor) checkNetwork(report *DiagnosticReport) {
	// Try to load config to get MCP URL
	v := viper.New()
	v.SetConfigFile(d.configPath)
	v.SetConfigType("toml")
	
	if err := v.ReadInConfig(); err != nil {
		// Config issues already reported in checkConfiguration
		return
	}
	
	// Check MCP server connectivity (basic check)
	mcpURL := v.GetString("services.mcp_agent_mail.url")
	if mcpURL != "" {
		// Note: We don't actually make HTTP requests here to avoid dependencies
		// This is a placeholder for where you'd check connectivity
		report.Issues = append(report.Issues, Issue{
			ID:          "network-check-info",
			Category:    CategoryNetwork,
			Severity:    SeverityInfo,
			Title:       "Network connectivity check",
			Description: fmt.Sprintf("MCP server configured at %s", mcpURL),
			Impact:      "None",
			Remediation: "Verify MCP server is accessible: curl " + mcpURL + "/health",
			AutoFixable: false,
			DetectedAt:  time.Now(),
		})
	}
}

// checkAgents validates agent configuration and state
func (d *Doctor) checkAgents(report *DiagnosticReport) {
	v := viper.New()
	v.SetConfigFile(d.configPath)
	v.SetConfigType("toml")
	
	if err := v.ReadInConfig(); err != nil {
		// Config issues already reported
		return
	}
	
	// Get all agent configurations
	agents := v.GetStringMap("agent")
	
	// Check each agent configuration
	for agentName := range agents {
		agentKey := fmt.Sprintf("agent.%s", agentName)
		
		// Check if agent command exists
		command := v.GetString(agentKey + ".command")
		if command == "" {
			report.Issues = append(report.Issues, Issue{
				ID:          fmt.Sprintf("agent-no-command-%s", agentName),
				Category:    CategoryAgent,
				Severity:    SeverityCritical,
				Title:       fmt.Sprintf("Agent '%s' has no command", agentName),
				Description: "Agent configuration is missing the command field",
				Impact:      "Agent cannot be started",
				Remediation: fmt.Sprintf("Add command field to [agent.%s] in asc.toml", agentName),
				AutoFixable: false,
				DetectedAt:  time.Now(),
			})
		}
		
		// Check if model is valid
		model := v.GetString(agentKey + ".model")
		validModels := []string{"claude", "gemini", "openai", "gpt-4", "codex"}
		modelValid := false
		for _, valid := range validModels {
			if model == valid {
				modelValid = true
				break
			}
		}
		if !modelValid && model != "" {
			report.Issues = append(report.Issues, Issue{
				ID:          fmt.Sprintf("agent-invalid-model-%s", agentName),
				Category:    CategoryAgent,
				Severity:    SeverityHigh,
				Title:       fmt.Sprintf("Agent '%s' has invalid model", agentName),
				Description: fmt.Sprintf("Model '%s' is not recognized", model),
				Impact:      "Agent may fail to start or authenticate",
				Remediation: fmt.Sprintf("Use a valid model: %v", validModels),
				AutoFixable: false,
				DetectedAt:  time.Now(),
			})
		}
		
		// Check if phases are defined
		phases := v.GetStringSlice(agentKey + ".phases")
		if len(phases) == 0 {
			report.Issues = append(report.Issues, Issue{
				ID:          fmt.Sprintf("agent-no-phases-%s", agentName),
				Category:    CategoryAgent,
				Severity:    SeverityHigh,
				Title:       fmt.Sprintf("Agent '%s' has no phases", agentName),
				Description: "Agent configuration is missing phases",
				Impact:      "Agent will not pick up any tasks",
				Remediation: fmt.Sprintf("Add phases to [agent.%s] in asc.toml", agentName),
				AutoFixable: false,
				DetectedAt:  time.Now(),
			})
		}
	}
}

// ApplyFixes attempts to automatically fix issues
func (d *Doctor) ApplyFixes(report *DiagnosticReport) ([]FixResult, error) {
	results := []FixResult{}
	
	for _, issue := range report.Issues {
		if !issue.AutoFixable {
			continue
		}
		
		logger.Info("Attempting to fix: %s", issue.Title)
		
		var success bool
		var message string
		
		switch issue.ID {
		case "env-permissions":
			success, message = d.fixEnvPermissions()
		case "asc-not-dir":
			success, message = d.fixAscNotDir()
		case "asc-not-writable":
			success, message = d.fixAscNotWritable()
		case "logs-large":
			success, message = d.fixLargeLogs()
		default:
			if len(issue.ID) > 13 && issue.ID[:13] == "pid-corrupted" {
				success, message = d.fixCorruptedPID(issue.ID)
			} else if len(issue.ID) > 13 && issue.ID[:13] == "pid-orphaned-" {
				success, message = d.fixOrphanedPID(issue.ID)
			} else if len(issue.ID) > 12 && issue.ID[:12] == "dir-missing-" {
				success, message = d.fixMissingDir(issue.ID)
			} else {
				continue
			}
		}
		
		results = append(results, FixResult{
			IssueID:   issue.ID,
			Success:   success,
			Message:   message,
			AppliedAt: time.Now(),
		})
		
		if success {
			logger.Info("Fixed: %s", issue.Title)
		} else {
			logger.Warn("Failed to fix %s: %s", issue.Title, message)
		}
	}
	
	return results, nil
}

// Fix functions
func (d *Doctor) fixEnvPermissions() (bool, string) {
	if err := os.Chmod(d.envPath, 0600); err != nil {
		return false, fmt.Sprintf("Failed to change permissions: %v", err)
	}
	return true, "Set .env permissions to 0600"
}

func (d *Doctor) fixAscNotDir() (bool, string) {
	ascDir := filepath.Join(d.homeDir, ".asc")
	if err := os.Remove(ascDir); err != nil {
		return false, fmt.Sprintf("Failed to remove file: %v", err)
	}
	if err := os.MkdirAll(ascDir, 0755); err != nil {
		return false, fmt.Sprintf("Failed to create directory: %v", err)
	}
	return true, "Removed file and created directory"
}

func (d *Doctor) fixAscNotWritable() (bool, string) {
	ascDir := filepath.Join(d.homeDir, ".asc")
	if err := os.Chmod(ascDir, 0755); err != nil {
		return false, fmt.Sprintf("Failed to change permissions: %v", err)
	}
	return true, "Set ~/.asc permissions to 0755"
}

func (d *Doctor) fixLargeLogs() (bool, string) {
	logDir := filepath.Join(d.homeDir, ".asc", "logs")
	
	// Delete logs older than 7 days
	deleted := 0
	err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".log" {
			if time.Since(info.ModTime()) > 7*24*time.Hour {
				if err := os.Remove(path); err == nil {
					deleted++
				}
			}
		}
		return nil
	})
	
	if err != nil {
		return false, fmt.Sprintf("Failed to clean logs: %v", err)
	}
	return true, fmt.Sprintf("Deleted %d old log files", deleted)
}

func (d *Doctor) fixCorruptedPID(issueID string) (bool, string) {
	// Extract filename from issue ID
	filename := issueID[14:] // Skip "pid-corrupted-"
	pidPath := filepath.Join(d.homeDir, ".asc", "pids", filename)
	
	if err := os.Remove(pidPath); err != nil {
		return false, fmt.Sprintf("Failed to remove file: %v", err)
	}
	return true, "Removed corrupted PID file"
}

func (d *Doctor) fixOrphanedPID(issueID string) (bool, string) {
	// Extract agent name from issue ID
	agentName := issueID[13:] // Skip "pid-orphaned-"
	pidPath := filepath.Join(d.homeDir, ".asc", "pids", agentName+".json")
	
	if err := os.Remove(pidPath); err != nil {
		return false, fmt.Sprintf("Failed to remove file: %v", err)
	}
	return true, "Removed orphaned PID file"
}

func (d *Doctor) fixMissingDir(issueID string) (bool, string) {
	// Extract directory name from issue ID
	dirName := issueID[12:] // Skip "dir-missing-"
	dirPath := filepath.Join(d.homeDir, ".asc", dirName)
	
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return false, fmt.Sprintf("Failed to create directory: %v", err)
	}
	return true, fmt.Sprintf("Created directory ~/.asc/%s", dirName)
}

// generateHealthSummary creates a summary of the diagnostic results
func (d *Doctor) generateHealthSummary(report *DiagnosticReport) string {
	if len(report.Issues) == 0 {
		return "✓ All checks passed - system is healthy"
	}
	
	criticalCount := 0
	highCount := 0
	mediumCount := 0
	lowCount := 0
	
	for _, issue := range report.Issues {
		switch issue.Severity {
		case SeverityCritical:
			criticalCount++
		case SeverityHigh:
			highCount++
		case SeverityMedium:
			mediumCount++
		case SeverityLow:
			lowCount++
		}
	}
	
	summary := fmt.Sprintf("Found %d issue(s): ", len(report.Issues))
	parts := []string{}
	if criticalCount > 0 {
		parts = append(parts, fmt.Sprintf("%d critical", criticalCount))
	}
	if highCount > 0 {
		parts = append(parts, fmt.Sprintf("%d high", highCount))
	}
	if mediumCount > 0 {
		parts = append(parts, fmt.Sprintf("%d medium", mediumCount))
	}
	if lowCount > 0 {
		parts = append(parts, fmt.Sprintf("%d low", lowCount))
	}
	
	for i, part := range parts {
		if i > 0 {
			summary += ", "
		}
		summary += part
	}
	
	return summary
}

// HasCriticalIssues returns true if there are any critical issues
func (r *DiagnosticReport) HasCriticalIssues() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

// ToJSON converts the report to JSON format
func (r *DiagnosticReport) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Format creates a human-readable report
func (r *DiagnosticReport) Format(verbose bool) string {
	output := "\n"
	output += "╔════════════════════════════════════════════════════════════════╗\n"
	output += "║              ASC DOCTOR - DIAGNOSTIC REPORT                    ║\n"
	output += "╚════════════════════════════════════════════════════════════════╝\n\n"
	
	output += fmt.Sprintf("Run at: %s\n", r.RunAt.Format("2006-01-02 15:04:05"))
	output += fmt.Sprintf("Status: %s\n\n", r.HealthSummary)
	
	if len(r.Issues) == 0 {
		output += "✓ No issues detected\n"
		return output
	}
	
	// Group issues by severity
	bySeverity := map[IssueSeverity][]Issue{
		SeverityCritical: {},
		SeverityHigh:     {},
		SeverityMedium:   {},
		SeverityLow:      {},
		SeverityInfo:     {},
	}
	
	for _, issue := range r.Issues {
		bySeverity[issue.Severity] = append(bySeverity[issue.Severity], issue)
	}
	
	// Display issues by severity
	severities := []IssueSeverity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}
	for _, severity := range severities {
		issues := bySeverity[severity]
		if len(issues) == 0 {
			continue
		}
		
		output += fmt.Sprintf("─── %s SEVERITY (%d) ───\n\n", severity, len(issues))
		
		for i, issue := range issues {
			icon := "●"
			switch severity {
			case SeverityCritical:
				icon = "✗"
			case SeverityHigh:
				icon = "⚠"
			case SeverityMedium:
				icon = "!"
			case SeverityLow:
				icon = "·"
			case SeverityInfo:
				icon = "ℹ"
			}
			
			output += fmt.Sprintf("%s %s\n", icon, issue.Title)
			output += fmt.Sprintf("  Category: %s\n", issue.Category)
			
			if verbose {
				output += fmt.Sprintf("  Description: %s\n", issue.Description)
				output += fmt.Sprintf("  Impact: %s\n", issue.Impact)
			}
			
			output += fmt.Sprintf("  Remediation: %s\n", issue.Remediation)
			
			if issue.AutoFixable {
				output += "  ✓ Auto-fixable with --fix flag\n"
			}
			
			if i < len(issues)-1 {
				output += "\n"
			}
		}
		output += "\n"
	}
	
	// Display fix results if any
	if len(r.FixesApplied) > 0 {
		output += "─── FIXES APPLIED ───\n\n"
		for _, fix := range r.FixesApplied {
			icon := "✓"
			if !fix.Success {
				icon = "✗"
			}
			output += fmt.Sprintf("%s %s: %s\n", icon, fix.IssueID, fix.Message)
		}
		output += "\n"
	}
	
	return output
}

// Helper functions
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
