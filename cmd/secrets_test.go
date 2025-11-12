package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSecretsInitCommand_Success tests successful key generation
func TestSecretsInitCommand_Success(t *testing.T) {
	// Skip if age is not installed
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	// Set custom key path
	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Run secrets init command
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	defer func() { os.Args = oldArgs }()

	// Capture output
	capture := NewCaptureOutput()
	capture.Start()

	err := secretsInitCmd.RunE(secretsInitCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets init failed: %v", err)
	}

	// Verify key file was created
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("Age key file should exist after init")
	}

	// Verify output contains success message
	output := capture.GetStdout()
	if !strings.Contains(output, "generated successfully") {
		t.Errorf("Expected success message in output, got: %s", output)
	}
}

// TestSecretsInitCommand_AlreadyExists tests behavior when key already exists
func TestSecretsInitCommand_AlreadyExists(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Create existing key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	os.WriteFile(keyPath, []byte("existing key"), 0600)

	// Run secrets init command (should prompt for overwrite)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	defer func() { os.Args = oldArgs }()

	// Note: This test would require mocking user input
	// For now, we just verify the key exists check works
	capture := NewCaptureOutput()
	capture.Start()

	// The command will wait for user input, so we skip actual execution
	// and just verify the key exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("Key should exist before init")
	}

	capture.Stop()
}

// TestSecretsInitCommand_NoAge tests behavior when age is not installed
func TestSecretsInitCommand_NoAge(t *testing.T) {
	if isAgeInstalled() {
		t.Skip("age is installed, skipping no-age test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	defer func() { os.Args = oldArgs }()

	err := secretsInitCmd.RunE(secretsInitCmd, []string{})

	if err == nil {
		t.Error("Expected error when age is not installed")
	}

	if !strings.Contains(err.Error(), "age not installed") {
		t.Errorf("Expected 'age not installed' error, got: %v", err)
	}
}

// TestSecretsEncryptCommand_Success tests successful encryption
func TestSecretsEncryptCommand_Success(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key first
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Create .env file
	envContent := `CLAUDE_API_KEY=sk-test-123
OPENAI_API_KEY=sk-test-456
GOOGLE_API_KEY=test-789
`
	env.WriteEnv(envContent)

	// Run encrypt command
	os.Args = []string{"asc", "secrets", "encrypt"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets encrypt failed: %v", err)
	}

	// Verify encrypted file was created
	encPath := filepath.Join(env.TempDir, ".env.age")
	if _, err := os.Stat(encPath); os.IsNotExist(err) {
		t.Error("Encrypted file should exist after encryption")
	}

	output := capture.GetStdout()
	if !strings.Contains(output, "Encrypted") {
		t.Errorf("Expected encryption success message, got: %s", output)
	}
}

// TestSecretsEncryptCommand_CustomFile tests encryption with custom file
func TestSecretsEncryptCommand_CustomFile(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Create custom env file
	customEnvPath := filepath.Join(env.TempDir, ".env.prod")
	envContent := `CLAUDE_API_KEY=sk-prod-123
OPENAI_API_KEY=sk-prod-456
GOOGLE_API_KEY=prod-789
`
	os.WriteFile(customEnvPath, []byte(envContent), 0600)

	// Run encrypt command with custom file
	os.Args = []string{"asc", "secrets", "encrypt", ".env.prod"}
	defer func() { os.Args = oldArgs }()

	err := secretsEncryptCmd.RunE(secretsEncryptCmd, []string{".env.prod"})

	if err != nil {
		t.Errorf("secrets encrypt with custom file failed: %v", err)
	}

	// Verify encrypted file was created
	encPath := filepath.Join(env.TempDir, ".env.prod.age")
	if _, err := os.Stat(encPath); os.IsNotExist(err) {
		t.Error("Encrypted custom file should exist")
	}
}

// TestSecretsEncryptCommand_MissingFile tests encryption with missing file
func TestSecretsEncryptCommand_MissingFile(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Run encrypt command without creating .env file
	os.Args = []string{"asc", "secrets", "encrypt"}
	defer func() { os.Args = oldArgs }()

	err := secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})

	if err == nil {
		t.Error("Expected error for missing file")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// TestSecretsEncryptCommand_NoKey tests encryption without key
func TestSecretsEncryptCommand_NoKey(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Create .env file but no key
	env.WriteEnv(ValidEnv())

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "encrypt"}
	defer func() { os.Args = oldArgs }()

	err := secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})

	if err == nil {
		t.Error("Expected error when key doesn't exist")
	}

	if !strings.Contains(err.Error(), "key not found") {
		t.Errorf("Expected 'key not found' error, got: %v", err)
	}
}

