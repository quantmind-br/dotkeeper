package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/restore"
)

// RestoreCommand handles the restore subcommand
func RestoreCommand(args []string) int {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	force := fs.Bool("force", false, "Overwrite existing files without prompting")
	passwordFile := fs.String("password-file", "", "Path to file containing password")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dotkeeper restore [backup-name] [--force] [--password-file PATH]\n\n")
		fmt.Fprintf(os.Stderr, "Restore dotfiles from a backup.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  DOTKEEPER_PASSWORD    Password for decryption (non-interactive mode)\n")
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 1
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	// Get backup name
	backupName := ""
	if fs.NArg() > 0 {
		backupName = fs.Arg(0)
	} else {
		fmt.Fprintf(os.Stderr, "Error: backup name required\n")
		fmt.Fprintf(os.Stderr, "Use 'dotkeeper list' to see available backups\n")
		return 1
	}

	// Construct backup path
	backupPath := filepath.Join(cfg.BackupDir, backupName)
	if !strings.HasSuffix(backupPath, ".tar.gz.enc") {
		backupPath += ".tar.gz.enc"
	}

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: backup not found: %s\n", backupPath)
		return 1
	}

	// Get password
	password, err := getPassword(*passwordFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting password: %v\n", err)
		return 1
	}

	// Perform restore
	fmt.Printf("Restoring from %s...\n", backupName)

	opts := restore.RestoreOptions{
		Force: *force,
	}

	result, err := restore.Restore(backupPath, password, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Restore failed: %v\n", err)
		return 1
	}

	// Print results
	if result.FilesSkipped > 0 {
		fmt.Printf("⚠ Restore completed with warnings\n")
		fmt.Printf("  Files restored: %d\n", result.FilesRestored)
		fmt.Printf("  Files skipped: %d\n", result.FilesSkipped)
		fmt.Printf("  Backup files created: %d\n", len(result.BackupFiles))
		return 2 // Partial success
	}

	fmt.Printf("✓ Restore completed successfully\n")
	fmt.Printf("  Files restored: %d\n", result.FilesRestored)
	fmt.Printf("  Backup files created: %d\n", len(result.BackupFiles))
	return 0
}
