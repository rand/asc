package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/rand/asc/internal/beads"
	"github.com/rand/asc/internal/config"
	"github.com/rand/asc/internal/mcp"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run an end-to-end test of the agent stack",
	Long: `Run an end-to-end test to verify all components are communicating correctly.
This creates a test task in beads, sends a test message to mcp_agent_mail,
and verifies that both systems are responding properly.`,
	Run: runTest,
}

func init() {
	rootCmd.AddCommand(testCmd)
}

// runTest executes the end-to-end test flow
func runTest(cmd *cobra.Command, args []string) {
	fmt.Println("Running agent stack test...")
	fmt.Println()

	// Load configuration
	cfg, err := config.Load(config.DefaultConfigPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		fmt.Fprintf(os.Stderr, "Solution: Ensure asc.toml exists and is valid\n")
		os.Exit(1)
	}

	// Initialize clients
	beadsClient := beads.NewClient(cfg.Core.BeadsDBPath, 5*time.Second)
	mcpClient := mcp.NewHTTPClient(cfg.Services.MCPAgentMail.URL)

	// Test 1: Create test beads task
	fmt.Print("1. Creating test beads task... ")
	testTask, err := beadsClient.CreateTask("asc test task")
	if err != nil {
		fmt.Println("✗ FAILED")
		fmt.Fprintf(os.Stderr, "   Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "   Solution: Ensure 'bd' CLI is installed and beads_db_path is correct\n")
		os.Exit(1)
	}
	fmt.Printf("✓ OK (ID: %s)\n", testTask.ID)

	// Test 2: Send test message to MCP server
	fmt.Print("2. Sending test message to MCP server... ")
	testMessage := mcp.Message{
		Timestamp: time.Now(),
		Type:      mcp.TypeMessage,
		Source:    "asc-test",
		Content:   "Test message from asc test command",
	}
	err = mcpClient.SendMessage(testMessage)
	if err != nil {
		fmt.Println("✗ FAILED")
		fmt.Fprintf(os.Stderr, "   Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "   Solution: Ensure mcp_agent_mail server is running (try 'asc services start')\n")
		
		// Clean up test task before exiting
		fmt.Print("   Cleaning up test task... ")
		if cleanupErr := beadsClient.DeleteTask(testTask.ID); cleanupErr != nil {
			fmt.Printf("✗ (failed: %v)\n", cleanupErr)
		} else {
			fmt.Println("✓")
		}
		os.Exit(1)
	}
	fmt.Println("✓ OK")

	// Test 3: Poll beads to confirm task exists
	fmt.Print("3. Verifying beads task retrieval... ")
	success := false
	timeout := 30 * time.Second
	pollInterval := 1 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		tasks, err := beadsClient.GetTasks([]string{"open", "in_progress", "done"})
		if err != nil {
			fmt.Println("✗ FAILED")
			fmt.Fprintf(os.Stderr, "   Error: %v\n", err)
			
			// Clean up test task before exiting
			fmt.Print("   Cleaning up test task... ")
			if cleanupErr := beadsClient.DeleteTask(testTask.ID); cleanupErr != nil {
				fmt.Printf("✗ (failed: %v)\n", cleanupErr)
			} else {
				fmt.Println("✓")
			}
			os.Exit(1)
		}

		// Check if our test task is in the list
		for _, task := range tasks {
			if task.ID == testTask.ID {
				success = true
				break
			}
		}

		if success {
			break
		}

		time.Sleep(pollInterval)
	}

	if !success {
		fmt.Println("✗ FAILED (timeout)")
		fmt.Fprintf(os.Stderr, "   Error: Test task not found in beads database after %v\n", timeout)
		
		// Clean up test task before exiting
		fmt.Print("   Cleaning up test task... ")
		if cleanupErr := beadsClient.DeleteTask(testTask.ID); cleanupErr != nil {
			fmt.Printf("✗ (failed: %v)\n", cleanupErr)
		} else {
			fmt.Println("✓")
		}
		os.Exit(1)
	}
	fmt.Println("✓ OK")

	// Test 4: Poll MCP to confirm message was received
	fmt.Print("4. Verifying MCP message retrieval... ")
	success = false
	deadline = time.Now().Add(timeout)
	messageCheckTime := testMessage.Timestamp.Add(-1 * time.Second) // Check slightly before to ensure we catch it

	for time.Now().Before(deadline) {
		messages, err := mcpClient.GetMessages(messageCheckTime)
		if err != nil {
			fmt.Println("✗ FAILED")
			fmt.Fprintf(os.Stderr, "   Error: %v\n", err)
			
			// Clean up test artifacts before exiting
			fmt.Print("   Cleaning up test task... ")
			if cleanupErr := beadsClient.DeleteTask(testTask.ID); cleanupErr != nil {
				fmt.Printf("✗ (failed: %v)\n", cleanupErr)
			} else {
				fmt.Println("✓")
			}
			os.Exit(1)
		}

		// Check if our test message is in the list
		for _, msg := range messages {
			if msg.Source == "asc-test" && msg.Type == mcp.TypeMessage {
				success = true
				break
			}
		}

		if success {
			break
		}

		time.Sleep(pollInterval)
	}

	if !success {
		fmt.Println("✗ FAILED (timeout)")
		fmt.Fprintf(os.Stderr, "   Error: Test message not found in MCP server after %v\n", timeout)
		
		// Clean up test task before exiting
		fmt.Print("   Cleaning up test task... ")
		if cleanupErr := beadsClient.DeleteTask(testTask.ID); cleanupErr != nil {
			fmt.Printf("✗ (failed: %v)\n", cleanupErr)
		} else {
			fmt.Println("✓")
		}
		os.Exit(1)
	}
	fmt.Println("✓ OK")

	// Test 5: Clean up test artifacts
	fmt.Print("5. Cleaning up test artifacts... ")
	
	// Delete test task
	if err := beadsClient.DeleteTask(testTask.ID); err != nil {
		fmt.Println("✗ FAILED")
		fmt.Fprintf(os.Stderr, "   Error: Failed to delete test task: %v\n", err)
		fmt.Fprintf(os.Stderr, "   Note: You may need to manually delete task '%s'\n", testTask.ID)
		os.Exit(1)
	}
	
	fmt.Println("✓ OK")

	// All tests passed
	fmt.Println()
	fmt.Println("✓ Stack is healthy")
	fmt.Println()
	fmt.Println("All components are communicating correctly:")
	fmt.Println("  • beads task database is accessible")
	fmt.Println("  • mcp_agent_mail server is responding")
	fmt.Println("  • Message passing is working")
	
	os.Exit(0)
}
