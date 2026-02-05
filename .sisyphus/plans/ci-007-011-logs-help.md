# CI-007 + CI-011: LogsView with Operation History & Help Overlay System

## TL;DR

> **Quick Summary**: Implement two independent TUI features: (1) a fully functional LogsView that displays persistent operation history stored as JSONL, with history logging hooks in both CLI and TUI, and a `dotkeeper history` CLI command; (2) a context-aware help overlay triggered by `?` key, rendered as a centered modal over the current view.
>
> **Deliverables**:
> - `internal/history/` package — JSONL-based operation history storage
> - `internal/history/history_test.go` — Full unit test coverage
> - Fully implemented `internal/tui/views/logs.go` LogsView with `list.Model`
> - History logging hooks in CLI backup/restore and TUI backup/restore flows
> - `dotkeeper history [--json] [--type backup|restore]` CLI command
> - Help overlay system in `internal/tui/` with per-view keybinding display
> - Updated tests for all modified files
>
> **Estimated Effort**: Medium (3-5 days)
> **Parallel Execution**: YES — 2 waves
> **Critical Path**: Task 1 → Task 2 → Task 3 → Task 4 (CI-007), Task 5 → Task 6 (CI-011), Task 7 (integration)

---

## Context

### Original Request
Implement CI-007 (LogsView with Operation History) and CI-011 (Help Overlay System) from `IDEATION_CODE_IMPROVEMENTS.md`.

### Interview Summary
**Key Discussions**:
- User confirmed both features in a single plan
- Both features are independent and can be parallelized
- History stored in XDG_STATE_HOME as JSONL

**Research Findings**:
- **BubbleTea help patterns**: Charm ecosystem uses `bubbles/help` for inline bottom-bar help; for overlay, use `lipgloss.Place()` with custom rendering (soft-serve demonstrates context-aware keybinding composition)
- **JSONL storage**: Append-only, one JSON object per line, skip corrupt lines during read. XDG_STATE_HOME (`~/.local/state`) is the correct XDG path for operation history per spec
- **Codebase conventions**: No `adrg/xdg` library used — project rolls its own XDG via `os.Getenv()` with fallback. Views use raw `msg.String()` comparisons, not `key.Binding` internally
- **Existing wiring**: LogsView skeleton already exists and is wired into model.go, update.go, view.go. `?` key defined in KeyMap but unhandled

### Metis Review
**Identified Gaps** (all addressed in this plan):
- **Hook points enumerated**: CLI backup.go:64/65, CLI restore.go:84/85, TUI backuplist.go:117/123, TUI restore.go:307-319
- **Failed operations must be logged** — HistoryEntry includes `Status` field (success/error) and `Error` field
- **No new dependencies** — Use `os.Getenv("XDG_STATE_HOME")` and `syscall.Flock` instead of third-party libraries
- **Views use raw strings, not key.Binding** — Help uses `HelpEntry{Key, Description string}` structs, not `key.Binding` extraction
- **BackupSuccessMsg interception** — Centralize logging in main `update.go` before propagating to views
- **LogsView refresh on tab** — Add refresh hook in tab handler similar to BackupListView/RestoreView
- **Edge cases**: empty history, corrupt lines, disk full, terminal too small for overlay

---

## Work Objectives

### Core Objective
Add persistent operation history with a TUI viewer and CLI command (CI-007), and a context-aware help overlay system (CI-011) to the dotkeeper TUI.

### Concrete Deliverables
- New package: `internal/history/history.go` + `history_test.go`
- New CLI command: `internal/cli/history.go`
- Updated: `internal/tui/views/logs.go` (full implementation replacing skeleton)
- Updated: `internal/tui/views/logs_test.go` (rewritten for new implementation)
- Updated: `internal/cli/backup.go` (history logging hooks)
- Updated: `internal/cli/restore.go` (history logging hooks)
- Updated: `internal/tui/model.go` (history store field, help state)
- Updated: `internal/tui/update.go` (BackupSuccessMsg interception, help key handler, LogsView refresh)
- Updated: `internal/tui/view.go` (help overlay rendering)
- New: `internal/tui/help.go` (HelpProvider interface, overlay rendering, per-view help data)
- Updated: `cmd/dotkeeper/main.go` (add `history` command)

### Definition of Done
- [ ] `make build` succeeds
- [ ] `make test` passes (all tests, including new ones)
- [ ] `dotkeeper backup` creates history entry in `~/.local/state/dotkeeper/history.jsonl`
- [ ] `dotkeeper restore` creates history entry
- [ ] `dotkeeper history` displays operation history
- [ ] `dotkeeper history --json` outputs valid JSON
- [ ] TUI LogsView shows history entries with filtering
- [ ] TUI `?` key shows help overlay with context-aware keybindings
- [ ] Help overlay dismisses on `?`, `Esc`, or any other key

### Must Have
- JSONL format for history storage (one JSON object per line)
- History logging for BOTH success and failure paths
- Both CLI and TUI operations logged
- History file at `XDG_STATE_HOME/dotkeeper/history.jsonl`
- File permissions 0600 for history file
- Corrupt line resilience (skip malformed JSON lines)
- `list.Model` from bubbles for LogsView (following backuplist.go pattern)
- Help overlay centered with `lipgloss.Place()`
- Context-aware keybindings (global + current view)
- LogsView refreshes when switched to via Tab

