package restore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupExisting_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("original content")

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	backupPath, err := BackupExisting(filePath)
	if err != nil {
		t.Fatalf("BackupExisting failed: %v", err)
	}

	// Verify backup was created
	if backupPath == "" {
		t.Error("Expected backup path, got empty string")
	}

	if !strings.Contains(backupPath, ".bak.") {
		t.Errorf("Expected backup path to contain '.bak.', got: %s", backupPath)
	}

	// Verify original file no longer exists
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Expected original file to be renamed")
	}

	// Verify backup content matches original
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}

	if string(backupContent) != string(content) {
		t.Errorf("Backup content mismatch: got %s, want %s", backupContent, content)
	}
}

func TestBackupExisting_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.txt")

	backupPath, err := BackupExisting(filePath)
	if err != nil {
		t.Fatalf("BackupExisting failed: %v", err)
	}

	if backupPath != "" {
		t.Errorf("Expected empty backup path for non-existent file, got: %s", backupPath)
	}
}

func TestEnsureUniquePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first backup
	path1 := filepath.Join(tmpDir, "test.bak")
	if err := os.WriteFile(path1, []byte("backup1"), 0644); err != nil {
		t.Fatal(err)
	}

	// Get unique path
	uniquePath := ensureUniquePath(path1)

	if uniquePath == path1 {
		t.Error("Expected unique path different from original")
	}

	if !strings.HasPrefix(uniquePath, path1) {
		t.Errorf("Expected unique path to start with %s, got %s", path1, uniquePath)
	}
}

func TestEnsureUniquePath_NoCollision(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.bak")

	uniquePath := ensureUniquePath(path)

	if uniquePath != path {
		t.Errorf("Expected same path when no collision: got %s, want %s", uniquePath, path)
	}
}

func TestResolveConflict_NoConflict(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.txt")

	action := ResolveConflict(filePath, RestoreOptions{})

	if action != ActionOverwrite {
		t.Errorf("Expected ActionOverwrite for non-existent file, got %v", action)
	}
}

func TestResolveConflict_WithConflict(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Without Force option
	action := ResolveConflict(filePath, RestoreOptions{Force: false})
	if action != ActionBackup {
		t.Errorf("Expected ActionBackup without Force, got %v", action)
	}

	// With Force option
	action = ResolveConflict(filePath, RestoreOptions{Force: true})
	if action != ActionOverwrite {
		t.Errorf("Expected ActionOverwrite with Force, got %v", action)
	}
}

func TestHasConflict(t *testing.T) {
	tmpDir := t.TempDir()

	// Non-existent file
	nonExistent := filepath.Join(tmpDir, "nonexistent.txt")
	if HasConflict(nonExistent) {
		t.Error("Expected no conflict for non-existent file")
	}

	// Existing file
	existing := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existing, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	if !HasConflict(existing) {
		t.Error("Expected conflict for existing file")
	}
}

func TestCleanupBackups(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "test.txt")

	// Create some backup files
	backup1 := basePath + ".bak.20240101"
	backup2 := basePath + ".bak.20240102"
	backup3 := basePath + ".bak.20240103"

	for _, path := range []string{backup1, backup2, backup3} {
		if err := os.WriteFile(path, []byte("backup"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := CleanupBackups(basePath); err != nil {
		t.Fatalf("CleanupBackups failed: %v", err)
	}

	// Verify all backups are removed
	for _, path := range []string{backup1, backup2, backup3} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Expected backup %s to be removed", path)
		}
	}
}

func TestRestoreFromBackup(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := filepath.Join(tmpDir, "test.txt")
	backupPath := filepath.Join(tmpDir, "test.txt.bak")

	// Create current file
	if err := os.WriteFile(originalPath, []byte("current"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create backup file
	if err := os.WriteFile(backupPath, []byte("backup"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RestoreFromBackup(originalPath, backupPath); err != nil {
		t.Fatalf("RestoreFromBackup failed: %v", err)
	}

	// Verify content was restored
	content, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(content) != "backup" {
		t.Errorf("Expected restored content 'backup', got '%s'", content)
	}

	// Verify backup file is gone
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Expected backup file to be removed after restore")
	}
}

func TestListBackups(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "test.txt")

	// Create some backup files
	backups := []string{
		basePath + ".bak.20240101",
		basePath + ".bak.20240102",
	}

	for _, path := range backups {
		if err := os.WriteFile(path, []byte("backup"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create an unrelated file
	if err := os.WriteFile(filepath.Join(tmpDir, "other.txt"), []byte("other"), 0644); err != nil {
		t.Fatal(err)
	}

	found, err := ListBackups(basePath)
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 backups, found %d", len(found))
	}
}
