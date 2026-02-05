package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/diogo/dotkeeper/internal/backup"
	"github.com/diogo/dotkeeper/internal/restore"
)

// HistoryEntry represents a single operation in the history log.
type HistoryEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Operation  string    `json:"operation"`
	Status     string    `json:"status"`
	FileCount  int       `json:"file_count"`
	TotalSize  int64     `json:"total_size"`
	DurationMs int64     `json:"duration_ms"`
	BackupPath string    `json:"backup_path,omitempty"`
	BackupName string    `json:"backup_name,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// Store manages reading and writing operation history in JSONL format.
type Store struct {
	path string
}

// NewStore creates a new Store using XDG_STATE_HOME with fallback to ~/.local/state.
func NewStore() (*Store, error) {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		stateHome = filepath.Join(home, ".local", "state")
	}

	dir := filepath.Join(stateHome, "dotkeeper")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	return &Store{
		path: filepath.Join(dir, "history.jsonl"),
	}, nil
}

// NewStoreWithPath creates a Store with a custom path, for testing.
func NewStoreWithPath(path string) *Store {
	return &Store{path: path}
}

// Append writes a HistoryEntry to the JSONL file with advisory file locking.
func (s *Store) Append(entry HistoryEntry) error {
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to lock history file: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal history entry: %w", err)
	}

	data = append(data, '\n')
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write history entry: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync history file: %w", err)
	}

	return nil
}

// Read returns history entries sorted newest-first. If limit > 0, at most limit entries
// are returned. If the file doesn't exist, an empty slice is returned.
func (s *Store) Read(limit int) ([]HistoryEntry, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoryEntry{}, nil
		}
		return nil, fmt.Errorf("failed to open history file: %w", err)
	}
	defer f.Close()

	var entries []HistoryEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry HistoryEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			log.Printf("warning: skipping corrupt history line: %s", err)
			continue
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	// Reverse to newest-first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

// ReadByType returns history entries filtered by operation type, sorted newest-first.
func (s *Store) ReadByType(opType string, limit int) ([]HistoryEntry, error) {
	all, err := s.Read(0)
	if err != nil {
		return nil, err
	}

	var filtered []HistoryEntry
	for _, entry := range all {
		if entry.Operation == opType {
			filtered = append(filtered, entry)
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

// EntryFromBackupResult creates a HistoryEntry from a successful backup result.
func EntryFromBackupResult(result *backup.BackupResult) HistoryEntry {
	return HistoryEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "backup",
		Status:     "success",
		FileCount:  result.FileCount,
		TotalSize:  result.TotalSize,
		DurationMs: result.Duration.Milliseconds(),
		BackupPath: result.BackupPath,
		BackupName: result.BackupName,
	}
}

// EntryFromBackupError creates a HistoryEntry from a failed backup.
func EntryFromBackupError(err error) HistoryEntry {
	return HistoryEntry{
		Timestamp: time.Now().UTC(),
		Operation: "backup",
		Status:    "error",
		Error:     err.Error(),
	}
}

// EntryFromRestoreResult creates a HistoryEntry from a successful restore result.
func EntryFromRestoreResult(result *restore.RestoreResult, backupPath string) HistoryEntry {
	return HistoryEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "restore",
		Status:     "success",
		FileCount:  result.FilesRestored,
		BackupPath: backupPath,
	}
}

// EntryFromRestoreError creates a HistoryEntry from a failed restore.
func EntryFromRestoreError(err error, backupPath string) HistoryEntry {
	return HistoryEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "restore",
		Status:     "error",
		BackupPath: backupPath,
		Error:      err.Error(),
	}
}
