package config

import (
	"fmt"
	"strings"
)

// ReloadManager handles configuration reload logic and agent lifecycle management
type ReloadManager struct {
	currentConfig  *Config
	processManager ProcessManager
	envVars        map[string]string // Environment variables (API keys, etc.)
}

// ProcessManager interface for managing agent processes
// This allows the reload manager to start/stop agents without circular dependencies
type ProcessManager interface {
	Start(name, command string, args []string, env []string) (int, error)
	Stop(pid int) error
	IsRunning(pid int) bool
	GetProcessInfo(name string) (ProcessInfoGetter, error)
}

// ProcessInfoGetter is a minimal interface for getting process information
// This avoids circular dependencies with the process package
type ProcessInfoGetter interface {
	GetPID() int
	GetName() string
	GetCommand() string
	GetArgs() []string
	GetEnv() map[string]string
}

// ProcessInfoAdapter adapts a concrete ProcessInfo to the interface
type ProcessInfoAdapter struct {
	Name    string
	PID     int
	Command string
	Args    []string
	Env     map[string]string
}

func (p *ProcessInfoAdapter) GetPID() int                     { return p.PID }
func (p *ProcessInfoAdapter) GetName() string                { return p.Name }
func (p *ProcessInfoAdapter) GetCommand() string             { return p.Command }
func (p *ProcessInfoAdapter) GetArgs() []string              { return p.Args }
func (p *ProcessInfoAdapter) GetEnv() map[string]string      { return p.Env }

// NewReloadManager creates a new reload manager
func NewReloadManager(currentConfig *Config, procManager ProcessManager, envVars map[string]string) *ReloadManager {
	return &ReloadManager{
		currentConfig: currentConfig,
		processManager: procManager,
		envVars:       envVars,
	}
}

// ReloadResult contains information about what changed during a reload
type ReloadResult struct {
	AgentsAdded   []string
	AgentsRemoved []string
	AgentsUpdated []string
	Errors        []error
}

// Reload compares the new configuration with the current one and applies changes
func (rm *ReloadManager) Reload(newConfig *Config) (*ReloadResult, error) {
	result := &ReloadResult{
		AgentsAdded:   []string{},
		AgentsRemoved: []string{},
		AgentsUpdated: []string{},
		Errors:        []error{},
	}

	// Find agents that were removed
	for oldName := range rm.currentConfig.Agents {
		if _, exists := newConfig.Agents[oldName]; !exists {
			result.AgentsRemoved = append(result.AgentsRemoved, oldName)
		}
	}

	// Find agents that were added or updated
	for newName, newAgent := range newConfig.Agents {
		oldAgent, exists := rm.currentConfig.Agents[newName]
		
		if !exists {
			// New agent
			result.AgentsAdded = append(result.AgentsAdded, newName)
		} else if rm.agentConfigChanged(oldAgent, newAgent) {
			// Agent configuration changed
			result.AgentsUpdated = append(result.AgentsUpdated, newName)
		}
	}

	// Apply changes: stop removed agents
	for _, agentName := range result.AgentsRemoved {
		if err := rm.stopAgent(agentName); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to stop agent %s: %w", agentName, err))
		}
	}

	// Apply changes: restart updated agents
	for _, agentName := range result.AgentsUpdated {
		if err := rm.stopAgent(agentName); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to stop agent %s for update: %w", agentName, err))
			continue
		}
		
		if err := rm.startAgent(agentName, newConfig.Agents[agentName], newConfig); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to restart agent %s: %w", agentName, err))
		}
	}

	// Apply changes: start new agents
	for _, agentName := range result.AgentsAdded {
		if err := rm.startAgent(agentName, newConfig.Agents[agentName], newConfig); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to start agent %s: %w", agentName, err))
		}
	}

	// Update current config
	rm.currentConfig = newConfig

	return result, nil
}

// agentConfigChanged checks if an agent's configuration has changed
func (rm *ReloadManager) agentConfigChanged(old, new AgentConfig) bool {
	// Check if command changed
	if old.Command != new.Command {
		return true
	}

	// Check if model changed
	if old.Model != new.Model {
		return true
	}

	// Check if phases changed
	if len(old.Phases) != len(new.Phases) {
		return true
	}

	// Check each phase
	oldPhases := make(map[string]bool)
	for _, phase := range old.Phases {
		oldPhases[phase] = true
	}

	for _, phase := range new.Phases {
		if !oldPhases[phase] {
			return true
		}
	}

	return false
}

// stopAgent stops a running agent
func (rm *ReloadManager) stopAgent(agentName string) error {
	// Get process info
	info, err := rm.processManager.GetProcessInfo(agentName)
	if err != nil {
		// Agent might not be running, which is fine
		return nil
	}

	// Check if running
	if !rm.processManager.IsRunning(info.GetPID()) {
		return nil
	}

	// Stop the process
	return rm.processManager.Stop(info.GetPID())
}

// startAgent starts a new agent with the given configuration
func (rm *ReloadManager) startAgent(agentName string, agentConfig AgentConfig, config *Config) error {
	// Parse command and args
	cmdParts := strings.Fields(agentConfig.Command)
	if len(cmdParts) == 0 {
		return fmt.Errorf("empty command")
	}

	command := cmdParts[0]
	args := cmdParts[1:]

	// Build environment variables
	env := rm.buildAgentEnv(agentName, agentConfig, config)

	// Start the process
	_, err := rm.processManager.Start(agentName, command, args, env)
	return err
}

// buildAgentEnv builds the environment variables for an agent
func (rm *ReloadManager) buildAgentEnv(agentName string, agentConfig AgentConfig, config *Config) []string {
	env := []string{
		fmt.Sprintf("AGENT_NAME=%s", agentName),
		fmt.Sprintf("AGENT_MODEL=%s", agentConfig.Model),
		fmt.Sprintf("AGENT_PHASES=%s", strings.Join(agentConfig.Phases, ",")),
		fmt.Sprintf("MCP_MAIL_URL=%s", config.Services.MCPAgentMail.URL),
		fmt.Sprintf("BEADS_DB_PATH=%s", config.Core.BeadsDBPath),
	}

	// Add API keys from environment
	for key, value := range rm.envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

// GetCurrentConfig returns the current configuration
func (rm *ReloadManager) GetCurrentConfig() *Config {
	return rm.currentConfig
}
