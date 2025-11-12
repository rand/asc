package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rand/asc/internal/beads"
	"github.com/rand/asc/internal/mcp"
)

// TestFrameworkCreation tests creating a new test framework
func TestFrameworkCreation(t *testing.T) {
	tf := NewTestFramework()

	if tf == nil {
		t.Fatal("Expected test framework to be created")
	}

	model := tf.GetModel()
	if model == nil {
		t.Fatal("Expected model to be created")
	}

	if model.width != 120 {
		t.Errorf("Expected width to be 120, got %d", model.width)
	}

	if model.height != 40 {
		t.Errorf("Expected height to be 40, got %d", model.height)
	}

	if len(model.config.Agents) != 2 {
		t.Errorf("Expected 2 agents in config, got %d", len(model.config.Agents))
	}
}

// TestFrameworkKeySimulation tests key press simulation
func TestFrameworkKeySimulation(t *testing.T) {
	tf := NewTestFramework()

	// Test sending a key
	model := tf.SendKey(tea.KeyDown)
	if model.selectedTaskIndex != 0 {
		t.Errorf("Expected selectedTaskIndex to be 0, got %d", model.selectedTaskIndex)
	}

	// Test sending a rune
	model = tf.SendKeyRune('q')
	// Model should still be valid after quit key
	if model.config.Core.BeadsDBPath == "" {
		t.Error("Expected model to still have config after quit key")
	}
}

// TestFrameworkResize tests terminal resize simulation
func TestFrameworkResize(t *testing.T) {
	tf := NewTestFramework()

	model := tf.Resize(80, 24)
	if model.width != 80 {
		t.Errorf("Expected width to be 80, got %d", model.width)
	}

	if model.height != 24 {
		t.Errorf("Expected height to be 24, got %d", model.height)
	}
}

// TestFrameworkRender tests rendering
func TestFrameworkRender(t *testing.T) {
	tf := NewTestFramework()

	output := tf.Render()
	if output == "" {
		t.Error("Expected non-empty render output")
	}

	// Should contain some basic UI elements
	if !strings.Contains(output, "Agent") && !strings.Contains(output, "Task") {
		t.Error("Expected output to contain UI elements")
	}
}

