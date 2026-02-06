package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/diogo/dotkeeper/internal/history"
)

func seedHistory(t *testing.T, stateHome string, entries []history.HistoryEntry) {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", stateHome)
	store, err := history.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	for _, e := range entries {
		if err := store.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
}

func TestHistoryCommand_JSONAndFilter(t *testing.T) {
	stateHome := t.TempDir()
	seedHistory(t, stateHome, []history.HistoryEntry{
		{Timestamp: time.Now().Add(-2 * time.Hour), Operation: "backup", Status: "success", FileCount: 2, TotalSize: 1024, DurationMs: 1200},
		{Timestamp: time.Now().Add(-1 * time.Hour), Operation: "restore", Status: "error", Error: "boom"},
	})

	var exit int
	stdout, stderr := captureStdoutStderr(t, func() {
		exit = HistoryCommand([]string{"--json"})
	})
	if exit != 0 {
		t.Fatalf("exit = %d, stderr=%s", exit, stderr)
	}

	var all []history.HistoryEntry
	if err := json.Unmarshal([]byte(stdout), &all); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}
	if len(all) != 2 {
		t.Fatalf("len(all) = %d", len(all))
	}

	stdout, stderr = captureStdoutStderr(t, func() {
		exit = HistoryCommand([]string{"--json", "--type", "backup"})
	})
	if exit != 0 {
		t.Fatalf("filtered exit = %d, stderr=%s", exit, stderr)
	}
	var filtered []history.HistoryEntry
	if err := json.Unmarshal([]byte(stdout), &filtered); err != nil {
		t.Fatalf("invalid filtered JSON output: %v\n%s", err, stdout)
	}
	if len(filtered) != 1 || filtered[0].Operation != "backup" {
		t.Fatalf("unexpected filtered entries: %#v", filtered)
	}
}

func TestHistoryCommand_EmptyPlain(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	var exit int
	stdout, stderr := captureStdoutStderr(t, func() {
		exit = HistoryCommand(nil)
	})
	if exit != 0 {
		t.Fatalf("exit = %d, stderr=%s", exit, stderr)
	}
	if !strings.Contains(stdout, "No operations recorded yet") {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestHistoryCommand_StoreError(t *testing.T) {
	base := t.TempDir()
	stateFile := filepath.Join(base, "state-file")
	if err := os.WriteFile(stateFile, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_STATE_HOME", stateFile)

	var exit int
	_, stderr := captureStdoutStderr(t, func() {
		exit = HistoryCommand(nil)
	})
	if exit != 1 {
		t.Fatalf("exit = %d", exit)
	}
	if !strings.Contains(stderr, "Error accessing history") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestHistoryCommand_TableFormat(t *testing.T) {
	stateHome := t.TempDir()
	// Seed with a successful backup entry
	seedHistory(t, stateHome, []history.HistoryEntry{
		{
			Timestamp:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			Operation:  "backup",
			Status:     "success",
			FileCount:  42,
			TotalSize:  1024 * 1024,
			DurationMs: 5000,
			BackupName: "backup-2025-01-01-120000",
		},
	})

	var exit int
	stdout, stderr := captureStdoutStderr(t, func() {
		exit = HistoryCommand(nil) // No --json flag, should trigger table format
	})

	if exit != 0 {
		t.Fatalf("exit = %d, stderr=%s", exit, stderr)
	}

	// Check for table format markers
	if !strings.Contains(stdout, "TIMESTAMP") {
		t.Error("expected table header with TIMESTAMP")
	}
	if !strings.Contains(stdout, "OPERATION") {
		t.Error("expected table header with OPERATION")
	}
	if !strings.Contains(stdout, "backup") {
		t.Error("expected 'backup' operation in output")
	}
	if !strings.Contains(stdout, "Total: 1 operation") {
		t.Error("expected total count")
	}
}

func TestHistoryCommand_TableFormatErrorEntry(t *testing.T) {
	stateHome := t.TempDir()
	// Seed with an error entry
	seedHistory(t, stateHome, []history.HistoryEntry{
		{
			Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			Operation: "backup",
			Status:    "error",
			Error:     "disk full",
			FileCount: 0,
			TotalSize: 0,
			DurationMs: 0,
		},
	})

	var exit int
	stdout, stderr := captureStdoutStderr(t, func() {
		exit = HistoryCommand(nil)
	})

	if exit != 0 {
		t.Fatalf("exit = %d, stderr=%s", exit, stderr)
	}

	// Check for error entry formatting (should have dashes for empty values)
	if !strings.Contains(stdout, "error") {
		t.Error("expected error status in output")
	}
	if !strings.Contains(stdout, "-") {
		t.Error("expected dash placeholder for error entry")
	}
}
