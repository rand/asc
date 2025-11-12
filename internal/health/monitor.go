// Package health provides comprehensive health monitoring for agents in the
// Agent Stack Controller. It tracks agent heartbeats, detects unresponsive,
// crashed, and stuck agents, and logs all health issues.
//
// Example usage:
//
//	monitor := health.NewMonitor(mcpClient, procManager, config)
//	monitor.Start()
//	defer monitor.Stop()
//
//	// Get current health status
//	issues := monitor.GetHealthIssues()
//	for _, issue := range issues {
//	    fmt.Printf("Agent %s: %s\n", issue.AgentName, issue.Description)
//	}
package health

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rand/asc/internal/config"
	"github.com/rand/asc/internal/logger"
	"github.com/rand/asc/internal/mcp"
	"github.com/rand/asc/internal/process"
)

// HealthIssueType represents the type of health issue detected
type HealthIssueType string

const (
	IssueUnresponsive HealthIssueType = "unresponsive" // No heartbeat for 2 minutes
	IssueCrashed      HealthIssueType = "crashed"      // Process exited unexpectedly
	IssueStuck        HealthIssueType = "stuck"        // Working on same task for >30 minutes
)

// HealthIssue represents a detected health problem with an agent
type HealthIssue struct {
	AgentName   string
	Type        HealthIssueType
	Description string
	DetectedAt  time.Time
	Severity    string // "warning" or "critical"
}

// AgentHealthState tracks the health state of a single agent
type AgentHealthState struct {
	Name            string
	LastHeartbeat   time.Time
	LastTask        string
	TaskStartTime   time.Time
	ProcessRunning  bool
	LastCheckTime   time.Time
	ConsecutiveFails int
}

// RecoveryAction represents an automatic recovery action taken
type RecoveryAction struct {
	AgentName   string
	Action      string // "restart", "release_leases"
	Reason      string
	Timestamp   time.Time
	Success     bool
	ErrorMsg    string
}

// Monitor provides comprehensive health monitoring for agents
type Monitor struct {
	mcpClient   mcp.MCPClient
	procManager process.ProcessManager
	config      config.Config
	
	// Health state tracking
	mu           sync.RWMutex
	agentStates  map[string]*AgentHealthState
	healthIssues []HealthIssue
	
	// Recovery tracking
	recoveryActions []RecoveryAction
	recoveryStats   map[string]*RecoveryStats
	
	// Health check configuration
	checkInterval       time.Duration
	unresponsiveTimeout time.Duration
	stuckTaskTimeout    time.Duration
	autoRecoveryEnabled bool
	
	// Control
	stopChan chan struct{}
	wg       sync.WaitGroup
	
	// Health log
	healthLogger *logger.Logger
}

// RecoveryStats tracks recovery success rate per agent
type RecoveryStats struct {
	TotalAttempts   int
	SuccessfulCount int
	LastAttempt     time.Time
	BackoffUntil    time.Time
}

// NewMonitor creates a new health monitor with the given clients and configuration
func NewMonitor(mcpClient mcp.MCPClient, procManager process.ProcessManager, cfg config.Config) (*Monitor, error) {
	// Create health log file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	logDir := filepath.Join(homeDir, ".asc", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	healthLogPath := filepath.Join(logDir, "health.log")
	healthLogger, err := logger.NewLogger(healthLogPath, 10*1024*1024, 5, logger.INFO)
	if err != nil {
		return nil, fmt.Errorf("failed to create health logger: %w", err)
	}
	
	m := &Monitor{
		mcpClient:           mcpClient,
		procManager:         procManager,
		config:              cfg,
		agentStates:         make(map[string]*AgentHealthState),
		healthIssues:        []HealthIssue{},
		recoveryActions:     []RecoveryAction{},
		recoveryStats:       make(map[string]*RecoveryStats),
		checkInterval:       30 * time.Second,
		unresponsiveTimeout: 2 * time.Minute,
		stuckTaskTimeout:    30 * time.Minute,
		autoRecoveryEnabled: true, // Enabled by default, can be disabled via SetAutoRecovery()
		stopChan:            make(chan struct{}),
		healthLogger:        healthLogger,
	}
	
	// Initialize agent states from config
	for agentName := range cfg.Agents {
		m.agentStates[agentName] = &AgentHealthState{
			Name:           agentName,
			LastHeartbeat:  time.Time{},
			LastTask:       "",
			TaskStartTime:  time.Time{},
			ProcessRunning: false,
			LastCheckTime:  time.Now(),
		}
	}
	
	return m, nil
}

