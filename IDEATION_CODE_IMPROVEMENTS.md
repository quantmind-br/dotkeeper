# Code Improvements Ideation Report

## Executive Summary

This report analyzes the `dotkeeper` codebase to identify improvement opportunities that naturally emerge from existing patterns and architecture. The analysis discovered **6 major pattern categories** across TUI, CLI, backup/restore, configuration, and infrastructure layers, resulting in **15 concrete improvement ideas** across all effort levels.

**Key Findings:**
- The TUI has a well-established view model pattern with several incomplete views (LogsView is a placeholder)
- The CLI has an unwired `schedule` command and duplicated restore logic
- The `internal/restore` package has advanced features not exposed in CLI
- Notification and validation patterns exist but are underutilized
- Centralized styling exists but is inconsistently applied

## Existing Patterns Discovered

### Pattern 1: BubbleTea View Model Architecture
- **Location:** `internal/tui/views/*.go`, `internal/tui/model.go`
- **Description:** Each view implements `tea.Model` with Init/Update/View, shares config via constructor, uses async commands for I/O
- **Extension Potential:** New views can be added following exact same pattern; LogsView is a skeleton waiting for implementation

### Pattern 2: CLI Command Handler Pattern
- **Location:** `internal/cli/*.go`, `cmd/dotkeeper/main.go`
- **Description:** `func XxxCommand(args []string) int` functions with `flag.NewFlagSet`, returns exit codes
- **Extension Potential:** `schedule.go` exists but isn't wired; new commands follow trivial pattern

### Pattern 3: Backup/Restore Pipeline
- **Location:** `internal/backup/`, `internal/restore/`
- **Description:** Collect -> Archive -> Encrypt -> Save (backup) and Decrypt -> Extract -> Conflict -> Atomic (restore)
- **Extension Potential:** RestoreOptions supports dry-run, diff preview, file selection - CLI doesn't expose these

### Pattern 4: Centralized Styles
- **Location:** `internal/tui/styles.go`
- **Description:** `Styles` struct with Title, Subtitle, Normal, Selected, Help, Error, Success styles
- **Extension Potential:** Most views define inline styles instead of using this; consolidation opportunity

### Pattern 5: Path Validation
- **Location:** `internal/tui/views/helpers.go`
- **Description:** `ValidatePath`, `ValidateFilePath`, `ValidateFolderPath` with detailed result struct
- **Extension Potential:** Only used in TUI Settings; could be used in CLI and backup collector

### Pattern 6: Notification System
- **Location:** `internal/notify/notify.go`
- **Description:** `Send`, `SendSuccess`, `SendError` wrappers around `notify-send`
- **Extension Potential:** Currently unused in CLI backup; pattern ready for integration

---

## Improvement Opportunities

### Trivial Effort (1-2 hours)

#### CI-001: Wire Schedule Command to CLI

**Builds Upon:** CLI command pattern, existing `schedule.go` implementation
**Affected Files:**
- `cmd/dotkeeper/main.go`
- `internal/cli/schedule.go`

**Description:**
The `schedule` command (enable/disable/status for systemd timer) is fully implemented in `internal/cli/schedule.go` but not wired in `main.go`.

**Rationale:**
The code reveals a complete implementation sitting unused. The pattern for wiring is identical to other commands (single case in switch statement).

**Existing Patterns:**
- `backup`, `restore`, `list`, `config` commands follow same wiring pattern

**Implementation Approach:**
Add `case "schedule": exitCode = cli.ScheduleCommand(args)` to the switch in `main.go`. The `ScheduleCommand` function likely already exists or can be created from the existing `EnableSchedule`/`DisableSchedule`/`StatusSchedule` functions.

---

#### CI-002: Add Notifications to CLI Backup

**Builds Upon:** Notification pattern in `internal/notify/`
**Affected Files:**
- `internal/cli/backup.go`

**Description:**
Send desktop notifications on backup success/failure when running from CLI.

**Rationale:**
The `notify.SendSuccess` and `notify.SendError` functions exist but are never called. The pattern is ready; just needs invocation.

**Existing Patterns:**
- `internal/notify/notify.go` already has `SendSuccess(backupName, duration)` and `SendError(err)`

