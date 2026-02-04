package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfigStruct verifies the Config struct has all required fields
func TestConfigStruct(t *testing.T) {
	cfg := &Config{
		BackupDir:     "/tmp/backup",
		GitRemote:     "https://github.com/user/repo.git",
		Files:         []string{".bashrc", ".zshrc"},
		Folders:       []string{".config"},
		Schedule:      "0 2 * * *",
		Notifications: true,
	}

	if cfg.BackupDir != "/tmp/backup" {
		t.Errorf("BackupDir not set correctly")
	}
	if cfg.GitRemote != "https://github.com/user/repo.git" {
		t.Errorf("GitRemote not set correctly")
	}
	if len(cfg.Files) != 2 {
		t.Errorf("Files not set correctly")
	}
	if len(cfg.Folders) != 1 {
		t.Errorf("Folders not set correctly")
	}
	if cfg.Schedule != "0 2 * * *" {
		t.Errorf("Schedule not set correctly")
	}
	if !cfg.Notifications {
		t.Errorf("Notifications not set correctly")
	}
}

// TestLoadConfig tests loading config from YAML file
func TestLoadConfig(t *testing.T) {
	// Create temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write test config file
	testConfig := `backup_dir: /tmp/backup
git_remote: https://github.com/user/repo.git
files:
  - .bashrc
  - .zshrc
folders:
  - .config
schedule: "0 2 * * *"
notifications: true
`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config
	if cfg.BackupDir != "/tmp/backup" {
		t.Errorf("Expected BackupDir '/tmp/backup', got '%s'", cfg.BackupDir)
	}
	if cfg.GitRemote != "https://github.com/user/repo.git" {
		t.Errorf("Expected GitRemote 'https://github.com/user/repo.git', got '%s'", cfg.GitRemote)
	}
	if len(cfg.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(cfg.Files))
	}
	if len(cfg.Folders) != 1 {
		t.Errorf("Expected 1 folder, got %d", len(cfg.Folders))
	}
	if cfg.Schedule != "0 2 * * *" {
		t.Errorf("Expected Schedule '0 2 * * *', got '%s'", cfg.Schedule)
	}
	if !cfg.Notifications {
		t.Errorf("Expected Notifications true, got false")
	}
}

// TestSaveConfig tests saving config to YAML file
func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		BackupDir:     "/tmp/backup",
		GitRemote:     "https://github.com/user/repo.git",
		Files:         []string{".bashrc", ".zshrc"},
		Folders:       []string{".config"},
		Schedule:      "0 2 * * *",
		Notifications: true,
	}

	// Save config
	if err := cfg.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	// Load it back and verify
	loaded, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.BackupDir != cfg.BackupDir {
		t.Errorf("BackupDir mismatch: expected '%s', got '%s'", cfg.BackupDir, loaded.BackupDir)
	}
	if loaded.GitRemote != cfg.GitRemote {
		t.Errorf("GitRemote mismatch: expected '%s', got '%s'", cfg.GitRemote, loaded.GitRemote)
	}
}

// TestValidateConfig tests config validation
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				BackupDir:     "/tmp/backup",
				GitRemote:     "https://github.com/user/repo.git",
				Files:         []string{".bashrc"},
				Folders:       []string{".config"},
				Schedule:      "0 2 * * *",
				Notifications: true,
			},
			wantErr: false,
		},
		{
			name: "empty backup dir",
			cfg: &Config{
				BackupDir:     "",
				GitRemote:     "https://github.com/user/repo.git",
				Files:         []string{".bashrc"},
				Folders:       []string{".config"},
				Schedule:      "0 2 * * *",
				Notifications: true,
			},
			wantErr: true,
		},
		{
			name: "empty git remote",
			cfg: &Config{
				BackupDir:     "/tmp/backup",
				GitRemote:     "",
				Files:         []string{".bashrc"},
				Folders:       []string{".config"},
				Schedule:      "0 2 * * *",
				Notifications: true,
			},
			wantErr: true,
		},
		{
			name: "no files or folders",
			cfg: &Config{
				BackupDir:     "/tmp/backup",
				GitRemote:     "https://github.com/user/repo.git",
				Files:         []string{},
				Folders:       []string{},
				Schedule:      "0 2 * * *",
				Notifications: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDefaultConfig tests loading default config
func TestDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Load default config (file doesn't exist)
	cfg, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	// Verify default values are set
	if cfg.BackupDir == "" {
		t.Errorf("Default BackupDir should not be empty")
	}
	if cfg.GitRemote == "" {
		t.Errorf("Default GitRemote should not be empty")
	}
}

// TestConfigRoundtrip tests saving and loading config
func TestConfigRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	original := &Config{
		BackupDir:     "/home/user/.dotfiles",
		GitRemote:     "https://github.com/user/dotfiles.git",
		Files:         []string{".bashrc", ".zshrc", ".vimrc"},
		Folders:       []string{".config", ".ssh"},
		Schedule:      "0 2 * * *",
		Notifications: true,
	}

	// Save
	if err := original.SaveToPath(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load
	loaded, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Compare
	if loaded.BackupDir != original.BackupDir {
		t.Errorf("BackupDir mismatch")
	}
	if loaded.GitRemote != original.GitRemote {
		t.Errorf("GitRemote mismatch")
	}
	if len(loaded.Files) != len(original.Files) {
		t.Errorf("Files count mismatch")
	}
	if len(loaded.Folders) != len(original.Folders) {
		t.Errorf("Folders count mismatch")
	}
	if loaded.Schedule != original.Schedule {
		t.Errorf("Schedule mismatch")
	}
	if loaded.Notifications != original.Notifications {
		t.Errorf("Notifications mismatch")
	}
}
