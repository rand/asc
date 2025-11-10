package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yourusername/asc/internal/beads"
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
		
	case taskActionMsg:
		return m.handleTaskAction(msg)
		
	case agentActionMsg:
		return m.handleAgentAction(msg)
		
	case logActionMsg:
		return m.handleLogAction(msg)
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle modal inputs first
	if m.showCreateModal {
		return m.handleCreateModalInput(msg)
	}
	
	if m.showConfirmModal {
		return m.handleConfirmModalInput(msg)
	}
	
	if m.searchMode {
		return m.handleSearchInput(msg)
	}
	
	// Handle task detail modal
	if m.showTaskModal {
		switch msg.String() {
		case "esc", "v":
			// Close modal
			m.showTaskModal = false
			return m, nil
		}
		return m, nil
	}
	
	// Normal key handling
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
		
	// Task interaction keys
	case "up":
		// Move selection up in task list
		if m.selectedTaskIndex > 0 {
			m.selectedTaskIndex--
		}
		return m, nil
		
	case "down":
		// Move selection down in task list
		filteredTasks := m.filterTasksByStatus([]string{"open", "in_progress"})
		if m.selectedTaskIndex < len(filteredTasks)-1 {
			m.selectedTaskIndex++
		}
		return m, nil
		
	case "c":
		// Claim selected task
		return m, claimTaskCmd(m)
		
	case "v":
		// View task details
		m.showTaskModal = true
		return m, nil
		
	case "n":
		// Create new task
		m.showCreateModal = true
		m.createTaskInput = ""
		return m, nil
		
	// Agent control keys
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Select agent by number
		agentNum := int(msg.String()[0] - '0')
		if agentNum > 0 && agentNum <= len(m.config.Agents) {
			m.selectedAgentIndex = agentNum - 1
		}
		return m, nil
		
	case "p":
		// Pause/resume selected agent
		return m, pauseAgentCmd(m)
		
	case "k":
		// Kill selected agent (with confirmation)
		m.showConfirmModal = true
		m.confirmAction = "kill"
		return m, nil
		
	case "R": // Shift+R for restart to avoid conflict with 'r' refresh
		// Restart selected agent (with confirmation)
		m.showConfirmModal = true
		m.confirmAction = "restart"
		return m, nil
		
	case "l":
		// View agent logs
		return m, viewAgentLogsCmd(m)
		
	// Log filtering keys
	case "/":
		// Enter search mode
		m.searchMode = true
		m.searchInput = ""
		return m, nil
		
	case "e":
		// Export logs
		return m, exportLogsCmd(m)
		
	case "a":
		// Cycle through agent filter
		return m.cycleAgentFilter(), nil
		
	case "m":
		// Cycle through message type filter
		return m.cycleMessageTypeFilter(), nil
		
	case "x":
		// Clear all filters
		m.searchInput = ""
		m.logFilterAgent = ""
		m.logFilterType = ""
		return m, nil
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

// handleTaskAction processes task action results
func (m Model) handleTaskAction(msg taskActionMsg) (tea.Model, tea.Cmd) {
	if !msg.success {
		m.err = fmt.Errorf("%s", msg.message)
	} else {
		// Refresh tasks after successful action
		return m, refreshDataCmd(m)
	}
	return m, nil
}

// handleAgentAction processes agent action results
func (m Model) handleAgentAction(msg agentActionMsg) (tea.Model, tea.Cmd) {
	if !msg.success {
		m.err = fmt.Errorf("%s", msg.message)
	}
	// Note: Agent status will be updated via WebSocket events
	return m, nil
}

// handleLogAction processes log action results
func (m Model) handleLogAction(msg logActionMsg) (tea.Model, tea.Cmd) {
	if !msg.success {
		m.err = fmt.Errorf("%s", msg.message)
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

// handleCreateModalInput handles input when create task modal is open
func (m Model) handleCreateModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel creation
		m.showCreateModal = false
		m.createTaskInput = ""
		return m, nil
		
	case "enter":
		// Create task
		if m.createTaskInput != "" {
			m.showCreateModal = false
			return m, createTaskCmd(m, m.createTaskInput)
		}
		return m, nil
		
	case "backspace":
		// Delete character
		if len(m.createTaskInput) > 0 {
			m.createTaskInput = m.createTaskInput[:len(m.createTaskInput)-1]
		}
		return m, nil
		
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			m.createTaskInput += msg.String()
		}
		return m, nil
	}
}

// handleConfirmModalInput handles input when confirmation modal is open
func (m Model) handleConfirmModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm action
		m.showConfirmModal = false
		switch m.confirmAction {
		case "kill":
			return m, killAgentCmd(m)
		case "restart":
			return m, restartAgentCmd(m)
		}
		return m, nil
		
	case "n", "N", "esc":
		// Cancel action
		m.showConfirmModal = false
		m.confirmAction = ""
		return m, nil
	}
	
	return m, nil
}