// TestMockBeadsClient tests the mock beads client
func TestMockBeadsClient(t *testing.T) {
	client := NewMockBeadsClient()

	// Test adding tasks
	task1 := beads.Task{ID: "1", Title: "Task 1", Status: "open"}
	client.AddTask(task1)

	tasks, err := client.GetTasks([]string{"open"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Test creating tasks
	newTask, err := client.CreateTask("New Task")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if newTask.Title != "New Task" {
		t.Errorf("Expected title 'New Task', got %s", newTask.Title)
	}

	// Test updating tasks
	assignee := "test-user"
	err = client.UpdateTask("1", beads.TaskUpdate{Assignee: &assignee})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	tasks, _ = client.GetTasks([]string{"open"})
	if tasks[0].Assignee != "test-user" {
		t.Errorf("Expected assignee 'test-user', got %s", tasks[0].Assignee)
	}

	// Test deleting tasks
	err = client.DeleteTask("1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	tasks, _ = client.GetTasks([]string{"open"})
	if len(tasks) != 1 { // Only the created task should remain
		t.Errorf("Expected 1 task after deletion, got %d", len(tasks))
	}
}

// TestMockMCPClient tests the mock MCP client
func TestMockMCPClient(t *testing.T) {
	client := NewMockMCPClient()

	// Test adding messages
	msg := mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "Test message",
	}
	client.AddMessage(msg)

	messages, err := client.GetMessages(time.Now().Add(-1 * time.Minute))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	// Test sending messages
	err = client.SendMessage(msg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	messages, _ = client.GetMessages(time.Now().Add(-1 * time.Minute))
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Test agent status
	status := mcp.AgentStatus{
		Name:  "test-agent",
		State: mcp.StateWorking,
	}
	client.SetAgentStatus(status)

	retrievedStatus, err := client.GetAgentStatus("test-agent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrievedStatus.State != mcp.StateWorking {
		t.Errorf("Expected state Working, got %v", retrievedStatus.State)
	}

	// Test getting all statuses
	statuses, err := client.GetAllAgentStatuses(2 * time.Minute)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(statuses) != 1 {
		t.Errorf("Expected 1 status, got %d", len(statuses))
	}
}

// TestMockProcessManager tests the mock process manager
func TestMockProcessManager(t *testing.T) {
	pm := NewMockProcessManager()

	// Test starting a process
	pid, err := pm.Start("test-agent", "python", []string{"agent.py"}, []string{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if pid <= 0 {
		t.Errorf("Expected positive PID, got %d", pid)
	}

	// Test checking if running
	if !pm.IsRunning(pid) {
		t.Error("Expected process to be running")
	}

	// Test getting status
	status := pm.GetStatus(pid)
	if status != "running" {
		t.Errorf("Expected status 'running', got %s", status)
	}

	// Test getting process info
	info, err := pm.GetProcessInfo("test-agent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if info.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got %s", info.Name)
	}

	// Test listing processes
	processes, err := pm.ListProcesses()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(processes) != 1 {
		t.Errorf("Expected 1 process, got %d", len(processes))
	}

	// Test stopping a process
	err = pm.Stop(pid)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if pm.IsRunning(pid) {
		t.Error("Expected process to be stopped")
	}

	// Test stopping all processes
	pm.Start("agent1", "python", []string{}, []string{})
	pm.Start("agent2", "python", []string{}, []string{})

	err = pm.StopAll()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	processes, _ = pm.ListProcesses()
	if len(processes) != 0 {
		t.Errorf("Expected 0 processes after StopAll, got %d", len(processes))
	}
}

// TestMockTerminal tests the mock terminal
func TestMockTerminal(t *testing.T) {
	terminal := NewMockTerminal(100, 30)

	if terminal.GetWidth() != 100 {
		t.Errorf("Expected width 100, got %d", terminal.GetWidth())
	}

	if terminal.GetHeight() != 30 {
		t.Errorf("Expected height 30, got %d", terminal.GetHeight())
	}

	tf := NewTestFramework()
	model := *tf.GetModel()

	output := terminal.Render(model)
	if output == "" {
		t.Error("Expected non-empty output")
	}

	if terminal.GetOutput() != output {
		t.Error("Expected GetOutput to return last rendered output")
	}
}

// TestTestHelper tests the test helper functions
func TestTestHelper(t *testing.T) {
	helper := NewTestHelper()

	// Test creating test task
	task := helper.CreateTestTask("1", "Test Task", "open")
	if task.ID != "1" {
		t.Errorf("Expected ID '1', got %s", task.ID)
	}
	if task.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got %s", task.Title)
	}

	// Test creating test message
	msg := helper.CreateTestMessage(mcp.TypeMessage, "test-agent", "Test content")
	if msg.Type != mcp.TypeMessage {
		t.Errorf("Expected type Message, got %v", msg.Type)
	}
	if msg.Source != "test-agent" {
		t.Errorf("Expected source 'test-agent', got %s", msg.Source)
	}

	// Test creating test agent status
	status := helper.CreateTestAgentStatus("test-agent", mcp.StateIdle)
	if status.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got %s", status.Name)
	}
	if status.State != mcp.StateIdle {
		t.Errorf("Expected state Idle, got %v", status.State)
	}

	// Test creating test config
	cfg := helper.CreateTestConfig()
	if cfg.Core.BeadsDBPath != "./test-beads" {
		t.Errorf("Expected beads path './test-beads', got %s", cfg.Core.BeadsDBPath)
	}
}

// TestTestHelperAssertions tests the assertion helper functions
func TestTestHelperAssertions(t *testing.T) {
	helper := NewTestHelper()

	// Create a mock testing.T
	mockT := &mockTestingT{}

	// Test AssertContains
	if !helper.AssertContains(mockT, "hello world", "world") {
		t.Error("Expected AssertContains to return true")
	}

	if mockT.errorCount != 0 {
		t.Errorf("Expected no errors, got %d", mockT.errorCount)
	}

	// Test AssertContains failure
	if helper.AssertContains(mockT, "hello world", "foo") {
		t.Error("Expected AssertContains to return false")
	}

	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error, got %d", mockT.errorCount)
	}

	// Reset mock
	mockT.errorCount = 0

	// Test AssertNotContains
	if !helper.AssertNotContains(mockT, "hello world", "foo") {
		t.Error("Expected AssertNotContains to return true")
	}

	if mockT.errorCount != 0 {
		t.Errorf("Expected no errors, got %d", mockT.errorCount)
	}

	// Test AssertNotContains failure
	if helper.AssertNotContains(mockT, "hello world", "world") {
		t.Error("Expected AssertNotContains to return false")
	}

	if mockT.errorCount != 1 {
		t.Errorf("Expected 1 error, got %d", mockT.errorCount)
	}
}

// TestFrameworkDataManipulation tests adding data to the framework
func TestFrameworkDataManipulation(t *testing.T) {
	tf := NewTestFramework()

	// Add tasks
	task := beads.Task{ID: "1", Title: "Test Task", Status: "open"}
	tf.AddTask(task)

	// Add messages
	msg := mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeMessage,
		Source:    "test-agent",
		Content:   "Test message",
	}
	tf.AddMessage(msg)

	// Set agent status
	status := mcp.AgentStatus{
		Name:  "test-agent",
		State: mcp.StateWorking,
	}
	tf.SetAgentStatus(status)

	// Refresh data
	model := tf.RefreshData()

	// Verify data was loaded
	if len(model.tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(model.tasks))
	}

	if len(model.messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(model.messages))
	}

	if len(model.agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(model.agents))
	}
}

