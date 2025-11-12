// Package secrets provides secure secrets management using age encryption.
// It supports encrypting/decrypting .env files and managing age keys.
//
// Example usage:
//
//	manager := secrets.NewManager()
//	
//	// Generate a new age key
//	if err := manager.GenerateKey(); err != nil {
//	    log.Fatal(err)
//	}
//	
//	// Encrypt secrets
//	if err := manager.Encrypt(".env", ".env.age"); err != nil {
//	    log.Fatal(err)
//	}
//	
//	// Decrypt secrets
//	if err := manager.Decrypt(".env.age", ".env"); err != nil {
//	    log.Fatal(err)
//	}
package secrets

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Manager handles secrets encryption and decryption using age
type Manager struct {
	keyPath string // Path to age key file
}

// NewManager creates a new secrets manager with the default key path
func NewManager() *Manager {
	home, _ := os.UserHomeDir()
	return &Manager{
		keyPath: filepath.Join(home, ".asc", "age.key"),
	}
}

// NewManagerWithKeyPath creates a secrets manager with a custom key path
func NewManagerWithKeyPath(keyPath string) *Manager {
	return &Manager{
		keyPath: keyPath,
	}
}

// GenerateKey generates a new age key and saves it to the key file
func (m *Manager) GenerateKey() error {
	// Check if age is installed
	if !m.IsAgeInstalled() {
		return fmt.Errorf("age is not installed. Install with: brew install age (macOS) or see https://github.com/FiloSottile/age")
	}

	// Create directory if it doesn't exist
	keyDir := filepath.Dir(m.keyPath)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Generate key
	cmd := exec.Command("age-keygen", "-o", m.keyPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate age key: %w (output: %s)", err, output)
	}

	// Set restrictive permissions
	if err := os.Chmod(m.keyPath, 0600); err != nil {
		return fmt.Errorf("failed to set key permissions: %w", err)
	}

	return nil
}

// GetPublicKey extracts the public key from the age key file
func (m *Manager) GetPublicKey() (string, error) {
	if !m.KeyExists() {
		return "", fmt.Errorf("age key not found at %s", m.keyPath)
	}

	file, err := os.Open(m.keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to open key file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# public key: ") {
			return strings.TrimPrefix(line, "# public key: "), nil
		}
	}

	return "", fmt.Errorf("public key not found in key file")
}

// Encrypt encrypts a file using age encryption
func (m *Manager) Encrypt(inputPath, outputPath string) error {
	if !m.IsAgeInstalled() {
		return fmt.Errorf("age is not installed")
	}

	if !m.KeyExists() {
		return fmt.Errorf("age key not found. Run 'asc secrets init' first")
	}

	// Get public key
	pubKey, err := m.GetPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	// Encrypt file
	cmd := exec.Command("age", "-r", pubKey, "-o", outputPath, inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to encrypt file: %w (output: %s)", err, output)
	}

	return nil
}

// Decrypt decrypts a file using age decryption
func (m *Manager) Decrypt(inputPath, outputPath string) error {
	if !m.IsAgeInstalled() {
		return fmt.Errorf("age is not installed")
	}

	if !m.KeyExists() {
		return fmt.Errorf("age key not found at %s", m.keyPath)
	}

	// Decrypt file
	cmd := exec.Command("age", "-d", "-i", m.keyPath, "-o", outputPath, inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to decrypt file: %w (output: %s)", err, output)
	}

	// Set restrictive permissions on decrypted file
	if err := os.Chmod(outputPath, 0600); err != nil {
		return fmt.Errorf("failed to set permissions on decrypted file: %w", err)
	}

	return nil
}

// EncryptEnv encrypts the .env file to .env.age
func (m *Manager) EncryptEnv(envPath string) error {
	if envPath == "" {
		envPath = ".env"
	}

	outputPath := envPath + ".age"
	
	if err := m.Encrypt(envPath, outputPath); err != nil {
		return err
	}

	fmt.Printf("✓ Encrypted %s → %s\n", envPath, outputPath)
	fmt.Printf("✓ You can now safely commit %s to git\n", outputPath)
	fmt.Printf("⚠ Remember to add %s to .gitignore\n", envPath)
	
	return nil
}

// DecryptEnv decrypts the .env.age file to .env
func (m *Manager) DecryptEnv(envPath string) error {
	if envPath == "" {
		envPath = ".env"
	}

	inputPath := envPath + ".age"
	
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("encrypted file %s not found", inputPath)
	}

	if err := m.Decrypt(inputPath, envPath); err != nil {
		return err
	}

	fmt.Printf("✓ Decrypted %s → %s\n", inputPath, envPath)
	fmt.Printf("✓ Secrets are now available in %s\n", envPath)
	
	return nil
}

// KeyExists checks if the age key file exists
func (m *Manager) KeyExists() bool {
	_, err := os.Stat(m.keyPath)
	return err == nil
}

// IsAgeInstalled checks if age is installed and available in PATH
func (m *Manager) IsAgeInstalled() bool {
	_, err := exec.LookPath("age")
	if err != nil {
		return false
	}
	_, err = exec.LookPath("age-keygen")
	return err == nil
}

// GetKeyPath returns the path to the age key file
func (m *Manager) GetKeyPath() string {
	return m.keyPath
}

// ValidateEnvFile checks if an env file has the required structure
func (m *Manager) ValidateEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close()

	requiredKeys := []string{"CLAUDE_API_KEY", "OPENAI_API_KEY", "GOOGLE_API_KEY"}
	foundKeys := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			for _, reqKey := range requiredKeys {
				if key == reqKey {
					foundKeys[key] = true
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file: %w", err)
	}

	// Check for missing keys
	var missing []string
	for _, key := range requiredKeys {
		if !foundKeys[key] {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required keys: %v", missing)
	}

	return nil
}

// RotateKey generates a new age key and re-encrypts all encrypted files
func (m *Manager) RotateKey(encryptedFiles []string) error {
	// Backup old key
	oldKeyPath := m.keyPath + ".old"
	if m.KeyExists() {
		if err := copyFile(m.keyPath, oldKeyPath); err != nil {
			return fmt.Errorf("failed to backup old key: %w", err)
		}
		fmt.Printf("✓ Backed up old key to %s\n", oldKeyPath)
		
		// Remove the old key file so we can generate a new one
		if err := os.Remove(m.keyPath); err != nil {
			return fmt.Errorf("failed to remove old key: %w", err)
		}
	}

	// Generate new key
	if err := m.GenerateKey(); err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}
	fmt.Printf("✓ Generated new age key\n")

	// Re-encrypt all files
	for _, encFile := range encryptedFiles {
		// Decrypt with old key
		tempFile := encFile + ".temp"
		oldManager := NewManagerWithKeyPath(oldKeyPath)
		if err := oldManager.Decrypt(encFile, tempFile); err != nil {
			return fmt.Errorf("failed to decrypt %s with old key: %w", encFile, err)
		}

		// Encrypt with new key
		if err := m.Encrypt(tempFile, encFile); err != nil {
			return fmt.Errorf("failed to encrypt %s with new key: %w", encFile, err)
		}

		// Clean up temp file
		os.Remove(tempFile)
		fmt.Printf("✓ Re-encrypted %s\n", encFile)
	}

	fmt.Printf("✓ Key rotation complete\n")
	fmt.Printf("⚠ Keep %s in a safe place in case you need to recover old encrypted files\n", oldKeyPath)

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Copy permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}