// handleSearchInput handles input when in search mode
func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Exit search mode
		m.searchMode = false
		m.searchInput = ""
		return m, nil
		
	case "enter":
		// Apply search filter
		m.searchMode = false
		// Search input is already stored in m.searchInput
		return m, nil
		
	case "backspace":
		// Delete character
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}
		return m, nil
		
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
		}
		return m, nil
	}
}

// claimTaskCmd claims the selected task for the current user
func claimTaskCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		filteredTasks := m.filterTasksByStatus([]string{"open", "in_progress"})
		if m.selectedTaskIndex < 0 || m.selectedTaskIndex >= len(filteredTasks) {
			return taskActionMsg{
				success: false,
				message: "No task selected",
			}
		}
		
		task := filteredTasks[m.selectedTaskIndex]
		
		// Get current user from environment
		user := "current-user" // TODO: Get from environment or config
		
		// Update task with assignee
		updates := beads.TaskUpdate{
			Assignee: &user,
		}
		
		if err := m.beadsClient.UpdateTask(task.ID, updates); err != nil {
			return taskActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to claim task: %v", err),
			}
		}
		
		return taskActionMsg{
			success: true,
			message: fmt.Sprintf("Claimed task #%s", task.ID),
		}
	}
}

// createTaskCmd creates a new task with the given title
func createTaskCmd(m Model, title string) tea.Cmd {
	return func() tea.Msg {
		task, err := m.beadsClient.CreateTask(title)
		if err != nil {
			return taskActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to create task: %v", err),
			}
		}
		
		return taskActionMsg{
			success: true,
			message: fmt.Sprintf("Created task #%s", task.ID),
		}
	}
}

// pauseAgentCmd pauses or resumes the selected agent
func pauseAgentCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		// Get selected agent name
		agentNames := m.getAgentNames()
		if m.selectedAgentIndex < 0 || m.selectedAgentIndex >= len(agentNames) {
			return agentActionMsg{
				success: false,
				message: "No agent selected",
			}
		}
		
		agentName := agentNames[m.selectedAgentIndex]
		
		// Get process info
		info, err := m.procManager.GetProcessInfo(agentName)
		if err != nil {
			return agentActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to get agent info: %v", err),
			}
		}
		
		// Check if running
		if !m.procManager.IsRunning(info.PID) {
			return agentActionMsg{
				success: false,
				message: fmt.Sprintf("Agent %s is not running", agentName),
			}
		}
		
		// Send SIGSTOP to pause (not implemented in basic version)
		// For now, just return a message
		return agentActionMsg{
			success: false,
			message: "Pause/resume not yet implemented",
		}
	}
}

// killAgentCmd kills the selected agent
func killAgentCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		// Get selected agent name
		agentNames := m.getAgentNames()
		if m.selectedAgentIndex < 0 || m.selectedAgentIndex >= len(agentNames) {
			return agentActionMsg{
				success: false,
				message: "No agent selected",
			}
		}
		
		agentName := agentNames[m.selectedAgentIndex]
		
		// Get process info
		info, err := m.procManager.GetProcessInfo(agentName)
		if err != nil {
			return agentActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to get agent info: %v", err),
			}
		}
		
		// Stop the process
		if err := m.procManager.Stop(info.PID); err != nil {
			return agentActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to kill agent: %v", err),
			}
		}
		
		return agentActionMsg{
			success: true,
			message: fmt.Sprintf("Killed agent %s", agentName),
		}
	}
}

// restartAgentCmd restarts the selected agent
func restartAgentCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		// Get selected agent name
		agentNames := m.getAgentNames()
		if m.selectedAgentIndex < 0 || m.selectedAgentIndex >= len(agentNames) {
			return agentActionMsg{
				success: false,
				message: "No agent selected",
			}
		}
		
		agentName := agentNames[m.selectedAgentIndex]
		
		// Get process info
		info, err := m.procManager.GetProcessInfo(agentName)
		if err != nil {
			return agentActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to get agent info: %v", err),
			}
		}
		
		// Stop the process
		if m.procManager.IsRunning(info.PID) {
			if err := m.procManager.Stop(info.PID); err != nil {
				return agentActionMsg{
					success: false,
					message: fmt.Sprintf("Failed to stop agent: %v", err),
				}
			}
		}
		
		// Restart with same command and environment
		env := make([]string, 0, len(info.Env))
		for k, v := range info.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		
		_, err = m.procManager.Start(info.Name, info.Command, info.Args, env)
		if err != nil {
			return agentActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to restart agent: %v", err),
			}
		}
		
		return agentActionMsg{
			success: true,
			message: fmt.Sprintf("Restarted agent %s", agentName),
		}
	}
}

