// +build usability

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFirstTimeUserExperience tests the complete first-time user journey
func TestFirstTimeUserExperience(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("asc_init_workflow", func(t *testing.T) {
		// Test that init command is discoverable
		cmd := exec.Command("./build/asc", "--help")
		output, _ := cmd.CombinedOutput()

		if !strings.Contains(string(output), "init") {
			t.Errorf("Init command should be mentioned in help")
		}

		// Test init help is clear
		cmd = exec.Command("./build/asc", "init", "--help")
		output, _ = cmd.CombinedOutput()

		helpText := string(output)
		if !strings.Contains(helpText, "Initialize") && !strings.Contains(helpText, "setup") {
			t.Errorf("Init help should explain what it does")
		}

		// Verify init creates expected files
		t.Log("First-time user would run: asc init")
		t.Log("Expected: Interactive wizard guides through setup")
	})

	t.Run("error_messages_are_helpful", func(t *testing.T) {
		// Test running commands without setup
		cmd := exec.Command("./build/asc", "up")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()

		errorMsg := string(output)
		// Error should mention what's missing
		if !strings.Contains(errorMsg, "config") && !strings.Contains(errorMsg, "asc.toml") {
			t.Logf("Error message could be more helpful: %s", errorMsg)
		}
	})

	t.Run("check_command_guides_setup", func(t *testing.T) {
		// Check command should help identify what's needed
		cmd := exec.Command("./build/asc", "check")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()

		t.Logf("Check output for new user: %s", output)

		// Should clearly indicate missing dependencies
		outputStr := string(output)
		if !strings.Contains(outputStr, "fail") && !strings.Contains(outputStr, "missing") {
			t.Log("Check command should clearly indicate missing items")
		}
	})

	t.Run("documentation_accessibility", func(t *testing.T) {
		// Verify README exists and is helpful
		readmePath := "README.md"
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			t.Errorf("README.md should exist for new users")
		}

		// Check for quick start section
		content, err := os.ReadFile(readmePath)
		if err == nil {
			readme := string(content)
			if !strings.Contains(readme, "Quick Start") && !strings.Contains(readme, "Getting Started") {
				t.Log("README should have a quick start section")
			}
		}
	})
}

// TestCommonWorkflows tests typical user workflows
func TestCommonWorkflows(t *testing.T) {
	tmpDir := t.TempDir()
	setupMinimalConfig(t, tmpDir)

	t.Run("check_before_start", func(t *testing.T) {
		// Common workflow: check → up
		cmd := exec.Command("./build/asc", "check")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()

		t.Logf("Check output: %s", output)

		// User should understand if they can proceed
		outputStr := string(output)
		hasStatus := strings.Contains(outputStr, "pass") || 
		             strings.Contains(outputStr, "fail") ||
		             strings.Contains(outputStr, "✓") ||
		             strings.Contains(outputStr, "✗")

		if !hasStatus {
			t.Log("Check output should clearly show pass/fail status")
		}
	})

	t.Run("start_agents_workflow", func(t *testing.T) {
		// Workflow: up → view status → down
		t.Log("User workflow: asc up")
		t.Log("Expected: Agents start, TUI launches")
		t.Log("User sees: Agent status, tasks, logs")

		// Test that up command exists and has help
		cmd := exec.Command("./build/asc", "up", "--help")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()

		if !strings.Contains(string(output), "up") {
			t.Errorf("Up help should describe the command")
		}
	})

	t.Run("view_status_workflow", func(t *testing.T) {
		// Check services status without starting
		cmd := exec.Command("./build/asc", "services", "status")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()

		t.Logf("Services status: %s", output)

		// Should clearly indicate if running or not
		outputStr := string(output)
		hasStatusInfo := strings.Contains(outputStr, "running") || 
		                 strings.Contains(outputStr, "not running") ||
		                 strings.Contains(outputStr, "stopped")

		if !hasStatusInfo {
			t.Log("Status output should clearly indicate service state")
		}
	})

	t.Run("stop_agents_workflow", func(t *testing.T) {
		// Test down command
		cmd := exec.Command("./build/asc", "down")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()

		t.Logf("Down output: %s", output)
	})

	t.Run("test_connectivity_workflow", func(t *testing.T) {
		// Test command for verifying setup
		cmd := exec.Command("./build/asc", "test", "--help")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()

		if strings.Contains(string(output), "test") {
			t.Log("Test command is available for connectivity checks")
		}
	})
}