**Implementation Approach:**
After successful backup, call `notify.SendSuccess(result.BackupName, result.Duration)`. On failure, call `notify.SendError(err)`. Add a `--notify` flag to control this (default on if config.Notifications is true).

---

#### CI-003: Use Centralized Styles in All TUI Views

**Builds Upon:** `internal/tui/styles.go` pattern
**Affected Files:**
- `internal/tui/views/dashboard.go`
- `internal/tui/views/backuplist.go`
- `internal/tui/views/restore.go`
- `internal/tui/views/settings.go`
- `internal/tui/views/logs.go`

**Description:**
Replace inline `lipgloss.NewStyle()` calls with the centralized `Styles` struct.

**Rationale:**
The code reveals a `DefaultStyles()` function that's defined but each view creates its own inline styles. This is inconsistent and harder to theme.

**Existing Patterns:**
- `internal/tui/styles.go` defines Title, Subtitle, Normal, Selected, Help, Error, Success

**Implementation Approach:**
Import and use `styles := tui.DefaultStyles()` in each view, replacing inline style definitions.

---

### Small Effort (Half day)

#### CI-004: Add JSON Output to Backup Command

**Builds Upon:** JSON output pattern in `list.go`
**Affected Files:**
- `internal/cli/backup.go`

**Description:**
Add `--json` flag to backup command for machine-readable output.

**Rationale:**
The `list` command already has `--json` with `json.MarshalIndent`. Backup returns a `BackupResult` struct that's already JSON-serializable.

**Existing Patterns:**
- `internal/cli/list.go` lines 29, 81-87 show exact pattern

**Implementation Approach:**
Add `jsonOutput := fs.Bool("json", false, "Output in JSON format")`, then marshal `result` to JSON if flag is set.

---

#### CI-005: Add Dry-Run Flag to CLI Restore

**Builds Upon:** RestoreOptions.DryRun in `internal/restore/`
**Affected Files:**
- `internal/cli/restore.go`

**Description:**
Add `--dry-run` flag to preview what would be restored without making changes.

**Rationale:**
The `RestoreOptions` struct already has `DryRun` and `ShowDiff` fields. The TUI uses them (via `PreviewRestore`), but CLI doesn't.

**Existing Patterns:**
- `internal/restore/restore.go` lines 63-73 handle DryRun mode
- `internal/restore/restore.go` line 278-281 `PreviewRestore` function

**Implementation Approach:**
Add `dryRun := fs.Bool("dry-run", false, "Preview restore without making changes")`. Refactor to use `restore.Restore()` with `RestoreOptions{DryRun: *dryRun}` instead of the inline extraction logic.

---

#### CI-006: Add Diff Preview to CLI Restore

**Builds Upon:** Diff generation in `internal/restore/diff.go`
**Affected Files:**
- `internal/cli/restore.go`

**Description:**
Add `--diff` flag to show differences between backup and current files.

**Rationale:**
The TUI restore view has full diff preview (phase 4). The pattern exists via `restore.GetFileDiff()` and `RestoreOptions.ShowDiff`.

**Existing Patterns:**
- `internal/restore/diff.go` has `GenerateDiff(content, targetPath)`
- `internal/restore/restore.go` lines 46-59 handle ShowDiff option

**Implementation Approach:**
Add `showDiff := fs.Bool("diff", false, "Show file differences before restore")`. Set `RestoreOptions{ShowDiff: *showDiff, DiffWriter: os.Stdout}`.

---

### Medium Effort (1-3 days)

#### CI-007: Implement LogsView with Operation History

**Builds Upon:** TUI view model pattern, backup/restore result structures
**Affected Files:**
- `internal/tui/views/logs.go`
- New file: `internal/logs/history.go`

**Description:**
Implement the placeholder LogsView to show backup/restore operation history.

**Rationale:**
The `logs.go` view exists as a skeleton with "implementation pending" comment. The pattern for views is well-established. Backup/restore operations return result structs with all needed metadata.

**Existing Patterns:**
- `BackupResult` struct has FileCount, TotalSize, Duration, Checksum
- `RestoreResult` struct has RestoredFiles, SkippedFiles, BackupFiles
- List-based views like `BackupListModel` show how to display items