// viewAgentLogsCmd opens the log file for the selected agent
func viewAgentLogsCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		// Get selected agent name
		agentNames := m.getAgentNames()
		if m.selectedAgentIndex < 0 || m.selectedAgentIndex >= len(agentNames) {
			return agentActionMsg{
				success: false,
				message: "No agent selected",
			}
		}
		
		agentName := agentNames[m.selectedAgentIndex]
		
		// Get process info
		info, err := m.procManager.GetProcessInfo(agentName)
		if err != nil {
			return agentActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to get agent info: %v", err),
			}
		}
		
		// For now, just return the log file path
		// In a full implementation, this could open the file in a pager
		return agentActionMsg{
			success: true,
			message: fmt.Sprintf("Log file: %s", info.LogFile),
		}
	}
}

// exportLogsCmd exports filtered logs to a file
func exportLogsCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		// Filter messages
		messages := m.getFilteredMessages()
		
		// Create export file
		filename := fmt.Sprintf("asc-logs-%s.txt", time.Now().Format("20060102-150405"))
		file, err := os.Create(filename)
		if err != nil {
			return logActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to create export file: %v", err),
			}
		}
		defer file.Close()
		
		// Write messages
		for _, msg := range messages {
			line := fmt.Sprintf("[%s] [%s] [%s] → %s\n",
				msg.Timestamp.Format("15:04:05"),
				msg.Type,
				msg.Source,
				msg.Content)
			if _, err := file.WriteString(line); err != nil {
				return logActionMsg{
					success: false,
					message: fmt.Sprintf("Failed to write to export file: %v", err),
				}
			}
		}
		
		return logActionMsg{
			success: true,
			message: fmt.Sprintf("Exported %d messages to %s", len(messages), filename),
		}
	}
}

// taskActionMsg is sent when a task action completes
type taskActionMsg struct {
	success bool
	message string
}

// agentActionMsg is sent when an agent action completes
type agentActionMsg struct {
	success bool
	message string
}

// logActionMsg is sent when a log action completes
type logActionMsg struct {
	success bool
	message string
}

// getFilteredMessages returns messages filtered by current filters
func (m Model) getFilteredMessages() []mcp.Message {
	var filtered []mcp.Message
	
	for _, msg := range m.messages {
		// Filter by agent name
		if m.logFilterAgent != "" && msg.Source != m.logFilterAgent {
			continue
		}
		
		// Filter by message type
		if m.logFilterType != "" && string(msg.Type) != m.logFilterType {
			continue
		}
		
		// Filter by search input
		if m.searchInput != "" {
			searchLower := strings.ToLower(m.searchInput)
			contentLower := strings.ToLower(msg.Content)
			sourceLower := strings.ToLower(msg.Source)
			
			if !strings.Contains(contentLower, searchLower) && !strings.Contains(sourceLower, searchLower) {
				continue
			}
		}
		
		filtered = append(filtered, msg)
	}
	
	return filtered
}

// cycleAgentFilter cycles through agent name filters
func (m Model) cycleAgentFilter() Model {
	agentNames := m.getAgentNames()
	
	if m.logFilterAgent == "" {
		// Start with first agent
		if len(agentNames) > 0 {
			m.logFilterAgent = agentNames[0]
		}
	} else {
		// Find current agent and move to next
		found := false
		for i, name := range agentNames {
			if name == m.logFilterAgent {
				if i+1 < len(agentNames) {
					m.logFilterAgent = agentNames[i+1]
				} else {
					// Cycle back to "all" (empty filter)
					m.logFilterAgent = ""
				}
				found = true
				break
			}
		}
		if !found {
			// Current filter not found, reset
			m.logFilterAgent = ""
		}
	}
	
	return m
}

// cycleMessageTypeFilter cycles through message type filters
func (m Model) cycleMessageTypeFilter() Model {
	messageTypes := []string{"lease", "beads", "error", "message"}
	
	if m.logFilterType == "" {
		// Start with first type
		m.logFilterType = messageTypes[0]
	} else {
		// Find current type and move to next
		found := false
		for i, msgType := range messageTypes {
			if msgType == m.logFilterType {
				if i+1 < len(messageTypes) {
					m.logFilterType = messageTypes[i+1]
				} else {
					// Cycle back to "all" (empty filter)
					m.logFilterType = ""
				}
				found = true
				break
			}
		}
		if !found {
			// Current filter not found, reset
			m.logFilterType = ""
		}
	}
	
	return m
}