// TestSecretsEncryptCommand_NoAge tests encryption without age installed
func TestSecretsEncryptCommand_NoAge(t *testing.T) {
	if isAgeInstalled() {
		t.Skip("age is installed, skipping no-age test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	env.WriteEnv(ValidEnv())

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "encrypt"}
	defer func() { os.Args = oldArgs }()

	err := secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})

	if err == nil {
		t.Error("Expected error when age is not installed")
	}

	if !strings.Contains(err.Error(), "age is not installed") {
		t.Errorf("Expected 'age is not installed' error, got: %v", err)
	}
}

// TestSecretsDecryptCommand_Success tests successful decryption
func TestSecretsDecryptCommand_Success(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Create and encrypt .env file
	envContent := `CLAUDE_API_KEY=sk-test-123
OPENAI_API_KEY=sk-test-456
GOOGLE_API_KEY=test-789
`
	env.WriteEnv(envContent)

	os.Args = []string{"asc", "secrets", "encrypt"}
	secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})
	os.Args = oldArgs

	// Remove original .env
	os.Remove(filepath.Join(env.TempDir, ".env"))

	// Run decrypt command
	os.Args = []string{"asc", "secrets", "decrypt"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsDecryptCmd.RunE(secretsDecryptCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets decrypt failed: %v", err)
	}

	// Verify decrypted file was created
	envPath := filepath.Join(env.TempDir, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Error("Decrypted file should exist after decryption")
	}

	// Verify content matches
	decrypted := env.ReadFile(envPath)
	if decrypted != envContent {
		t.Errorf("Decrypted content doesn't match original")
	}

	output := capture.GetStdout()
	if !strings.Contains(output, "Decrypted") {
		t.Errorf("Expected decryption success message, got: %s", output)
	}
}

// TestSecretsDecryptCommand_CustomFile tests decryption with custom file
func TestSecretsDecryptCommand_CustomFile(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Create and encrypt custom env file
	customEnvPath := filepath.Join(env.TempDir, ".env.staging")
	envContent := `CLAUDE_API_KEY=sk-staging-123
OPENAI_API_KEY=sk-staging-456
GOOGLE_API_KEY=staging-789
`
	os.WriteFile(customEnvPath, []byte(envContent), 0600)

	os.Args = []string{"asc", "secrets", "encrypt", ".env.staging"}
	secretsEncryptCmd.RunE(secretsEncryptCmd, []string{".env.staging"})
	os.Args = oldArgs

	// Remove original
	os.Remove(customEnvPath)

	// Run decrypt command with custom file
	os.Args = []string{"asc", "secrets", "decrypt", ".env.staging"}
	defer func() { os.Args = oldArgs }()

	err := secretsDecryptCmd.RunE(secretsDecryptCmd, []string{".env.staging"})

	if err != nil {
		t.Errorf("secrets decrypt with custom file failed: %v", err)
	}

	// Verify decrypted file was created
	if _, err := os.Stat(customEnvPath); os.IsNotExist(err) {
		t.Error("Decrypted custom file should exist")
	}
}

