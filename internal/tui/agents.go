package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yourusername/asc/internal/mcp"
)

// Agent status icons
const (
	iconIdle    = "●" // Filled circle for idle
	iconWorking = "⟳" // Rotating arrow for working
	iconError   = "!" // Exclamation for error
	iconOffline = "○" // Empty circle for offline
)

// Color styles for agent states
var (
	styleIdle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))  // Green
	styleWorking = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))  // Blue
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))   // Red
	styleOffline = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Gray
)

// Border style for the agent pane
var agentPaneBorder = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(0, 1)

// renderAgentPane renders the agent status pane
func (m Model) renderAgentPane(width, height int) string {
	// Calculate content dimensions (accounting for border and padding)
	contentWidth := width - 4  // 2 for border + 2 for padding
	contentHeight := height - 2 // 2 for border

	var lines []string
	
	// Build agent status map from current agent statuses
	statusMap := make(map[string]mcp.AgentStatus)
	for _, agent := range m.agents {
		statusMap[agent.Name] = agent
	}
	
	// Get sorted agent names for consistent ordering
	agentNames := m.getAgentNames()
	
	// Iterate through agents from config to maintain consistent ordering
	for i, agentName := range agentNames {
		status, exists := statusMap[agentName]
		if !exists {
			// Agent not found in status updates - mark as offline
			status = mcp.AgentStatus{
				Name:  agentName,
				State: mcp.StateOffline,
			}
		}
		
		line := m.formatAgentLine(status, contentWidth, i+1, i == m.selectedAgentIndex)
		lines = append(lines, line)
	}
	
	// If no agents configured, show a message
	if len(lines) == 0 {
		lines = append(lines, styleOffline.Render("No agents configured"))
	}
	
	// Pad or truncate to fit height
	content := m.fitContent(lines, contentHeight)
	
	// Join lines and apply border with title
	contentStr := strings.Join(content, "\n")
	
	// Add keybindings hint
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("1-9:select p:pause k:kill R:restart l:logs")
	
	return agentPaneBorder.
		Width(width - 2).
		Height(height - 2).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("Agent Status"),
			hint,
			contentStr,
		))
}

// formatAgentLine formats a single agent status line
func (m Model) formatAgentLine(status mcp.AgentStatus, maxWidth int, number int, selected bool) string {
	// Get icon and style based on state
	icon, style := m.getAgentIconAndStyle(status.State)
	
	// Format the status text
	var statusText string
	switch status.State {
	case mcp.StateIdle:
		statusText = "Idle"
	case mcp.StateWorking:
		if status.CurrentTask != "" {
			statusText = fmt.Sprintf("Working on #%s", status.CurrentTask)
		} else {
			statusText = "Working"
		}
	case mcp.StateError:
		statusText = "Error"
	case mcp.StateOffline:
		statusText = "Offline"
	default:
		statusText = "Unknown"
	}
	
	// Add selection indicator
	prefix := fmt.Sprintf("%d ", number)
	if selected {
		prefix = fmt.Sprintf("%d▶", number)
		style = style.Background(lipgloss.Color("237")) // Highlight background
	}
	
	// Build the line: number + icon + name + status
	line := fmt.Sprintf("%s %s %s - %s", prefix, icon, status.Name, statusText)
	
	// Truncate if too long
	if len(line) > maxWidth {
		if maxWidth > 3 {
			line = line[:maxWidth-3] + "..."
		} else {
			line = line[:maxWidth]
		}
	}
	
	// Apply color styling
	return style.Render(line)
}

// getAgentNames returns a sorted list of agent names from config
// This is defined here to avoid import cycles
func (m Model) getAgentNames() []string {
	names := make([]string, 0, len(m.config.Agents))
	for name := range m.config.Agents {
		names = append(names, name)
	}
	return names
}

// getAgentIconAndStyle returns the icon and style for a given agent state
func (m Model) getAgentIconAndStyle(state mcp.AgentState) (string, lipgloss.Style) {
	switch state {
	case mcp.StateIdle:
		return iconIdle, styleIdle
	case mcp.StateWorking:
		return iconWorking, styleWorking
	case mcp.StateError:
		return iconError, styleError
	case mcp.StateOffline:
		return iconOffline, styleOffline
	default:
		return iconOffline, styleOffline
	}
}

// fitContent pads or truncates lines to fit the target height
func (m Model) fitContent(lines []string, targetHeight int) []string {
	if len(lines) > targetHeight {
		// Truncate if too many lines
		return lines[:targetHeight]
	}
	
	// Pad with empty lines if needed
	for len(lines) < targetHeight {
		lines = append(lines, "")
	}
	
	return lines
}
