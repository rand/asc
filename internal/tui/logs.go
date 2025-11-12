package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/rand/asc/internal/mcp"
)

// Color styles for message types
var (
	styleLease   = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))  // Blue
	styleBeads   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))  // Green
	styleMsgError = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Red
	styleMessage = lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // Default/White
)

// Border style for the log pane
var logPaneBorder = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(0, 1)

// Maximum number of messages to display
const maxLogMessages = 100

// renderLogPane renders the MCP interaction log pane
func (m Model) renderLogPane(width, height int) string {
	// Calculate content dimensions (accounting for border and padding)
	contentWidth := width - 4  // 2 for border + 2 for padding
	contentHeight := height - 2 // 2 for border

	var lines []string
	
	// Get filtered messages
	messages := m.getFilteredMessages()
	
	// Limit to recent messages
	if len(messages) > maxLogMessages {
		messages = messages[len(messages)-maxLogMessages:]
	}
	
	// Build message lines
	for _, msg := range messages {
		line := m.formatMessageLine(msg, contentWidth)
		lines = append(lines, line)
	}
	
	// If no messages, show a message
	if len(lines) == 0 {
		lines = append(lines, styleMessage.Render("No messages yet"))
	}
	
	// Auto-scroll to bottom: take the last contentHeight lines
	if len(lines) > contentHeight {
		lines = lines[len(lines)-contentHeight:]
	} else {
		// Pad with empty lines at the top if needed
		for len(lines) < contentHeight {
			lines = append([]string{""}, lines...)
		}
	}
	
	// Join lines and apply border with title
	contentStr := strings.Join(lines, "\n")
	
	// Build title with active filters
	title := "MCP Interaction Log"
	var filterParts []string
	if m.searchInput != "" {
		filterParts = append(filterParts, fmt.Sprintf("search:%s", m.searchInput))
	}
	if m.logFilterAgent != "" {
		filterParts = append(filterParts, fmt.Sprintf("agent:%s", m.logFilterAgent))
	}
	if m.logFilterType != "" {
		filterParts = append(filterParts, fmt.Sprintf("type:%s", m.logFilterType))
	}
	if len(filterParts) > 0 {
		title += " [" + strings.Join(filterParts, " ") + "]"
	}
	
	// Add keybindings hint
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("/:search a:agent m:type x:clear e:export")
	
	return logPaneBorder.
		Width(width - 2).
		Height(height - 2).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render(title),
			hint,
			contentStr,
		))
}

// getRecentMessages returns the last N messages from the message list
func (m Model) getRecentMessages(limit int) []mcp.Message {
	if len(m.messages) <= limit {
		return m.messages
	}
	return m.messages[len(m.messages)-limit:]
}

// formatMessageLine formats a single message line
// Format: [HH:MM:SS] [Type] [Source] → [Content]
func (m Model) formatMessageLine(msg mcp.Message, maxWidth int) string {
	// Format timestamp as HH:MM:SS
	timestamp := msg.Timestamp.Format("15:04:05")
	
	// Format the message type
	msgType := string(msg.Type)
	
	// Build the line
	line := fmt.Sprintf("[%s] [%s] [%s] → %s", timestamp, msgType, msg.Source, msg.Content)
	
	// Truncate if too long
	if len(line) > maxWidth {
		if maxWidth > 3 {
			line = line[:maxWidth-3] + "..."
		} else {
			line = line[:maxWidth]
		}
	}
	
	// Apply color styling based on message type
	style := m.getMessageStyle(msg.Type)
	return style.Render(line)
}

// getMessageStyle returns the style for a given message type
func (m Model) getMessageStyle(msgType mcp.MessageType) lipgloss.Style {
	switch msgType {
	case mcp.TypeLease:
		return styleLease
	case mcp.TypeBeads:
		return styleBeads
	case mcp.TypeError:
		return styleMsgError
	case mcp.TypeMessage:
		return styleMessage
	default:
		return styleMessage
	}
}
