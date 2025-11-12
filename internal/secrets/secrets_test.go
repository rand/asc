package secrets

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.keyPath == "" {
		t.Error("keyPath should not be empty")
	}
}

func TestNewManagerWithKeyPath(t *testing.T) {
	customPath := "/custom/path/age.key"
	manager := NewManagerWithKeyPath(customPath)
	
	if manager.GetKeyPath() != customPath {
		t.Errorf("keyPath = %v, want %v", manager.GetKeyPath(), customPath)
	}
}

func TestKeyExists(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")
	
	manager := NewManagerWithKeyPath(keyPath)
	
	// Should not exist initially
	if manager.KeyExists() {
		t.Error("Key should not exist initially")
	}
	
	// Create a dummy key file
	err := os.WriteFile(keyPath, []byte("test key"), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key: %v", err)
	}
	
	// Should exist now
	if !manager.KeyExists() {
		t.Error("Key should exist after creation")
	}
}

func TestIsAgeInstalled(t *testing.T) {
	manager := NewManager()
	
	// This test will pass or fail depending on whether age is installed
	// We just verify the method doesn't panic
	installed := manager.IsAgeInstalled()
	t.Logf("age installed: %v", installed)
}

func TestValidateEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()
	
	tests := []struct {
		name      string
		content   string
		shouldErr bool
	}{
		{
			name: "valid env file",
			content: `CLAUDE_API_KEY=sk-ant-123
OPENAI_API_KEY=sk-456
GOOGLE_API_KEY=789
`,
			shouldErr: false,
		},
		{
			name: "missing keys",
			content: `CLAUDE_API_KEY=sk-ant-123
`,
			shouldErr: true,
		},
		{
			name: "with comments",
			content: `# API Keys
CLAUDE_API_KEY=sk-ant-123
OPENAI_API_KEY=sk-456
# Google key
GOOGLE_API_KEY=789
`,
			shouldErr: false,
		},
		{
			name: "empty file",
			content: ``,
			shouldErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envPath := filepath.Join(tmpDir, tt.name+".env")
			err := os.WriteFile(envPath, []byte(tt.content), 0600)
			if err != nil {
				t.Fatalf("Failed to write test env file: %v", err)
			}
			
			err = manager.ValidateEnvFile(envPath)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateEnvFileMissing(t *testing.T) {
	manager := NewManager()
	err := manager.ValidateEnvFile("/nonexistent/file.env")
	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestGetKeyPath(t *testing.T) {
	customPath := "/test/path/key"
	manager := NewManagerWithKeyPath(customPath)
	
	if manager.GetKeyPath() != customPath {
		t.Errorf("GetKeyPath() = %v, want %v", manager.GetKeyPath(), customPath)
	}
}

func TestEncryptDecryptFlow(t *testing.T) {
	// Skip if age is not installed
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping encryption test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Verify key exists
	if !manager.KeyExists() {
		t.Error("Key should exist after generation")
	}
	
	// Create test env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=test-key-123
OPENAI_API_KEY=test-key-456
GOOGLE_API_KEY=test-key-789
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test env: %v", err)
	}
	
	// Encrypt
	encPath := envPath + ".age"
	if err := manager.Encrypt(envPath, encPath); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}
	
	// Verify encrypted file exists
	if _, err := os.Stat(encPath); os.IsNotExist(err) {
		t.Error("Encrypted file should exist")
	}
	
	// Remove original
	os.Remove(envPath)
	
	// Decrypt
	if err := manager.Decrypt(encPath, envPath); err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	
	// Verify decrypted content
	decrypted, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}
	
	if string(decrypted) != envContent {
		t.Errorf("Decrypted content doesn't match original")
	}
}

