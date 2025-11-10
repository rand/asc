package config

// Config represents the complete asc configuration
type Config struct {
	Core     CoreConfig                `mapstructure:"core"`
	Services ServicesConfig            `mapstructure:"services"`
	Agents   map[string]AgentConfig    `mapstructure:"agent"`
}

// CoreConfig contains core system configuration
type CoreConfig struct {
	BeadsDBPath string `mapstructure:"beads_db_path"`
}

// ServicesConfig contains configuration for external services
type ServicesConfig struct {
	MCPAgentMail MCPConfig `mapstructure:"mcp_agent_mail"`
}

// MCPConfig contains MCP agent mail server configuration
type MCPConfig struct {
	StartCommand string `mapstructure:"start_command"`
	URL          string `mapstructure:"url"`
}

// AgentConfig contains configuration for a single agent
type AgentConfig struct {
	Command string   `mapstructure:"command"`
	Model   string   `mapstructure:"model"`
	Phases  []string `mapstructure:"phases"`
}
