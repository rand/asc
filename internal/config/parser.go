package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// DefaultConfigPath returns the default path for the asc.toml configuration file.
// This is typically "asc.toml" in the current working directory.
func DefaultConfigPath() string {
	return "asc.toml"
}

// Load reads and parses the configuration file from the given path.
// It validates required fields, applies defaults, and expands paths.
// Returns an error if the file doesn't exist, has invalid syntax, or fails validation.
//
// Example:
//
//	cfg, err := config.Load("asc.toml")
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
func Load(configPath string) (*Config, error) {
	// Set up viper
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("toml")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	applyDefaults(&cfg)

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// applyDefaults sets default values for optional configuration fields
func applyDefaults(cfg *Config) {
	// Default beads DB path
	if cfg.Core.BeadsDBPath == "" {
		cfg.Core.BeadsDBPath = "./project-repo"
	}

	// Default auto-recovery to true (enabled by default)
	// Note: In TOML, if the field is not specified, it defaults to false (zero value for bool)
	// We want it enabled by default, so we need to check if it was explicitly set
	// Since we can't distinguish between "not set" and "set to false" with a bool,
	// we'll enable it by default in the monitor initialization instead
	// For now, we'll document that auto_recovery defaults to true if not specified

	// Default MCP agent mail URL
	if cfg.Services.MCPAgentMail.URL == "" {
		cfg.Services.MCPAgentMail.URL = "http://localhost:8765"
	}

	// Default MCP agent mail start command
	if cfg.Services.MCPAgentMail.StartCommand == "" {
		cfg.Services.MCPAgentMail.StartCommand = "python -m mcp_agent_mail.server"
	}
}

// validate checks that all required configuration fields are present and valid
func validate(cfg *Config) error {
	// Validate beads DB path
	if cfg.Core.BeadsDBPath == "" {
		return fmt.Errorf("core.beads_db_path is required")
	}

	// Expand and validate beads DB path
	beadsPath, err := expandPath(cfg.Core.BeadsDBPath)
	if err != nil {
		return fmt.Errorf("invalid beads_db_path: %w", err)
	}
	cfg.Core.BeadsDBPath = beadsPath

	// Validate MCP configuration
	if cfg.Services.MCPAgentMail.StartCommand == "" {
		return fmt.Errorf("services.mcp_agent_mail.start_command is required")
	}
	if cfg.Services.MCPAgentMail.URL == "" {
		return fmt.Errorf("services.mcp_agent_mail.url is required")
	}

	// Validate agents
	if len(cfg.Agents) == 0 {
		return fmt.Errorf("at least one agent must be defined")
	}

	// Check for duplicate agent names (case-insensitive)
	agentNames := make(map[string]string)
	for name := range cfg.Agents {
		lowerName := strings.ToLower(name)
		if existingName, exists := agentNames[lowerName]; exists {
			return fmt.Errorf("duplicate agent name detected: '%s' and '%s' (agent names are case-insensitive)", existingName, name)
		}
		agentNames[lowerName] = name
	}

	// Validate each agent
	for name, agent := range cfg.Agents {
		if err := validateAgent(name, agent); err != nil {
			return err
		}
	}

	return nil
}

// validateAgent validates a single agent configuration with detailed error messages and suggestions
func validateAgent(name string, agent AgentConfig) error {
	// Validate command is present
	if agent.Command == "" {
		return fmt.Errorf("agent '%s': command is required", name)
	}

	// Validate command exists in PATH
	cmdParts := strings.Fields(agent.Command)
	if len(cmdParts) == 0 {
		return fmt.Errorf("agent '%s': command is empty", name)
	}
	
	cmdName := cmdParts[0]
	if _, err := exec.LookPath(cmdName); err != nil {
		return fmt.Errorf("agent '%s': command '%s' not found in PATH\n  Suggestion: Install the required binary or check your PATH environment variable", name, cmdName)
	}

	// Validate model is present
	if agent.Model == "" {
		return fmt.Errorf("agent '%s': model is required", name)
	}

	// Validate model is supported
	supportedModels := []string{"claude", "gemini", "gpt-4", "codex", "openai"}
	if !isValidModel(agent.Model) {
		return fmt.Errorf("agent '%s': unsupported model '%s'\n  Supported models: %s\n  Suggestion: Use one of the supported models or check for typos", 
			name, agent.Model, strings.Join(supportedModels, ", "))
	}

	// Validate phases are present
	if len(agent.Phases) == 0 {
		return fmt.Errorf("agent '%s': at least one phase is required", name)
	}

	// Validate each phase
	validPhases := []string{
		"planning", "design", "implementation", "coding", 
		"testing", "review", "refactor", "documentation",
		"debugging", "optimization", "deployment",
	}
	
	for _, phase := range agent.Phases {
		if !isValidPhase(phase) {
			suggestion := findClosestPhase(phase, validPhases)
			errMsg := fmt.Errorf("agent '%s': invalid phase '%s'\n  Valid phases: %s", 
				name, phase, strings.Join(validPhases, ", "))
			if suggestion != "" {
				errMsg = fmt.Errorf("agent '%s': invalid phase '%s'\n  Valid phases: %s\n  Suggestion: Did you mean '%s'?", 
					name, phase, strings.Join(validPhases, ", "), suggestion)
			}
			return errMsg
		}
	}

	return nil
}

// isValidModel checks if the model name is supported
func isValidModel(model string) bool {
	supportedModels := map[string]bool{
		"claude":  true,
		"gemini":  true,
		"gpt-4":   true,
		"codex":   true,
		"openai":  true,
	}
	return supportedModels[strings.ToLower(model)]
}

// isValidPhase checks if the phase name is valid
func isValidPhase(phase string) bool {
	validPhases := map[string]bool{
		"planning":       true,
		"design":         true,
		"implementation": true,
		"coding":         true,
		"testing":        true,
		"review":         true,
		"refactor":       true,
		"documentation":  true,
		"debugging":      true,
		"optimization":   true,
		"deployment":     true,
	}
	return validPhases[strings.ToLower(phase)]
}

// findClosestPhase finds the closest matching phase using simple string similarity
func findClosestPhase(input string, validPhases []string) string {
	input = strings.ToLower(input)
	
	// Check for substring matches first
	for _, valid := range validPhases {
		if strings.Contains(valid, input) || strings.Contains(input, valid) {
			return valid
		}
	}
	
	// Check for common typos and abbreviations
	commonMappings := map[string]string{
		"plan":    "planning",
		"impl":    "implementation",
		"code":    "coding",
		"test":    "testing",
		"doc":     "documentation",
		"docs":    "documentation",
		"debug":   "debugging",
		"opt":     "optimization",
		"deploy":  "deployment",
		"refact":  "refactor",
	}
	
	if suggestion, exists := commonMappings[input]; exists {
		return suggestion
	}
	
	return ""
}

// expandPath expands ~ and environment variables in a path
func expandPath(path string) (string, error) {
	// Expand ~ to home directory
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return absPath, nil
}

// ValidationWarning represents a non-fatal configuration issue
type ValidationWarning struct {
	Message    string
	Suggestion string
}

// ValidateWithWarnings performs validation and returns both errors and warnings
// This allows for more detailed feedback without failing the configuration load
func ValidateWithWarnings(cfg *Config) ([]ValidationWarning, error) {
	warnings := []ValidationWarning{}
	
	// Run standard validation first
	if err := validate(cfg); err != nil {
		return warnings, err
	}
	
	// Check for common configuration issues that are warnings, not errors
	
	// Warn if all agents use the same model
	modelCounts := make(map[string]int)
	for _, agent := range cfg.Agents {
		modelCounts[agent.Model]++
	}
	if len(modelCounts) == 1 && len(cfg.Agents) > 1 {
		for model := range modelCounts {
			warnings = append(warnings, ValidationWarning{
				Message:    fmt.Sprintf("All agents are using the same model (%s)", model),
				Suggestion: "Consider using different models for different agents to leverage their unique strengths",
			})
		}
	}
	
	// Warn if multiple agents have overlapping phases
	phaseAgents := make(map[string][]string)
	for name, agent := range cfg.Agents {
		for _, phase := range agent.Phases {
			phaseAgents[phase] = append(phaseAgents[phase], name)
		}
	}
	for phase, agents := range phaseAgents {
		if len(agents) > 3 {
			warnings = append(warnings, ValidationWarning{
				Message:    fmt.Sprintf("Phase '%s' has %d agents assigned: %s", phase, len(agents), strings.Join(agents, ", ")),
				Suggestion: "Having too many agents on the same phase may cause resource contention",
			})
		}
	}
	
	// Warn if no agent covers certain critical phases
	criticalPhases := []string{"planning", "implementation", "testing"}
	for _, critical := range criticalPhases {
		if agents, exists := phaseAgents[critical]; !exists || len(agents) == 0 {
			warnings = append(warnings, ValidationWarning{
				Message:    fmt.Sprintf("No agent is assigned to the '%s' phase", critical),
				Suggestion: fmt.Sprintf("Consider adding an agent for the '%s' phase for better workflow coverage", critical),
			})
		}
	}
	
	return warnings, nil
}
