package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad_DefaultLocation tests loading config from XDG default location
func TestLoad_DefaultLocation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, "dotkeeper")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write test config
	configPath := filepath.Join(configDir, "config.yaml")
	testConfig := `backup_dir: /tmp/backup
files:
  - .bashrc
`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load using default location
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.BackupDir != "/tmp/backup" {
		t.Errorf("Expected BackupDir '/tmp/backup', got '%s'", cfg.BackupDir)
	}
}

// TestLoad_ConfigNotExists tests loading when config doesn't exist
func TestLoad_ConfigNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir+"/nonexistent")

	_, err := Load()
	if err == nil {
		t.Error("Load should fail when config doesn't exist")
	}
}

// TestLoad_InvalidYAML tests loading invalid YAML
func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, "dotkeeper")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write invalid YAML
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("backup_dir: [invalid"), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load should fail with invalid YAML")
	}
}

// TestLoad_GetConfigPathFallsBackToHome tests GetConfigPath when XDG_CONFIG_HOME is not set
func TestLoad_GetConfigPathFallsBackToHome(t *testing.T) {
	// Clear XDG_CONFIG_HOME to test fallback to ~/.config
	tmpDir := t.TempDir()
	homeBackup := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		if homeBackup != "" {
			os.Setenv("HOME", homeBackup)
		}
	})
	t.Setenv("XDG_CONFIG_HOME", "")

	// Create a config in the expected fallback location
	configDir := filepath.Join(tmpDir, ".config", "dotkeeper")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	testConfig := `backup_dir: /tmp/backup
files:
  - .bashrc
`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load should work even without XDG_CONFIG_HOME
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config without XDG_CONFIG_HOME: %v", err)
	}

	if cfg.BackupDir != "/tmp/backup" {
		t.Errorf("Expected BackupDir '/tmp/backup', got '%s'", cfg.BackupDir)
	}
}

// TestLoad_GetConfigPathFails tests GetConfigPath when home directory cannot be determined
func TestLoad_GetConfigPathFails(t *testing.T) {
	// Set HOME to an invalid path and clear XDG_CONFIG_HOME
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/nonexistent/directory/that/does/not/exist")

	_, err := Load()
	if err == nil {
		t.Error("Load should fail when home directory cannot be determined")
	}
}

// TestSave_DefaultLocation tests saving config to XDG default location
func TestSave_DefaultLocation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := &Config{
		BackupDir:     "/tmp/backup",
		Files:         []string{".bashrc"},
		Folders:       []string{".config"},
		Schedule:      "0 2 * * *",
		Notifications: true,
	}

	// Save using default location
	if err := cfg.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, "dotkeeper", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file was not created: %v", err)
	}

	// Verify content
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.BackupDir != cfg.BackupDir {
		t.Errorf("BackupDir mismatch after roundtrip")
	}
}

// TestSave_DirectoryCreation tests that Save creates the config directory
func TestSave_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := &Config{
		BackupDir: "/tmp/backup",
		Files:     []string{".bashrc"},
	}

	// Config directory doesn't exist yet
	configDir := filepath.Join(tmpDir, "dotkeeper")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// Expected - directory doesn't exist yet
	} else {
		t.Fatalf("Config directory should not exist yet")
	}

	// Save should create the directory
	if err := cfg.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(configDir); err != nil {
		t.Fatalf("Config directory was not created: %v", err)
	}
}

// TestSave_InvalidDirectory tests saving when config directory cannot be created
func TestSave_InvalidDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a file where XDG_CONFIG_HOME should be a directory
	invalidPath := filepath.Join(tmpDir, "not-a-directory")
	if err := os.WriteFile(invalidPath, []byte("x"), 0644); err != nil {
		t.Fatalf("Failed to create invalid path: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", invalidPath)

	cfg := &Config{
		BackupDir: "/tmp/backup",
		Files:     []string{".bashrc"},
	}

	err := cfg.Save()
	if err == nil {
		t.Error("Save should fail when config directory cannot be created")
	}
}

// TestLoadFromPath_InvalidPath tests loading from a non-existent path
func TestLoadFromPath_InvalidPath(t *testing.T) {
	_, err := LoadFromPath("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadFromPath should fail for non-existent file")
	}
}

// TestLoadOrDefault_FileExists tests LoadOrDefault when file exists
func TestLoadOrDefault_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create a config file
	testConfig := `backup_dir: /my/backup
files:
  - .bashrc
`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// LoadOrDefault should load the existing file
	cfg, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("LoadOrDefault failed: %v", err)
	}

	if cfg.BackupDir != "/my/backup" {
		t.Errorf("Expected BackupDir '/my/backup', got '%s'", cfg.BackupDir)
	}
}

// TestLoadOrDefault_FileNotExists tests LoadOrDefault when file doesn't exist
func TestLoadOrDefault_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// File doesn't exist
	cfg, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("LoadOrDefault should return default config: %v", err)
	}

	// Should have default values
	if cfg.BackupDir == "" {
		t.Error("Default BackupDir should not be empty")
	}
	if cfg.GitRemote == "" {
		t.Error("Default GitRemote should not be empty")
	}
	if cfg.Notifications != true {
		t.Error("Default Notifications should be true")
	}
}
