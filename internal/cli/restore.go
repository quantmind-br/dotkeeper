package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/crypto"
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
	filesRestored, partialFailures, err := restoreBackup(backupPath, password, *force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Restore failed: %v\n", err)
		return 1
	}

	// Print results
	if partialFailures > 0 {
		fmt.Printf("⚠ Restore completed with warnings\n")
		fmt.Printf("  Files restored: %d\n", filesRestored)
		fmt.Printf("  Partial failures: %d\n", partialFailures)
		return 2 // Partial success
	}

	fmt.Printf("✓ Restore completed successfully\n")
	fmt.Printf("  Files restored: %d\n", filesRestored)
	return 0
}

// restoreBackup decrypts and extracts a backup
func restoreBackup(backupPath, password string, force bool) (int, int, error) {
	// Read metadata
	metadataPath := backupPath + ".meta.json"
	metadataData, err := os.ReadFile(metadataPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata crypto.EncryptionMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return 0, 0, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Read encrypted backup
	encryptedData, err := os.ReadFile(backupPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read backup: %w", err)
	}

	// Derive key
	key := crypto.DeriveKey(password, metadata.Salt)

	// Decrypt
	decryptedData, err := crypto.Decrypt(encryptedData, key)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decrypt (wrong password?): %w", err)
	}

	// Extract archive
	filesRestored, partialFailures, err := extractArchive(decryptedData, force)
	if err != nil {
		return filesRestored, partialFailures, fmt.Errorf("failed to extract archive: %w", err)
	}

	return filesRestored, partialFailures, nil
}

// extractArchive extracts files from a tar.gz archive
func extractArchive(data []byte, force bool) (int, int, error) {
	// Create gzip reader
	gzr, err := gzip.NewReader(strings.NewReader(string(data)))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	filesRestored := 0
	partialFailures := 0

	// Extract each file
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return filesRestored, partialFailures, fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Check if file exists
		if !force {
			if _, err := os.Stat(header.Name); err == nil {
				fmt.Printf("Skipping existing file: %s (use --force to overwrite)\n", header.Name)
				partialFailures++
				continue
			}
		}

		// Create parent directories
		dir := filepath.Dir(header.Name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create directory %s: %v\n", dir, err)
			partialFailures++
			continue
		}

		// Create file
		outFile, err := os.OpenFile(header.Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create file %s: %v\n", header.Name, err)
			partialFailures++
			continue
		}

		// Copy content
		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			fmt.Fprintf(os.Stderr, "Warning: failed to write file %s: %v\n", header.Name, err)
			partialFailures++
			continue
		}
		outFile.Close()

		filesRestored++
	}

	return filesRestored, partialFailures, nil
}
