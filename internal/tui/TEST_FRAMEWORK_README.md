# TUI Integration Test Framework

This document describes the TUI integration test framework for the Agent Stack Controller.

## Overview

The TUI integration test framework provides a comprehensive set of utilities for testing Terminal User Interface components in isolation and integration. It includes mock implementations of all external dependencies, helper functions for simulating user interactions, and utilities for asserting on rendered output.

## Components

### TestFramework

The main test framework class that orchestrates testing of TUI components.

```go
tf := NewTestFramework()
```

**Features:**
- Pre-configured with mock clients and default configuration
- Simulates terminal dimensions (default: 120x40)
- Provides methods for simulating user input
- Manages model state across interactions

**Methods:**

- `GetModel() *Model` - Returns the current model state
- `SendKey(key tea.KeyType) Model` - Simulates a key press
- `SendKeyRune(r rune) Model` - Simulates a rune key press
- `SendKeyString(s string) Model` - Simulates typing a string
- `Resize(width, height int) Model` - Simulates terminal resize
- `Tick() Model` - Simulates a tick event
- `Render() string` - Returns the current view output
- `AddTask(task beads.Task)` - Adds a task to mock beads client
- `AddMessage(msg mcp.Message)` - Adds a message to mock MCP client
- `SetAgentStatus(status mcp.AgentStatus)` - Sets agent status in mock MCP client
- `RefreshData() Model` - Forces a data refresh

### Mock Clients

#### MockBeadsClient

Mock implementation of the BeadsClient interface for testing task management.

```go
client := NewMockBeadsClient()
client.AddTask(beads.Task{ID: "1", Title: "Test", Status: "open"})
tasks, _ := client.GetTasks([]string{"open"})
```

**Supported Operations:**
- `GetTasks(statuses []string)` - Filter tasks by status
- `CreateTask(title string)` - Create a new task
- `UpdateTask(id string, updates TaskUpdate)` - Update task fields
- `DeleteTask(id string)` - Delete a task
- `Refresh()` - No-op for testing

#### MockMCPClient

Mock implementation of the MCPClient interface for testing agent communication.

```go
client := NewMockMCPClient()
client.SetAgentStatus(mcp.AgentStatus{Name: "agent1", State: mcp.StateWorking})
status, _ := client.GetAgentStatus("agent1")
```

**Supported Operations:**
- `GetMessages(since time.Time)` - Get messages since timestamp
- `SendMessage(msg Message)` - Send a message
- `GetAgentStatus(agentName string)` - Get single agent status
- `GetAllAgentStatuses(offlineThreshold time.Duration)` - Get all agent statuses
- `ReleaseAgentLeases(agentName string)` - Release agent leases

#### MockProcessManager

Mock implementation of the ProcessManager interface for testing process lifecycle.

```go
pm := NewMockProcessManager()
pid, _ := pm.Start("agent1", "python", []string{"agent.py"}, []string{})
pm.IsRunning(pid) // true
```

**Supported Operations:**
- `Start(name, command string, args, env []string)` - Start a process
- `Stop(pid int)` - Stop a process
- `StopAll()` - Stop all processes
- `IsRunning(pid int)` - Check if process is running
- `GetStatus(pid int)` - Get process status
- `GetProcessInfo(name string)` - Get process information
- `ListProcesses()` - List all processes

### MockTerminal

Simulates a terminal for testing rendering output.

```go
terminal := NewMockTerminal(100, 30)
output := terminal.Render(model)
```

**Methods:**
- `Render(m Model) string` - Render model to string
- `GetOutput() string` - Get last rendered output
- `GetWidth() int` - Get terminal width
- `GetHeight() int` - Get terminal height

### TestHelper

Provides utility functions for creating test data and assertions.

```go
helper := NewTestHelper()
task := helper.CreateTestTask("1", "Test Task", "open")
```

**Factory Methods:**
- `CreateTestTask(id, title, status string)` - Create a test task
- `CreateTestMessage(msgType, source, content string)` - Create a test message
- `CreateTestAgentStatus(name string, state AgentState)` - Create agent status
- `CreateTestConfig()` - Create a test configuration

**Assertion Methods:**
- `AssertContains(t, haystack, needle string)` - Assert string contains substring
- `AssertNotContains(t, haystack, needle string)` - Assert string doesn't contain substring

