package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/crypto"
)

func TestBackup(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create config
	cfg := &config.Config{
		BackupDir: backupDir,
		Files:     []string{file1, file2},
		Folders:   []string{},
	}

	password := "test-password-123"

	// Run backup
	result, err := Backup(cfg, password)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.FileCount != 2 {
		t.Errorf("Expected 2 files, got %d", result.FileCount)
	}

	if result.TotalSize == 0 {
		t.Error("Expected non-zero total size")
	}

	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}

	if result.Checksum == "" {
		t.Error("Expected non-empty checksum")
	}

	// Verify backup file exists
	if _, err := os.Stat(result.BackupPath); os.IsNotExist(err) {
		t.Errorf("Backup file does not exist: %s", result.BackupPath)
	}

	// Verify metadata file exists
	if _, err := os.Stat(result.MetadataPath); os.IsNotExist(err) {
		t.Errorf("Metadata file does not exist: %s", result.MetadataPath)
	}

	// Verify backup name format
	if !strings.HasPrefix(result.BackupName, "backup-") {
		t.Errorf("Expected backup name to start with 'backup-', got %s", result.BackupName)
	}
	if !strings.HasSuffix(result.BackupName, ".tar.gz.enc") {
		t.Errorf("Expected backup name to end with '.tar.gz.enc', got %s", result.BackupName)
	}

	// Verify metadata content
	metadataBytes, err := os.ReadFile(result.MetadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	var metadata crypto.EncryptionMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}

	if metadata.Version != crypto.Version {
		t.Errorf("Expected version %d, got %d", crypto.Version, metadata.Version)
	}

	if metadata.Algorithm != "AES-256-GCM" {
		t.Errorf("Expected algorithm AES-256-GCM, got %s", metadata.Algorithm)
	}

	if metadata.KDF != "Argon2id" {
		t.Errorf("Expected KDF Argon2id, got %s", metadata.KDF)
	}

	if len(metadata.Salt) == 0 {
		t.Error("Expected non-empty salt")
	}

	if metadata.OriginalSize == 0 {
		t.Error("Expected non-zero original size")
	}

	// Verify backup file permissions
	info, err := os.Stat(result.BackupPath)
	if err != nil {
		t.Fatalf("Failed to stat backup file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected backup file permissions 0600, got %o", info.Mode().Perm())
	}

	// Verify we can decrypt the backup
	encryptedData, err := os.ReadFile(result.BackupPath)
	if err != nil {
		t.Fatalf("Failed to read encrypted backup: %v", err)
	}

	key := crypto.DeriveKey(password, metadata.Salt)
	decrypted, err := crypto.Decrypt(encryptedData, key)
	if err != nil {
		t.Fatalf("Failed to decrypt backup: %v", err)
	}

	if len(decrypted) == 0 {
		t.Error("Expected non-empty decrypted data")
	}
}

func TestBackup_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")

	cfg := &config.Config{
		BackupDir: backupDir,
		Files:     []string{},
		Folders:   []string{},
	}

	password := "test-password-123"

	result, err := Backup(cfg, password)

	if err == nil {
		t.Error("Expected error when no files to backup")
	}

	if result != nil {
		t.Error("Expected nil result when backup fails")
	}

	if !strings.Contains(err.Error(), "no files to backup") {
		t.Errorf("Expected 'no files to backup' error, got: %v", err)
	}
}

func TestBackup_InvalidBackupDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file where backup dir should be
	invalidDir := filepath.Join(tmpDir, "not-a-dir")
	if err := os.WriteFile(invalidDir, []byte("file"), 0644); err != nil {
		t.Fatal(err)
	}

	file1 := filepath.Join(tmpDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		BackupDir: filepath.Join(invalidDir, "backups"), // Can't create dir inside file
		Files:     []string{file1},
		Folders:   []string{},
	}

	password := "test-password-123"

	result, err := Backup(cfg, password)

	if err == nil {
		t.Error("Expected error when backup directory cannot be created")
	}

	if result != nil {
		t.Error("Expected nil result when backup fails")
	}
}

func TestBackupCleanupOnFailure(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")

	// Isolate TMPDIR so os.CreateTemp inside Backup() writes here,
	// not to the global /tmp where parallel tests leave temp files.
	isolatedTmp := filepath.Join(tmpDir, "tmp")
	if err := os.Mkdir(isolatedTmp, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", isolatedTmp)

	file1 := filepath.Join(tmpDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		BackupDir: backupDir,
		Files:     []string{file1},
		Folders:   []string{},
	}

	password := ""

	_, _ = Backup(cfg, password)

	entries, err := os.ReadDir(isolatedTmp)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), "dotkeeper-backup-") {
			t.Errorf("Found leftover temp file: %s", entry.Name())
		}
	}
}

func TestBackup_WithFolder(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")

	// Create folder with files
	folder := filepath.Join(tmpDir, "myfolder")
	if err := os.Mkdir(folder, 0755); err != nil {
		t.Fatal(err)
	}

	file1 := filepath.Join(folder, "file1.txt")
	file2 := filepath.Join(folder, "file2.txt")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		BackupDir: backupDir,
		Files:     []string{},
		Folders:   []string{folder},
	}

	password := "test-password-123"

	result, err := Backup(cfg, password)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	if result.FileCount != 2 {
		t.Errorf("Expected 2 files from folder, got %d", result.FileCount)
	}
}