### Must NOT Have (Guardrails)
- **NO new dependencies** — No `github.com/adrg/xdg`, no `github.com/gofrs/flock`. Use `os.Getenv()` + fallback and `syscall.Flock`
- **NO `bubbles/help` component** — Build custom overlay with lipgloss (bubbles/help is designed for inline bar, not overlay)
- **NO new ViewState for help** — It's `showingHelp bool` on Model. `viewCount` stays 6
- **NO history rotation/pruning in v1** — Append only, read last N entries
- **NO SQLite or structured DB** — JSONL only
- **NO date range filtering** — Only filter by operation type (backup/restore). Tab to cycle filter
- **NO help overlay animation** — Instant show/hide
- **NO `dotkeeper history clear`** or export subcommands — Only `dotkeeper history [--json] [--type TYPE]`
- **NO modification to `backup.Backup()` or `restore.Restore()` function signatures** — Logging happens at caller sites
- **NO hostname, OS version, git hash in history entries** — Only: timestamp, operation, status, file count, size, duration, backup path, error
- **NO `key.Binding` extraction from views** — Views use raw string matching; help uses `HelpEntry{Key, Description}` structs

---

## Verification Strategy

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
> ALL tasks MUST be verifiable WITHOUT any human action.

### Test Decision
- **Infrastructure exists**: YES (`make test` → `go test -v -race ./...`)
- **Automated tests**: YES (tests-after, matching project convention)
- **Framework**: `go test` (standard library)

### Agent-Executed QA Scenarios (MANDATORY — ALL tasks)

Verification tools:
| Type | Tool | How Agent Verifies |
|------|------|-------------------|
| **TUI** | interactive_bash (tmux) | Launch TUI, send keystrokes, validate output |
| **CLI** | Bash | Run commands, parse output, assert fields |
| **Unit tests** | Bash | `go test -v -race ./...`, check exit code |
| **Build** | Bash | `make build`, check exit code |

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately — independent foundations):
├── Task 1: Create internal/history package (CI-007 foundation)
└── Task 5: Implement help overlay system (CI-011 — fully independent)

Wave 2 (After Task 1):
├── Task 2: Add CLI history command + hook CLI backup/restore
├── Task 3: Implement LogsView TUI (depends on history package)
└── Task 6: Add per-view help data to all views (after Task 5)

Wave 3 (After Tasks 2, 3, 5, 6):
├── Task 4: Hook TUI backup/restore to history + LogsView refresh on Tab
└── Task 7: Final integration, build verification, all tests pass

Critical Path: Task 1 → Task 3 → Task 4 → Task 7
Parallel Speedup: ~40% faster than sequential
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2, 3, 4 | 5 |
| 2 | 1 | 7 | 3, 5, 6 |
| 3 | 1 | 4 | 2, 5, 6 |
| 4 | 3 | 7 | 6 |
| 5 | None | 6 | 1 |
| 6 | 5 | 7 | 2, 3, 4 |
| 7 | 2, 3, 4, 6 | None | None (final) |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Agents |
|------|-------|-------------------|
| 1 | 1, 5 | dispatch parallel |
| 2 | 2, 3, 6 | dispatch parallel after Wave 1 |
| 3 | 4, 7 | sequential final integration |

---

## TODOs

