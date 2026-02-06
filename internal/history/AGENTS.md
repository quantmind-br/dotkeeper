# HISTORY PACKAGE KNOWLEDGE BASE

**Scope:** `internal/history/`

## OVERVIEW

Operation history tracking in JSONL format with file locking. 2 files, ~610 lines.

## STRUCTURE

```
history/
├── history.go       # Store, HistoryEntry, Append/Read
└── history_test.go  # Unit tests
```

## STORAGE FORMAT

Location: `$XDG_STATE_HOME/dotkeeper/history.jsonl`

**JSONL** (JSON Lines): One JSON object per line, newline-delimited

```json
{"timestamp":"2026-01-15T10:30:00Z","operation":"backup","status":"success","file_count":42,"total_size":1048576,"duration_ms":1234,"backup_path":"/path/to/backup","backup_name":"backup-2026-01-15-103000.tar.gz.enc"}
{"timestamp":"2026-01-15T11:00:00Z","operation":"restore","status":"error","backup_path":"/path/to/backup","error":"decryption failed"}
```

## HISTORY ENTRY

```go
type HistoryEntry struct {
    Timestamp  time.Time `json:"timestamp"`   // UTC
    Operation  string    `json:"operation"`   // "backup" | "restore"
    Status     string    `json:"status"`      // "success" | "error"
    FileCount  int       `json:"file_count"`
    TotalSize  int64     `json:"total_size"`  // Bytes
    DurationMs int64     `json:"duration_ms"`
    BackupPath string    `json:"backup_path,omitempty"`
    BackupName string    `json:"backup_name,omitempty"`
    Error      string    `json:"error,omitempty"`  // Only on failure
}
```

## API

```go
// Create store with default XDG path
store, err := history.NewStore()

// Create with custom path (for testing)
store := history.NewStoreWithPath("/custom/path/history.jsonl")

// Append entry with file locking
err := store.Append(entry)

// Read all entries (newest first)
entries, err := store.Read(0)  // 0 = no limit

// Read limited entries
entries, err := store.Read(100)  // Last 100 entries

// Read filtered by operation
entries, err := store.ReadByType("backup", 10)
```

## FILE LOCKING

Uses advisory file locking via `syscall.Flock`:

```go
// Exclusive lock for writes
syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
```

**Why?** Prevents corruption when multiple dotkeeper instances run concurrently (e.g., manual + systemd timer).

## ENTRY FACTORIES

```go
// From successful backup
entry := history.EntryFromBackupResult(result)

// From backup error
entry := history.EntryFromBackupError(err)

// From successful restore
entry := history.EntryFromRestoreResult(result, backupPath)

// From restore error
entry := history.EntryFromRestoreError(err, backupPath)
```

## READ BEHAVIOR

- Returns newest-first (reversed from file order)
- Corrupt lines are logged and skipped (doesn't fail entire read)
- Missing file returns empty slice (not error)

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Change entry format | `history.go:HistoryEntry` | JSON struct |
| Modify file locking | `history.go:Append()` | syscall.Flock |
| Add filter method | `history.go` | Like ReadByType |
| Change storage path | `history.go:NewStore()` | XDG_STATE_HOME |
| Export/import | New methods | JSONL is easy to parse |

## ANTI-PATTERNS

- **Never** write without file locking (corruption risk)
- **Never** store passwords in history entries
- **Never** parse JSONL manually - use the Store API
- **Don't** hold lock longer than necessary

## CONCURRENCY

File locking ensures:
- Multiple readers: OK
- One writer: OK (blocks readers/writers)
- Multiple writers: Sequential (via lock)

## TESTING

```go
// Always use temp path in tests
store := history.NewStoreWithPath(t.TempDir() + "/history.jsonl")

// Cleanup automatic with t.TempDir()
```
