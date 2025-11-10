package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RequiredAPIKeys lists the API keys that must be present in the .env file
var RequiredAPIKeys = []string{
	"CLAUDE_API_KEY",
	"OPENAI_API_KEY",
	"GOOGLE_API_KEY",
}

// DefaultEnvPath returns the default path for the .env file
func DefaultEnvPath() string {
	return ".env"
}

// LoadEnv reads the .env file and loads API keys into the environment
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

// ValidateEnv checks that all required API keys are present in the environment
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

// LoadAndValidateEnv loads the .env file and validates required keys
func LoadAndValidateEnv(envPath string) error {
	if err := LoadEnv(envPath); err != nil {
		return err
	}

	if err := ValidateEnv(); err != nil {
		return err
	}

	return nil
}