**Implementation Approach:**
1. Create `internal/logs/history.go` with a JSON file to persist operation history
2. Add logging calls after backup/restore operations
3. Implement LogsView using `list.Model` to display entries
4. Add filtering by operation type (backup/restore), date range

---

#### CI-008: Refactor CLI Restore to Use internal/restore Package

**Builds Upon:** Advanced restore features in `internal/restore/`
**Affected Files:**
- `internal/cli/restore.go`

**Description:**
Replace the naive `extractArchive` implementation with the full-featured `restore.Restore()` function.

**Rationale:**
There's significant code duplication. `internal/cli/restore.go` reimplements tar extraction (lines 133-200) while `internal/restore/restore.go` has atomic writes, conflict resolution, diff generation, and dry-run support.

**Existing Patterns:**
- `internal/restore/restore.go` is production-ready with all edge cases handled
- TUI's `RestoreModel` already uses this pattern correctly

**Implementation Approach:**
Replace `restoreBackup()` function body with:
```go
opts := restore.RestoreOptions{
    ConflictAction: restore.ActionOverwrite, // if force
    ShowDiff: false,
}
result, err := restore.Restore(backupPath, password, opts)
```

---

#### CI-009: Add File Exclusion Patterns to Collector

**Builds Upon:** CollectFiles pattern in `internal/backup/collector.go`
**Affected Files:**
- `internal/backup/collector.go`
- `internal/config/config.go`

**Description:**
Add support for glob-based exclusion patterns (e.g., `*.log`, `node_modules/`).

**Rationale:**
The collector recursively collects all files but has no exclusion mechanism. Config already has Files and Folders slices; adding Excludes follows the same pattern.

**Existing Patterns:**
- `CollectFiles` iterates paths and uses `filepath.Glob` semantics
- Config struct uses yaml tags for serialization

**Implementation Approach:**
1. Add `Excludes []string` to Config struct
2. In `collectPath`, check if path matches any exclude pattern using `filepath.Match`
3. Skip matching paths with a log warning

---

#### CI-010: Add Interactive Password Prompt for CLI

**Builds Upon:** Password handling in `getPassword()`
**Affected Files:**
- `internal/cli/backup.go`

**Description:**
Add fallback interactive password prompt when no password source is available.

**Rationale:**
The help text mentions "non-interactive mode" for env vars, implying interactive mode should exist. Currently, if file/env/keyring all fail, the command fails.

**Existing Patterns:**
- `getPassword()` already has priority chain (file -> env -> keyring)
- TUI uses `textinput` with `EchoPassword` mode

**Implementation Approach:**
Add a fourth priority: use `golang.org/x/term.ReadPassword()` to prompt for password. Detect if stdin is a TTY first; if not, return the existing error.

---

### Large Effort (3-7 days)

#### CI-011: Add Help Overlay System to TUI

**Builds Upon:** KeyMap pattern in `internal/tui/update.go`
**Affected Files:**
- `internal/tui/update.go`
- `internal/tui/model.go`
- `internal/tui/views/*.go`

**Description:**
Implement a consistent help overlay triggered by `?` key showing context-aware shortcuts.

**Rationale:**
The `?` key binding exists in `DefaultKeyMap()` but doesn't do anything. Each view has its own shortcuts documented in the View() help text, but there's no overlay.

**Existing Patterns:**
- `key.Binding` with `WithHelp()` already documents shortcuts
- Each view has help text in its View() function

**Implementation Approach:**
1. Add `showingHelp bool` to main Model
2. Create `HelpModel` view that renders all keybindings
3. Toggle on `?` press, dismiss on any key
4. Collect shortcuts from each view's KeyMap

---

#### CI-012: Wire FileBrowser Selection to Config

**Builds Upon:** FileBrowser view pattern, Settings config editing
**Affected Files:**
- `internal/tui/views/filebrowser.go`
- `internal/tui/model.go`
- `internal/tui/update.go`

**Description:**
Complete the FileBrowser view to allow selecting files/folders that get added to the config.

