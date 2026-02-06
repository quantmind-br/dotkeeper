package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/diogo/dotkeeper/internal/config"
)

func DeleteCommand(args []string) int {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	force := fs.Bool("force", false, "Skip confirmation prompt")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dotkeeper delete <backup-name> [--force]\n\n")
		fmt.Fprintf(os.Stderr, "Delete a backup and its metadata.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 1
	}

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Error: backup name required\n")
		fmt.Fprintf(os.Stderr, "Usage: dotkeeper delete <backup-name>\n")
		return 1
	}

	name := fs.Arg(0)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	if !strings.HasSuffix(name, ".tar.gz.enc") {
		name = name + ".tar.gz.enc"
	}

	encPath := filepath.Join(cfg.BackupDir, name)
	if _, err := os.Stat(encPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: backup not found: %s\n", name)
		return 1
	}

	if !*force {
		fmt.Printf("Delete %s? [y/N] ", name)
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Cancelled")
			return 0
		}
	}

	if err := os.Remove(encPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting backup: %v\n", err)
		return 1
	}

	metaPath := encPath + ".meta.json"
	os.Remove(metaPath)

	fmt.Printf("Deleted %s\n", name)
	return 0
}
