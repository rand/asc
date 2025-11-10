package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yourusername/asc/internal/mcp"
)

// refreshDataCmd returns a command that refreshes all data sources
func refreshDataCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		// Perform data refresh
		if err := m.refreshData(); err != nil {
			return refreshDataMsg{err: err}
		}
		return refreshDataMsg{err: nil}
	}
}

// refreshData fetches fresh data from all sources
func (m *Model) refreshData() error {
	// Track when this refresh started
	refreshTime := time.Now()

	// Fetch agent statuses from MCP client
	// Use 30 second offline threshold
	var agents []mcp.AgentStatus
	var err error
	
	// Check if the client supports GetAllAgentStatuses
	if httpClient, ok := m.mcpClient.(*mcp.HTTPClient); ok {
		agents, err = httpClient.GetAllAgentStatuses(30 * time.Second)
	} else {
		// Fallback: build agent list from config and query each individually
		agents = make([]mcp.AgentStatus, 0, len(m.config.Agents))
		for agentName := range m.config.Agents {
			status, statusErr := m.mcpClient.GetAgentStatus(agentName)
			if statusErr != nil {
				// Agent not found or error - mark as offline
				agents = append(agents, mcp.AgentStatus{
					Name:  agentName,
					State: mcp.StateOffline,
				})
			} else {
				agents = append(agents, status)
			}
		}
	}
	
	if err != nil {
		// Don't fail completely on agent status errors - just log and continue
		// This allows the TUI to remain functional even if MCP is temporarily unavailable
		m.err = err
	} else {
		m.agents = agents
		m.err = nil
	}

	// Fetch tasks from beads client with statuses "open" and "in_progress"
	tasks, err := m.beadsClient.GetTasks([]string{"open", "in_progress"})
	if err != nil {
		// Don't fail completely on task fetch errors
		m.err = err
	} else {
		m.tasks = tasks
	}

	// Fetch messages from MCP client since last refresh
	messages, err := m.mcpClient.GetMessages(m.lastRefresh)
	if err != nil {
		// Don't fail completely on message fetch errors
		m.err = err
	} else {
		// Append new messages to existing messages
		m.messages = append(m.messages, messages...)
		
		// Limit message buffer to last 100 messages
		if len(m.messages) > 100 {
			m.messages = m.messages[len(m.messages)-100:]
		}
	}

	// Update last refresh time
	m.lastRefresh = refreshTime

	return nil
}