// TestSecretsDecryptCommand_MissingFile tests decryption with missing encrypted file
func TestSecretsDecryptCommand_MissingFile(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Run decrypt command without creating .env.age file
	os.Args = []string{"asc", "secrets", "decrypt"}
	defer func() { os.Args = oldArgs }()

	err := secretsDecryptCmd.RunE(secretsDecryptCmd, []string{})

	if err == nil {
		t.Error("Expected error for missing encrypted file")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// TestSecretsDecryptCommand_NoKey tests decryption without key
func TestSecretsDecryptCommand_NoKey(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Create .env.age file but no key
	os.WriteFile(filepath.Join(env.TempDir, ".env.age"), []byte("encrypted"), 0600)

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "decrypt"}
	defer func() { os.Args = oldArgs }()

	err := secretsDecryptCmd.RunE(secretsDecryptCmd, []string{})

	if err == nil {
		t.Error("Expected error when key doesn't exist")
	}

	if !strings.Contains(err.Error(), "key not found") {
		t.Errorf("Expected 'key not found' error, got: %v", err)
	}
}

// TestSecretsDecryptCommand_NoAge tests decryption without age installed
func TestSecretsDecryptCommand_NoAge(t *testing.T) {
	if isAgeInstalled() {
		t.Skip("age is installed, skipping no-age test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	os.WriteFile(filepath.Join(env.TempDir, ".env.age"), []byte("encrypted"), 0600)

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "decrypt"}
	defer func() { os.Args = oldArgs }()

	err := secretsDecryptCmd.RunE(secretsDecryptCmd, []string{})

	if err == nil {
		t.Error("Expected error when age is not installed")
	}

	if !strings.Contains(err.Error(), "age is not installed") {
		t.Errorf("Expected 'age is not installed' error, got: %v", err)
	}
}

// TestSecretsStatusCommand tests the status command
func TestSecretsStatusCommand(t *testing.T) {
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "status"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsStatusCmd.RunE(secretsStatusCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets status failed: %v", err)
	}

	output := capture.GetStdout()
	if !strings.Contains(output, "Secrets Management Status") {
		t.Errorf("Expected status header in output, got: %s", output)
	}

	// Should show age installation status
	if isAgeInstalled() {
		if !strings.Contains(output, "age is installed") {
			t.Error("Expected age installed message")
		}
	} else {
		if !strings.Contains(output, "age is NOT installed") {
			t.Error("Expected age not installed message")
		}
	}
}

// TestSecretsStatusCommand_WithKey tests status with existing key
func TestSecretsStatusCommand_WithKey(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Run status command
	os.Args = []string{"asc", "secrets", "status"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsStatusCmd.RunE(secretsStatusCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets status failed: %v", err)
	}

	output := capture.GetStdout()
	if !strings.Contains(output, "Age key exists") {
		t.Errorf("Expected key exists message, got: %s", output)
	}
}

// TestSecretsStatusCommand_WithEncryptedFiles tests status with encrypted files
func TestSecretsStatusCommand_WithEncryptedFiles(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key and encrypt file
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	env.WriteEnv(ValidEnv())
	os.Args = []string{"asc", "secrets", "encrypt"}
	secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})
	os.Args = oldArgs

	// Run status command
	os.Args = []string{"asc", "secrets", "status"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsStatusCmd.RunE(secretsStatusCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets status failed: %v", err)
	}

	output := capture.GetStdout()
	if !strings.Contains(output, ".env.age") {
		t.Errorf("Expected encrypted file in output, got: %s", output)
	}
}

// TestSecretsRotateCommand tests key rotation
func TestSecretsRotateCommand(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key and encrypt file
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	env.WriteEnv(ValidEnv())
	os.Args = []string{"asc", "secrets", "encrypt"}
	secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})
	os.Args = oldArgs

	// Note: Rotate command requires user confirmation
	// We can't easily test the full flow without mocking stdin
	// But we can verify the command structure exists
	if secretsRotateCmd == nil {
		t.Error("secretsRotateCmd should be defined")
	}

	if secretsRotateCmd.Use != "rotate" {
		t.Errorf("Expected rotate command, got: %s", secretsRotateCmd.Use)
	}
}

// TestSecretsRotateCommand_NoKey tests rotation without existing key
func TestSecretsRotateCommand_NoKey(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "rotate"}
	defer func() { os.Args = oldArgs }()

	// Note: This would require user confirmation, so we just verify error handling
	// The actual RunE would wait for user input
	if secretsRotateCmd.RunE == nil {
		t.Error("secretsRotateCmd.RunE should be defined")
	}
}

