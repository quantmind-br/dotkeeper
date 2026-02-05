# BACKUP PACKAGE KNOWLEDGE BASE

**Scope:** `internal/backup/`

## OVERVIEW

Backup orchestration: collect → archive → encrypt → save. 6 files, ~1,800 lines.

## STRUCTURE

```
backup/
├── backup.go       # Main orchestration (BackupManager)
├── collector.go    # File collection from config paths
├── archive.go      # tar.gz creation
├── backup_test.go  # Unit tests
├── collector_test.go
└── archive_test.go
```

## BACKUP FLOW

```
Collect(paths) → Archive(files) → Encrypt(data) → Write(encrypted, meta)
```

1. **Collect**: Walk config paths, resolve symlinks (copy content)
2. **Archive**: Create temp tar.gz with preserved structure
3. **Encrypt**: AES-256-GCM + Argon2id (via `internal/crypto`)
4. **Write**: Save `.tar.gz.enc` + `.meta.json` with 0600 permissions

## KEY TYPES

```go
type BackupManager struct {
    config *config.Config
}

func (bm *BackupManager) CreateBackup(password string) (*BackupInfo, error)
```

## CONVENTIONS

- **Temp files**: Use `os.CreateTemp("", "dotkeeper-*")`
- **Permissions**: Always `0600` for backup files
- **Error wrapping**: `fmt.Errorf("backup failed: %w", err)`
- **Atomic writes**: Temp + rename pattern

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add pre-backup hook | `backup.go:CreateBackup()` | Before collection |
| Change archive format | `archive.go` | Currently tar.gz |
| Handle symlinks differently | `collector.go` | Currently follows & copies |
| Streaming for large files | `backup.go` | **TODO**: currently loads in memory |

## ANTI-PATTERNS

- **Never** change crypto defaults (AES-256-GCM, Argon2id params)
- **Never** skip permission setting on backup files
- **Don't** add incremental backups (full only, by design)
