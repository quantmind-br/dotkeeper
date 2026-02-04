// Package e2e provides end-to-end integration tests for dotkeeper.
// These tests verify the complete backup and restore cycle.
//
// To run these tests:
//
//	go test ./e2e -v
//
// To skip e2e tests (e.g., in CI without full environment):
//
//	go test ./e2e -v -short
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/diogo/dotkeeper/internal/backup"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/restore"
)

const testPassword = "e2e-test-password-secure"

// skipIfShort skips the test if running with -short flag
func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
}

// TestBackupRestoreCycle tests the complete backup -> restore flow
func TestBackupRestoreCycle(t *testing.T) {
	skipIfShort(t)

	// Create temporary test environment
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	backupDir := filepath.Join(tempDir, "backups")
	restoreDir := filepath.Join(tempDir, "restore")

	// Create source directory with test files
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	// Create test files with various content
	testFiles := map[string]string{
		"config.yaml": "key: value\nname: test",
		"script.sh":   "#!/bin/bash\necho 'hello'",
		".bashrc":     "export PATH=$PATH:/usr/local/bin",
	}

	for name, content := range testFiles {
		path := filepath.Join(sourceDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file %s: %v", name, err)
		}
	}

	// Create config
	cfg := &config.Config{
		BackupDir: backupDir,
		GitRemote: "https://github.com/test/dotfiles.git",
		Files:     []string{filepath.Join(sourceDir, "config.yaml")},
		Folders:   []string{sourceDir},
	}

	// Step 1: Run backup
	t.Log("Step 1: Running backup...")
	result, err := backup.Backup(cfg, testPassword)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Verify backup was created
	if result.FileCount == 0 {
		t.Error("backup created no files")
	}
	if _, err := os.Stat(result.BackupPath); os.IsNotExist(err) {
		t.Errorf("backup file not created: %s", result.BackupPath)
	}
	if _, err := os.Stat(result.MetadataPath); os.IsNotExist(err) {
		t.Errorf("metadata file not created: %s", result.MetadataPath)
	}

	t.Logf("Backup created: %s (%d files, %d bytes)",
		result.BackupName, result.FileCount, result.TotalSize)

	// Step 2: List backup contents
	t.Log("Step 2: Listing backup contents...")
	contents, err := restore.ListBackupContents(result.BackupPath, testPassword)
	if err != nil {
		t.Fatalf("failed to list backup contents: %v", err)
	}
	if len(contents) == 0 {
		t.Error("backup appears empty")
	}
	t.Logf("Backup contains %d files", len(contents))

	// Step 3: Restore to new directory
	t.Log("Step 3: Restoring backup...")
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatalf("failed to create restore dir: %v", err)
	}

	opts := restore.RestoreOptions{
		TargetDir: restoreDir,
		Force:     true,
	}

	restoreResult, err := restore.Restore(result.BackupPath, testPassword, opts)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	t.Logf("Restored %d files", restoreResult.FilesRestored)

	// Step 4: Verify restored files match originals
	t.Log("Step 4: Verifying restored files...")
	for name, expectedContent := range testFiles {
		restoredPath := filepath.Join(restoreDir, name)
		content, err := os.ReadFile(restoredPath)
		if err != nil {
			t.Errorf("failed to read restored file %s: %v", name, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("content mismatch for %s:\n  expected: %q\n  got: %q",
				name, expectedContent, string(content))
		}
	}

	t.Log("Backup/restore cycle completed successfully")
}

// TestBackupRestoreWithSubdirectories tests backup/restore with nested directories
func TestBackupRestoreWithSubdirectories(t *testing.T) {
	skipIfShort(t)

	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	backupDir := filepath.Join(tempDir, "backups")
	restoreDir := filepath.Join(tempDir, "restore")

	// Create nested directory structure
	dirs := []string{
		filepath.Join(sourceDir, "config", "nvim"),
		filepath.Join(sourceDir, "config", "tmux"),
		filepath.Join(sourceDir, "scripts"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	// Create files in nested directories
	nestedFiles := map[string]string{
		filepath.Join(sourceDir, "config", "nvim", "init.lua"):  "vim.opt.number = true",
		filepath.Join(sourceDir, "config", "tmux", "tmux.conf"): "set -g prefix C-a",
		filepath.Join(sourceDir, "scripts", "backup.sh"):        "#!/bin/bash\necho backup",
	}

	for path, content := range nestedFiles {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", path, err)
		}
	}

	// Create config to back up the source folder
	cfg := &config.Config{
		BackupDir: backupDir,
		GitRemote: "https://github.com/test/dotfiles.git",
		Folders:   []string{sourceDir},
	}

	// Backup
	result, err := backup.Backup(cfg, testPassword)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Restore
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatalf("failed to create restore dir: %v", err)
	}

	opts := restore.RestoreOptions{
		TargetDir: restoreDir,
		Force:     true,
	}

	_, err = restore.Restore(result.BackupPath, testPassword, opts)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	// Verify nested files were restored
	for _, content := range nestedFiles {
		// Files are restored to targetDir with just the basename
		// This is expected behavior per the restore implementation
		found := false
		err := filepath.Walk(restoreDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			data, _ := os.ReadFile(path)
			if string(data) == content {
				found = true
			}
			return nil
		})
		if err != nil {
			t.Errorf("walk error: %v", err)
		}
		if !found {
			t.Errorf("content not found in restore: %q", content)
		}
	}
}

// TestBackupValidation tests backup validation
func TestBackupValidation(t *testing.T) {
	skipIfShort(t)

	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	backupDir := filepath.Join(tempDir, "backups")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	// Create test file
	testFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := &config.Config{
		BackupDir: backupDir,
		GitRemote: "https://github.com/test/dotfiles.git",
		Files:     []string{testFile},
	}

	// Create backup
	result, err := backup.Backup(cfg, testPassword)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Test validation with correct password
	t.Log("Testing validation with correct password...")
	if err := restore.ValidateBackup(result.BackupPath, testPassword); err != nil {
		t.Errorf("validation should pass with correct password: %v", err)
	}

	// Test validation with wrong password
	t.Log("Testing validation with wrong password...")
	err = restore.ValidateBackup(result.BackupPath, "wrong-password")
	if err == nil {
		t.Error("validation should fail with wrong password")
	}
}

// TestDryRun tests dry-run mode
func TestDryRun(t *testing.T) {
	skipIfShort(t)

	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	backupDir := filepath.Join(tempDir, "backups")
	restoreDir := filepath.Join(tempDir, "restore")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatalf("failed to create restore dir: %v", err)
	}

	// Create test file
	testFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("original content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := &config.Config{
		BackupDir: backupDir,
		GitRemote: "https://github.com/test/dotfiles.git",
		Files:     []string{testFile},
	}

	// Create backup
	result, err := backup.Backup(cfg, testPassword)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Run restore in dry-run mode
	opts := restore.RestoreOptions{
		TargetDir: restoreDir,
		DryRun:    true,
	}

	restoreResult, err := restore.Restore(result.BackupPath, testPassword, opts)
	if err != nil {
		t.Fatalf("dry-run restore failed: %v", err)
	}

	// Verify no files were actually restored
	if restoreResult.FilesRestored != 0 {
		t.Errorf("dry-run should not restore files, got %d", restoreResult.FilesRestored)
	}

	// Verify target directory is empty (no files created)
	entries, err := os.ReadDir(restoreDir)
	if err != nil {
		t.Fatalf("failed to read restore dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("dry-run should not create files, found %d entries", len(entries))
	}
}
