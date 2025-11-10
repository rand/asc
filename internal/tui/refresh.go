package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yourusername/asc/internal/mcp"
)

// refreshDataCmd returns a command that refreshes all data sources
// This is used for initial load and manual refresh
func refreshDataCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		// Perform data refresh
		if err := m.refreshData(); err != nil {
			return refreshDataMsg{err: err}
		}
		return refreshDataMsg{err: nil}
	}
}

// refreshBeadsCmd returns a command that refreshes only beads data
// This is used for periodic polling since beads is git-based and cannot be real-time
func refreshBeadsCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		// Perform beads refresh only
		if err := m.refreshBeadsData(); err != nil {
			return refreshDataMsg{err: err}
		}
		return refreshDataMsg{err: nil}
	}
}

// refreshData fetches fresh data from all sources
// Used for initial load and manual refresh
func (m *Model) refreshData() error {
	// Track when this refresh started
	refreshTime := time.Now()

	// Fetch agent statuses from MCP client (only if WebSocket is not connected)
	if !m.wsConnected {
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

		// Fetch messages from MCP client since last refresh (only if WebSocket is not connected)
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
	}

	// Fetch tasks from beads client with statuses "open" and "in_progress"
	if err := m.refreshBeadsData(); err != nil {
		return err
	}

	// Update last refresh time
	m.lastRefresh = refreshTime

	return nil
}

// refreshBeadsData fetches fresh data from beads only
// Used for periodic polling since beads is git-based
func (m *Model) refreshBeadsData() error {
	// Fetch tasks from beads client with statuses "open" and "in_progress"
	tasks, err := m.beadsClient.GetTasks([]string{"open", "in_progress"})
	if err != nil {
		// Don't fail completely on task fetch errors
		m.err = err
		m.beadsConnected = false
	} else {
		m.tasks = tasks
		m.beadsConnected = true
	}

	return nil
}