// logHealth logs to the health logger if it's not nil
func (m *Monitor) logHealth(level logger.LogLevel, format string, args ...interface{}) {
	if m.healthLogger != nil {
		switch level {
		case logger.INFO:
			m.healthLogger.Info(format, args...)
		case logger.WARN:
			m.healthLogger.Warn(format, args...)
		case logger.ERROR:
			m.healthLogger.Error(format, args...)
		case logger.DEBUG:
			m.healthLogger.Debug(format, args...)
		}
	}
}

// Start begins the health monitoring loop
func (m *Monitor) Start() {
	m.wg.Add(1)
	go m.monitorLoop()
	logger.Info("Health monitor started")
	m.logHealth(logger.INFO, "Health monitor started")
}

// Stop stops the health monitoring loop
func (m *Monitor) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	if m.healthLogger != nil {
		m.healthLogger.Close()
	}
	logger.Info("Health monitor stopped")
}

// monitorLoop runs the periodic health check
func (m *Monitor) monitorLoop() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()
	
	// Run initial check immediately
	m.performHealthCheck()
	
	for {
		select {
		case <-ticker.C:
			m.performHealthCheck()
		case <-m.stopChan:
			return
		}
	}
}

// performHealthCheck executes a complete health check on all agents
func (m *Monitor) performHealthCheck() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	newIssues := []HealthIssue{}
	
	// Get current agent statuses from MCP
	statuses, err := m.mcpClient.GetAllAgentStatuses(m.unresponsiveTimeout)
	if err != nil {
		logger.Warn("Failed to get agent statuses during health check: %v", err)
		m.logHealth(logger.WARN, "Failed to get agent statuses: %v", err)
		return
	}
	
	// Create a map for quick lookup
	statusMap := make(map[string]mcp.AgentStatus)
	for _, status := range statuses {
		statusMap[status.Name] = status
	}
	
	// Check each configured agent
	for agentName, state := range m.agentStates {
		// Get process info
		procInfo, err := m.procManager.GetProcessInfo(agentName)
		processRunning := err == nil && m.procManager.IsRunning(procInfo.PID)
		state.ProcessRunning = processRunning
		state.LastCheckTime = now
		
		// Get MCP status
		mcpStatus, hasMCPStatus := statusMap[agentName]
		
		// Check for crashed agent (process not running but should be)
		if !processRunning {
			issue := HealthIssue{
				AgentName:   agentName,
				Type:        IssueCrashed,
				Description: "Process has exited unexpectedly",
				DetectedAt:  now,
				Severity:    "critical",
			}
			newIssues = append(newIssues, issue)
			m.logHealth(logger.ERROR, "Agent %s crashed: process not running", agentName)
			continue
		}
		
		// Check for unresponsive agent (no heartbeat)
		if hasMCPStatus {
			state.LastHeartbeat = mcpStatus.LastSeen
			
			if mcpStatus.State == mcp.StateOffline || now.Sub(mcpStatus.LastSeen) > m.unresponsiveTimeout {
				issue := HealthIssue{
					AgentName:   agentName,
					Type:        IssueUnresponsive,
					Description: fmt.Sprintf("No heartbeat for %v", now.Sub(mcpStatus.LastSeen).Round(time.Second)),
					DetectedAt:  now,
					Severity:    "critical",
				}
				newIssues = append(newIssues, issue)
				m.logHealth(logger.WARN, "Agent %s unresponsive: no heartbeat for %v", agentName, now.Sub(mcpStatus.LastSeen).Round(time.Second))
			}
			
			// Check for stuck agent (working on same task too long)
			if mcpStatus.State == mcp.StateWorking && mcpStatus.CurrentTask != "" {
				// Track task changes
				if state.LastTask != mcpStatus.CurrentTask {
					state.LastTask = mcpStatus.CurrentTask
					state.TaskStartTime = now
				} else if !state.TaskStartTime.IsZero() {
					taskDuration := now.Sub(state.TaskStartTime)
					if taskDuration > m.stuckTaskTimeout {
						issue := HealthIssue{
							AgentName:   agentName,
							Type:        IssueStuck,
							Description: fmt.Sprintf("Working on task %s for %v", mcpStatus.CurrentTask, taskDuration.Round(time.Minute)),
							DetectedAt:  now,
							Severity:    "warning",
						}
						newIssues = append(newIssues, issue)
						m.logHealth(logger.WARN, "Agent %s stuck: working on task %s for %v", agentName, mcpStatus.CurrentTask, taskDuration.Round(time.Minute))
					}
				}
			} else {
				// Agent is idle or in another state, reset task tracking
				state.LastTask = ""
				state.TaskStartTime = time.Time{}
			}
		} else {
			// No MCP status found - agent might not have sent heartbeat yet
			if !state.LastHeartbeat.IsZero() && now.Sub(state.LastHeartbeat) > m.unresponsiveTimeout {
				issue := HealthIssue{
					AgentName:   agentName,
					Type:        IssueUnresponsive,
					Description: fmt.Sprintf("No heartbeat data available for %v", now.Sub(state.LastHeartbeat).Round(time.Second)),
					DetectedAt:  now,
					Severity:    "critical",
				}
				newIssues = append(newIssues, issue)
				m.logHealth(logger.WARN, "Agent %s unresponsive: no MCP status available", agentName)
			}
		}
	}
	
	// Update health issues
	m.healthIssues = newIssues
	
	// Log summary if there are issues
	if len(newIssues) > 0 {
		logger.Warn("Health check found %d issue(s)", len(newIssues))
		
		// Attempt automatic recovery if enabled
		m.attemptRecovery()
	} else {
		logger.Debug("Health check: all agents healthy")
	}
}