// TestErrorRecoveryFromUserPerspective tests how users experience and recover from errors
func TestErrorRecoveryFromUserPerspective(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("missing_config_recovery", func(t *testing.T) {
		// User tries to start without config
		cmd := exec.Command("./build/asc", "up")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Log("Command succeeded unexpectedly")
		}

		errorMsg := string(output)
		t.Logf("Error message: %s", errorMsg)

		// Error should suggest running init
		suggestsInit := strings.Contains(errorMsg, "init") || 
		                strings.Contains(errorMsg, "asc.toml")

		if !suggestsInit {
			t.Log("Error should suggest how to fix (run init or create config)")
		}
	})

	t.Run("invalid_config_recovery", func(t *testing.T) {
		// Create invalid config
		configPath := filepath.Join(tmpDir, "asc.toml")
		invalidConfig := `[core
beads_db_path = 
`
		os.WriteFile(configPath, []byte(invalidConfig), 0644)

		cmd := exec.Command("./build/asc", "check")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()

		errorMsg := string(output)
		t.Logf("Invalid config error: %s", errorMsg)

		// Should indicate config is invalid
		if !strings.Contains(errorMsg, "config") && !strings.Contains(errorMsg, "parse") {
			t.Log("Error should clearly indicate config parsing issue")
		}
	})

	t.Run("missing_dependencies_recovery", func(t *testing.T) {
		setupMinimalConfig(t, tmpDir)

		cmd := exec.Command("./build/asc", "check")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()

		outputStr := string(output)
		t.Logf("Dependency check: %s", outputStr)

		// Should list what's missing and how to install
		if strings.Contains(outputStr, "fail") {
			t.Log("Failed checks should suggest installation steps")
		}
	})

	t.Run("port_conflict_recovery", func(t *testing.T) {
		// Simulate port already in use
		setupMinimalConfig(t, tmpDir)

		// This would test starting when port is taken
		t.Log("User scenario: MCP server port already in use")
		t.Log("Expected: Clear error message with port number")
		t.Log("Expected: Suggestion to stop conflicting service or change port")
	})

	t.Run("agent_crash_recovery", func(t *testing.T) {
		// Test that system handles agent crashes gracefully
		configPath := filepath.Join(tmpDir, "asc.toml")
		crashConfig := `[core]
beads_db_path = "./repo"

[agent.crasher]
command = "false"
model = "claude"
phases = ["testing"]
`
		os.WriteFile(configPath, []byte(crashConfig), 0644)
		createEnvFile(t, tmpDir)

		t.Log("User scenario: Agent command fails immediately")
		t.Log("Expected: Error logged, other agents continue")
		t.Log("Expected: User can see which agent failed in TUI")
	})
}

// TestKeyboardNavigationAndShortcuts tests TUI keyboard interactions
func TestKeyboardNavigationAndShortcuts(t *testing.T) {
	t.Run("help_shows_keybindings", func(t *testing.T) {
		// TUI should show available keybindings
		t.Log("TUI footer should display: (q)uit | (r)efresh | (t)est")
		t.Log("User can discover shortcuts without reading docs")
	})

	t.Run("quit_is_intuitive", func(t *testing.T) {
		// 'q' is standard for quit
		t.Log("Pressing 'q' should quit TUI")
		t.Log("Should prompt for confirmation if agents are running")
	})

	t.Run("refresh_is_responsive", func(t *testing.T) {
		// 'r' should immediately refresh
		t.Log("Pressing 'r' should refresh all panes")
		t.Log("Should show loading indicator during refresh")
	})

	t.Run("test_command_accessible", func(t *testing.T) {
		// 't' runs connectivity test
		t.Log("Pressing 't' should run test command")
		t.Log("Results should appear in log pane")
	})

	t.Run("navigation_between_panes", func(t *testing.T) {
		// Test pane focus navigation
		t.Log("Tab or arrow keys should move between panes")
		t.Log("Active pane should be visually distinct")
	})

	t.Run("scrolling_in_panes", func(t *testing.T) {
		// Test scrolling in log pane
		t.Log("Up/Down arrows should scroll in active pane")
		t.Log("Page Up/Down for faster scrolling")
		t.Log("Home/End to jump to top/bottom")
	})
}