## Usage Examples

### Basic Test Setup

```go
func TestMyFeature(t *testing.T) {
    // Create test framework
    tf := NewTestFramework()
    
    // Add test data
    tf.AddTask(beads.Task{ID: "1", Title: "Test Task", Status: "open"})
    tf.RefreshData()
    
    // Simulate user interaction
    tf.SendKeyRune('v') // Open task modal
    
    // Get model state
    model := tf.GetModel()
    
    // Assert on state
    if !model.showTaskModal {
        t.Error("Expected task modal to be open")
    }
}
```

### Testing Keyboard Navigation

```go
func TestTaskNavigation(t *testing.T) {
    tf := NewTestFramework()
    
    // Add multiple tasks
    tf.AddTask(beads.Task{ID: "1", Title: "Task 1", Status: "open"})
    tf.AddTask(beads.Task{ID: "2", Title: "Task 2", Status: "open"})
    tf.RefreshData()
    
    // Navigate down
    tf.SendKey(tea.KeyDown)
    model := tf.GetModel()
    
    if model.selectedTaskIndex != 1 {
        t.Errorf("Expected index 1, got %d", model.selectedTaskIndex)
    }
    
    // Navigate up
    tf.SendKey(tea.KeyUp)
    model = tf.GetModel()
    
    if model.selectedTaskIndex != 0 {
        t.Errorf("Expected index 0, got %d", model.selectedTaskIndex)
    }
}
```

### Testing Rendering

```go
func TestAgentPaneRendering(t *testing.T) {
    tf := NewTestFramework()
    helper := NewTestHelper()
    
    // Set up agent status
    tf.SetAgentStatus(mcp.AgentStatus{
        Name:  "test-agent",
        State: mcp.StateWorking,
    })
    tf.RefreshData()
    
    // Render
    output := tf.Render()
    
    // Assert on output
    helper.AssertContains(t, output, "test-agent")
    helper.AssertContains(t, output, "Working")
}
```

### Testing Modal Interactions

```go
func TestCreateTaskModal(t *testing.T) {
    tf := NewTestFramework()
    
    // Open create modal
    tf.SendKeyRune('n')
    model := tf.GetModel()
    
    if !model.showCreateModal {
        t.Error("Expected create modal to be open")
    }
    
    // Type task title
    tf.SendKeyString("New Task")
    
    // Submit (implementation dependent)
    tf.SendKey(tea.KeyEnter)
    
    // Verify task was created
    model = tf.GetModel()
    // Add assertions based on your implementation
}
```

### Testing Terminal Resize

```go
func TestResponsiveLayout(t *testing.T) {
    tf := NewTestFramework()
    
    // Test with small terminal
    tf.Resize(80, 24)
    output := tf.Render()
    
    if len(output) == 0 {
        t.Error("Expected output for small terminal")
    }
    
    // Test with large terminal
    tf.Resize(200, 60)
    output = tf.Render()
    
    if len(output) == 0 {
        t.Error("Expected output for large terminal")
    }
}
```

### Testing Data Refresh

```go
func TestDataRefresh(t *testing.T) {
    tf := NewTestFramework()
    
    // Add initial data
    tf.AddTask(beads.Task{ID: "1", Title: "Task 1", Status: "open"})
    tf.RefreshData()
    
    model := tf.GetModel()
    if len(model.tasks) != 1 {
        t.Errorf("Expected 1 task, got %d", len(model.tasks))
    }
    
    // Add more data
    tf.AddTask(beads.Task{ID: "2", Title: "Task 2", Status: "open"})
    tf.RefreshData()
    
    model = tf.GetModel()
    if len(model.tasks) != 2 {
        t.Errorf("Expected 2 tasks, got %d", len(model.tasks))
    }
}
```

### Testing Agent Status Updates

```go
func TestAgentStatusUpdates(t *testing.T) {
    tf := NewTestFramework()
    
    // Set initial status
    tf.SetAgentStatus(mcp.AgentStatus{
        Name:  "agent1",
        State: mcp.StateIdle,
    })
    tf.RefreshData()
    
    model := tf.GetModel()
    if len(model.agents) != 1 {
        t.Fatalf("Expected 1 agent, got %d", len(model.agents))
    }
    
    if model.agents[0].State != mcp.StateIdle {
        t.Errorf("Expected Idle state, got %v", model.agents[0].State)
    }
    
    // Update status
    tf.SetAgentStatus(mcp.AgentStatus{
        Name:  "agent1",
        State: mcp.StateWorking,
    })
    tf.RefreshData()
    
    model = tf.GetModel()
    if model.agents[0].State != mcp.StateWorking {
        t.Errorf("Expected Working state, got %v", model.agents[0].State)
    }
}
```