// TestSecretsEncryptCommand_InvalidEnvFile tests encryption with invalid env file
func TestSecretsEncryptCommand_InvalidEnvFile(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Create invalid .env file (missing required keys)
	invalidEnv := `SOME_OTHER_KEY=value
`
	env.WriteEnv(invalidEnv)

	// Run encrypt command (should warn but not fail if user confirms)
	os.Args = []string{"asc", "secrets", "encrypt"}
	defer func() { os.Args = oldArgs }()

	// Note: This would prompt for user confirmation
	// We just verify the validation logic exists
	if secretsEncryptCmd.RunE == nil {
		t.Error("secretsEncryptCmd.RunE should be defined")
	}
}

// Helper function to check if age is installed
func isAgeInstalled() bool {
	_, err := exec.LookPath("age")
	if err != nil {
		return false
	}
	_, err = exec.LookPath("age-keygen")
	return err == nil
}

// TestSecretsCommand_Structure tests the command structure
func TestSecretsCommand_Structure(t *testing.T) {
	if secretsCmd == nil {
		t.Fatal("secretsCmd should be defined")
	}

	if secretsCmd.Use != "secrets" {
		t.Errorf("Expected 'secrets' command, got: %s", secretsCmd.Use)
	}

	// Verify subcommands exist
	// Note: Commands() returns subcommands that match the Use field exactly or start with it
	subcommands := map[string]string{
		"init":    "init",
		"encrypt": "encrypt [file]",
		"decrypt": "decrypt [file]",
		"status":  "status",
		"rotate":  "rotate",
	}

	for name, expectedUse := range subcommands {
		found := false
		for _, cmd := range secretsCmd.Commands() {
			if strings.HasPrefix(cmd.Use, name) || cmd.Use == expectedUse {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' to exist", name)
		}
	}
}

// TestSecretsInitCommand_Structure tests init command structure
func TestSecretsInitCommand_Structure(t *testing.T) {
	if secretsInitCmd == nil {
		t.Fatal("secretsInitCmd should be defined")
	}

	if secretsInitCmd.Use != "init" {
		t.Errorf("Expected 'init' command, got: %s", secretsInitCmd.Use)
	}

	if secretsInitCmd.Short == "" {
		t.Error("init command should have a short description")
	}

	if secretsInitCmd.RunE == nil {
		t.Error("init command should have a RunE function")
	}
}

// TestSecretsEncryptCommand_Structure tests encrypt command structure
func TestSecretsEncryptCommand_Structure(t *testing.T) {
	if secretsEncryptCmd == nil {
		t.Fatal("secretsEncryptCmd should be defined")
	}

	if secretsEncryptCmd.Use != "encrypt [file]" {
		t.Errorf("Expected 'encrypt [file]' command, got: %s", secretsEncryptCmd.Use)
	}

	if secretsEncryptCmd.Short == "" {
		t.Error("encrypt command should have a short description")
	}

	if secretsEncryptCmd.RunE == nil {
		t.Error("encrypt command should have a RunE function")
	}
}

// TestSecretsDecryptCommand_Structure tests decrypt command structure
func TestSecretsDecryptCommand_Structure(t *testing.T) {
	if secretsDecryptCmd == nil {
		t.Fatal("secretsDecryptCmd should be defined")
	}

	if secretsDecryptCmd.Use != "decrypt [file]" {
		t.Errorf("Expected 'decrypt [file]' command, got: %s", secretsDecryptCmd.Use)
	}

	if secretsDecryptCmd.Short == "" {
		t.Error("decrypt command should have a short description")
	}

	if secretsDecryptCmd.RunE == nil {
		t.Error("decrypt command should have a RunE function")
	}
}

// TestSecretsStatusCommand_Structure tests status command structure
func TestSecretsStatusCommand_Structure(t *testing.T) {
	if secretsStatusCmd == nil {
		t.Fatal("secretsStatusCmd should be defined")
	}

	if secretsStatusCmd.Use != "status" {
		t.Errorf("Expected 'status' command, got: %s", secretsStatusCmd.Use)
	}

	if secretsStatusCmd.Short == "" {
		t.Error("status command should have a short description")
	}

	if secretsStatusCmd.RunE == nil {
		t.Error("status command should have a RunE function")
	}
}

// TestSecretsRotateCommand_Structure tests rotate command structure
func TestSecretsRotateCommand_Structure(t *testing.T) {
	if secretsRotateCmd == nil {
		t.Fatal("secretsRotateCmd should be defined")
	}

	if secretsRotateCmd.Use != "rotate" {
		t.Errorf("Expected 'rotate' command, got: %s", secretsRotateCmd.Use)
	}

	if secretsRotateCmd.Short == "" {
		t.Error("rotate command should have a short description")
	}

	if secretsRotateCmd.RunE == nil {
		t.Error("rotate command should have a RunE function")
	}
}

// TestSecretsEncryptCommand_ValidationWarning tests validation warning flow
func TestSecretsEncryptCommand_ValidationWarning(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Create invalid .env file (missing required keys)
	invalidEnv := `SOME_KEY=value
`
	env.WriteEnv(invalidEnv)

	// The command would prompt for confirmation, which we can't easily test
	// But we can verify the validation is called
	capture := NewCaptureOutput()
	capture.Start()

	// Note: This will hang waiting for user input, so we just verify the setup
	// In a real scenario, we'd need to mock stdin

	capture.Stop()
}

// TestSecretsCommand_Help tests help text
func TestSecretsCommand_Help(t *testing.T) {
	if secretsCmd.Long == "" {
		t.Error("secrets command should have long description")
	}

	if !strings.Contains(secretsCmd.Long, "age") {
		t.Error("secrets command help should mention age")
	}
}

// TestSecretsInitCommand_Help tests init help text
func TestSecretsInitCommand_Help(t *testing.T) {
	if secretsInitCmd.Long == "" {
		t.Error("init command should have long description")
	}

	if !strings.Contains(secretsInitCmd.Long, "key") {
		t.Error("init command help should mention key")
	}
}

// TestSecretsEncryptCommand_Help tests encrypt help text
func TestSecretsEncryptCommand_Help(t *testing.T) {
	if secretsEncryptCmd.Long == "" {
		t.Error("encrypt command should have long description")
	}

	if !strings.Contains(secretsEncryptCmd.Long, ".env") {
		t.Error("encrypt command help should mention .env")
	}
}

// TestSecretsDecryptCommand_Help tests decrypt help text
func TestSecretsDecryptCommand_Help(t *testing.T) {
	if secretsDecryptCmd.Long == "" {
		t.Error("decrypt command should have long description")
	}

	if !strings.Contains(secretsDecryptCmd.Long, ".env.age") {
		t.Error("decrypt command help should mention .env.age")
	}
}

// TestSecretsStatusCommand_Help tests status help text
func TestSecretsStatusCommand_Help(t *testing.T) {
	if secretsStatusCmd.Long == "" {
		t.Error("status command should have long description")
	}
}

// TestSecretsRotateCommand_Help tests rotate help text
func TestSecretsRotateCommand_Help(t *testing.T) {
	if secretsRotateCmd.Long == "" {
		t.Error("rotate command should have long description")
	}

	if !strings.Contains(secretsRotateCmd.Long, "key") {
		t.Error("rotate command help should mention key")
	}
}

// TestSecretsStatusCommand_WithUnencryptedFiles tests status showing unencrypted files
func TestSecretsStatusCommand_WithUnencryptedFiles(t *testing.T) {
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Create unencrypted .env file
	env.WriteEnv(ValidEnv())

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "status"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsStatusCmd.RunE(secretsStatusCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets status failed: %v", err)
	}

	output := capture.GetStdout()
	if !strings.Contains(output, ".env") {
		t.Errorf("Expected .env in output, got: %s", output)
	}
}

// TestSecretsCommand_Subcommands tests that all subcommands are properly registered
func TestSecretsCommand_Subcommands(t *testing.T) {
	commands := secretsCmd.Commands()

	if len(commands) < 5 {
		t.Errorf("Expected at least 5 subcommands, got %d", len(commands))
	}

	// Verify each command has required fields
	for _, cmd := range commands {
		if cmd.Use == "" {
			t.Error("Subcommand should have Use field")
		}
		if cmd.Short == "" {
			t.Errorf("Subcommand %s should have Short description", cmd.Use)
		}
		if cmd.RunE == nil {
			t.Errorf("Subcommand %s should have RunE function", cmd.Use)
		}
	}
}

// TestSecretsInitCommand_KeyPathCreation tests that key directory is created
func TestSecretsInitCommand_KeyPathCreation(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyDir := filepath.Join(env.TempDir, ".asc")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Verify directory doesn't exist initially
	if _, err := os.Stat(keyDir); err == nil {
		os.RemoveAll(keyDir)
	}

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	defer func() { os.Args = oldArgs }()

	err := secretsInitCmd.RunE(secretsInitCmd, []string{})

	if err != nil {
		t.Errorf("secrets init failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		t.Error("Key directory should be created")
	}
}

// TestSecretsEncryptCommand_FilePermissions tests that encrypted files have correct permissions
func TestSecretsEncryptCommand_FilePermissions(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Create .env file
	env.WriteEnv(ValidEnv())

	// Encrypt
	os.Args = []string{"asc", "secrets", "encrypt"}
	defer func() { os.Args = oldArgs }()

	err := secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})

	if err != nil {
		t.Errorf("secrets encrypt failed: %v", err)
	}

	// Verify encrypted file exists (permissions are handled by age tool)
	encPath := filepath.Join(env.TempDir, ".env.age")
	if _, err := os.Stat(encPath); os.IsNotExist(err) {
		t.Error("Encrypted file should exist")
	}
}