// GetHealthIssues returns the current list of health issues
func (m *Monitor) GetHealthIssues() []HealthIssue {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	issues := make([]HealthIssue, len(m.healthIssues))
	copy(issues, m.healthIssues)
	return issues
}

// GetAgentState returns the health state of a specific agent
func (m *Monitor) GetAgentState(agentName string) (*AgentHealthState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	state, exists := m.agentStates[agentName]
	if !exists {
		return nil, false
	}
	
	// Return a copy
	stateCopy := *state
	return &stateCopy, true
}

// GetAllAgentStates returns the health state of all agents
func (m *Monitor) GetAllAgentStates() map[string]*AgentHealthState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy
	states := make(map[string]*AgentHealthState)
	for name, state := range m.agentStates {
		stateCopy := *state
		states[name] = &stateCopy
	}
	return states
}

// IsHealthy returns true if there are no critical health issues
func (m *Monitor) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, issue := range m.healthIssues {
		if issue.Severity == "critical" {
			return false
		}
	}
	return true
}

// GetHealthSummary returns a human-readable summary of health status
func (m *Monitor) GetHealthSummary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.healthIssues) == 0 {
		return "All agents healthy"
	}
	
	criticalCount := 0
	warningCount := 0
	for _, issue := range m.healthIssues {
		if issue.Severity == "critical" {
			criticalCount++
		} else {
			warningCount++
		}
	}
	
	return fmt.Sprintf("%d critical, %d warning", criticalCount, warningCount)
}

// SetAutoRecovery enables or disables automatic recovery
func (m *Monitor) SetAutoRecovery(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.autoRecoveryEnabled = enabled
	status := map[bool]string{true: "enabled", false: "disabled"}[enabled]
	logger.Info("Auto-recovery %s", status)
	m.logHealth(logger.INFO, "Auto-recovery %s", status)
}

// IsAutoRecoveryEnabled returns whether automatic recovery is enabled
func (m *Monitor) IsAutoRecoveryEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.autoRecoveryEnabled
}

// GetRecoveryActions returns the list of recovery actions taken
func (m *Monitor) GetRecoveryActions() []RecoveryAction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy
	actions := make([]RecoveryAction, len(m.recoveryActions))
	copy(actions, m.recoveryActions)
	return actions
}

// GetRecoveryStats returns recovery statistics for an agent
func (m *Monitor) GetRecoveryStats(agentName string) *RecoveryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats, exists := m.recoveryStats[agentName]
	if !exists {
		return nil
	}
	
	// Return a copy
	statsCopy := *stats
	return &statsCopy
}

