package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/rand/asc/internal/secrets"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage encrypted secrets using age",
	Long: `Securely manage API keys and secrets using age encryption.

Age (https://github.com/FiloSottile/age) provides simple, secure file encryption.
This command helps you encrypt your .env file so you can safely commit it to git.

Workflow:
  1. asc secrets init          # Generate age key
  2. Create .env with your API keys
  3. asc secrets encrypt        # Encrypt .env → .env.age
  4. git add .env.age           # Commit encrypted file
  5. asc secrets decrypt        # Decrypt when needed

The age key is stored in ~/.asc/age.key and should NEVER be committed to git.`,
}

var secretsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize age encryption (generate key)",
	Long: `Generate a new age encryption key for secrets management.

The key will be stored in ~/.asc/age.key with restrictive permissions (0600).
This key is used to encrypt and decrypt your .env files.

IMPORTANT: Keep this key safe and NEVER commit it to git!`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := secrets.NewManager()

		if manager.KeyExists() {
			fmt.Println("⚠ Age key already exists at", manager.GetKeyPath())
			fmt.Print("Do you want to overwrite it? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		if !manager.IsAgeInstalled() {
			fmt.Println("✗ age is not installed")
			fmt.Println("\nInstall age:")
			fmt.Println("  macOS:   brew install age")
			fmt.Println("  Linux:   apt install age  (or download from https://github.com/FiloSottile/age)")
			fmt.Println("  Windows: scoop install age")
			return fmt.Errorf("age not installed")
		}

		fmt.Println("Generating age key...")
		if err := manager.GenerateKey(); err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}

		pubKey, err := manager.GetPublicKey()
		if err != nil {
			return fmt.Errorf("failed to get public key: %w", err)
		}

		fmt.Println("✓ Age key generated successfully")
		fmt.Println("\nKey location:", manager.GetKeyPath())
		fmt.Println("Public key:", pubKey)
		fmt.Println("\n⚠ IMPORTANT: Keep your key safe and NEVER commit it to git!")
		fmt.Println("✓ The key file has been set to permissions 0600")

		return nil
	},
}

var secretsEncryptCmd = &cobra.Command{
	Use:   "encrypt [file]",
	Short: "Encrypt secrets file (default: .env)",
	Long: `Encrypt your secrets file using age encryption.

By default, encrypts .env to .env.age. You can specify a different file.

The encrypted .env.age file is safe to commit to git, while .env should
be added to .gitignore.

Example:
  asc secrets encrypt           # Encrypts .env → .env.age
  asc secrets encrypt .env.prod # Encrypts .env.prod → .env.prod.age`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := secrets.NewManager()

		if !manager.IsAgeInstalled() {
			return fmt.Errorf("age is not installed. Run 'asc secrets init' for installation instructions")
		}

		if !manager.KeyExists() {
			return fmt.Errorf("age key not found. Run 'asc secrets init' first")
		}

		envPath := ".env"
		if len(args) > 0 {
			envPath = args[0]
		}

		// Check if file exists
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			return fmt.Errorf("file %s not found", envPath)
		}

		// Validate env file structure
		if err := manager.ValidateEnvFile(envPath); err != nil {
			fmt.Printf("⚠ Warning: %v\n", err)
			fmt.Print("Continue anyway? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		fmt.Printf("Encrypting %s...\n", envPath)
		if err := manager.EncryptEnv(envPath); err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}

		return nil
	},
}

var secretsDecryptCmd = &cobra.Command{
	Use:   "decrypt [file]",
	Short: "Decrypt secrets file (default: .env.age)",
	Long: `Decrypt your encrypted secrets file using age decryption.

By default, decrypts .env.age to .env. You can specify a different file.

The decrypted file will have restrictive permissions (0600) set automatically.

Example:
  asc secrets decrypt           # Decrypts .env.age → .env
  asc secrets decrypt .env.prod # Decrypts .env.prod.age → .env.prod`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := secrets.NewManager()

		if !manager.IsAgeInstalled() {
			return fmt.Errorf("age is not installed. Run 'asc secrets init' for installation instructions")
		}

		if !manager.KeyExists() {
			return fmt.Errorf("age key not found at %s", manager.GetKeyPath())
		}

		envPath := ".env"
		if len(args) > 0 {
			envPath = args[0]
		}

		fmt.Printf("Decrypting %s.age...\n", envPath)
		if err := manager.DecryptEnv(envPath); err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}

		return nil
	},
}

var secretsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show secrets management status",
	Long:  `Display the current status of secrets management including key location and encrypted files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := secrets.NewManager()

		fmt.Println("Secrets Management Status")
		fmt.Println("========================")
		fmt.Println()

		// Check age installation
		if manager.IsAgeInstalled() {
			fmt.Println("✓ age is installed")
		} else {
			fmt.Println("✗ age is NOT installed")
			fmt.Println("  Install: brew install age (macOS) or see https://github.com/FiloSottile/age")
		}

		// Check key
		if manager.KeyExists() {
			fmt.Println("✓ Age key exists at", manager.GetKeyPath())
			if pubKey, err := manager.GetPublicKey(); err == nil {
				fmt.Println("  Public key:", pubKey)
			}
		} else {
			fmt.Println("✗ Age key NOT found")
			fmt.Println("  Run: asc secrets init")
		}

		fmt.Println()

		// Check for encrypted files
		encryptedFiles := []string{".env.age", ".env.prod.age", ".env.staging.age"}
		fmt.Println("Encrypted Files:")
		foundAny := false
		for _, file := range encryptedFiles {
			if _, err := os.Stat(file); err == nil {
				fmt.Println("  ✓", file)
				foundAny = true
			}
		}
		if !foundAny {
			fmt.Println("  (none found)")
		}

		fmt.Println()

		// Check for unencrypted files
		unencryptedFiles := []string{".env", ".env.prod", ".env.staging"}
		fmt.Println("Unencrypted Files:")
		foundAny = false
		for _, file := range unencryptedFiles {
			if _, err := os.Stat(file); err == nil {
				fmt.Println("  ⚠", file, "(should be encrypted and gitignored)")
				foundAny = true
			}
		}
		if !foundAny {
			fmt.Println("  (none found)")
		}

		return nil
	},
}

var secretsRotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate age key and re-encrypt all files",
	Long: `Generate a new age key and re-encrypt all encrypted files.

This is useful if you suspect your key has been compromised or as part
of regular security maintenance.

The old key will be backed up to ~/.asc/age.key.old`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := secrets.NewManager()

		if !manager.IsAgeInstalled() {
			return fmt.Errorf("age is not installed")
		}

		if !manager.KeyExists() {
			return fmt.Errorf("no existing key to rotate")
		}

		fmt.Println("⚠ This will generate a new key and re-encrypt all files")
		fmt.Print("Continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted.")
			return nil
		}

		// Find encrypted files
		encryptedFiles := []string{}
		candidates := []string{".env.age", ".env.prod.age", ".env.staging.age"}
		for _, file := range candidates {
			if _, err := os.Stat(file); err == nil {
				encryptedFiles = append(encryptedFiles, file)
			}
		}

		if len(encryptedFiles) == 0 {
			fmt.Println("No encrypted files found to re-encrypt")
		}

		if err := manager.RotateKey(encryptedFiles); err != nil {
			return fmt.Errorf("key rotation failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(secretsCmd)
	secretsCmd.AddCommand(secretsInitCmd)
	secretsCmd.AddCommand(secretsEncryptCmd)
	secretsCmd.AddCommand(secretsDecryptCmd)
	secretsCmd.AddCommand(secretsStatusCmd)
	secretsCmd.AddCommand(secretsRotateCmd)
}
