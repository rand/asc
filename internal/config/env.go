package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RequiredAPIKeys lists the API keys that must be present in the .env file
// for the agent stack to function properly. At least one key should be present
// depending on which LLM models are configured.
var RequiredAPIKeys = []string{
	"CLAUDE_API_KEY",
	"OPENAI_API_KEY",
	"GOOGLE_API_KEY",
}

// DefaultEnvPath returns the default path for the .env file.
// This is typically ".env" in the current working directory.
func DefaultEnvPath() string {
	return ".env"
}

// GetDefaultPIDDir returns the default directory for storing process ID files.
// This is typically "~/.asc/pids" in the user's home directory.
func GetDefaultPIDDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return fmt.Sprintf("%s/.asc/pids", home), nil
}

// GetDefaultLogDir returns the default directory for storing log files.
// This is typically "~/.asc/logs" in the user's home directory.
func GetDefaultLogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return fmt.Sprintf("%s/.asc/logs", home), nil
}

// LoadEnv reads the .env file and loads API keys into the environment.
// It parses KEY=VALUE format, skips comments and empty lines, and sets
// environment variables for each entry. Returns an error if the file
// doesn't exist or has invalid syntax.
//
// Example:
//
//	if err := config.LoadEnv(".env"); err != nil {
//	    log.Fatalf("Failed to load environment: %v", err)
//	}
func LoadEnv(envPath string) error {
	// Check if file exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf("environment file not found: %s", envPath)
	}

	// Open the file
	file, err := os.Open(envPath)
	if err != nil {
		return fmt.Errorf("failed to open environment file: %w", err)
	}
	defer file.Close()

	// Parse the file line by line
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format at line %d: expected KEY=VALUE", lineNum)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Set environment variable
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading environment file: %w", err)
	}

	return nil
}

// ValidateEnv checks that all required API keys are present in the environment.
// Returns an error listing any missing keys. This should be called after LoadEnv.
func ValidateEnv() error {
	missing := []string{}

	for _, key := range RequiredAPIKeys {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required API keys: %s", strings.Join(missing, ", "))
	}

	return nil
}

// LoadAndValidateEnv loads the .env file and validates required keys.
// This is a convenience function that combines LoadEnv and ValidateEnv.
// Returns an error if the file doesn't exist, has invalid syntax, or
// is missing required API keys.
func LoadAndValidateEnv(envPath string) error {
	if err := LoadEnv(envPath); err != nil {
		return err
	}

	if err := ValidateEnv(); err != nil {
		return err
	}

	return nil
}
