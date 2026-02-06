package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/crypto"
	"github.com/diogo/dotkeeper/internal/pathutil"
)

// BackupInfo contains information about a backup
type BackupInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	Created      time.Time `json:"created"`
	OriginalSize int64     `json:"original_size,omitempty"`
}

// ListCommand handles the list subcommand
func ListCommand(args []string) int {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dotkeeper list [--json]\n\n")
		fmt.Fprintf(os.Stderr, "List available backups.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
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

	// Check if backup directory exists
	if _, err := os.Stat(cfg.BackupDir); os.IsNotExist(err) {
		if *jsonOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("No backups found (backup directory does not exist)")
		}
		return 0
	}

	// Find all backup files
	backups, err := findBackups(cfg.BackupDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding backups: %v\n", err)
		return 1
	}

	if len(backups) == 0 {
		if *jsonOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("No backups found")
		}
		return 0
	}

	// Sort by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Created.After(backups[j].Created)
	})

	// Output
	if *jsonOutput {
		data, err := json.MarshalIndent(backups, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			return 1
		}
		fmt.Println(string(data))
	} else {
		printBackupTable(backups)
	}

	return 0
}

// findBackups finds all backup files in the backup directory
func findBackups(backupDir string) ([]BackupInfo, error) {
	var backups []BackupInfo

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".tar.gz.enc") {
			continue
		}

		path := filepath.Join(backupDir, name)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		// Try to read metadata for more info
		metadataPath := path + ".meta.json"
		var originalSize int64
		var created time.Time = info.ModTime()

		if metadataData, err := os.ReadFile(metadataPath); err == nil {
			var metadata crypto.EncryptionMetadata
			if err := json.Unmarshal(metadataData, &metadata); err == nil {
				originalSize = metadata.OriginalSize
				created = metadata.Timestamp
			}
		}

		backups = append(backups, BackupInfo{
			Name:         name,
			Path:         path,
			Size:         info.Size(),
			Created:      created,
			OriginalSize: originalSize,
		})
	}

	return backups, nil
}

// printBackupTable prints backups in a formatted table
func printBackupTable(backups []BackupInfo) {
	fmt.Printf("%-40s %-20s %-12s %-12s\n", "NAME", "CREATED", "SIZE", "ORIGINAL")
	fmt.Println(strings.Repeat("-", 88))

	for _, backup := range backups {
		created := backup.Created.Format("2006-01-02 15:04:05")
		size := pathutil.FormatSize(backup.Size)
		originalSize := "-"
		if backup.OriginalSize > 0 {
			originalSize = pathutil.FormatSize(backup.OriginalSize)
		}

		fmt.Printf("%-40s %-20s %-12s %-12s\n",
			backup.Name,
			created,
			size,
			originalSize,
		)
	}

	fmt.Printf("\nTotal: %d backup(s)\n", len(backups))
}