func TestGetPublicKey(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping public key test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Should fail before key generation
	_, err := manager.GetPublicKey()
	if err == nil {
		t.Error("Expected error when key doesn't exist")
	}
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Should succeed after generation
	pubKey, err := manager.GetPublicKey()
	if err != nil {
		t.Errorf("Failed to get public key: %v", err)
	}
	
	if pubKey == "" {
		t.Error("Public key should not be empty")
	}
	
	// Public key should start with "age1"
	if len(pubKey) < 4 || pubKey[:4] != "age1" {
		t.Errorf("Public key should start with 'age1', got: %s", pubKey)
	}
}

func TestEncryptEnvHelperMethod(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Create test env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=test
OPENAI_API_KEY=test
GOOGLE_API_KEY=test
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test env: %v", err)
	}
	
	// Change to temp dir for relative paths
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)
	
	// Encrypt using helper method
	if err := manager.EncryptEnv(".env"); err != nil {
		t.Errorf("EncryptEnv failed: %v", err)
	}
	
	// Verify encrypted file exists
	if _, err := os.Stat(".env.age"); os.IsNotExist(err) {
		t.Error("Encrypted file should exist")
	}
}

func TestDecryptEnvHelperMethod(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Create and encrypt test env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=test
OPENAI_API_KEY=test
GOOGLE_API_KEY=test
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test env: %v", err)
	}
	
	encPath := envPath + ".age"
	if err := manager.Encrypt(envPath, encPath); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}
	
	// Remove original
	os.Remove(envPath)
	
	// Change to temp dir for relative paths
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)
	
	// Decrypt using helper method
	if err := manager.DecryptEnv(".env"); err != nil {
		t.Errorf("DecryptEnv failed: %v", err)
	}
	
	// Verify decrypted file exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		t.Error("Decrypted file should exist")
	}
}

func TestDecryptEnvMissingFile(t *testing.T) {
	manager := NewManager()
	tmpDir := t.TempDir()
	
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)
	
	err := manager.DecryptEnv(".env")
	if err == nil {
		t.Error("Expected error for missing encrypted file")
	}
}

// Additional comprehensive tests for better coverage

func TestGenerateKey_NoAge(t *testing.T) {
	// Test when age is not installed
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager := NewManagerWithKeyPath(keyPath)
	
	// If age is installed, skip this test
	if manager.IsAgeInstalled() {
		t.Skip("age is installed, skipping no-age test")
	}
	
	err := manager.GenerateKey()
	if err == nil {
		t.Error("Expected error when age is not installed")
	}
}

func TestEncrypt_NoKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "nonexistent.key")
	manager := NewManagerWithKeyPath(keyPath)
	
	inputPath := filepath.Join(tmpDir, "input.txt")
	outputPath := filepath.Join(tmpDir, "output.age")
	
	os.WriteFile(inputPath, []byte("test"), 0600)
	
	err := manager.Encrypt(inputPath, outputPath)
	if err == nil {
		t.Error("Expected error when key doesn't exist")
	}
}

func TestDecrypt_NoKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "nonexistent.key")
	manager := NewManagerWithKeyPath(keyPath)
	
	inputPath := filepath.Join(tmpDir, "input.age")
	outputPath := filepath.Join(tmpDir, "output.txt")
	
	err := manager.Decrypt(inputPath, outputPath)
	if err == nil {
		t.Error("Expected error when key doesn't exist")
	}
}

func TestValidateEnvFile_WithExtraKeys(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()
	
	content := `CLAUDE_API_KEY=sk-ant-123
OPENAI_API_KEY=sk-456
GOOGLE_API_KEY=789
EXTRA_KEY=value
ANOTHER_KEY=value2
`
	
	envPath := filepath.Join(tmpDir, "test.env")
	err := os.WriteFile(envPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write test env file: %v", err)
	}
	
	err = manager.ValidateEnvFile(envPath)
	if err != nil {
		t.Errorf("Should accept env file with extra keys: %v", err)
	}
}

