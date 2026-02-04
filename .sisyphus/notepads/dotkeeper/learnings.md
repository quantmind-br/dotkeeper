# Learnings - dotkeeper

## File Collection and Archive Creation (Task 4)

### Implementation Approach
- Used TDD workflow: write failing tests first, implement minimum code to pass, refactor
- Separated concerns: collector.go for file gathering, archive.go for tar.gz creation
- Streaming approach: use io.Copy to avoid loading entire files in memory

### Symlink Handling
- Follow symlinks and copy content (don't preserve as links)
- Detect circular symlinks by tracking visited real paths
- Implement max depth check (20 levels) to prevent infinite loops
- Use countSymlinkDepth() to count actual symlink chain depth
- Key insight: filepath.EvalSymlinks() resolves entire chain at once, so need manual depth counting

### Error Handling Strategy
- Skip unreadable files with warning (don't fail entire operation)
- Skip special files (sockets, FIFOs, devices)
- Log warnings but continue processing other files
- Return errors only for critical failures

### Archive Format
- Use tar.gz format (tar + gzip compression)
- Preserve file permissions (mode bits)
- Preserve modification timestamps
- Stream files directly to archive (efficient for large files)

### Test Coverage
- Basic file collection and archiving
- Symlink following and circular detection
- Max depth enforcement
- Unreadable file handling
- Directory traversal
- Permission preservation
- Timestamp preservation
- Large file streaming (10MB test)
- Empty file list handling

## Backup Orchestration Module (Task 5)

**Date:** 2026-02-04

**Implementation:**
- Created `internal/backup/backup.go` with `Backup()` function that orchestrates the full workflow
- Workflow: collect files → create archive → encrypt → save
- Generates timestamped backup names: `backup-YYYY-MM-DD-HHMMSS.tar.gz.enc`
- Calculates SHA-256 checksum of plaintext archive before encryption
- Stores metadata in separate `.meta.json` file alongside encrypted backup
- Sets backup file permissions to 0600 (owner read/write only) for security
- Properly cleans up temp files using `defer os.Remove()`

**TDD Approach:**
- Wrote comprehensive tests first in `backup_test.go`
- Tests cover: basic backup, no files error, invalid backup dir, cleanup on failure, folder backup
- All tests pass including `TestBackupCleanupOnFailure` which verifies no temp files left behind

**Key Design Decisions:**
1. **Temp file cleanup:** Used `defer os.Remove(tempFile.Name())` immediately after creation to ensure cleanup even on errors
2. **Metadata separation:** Store metadata in separate JSON file for easy inspection without decryption
3. **Checksum calculation:** Calculate on plaintext before encryption for integrity verification
4. **File permissions:** Backup files get 0600, metadata gets 0644 (readable for inspection)
5. **Error handling:** Return descriptive errors at each step with context

**Integration:**
- Uses `CollectFiles()` from collector.go to gather files from both Files and Folders config
- Uses `CreateArchive()` from archive.go to create tar.gz archive
- Uses `crypto.GenerateSalt()`, `crypto.DeriveKey()`, `crypto.Encrypt()` for encryption
- Uses `crypto.EncryptionMetadata` struct for metadata format

**Test Results:**
```
=== RUN   TestBackup
--- PASS: TestBackup (0.06s)
=== RUN   TestBackup_NoFiles
--- PASS: TestBackup_NoFiles (0.00s)
=== RUN   TestBackup_InvalidBackupDir
--- PASS: TestBackup_InvalidBackupDir (0.00s)
=== RUN   TestBackupCleanupOnFailure
--- PASS: TestBackupCleanupOnFailure (0.03s)
=== RUN   TestBackup_WithFolder
--- PASS: TestBackup_WithFolder (0.02s)
PASS
ok  	github.com/diogo/dotkeeper/internal/backup	0.109s
```

**Commit:** `feat(backup): add backup orchestration module` (91e47da)

## Git Integration Module (2026-02-04)

### Implementation Approach
- Used TDD: wrote tests first, then implementation
- go-git library provides pure Go git operations (no shell commands)
- Split into three files: repo.go (init/open), operations.go (add/commit/push), status.go (status checking)

### Key Patterns
- Repository struct wraps go-git's Repository with path tracking
- InitOrOpen pattern: try to open existing, fallback to init
- Status detection requires checking Worktree field for untracked files separately
- go-git's Add() returns (hash, error) not just error

### Testing Strategy
- Use t.TempDir() for isolated test repositories
- Each test creates fresh repo to avoid state pollution
- Test both success and error paths (e.g., OpenNonExistent)
- Verify status after operations (e.g., IsClean after commit)

### Dependencies
- github.com/go-git/go-git/v5 - main git library
- github.com/go-git/go-git/v5/config - for remote configuration
- github.com/go-git/go-git/v5/plumbing/object - for commit signatures


## TUI Architecture Pattern
- Adopted the "Bubbletea" architecture for the TUI.
- Split the model into separate files for better maintainability:
    - **model.go**: Defines the data structure and state.
    - **update.go**: Handles event loops and state transitions.
    - **view.go**: Pure function rendering the UI based on state.
    - **styles.go**: Centralized styling using Lipgloss.
- Using `ViewState` enum to manage top-level navigation (Dashboard, FileBrowser, etc.).

## Project Completion Summary (2026-02-04)

**All 19 Tasks Completed Successfully!**

### Final Statistics
- **Total Tasks**: 19/19 (100%)
- **Total Test Files**: 20+ test files
- **Total Tests**: 100+ tests (all passing)
- **Binary Size**: 7.4MB
- **Build Status**: ✅ SUCCESS
- **Test Status**: ✅ ALL PASSING

### Completed Waves
1. **Wave 1 (Foundation)**: Tasks 1-4
   - Project scaffolding, Config, Crypto, Backup collection
   
2. **Wave 2 (Core Logic)**: Tasks 5-7
   - Backup orchestration, Git integration, Keyring
   
3. **Wave 3 (TUI)**: Tasks 8-14
   - TUI base + 6 screens (Dashboard, File Browser, Backup List, Restore, Settings, Logs)
   
4. **Wave 4 (Integration)**: Tasks 15-19
   - CLI commands, Systemd timer, Notifications, Restore module, E2E tests

### Key Features Delivered
- AES-256-GCM encryption with Argon2id KDF
- File browser with bubbles/filepicker
- Git integration with go-git
- Keyring integration for secure password storage
- 6-screen TUI with Bubbletea
- CLI commands: backup, restore, list, config, schedule
- Systemd timer for scheduled backups
- Desktop notifications via notify-send
- Restore with diff preview and conflict resolution
- 17 end-to-end integration tests

### Architecture Highlights
- Clean separation of concerns (internal/ packages)
- TDD approach throughout
- Streaming operations for memory efficiency
- Comprehensive error handling
- Atomic file operations
- Secure defaults (0600 permissions, no password storage)

### Verification Commands
```bash
# Build
go build -o dotkeeper ./cmd/dotkeeper

# Run tests
go test ./...

# Run TUI
./dotkeeper

# Run CLI commands
./dotkeeper backup
./dotkeeper list
./dotkeeper restore <backup-name>
./dotkeeper config get backup_dir
```

**Status**: ✅ PRODUCTION READY