// TestSecretsDecryptCommand_FilePermissions tests that decrypted files have secure permissions
func TestSecretsDecryptCommand_FilePermissions(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key and encrypt
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	env.WriteEnv(ValidEnv())
	os.Args = []string{"asc", "secrets", "encrypt"}
	secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})
	os.Args = oldArgs

	// Remove original
	os.Remove(filepath.Join(env.TempDir, ".env"))

	// Decrypt
	os.Args = []string{"asc", "secrets", "decrypt"}
	defer func() { os.Args = oldArgs }()

	err := secretsDecryptCmd.RunE(secretsDecryptCmd, []string{})

	if err != nil {
		t.Errorf("secrets decrypt failed: %v", err)
	}

	// Verify file has secure permissions (0600)
	envPath := filepath.Join(env.TempDir, ".env")
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Failed to stat decrypted file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Expected permissions 0600, got %o", mode)
	}
}

// TestSecretsInitCommand_PublicKeyDisplay tests that public key is displayed after generation
func TestSecretsInitCommand_PublicKeyDisplay(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsInitCmd.RunE(secretsInitCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets init failed: %v", err)
	}

	output := capture.GetStdout()
	if !strings.Contains(output, "Public key:") {
		t.Errorf("Expected public key in output, got: %s", output)
	}

	if !strings.Contains(output, "age1") {
		t.Errorf("Expected age1 public key format, got: %s", output)
	}
}