func TestValidateEnvFile_WithWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()
	
	content := `
CLAUDE_API_KEY = sk-ant-123
  OPENAI_API_KEY=sk-456  
GOOGLE_API_KEY=789

`
	
	envPath := filepath.Join(tmpDir, "test.env")
	err := os.WriteFile(envPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write test env file: %v", err)
	}
	
	err = manager.ValidateEnvFile(envPath)
	if err != nil {
		t.Errorf("Should handle whitespace: %v", err)
	}
}

func TestValidateEnvFile_MissingOneKey(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()
	
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "missing CLAUDE_API_KEY",
			content: `OPENAI_API_KEY=sk-456
GOOGLE_API_KEY=789
`,
		},
		{
			name: "missing OPENAI_API_KEY",
			content: `CLAUDE_API_KEY=sk-ant-123
GOOGLE_API_KEY=789
`,
		},
		{
			name: "missing GOOGLE_API_KEY",
			content: `CLAUDE_API_KEY=sk-ant-123
OPENAI_API_KEY=sk-456
`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envPath := filepath.Join(tmpDir, tt.name+".env")
			err := os.WriteFile(envPath, []byte(tt.content), 0600)
			if err != nil {
				t.Fatalf("Failed to write test env file: %v", err)
			}
			
			err = manager.ValidateEnvFile(envPath)
			if err == nil {
				t.Error("Expected error for missing key")
			}
		})
	}
}

func TestValidateEnvFile_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()
	
	content := `CLAUDE_API_KEY=sk-ant-123
OPENAI_API_KEY=sk-456
GOOGLE_API_KEY=789
INVALID LINE WITHOUT EQUALS
`
	
	envPath := filepath.Join(tmpDir, "test.env")
	err := os.WriteFile(envPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write test env file: %v", err)
	}
	
	// Should still pass because required keys are present
	err = manager.ValidateEnvFile(envPath)
	if err != nil {
		t.Errorf("Should ignore invalid lines: %v", err)
	}
}

