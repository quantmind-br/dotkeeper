package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/diogo/dotkeeper/internal/backup"
	"github.com/diogo/dotkeeper/internal/restore"
)

func TestAppendWritesValidJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	store := NewStoreWithPath(path)

	entry := HistoryEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "backup",
		Status:     "success",
		FileCount:  10,
		TotalSize:  1024,
		DurationMs: 500,
		BackupPath: "/tmp/backup.tar.gz.enc",
		BackupName: "backup-2025-01-01-120000",
	}

	if err := store.Append(entry); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	var readBack HistoryEntry
	if err := json.Unmarshal(data[:len(data)-1], &readBack); err != nil {
		t.Fatalf("Failed to unmarshal JSONL line: %v", err)
	}

	if readBack.Operation != "backup" {
		t.Errorf("expected operation 'backup', got %q", readBack.Operation)
	}
	if readBack.FileCount != 10 {
		t.Errorf("expected file_count 10, got %d", readBack.FileCount)
	}
}

func TestReadReturnsNewestFirst(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	store := NewStoreWithPath(path)

	for i := 0; i < 5; i++ {
		entry := HistoryEntry{
			Timestamp: time.Date(2025, 1, 1+i, 0, 0, 0, 0, time.UTC),
			Operation: "backup",
			Status:    "success",
			FileCount: i,
		}
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append %d failed: %v", i, err)
		}
	}

	entries, err := store.Read(0)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}

	// Newest first: FileCount 4 should be first
	if entries[0].FileCount != 4 {
		t.Errorf("expected first entry FileCount=4, got %d", entries[0].FileCount)
	}
	if entries[4].FileCount != 0 {
		t.Errorf("expected last entry FileCount=0, got %d", entries[4].FileCount)
	}
}

func TestReadWithLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	store := NewStoreWithPath(path)

	for i := 0; i < 10; i++ {
		entry := HistoryEntry{
			Timestamp: time.Date(2025, 1, 1+i, 0, 0, 0, 0, time.UTC),
			Operation: "backup",
			Status:    "success",
			FileCount: i,
		}
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append %d failed: %v", i, err)
		}
	}

	entries, err := store.Read(3)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Should be the 3 newest (FileCount 9, 8, 7)
	if entries[0].FileCount != 9 {
		t.Errorf("expected first entry FileCount=9, got %d", entries[0].FileCount)
	}
}

func TestReadByTypeFiltering(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	store := NewStoreWithPath(path)

	ops := []string{"backup", "restore", "backup", "restore", "backup"}
	for i, op := range ops {
		entry := HistoryEntry{
			Timestamp: time.Date(2025, 1, 1+i, 0, 0, 0, 0, time.UTC),
			Operation: op,
			Status:    "success",
			FileCount: i,
		}
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append %d failed: %v", i, err)
		}
	}

	backups, err := store.ReadByType("backup", 0)
	if err != nil {
		t.Fatalf("ReadByType failed: %v", err)
	}
	if len(backups) != 3 {
		t.Fatalf("expected 3 backup entries, got %d", len(backups))
	}

	restores, err := store.ReadByType("restore", 1)
	if err != nil {
		t.Fatalf("ReadByType failed: %v", err)
	}
	if len(restores) != 1 {
		t.Fatalf("expected 1 restore entry (limited), got %d", len(restores))
	}
}

func TestCorruptLineResilience(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	store := NewStoreWithPath(path)

	// Write first valid entry
	entry1 := HistoryEntry{
		Timestamp: time.Now().UTC(),
		Operation: "backup",
		Status:    "success",
		FileCount: 1,
	}
	if err := store.Append(entry1); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Manually write corrupt line
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	f.Write([]byte("not-json\n"))
	f.Close()

	// Write second valid entry
	entry2 := HistoryEntry{
		Timestamp: time.Now().UTC(),
		Operation: "restore",
		Status:    "success",
		FileCount: 2,
	}
	if err := store.Append(entry2); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	entries, err := store.Read(0)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (skipping corrupt), got %d", len(entries))
	}
}

func TestEmptyNonexistentFileReturnsEmptySlice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.jsonl")
	store := NewStoreWithPath(path)

	entries, err := store.Read(0)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}

	// Also test empty file
	os.WriteFile(path, []byte(""), 0600)
	entries, err = store.Read(0)
	if err != nil {
		t.Fatalf("Read empty file failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries from empty file, got %d", len(entries))
	}
}

func TestFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	store := NewStoreWithPath(path)

	entry := HistoryEntry{
		Timestamp: time.Now().UTC(),
		Operation: "backup",
		Status:    "success",
	}
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected permissions 0600, got %04o", perm)
	}
}

func TestConcurrentAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	store := NewStoreWithPath(path)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			entry := HistoryEntry{
				Timestamp: time.Now().UTC(),
				Operation: "backup",
				Status:    "success",
				FileCount: n,
			}
			if err := store.Append(entry); err != nil {
				t.Errorf("Concurrent Append %d failed: %v", n, err)
			}
		}(i)
	}
	wg.Wait()

	entries, err := store.Read(0)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(entries) != goroutines {
		t.Errorf("expected %d entries, got %d", goroutines, len(entries))
	}

	// Verify each line is valid JSON by checking we got all entries
	seen := make(map[int]bool)
	for _, e := range entries {
		seen[e.FileCount] = true
	}
	if len(seen) != goroutines {
		t.Errorf("expected %d unique FileCount values, got %d", goroutines, len(seen))
	}
}

func TestNewStoreCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "nested", "deep", "dotkeeper")
	path := filepath.Join(stateDir, "history.jsonl")

	// Ensure parent doesn't exist
	if _, err := os.Stat(filepath.Join(dir, "nested")); !os.IsNotExist(err) {
		t.Fatal("expected nested dir to not exist yet")
	}

	// Use NewStore with XDG_STATE_HOME override
	t.Setenv("XDG_STATE_HOME", filepath.Join(dir, "nested", "deep"))

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(stateDir)
	if err != nil {
		t.Fatalf("Directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory, got file")
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("expected directory permissions 0700, got %04o", info.Mode().Perm())
	}

	// Verify the store path is correct
	entry := HistoryEntry{Timestamp: time.Now().UTC(), Operation: "backup", Status: "success"}
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append to new store failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("History file not at expected path %s: %v", path, err)
	}
}

func TestEntryFromBackupResult(t *testing.T) {
	result := &backup.BackupResult{
		BackupPath: "/backups/backup-2025-01-01.tar.gz.enc",
		BackupName: "backup-2025-01-01-120000",
		FileCount:  42,
		TotalSize:  123456,
		Duration:   2500 * time.Millisecond,
	}

	entry := EntryFromBackupResult(result)

	if entry.Operation != "backup" {
		t.Errorf("expected operation 'backup', got %q", entry.Operation)
	}
	if entry.Status != "success" {
		t.Errorf("expected status 'success', got %q", entry.Status)
	}
	if entry.FileCount != 42 {
		t.Errorf("expected FileCount 42, got %d", entry.FileCount)
	}
	if entry.DurationMs != 2500 {
		t.Errorf("expected DurationMs 2500, got %d", entry.DurationMs)
	}
	if entry.BackupPath != result.BackupPath {
		t.Errorf("expected BackupPath %q, got %q", result.BackupPath, entry.BackupPath)
	}
}

func TestEntryFromBackupError(t *testing.T) {
	err := fmt.Errorf("encryption failed: key too short")
	entry := EntryFromBackupError(err)

	if entry.Operation != "backup" {
		t.Errorf("expected operation 'backup', got %q", entry.Operation)
	}
	if entry.Status != "error" {
		t.Errorf("expected status 'error', got %q", entry.Status)
	}
	if entry.Error != "encryption failed: key too short" {
		t.Errorf("unexpected error message: %q", entry.Error)
	}
}

func TestEntryFromRestoreResult(t *testing.T) {
	result := &restore.RestoreResult{
		FilesRestored: 15,
		TotalFiles:    20,
	}

	entry := EntryFromRestoreResult(result, "/backups/backup.tar.gz.enc")

	if entry.Operation != "restore" {
		t.Errorf("expected operation 'restore', got %q", entry.Operation)
	}
	if entry.Status != "success" {
		t.Errorf("expected status 'success', got %q", entry.Status)
	}
	if entry.FileCount != 15 {
		t.Errorf("expected FileCount 15, got %d", entry.FileCount)
	}
	if entry.BackupPath != "/backups/backup.tar.gz.enc" {
		t.Errorf("unexpected BackupPath: %q", entry.BackupPath)
	}
}

func TestEntryFromRestoreError(t *testing.T) {
	err := fmt.Errorf("decryption failed: wrong password")
	entry := EntryFromRestoreError(err, "/backups/backup.tar.gz.enc")

	if entry.Operation != "restore" {
		t.Errorf("expected operation 'restore', got %q", entry.Operation)
	}
	if entry.Status != "error" {
		t.Errorf("expected status 'error', got %q", entry.Status)
	}
	if entry.BackupPath != "/backups/backup.tar.gz.enc" {
		t.Errorf("unexpected BackupPath: %q", entry.BackupPath)
	}
	if entry.Error != "decryption failed: wrong password" {
		t.Errorf("unexpected error message: %q", entry.Error)
	}
}