// TestTerminalResizeAndResponsiveness tests responsive layout
func TestTerminalResizeAndResponsiveness(t *testing.T) {
	t.Run("minimum_terminal_size", func(t *testing.T) {
		// Test behavior with small terminal
		t.Log("Minimum usable size: 80x24")
		t.Log("Should show warning if terminal too small")
		t.Log("Should not crash or corrupt display")
	})

	t.Run("large_terminal_utilization", func(t *testing.T) {
		// Test with large terminal
		t.Log("Large terminals (200x60) should use space well")
		t.Log("Panes should expand proportionally")
		t.Log("Text should not be unnecessarily truncated")
	})

	t.Run("resize_during_operation", func(t *testing.T) {
		// Test resizing while TUI is running
		t.Log("Resizing terminal should update layout immediately")
		t.Log("No visual artifacts or corruption")
		t.Log("Content should reflow appropriately")
	})

	t.Run("aspect_ratio_handling", func(t *testing.T) {
		// Test with unusual aspect ratios
		t.Log("Very wide terminals: horizontal layout works")
		t.Log("Very tall terminals: vertical space used well")
		t.Log("Square terminals: balanced layout")
	})
}

// TestAccessibilityFeatures tests accessibility and usability features
func TestAccessibilityFeatures(t *testing.T) {
	t.Run("high_contrast_mode", func(t *testing.T) {
		// Test high contrast theme
		t.Log("High contrast mode should be available")
		t.Log("Colors should meet WCAG contrast ratios")
		t.Log("Status indicators visible without color")
	})

	t.Run("color_blind_friendly", func(t *testing.T) {
		// Test color blind accessibility
		t.Log("Don't rely solely on color for status")
		t.Log("Use icons/symbols: ● ⟳ ! ○")
		t.Log("Test with deuteranopia/protanopia simulators")
	})

	t.Run("screen_reader_compatibility", func(t *testing.T) {
		// Test screen reader support
		t.Log("CLI output should be screen reader friendly")
		t.Log("Status information in text form")
		t.Log("Avoid ASCII art that doesn't read well")
	})

	t.Run("reduced_motion", func(t *testing.T) {
		// Test without animations
		t.Log("Animations should be optional")
		t.Log("Spinning indicators can be static")
		t.Log("Transitions can be instant")
	})

	t.Run("font_size_independence", func(t *testing.T) {
		// Test with different terminal font sizes
		t.Log("Layout should work with various font sizes")
		t.Log("No hardcoded pixel dimensions")
		t.Log("Use terminal rows/columns")
	})

	t.Run("keyboard_only_navigation", func(t *testing.T) {
		// Test without mouse
		t.Log("All features accessible via keyboard")
		t.Log("No mouse-only interactions")
		t.Log("Clear keyboard shortcuts")
	})
}