// attemptRecovery attempts to recover from health issues
func (m *Monitor) attemptRecovery() {
	if !m.autoRecoveryEnabled {
		return
	}
	
	now := time.Now()
	
	for _, issue := range m.healthIssues {
		// Check if we should attempt recovery for this agent
		stats := m.getOrCreateRecoveryStats(issue.AgentName)
		
		// Check if we're in backoff period
		if now.Before(stats.BackoffUntil) {
			logger.Debug("Skipping recovery for %s: in backoff period", issue.AgentName)
			continue
		}
		
		// Attempt recovery based on issue type
		switch issue.Type {
		case IssueCrashed:
			m.recoverCrashedAgent(issue.AgentName, stats)
		case IssueStuck:
			m.recoverStuckAgent(issue.AgentName, stats)
		case IssueUnresponsive:
			// For unresponsive agents, try restarting if process is still running
			m.recoverUnresponsiveAgent(issue.AgentName, stats)
		}
	}
}

// getOrCreateRecoveryStats gets or creates recovery stats for an agent
func (m *Monitor) getOrCreateRecoveryStats(agentName string) *RecoveryStats {
	stats, exists := m.recoveryStats[agentName]
	if !exists {
		stats = &RecoveryStats{
			TotalAttempts:   0,
			SuccessfulCount: 0,
			LastAttempt:     time.Time{},
			BackoffUntil:    time.Time{},
		}
		m.recoveryStats[agentName] = stats
	}
	return stats
}

// recoverCrashedAgent attempts to restart a crashed agent
func (m *Monitor) recoverCrashedAgent(agentName string, stats *RecoveryStats) {
	logger.Info("Attempting to restart crashed agent: %s", agentName)
	m.logHealth(logger.INFO, "Attempting to restart crashed agent: %s", agentName)
	
	// Get agent config
	agentConfig, exists := m.config.Agents[agentName]
	if !exists {
		m.recordRecoveryAction(agentName, "restart", "crashed", false, "agent not found in config")
		return
	}
	
	// Build environment variables
	env := m.buildAgentEnv(agentName, agentConfig)
	
	// Start the agent process
	pid, err := m.procManager.Start(agentName, agentConfig.Command, []string{}, env)
	if err != nil {
		m.recordRecoveryAction(agentName, "restart", "crashed", false, err.Error())
		m.updateRecoveryStats(stats, false)
		return
	}
	
	logger.Info("Successfully restarted agent %s with PID %d", agentName, pid)
	m.logHealth(logger.INFO, "Successfully restarted agent %s with PID %d", agentName, pid)
	m.recordRecoveryAction(agentName, "restart", "crashed", true, "")
	m.updateRecoveryStats(stats, true)
}

// recoverStuckAgent attempts to recover a stuck agent by releasing leases
func (m *Monitor) recoverStuckAgent(agentName string, stats *RecoveryStats) {
	logger.Info("Attempting to recover stuck agent: %s", agentName)
	m.logHealth(logger.INFO, "Attempting to recover stuck agent: %s", agentName)
	
	// Try to release file leases via MCP
	err := m.mcpClient.ReleaseAgentLeases(agentName)
	if err != nil {
		logger.Error("Failed to release leases for stuck agent %s: %v", agentName, err)
		m.logHealth(logger.ERROR, "Failed to release leases for stuck agent %s: %v", agentName, err)
		m.recordRecoveryAction(agentName, "release_leases", "stuck", false, err.Error())
		m.updateRecoveryStats(stats, false)
		return
	}
	
	logger.Info("Successfully released leases for stuck agent %s", agentName)
	m.logHealth(logger.INFO, "Successfully released leases for stuck agent %s", agentName)
	m.recordRecoveryAction(agentName, "release_leases", "stuck", true, "")
	m.updateRecoveryStats(stats, true)
}