## Best Practices

### 1. Use TestFramework for Integration Tests

The TestFramework is designed for integration testing of TUI components. Use it when you need to test:
- User interaction flows
- State transitions
- Rendering output
- Multiple components working together

### 2. Use Mock Clients Directly for Unit Tests

For unit testing individual functions, you can use the mock clients directly:

```go
func TestSomeFunction(t *testing.T) {
    client := NewMockBeadsClient()
    client.AddTask(beads.Task{ID: "1", Title: "Test", Status: "open"})
    
    // Test your function with the mock client
    result := SomeFunction(client)
    
    // Assert on result
}
```

### 3. Use TestHelper for Common Patterns

The TestHelper provides factory methods and assertions to reduce boilerplate:

```go
func TestWithHelper(t *testing.T) {
    helper := NewTestHelper()
    
    task := helper.CreateTestTask("1", "Test", "open")
    msg := helper.CreateTestMessage(mcp.TypeMessage, "agent", "content")
    
    output := "some rendered output"
    helper.AssertContains(t, output, "expected text")
}
```

### 4. Test State Transitions

Always verify that state transitions work correctly:

```go
func TestStateTransition(t *testing.T) {
    tf := NewTestFramework()
    
    // Initial state
    model := tf.GetModel()
    if model.showTaskModal {
        t.Error("Modal should be closed initially")
    }
    
    // Trigger transition
    tf.SendKeyRune('v')
    model = tf.GetModel()
    
    // Verify new state
    if !model.showTaskModal {
        t.Error("Modal should be open after 'v' key")
    }
}
```

### 5. Test Error Handling

Test how the TUI handles errors and edge cases:

```go
func TestErrorHandling(t *testing.T) {
    tf := NewTestFramework()
    
    // Simulate error condition
    // (implementation specific)
    
    // Verify error is handled gracefully
    output := tf.Render()
    if output == "" {
        t.Error("TUI should still render on error")
    }
}
```

## Running Tests

Run all TUI tests:
```bash
go test ./internal/tui -v
```

Run specific test:
```bash
go test ./internal/tui -v -run TestFrameworkCreation
```

Run with coverage:
```bash
go test ./internal/tui -v -cover
```

## Coverage Goals

The test framework is designed to help achieve the following coverage targets:

- **wizard.go**: 60%+ coverage
- **view.go, agents.go, tasks.go, logs.go**: 40%+ coverage
- **update.go, modals.go**: 40%+ coverage
- **theme.go, animations.go, performance.go**: 30%+ coverage

## Extending the Framework

### Adding New Mock Methods

To add new methods to mock clients:

1. Add the method to the mock struct
2. Implement the method logic
3. Add tests for the new method
4. Update this documentation

### Adding New Test Utilities

To add new test utilities:

1. Add the utility function to TestHelper
2. Add tests for the utility
3. Document the utility in this README
4. Provide usage examples

## Troubleshooting

### Tests Fail with "nil pointer dereference"

Make sure to call `RefreshData()` after adding test data:

```go
tf.AddTask(task)
tf.RefreshData() // Important!
```

### Rendering Output is Empty

Ensure the model has been initialized with proper dimensions:

```go
tf := NewTestFramework() // Sets default 120x40
// or
tf.Resize(100, 30)
```

### Mock Data Not Appearing

Verify you're using the correct mock client methods:

```go
// Correct
tf.AddTask(task)
tf.RefreshData()

// Incorrect - bypasses framework
tf.beadsClient.AddTask(task) // Don't access directly
```

## Related Documentation

- [Bubbletea Testing Guide](https://github.com/charmbracelet/bubbletea#testing)
- [Go Testing Package](https://pkg.go.dev/testing)
- [TUI Design Document](../../docs/VAPORWAVE_DESIGN.md)
- [Integration Testing Guide](../../docs/INTEGRATION_TESTING.md)