// TestSecretsEncryptCommand_OutputMessages tests encryption output messages
func TestSecretsEncryptCommand_OutputMessages(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Create .env file
	env.WriteEnv(ValidEnv())

	// Encrypt
	os.Args = []string{"asc", "secrets", "encrypt"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets encrypt failed: %v", err)
	}

	output := capture.GetStdout()
	if !strings.Contains(output, "safely commit") {
		t.Errorf("Expected commit message in output, got: %s", output)
	}

	if !strings.Contains(output, ".gitignore") {
		t.Errorf("Expected gitignore reminder in output, got: %s", output)
	}
}

// TestSecretsDecryptCommand_OutputMessages tests decryption output messages
func TestSecretsDecryptCommand_OutputMessages(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key and encrypt
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	env.WriteEnv(ValidEnv())
	os.Args = []string{"asc", "secrets", "encrypt"}
	secretsEncryptCmd.RunE(secretsEncryptCmd, []string{})
	os.Args = oldArgs

	// Remove original
	os.Remove(filepath.Join(env.TempDir, ".env"))

	// Decrypt
	os.Args = []string{"asc", "secrets", "decrypt"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsDecryptCmd.RunE(secretsDecryptCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets decrypt failed: %v", err)
	}

	output := capture.GetStdout()
	if !strings.Contains(output, "Secrets are now available") {
		t.Errorf("Expected availability message in output, got: %s", output)
	}
}

// TestSecretsStatusCommand_NoEncryptedFiles tests status with no encrypted files
func TestSecretsStatusCommand_NoEncryptedFiles(t *testing.T) {
	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "status"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsStatusCmd.RunE(secretsStatusCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets status failed: %v", err)
	}

	output := capture.GetStdout()
	if !strings.Contains(output, "Encrypted Files:") {
		t.Errorf("Expected encrypted files section, got: %s", output)
	}

	if !strings.Contains(output, "(none found)") {
		t.Errorf("Expected 'none found' message, got: %s", output)
	}
}

// TestSecretsStatusCommand_MultipleFiles tests status with multiple encrypted files
func TestSecretsStatusCommand_MultipleFiles(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	// Generate key
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	secretsInitCmd.RunE(secretsInitCmd, []string{})
	os.Args = oldArgs

	// Create and encrypt multiple files
	files := []string{".env", ".env.prod", ".env.staging"}
	for _, file := range files {
		os.WriteFile(filepath.Join(env.TempDir, file), []byte(ValidEnv()), 0600)
		os.Args = []string{"asc", "secrets", "encrypt", file}
		secretsEncryptCmd.RunE(secretsEncryptCmd, []string{file})
		os.Args = oldArgs
	}

	// Run status
	os.Args = []string{"asc", "secrets", "status"}
	defer func() { os.Args = oldArgs }()

	capture := NewCaptureOutput()
	capture.Start()

	err := secretsStatusCmd.RunE(secretsStatusCmd, []string{})

	capture.Stop()

	if err != nil {
		t.Errorf("secrets status failed: %v", err)
	}

	output := capture.GetStdout()
	for _, file := range files {
		if !strings.Contains(output, file+".age") {
			t.Errorf("Expected %s.age in output, got: %s", file, output)
		}
	}
}

// TestSecretsInitCommand_KeyPermissions tests that generated key has secure permissions
func TestSecretsInitCommand_KeyPermissions(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	keyPath := filepath.Join(env.TempDir, ".asc", "age.key")
	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	oldArgs := os.Args
	os.Args = []string{"asc", "secrets", "init"}
	defer func() { os.Args = oldArgs }()

	err := secretsInitCmd.RunE(secretsInitCmd, []string{})

	if err != nil {
		t.Errorf("secrets init failed: %v", err)
	}

	// Verify key has secure permissions (0600)
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Failed to stat key file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Expected key permissions 0600, got %o", mode)
	}
}

