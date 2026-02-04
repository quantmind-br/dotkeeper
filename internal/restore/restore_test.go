package restore

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/backup"
	"github.com/diogo/dotkeeper/internal/config"
)

// createTestBackup creates a test backup for use in restore tests
func createTestBackup(t *testing.T, tmpDir string, files map[string]string) (string, string) {
	t.Helper()

	backupDir := filepath.Join(tmpDir, "backups")
	sourceDir := filepath.Join(tmpDir, "source")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	var filePaths []string
	for name, content := range files {
		filePath := filepath.Join(sourceDir, name)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		filePaths = append(filePaths, filePath)
	}

	cfg := &config.Config{
		BackupDir: backupDir,
		Files:     filePaths,
		Folders:   []string{},
	}

	password := "test-password-123"

	result, err := backup.Backup(cfg, password)
	if err != nil {
		t.Fatalf("Failed to create test backup: %v", err)
	}

	return result.BackupPath, password
}

func TestRestore_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test backup
	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	// Create a restore directory
	restoreDir := filepath.Join(tmpDir, "restore")
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatal(err)
	}

	opts := RestoreOptions{
		TargetDir: restoreDir,
	}

	result, err := Restore(backupPath, password, opts)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify result
	if result.TotalFiles != 2 {
		t.Errorf("Expected 2 total files, got %d", result.TotalFiles)
	}

	if result.FilesRestored != 2 {
		t.Errorf("Expected 2 restored files, got %d", result.FilesRestored)
	}

	// Verify files were restored
	for name := range files {
		restoredPath := filepath.Join(restoreDir, name)
		if _, err := os.Stat(restoredPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be restored", name)
		}
	}
}

func TestRestore_DryRun(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "content1",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	restoreDir := filepath.Join(tmpDir, "restore")
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatal(err)
	}

	opts := RestoreOptions{
		TargetDir: restoreDir,
		DryRun:    true,
	}

	result, err := Restore(backupPath, password, opts)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// In dry run, files should be skipped, not restored
	if result.FilesRestored != 0 {
		t.Errorf("Expected 0 restored files in dry run, got %d", result.FilesRestored)
	}

	if result.FilesSkipped != 1 {
		t.Errorf("Expected 1 skipped file in dry run, got %d", result.FilesSkipped)
	}

	// Verify no files were actually created
	restoredPath := filepath.Join(restoreDir, "file1.txt")
	if _, err := os.Stat(restoredPath); !os.IsNotExist(err) {
		t.Error("Expected no files to be created in dry run")
	}
}

