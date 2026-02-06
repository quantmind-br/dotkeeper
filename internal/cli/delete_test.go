package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupDeleteTest(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Set up config with backup dir
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	configDir := filepath.Join(tmpDir, "dotkeeper")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := "backup_dir: " + tmpDir + "\nfiles:\n  - .bashrc\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create a backup file
	backupPath := filepath.Join(tmpDir, "backup-2025-01-01-120000.tar.gz.enc")
	if err := os.WriteFile(backupPath, []byte("dummy backup"), 0644); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Create a metadata file
	metaPath := backupPath + ".meta.json"
	if err := os.WriteFile(metaPath, []byte(`{"version": "1.0"}`), 0644); err != nil {
		t.Fatalf("Failed to create metadata: %v", err)
	}

	return tmpDir
}

func TestDeleteCommand_WithForce(t *testing.T) {
	tmpDir := setupDeleteTest(t)

	// Test with force flag (flag before backup name)
	var exit int
	stdout, stderr := captureStdoutStderr(t, func() {
		exit = DeleteCommand([]string{"--force", "backup-2025-01-01-120000"})
	})

	if exit != 0 {
		t.Fatalf("exit = %d, stderr=%s", exit, stderr)
	}

	// Check backup was deleted
	backupPath := filepath.Join(tmpDir, "backup-2025-01-01-120000.tar.gz.enc")
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("backup file should be deleted")
	}

	// Check metadata was deleted
	metaPath := backupPath + ".meta.json"
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Error("metadata file should be deleted")
	}

	if !strings.Contains(stdout, "Deleted backup-2025-01-01-120000.tar.gz.enc") {
		t.Errorf("expected deletion message in stdout: %s", stdout)
	}
}

func TestDeleteCommand_WithoutExtension(t *testing.T) {
	_ = setupDeleteTest(t)

	// Test without .tar.gz.enc extension
	var exit int
	stdout, stderr := captureStdoutStderr(t, func() {
		exit = DeleteCommand([]string{"--force", "backup-2025-01-01-120000"})
	})

	if exit != 0 {
		t.Fatalf("exit = %d, stderr=%s", exit, stderr)
	}

	if !strings.Contains(stdout, "Deleted backup-2025-01-01-120000.tar.gz.enc") {
		t.Error("should add extension automatically")
	}
}

func TestDeleteCommand_NonExistentBackup(t *testing.T) {
	_ = setupDeleteTest(t)

	var exit int
	_, stderr := captureStdoutStderr(t, func() {
		exit = DeleteCommand([]string{"--force", "backup-does-not-exist"})
	})

	if exit != 1 {
		t.Fatalf("expected exit 1, got %d", exit)
	}

	if !strings.Contains(stderr, "backup not found") {
		t.Errorf("expected 'backup not found' error, got: %s", stderr)
	}
}

func TestDeleteCommand_NoArguments(t *testing.T) {
	_ = setupDeleteTest(t)

	var exit int
	_, stderr := captureStdoutStderr(t, func() {
		exit = DeleteCommand([]string{})
	})

	if exit != 1 {
		t.Fatalf("expected exit 1, got %d", exit)
	}

	if !strings.Contains(stderr, "backup name required") {
		t.Errorf("expected 'backup name required' error, got: %s", stderr)
	}
}

func TestDeleteCommand_MetadataOnly(t *testing.T) {
	tmpDir := setupDeleteTest(t)

	// Remove the backup file but keep metadata
	backupPath := filepath.Join(tmpDir, "backup-2025-01-01-120000.tar.gz.enc")
	if err := os.Remove(backupPath); err != nil {
		t.Logf("Warning: could not remove backup for test: %v", err)
	}

	var exit int
	_, stderr := captureStdoutStderr(t, func() {
		exit = DeleteCommand([]string{"--force", "backup-2025-01-01-120000"})
	})

	if exit != 1 {
		t.Fatalf("expected exit 1, got %d", exit)
	}

	if !strings.Contains(stderr, "backup not found") {
		t.Errorf("expected 'backup not found' error, got: %s", stderr)
	}
}
