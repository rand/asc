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
	// Schedule next tick and refresh data
	return m, tea.Batch(
		tickCmd(),
		refreshDataCmd(m),
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
