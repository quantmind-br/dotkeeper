# RESTORE PACKAGE KNOWLEDGE BASE

**Scope:** `internal/restore/`

## OVERVIEW

Restore orchestration: decrypt → extract → conflict resolution → atomic write. 7 files, ~2,800 lines (largest package).

## STRUCTURE

```
restore/
├── restore.go        # Main orchestration (RestoreManager)
├── conflict.go       # Conflict detection & resolution
├── diff.go           # File diff preview
├── types.go          # Restore types & options
├── restore_test.go   # Unit tests (largest: 539 lines)
├── conflict_test.go
└── diff_test.go
```

## RESTORE FLOW

```
LoadMeta() → Decrypt() → Extract() → [ConflictCheck] → AtomicWrite()
```

1. **Load**: Read `.meta.json` for salt/KDF params
2. **Decrypt**: AES-256-GCM via `internal/crypto`
3. **Extract**: Parse tar.gz stream
4. **Conflict**: Check existing files, offer: skip / backup (.bak) / overwrite
5. **Write**: Temp file + rename for atomicity

## CONFLICT RESOLUTION

```go
type ConflictResolution int
const (
    Skip ConflictResolution = iota
    BackupOriginal
    Overwrite
)
```

UI shows diff preview before resolution.

## KEY TYPES

```go
type RestoreManager struct {
    config *config.Config
}

type RestoreOptions struct {
    BackupID    string
    Password    string
    Resolution  ConflictResolution
    DryRun      bool
}
```

## CONVENTIONS

- **Atomic writes**: `restoreFileAtomic()` - write to temp, then rename
- **Backup suffix**: `.bak` for existing files
- **Validation**: `ValidateBackup()` does trial decryption

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Change conflict behavior | `conflict.go` | Resolution strategies |
| Add diff features | `diff.go` | Currently basic line diff |
| Modify atomic write | `restore.go:restoreFileAtomic()` | Temp + rename pattern |
| Restore to different path | `restore.go` | Currently restores to original paths |

## ANTI-PATTERNS

- **Never** restore without atomic write (prevents corruption)
- **Never** delete existing files without backup option
- **Don't** skip conflict resolution in TUI (always prompt)
