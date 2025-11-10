package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yourusername/asc/internal/mcp"
)

// refreshDataMsg is sent when data refresh is complete
type refreshDataMsg struct {
	err error
}

// testResultMsg is sent when test command completes
type testResultMsg struct {
	success bool
	message string
}

// Update handles incoming messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		return m.handleResize(msg)

	case tickMsg:
		return m.handleTick()

	case refreshDataMsg:
		return m.handleRefresh(msg)

	case testResultMsg:
		return m.handleTestResult(msg)

	case wsEventMsg:
		return m.handleWSEvent(msg)
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		// Quit - trigger shutdown sequence
		return m, tea.Quit

	case "r":
		// Force refresh
		return m, refreshDataCmd(m)

	case "t":
		// Run test command
		return m, runTestCmd(m)
	}

	return m, nil
}

// handleResize processes terminal resize events
func (m Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	return m, nil
}

// handleTick processes periodic tick events
func (m Model) handleTick() (tea.Model, tea.Cmd) {
	// Schedule next tick and refresh beads data only
	// MCP data is updated via WebSocket events
	return m, tea.Batch(
		tickCmd(),
		refreshBeadsCmd(m),
	)
}

// handleRefresh processes data refresh completion
func (m Model) handleRefresh(msg refreshDataMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
	}
	return m, nil
}

// handleTestResult processes test command results
func (m Model) handleTestResult(msg testResultMsg) (tea.Model, tea.Cmd) {
	// Update error state based on test result
	// The test result will be displayed in the log pane
	m.err = nil
	if !msg.success {
		m.err = fmt.Errorf("test failed: %s", msg.message)
	}
	
	return m, nil
}

// handleWSEvent processes WebSocket events from the MCP server
func (m Model) handleWSEvent(msg wsEventMsg) (tea.Model, tea.Cmd) {
	event := mcp.Event(msg)
	
	switch event.Type {
	case mcp.EventConnected:
		// WebSocket connected successfully
		m.wsConnected = true
		m.err = nil
		
	case mcp.EventDisconnected:
		// WebSocket disconnected - will auto-reconnect
		m.wsConnected = false
		
	case mcp.EventAgentStatus:
		// Agent status changed - update the agent in our list
		if event.AgentStatus != nil {
			m.updateAgentStatus(*event.AgentStatus)
		}
		
	case mcp.EventNewMessage:
		// New message received - add to message list
		if event.Message != nil {
			m.messages = append(m.messages, *event.Message)
			
			// Limit message buffer to last 100 messages
			if len(m.messages) > 100 {
				m.messages = m.messages[len(m.messages)-100:]
			}
		}
		
	case mcp.EventError:
		// WebSocket error - log but don't fail
		// We'll continue with polling fallback
		if event.Error != "" {
			m.err = fmt.Errorf("websocket error: %s", event.Error)
		}
	}
	
	// Continue listening for next WebSocket event
	return m, waitForWSEventCmd(m.wsClient)
}

// updateAgentStatus updates or adds an agent status in the agents list
func (m *Model) updateAgentStatus(newStatus mcp.AgentStatus) {
	// Find and update existing agent
	for i, agent := range m.agents {
		if agent.Name == newStatus.Name {
			m.agents[i] = newStatus
			return
		}
	}
	
	// Agent not found - add it
	m.agents = append(m.agents, newStatus)
}

// runTestCmd executes the test command asynchronously
func runTestCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		// Create test task
		task, err := m.beadsClient.CreateTask("asc test task")
		if err != nil {
			return testResultMsg{
				success: false,
				message: fmt.Sprintf("Failed to create test task: %v", err),
			}
		}

		// Send test message to MCP
		testMsg := mcp.Message{
			Type:    mcp.TypeMessage,
			Source:  "asc-test",
			Content: "test message",
		}
		if err := m.mcpClient.SendMessage(testMsg); err != nil {
			// Clean up test task
			_ = m.beadsClient.DeleteTask(task.ID)
			return testResultMsg{
				success: false,
				message: fmt.Sprintf("Failed to send test message: %v", err),
			}
		}

		// Clean up test task
		if err := m.beadsClient.DeleteTask(task.ID); err != nil {
			return testResultMsg{
				success: false,
				message: fmt.Sprintf("Failed to delete test task: %v", err),
			}
		}

		return testResultMsg{
			success: true,
			message: "Stack is healthy",
		}
	}
}
