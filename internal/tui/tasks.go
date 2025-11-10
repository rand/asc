package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yourusername/asc/internal/beads"
)

// Task status icons
const (
	iconOpen       = "○" // Empty circle for open tasks
	iconInProgress = "◉" // Filled circle with dot for in-progress tasks
)

// Color styles for task states
var (
	styleOpen       = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // Gray
	styleInProgress = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true) // Yellow/Bold
)

// Border style for the task pane
var taskPaneBorder = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(0, 1)

// renderTaskPane renders the beads task stream pane
func (m Model) renderTaskPane(width, height int) string {
	// Calculate content dimensions (accounting for border and padding)
	contentWidth := width - 4  // 2 for border + 2 for padding
	contentHeight := height - 2 // 2 for border

	var lines []string
	
	// Filter tasks by status (open, in_progress)
	filteredTasks := m.filterTasksByStatus([]string{"open", "in_progress"})
	
	// Build task lines
	for _, task := range filteredTasks {
		line := m.formatTaskLine(task, contentWidth)
		lines = append(lines, line)
	}
	
	// If no tasks, show a message
	if len(lines) == 0 {
		lines = append(lines, styleOpen.Render("No open or in-progress tasks"))
	}
	
	// Pad or truncate to fit height
	content := m.fitContent(lines, contentHeight)
	
	// Join lines and apply border with title
	contentStr := strings.Join(content, "\n")
	
	return taskPaneBorder.
		Width(width - 2).
		Height(height - 2).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("Task Stream"),
			"",
			contentStr,
		))
}

// filterTasksByStatus filters tasks by the given statuses
func (m Model) filterTasksByStatus(statuses []string) []beads.Task {
	statusMap := make(map[string]bool)
	for _, status := range statuses {
		statusMap[status] = true
	}
	
	var filtered []beads.Task
	for _, task := range m.tasks {
		if statusMap[task.Status] {
			filtered = append(filtered, task)
		}
	}
	
	return filtered
}

// formatTaskLine formats a single task line
func (m Model) formatTaskLine(task beads.Task, maxWidth int) string {
	// Get icon and style based on status
	icon, style := m.getTaskIconAndStyle(task.Status)
	
	// Build the line: icon + ID + title
	line := fmt.Sprintf("%s #%s %s", icon, task.ID, task.Title)
	
	// Truncate if too long
	if len(line) > maxWidth {
		if maxWidth > 3 {
			line = line[:maxWidth-3] + "..."
		} else {
			line = line[:maxWidth]
		}
	}
	
	// Apply styling
	return style.Render(line)
}

// getTaskIconAndStyle returns the icon and style for a given task status
func (m Model) getTaskIconAndStyle(status string) (string, lipgloss.Style) {
	switch status {
	case "in_progress":
		return iconInProgress, styleInProgress
	case "open":
		return iconOpen, styleOpen
	default:
		return iconOpen, styleOpen
	}
}
