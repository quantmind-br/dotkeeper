package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the dotkeeper configuration
type Config struct {
	BackupDir       string   `yaml:"backup_dir"`
	GitRemote       string   `yaml:"git_remote"`
	Files           []string `yaml:"files"`
	Folders         []string `yaml:"folders"`
	Schedule        string   `yaml:"schedule"` // cron format
	Notifications   bool     `yaml:"notifications"`
	Exclude         []string `yaml:"exclude,omitempty"`
	DisabledFiles   []string `yaml:"disabled_files,omitempty"`
	DisabledFolders []string `yaml:"disabled_folders,omitempty"`
}

// GetConfigDir returns the XDG config directory for dotkeeper
func GetConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
	}

	dotKeeperDir := filepath.Join(configHome, "dotkeeper")
	return dotKeeperDir, nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

// LoadFromPath loads config from a specific path
func LoadFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Load loads config from the default XDG location
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadFromPath(configPath)
}

// SaveToPath saves config to a specific path
func (c *Config) SaveToPath(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Save saves config to the default XDG location
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	return c.SaveToPath(configPath)
}

// Validate validates the config
func (c *Config) Validate() error {
	if c.BackupDir == "" {
		return fmt.Errorf("backup_dir is required")
	}

	if len(c.Files) == 0 && len(c.Folders) == 0 {
		return fmt.Errorf("at least one file or folder must be specified")
	}

	return nil
}

// ActiveFiles returns Files minus any entries in DisabledFiles.
func (c *Config) ActiveFiles() []string {
	if len(c.DisabledFiles) == 0 {
		return c.Files
	}
	disabled := make(map[string]bool, len(c.DisabledFiles))
	for _, f := range c.DisabledFiles {
		disabled[f] = true
	}
	active := make([]string, 0, len(c.Files))
	for _, f := range c.Files {
		if !disabled[f] {
			active = append(active, f)
		}
	}
	return active
}

// ActiveFolders returns Folders minus any entries in DisabledFolders.
func (c *Config) ActiveFolders() []string {
	if len(c.DisabledFolders) == 0 {
		return c.Folders
	}
	disabled := make(map[string]bool, len(c.DisabledFolders))
	for _, f := range c.DisabledFolders {
		disabled[f] = true
	}
	active := make([]string, 0, len(c.Folders))
	for _, f := range c.Folders {
		if !disabled[f] {
			active = append(active, f)
		}
	}
	return active
}

// LoadOrDefault loads config from the default location or returns a default config
func LoadOrDefault(path string) (*Config, error) {
	home, _ := os.UserHomeDir()

	cfg, err := LoadFromPath(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{
				BackupDir:       filepath.Join(home, ".dotfiles"),
				GitRemote:       "https://github.com/user/dotfiles.git",
				Files:           []string{},
				Folders:         []string{".config"},
				Schedule:        "0 2 * * *",
				Notifications:   true,
				Exclude:         []string{},
				DisabledFiles:   []string{},
				DisabledFolders: []string{},
			}, nil
		}
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return cfg, nil
}
