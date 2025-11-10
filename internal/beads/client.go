package beads

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// BeadsClient defines the interface for interacting with the beads task database
type BeadsClient interface {
	GetTasks(statuses []string) ([]Task, error)
	CreateTask(title string) (Task, error)
	UpdateTask(id string, updates TaskUpdate) error
	DeleteTask(id string) error
	Refresh() error
}

// Task represents a beads task
type Task struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Phase    string `json:"phase"`
	Assignee string `json:"assignee,omitempty"`
}

// TaskUpdate represents fields that can be updated on a task
type TaskUpdate struct {
	Title    *string `json:"title,omitempty"`
	Status   *string `json:"status,omitempty"`
	Phase    *string `json:"phase,omitempty"`
	Assignee *string `json:"assignee,omitempty"`
}

// Client implements the BeadsClient interface using the bd CLI
type Client struct {
	dbPath         string
	refreshInterval time.Duration
}

// NewClient creates a new beads client
func NewClient(dbPath string, refreshInterval time.Duration) *Client {
	return &Client{
		dbPath:         dbPath,
		refreshInterval: refreshInterval,
	}
}

// GetTasks retrieves tasks filtered by status
func (c *Client) GetTasks(statuses []string) ([]Task, error) {
	args := []string{"--json", "list"}
	
	// Add status filters if provided
	if len(statuses) > 0 {
		args = append(args, "--status", strings.Join(statuses, ","))
	}
	
	cmd := exec.Command("bd", args...)
	if c.dbPath != "" {
		cmd.Dir = c.dbPath
	}
	
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("bd list failed: %w (stderr: %s)", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("bd list failed: %w", err)
	}
	
	var tasks []Task
	if len(output) > 0 {
		if err := json.Unmarshal(output, &tasks); err != nil {
			return nil, fmt.Errorf("failed to parse bd output: %w", err)
		}
	}
	
	return tasks, nil
}

// CreateTask creates a new task with the given title
func (c *Client) CreateTask(title string) (Task, error) {
	cmd := exec.Command("bd", "--json", "create", title)
	if c.dbPath != "" {
		cmd.Dir = c.dbPath
	}
	
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return Task{}, fmt.Errorf("bd create failed: %w (stderr: %s)", err, string(exitErr.Stderr))
		}
		return Task{}, fmt.Errorf("bd create failed: %w", err)
	}
	
	var task Task
	if err := json.Unmarshal(output, &task); err != nil {
		return Task{}, fmt.Errorf("failed to parse bd output: %w", err)
	}
	
	return task, nil
}

// UpdateTask updates a task with the given ID
func (c *Client) UpdateTask(id string, updates TaskUpdate) error {
	args := []string{"update", id}
	
	if updates.Title != nil {
		args = append(args, "--title", *updates.Title)
	}
	if updates.Status != nil {
		args = append(args, "--status", *updates.Status)
	}
	if updates.Phase != nil {
		args = append(args, "--phase", *updates.Phase)
	}
	if updates.Assignee != nil {
		args = append(args, "--assignee", *updates.Assignee)
	}
	
	cmd := exec.Command("bd", args...)
	if c.dbPath != "" {
		cmd.Dir = c.dbPath
	}
	
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("bd update failed: %w (stderr: %s)", err, string(exitErr.Stderr))
		}
		return fmt.Errorf("bd update failed: %w", err)
	}
	
	return nil
}

// DeleteTask deletes a task with the given ID
func (c *Client) DeleteTask(id string) error {
	cmd := exec.Command("bd", "delete", id)
	if c.dbPath != "" {
		cmd.Dir = c.dbPath
	}
	
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("bd delete failed: %w (stderr: %s)", err, string(exitErr.Stderr))
		}
		return fmt.Errorf("bd delete failed: %w", err)
	}
	
	return nil
}

// Refresh executes git pull on the beads repository
func (c *Client) Refresh() error {
	if c.dbPath == "" {
		return fmt.Errorf("dbPath not configured")
	}
	
	cmd := exec.Command("git", "pull")
	cmd.Dir = c.dbPath
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a merge conflict
		if strings.Contains(string(output), "CONFLICT") {
			return fmt.Errorf("git pull failed with merge conflict: %s", string(output))
		}
		return fmt.Errorf("git pull failed: %w (output: %s)", err, string(output))
	}
	
	return nil
}