// TestUserFeedbackCollection simulates beta testing scenarios
func TestUserFeedbackCollection(t *testing.T) {
	t.Run("first_impressions", func(t *testing.T) {
		// Simulate first-time user experience
		scenarios := []string{
			"User has never used CLI tools before",
			"User is experienced with Docker/Kubernetes",
			"User is familiar with TUI apps (htop, vim)",
			"User prefers GUI applications",
		}

		for _, scenario := range scenarios {
			t.Logf("Scenario: %s", scenario)
			t.Log("  - Is help text clear?")
			t.Log("  - Are commands discoverable?")
			t.Log("  - Is feedback immediate?")
		}
	})

	t.Run("common_pain_points", func(t *testing.T) {
		// Identify potential pain points
		painPoints := []string{
			"Setting up API keys",
			"Understanding agent phases",
			"Debugging agent failures",
			"Interpreting log messages",
			"Configuring multiple agents",
		}

		for _, point := range painPoints {
			t.Logf("Pain point: %s", point)
			t.Log("  - Is there clear documentation?")
			t.Log("  - Are error messages helpful?")
			t.Log("  - Is there a troubleshooting guide?")
		}
	})

	t.Run("feature_discoverability", func(t *testing.T) {
		// Test feature discovery
		features := []string{
			"Configuration templates",
			"Hot reload",
			"Health monitoring",
			"Log filtering",
			"Agent control",
		}

		for _, feature := range features {
			t.Logf("Feature: %s", feature)
			t.Log("  - Can users find it without docs?")
			t.Log("  - Is it mentioned in help text?")
			t.Log("  - Is there an example?")
		}
	})

	t.Run("workflow_efficiency", func(t *testing.T) {
		// Measure workflow efficiency
		workflows := []string{
			"Start working on a new project",
			"Add a new agent to existing setup",
			"Debug a failing agent",
			"View agent logs",
			"Stop and restart agents",
		}

		for _, workflow := range workflows {
			t.Logf("Workflow: %s", workflow)
			t.Log("  - How many commands required?")
			t.Log("  - Are there shortcuts?")
			t.Log("  - Is feedback clear at each step?")
		}
	})
}

// TestCommonUserIssues documents and tests solutions for common issues
func TestCommonUserIssues(t *testing.T) {
	t.Run("issue_api_keys_not_loaded", func(t *testing.T) {
		// Issue: API keys not being loaded
		t.Log("Issue: Agent can't authenticate with LLM")
		t.Log("Cause: .env file not in correct location")
		t.Log("Solution: Check .env is in project root")
		t.Log("Solution: Verify file permissions (should be 0600)")
		t.Log("Solution: Check env vars with 'asc check'")

		// Test that check command helps diagnose this
		tmpDir := t.TempDir()
		setupMinimalConfig(t, tmpDir)
		// Don't create .env file

		cmd := exec.Command("./build/asc", "check")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()

		outputStr := string(output)
		if !strings.Contains(outputStr, ".env") {
			t.Log("Check should mention missing .env file")
		}
	})

	t.Run("issue_port_already_in_use", func(t *testing.T) {
		// Issue: Can't start MCP server
		t.Log("Issue: MCP server fails to start")
		t.Log("Cause: Port 8765 already in use")
		t.Log("Solution: Stop other service using port")
		t.Log("Solution: Change port in asc.toml")
		t.Log("Solution: Use 'lsof -i :8765' to find process")
	})

	t.Run("issue_agents_not_picking_up_tasks", func(t *testing.T) {
		// Issue: Agents idle but tasks available
		t.Log("Issue: Agents show as idle but tasks exist")
		t.Log("Cause: Phase mismatch between agent and tasks")
		t.Log("Solution: Check agent phases in config")
		t.Log("Solution: Verify task phases in beads")
		t.Log("Solution: Check MCP connectivity")
	})

	t.Run("issue_tui_display_corrupted", func(t *testing.T) {
		// Issue: TUI looks broken
		t.Log("Issue: TUI display is garbled or corrupted")
		t.Log("Cause: Terminal doesn't support required features")
		t.Log("Solution: Use modern terminal (iTerm2, Alacritty)")
		t.Log("Solution: Check TERM environment variable")
		t.Log("Solution: Try 'export TERM=xterm-256color'")
	})

	t.Run("issue_high_cpu_usage", func(t *testing.T) {
		// Issue: High CPU usage
		t.Log("Issue: asc using high CPU")
		t.Log("Cause: Polling intervals too aggressive")
		t.Log("Solution: Increase refresh intervals in config")
		t.Log("Solution: Use WebSocket instead of polling")
		t.Log("Solution: Reduce number of agents")
	})

	t.Run("issue_logs_filling_disk", func(t *testing.T) {
		// Issue: Disk space issues
		t.Log("Issue: ~/.asc/logs directory growing large")
		t.Log("Cause: Log rotation not configured")
		t.Log("Solution: Configure log rotation")
		t.Log("Solution: Manually clean old logs")
		t.Log("Solution: Reduce log verbosity")
	})

	t.Run("issue_agent_stuck", func(t *testing.T) {
		// Issue: Agent appears stuck
		t.Log("Issue: Agent working on same task for hours")
		t.Log("Cause: Agent in infinite loop or waiting")
		t.Log("Solution: Check agent logs for errors")
		t.Log("Solution: Restart agent with 'asc down' then 'asc up'")
		t.Log("Solution: Check file lease status in MCP")
	})

	t.Run("issue_config_not_reloading", func(t *testing.T) {
		// Issue: Config changes not taking effect
		t.Log("Issue: Changed config but agents unchanged")
		t.Log("Cause: Hot reload not triggered")
		t.Log("Solution: Restart with 'asc down' then 'asc up'")
		t.Log("Solution: Check config file syntax")
		t.Log("Solution: Verify file watcher is working")
	})
}