- [x] 1. Create `internal/history` Package — JSONL Operation History Storage

  **What to do**:
  - Create `internal/history/history.go` with:
    - `HistoryEntry` struct: `Timestamp time.Time`, `Operation string` (backup/restore), `Status string` (success/error), `FileCount int`, `TotalSize int64`, `DurationMs int64`, `BackupPath string`, `BackupName string`, `Error string` (omitempty). All fields with `json` tags.
    - `Store` struct: holds `path string` (full path to history.jsonl)
    - `NewStore() (*Store, error)`: resolve XDG_STATE_HOME using same pattern as `config.go:22-33` — `os.Getenv("XDG_STATE_HOME")` with fallback to `$HOME/.local/state`, then append `dotkeeper/history.jsonl`. Call `os.MkdirAll` on directory with `0700`.
    - `NewStoreWithPath(path string) *Store`: for testing with custom paths
    - `(s *Store) Append(entry HistoryEntry) error`: Open file with `os.O_APPEND|os.O_WRONLY|os.O_CREATE`, perm `0600`. Use `syscall.Flock` with `LOCK_EX` for advisory locking. Marshal entry to JSON, write `json_bytes + "\n"`. Call `f.Sync()`. Unlock. History write is **best-effort** — errors are logged but never crash the caller.
    - `(s *Store) Read(limit int) ([]HistoryEntry, error)`: Open file, scan line-by-line with `bufio.Scanner`. Unmarshal each line, **skip corrupt lines** (log warning, continue). Return latest `limit` entries (if limit > 0) sorted newest-first. If file doesn't exist, return empty slice (not error).
    - `(s *Store) ReadByType(opType string, limit int) ([]HistoryEntry, error)`: Same as Read but filter by operation type.
    - Helper: `EntryFromBackupResult(result *backup.BackupResult) HistoryEntry` — converts BackupResult to HistoryEntry with Status="success"
    - Helper: `EntryFromBackupError(err error) HistoryEntry` — creates error HistoryEntry with Status="error"
    - Helper: `EntryFromRestoreResult(result *restore.RestoreResult, backupPath string) HistoryEntry` — converts RestoreResult to HistoryEntry
    - Helper: `EntryFromRestoreError(err error, backupPath string) HistoryEntry`
  - Create `internal/history/history_test.go` with:
    - Test Append writes valid JSONL (read back and unmarshal)
    - Test Read returns entries newest-first
    - Test Read with limit
    - Test ReadByType filtering
    - Test corrupt line resilience (write valid entry, append "not-json\n", write another valid entry, Read should return 2 entries)
    - Test empty/nonexistent file returns empty slice
    - Test file permissions are 0600
    - Test concurrent Append (goroutines) doesn't corrupt file
    - Test NewStore creates directory if not exists
    - All tests use `t.TempDir()` for isolation

  **Must NOT do**:
  - Do NOT import `github.com/adrg/xdg` or any new dependency
  - Do NOT add rotation, pruning, or cleanup logic
  - Do NOT add any history entry fields beyond what's specified

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: New package creation with file I/O, concurrency, and comprehensive tests. Not UI work (not visual-engineering), not architecturally complex (not ultrabrain).
  - **Skills**: []
    - No special skills needed — pure Go package development

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 5)
  - **Blocks**: Tasks 2, 3, 4
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/config/config.go:22-33` — XDG path resolution pattern (os.Getenv + fallback). Follow this EXACT pattern for XDG_STATE_HOME.
  - `internal/config/config.go:46-57` — LoadFromPath pattern with error handling. Follow for Read().
  - `internal/backup/backup.go:17-25` — BackupResult struct definition. Source fields for EntryFromBackupResult helper.
  - `internal/restore/types.go:30-39` — RestoreResult struct definition. Source fields for EntryFromRestoreResult helper.

  **Test References** (testing patterns to follow):
  - `internal/tui/views/backuplist_test.go:14-34` — Pattern for using t.TempDir() and creating test files
  - `internal/backup/backup_test.go` — Package test patterns in this project

  **External References**:
  - XDG Base Directory Specification: `XDG_STATE_HOME` defaults to `$HOME/.local/state`, designed for "actions history (logs, history)"
  - JSONL format: One JSON object per line, newline-delimited. `encoding/json.Marshal()` + `"\n"`

  **Acceptance Criteria**:

  - [ ] File `internal/history/history.go` exists with Store, HistoryEntry, Append, Read, ReadByType, helper functions
  - [ ] File `internal/history/history_test.go` exists with ≥8 test functions
  - [ ] `go test -v -race ./internal/history/...` → PASS (all tests, 0 failures)
  - [ ] No new dependencies added to `go.mod`

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: History package unit tests pass
    Tool: Bash
    Preconditions: Go toolchain available
    Steps:
      1. go test -v -race ./internal/history/...
      2. Assert: exit code 0
      3. Assert: output contains "PASS"
      4. Assert: output does NOT contain "FAIL"
    Expected Result: All tests pass with race detector
    Evidence: Test output captured

  Scenario: History package compiles without new dependencies
    Tool: Bash
    Preconditions: Current go.mod
    Steps:
      1. go build ./internal/history/...
      2. Assert: exit code 0
      3. grep -c "adrg/xdg\|gofrs/flock" go.mod
      4. Assert: count is 0 (no new deps)
    Expected Result: Compiles, no new dependencies
    Evidence: Build output + grep result
  ```

  **Commit**: YES
  - Message: `feat(history): add JSONL-based operation history storage package`
  - Files: `internal/history/history.go`, `internal/history/history_test.go`
  - Pre-commit: `go test -v -race ./internal/history/...`

---

