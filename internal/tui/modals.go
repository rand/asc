package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Modal styles
var (
	modalOverlayStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("0")).
				Foreground(lipgloss.Color("15"))

	modalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("15"))

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12"))

	modalLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	modalInputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("237")).
			Padding(0, 1)
)

// renderTaskDetailModal renders a modal showing task details
func (m Model) renderTaskDetailModal() string {
	filteredTasks := m.filterTasksByStatus([]string{"open", "in_progress"})
	if m.selectedTaskIndex < 0 || m.selectedTaskIndex >= len(filteredTasks) {
		return ""
	}

	task := filteredTasks[m.selectedTaskIndex]

	// Build modal content
	var content strings.Builder
	content.WriteString(modalTitleStyle.Render(fmt.Sprintf("Task #%s", task.ID)))
	content.WriteString("\n\n")
	content.WriteString(modalLabelStyle.Render("Title: "))
	content.WriteString(task.Title)
	content.WriteString("\n\n")
	content.WriteString(modalLabelStyle.Render("Status: "))
	content.WriteString(task.Status)
	content.WriteString("\n\n")
	content.WriteString(modalLabelStyle.Render("Phase: "))
	content.WriteString(task.Phase)
	content.WriteString("\n\n")
	if task.Assignee != "" {
		content.WriteString(modalLabelStyle.Render("Assignee: "))
		content.WriteString(task.Assignee)
		content.WriteString("\n\n")
	}
	content.WriteString(modalLabelStyle.Render("Press 'v' or 'esc' to close"))

	// Render modal box
	modalContent := modalBoxStyle.Render(content.String())

	// Center the modal
	return m.centerModal(modalContent)
}

// renderCreateTaskModal renders a modal for creating a new task
func (m Model) renderCreateTaskModal() string {
	// Build modal content
	var content strings.Builder
	content.WriteString(modalTitleStyle.Render("Create New Task"))
	content.WriteString("\n\n")
	content.WriteString(modalLabelStyle.Render("Title:"))
	content.WriteString("\n")
	content.WriteString(modalInputStyle.Render(m.createTaskInput + "█"))
	content.WriteString("\n\n")
	content.WriteString(modalLabelStyle.Render("Press 'enter' to create, 'esc' to cancel"))

	// Render modal box
	modalContent := modalBoxStyle.Render(content.String())

	// Center the modal
	return m.centerModal(modalContent)
}

// renderConfirmModal renders a confirmation dialog
func (m Model) renderConfirmModal() string {
	// Get selected agent name
	agentNames := m.getAgentNames()
	agentName := "unknown"
	if m.selectedAgentIndex >= 0 && m.selectedAgentIndex < len(agentNames) {
		agentName = agentNames[m.selectedAgentIndex]
	}

	// Build modal content based on action
	var content strings.Builder
	content.WriteString(modalTitleStyle.Render("Confirm Action"))
	content.WriteString("\n\n")

	switch m.confirmAction {
	case "kill":
		content.WriteString(fmt.Sprintf("Kill agent '%s'?", agentName))
	case "restart":
		content.WriteString(fmt.Sprintf("Restart agent '%s'?", agentName))
	default:
		content.WriteString("Confirm this action?")
	}

	content.WriteString("\n\n")
	content.WriteString(modalLabelStyle.Render("Press 'y' to confirm, 'n' or 'esc' to cancel"))

	// Render modal box
	modalContent := modalBoxStyle.Render(content.String())

	// Center the modal
	return m.centerModal(modalContent)
}

// renderSearchInput renders the search input bar
func (m Model) renderSearchInput() string {
	// Build search bar
	var content strings.Builder
	content.WriteString(modalLabelStyle.Render("Search: "))
	content.WriteString(modalInputStyle.Render(m.searchInput + "█"))
	content.WriteString("  ")
	content.WriteString(modalLabelStyle.Render("Press 'enter' to apply, 'esc' to cancel"))

	return modalBoxStyle.
		Width(m.width - 4).
		Render(content.String())
}

// centerModal centers a modal content string on the screen
func (m Model) centerModal(modalContent string) string {
	// Calculate modal dimensions
	modalWidth := lipgloss.Width(modalContent)
	modalHeight := lipgloss.Height(modalContent)

	// Calculate centering offsets
	horizontalPadding := (m.width - modalWidth) / 2
	verticalPadding := (m.height - modalHeight) / 2

	if horizontalPadding < 0 {
		horizontalPadding = 0
	}
	if verticalPadding < 0 {
		verticalPadding = 0
	}

	// Add vertical padding
	var paddedContent strings.Builder
	for i := 0; i < verticalPadding; i++ {
		paddedContent.WriteString("\n")
	}
	paddedContent.WriteString(modalContent)

	// Add horizontal padding
	centeredContent := lipgloss.NewStyle().
		PaddingLeft(horizontalPadding).
		Render(paddedContent.String())

	return centeredContent
}
