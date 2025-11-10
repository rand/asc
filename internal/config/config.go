// Package config provides configuration management for the Agent Stack Controller.
// It handles parsing of TOML configuration files and environment variables,
// with support for validation and default values.
//
// Example usage:
//
//	cfg, err := config.Load("asc.toml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := config.LoadAndValidateEnv(".env"); err != nil {
//	    log.Fatal(err)
//	}
package config

// Config represents the complete asc configuration loaded from asc.toml.
// It contains core settings, service configurations, and agent definitions.
type Config struct {
	Core     CoreConfig                `mapstructure:"core"`
	Services ServicesConfig            `mapstructure:"services"`
	Agents   map[string]AgentConfig    `mapstructure:"agent"`
}

// CoreConfig contains core system configuration including paths to
// essential components like the beads task database.
type CoreConfig struct {
	BeadsDBPath     string `mapstructure:"beads_db_path"`     // Path to the beads task database repository
	AutoRecovery    *bool  `mapstructure:"auto_recovery"`     // Enable automatic agent recovery (default: true if nil)
}

// ServicesConfig contains configuration for external services that
// the agent stack depends on, such as the MCP agent mail server.
type ServicesConfig struct {
	MCPAgentMail MCPConfig `mapstructure:"mcp_agent_mail"` // MCP agent mail server configuration
}

// MCPConfig contains MCP agent mail server configuration including
// the command to start the server and its HTTP endpoint URL.
type MCPConfig struct {
	StartCommand string `mapstructure:"start_command"` // Command to start the MCP server (e.g., "python -m mcp_agent_mail.server")
	URL          string `mapstructure:"url"`           // HTTP endpoint URL (e.g., "http://localhost:8765")
}

// AgentConfig contains configuration for a single agent including
// the command to execute, the LLM model to use, and the workflow phases it handles.
type AgentConfig struct {
	Command string   `mapstructure:"command"` // Command to execute the agent (e.g., "python agent_adapter.py")
	Model   string   `mapstructure:"model"`   // LLM model: "claude", "gemini", "gpt-4", "codex"
	Phases  []string `mapstructure:"phases"`  // Workflow phases: "planning", "implementation", "testing", etc.
}