func TestRestore_WithConflict(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "backup content",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	restoreDir := filepath.Join(tmpDir, "restore")
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create an existing file that will conflict
	existingPath := filepath.Join(restoreDir, "file1.txt")
	if err := os.WriteFile(existingPath, []byte("existing content"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := RestoreOptions{
		TargetDir: restoreDir,
	}

	result, err := Restore(backupPath, password, opts)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify backup was created
	if result.FilesConflict != 1 {
		t.Errorf("Expected 1 conflict, got %d", result.FilesConflict)
	}

	if len(result.BackupFiles) != 1 {
		t.Errorf("Expected 1 backup file, got %d", len(result.BackupFiles))
	}

	// Verify the restored content
	restoredContent, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(restoredContent) != "backup content" {
		t.Errorf("Expected restored content 'backup content', got '%s'", restoredContent)
	}

	// Verify .bak file exists
	backupFound := false
	for _, bakPath := range result.BackupFiles {
		if strings.Contains(bakPath, ".bak.") {
			backupFound = true
			content, err := os.ReadFile(bakPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(content) != "existing content" {
				t.Errorf("Backup content mismatch: got '%s'", content)
			}
		}
	}

	if !backupFound {
		t.Error("Expected .bak file to be created")
	}
}

func TestRestore_WithForce(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "backup content",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	restoreDir := filepath.Join(tmpDir, "restore")
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create an existing file
	existingPath := filepath.Join(restoreDir, "file1.txt")
	if err := os.WriteFile(existingPath, []byte("existing content"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := RestoreOptions{
		TargetDir: restoreDir,
		Force:     true,
	}

	result, err := Restore(backupPath, password, opts)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// With Force, no backup should be created
	if len(result.BackupFiles) != 0 {
		t.Errorf("Expected 0 backup files with Force option, got %d", len(result.BackupFiles))
	}

	// Verify content was overwritten
	content, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "backup content" {
		t.Errorf("Expected content to be overwritten, got '%s'", content)
	}
}

func TestRestore_ShowDiff(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "new line\n",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	restoreDir := filepath.Join(tmpDir, "restore")
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create existing file with different content
	existingPath := filepath.Join(restoreDir, "file1.txt")
	if err := os.WriteFile(existingPath, []byte("old line\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var diffOutput bytes.Buffer
	opts := RestoreOptions{
		TargetDir:  restoreDir,
		ShowDiff:   true,
		DiffWriter: &diffOutput,
		DryRun:     true,
	}

	result, err := Restore(backupPath, password, opts)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify diff was generated
	if len(result.DiffResults) == 0 {
		t.Error("Expected diff results to be generated")
	}

	diffStr := diffOutput.String()
	if diffStr == "" {
		t.Error("Expected diff output to be written")
	}
}

func TestRestore_SelectedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
		"file3.txt": "content3",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	restoreDir := filepath.Join(tmpDir, "restore")
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatal(err)
	}

	opts := RestoreOptions{
		TargetDir:     restoreDir,
		SelectedFiles: []string{"file1.txt", "file3.txt"},
	}

	result, err := Restore(backupPath, password, opts)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Only 2 files should be restored
	if result.FilesRestored != 2 {
		t.Errorf("Expected 2 restored files, got %d", result.FilesRestored)
	}

	// file2.txt should not exist
	file2Path := filepath.Join(restoreDir, "file2.txt")
	if _, err := os.Stat(file2Path); !os.IsNotExist(err) {
		t.Error("file2.txt should not be restored when not in SelectedFiles")
	}
}

func TestRestore_WrongPassword(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "content1",
	}
	backupPath, _ := createTestBackup(t, tmpDir, files)

	restoreDir := filepath.Join(tmpDir, "restore")

	opts := RestoreOptions{
		TargetDir: restoreDir,
	}

	_, err := Restore(backupPath, "wrong-password", opts)
	if err == nil {
		t.Error("Expected error with wrong password")
	}

	if !strings.Contains(err.Error(), "decryption failed") {
		t.Errorf("Expected decryption error, got: %v", err)
	}
}

func TestRestore_MissingBackupFile(t *testing.T) {
	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "nonexistent.tar.gz.enc")

	opts := RestoreOptions{}

	_, err := Restore(backupPath, "password", opts)
	if err == nil {
		t.Error("Expected error with missing backup file")
	}
}

func TestListBackupContents(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	entries, err := ListBackupContents(backupPath, password)
	if err != nil {
		t.Fatalf("ListBackupContents failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestValidateBackup(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "content1",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	// Valid backup with correct password
	if err := ValidateBackup(backupPath, password); err != nil {
		t.Errorf("ValidateBackup failed for valid backup: %v", err)
	}

	// Wrong password
	if err := ValidateBackup(backupPath, "wrong"); err == nil {
		t.Error("Expected error for wrong password")
	}

	// Missing backup
	if err := ValidateBackup(filepath.Join(tmpDir, "nonexistent"), password); err == nil {
		t.Error("Expected error for missing backup")
	}
}

func TestExtractTarGz(t *testing.T) {
	// Create a simple tar.gz in memory for testing
	// This tests the extractTarGz function with known good data
	tmpDir := t.TempDir()

	files := map[string]string{
		"test.txt": "test content",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	// Use ListBackupContents which internally calls extractTarGz
	entries, err := ListBackupContents(backupPath, password)
	if err != nil {
		t.Fatalf("extractTarGz (via ListBackupContents) failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	// Verify content
	found := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Path, "test.txt") {
			found = true
			if string(entry.Content) != "test content" {
				t.Errorf("Content mismatch: got %s", entry.Content)
			}
		}
	}

	if !found {
		t.Error("Expected to find test.txt in entries")
	}
}

func TestRestoreFileAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "subdir", "test.txt")
	content := []byte("test content")
	mode := int64(0644)

	if err := restoreFileAtomic(filePath, content, mode); err != nil {
		t.Fatalf("restoreFileAtomic failed: %v", err)
	}

	// Verify file was created
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !bytes.Equal(readContent, content) {
		t.Error("Content mismatch")
	}

	// Verify permissions
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != os.FileMode(mode) {
		t.Errorf("Permission mismatch: got %o, want %o", info.Mode().Perm(), mode)
	}
}

func TestFilterEntries(t *testing.T) {
	entries := []FileEntry{
		{Path: "/home/user/file1.txt"},
		{Path: "/home/user/file2.txt"},
		{Path: "/home/user/dir/file3.txt"},
	}

	// Filter by exact path
	filtered := filterEntries(entries, []string{"/home/user/file1.txt"})
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered entry, got %d", len(filtered))
	}

	// Filter by base name
	filtered = filterEntries(entries, []string{"file2.txt"})
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered entry by basename, got %d", len(filtered))
	}

	// Empty filter returns all
	filtered = filterEntries(entries, []string{})
	if len(filtered) != 3 {
		t.Errorf("Expected all entries with empty filter, got %d", len(filtered))
	}
}

func TestProgressCallback(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "content1",
	}
	backupPath, password := createTestBackup(t, tmpDir, files)

	restoreDir := filepath.Join(tmpDir, "restore")
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		t.Fatal(err)
	}

	var callbackCalls []string
	opts := RestoreOptions{
		TargetDir: restoreDir,
		ProgressCallback: func(file, action string) {
			callbackCalls = append(callbackCalls, action)
		},
	}

	_, err := Restore(backupPath, password, opts)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if len(callbackCalls) == 0 {
		t.Error("Expected progress callback to be called")
	}

	foundRestored := false
	for _, call := range callbackCalls {
		if call == "restored" {
			foundRestored = true
		}
	}

	if !foundRestored {
		t.Error("Expected 'restored' callback")
	}
}