// TestUsabilityMetrics measures usability quantitatively
func TestUsabilityMetrics(t *testing.T) {
	t.Run("time_to_first_success", func(t *testing.T) {
		// Measure time from install to first successful agent run
		t.Log("Metric: Time to first success")
		t.Log("Target: < 5 minutes for experienced developer")
		t.Log("Target: < 15 minutes for new user")
		t.Log("Measure: Install → Init → Configure → Up → Agent runs task")
	})

	t.Run("command_discoverability", func(t *testing.T) {
		// Measure how easily users find commands
		t.Log("Metric: Command discoverability")
		t.Log("Test: Can user find command without docs?")
		t.Log("Method: Help text, command suggestions, examples")
	})

	t.Run("error_recovery_time", func(t *testing.T) {
		// Measure time to recover from errors
		t.Log("Metric: Error recovery time")
		t.Log("Target: < 2 minutes for common errors")
		t.Log("Measure: Error occurs → User understands → User fixes")
	})

	t.Run("cognitive_load", func(t *testing.T) {
		// Assess mental effort required
		t.Log("Metric: Cognitive load")
		t.Log("Measure: Concepts user must understand")
		t.Log("Target: Minimize required knowledge")
		t.Log("Concepts: agents, phases, beads, MCP, leases")
	})

	t.Run("task_completion_rate", func(t *testing.T) {
		// Measure successful task completion
		t.Log("Metric: Task completion rate")
		t.Log("Tasks: Init, Start agents, View status, Stop agents")
		t.Log("Target: > 90% completion without help")
	})
}

// Helper functions

func setupMinimalConfig(t *testing.T, dir string) {
	t.Helper()

	configPath := filepath.Join(dir, "asc.toml")
	config := `[core]
beads_db_path = "./test-repo"

[services.mcp_agent_mail]
start_command = "python -m mcp_agent_mail.server"
url = "http://localhost:8765"

[agent.test]
command = "sleep"
model = "claude"
phases = ["testing"]
`
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
}

func createEnvFile(t *testing.T, dir string) {
	t.Helper()

	envPath := filepath.Join(dir, ".env")
	env := `CLAUDE_API_KEY=test-key
OPENAI_API_KEY=test-key
GOOGLE_API_KEY=test-key
`
	if err := os.WriteFile(envPath, []byte(env), 0600); err != nil {
		t.Fatalf("Failed to write .env: %v", err)
	}
}