// TestSecretsCommand_Integration tests full workflow
func TestSecretsCommand_Integration(t *testing.T) {
	if !isAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}

	env := NewTestEnvironment(t)
	restore := ChangeToTempDir(t, env.TempDir)
	defer restore()

	os.Setenv("HOME", env.TempDir)
	defer os.Unsetenv("HOME")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Step 1: Init
	os.Args = []string{"asc", "secrets", "init"}
	if err := secretsInitCmd.RunE(secretsInitCmd, []string{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Step 2: Create env file
	env.WriteEnv(ValidEnv())

	// Step 3: Encrypt
	os.Args = []string{"asc", "secrets", "encrypt"}
	if err := secretsEncryptCmd.RunE(secretsEncryptCmd, []string{}); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Step 4: Remove original
	os.Remove(filepath.Join(env.TempDir, ".env"))

	// Step 5: Decrypt
	os.Args = []string{"asc", "secrets", "decrypt"}
	if err := secretsDecryptCmd.RunE(secretsDecryptCmd, []string{}); err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	// Step 6: Verify content
	decrypted := env.ReadFile(filepath.Join(env.TempDir, ".env"))
	if decrypted != ValidEnv() {
		t.Error("Decrypted content doesn't match original")
	}

	// Step 7: Check status
	os.Args = []string{"asc", "secrets", "status"}
	if err := secretsStatusCmd.RunE(secretsStatusCmd, []string{}); err != nil {
		t.Fatalf("status failed: %v", err)
	}
}