- [x] 2. Add CLI `history` Command and Hook CLI Backup/Restore to History

  **What to do**:
  - Create `internal/cli/history.go`:
    - `HistoryCommand(args []string) int` following exact pattern of `list.go`
    - Flags: `--json` (bool, JSON output), `--type` (string, filter: "backup" or "restore")
    - Create `history.NewStore()`, call `Read(50)` or `ReadByType(type, 50)` depending on flags
    - Default output: formatted table with columns: Timestamp (local time), Operation, Status, Files, Size, Duration
    - JSON output: `json.MarshalIndent(entries, "", "  ")`
    - Handle empty history: print "No operations recorded yet."
    - Error handling: if Store creation fails, print error to stderr, return 1
  - Update `cmd/dotkeeper/main.go`:
    - Add `case "history": exitCode = cli.HistoryCommand(args)` in the command switch
    - Add `history` to the help text in `printHelp()` under Commands
  - Update `internal/cli/backup.go`:
    - After `backup.Backup()` success (line 72-79): create store, call `store.Append(history.EntryFromBackupResult(result))`
    - After `backup.Backup()` error (line 65-70): create store, call `store.Append(history.EntryFromBackupError(err))`
    - History write errors are logged to stderr but do NOT change exit code (best-effort)
  - Update `internal/cli/restore.go`:
    - After `restore.Restore()` success (line 89-111): create store, call `store.Append(history.EntryFromRestoreResult(result, backupPath))`
    - After `restore.Restore()` error (line 85-87): create store, call `store.Append(history.EntryFromRestoreError(err, backupPath))`
    - History write errors are logged to stderr but do NOT change exit code

  **Must NOT do**:
  - Do NOT add `history clear`, `history export`, or any subcommands
  - Do NOT add `--limit` flag (hardcode 50 for now, future enhancement)
  - Do NOT add date range filtering to CLI
  - Do NOT change backup/restore exit codes due to history logging failures

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Mostly boilerplate CLI code following existing patterns. Hook insertions are 2-3 lines each.
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 6)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `internal/cli/list.go` — CLI command pattern with `--json` flag. Follow this EXACTLY for HistoryCommand: flag.NewFlagSet, fs.Bool for json, fs.Usage, json.MarshalIndent.
  - `internal/cli/backup.go:62-85` — Backup success/error paths where history hooks go. Insert after line 64 (result available) and after line 65 (err available).
  - `internal/cli/restore.go:84-111` — Restore success/error paths where history hooks go. Insert after line 84 (result available) and after line 85 (err available).
  - `cmd/dotkeeper/main.go:45-55` — Command switch where `history` case goes. Follow pattern of line 54-55 (schedule).
  - `cmd/dotkeeper/main.go:68-91` — Help text to update. Add `history` between `config` and `schedule`.

  **Acceptance Criteria**:

  - [ ] `go build ./cmd/dotkeeper/...` succeeds
  - [ ] `./bin/dotkeeper history` executes without error (may show empty list)
  - [ ] `./bin/dotkeeper history --json` outputs valid JSON (empty array `[]` if no history)
  - [ ] `./bin/dotkeeper --help` shows `history` command
  - [ ] History hooks in backup.go and restore.go don't affect exit codes

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: History CLI command shows help
    Tool: Bash
    Preconditions: Binary built
    Steps:
      1. make build
      2. ./bin/dotkeeper --help
      3. Assert: output contains "history"
    Expected Result: History command listed in help
    Evidence: Help output captured

  Scenario: History --json outputs valid JSON when empty
    Tool: Bash
    Preconditions: No history file exists, binary built
    Steps:
      1. XDG_STATE_HOME=$(mktemp -d) ./bin/dotkeeper history --json
      2. Pipe output to: python3 -c "import json,sys; json.load(sys.stdin); print('VALID')"
      3. Assert: output contains "VALID"
    Expected Result: Valid JSON even with no history
    Evidence: Output captured

  Scenario: History command with empty state shows friendly message
    Tool: Bash
    Preconditions: No history file
    Steps:
      1. XDG_STATE_HOME=$(mktemp -d) ./bin/dotkeeper history
      2. Assert: output contains "No operations recorded"
    Expected Result: User-friendly empty state
    Evidence: Output captured
  ```

  **Commit**: YES
  - Message: `feat(cli): add history command and hook backup/restore to history logging`
  - Files: `internal/cli/history.go`, `cmd/dotkeeper/main.go`, `internal/cli/backup.go`, `internal/cli/restore.go`
  - Pre-commit: `go build ./cmd/dotkeeper/...`

---

- [x] 3. Implement LogsView TUI — Replace Skeleton with Full Implementation

  **What to do**:
  - Rewrite `internal/tui/views/logs.go`:
    - Import `internal/history` and `charmbracelet/bubbles/list`
    - `LogsModel` struct: `config *config.Config`, `store *history.Store`, `list list.Model`, `width int`, `height int`, `filter string` (all/backup/restore), `loading bool`, `err string`
    - Define `logItem` struct implementing `list.Item` interface: `Title()` returns formatted timestamp + operation + status, `Description()` returns file count + size + duration, `FilterValue()` returns operation type
    - `NewLogs(cfg *config.Config, store *history.Store) LogsModel`: initialize list.Model with empty items, title "Operation History", SetShowHelp(false). Store reference to history store.
    - `Init() tea.Cmd`: return `m.LoadHistory()` command
    - `LoadHistory() tea.Cmd`: async command that calls `store.Read(100)` or `store.ReadByType(filter, 100)`, returns `logsLoadedMsg`
    - Define private messages: `logsLoadedMsg []history.HistoryEntry`, `logsErrorMsg error`
    - `Update()`: Handle WindowSizeMsg (resize list), logsLoadedMsg (convert entries to list items, set on list), logsErrorMsg (set error string), KeyMsg: `f` to cycle filter (all→backup→restore→all), `r` to refresh. Delegate unhandled to list.Model.
    - `View()`: Use `strings.Builder`, `DefaultStyles()`. Show title "Operation History" + active filter indicator. Show list. Show status/error. Show help: "f: filter | r: refresh | ↑/↓: navigate". Handle empty state: "No operations recorded yet. Run a backup to get started!"
  - Rewrite `internal/tui/views/logs_test.go`:
    - Test NewLogs creates valid model with store
    - Test Init returns a command (LoadHistory)
    - Test Update with logsLoadedMsg populates list
    - Test Update with empty logsLoadedMsg shows empty state
    - Test filter cycling (f key)
    - Test WindowSizeMsg resizes list
    - Test View contains "Operation History"
    - Use t.TempDir() for history store isolation

  **Must NOT do**:
  - Do NOT add date range filtering
  - Do NOT add search functionality
  - Do NOT add detail view/expand for individual entries
  - Do NOT add pagination — list.Model handles scrolling natively

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: TUI view implementation following BubbleTea patterns. Involves UI layout and component integration.
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 6)
  - **Blocks**: Task 4
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `internal/tui/views/backuplist.go:16-23` — `backupItem` struct implementing `list.Item` (Title, Description, FilterValue). Follow this EXACT pattern for `logItem`.
  - `internal/tui/views/backuplist.go:47-63` — NewBackupList constructor with `list.New()`, `l.Title`, `l.SetShowHelp(false)`. Follow this for NewLogs.
  - `internal/tui/views/backuplist.go:69-87` — Async Refresh() command returning custom message. Follow this for LoadHistory().
  - `internal/tui/views/backuplist.go:102-170` — Update pattern: WindowSizeMsg resize, custom message handling, key handling, delegating to list.Model. Follow this structure.
  - `internal/tui/views/backuplist.go:172-198` — View pattern with strings.Builder, DefaultStyles(), status/error display, help text. Follow this.
  - `internal/tui/views/dashboard.go:49-71` — Simple View pattern with DefaultStyles() usage.

  **Test References**:
  - `internal/tui/views/backuplist_test.go` — TUI view test pattern: create model, call Init, execute command, cast message, call Update, verify View output.
  - `internal/tui/views/logs_test.go` — Current skeleton tests to REPLACE (not extend).

  **Acceptance Criteria**:

  - [ ] `internal/tui/views/logs.go` is fully implemented (>100 lines)
  - [ ] `internal/tui/views/logs_test.go` has ≥5 test functions
  - [ ] `go test -v -race ./internal/tui/views/...` → PASS
  - [ ] LogsView shows "No operations recorded" when history is empty
  - [ ] LogsView list shows entries from history file when populated

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: LogsView unit tests pass
    Tool: Bash
    Preconditions: Go toolchain
    Steps:
      1. go test -v -race ./internal/tui/views/...
      2. Assert: exit code 0
      3. Assert: output contains "PASS"
      4. Assert: output does NOT contain "FAIL"
    Expected Result: All view tests pass
    Evidence: Test output captured

  Scenario: LogsView renders with empty history
    Tool: Bash
    Preconditions: history package available
    Steps:
      1. go test -v -run TestLogsModel ./internal/tui/views/...
      2. Assert: test passes
    Expected Result: Empty state rendering works
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): implement LogsView with operation history display and filtering`
  - Files: `internal/tui/views/logs.go`, `internal/tui/views/logs_test.go`
  - Pre-commit: `go test -v -race ./internal/tui/views/...`

