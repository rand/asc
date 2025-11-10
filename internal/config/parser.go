package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// DefaultConfigPath returns the default path for the asc.toml file
func DefaultConfigPath() string {
	return "asc.toml"
}

// Load reads and parses the configuration file from the given path
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

	for name, agent := range cfg.Agents {
		if agent.Command == "" {
			return fmt.Errorf("agent '%s': command is required", name)
		}
		if agent.Model == "" {
			return fmt.Errorf("agent '%s': model is required", name)
		}
		if len(agent.Phases) == 0 {
			return fmt.Errorf("agent '%s': at least one phase is required", name)
		}
	}

	return nil
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