func TestGetPublicKey_InvalidKeyFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "invalid.key")
	manager := NewManagerWithKeyPath(keyPath)
	
	// Create a key file without public key comment
	err := os.WriteFile(keyPath, []byte("invalid key content"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test key: %v", err)
	}
	
	_, err = manager.GetPublicKey()
	if err == nil {
		t.Error("Expected error for invalid key file")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")
	
	content := []byte("test content")
	err := os.WriteFile(srcPath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}
	
	err = copyFile(srcPath, dstPath)
	if err != nil {
		t.Errorf("copyFile failed: %v", err)
	}
	
	// Verify content
	copied, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	
	if string(copied) != string(content) {
		t.Error("Copied content doesn't match original")
	}
	
	// Verify permissions
	srcInfo, _ := os.Stat(srcPath)
	dstInfo, _ := os.Stat(dstPath)
	
	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("Permissions don't match: src=%v, dst=%v", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyFile_NonexistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	
	srcPath := filepath.Join(tmpDir, "nonexistent.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")
	
	err := copyFile(srcPath, dstPath)
	if err == nil {
		t.Error("Expected error for nonexistent source file")
	}
}

func TestCopyFile_InvalidDestination(t *testing.T) {
	tmpDir := t.TempDir()
	
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := "/invalid/path/dest.txt"
	
	err := os.WriteFile(srcPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}
	
	err = copyFile(srcPath, dstPath)
	if err == nil {
		t.Error("Expected error for invalid destination path")
	}
}

func TestRotateKey(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping key rotation test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate initial key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate initial key: %v", err)
	}
	
	// Get initial public key
	initialPubKey, err := manager.GetPublicKey()
	if err != nil {
		t.Fatalf("Failed to get initial public key: %v", err)
	}
	
	// Create and encrypt a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	encFile := testFile + ".age"
	
	err = os.WriteFile(testFile, []byte("secret data"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	if err := manager.Encrypt(testFile, encFile); err != nil {
		t.Fatalf("Failed to encrypt test file: %v", err)
	}
	
	// Rotate key
	if err := manager.RotateKey([]string{encFile}); err != nil {
		t.Errorf("RotateKey failed: %v", err)
	}
	
	// Verify old key backup exists
	oldKeyPath := keyPath + ".old"
	if _, err := os.Stat(oldKeyPath); os.IsNotExist(err) {
		t.Error("Old key backup should exist")
	}
	
	// Verify new key is different
	newPubKey, err := manager.GetPublicKey()
	if err != nil {
		t.Fatalf("Failed to get new public key: %v", err)
	}
	
	if newPubKey == initialPubKey {
		t.Error("New public key should be different from initial key")
	}
	
	// Verify we can decrypt with new key
	decFile := filepath.Join(tmpDir, "decrypted.txt")
	if err := manager.Decrypt(encFile, decFile); err != nil {
		t.Errorf("Failed to decrypt with new key: %v", err)
	}
	
	// Verify content
	decrypted, err := os.ReadFile(decFile)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}
	
	if string(decrypted) != "secret data" {
		t.Error("Decrypted content doesn't match original")
	}
}

func TestRotateKey_NoExistingKey(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Rotate without existing key (should generate new one)
	if err := manager.RotateKey([]string{}); err != nil {
		t.Errorf("RotateKey should work without existing key: %v", err)
	}
	
	// Verify new key exists
	if !manager.KeyExists() {
		t.Error("New key should exist after rotation")
	}
}

func TestEncrypt_NoAge(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager := NewManagerWithKeyPath(keyPath)
	
	// If age is installed, skip this test
	if manager.IsAgeInstalled() {
		t.Skip("age is installed, skipping no-age test")
	}
	
	inputPath := filepath.Join(tmpDir, "input.txt")
	outputPath := filepath.Join(tmpDir, "output.age")
	
	os.WriteFile(inputPath, []byte("test"), 0600)
	
	err := manager.Encrypt(inputPath, outputPath)
	if err == nil {
		t.Error("Expected error when age is not installed")
	}
}

func TestDecrypt_NoAge(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager := NewManagerWithKeyPath(keyPath)
	
	// If age is installed, skip this test
	if manager.IsAgeInstalled() {
		t.Skip("age is installed, skipping no-age test")
	}
	
	inputPath := filepath.Join(tmpDir, "input.age")
	outputPath := filepath.Join(tmpDir, "output.txt")
	
	err := manager.Decrypt(inputPath, outputPath)
	if err == nil {
		t.Error("Expected error when age is not installed")
	}
}

func TestEncryptEnv_EmptyPath(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Create .env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=test
OPENAI_API_KEY=test
GOOGLE_API_KEY=test
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test env: %v", err)
	}
	
	// Change to temp dir
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)
	
	// Encrypt with empty path (should default to .env)
	if err := manager.EncryptEnv(""); err != nil {
		t.Errorf("EncryptEnv with empty path failed: %v", err)
	}
}