**Rationale:**
The FileBrowser view exists and uses `filepicker` component, but selected paths aren't wired back to `config.Files/Folders`. The Settings view already shows how to modify config.

**Existing Patterns:**
- Settings view modifies `config.Files` and `config.Folders` directly
- `RefreshBackupListMsg` shows cross-view communication pattern

**Implementation Approach:**
1. Add message type `FileSelectedMsg{path string, isDir bool}`
2. Handle in main Update to add to appropriate config slice
3. Add confirmation UI before adding
4. Save config after selection

---

### Complex Effort (1-2 weeks)

#### CI-013: Streaming Encryption for Large Backups

**Builds Upon:** Crypto patterns in `internal/crypto/`, backup pipeline
**Affected Files:**
- `internal/crypto/aes.go`
- `internal/backup/backup.go`

**Description:**
Replace in-memory encryption with streaming `io.Reader`/`io.Writer` to handle large backup archives without memory pressure.

**Rationale:**
AGENTS.md explicitly mentions this as a TODO: "Large files: Currently loads in memory; streaming TODO". The current `os.ReadFile(tempFile.Name())` loads entire archive into RAM.

**Existing Patterns:**
- Crypto interface is already byte-based; needs Reader/Writer variants
- Archive creation already writes to a temp file

**Implementation Approach:**
1. Add `EncryptStream(src io.Reader, dst io.Writer, key, salt []byte)` to crypto package
2. Modify backup to pipe tar.gz writer -> crypto writer -> file writer
3. Preserve existing in-memory API for backward compatibility
4. Add streaming decrypt for restore

---

#### CI-014: Git Auto-Sync After Backup

**Builds Upon:** Git operations in `internal/git/`, backup flow
**Affected Files:**
- `internal/git/operations.go`
- `internal/backup/backup.go`
- `internal/config/config.go`

**Description:**
Automatically commit and push backups to the configured git remote after successful backup.

**Rationale:**
The git package has `Add`, `Commit`, `Push` operations ready. Config has `GitRemote` field. The pieces exist but aren't connected.

**Existing Patterns:**
- `internal/git/operations.go` has `AddAll()`, `Commit(message)`, `Push()`
- `internal/git/repo.go` has `InitOrOpen()`

**Implementation Approach:**
1. Add `AutoSync bool` to Config
2. After backup completes, if AutoSync:
   - InitOrOpen the backup directory
   - AddAll, Commit with timestamp message, Push
3. Add `--no-sync` flag to CLI backup
4. Handle push failures gracefully (log warning, don't fail backup)

---

#### CI-015: Notifier Interface with Multiple Backends

**Builds Upon:** Notification pattern in `internal/notify/`
**Affected Files:**
- `internal/notify/notify.go`
- New file: `internal/notify/webhook.go`
- `internal/config/config.go`

**Description:**
Replace hardcoded `notify-send` with a Notifier interface supporting multiple backends (notify-send, webhook, Discord, Slack).

**Rationale:**
The current implementation only works on Linux with `notify-send`. Scheduled backups (headless) can't receive notifications. The pattern of Send/SendSuccess/SendError is solid; just needs abstraction.

**Existing Patterns:**
- `Send(title, message)` interface is already clean
- Config already has `Notifications bool`

**Implementation Approach:**
1. Define `type Notifier interface { Send(title, message) error }`
2. Create `NotifySend`, `WebhookNotifier` implementations
3. Add `NotifyType string` and `NotifyURL string` to Config
4. Factory function to create appropriate notifier based on config

---

## Summary

| Effort Level | Count |
|--------------|-------|
| Trivial | 3 |
| Small | 3 |
| Medium | 4 |
| Large | 2 |
| Complex | 3 |

**Total Ideas:** 15
**Patterns Discovered:** 6

### Priority Recommendations

1. **Quick Wins (CI-001, CI-002, CI-003):** Wire schedule command, add notifications, consolidate styles
2. **High Value (CI-008, CI-005, CI-006):** Refactor CLI restore to use advanced features
3. **User-Visible (CI-007, CI-011):** Implement LogsView and Help overlay
4. **Infrastructure (CI-013, CI-014):** Streaming encryption and git auto-sync