// TestFrameworkKeySequence tests a sequence of key presses
func TestFrameworkKeySequence(t *testing.T) {
	tf := NewTestFramework()

	// Add some tasks
	tf.AddTask(beads.Task{ID: "1", Title: "Task 1", Status: "open"})
	tf.AddTask(beads.Task{ID: "2", Title: "Task 2", Status: "in_progress"})
	tf.RefreshData()

	// Navigate down
	tf.SendKey(tea.KeyDown)
	model := tf.GetModel()
	if model.selectedTaskIndex != 1 {
		t.Errorf("Expected selectedTaskIndex to be 1, got %d", model.selectedTaskIndex)
	}

	// Navigate up
	tf.SendKey(tea.KeyUp)
	model = tf.GetModel()
	if model.selectedTaskIndex != 0 {
		t.Errorf("Expected selectedTaskIndex to be 0, got %d", model.selectedTaskIndex)
	}

	// Open modal
	tf.SendKeyRune('v')
	model = tf.GetModel()
	if !model.showTaskModal {
		t.Error("Expected task modal to be open")
	}

	// Close modal
	tf.SendKeyRune('v')
	model = tf.GetModel()
	if model.showTaskModal {
		t.Error("Expected task modal to be closed")
	}
}

// TestFrameworkStringInput tests typing a string
func TestFrameworkStringInput(t *testing.T) {
	tf := NewTestFramework()

	// Open create task modal
	tf.SendKeyRune('n')
	model := tf.GetModel()
	if !model.showCreateModal {
		t.Error("Expected create modal to be open")
	}

	// Type a task title
	tf.SendKeyString("New Task Title")
	model = tf.GetModel()

	// Note: The actual input handling depends on the Update function implementation
	// This test verifies the framework can send the key sequence
	if model.showCreateModal != true {
		t.Error("Expected create modal to still be open after typing")
	}
}

// mockTestingT is a mock implementation of testing.T for testing assertions
type mockTestingT struct {
	errorCount int
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errorCount++
}