// recoverUnresponsiveAgent attempts to restart an unresponsive agent
func (m *Monitor) recoverUnresponsiveAgent(agentName string, stats *RecoveryStats) {
	logger.Info("Attempting to recover unresponsive agent: %s", agentName)
	m.logHealth(logger.INFO, "Attempting to recover unresponsive agent: %s", agentName)
	
	// Get process info
	procInfo, err := m.procManager.GetProcessInfo(agentName)
	if err != nil {
		// Process not found, treat as crashed
		m.recoverCrashedAgent(agentName, stats)
		return
	}
	
	// Stop the unresponsive process
	if err := m.procManager.Stop(procInfo.PID); err != nil {
		m.recordRecoveryAction(agentName, "restart", "unresponsive", false, fmt.Sprintf("failed to stop: %v", err))
		m.updateRecoveryStats(stats, false)
		return
	}
	
	// Wait a moment for cleanup
	time.Sleep(1 * time.Second)
	
	// Restart the agent
	m.recoverCrashedAgent(agentName, stats)
}

// buildAgentEnv builds environment variables for an agent
func (m *Monitor) buildAgentEnv(agentName string, agentConfig config.AgentConfig) []string {
	env := []string{
		fmt.Sprintf("AGENT_NAME=%s", agentName),
		fmt.Sprintf("AGENT_MODEL=%s", agentConfig.Model),
		fmt.Sprintf("AGENT_PHASES=%s", joinPhases(agentConfig.Phases)),
		fmt.Sprintf("MCP_MAIL_URL=%s", m.config.Services.MCPAgentMail.URL),
		fmt.Sprintf("BEADS_DB_PATH=%s", m.config.Core.BeadsDBPath),
	}
	
	// Add API keys from environment
	if apiKey := os.Getenv("CLAUDE_API_KEY"); apiKey != "" {
		env = append(env, fmt.Sprintf("CLAUDE_API_KEY=%s", apiKey))
	}
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		env = append(env, fmt.Sprintf("OPENAI_API_KEY=%s", apiKey))
	}
	if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		env = append(env, fmt.Sprintf("GOOGLE_API_KEY=%s", apiKey))
	}
	
	return env
}

// joinPhases joins phase names with commas
func joinPhases(phases []string) string {
	result := ""
	for i, phase := range phases {
		if i > 0 {
			result += ","
		}
		result += phase
	}
	return result
}

// recordRecoveryAction records a recovery action
func (m *Monitor) recordRecoveryAction(agentName, action, reason string, success bool, errorMsg string) {
	recoveryAction := RecoveryAction{
		AgentName: agentName,
		Action:    action,
		Reason:    reason,
		Timestamp: time.Now(),
		Success:   success,
		ErrorMsg:  errorMsg,
	}
	
	m.recoveryActions = append(m.recoveryActions, recoveryAction)
	
	// Limit recovery action history to last 100 actions
	if len(m.recoveryActions) > 100 {
		m.recoveryActions = m.recoveryActions[len(m.recoveryActions)-100:]
	}
	
	// Log to health log
	if success {
		m.logHealth(logger.INFO, "Recovery action succeeded: %s - %s for %s", agentName, action, reason)
	} else {
		m.logHealth(logger.ERROR, "Recovery action failed: %s - %s for %s: %s", agentName, action, reason, errorMsg)
	}
}

// updateRecoveryStats updates recovery statistics with exponential backoff
func (m *Monitor) updateRecoveryStats(stats *RecoveryStats, success bool) {
	stats.TotalAttempts++
	stats.LastAttempt = time.Now()
	
	if success {
		stats.SuccessfulCount++
		// Reset backoff on success
		stats.BackoffUntil = time.Time{}
	} else {
		// Calculate exponential backoff: 1min, 2min, 4min, 8min, max 15min
		backoffMinutes := 1 << uint(stats.TotalAttempts-stats.SuccessfulCount-1)
		if backoffMinutes > 15 {
			backoffMinutes = 15
		}
		stats.BackoffUntil = time.Now().Add(time.Duration(backoffMinutes) * time.Minute)
		logger.Warn("Recovery failed, backing off for %d minutes", backoffMinutes)
	}
}

// GetRecoverySuccessRate returns the success rate for an agent's recovery attempts
func (m *Monitor) GetRecoverySuccessRate(agentName string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats, exists := m.recoveryStats[agentName]
	if !exists || stats.TotalAttempts == 0 {
		return 0.0
	}
	
	return float64(stats.SuccessfulCount) / float64(stats.TotalAttempts)
}