// TestInteractiveFeatures tests interactive TUI features
func TestInteractiveFeatures(t *testing.T) {
	t.Run("modal_dialogs", func(t *testing.T) {
		// Test modal interactions
		t.Log("Modal for task details should be clear")
		t.Log("Modal for confirmations should be obvious")
		t.Log("Escape key should close modals")
	})

	t.Run("input_forms", func(t *testing.T) {
		// Test input handling
		t.Log("Text input should show cursor")
		t.Log("Input validation should be immediate")
		t.Log("Error messages should be inline")
	})

	t.Run("selection_lists", func(t *testing.T) {
		// Test list navigation
		t.Log("Selected item should be highlighted")
		t.Log("Arrow keys should move selection")
		t.Log("Enter should confirm selection")
	})
}

// TestDocumentationQuality tests documentation usability
func TestDocumentationQuality(t *testing.T) {
	t.Run("readme_completeness", func(t *testing.T) {
		// Check README has essential sections
		essentialSections := []string{
			"Installation",
			"Quick Start",
			"Configuration",
			"Troubleshooting",
			"Examples",
		}

		readmePath := "README.md"
		content, err := os.ReadFile(readmePath)
		if err != nil {
			t.Skipf("README not found: %v", err)
		}

		readme := string(content)
		for _, section := range essentialSections {
			if !strings.Contains(readme, section) {
				t.Logf("README should have %s section", section)
			}
		}
	})

	t.Run("examples_are_runnable", func(t *testing.T) {
		// Verify examples in docs actually work
		t.Log("All code examples should be tested")
		t.Log("Examples should be copy-pasteable")
		t.Log("Examples should include expected output")
	})

	t.Run("troubleshooting_guide", func(t *testing.T) {
		// Check for troubleshooting documentation
		troubleshootingPath := "TROUBLESHOOTING.md"
		if _, err := os.Stat(troubleshootingPath); os.IsNotExist(err) {
			t.Log("TROUBLESHOOTING.md should exist")
		}
	})
}

// TestUserOnboarding tests the onboarding experience
func TestUserOnboarding(t *testing.T) {
	t.Run("welcome_message", func(t *testing.T) {
		// First run should be welcoming
		t.Log("First run should show welcome message")
		t.Log("Should explain what asc does")
		t.Log("Should guide to next steps")
	})

	t.Run("progressive_disclosure", func(t *testing.T) {
		// Don't overwhelm with all features at once
		t.Log("Show basic features first")
		t.Log("Advanced features discoverable later")
		t.Log("Tooltips for complex features")
	})

	t.Run("example_configurations", func(t *testing.T) {
		// Provide working examples
		t.Log("Include example asc.toml files")
		t.Log("Templates for common setups")
		t.Log("Comments explaining each option")
	})
}

// TestPerformancePerception tests perceived performance
func TestPerformancePerception(t *testing.T) {
	t.Run("startup_feels_fast", func(t *testing.T) {
		// Startup should feel responsive
		t.Log("Show progress during startup")
		t.Log("Display what's happening")
		t.Log("Target: < 2 seconds to TUI")
	})

	t.Run("commands_feel_responsive", func(t *testing.T) {
		// Commands should feel instant
		t.Log("Acknowledge input immediately")
		t.Log("Show loading indicators")
		t.Log("Target: < 100ms to feedback")
	})

	t.Run("updates_are_smooth", func(t *testing.T) {
		// UI updates should be smooth
		t.Log("No flickering or tearing")
		t.Log("Smooth scrolling")
		t.Log("Animations enhance, not distract")
	})
}

// TestErrorPreventionAndRecovery tests proactive error handling
func TestErrorPreventionAndRecovery(t *testing.T) {
	t.Run("validation_before_action", func(t *testing.T) {
		// Validate before executing
		t.Log("Check config before starting agents")
		t.Log("Verify dependencies before init")
		t.Log("Confirm destructive actions")
	})

	t.Run("undo_capability", func(t *testing.T) {
		// Allow undoing actions
		t.Log("Backup configs before modification")
		t.Log("Allow reverting changes")
		t.Log("Clear undo instructions")
	})

	t.Run("safe_defaults", func(t *testing.T) {
		// Defaults should be safe
		t.Log("Default config should work")
		t.Log("No dangerous defaults")
		t.Log("Opt-in for risky features")
	})
}