func TestDecryptEnv_EmptyPath(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Create and encrypt .env file
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `CLAUDE_API_KEY=test
OPENAI_API_KEY=test
GOOGLE_API_KEY=test
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test env: %v", err)
	}
	
	encPath := envPath + ".age"
	if err := manager.Encrypt(envPath, encPath); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}
	
	// Remove original
	os.Remove(envPath)
	
	// Change to temp dir
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)
	
	// Decrypt with empty path (should default to .env)
	if err := manager.DecryptEnv(""); err != nil {
		t.Errorf("DecryptEnv with empty path failed: %v", err)
	}
}

func TestManager_MultipleOperations(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Create multiple test files
	for i := 1; i <= 3; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.txt", i))
		encFile := testFile + ".age"
		
		content := fmt.Sprintf("secret data %d", i)
		err := os.WriteFile(testFile, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to write test file %d: %v", i, err)
		}
		
		// Encrypt
		if err := manager.Encrypt(testFile, encFile); err != nil {
			t.Errorf("Failed to encrypt file %d: %v", i, err)
		}
		
		// Verify encrypted file exists
		if _, err := os.Stat(encFile); os.IsNotExist(err) {
			t.Errorf("Encrypted file %d should exist", i)
		}
		
		// Remove original
		os.Remove(testFile)
		
		// Decrypt
		if err := manager.Decrypt(encFile, testFile); err != nil {
			t.Errorf("Failed to decrypt file %d: %v", i, err)
		}
		
		// Verify content
		decrypted, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read decrypted file %d: %v", i, err)
		}
		
		if string(decrypted) != content {
			t.Errorf("Decrypted content %d doesn't match original", i)
		}
	}
}

func TestValidateEnvFile_EmptyValues(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()
	
	content := `CLAUDE_API_KEY=
OPENAI_API_KEY=
GOOGLE_API_KEY=
`
	
	envPath := filepath.Join(tmpDir, "test.env")
	err := os.WriteFile(envPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write test env file: %v", err)
	}
	
	// Should pass because keys are present (even if values are empty)
	err = manager.ValidateEnvFile(envPath)
	if err != nil {
		t.Errorf("Should accept env file with empty values: %v", err)
	}
}

func TestValidateEnvFile_OnlyComments(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()
	
	content := `# This is a comment
# Another comment
# CLAUDE_API_KEY=commented-out
`
	
	envPath := filepath.Join(tmpDir, "test.env")
	err := os.WriteFile(envPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write test env file: %v", err)
	}
	
	err = manager.ValidateEnvFile(envPath)
	if err == nil {
		t.Error("Expected error for file with only comments")
	}
}

// Additional tests for key rotation functionality

func TestRotateKey_MultipleFiles(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate initial key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate initial key: %v", err)
	}
	
	// Create and encrypt multiple test files
	testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
	encFiles := make([]string, len(testFiles))
	
	for i, testFile := range testFiles {
		fullPath := filepath.Join(tmpDir, testFile)
		encPath := fullPath + ".age"
		encFiles[i] = encPath
		
		content := fmt.Sprintf("secret data %d", i+1)
		err := os.WriteFile(fullPath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to write test file %d: %v", i, err)
		}
		
		if err := manager.Encrypt(fullPath, encPath); err != nil {
			t.Fatalf("Failed to encrypt test file %d: %v", i, err)
		}
	}
	
	// Rotate key with multiple files
	if err := manager.RotateKey(encFiles); err != nil {
		t.Errorf("RotateKey with multiple files failed: %v", err)
	}
	
	// Verify all files can be decrypted with new key
	for i, encFile := range encFiles {
		decFile := filepath.Join(tmpDir, fmt.Sprintf("dec%d.txt", i))
		if err := manager.Decrypt(encFile, decFile); err != nil {
			t.Errorf("Failed to decrypt file %d with new key: %v", i, err)
		}
		
		// Verify content
		decrypted, err := os.ReadFile(decFile)
		if err != nil {
			t.Fatalf("Failed to read decrypted file %d: %v", i, err)
		}
		
		expected := fmt.Sprintf("secret data %d", i+1)
		if string(decrypted) != expected {
			t.Errorf("Decrypted content %d doesn't match: got %s, want %s", i, string(decrypted), expected)
		}
	}
}

func TestRotateKey_FailedDecryption(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate initial key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate initial key: %v", err)
	}
	
	// Create a fake encrypted file (not actually encrypted)
	fakeEncFile := filepath.Join(tmpDir, "fake.age")
	err := os.WriteFile(fakeEncFile, []byte("not encrypted"), 0600)
	if err != nil {
		t.Fatalf("Failed to write fake encrypted file: %v", err)
	}
	
	// Attempt to rotate key with invalid encrypted file
	err = manager.RotateKey([]string{fakeEncFile})
	if err == nil {
		t.Error("Expected error when rotating with invalid encrypted file")
	}
}

// Tests for public key extraction

func TestGetPublicKey_EmptyKeyFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "empty.key")
	manager := NewManagerWithKeyPath(keyPath)
	
	// Create empty key file
	err := os.WriteFile(keyPath, []byte(""), 0600)
	if err != nil {
		t.Fatalf("Failed to write empty key file: %v", err)
	}
	
	_, err = manager.GetPublicKey()
	if err == nil {
		t.Error("Expected error for empty key file")
	}
}

func TestGetPublicKey_MultilineKeyFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "multiline.key")
	manager := NewManagerWithKeyPath(keyPath)
	
	// Create key file with multiple lines but no public key
	content := `# This is a comment
AGE-SECRET-KEY-1234567890
# Another comment
`
	err := os.WriteFile(keyPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}
	
	_, err = manager.GetPublicKey()
	if err == nil {
		t.Error("Expected error for key file without public key")
	}
}

func TestGetPublicKey_ValidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "valid.key")
	manager := NewManagerWithKeyPath(keyPath)
	
	// Create key file with valid public key format
	content := `# created: 2024-01-01T00:00:00Z
# public key: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
AGE-SECRET-KEY-1234567890ABCDEF
`
	err := os.WriteFile(keyPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}
	
	pubKey, err := manager.GetPublicKey()
	if err != nil {
		t.Errorf("Failed to get public key: %v", err)
	}
	
	expected := "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p"
	if pubKey != expected {
		t.Errorf("Public key = %s, want %s", pubKey, expected)
	}
}

// Tests for error handling edge cases

func TestEncrypt_InvalidInputFile(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Try to encrypt non-existent file
	inputPath := filepath.Join(tmpDir, "nonexistent.txt")
	outputPath := filepath.Join(tmpDir, "output.age")
	
	err := manager.Encrypt(inputPath, outputPath)
	if err == nil {
		t.Error("Expected error when encrypting non-existent file")
	}
}

func TestDecrypt_InvalidInputFile(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Try to decrypt non-existent file
	inputPath := filepath.Join(tmpDir, "nonexistent.age")
	outputPath := filepath.Join(tmpDir, "output.txt")
	
	err := manager.Decrypt(inputPath, outputPath)
	if err == nil {
		t.Error("Expected error when decrypting non-existent file")
	}
}

func TestDecrypt_CorruptedFile(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Create a corrupted encrypted file
	inputPath := filepath.Join(tmpDir, "corrupted.age")
	err := os.WriteFile(inputPath, []byte("corrupted data"), 0600)
	if err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}
	
	outputPath := filepath.Join(tmpDir, "output.txt")
	
	err = manager.Decrypt(inputPath, outputPath)
	if err == nil {
		t.Error("Expected error when decrypting corrupted file")
	}
}

func TestEncrypt_WrongKeyFormat(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "wrong.key")
	manager := NewManagerWithKeyPath(keyPath)
	
	// Create key file with wrong format (no public key)
	err := os.WriteFile(keyPath, []byte("wrong key format"), 0600)
	if err != nil {
		t.Fatalf("Failed to write wrong key file: %v", err)
	}
	
	// Create test input file
	inputPath := filepath.Join(tmpDir, "input.txt")
	err = os.WriteFile(inputPath, []byte("test data"), 0600)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}
	
	outputPath := filepath.Join(tmpDir, "output.age")
	
	err = manager.Encrypt(inputPath, outputPath)
	if err == nil {
		t.Error("Expected error when encrypting with wrong key format")
	}
}

// Tests for key file management

func TestGenerateKey_DirectoryCreation(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "nested", "dir", "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key in nested directory
	if err := manager.GenerateKey(); err != nil {
		t.Errorf("Failed to generate key in nested directory: %v", err)
	}
	
	// Verify directory was created
	keyDir := filepath.Dir(keyPath)
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		t.Error("Key directory should have been created")
	}
	
	// Verify key file exists
	if !manager.KeyExists() {
		t.Error("Key file should exist after generation")
	}
}

func TestGenerateKey_Permissions(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Check file permissions
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Failed to stat key file: %v", err)
	}
	
	mode := info.Mode()
	if mode.Perm() != 0600 {
		t.Errorf("Key file permissions = %o, want 0600", mode.Perm())
	}
}

func TestDecrypt_OutputPermissions(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Create and encrypt test file
	inputPath := filepath.Join(tmpDir, "input.txt")
	encPath := filepath.Join(tmpDir, "input.age")
	
	err := os.WriteFile(inputPath, []byte("test data"), 0600)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}
	
	if err := manager.Encrypt(inputPath, encPath); err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}
	
	// Decrypt
	outputPath := filepath.Join(tmpDir, "output.txt")
	if err := manager.Decrypt(encPath, outputPath); err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	
	// Check output file permissions
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	
	mode := info.Mode()
	if mode.Perm() != 0600 {
		t.Errorf("Decrypted file permissions = %o, want 0600", mode.Perm())
	}
}

// Tests for concurrent encryption/decryption

func TestConcurrentEncryption(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Create multiple test files
	numFiles := 5
	testFiles := make([]string, numFiles)
	encFiles := make([]string, numFiles)
	
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.txt", i))
		encFile := testFile + ".age"
		testFiles[i] = testFile
		encFiles[i] = encFile
		
		content := fmt.Sprintf("test data %d", i)
		err := os.WriteFile(testFile, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to write test file %d: %v", i, err)
		}
	}
	
	// Encrypt files concurrently
	errChan := make(chan error, numFiles)
	for i := 0; i < numFiles; i++ {
		go func(idx int) {
			err := manager.Encrypt(testFiles[idx], encFiles[idx])
			errChan <- err
		}(i)
	}
	
	// Wait for all encryptions to complete
	for i := 0; i < numFiles; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent encryption %d failed: %v", i, err)
		}
	}
	
	// Verify all encrypted files exist
	for i, encFile := range encFiles {
		if _, err := os.Stat(encFile); os.IsNotExist(err) {
			t.Errorf("Encrypted file %d should exist", i)
		}
	}
}

func TestConcurrentDecryption(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Create and encrypt multiple test files
	numFiles := 5
	encFiles := make([]string, numFiles)
	decFiles := make([]string, numFiles)
	expectedContent := make([]string, numFiles)
	
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.txt", i))
		encFile := testFile + ".age"
		decFile := filepath.Join(tmpDir, fmt.Sprintf("dec%d.txt", i))
		encFiles[i] = encFile
		decFiles[i] = decFile
		
		content := fmt.Sprintf("test data %d", i)
		expectedContent[i] = content
		
		err := os.WriteFile(testFile, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to write test file %d: %v", i, err)
		}
		
		if err := manager.Encrypt(testFile, encFile); err != nil {
			t.Fatalf("Failed to encrypt test file %d: %v", i, err)
		}
	}
	
	// Decrypt files concurrently
	errChan := make(chan error, numFiles)
	for i := 0; i < numFiles; i++ {
		go func(idx int) {
			err := manager.Decrypt(encFiles[idx], decFiles[idx])
			errChan <- err
		}(i)
	}
	
	// Wait for all decryptions to complete
	for i := 0; i < numFiles; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent decryption %d failed: %v", i, err)
		}
	}
	
	// Verify all decrypted files have correct content
	for i, decFile := range decFiles {
		content, err := os.ReadFile(decFile)
		if err != nil {
			t.Errorf("Failed to read decrypted file %d: %v", i, err)
			continue
		}
		
		if string(content) != expectedContent[i] {
			t.Errorf("Decrypted content %d = %s, want %s", i, string(content), expectedContent[i])
		}
	}
}

func TestConcurrentMixedOperations(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Perform mixed operations concurrently
	numOps := 10
	errChan := make(chan error, numOps)
	
	for i := 0; i < numOps; i++ {
		go func(idx int) {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("mixed%d.txt", idx))
			encFile := testFile + ".age"
			decFile := filepath.Join(tmpDir, fmt.Sprintf("mixed_dec%d.txt", idx))
			
			content := fmt.Sprintf("mixed data %d", idx)
			
			// Write
			if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
				errChan <- fmt.Errorf("write failed: %w", err)
				return
			}
			
			// Encrypt
			if err := manager.Encrypt(testFile, encFile); err != nil {
				errChan <- fmt.Errorf("encrypt failed: %w", err)
				return
			}
			
			// Decrypt
			if err := manager.Decrypt(encFile, decFile); err != nil {
				errChan <- fmt.Errorf("decrypt failed: %w", err)
				return
			}
			
			// Verify
			decrypted, err := os.ReadFile(decFile)
			if err != nil {
				errChan <- fmt.Errorf("read failed: %w", err)
				return
			}
			
			if string(decrypted) != content {
				errChan <- fmt.Errorf("content mismatch: got %s, want %s", string(decrypted), content)
				return
			}
			
			errChan <- nil
		}(i)
	}
	
	// Wait for all operations to complete
	for i := 0; i < numOps; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent mixed operation %d failed: %v", i, err)
		}
	}
}

// Additional edge case tests

func TestEncryptEnv_NonexistentFile(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Change to temp dir
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)
	
	// Try to encrypt non-existent .env file
	err := manager.EncryptEnv("nonexistent.env")
	if err == nil {
		t.Error("Expected error when encrypting non-existent file")
	}
}

func TestValidateEnvFile_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()
	
	// Create a large env file with many keys
	var content string
	content += "CLAUDE_API_KEY=sk-ant-123\n"
	content += "OPENAI_API_KEY=sk-456\n"
	content += "GOOGLE_API_KEY=789\n"
	
	// Add many extra keys
	for i := 0; i < 1000; i++ {
		content += fmt.Sprintf("EXTRA_KEY_%d=value%d\n", i, i)
	}
	
	envPath := filepath.Join(tmpDir, "large.env")
	err := os.WriteFile(envPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write large env file: %v", err)
	}
	
	err = manager.ValidateEnvFile(envPath)
	if err != nil {
		t.Errorf("Should handle large env file: %v", err)
	}
}

func TestKeyExists_SymbolicLink(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create actual key file
	actualKeyPath := filepath.Join(tmpDir, "actual.key")
	err := os.WriteFile(actualKeyPath, []byte("test key"), 0600)
	if err != nil {
		t.Fatalf("Failed to write actual key: %v", err)
	}
	
	// Create symbolic link
	linkPath := filepath.Join(tmpDir, "link.key")
	err = os.Symlink(actualKeyPath, linkPath)
	if err != nil {
		t.Skip("Cannot create symbolic link, skipping test")
	}
	
	manager := NewManagerWithKeyPath(linkPath)
	
	if !manager.KeyExists() {
		t.Error("KeyExists should return true for symbolic link")
	}
}

func TestGenerateKey_ExistingFile(t *testing.T) {
	manager := NewManager()
	if !manager.IsAgeInstalled() {
		t.Skip("age not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "age.key")
	manager = NewManagerWithKeyPath(keyPath)
	
	// Generate key first time
	if err := manager.GenerateKey(); err != nil {
		t.Fatalf("Failed to generate key first time: %v", err)
	}
	
	// Get first public key
	firstPubKey, err := manager.GetPublicKey()
	if err != nil {
		t.Fatalf("Failed to get first public key: %v", err)
	}
	
	// Try to generate key again (should fail or overwrite)
	err = manager.GenerateKey()
	// age-keygen will fail if file exists, which is expected behavior
	if err == nil {
		// If it succeeded, verify it's a different key
		secondPubKey, err := manager.GetPublicKey()
		if err != nil {
			t.Fatalf("Failed to get second public key: %v", err)
		}
		
		if firstPubKey == secondPubKey {
			t.Error("Second key generation should create a different key")
		}
	}
	// If it failed, that's also acceptable behavior
}
