package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the complete TUI layout
func (m Model) View() string {
	// If terminal size is not set yet, return empty
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Calculate pane dimensions based on terminal size
	// Reserve 3 lines for footer (1 line content + 2 for spacing/border)
	availableHeight := m.height - 3
	
	// Left pane (Agent Status): 1/3 width, full available height
	leftWidth := m.width / 3
	leftHeight := availableHeight
	
	// Right panes: 2/3 width
	rightWidth := m.width - leftWidth
	
	// Right top pane (Task Stream): 2/3 width, 1/2 of available height
	rightTopHeight := availableHeight / 2
	
	// Right bottom pane (MCP Log): 2/3 width, remaining height
	rightBottomHeight := availableHeight - rightTopHeight
	
	// Render individual panes
	agentPane := m.renderAgentPane(leftWidth, leftHeight)
	taskPane := m.renderTaskPane(rightWidth, rightTopHeight)
	logPane := m.renderLogPane(rightWidth, rightBottomHeight)
	footer := m.renderFooter(m.width)
	
	// Compose right column (task stream on top, log on bottom)
	rightColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		taskPane,
		logPane,
	)
	
	// Compose main layout (agent pane on left, right column on right)
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		agentPane,
		rightColumn,
	)
	
	// Compose final view with footer
	return lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		footer,
	)
}

// renderFooter renders the footer with keybindings and connection status
func (m Model) renderFooter(width int) string {
	// Define footer styles
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)
	
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true)
	
	// Build keybindings section
	keybindings := lipgloss.JoinHorizontal(
		lipgloss.Left,
		keyStyle.Render("(q)"),
		" quit | ",
		keyStyle.Render("(r)"),
		" refresh | ",
		keyStyle.Render("(t)"),
		" test",
	)
	
	// Check connection status for beads and MCP
	beadsStatus := m.getBeadsConnectionStatus()
	mcpStatus := m.getMCPConnectionStatus()
	
	// Build connection status section
	connectionStatus := lipgloss.JoinHorizontal(
		lipgloss.Left,
		"beads: ",
		beadsStatus,
		" | mcp: ",
		mcpStatus,
	)
	
	// Calculate spacing to push connection status to the right
	keybindingsWidth := lipgloss.Width(keybindings)
	connectionWidth := lipgloss.Width(connectionStatus)
	spacerWidth := width - keybindingsWidth - connectionWidth - 4 // 4 for padding
	
	if spacerWidth < 1 {
		spacerWidth = 1
	}
	
	spacer := lipgloss.NewStyle().Width(spacerWidth).Render("")
	
	// Compose footer content
	footerContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		keybindings,
		spacer,
		connectionStatus,
	)
	
	return footerStyle.Width(width - 2).Render(footerContent)
}

// getBeadsConnectionStatus returns a styled connection status for beads
func (m Model) getBeadsConnectionStatus() string {
	connectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	disconnectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	
	// Check beads connection status
	if m.beadsConnected {
		return connectedStyle.Render("●")
	}
	return disconnectedStyle.Render("○")
}

// getMCPConnectionStatus returns a styled connection status for MCP
func (m Model) getMCPConnectionStatus() string {
	connectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	disconnectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	
	// Check WebSocket connection status first (preferred)
	if m.wsConnected {
		return connectedStyle.Render("● ws")
	}
	
	// Fallback to HTTP polling
	if m.mcpClient != nil {
		return warningStyle.Render("● http")
	}
	
	return disconnectedStyle.Render("○")
}
