package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/diogo/dotkeeper/internal/backup"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/keyring"
	"github.com/diogo/dotkeeper/internal/notify"
)

// BackupCommand handles the backup subcommand
func BackupCommand(args []string) int {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	passwordFile := fs.String("password-file", "", "Path to file containing password")
	notifyPtr := fs.Bool("notify", true, "Send desktop notifications on completion")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dotkeeper backup [--password-file PATH] [--notify]\n\n")
		fmt.Fprintf(os.Stderr, "Create a backup of dotfiles.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  DOTKEEPER_PASSWORD    Password for encryption (non-interactive mode)\n")
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 1
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'dotkeeper config' to set up configuration.\n")
		return 1
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		return 1
	}

	notifyFlag := *notifyPtr || cfg.Notifications

	// Get password from various sources
	password, err := getPassword(*passwordFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting password: %v\n", err)
		return 1
	}

	// Perform backup
	fmt.Println("Starting backup...")
	result, err := backup.Backup(cfg, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
		if notifyFlag {
			notify.SendError(err)
		}
		return 1
	}

	// Print results
	fmt.Printf("✓ Backup completed successfully\n")
	fmt.Printf("  Files backed up: %d\n", result.FileCount)
	fmt.Printf("  Total size: %d bytes\n", result.TotalSize)
	fmt.Printf("  Duration: %v\n", result.Duration)
	fmt.Printf("  Backup file: %s\n", result.BackupPath)
	fmt.Printf("  Checksum: %s\n", result.Checksum)

	if notifyFlag {
		notify.SendSuccess(result.BackupName, result.Duration)
	}

	return 0
}

// getPassword retrieves password from file, env var, or keyring
func getPassword(passwordFile string) (string, error) {
	// Priority 1: Password file
	if passwordFile != "" {
		data, err := os.ReadFile(passwordFile)
		if err != nil {
			return "", fmt.Errorf("failed to read password file: %w", err)
		}
		// Trim trailing newline if present
		password := string(data)
		if len(password) > 0 && password[len(password)-1] == '\n' {
			password = password[:len(password)-1]
		}
		return password, nil
	}

	// Priority 2: Environment variable
	if password := os.Getenv("DOTKEEPER_PASSWORD"); password != "" {
		return password, nil
	}

	// Priority 3: System keyring
	password, err := keyring.Retrieve()
	if err != nil {
		return "", fmt.Errorf("no password provided and keyring access failed: %w", err)
	}

	return password, nil
}
