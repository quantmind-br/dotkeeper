# Draft: CI-005, CI-006, CI-008 - CLI Restore Refactoring

## Requirements (confirmed)

### CI-008: Refactor CLI Restore to Use internal/restore Package
- **Current state**: `internal/cli/restore.go` has ~100 lines of duplicated logic
  - `restoreBackup()` (lines 95-130): decrypt orchestration
  - `extractArchive()` (lines 133-200): naive tar extraction
- **Target**: Replace with `restore.Restore(backupPath, password, opts)` call
- **Benefits gained**:
  - Atomic writes (temp file + rename) - prevents corruption
  - Conflict resolution with `.bak` backups (currently only skip or overwrite)
  - Foundation for dry-run and diff features

### CI-005: Add Dry-Run Flag to CLI Restore
- **Depends on**: CI-008 (must use internal/restore first)
- **Implementation**: Add `--dry-run` flag → `RestoreOptions{DryRun: true}`
- **Behavior**: Shows what would be restored without making changes

### CI-006: Add Diff Preview to CLI Restore
- **Depends on**: CI-008 (must use internal/restore first)
- **Implementation**: Add `--diff` flag → `RestoreOptions{ShowDiff: true, DiffWriter: os.Stdout}`
- **Behavior**: Shows unified diffs before restore

## Technical Decisions

### Dependency Order
1. **CI-008 first** - refactor to use internal/restore
2. **CI-005 + CI-006** - can be done together after CI-008

### Available RestoreOptions
```go
type RestoreOptions struct {
    DryRun           bool                          // CI-005
    ShowDiff         bool                          // CI-006
    Force            bool                          // Already have --force
    TargetDir        string                        // Not exposing (keep simple)
    SelectedFiles    []string                      // Not exposing (keep simple)
    DiffWriter       io.Writer                     // CI-006: os.Stdout
    ProgressCallback func(file string, action string) // QUESTION: expose?
}
```

## Research Findings

### Current CLI Flow
```
RestoreCommand(args)
  ├─ Parse flags: --force, --password-file
  ├─ Load config
  ├─ Get backup name
  ├─ Construct backup path
  ├─ Verify backup exists
  ├─ Get password
  └─ restoreBackup() ← REPLACE THIS
       ├─ Read metadata
       ├─ Read encrypted backup
       ├─ Decrypt
       └─ extractArchive() ← DUPLICATED LOGIC
```

### internal/restore Package API
```go
func Restore(backupPath, password string, opts RestoreOptions) (*RestoreResult, error)

// RestoreResult provides rich information:
type RestoreResult struct {
    RestoredFiles []string          // Paths restored
    SkippedFiles  []string          // Paths skipped
    BackupFiles   []string          // .bak files created
    DiffResults   map[string]string // Diffs if ShowDiff=true
    TotalFiles    int
    FilesRestored int
    FilesSkipped  int
    FilesConflict int
}
```

### TUI Integration Pattern
```go
// TUI uses same pattern we'll use in CLI:
opts := restore.RestoreOptions{
    SelectedFiles: m.getSelectedFilePaths(),
}
result, err := restore.Restore(m.selectedBackup, m.password, opts)
```

## Decisions Made (User Confirmed)

1. **Progress callback**: NO - summary only, don't show per-file progress
2. **Exit codes**: Keep current semantics (0=success, 2=partial, 1=error)
3. **--diff without --dry-run**: YES - allow --diff alone (show diffs, then restore)
4. **Test strategy**: Add CLI unit tests in `internal/cli/restore_test.go`

### CRITICAL Behavior Change Decisions (from Metis review)

5. **Conflict behavior**: NEW BEHAVIOR - Create .bak backups then overwrite
   - OLD: CLI skipped existing files (no changes made)
   - NEW: CLI creates `.bak.TIMESTAMP` backup, then overwrites
   - Implementation: `RestoreOptions{Force: false}` → ActionBackup

6. **--force semantics**: Overwrite WITH backup (safer)
   - --force flag kept for backward compat, but now ALWAYS creates .bak
   - This is a behavior change: old --force meant "overwrite without backup"
   - New --force means "proceed with overwrite" (backup still created)
   - Implementation: `RestoreOptions{Force: false}` regardless of --force flag

## Scope Boundaries

### INCLUDE
- Refactor `restoreBackup()` and `extractArchive()` to use `restore.Restore()`
- Add `--dry-run` flag
- Add `--diff` flag
- Update help text
- Improved output (show .bak files created)

### EXCLUDE (keep simple)
- `--target-dir` flag (restore to different location)
- `--select-files` flag (restore specific files only)
- Interactive prompts for conflicts
- JSON output for restore command

## Test Strategy Decision
- **Infrastructure exists**: YES (Go tests in `internal/restore/*_test.go`)
- **Automated tests**: TBD - need user input
- **Framework**: Standard Go testing (`go test`)

## Test Coverage Analysis

### Existing Tests (969 lines in internal/restore/)
- `restore_test.go` (540 lines): Restore(), dry-run, conflicts, force, diffs, file selection
- `conflict_test.go` (239 lines): BackupExisting(), ResolveConflict(), HasConflict()
- `diff_test.go` (190 lines): GenerateDiff(), IsBinaryFile(), LCS algorithm

### E2E Tests
- `e2e/e2e_test.go`: TestBackupRestoreCycle(), TestBackupRestoreWithSubdirectories()
- `e2e/e2e_cli_test.go`: TestCLIRestoreWorkflow()

### Gap Identified
- **No CLI unit tests**: `internal/cli/restore.go` has no dedicated test file
- E2E tests cover basic CLI flow but not new flags

### Test Patterns in Project
- Fixture-based: `createTestBackup()` helper creates real backups
- Table-driven: Used in diff_test.go
- Temp directories: `t.TempDir()` everywhere
- No mocks: Real file I/O with real encryption
