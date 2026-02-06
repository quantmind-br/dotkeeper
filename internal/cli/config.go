package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/diogo/dotkeeper/internal/config"
)

// ConfigCommand handles the config subcommand
func ConfigCommand(args []string) int {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dotkeeper config [get|set] KEY [VALUE]\n\n")
		fmt.Fprintf(os.Stderr, "Manage configuration.\n\n")
		fmt.Fprintf(os.Stderr, "Subcommands:\n")
		fmt.Fprintf(os.Stderr, "  get KEY        Get a configuration value\n")
		fmt.Fprintf(os.Stderr, "  set KEY VALUE  Set a configuration value\n")
		fmt.Fprintf(os.Stderr, "  list           List all configuration values\n")
		fmt.Fprintf(os.Stderr, "\nAvailable keys:\n")
		fmt.Fprintf(os.Stderr, "  backup_dir     Directory for storing backups\n")
		fmt.Fprintf(os.Stderr, "  git_remote     Git remote URL\n")
		fmt.Fprintf(os.Stderr, "  schedule       Backup schedule (cron format)\n")
		fmt.Fprintf(os.Stderr, "  notifications  Enable/disable notifications (true/false)\n")
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 1
	}

	if fs.NArg() < 1 {
		fs.Usage()
		return 1
	}

	subcommand := fs.Arg(0)

	switch subcommand {
	case "get":
		return configGet(fs.Args()[1:])
	case "set":
		return configSet(fs.Args()[1:])
	case "list":
		return configList()
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		fs.Usage()
		return 1
	}
}

// configGet gets a configuration value
func configGet(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: KEY required\n")
		fmt.Fprintf(os.Stderr, "Usage: dotkeeper config get KEY\n")
		return 1
	}

	key := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	// Get value
	value, err := getConfigValue(cfg, key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println(value)
	return 0
}

// configSet sets a configuration value
func configSet(args []string) int {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: KEY and VALUE required\n")
		fmt.Fprintf(os.Stderr, "Usage: dotkeeper config set KEY VALUE\n")
		return 1
	}

	key := args[0]
	value := args[1]

	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		// Create default config if it doesn't exist
		cfg = &config.Config{}
	}

	// Set value
	if err := setConfigValue(cfg, key, value); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Save config
	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Set %s = %s\n", key, value)
	return 0
}

// configList lists all configuration values
func configList() int {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	// Print all values
	fmt.Println("Configuration:")
	fmt.Printf("  backup_dir:     %s\n", cfg.BackupDir)
	fmt.Printf("  git_remote:     %s\n", cfg.GitRemote)
	fmt.Printf("  schedule:       %s\n", cfg.Schedule)
	fmt.Printf("  notifications:  %t\n", cfg.Notifications)
	fmt.Printf("  files:          %v\n", cfg.Files)
	fmt.Printf("  folders:        %v\n", cfg.Folders)

	return 0
}

// getConfigValue gets a value from the config by key
func getConfigValue(cfg *config.Config, key string) (string, error) {
	key = normalizeKey(key)

	switch key {
	case "backup_dir":
		return cfg.BackupDir, nil
	case "git_remote":
		return cfg.GitRemote, nil
	case "schedule":
		return cfg.Schedule, nil
	case "notifications":
		return fmt.Sprintf("%t", cfg.Notifications), nil
	case "files":
		return strings.Join(cfg.Files, ","), nil
	case "folders":
		return strings.Join(cfg.Folders, ","), nil
	default:
		return "", fmt.Errorf("unknown key: %s", key)
	}
}

// setConfigValue sets a value in the config by key
func setConfigValue(cfg *config.Config, key, value string) error {
	key = normalizeKey(key)

	switch key {
	case "backup_dir":
		cfg.BackupDir = value
	case "git_remote":
		cfg.GitRemote = value
	case "schedule":
		cfg.Schedule = value
	case "notifications":
		switch strings.ToLower(value) {
		case "true", "yes", "1", "on":
			cfg.Notifications = true
		case "false", "no", "0", "off":
			cfg.Notifications = false
		default:
			return fmt.Errorf("invalid boolean value: %s (use true/false)", value)
		}
	case "files":
		if value == "" {
			cfg.Files = []string{}
		} else {
			cfg.Files = strings.Split(value, ",")
		}
	case "folders":
		if value == "" {
			cfg.Folders = []string{}
		} else {
			cfg.Folders = strings.Split(value, ",")
		}
	default:
		return fmt.Errorf("unknown key: %s", key)
	}

	return nil
}

// normalizeKey normalizes a config key (converts to lowercase, replaces - with _)
func normalizeKey(key string) string {
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "-", "_")
	return key
}