---

- [x] 4. Wire TUI Backup/Restore to History + LogsView Refresh on Tab

  **What to do**:
  - Update `internal/tui/model.go`:
    - Add `history *history.Store` field to Model struct
    - In `NewModel()`: create store via `history.NewStore()`. If error, log warning (don't crash TUI). Pass store to `views.NewLogs(cfg, store)`.
    - Update import to include `"github.com/diogo/dotkeeper/internal/history"`
    - After `SetupCompleteMsg` handler (line 66-76): also create store and pass to NewLogs
  - Update `internal/tui/update.go`:
    - Add handling for `views.BackupSuccessMsg` in main Update (BEFORE the view-specific routing switch at line 130): When BackupSuccessMsg is received, call `m.history.Append(history.EntryFromBackupResult(msg.Result))` — ignore errors (best-effort). Then let message continue flowing to BackupListModel.
    - Add handling for `views.BackupErrorMsg` similarly: call `m.history.Append(history.EntryFromBackupError(msg.Error))`.
    - Add handling for restore completion: When `restoreCompleteMsg` arrives (it's private to restore.go, so intercept via a NEW public message `RestoreSuccessMsg` or handle it differently — see note below).
    - **For restore history logging**: Since `restoreCompleteMsg` is private to `views/restore.go`, add a public `RestoreCompleteMsg` to helpers.go (like `RefreshBackupListMsg`). In `restore.go`, after handling `restoreCompleteMsg` at line 307-312, emit a `RestoreCompleteMsg` cmd. In main `update.go`, handle it for history logging.
    - Alternatively (simpler): Pass the history store to RestoreModel and log inside `restore.go`'s `restoreCompleteMsg` handler. This avoids message refactoring.
    - **Recommended approach**: Pass `*history.Store` to `NewRestore(cfg, store)` and `NewBackupList(cfg, store)`. Log directly in the view's Update() handler where success/error messages are received. This is simpler and follows the existing pattern where views handle their own messages.
    - Add LogsView refresh on Tab: In the tab handler (line 102-111), add:
      ```
      if m.state == LogsView && prevState != LogsView {
          cmds = append(cmds, m.logs.LoadHistory())
      }
      ```
    - Expose `LoadHistory()` as a public method on LogsModel (it already will be from Task 3, but verify it's exported).

  **Must NOT do**:
  - Do NOT modify `backup.Backup()` or `restore.Restore()` function signatures
  - Do NOT let history logging errors crash the TUI
  - Do NOT add complex message forwarding chains — keep it simple with store references

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Integration task touching multiple files with subtle wiring concerns. Needs understanding of message flow.
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (sequential after Tasks 2, 3)
  - **Blocks**: Task 7
  - **Blocked By**: Tasks 1, 3

  **References**:

  **Pattern References**:
  - `internal/tui/model.go:56-66` — NewModel constructor where views are initialized. Add store creation here.
  - `internal/tui/update.go:92-94` — RefreshBackupListMsg handling pattern. Follow this for centralized message handling.
  - `internal/tui/update.go:102-111` — Tab handler with view-specific refresh. Add LogsView refresh here.
  - `internal/tui/views/backuplist.go:115-127` — BackupSuccessMsg/BackupErrorMsg handling. Hook history logging here.
  - `internal/tui/views/restore.go:307-319` — restoreCompleteMsg/restoreErrorMsg handling. Hook history logging here.
  - `internal/tui/update.go:64-82` — SetupCompleteMsg handler that re-creates all views. Must also create history store here.

  **Acceptance Criteria**:

  - [ ] `make build` succeeds
  - [ ] `go test -v -race ./internal/tui/...` → PASS
  - [ ] Model struct has `history` field
  - [ ] NewModel creates history store and passes to views
  - [ ] Tab to LogsView triggers history reload

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: TUI builds with history integration
    Tool: Bash
    Preconditions: All previous tasks committed
    Steps:
      1. make build
      2. Assert: exit code 0
      3. go vet ./internal/tui/...
      4. Assert: exit code 0
    Expected Result: Clean build with no vet warnings
    Evidence: Build output captured

  Scenario: All TUI tests pass after integration
    Tool: Bash
    Preconditions: All changes applied
    Steps:
      1. go test -v -race ./internal/tui/...
      2. Assert: exit code 0
      3. Assert: output does NOT contain "FAIL"
    Expected Result: All tests pass
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): wire history logging to TUI backup/restore and add LogsView refresh on tab`
  - Files: `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/views/backuplist.go`, `internal/tui/views/restore.go`
  - Pre-commit: `go test -v -race ./internal/tui/...`

---

- [x] 5. Implement Help Overlay System — Core Infrastructure

  **What to do**:
  - Create `internal/tui/help.go`:
    - Define `HelpEntry struct { Key string; Description string }` — plain struct, NOT key.Binding
    - Define `HelpProvider interface { HelpBindings() []HelpEntry }` — views optionally implement this
    - Define `globalHelp() []HelpEntry` returning: `{"Tab", "Switch views"}`, `{"q", "Quit"}`, `{"?", "Toggle help"}`
    - Define `renderHelpOverlay(global []HelpEntry, viewHelp []HelpEntry, width, height int) string`:
      - Calculate overlay dimensions: width = min(60, termWidth-4), height = min(20, termHeight-4)
      - If terminal too small (width < 40 or height < 15): render full-screen help instead of overlay
      - Build content: "⌨ Keyboard Shortcuts" title, then "Global" section with global keys, then "Current View" section with view keys
      - Style: `lipgloss.RoundedBorder()`, `BorderForeground(#7D56F4)`, `Background(#1e1e2e)`, `Padding(1, 2)`
      - Center using `lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, overlay)`
      - Return rendered string
  - Update `internal/tui/model.go`:
    - Add `showingHelp bool` field to Model struct
  - Update `internal/tui/update.go`:
    - In KeyMsg handler, BEFORE quit/tab handling (after line 96): check `if m.showingHelp`. If showing help, dismiss on ANY key press: `m.showingHelp = false; return m, nil` (consumes the key, does nothing else)
    - Add `if key.Matches(msg, keys.Help)` handler: `m.showingHelp = !m.showingHelp; return m, nil`
    - Order is important: check showingHelp dismissal FIRST, then Help toggle, then quit/tab
  - Update `internal/tui/view.go`:
    - After building the normal view but BEFORE returning (before line 52): check `if m.showingHelp`. If true, call `renderHelpOverlay()` passing global help + current view's help (via HelpProvider interface assertion). Return the overlay instead of normal view.
    - Get view help: type-assert current view to `HelpProvider`. If it implements it, use its bindings. Otherwise, pass empty slice.

  **Must NOT do**:
  - Do NOT import `bubbles/help` — custom implementation with lipgloss
  - Do NOT add new ViewState — help is a bool overlay, not a state
  - Do NOT add animation or transitions
  - Do NOT modify any view files in this task — per-view help data is Task 6

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Involves careful message flow ordering and overlay rendering logic. Not purely visual (needs correct key interception).
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 1)
  - **Blocks**: Task 6
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References**:
  - `internal/tui/update.go:96-112` — Key handling order in main Update. Help overlay dismiss MUST come before quit/tab to prevent keys leaking through.
  - `internal/tui/view.go:9-53` — Main View function structure. Help overlay intercepts at the end, replacing normal output.
  - `internal/tui/views/styles.go:20-50` — Color palette. Use `#7D56F4` (purple) for overlay border to match existing theme.
  - `internal/tui/model.go:22-42` — Model struct where `showingHelp` goes.
  - `internal/tui/update.go:37-57` — propagateWindowSize pattern. WindowSizeMsg must still be processed when help is showing (for resize support).

  **External References**:
  - Soft-serve context-aware help pattern: Parent composes common + tab-specific keybindings
  - `lipgloss.Place()` documentation: Centers content in a given width/height area

  **Acceptance Criteria**:

  - [ ] `internal/tui/help.go` exists with HelpEntry, HelpProvider, globalHelp, renderHelpOverlay
  - [ ] Model struct has `showingHelp bool`
  - [ ] `?` key toggles help overlay visibility
  - [ ] Any key dismisses help when showing
  - [ ] WindowSizeMsg still processed when help is showing
  - [ ] `make build` succeeds
  - [ ] `go test -v -race ./internal/tui/...` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Help overlay builds and tests pass
    Tool: Bash
    Preconditions: Go toolchain
    Steps:
      1. make build
      2. Assert: exit code 0
      3. go test -v -race ./internal/tui/...
      4. Assert: exit code 0
    Expected Result: Clean build, all tests pass
    Evidence: Build/test output captured

  Scenario: Help infrastructure exists in code
    Tool: Bash (grep)
    Preconditions: Task code committed
    Steps:
      1. grep -c "showingHelp" internal/tui/model.go
      2. Assert: count >= 1
      3. grep -c "HelpProvider" internal/tui/help.go
      4. Assert: count >= 1
      5. grep "keys.Help" internal/tui/update.go | grep -c "Matches"
      6. Assert: count >= 1 (? key now handled)
    Expected Result: All help infrastructure present
    Evidence: grep output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): add help overlay system with HelpProvider interface and ? key toggle`
  - Files: `internal/tui/help.go`, `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/view.go`
  - Pre-commit: `go test -v -race ./internal/tui/...`

---

- [x] 6. Add Per-View Help Bindings to All TUI Views

  **What to do**:
  - Implement `HelpProvider` interface on views that have custom keybindings:
  - `internal/tui/views/dashboard.go`: Add `func (m DashboardModel) HelpBindings() []tui.HelpEntry` returning: `{"b", "Go to backups"}`, `{"r", "Go to restore"}`, `{"s", "Go to settings"}`
    - Note: Import cycle issue — `views` package can't import `tui` package. **Solution**: Define `HelpEntry` in `views/helpers.go` instead of `tui/help.go`. Then `help.go` uses the same type from views package. OR define HelpEntry in a shared location. **Simplest**: Define `HelpEntry` directly in `internal/tui/views/helpers.go` and have `tui/help.go` import it from views. This follows the existing pattern where `tui` imports `views`.
  - `internal/tui/views/backuplist.go`: Add HelpBindings returning: `{"n/c", "New backup"}`, `{"r", "Refresh list"}`, `{"↑/↓", "Navigate"}`, `{"Enter", "Select"}`
  - `internal/tui/views/restore.go`: Add HelpBindings that are PHASE-AWARE. Check `m.phase` and return appropriate bindings:
    - Phase 0: `{"Enter", "Select backup"}`, `{"r", "Refresh"}`, `{"↑/↓", "Navigate"}`
    - Phase 1: `{"Enter", "Submit password"}`, `{"Esc", "Back"}`
    - Phase 2: `{"Space", "Toggle file"}`, `{"a", "Select all"}`, `{"n", "Select none"}`, `{"d", "View diff"}`, `{"Enter", "Restore"}`, `{"Esc", "Back"}`
    - Phase 4: `{"j/k", "Scroll"}`, `{"g/G", "Top/Bottom"}`, `{"Esc", "Back"}`
    - Phase 5: `{"any key", "Continue"}`
  - `internal/tui/views/settings.go`: Add HelpBindings that are MODE-AWARE:
    - Read-only: `{"e", "Edit mode"}`
    - Edit mode: `{"↑/↓", "Navigate"}`, `{"Enter", "Edit field"}`, `{"a", "Add item"}`, `{"d", "Delete item"}`, `{"s", "Save"}`, `{"Esc", "Exit edit"}`
  - `internal/tui/views/logs.go`: Add HelpBindings: `{"f", "Cycle filter"}`, `{"r", "Refresh"}`, `{"↑/↓", "Navigate"}`
  - `internal/tui/views/filebrowser.go`: Add HelpBindings: `{"Enter", "Select"}`, `{"↑/↓", "Navigate"}`
  - Update `internal/tui/view.go` (or `help.go`): type-assert each view to HelpProvider when rendering overlay. Views that don't implement it get empty view-specific section.

  **Must NOT do**:
  - Do NOT refactor views to use `key.Binding` internally — keep existing raw string matching
  - Do NOT add new key bindings — only document existing ones
  - Do NOT modify any view's Update() or View() methods (only add HelpBindings method)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Mechanical task — adding a method to each view that returns static data. No logic changes.
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3)
  - **Blocks**: Task 7
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `internal/tui/views/backuplist.go:129-160` — Keyboard shortcuts used in BackupListModel. Source of truth for help entries.
  - `internal/tui/views/restore.go:322-441` — Phase-specific keyboard handling. Source for restore view help entries.
  - `internal/tui/views/settings.go:68-193` — Mode-specific keyboard handling. Source for settings view help entries.
  - `internal/tui/views/dashboard.go:49-71` — Dashboard view shortcuts (b, r, s) shown in View(). Source for dashboard help entries.
  - `internal/tui/update.go:114-127` — Dashboard shortcuts handled in main Update. These are the actual key bindings.
  - `internal/tui/views/helpers.go` — Shared types location. Put HelpEntry here to avoid import cycles.

  **Acceptance Criteria**:

  - [ ] All 6 views have HelpBindings method (Dashboard, BackupList, Restore, Settings, Logs, FileBrowser)
  - [ ] RestoreModel.HelpBindings is phase-aware
  - [ ] SettingsModel.HelpBindings is mode-aware
  - [ ] `make build` succeeds (no import cycles)
  - [ ] `go test -v -race ./internal/tui/...` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: All views implement HelpProvider
    Tool: Bash (grep)
    Preconditions: Code committed
    Steps:
      1. grep -l "HelpBindings" internal/tui/views/*.go
      2. Assert: at least 6 files listed (dashboard, backuplist, restore, settings, logs, filebrowser)
      3. make build
      4. Assert: exit code 0 (no import cycles)
    Expected Result: All views have HelpBindings, build succeeds
    Evidence: grep + build output

  Scenario: All TUI tests pass
    Tool: Bash
    Preconditions: All changes applied
    Steps:
      1. go test -v -race ./internal/tui/...
      2. Assert: exit code 0
    Expected Result: No regressions
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): add context-aware help bindings to all views`
  - Files: `internal/tui/views/dashboard.go`, `internal/tui/views/backuplist.go`, `internal/tui/views/restore.go`, `internal/tui/views/settings.go`, `internal/tui/views/logs.go`, `internal/tui/views/filebrowser.go`, `internal/tui/views/helpers.go`, `internal/tui/help.go`
  - Pre-commit: `go test -v -race ./internal/tui/...`

---

- [x] 7. Final Integration — Build, Full Test Suite, TUI Smoke Test

  **What to do**:
  - Run `make build` — must succeed with zero warnings
  - Run `make test` — ALL tests must pass (including new ones)
  - Run `go vet ./...` — no issues
  - TUI smoke test via tmux:
    - Launch `./bin/dotkeeper` in tmux
    - Press Tab 5 times to reach LogsView — verify it shows "No operations recorded" or history entries
    - Press `?` — verify help overlay appears with keybindings
    - Press any key — verify help overlay dismisses
    - Press `q` — verify TUI exits cleanly
  - CLI smoke test:
    - `./bin/dotkeeper --help` — verify `history` command listed
    - `./bin/dotkeeper history` — verify outputs empty message or history
    - `./bin/dotkeeper history --json` — verify valid JSON output
  - Fix any failing tests or build issues discovered during integration

  **Must NOT do**:
  - Do NOT add new features during integration
  - Do NOT refactor working code for style preferences

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Integration testing requires careful verification across all components. May need debugging.
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (final, sequential)
  - **Blocks**: None (final task)
  - **Blocked By**: Tasks 2, 3, 4, 6

  **References**:

  **Pattern References**:
  - `Makefile:29-32` — `make test` command: `go test -v -race -coverprofile=coverage.out ./...`
  - `Makefile:22-26` — `make build` command
  - `cmd/dotkeeper/main.go:31-37` — TUI launch pattern for understanding how to test via tmux

  **Acceptance Criteria**:

  - [ ] `make build` → exit code 0
  - [ ] `make test` → exit code 0, all tests PASS
  - [ ] `go vet ./...` → exit code 0
  - [ ] TUI launches without crash
  - [ ] LogsView reachable via Tab and displays content
  - [ ] `?` key shows help overlay
  - [ ] `dotkeeper history` works from CLI

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Full build succeeds
    Tool: Bash
    Preconditions: All tasks committed
    Steps:
      1. make clean && make build
      2. Assert: exit code 0
      3. test -f ./bin/dotkeeper
      4. Assert: binary exists
    Expected Result: Clean build produces binary
    Evidence: Build output captured

  Scenario: Full test suite passes
    Tool: Bash
    Preconditions: All code committed
    Steps:
      1. make test
      2. Assert: exit code 0
      3. Assert: output contains "ok" for each package
      4. Assert: output does NOT contain "FAIL"
    Expected Result: All packages pass with race detector
    Evidence: Full test output captured

  Scenario: TUI launches and help overlay works
    Tool: interactive_bash (tmux)
    Preconditions: Binary built
    Steps:
      1. tmux new-session -d -s dotkeeper-test "./bin/dotkeeper"
      2. Wait 2 seconds for TUI to start
      3. Send keys: "?" (help toggle)
      4. Wait 1 second
      5. Capture pane content
      6. Assert: output contains "Keyboard Shortcuts" or "Global" (help overlay visible)
      7. Send keys: "Escape" (dismiss help)
      8. Wait 1 second
      9. Send keys: "q" (quit)
    Expected Result: Help overlay appears and dismisses
    Evidence: Pane captures at each step

  Scenario: TUI LogsView reachable
    Tool: interactive_bash (tmux)
    Preconditions: Binary built
    Steps:
      1. tmux new-session -d -s dotkeeper-logs "./bin/dotkeeper"
      2. Wait 2 seconds
      3. Send Tab key 5 times (Dashboard→FileBrowser→BackupList→Restore→Settings→Logs)
      4. Wait 1 second
      5. Capture pane content
      6. Assert: output contains "Operation History" or "No operations recorded"
      7. Send keys: "q"
    Expected Result: LogsView is reachable and renders
    Evidence: Pane capture showing LogsView

  Scenario: CLI history command works
    Tool: Bash
    Preconditions: Binary built
    Steps:
      1. ./bin/dotkeeper --help
      2. Assert: output contains "history"
      3. XDG_STATE_HOME=$(mktemp -d) ./bin/dotkeeper history
      4. Assert: exit code 0
      5. XDG_STATE_HOME=$(mktemp -d) ./bin/dotkeeper history --json
      6. Assert: exit code 0
      7. Pipe --json output to python3 -c "import json,sys; d=json.load(sys.stdin); print('VALID')"
      8. Assert: output contains "VALID"
    Expected Result: History command works with both formats
    Evidence: Command outputs captured
  ```

  **Commit**: YES (if any fixes were needed)
  - Message: `fix(integration): resolve integration issues for CI-007 and CI-011`
  - Files: any files fixed during integration
  - Pre-commit: `make test`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `feat(history): add JSONL-based operation history storage package` | internal/history/* | `go test ./internal/history/...` |
| 2 | `feat(cli): add history command and hook backup/restore to history logging` | internal/cli/*, cmd/dotkeeper/main.go | `go build ./cmd/dotkeeper/...` |
| 3 | `feat(tui): implement LogsView with operation history display and filtering` | internal/tui/views/logs*.go | `go test ./internal/tui/views/...` |
| 4 | `feat(tui): wire history logging to TUI backup/restore and add LogsView refresh on tab` | internal/tui/*.go, views/backuplist.go, views/restore.go | `go test ./internal/tui/...` |
| 5 | `feat(tui): add help overlay system with HelpProvider interface and ? key toggle` | internal/tui/*.go | `go test ./internal/tui/...` |
| 6 | `feat(tui): add context-aware help bindings to all views` | internal/tui/views/*.go, internal/tui/help.go | `go test ./internal/tui/...` |
| 7 | `fix(integration): resolve integration issues for CI-007 and CI-011` (if needed) | any | `make test` |

---

## Success Criteria

### Verification Commands
```bash
make build                          # Expected: exit 0, binary at ./bin/dotkeeper
make test                           # Expected: exit 0, all tests PASS
go vet ./...                        # Expected: exit 0, no issues
./bin/dotkeeper --help              # Expected: shows "history" command
./bin/dotkeeper history             # Expected: exit 0, shows history or empty message
./bin/dotkeeper history --json      # Expected: exit 0, valid JSON
```

### Final Checklist
- [ ] All "Must Have" items present
- [ ] All "Must NOT Have" guardrails respected (no new deps, no bubbles/help, no new ViewState, etc.)
- [ ] `internal/history/` package with full tests
- [ ] `dotkeeper history` CLI command
- [ ] LogsView fully implemented (not skeleton)
- [ ] Help overlay works with `?` key
- [ ] Help overlay shows context-aware bindings per view
- [ ] All existing tests still pass
- [ ] No new dependencies in `go.mod`
