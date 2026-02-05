# Restore TUI Implementation

## TL;DR

> **Quick Summary**: Implement the RestoreModel TUI view to allow users to select backups, enter password, choose files to restore, preview diffs, and execute restore with automatic conflict handling.
> 
> **Deliverables**:
> - Complete `internal/tui/views/restore.go` implementation
> - Unit tests in `internal/tui/views/restore_test.go`
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: NO - sequential (each task depends on previous)
> **Critical Path**: Task 1 → Task 2 → Task 3 → Task 4 → Task 5 → Task 6

---

## Context

### Original Request
User requested implementation of the restore functionality TUI, which is currently a stub showing "Select a backup to restore (implementation pending)".

### Interview Summary
**Key Discussions**:
- **Preview diff**: Optional - press 'd' to see diff, 'enter' to restore directly
- **File selection**: Individual files can be selected (multi-select with space key)
- **Conflict resolution**: Automatic with .bak backup creation
- **Tests**: Yes, after implementation

**Research Findings**:
- **Restore package is COMPLETE**: All functions needed already exist (`restore.Restore()`, `restore.PreviewRestore()`, `restore.ListBackupContents()`, `restore.ValidateBackup()`)
- **TUI RestoreModel is STUB**: Only handles WindowSizeMsg, returns placeholder text
- **BackupListModel is reference**: Follow its patterns for consistency
- **View already wired**: RestoreView exists in ViewState enum, routing exists in update.go

### Metis Review
**Identified Gaps** (addressed):
- **Password retry**: Default to 3 attempts before returning to backup list
- **Zero files selected**: Disable restore, show "Select at least one file"
- **Progress indication**: Use `ProgressCallback` for real-time file-by-file updates
- **Navigation flow**: ESC returns to previous phase at each step
- **Diff display**: Scrollable viewport within same view (not separate modal)
- **Cancel mid-restore**: Disabled (atomic file writes prevent partial state)
- **Key conflicts**: Use unique keys: space (select), d (diff), enter (confirm), ESC (back)

---

## Work Objectives

### Core Objective
Implement a fully functional restore TUI view that allows users to:
1. Select a backup from the list
2. Enter decryption password
3. Choose specific files to restore (multi-select)
4. Optionally preview diffs before restoring
5. Execute restore with automatic .bak for conflicts
6. See progress and results

### Concrete Deliverables
- `internal/tui/views/restore.go` - Complete implementation (~350-400 lines)
- `internal/tui/views/restore_test.go` - Unit tests (~150-200 lines)

### Definition of Done
- [x] `go build ./...` compiles without errors
- [x] `go test ./internal/tui/views/... -v` passes all tests
- [x] TUI navigates through all restore phases correctly
- [x] Restore operation successfully restores files with .bak conflict handling

### Must Have
- Backup list view (reuse pattern from BackupListModel)
- Password entry with validation via `restore.ValidateBackup()`
- File list with multi-select (space key toggles selection)
- Diff preview on 'd' key via `restore.PreviewRestore()`
- Restore execution via `restore.Restore()` with progress
- Success/error messages
- ESC navigation to previous phase

### Must NOT Have (Guardrails)
- **DO NOT modify `internal/restore/` package** - it's complete
- **DO NOT modify `internal/tui/model.go`** - view already wired
- **DO NOT modify `internal/tui/update.go`** - routing already exists
- **DO NOT add per-file conflict resolution UI** - always automatic .bak
- **DO NOT add backup comparison features** - out of scope
- **DO NOT cache password between operations** - security concern
- **DO NOT add new BubbleTea dependencies** - use existing list.Model, textinput.Model, viewport.Model

---

## Verification Strategy

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
> ALL verification is executed by the agent using tools (Playwright for browser, tmux for TUI, Bash for tests).

### Test Decision
- **Infrastructure exists**: YES (go test)
- **Automated tests**: YES (Tests-after)
- **Framework**: go test with testing package

### Agent-Executed QA Scenarios (MANDATORY)

**Tool Selection**:
| Type | Tool |
|------|------|
| Unit tests | Bash (`go test`) |
| TUI interaction | interactive_bash (tmux) |
| Build verification | Bash (`go build`) |

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Sequential - dependencies):
Task 1 → Task 2 → Task 3 → Task 4 → Task 5 → Task 6

No parallelization: Each task builds on the previous phase implementation.
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2, 3, 4, 5, 6 | None |
| 2 | 1 | 3, 4, 5, 6 | None |
| 3 | 2 | 4, 5, 6 | None |
| 4 | 3 | 5, 6 | None |
| 5 | 4 | 6 | None |
| 6 | 5 | None | None |

---

## TODOs

- [x] 1. Implement RestoreModel struct and backup list loading

  **What to do**:
  - Define RestoreModel struct following BackupListModel pattern:
    - `config *config.Config`
    - `width, height int`
    - `backupList list.Model` - for backup selection
    - `phase int` - current phase (0=backup list, 1=password, 2=file select, 3=restoring, 4=diff preview)
    - `selectedBackup string` - path to selected backup
    - `passwordInput textinput.Model`
    - `fileList list.Model` - for file selection
    - `selectedFiles map[string]bool` - tracks selected files
    - `restoreStatus string`
    - `restoreError string`
    - `passwordAttempts int`
    - `viewport viewport.Model` - for diff display
    - `currentDiff string` - diff content
  - Implement `NewRestore(cfg *config.Config) RestoreModel` constructor
  - Implement `Init() tea.Cmd` - returns `m.refreshBackups()`
  - Implement `refreshBackups() tea.Cmd` - scans backup directory, returns `backupsLoadedMsg`
  - Define message type: `type backupsLoadedMsg []list.Item`
  - Handle WindowSizeMsg and backupsLoadedMsg in Update()
  - Render backup list in View() for phase 0

  **Must NOT do**:
  - Don't modify internal/tui/model.go
  - Don't add new dependencies
  - Don't implement password or file selection yet (next tasks)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single file modification with clear pattern to follow
  - **Skills**: [`git-master`]
    - `git-master`: For atomic commits after task completion

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential
  - **Blocks**: Tasks 2, 3, 4, 5, 6
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/tui/views/backuplist.go:14-32` - BackupListModel struct definition (follow this pattern)
  - `internal/tui/views/backuplist.go:34-52` - NewBackupList constructor (follow this pattern)
  - `internal/tui/views/backuplist.go:54-66` - Init() and Refresh() methods (follow this pattern)
  - `internal/tui/views/backuplist.go:68-78` - backupsLoadedMsg handling (follow this pattern)

  **API/Type References**:
  - `internal/config/config.go:Config` - Config struct to use
  - `internal/tui/views/restore.go` - Current stub to replace

  **Test References**:
  - `internal/tui/views/backuplist_test.go:TestNewBackupList` - Test pattern to follow later

  **WHY Each Reference Matters**:
  - `backuplist.go` struct: Shows exact field types and patterns for list + textinput combo
  - `backuplist.go` constructor: Shows how to initialize list.Model with custom delegate
  - `backuplist.go` Refresh: Shows async pattern for loading backup files

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Build succeeds with new RestoreModel
    Tool: Bash
    Preconditions: None
    Steps:
      1. cd /home/diogo/dev/backup-dotfiles
      2. go build ./...
    Expected Result: Exit code 0, no errors
    Evidence: Command output captured

  Scenario: RestoreModel initializes with backup list
    Tool: interactive_bash (tmux)
    Preconditions: At least one backup exists in BackupDir
    Steps:
      1. tmux new-session -d -s dotkeeper-test
      2. tmux send-keys -t dotkeeper-test "./bin/dotkeeper" Enter
      3. Wait 2s for TUI to load
      4. tmux send-keys -t dotkeeper-test "Tab" (cycle to RestoreView)
      5. tmux capture-pane -t dotkeeper-test -p > /tmp/restore-view.txt
      6. Assert: /tmp/restore-view.txt contains "Restore" title
      7. Assert: /tmp/restore-view.txt contains backup name pattern (backup-*)
      8. tmux send-keys -t dotkeeper-test "q"
      9. tmux kill-session -t dotkeeper-test
    Expected Result: Restore view shows list of available backups
    Evidence: /tmp/restore-view.txt
  ```

  **Commit**: YES
  - Message: `feat(tui): add RestoreModel struct and backup list loading`
  - Files: `internal/tui/views/restore.go`
  - Pre-commit: `go build ./... && go test ./internal/tui/views/... -v`

---

- [x] 2. Implement password entry and validation

  **What to do**:
  - Add phase 1 handling: password entry modal
  - On backup selection (enter key in phase 0): transition to phase 1, focus passwordInput
  - On enter in phase 1: validate password via `restore.ValidateBackup(selectedBackup, password)`
  - If valid: transition to phase 2
  - If invalid: increment passwordAttempts, show error, stay in phase 1
  - If passwordAttempts >= 3: return to phase 0 with error "Too many attempts"
  - On ESC in phase 1: return to phase 0, clear passwordInput
  - Define message types:
    - `type passwordValidMsg struct{}` - password correct
    - `type passwordInvalidMsg struct{ err error }` - password wrong

  **Must NOT do**:
  - Don't cache password after validation
  - Don't modify restore package
  - Don't implement file selection yet

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Focused modification adding one phase to existing structure
  - **Skills**: [`git-master`]
    - `git-master`: For atomic commits

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential
  - **Blocks**: Tasks 3, 4, 5, 6
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `internal/tui/views/backuplist.go:92-120` - Password input mode handling (creatingBackup flag pattern)
  - `internal/tui/views/backuplist.go:138-156` - KeyMsg handling with modal state

  **API/Type References**:
  - `internal/restore/restore.go:ValidateBackup` - Function signature: `ValidateBackup(backupPath, password string) error`

  **WHY Each Reference Matters**:
  - `backuplist.go` password handling: Shows how to switch between list mode and input mode
  - `ValidateBackup`: The function to call for password verification

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Correct password advances to file selection
    Tool: interactive_bash (tmux)
    Preconditions: Test backup exists with known password "test123"
    Steps:
      1. tmux new-session -d -s dotkeeper-test
      2. tmux send-keys -t dotkeeper-test "./bin/dotkeeper" Enter
      3. Wait 2s
      4. tmux send-keys -t dotkeeper-test "Tab" (to RestoreView)
      5. tmux send-keys -t dotkeeper-test "Enter" (select first backup)
      6. Wait 500ms
      7. tmux capture-pane -t dotkeeper-test -p > /tmp/password-prompt.txt
      8. Assert: /tmp/password-prompt.txt contains "password" or "Password"
      9. tmux send-keys -t dotkeeper-test "test123" Enter
      10. Wait 2s (Argon2id is slow)
      11. tmux capture-pane -t dotkeeper-test -p > /tmp/file-select.txt
      12. Assert: /tmp/file-select.txt shows file list (not password prompt)
      13. tmux send-keys -t dotkeeper-test "q"
      14. tmux kill-session -t dotkeeper-test
    Expected Result: After correct password, shows file selection
    Evidence: /tmp/password-prompt.txt, /tmp/file-select.txt

  Scenario: Wrong password shows error and allows retry
    Tool: interactive_bash (tmux)
    Preconditions: Test backup exists
    Steps:
      1. tmux new-session -d -s dotkeeper-test
      2. tmux send-keys -t dotkeeper-test "./bin/dotkeeper" Enter
      3. Wait 2s
      4. tmux send-keys -t dotkeeper-test "Tab" Enter
      5. tmux send-keys -t dotkeeper-test "wrongpassword" Enter
      6. Wait 2s
      7. tmux capture-pane -t dotkeeper-test -p > /tmp/wrong-pw.txt
      8. Assert: /tmp/wrong-pw.txt contains "error" or "invalid" or "failed"
      9. Assert: /tmp/wrong-pw.txt still shows password input (retry possible)
      10. tmux send-keys -t dotkeeper-test "q"
      11. tmux kill-session -t dotkeeper-test
    Expected Result: Error shown, can retry password
    Evidence: /tmp/wrong-pw.txt

  Scenario: ESC returns to backup list
    Tool: interactive_bash (tmux)
    Preconditions: In password entry phase
    Steps:
      1. (Start from password prompt)
      2. tmux send-keys -t dotkeeper-test Escape
      3. tmux capture-pane -t dotkeeper-test -p > /tmp/esc-back.txt
      4. Assert: /tmp/esc-back.txt shows backup list (not password prompt)
    Expected Result: Returns to backup selection
    Evidence: /tmp/esc-back.txt
  ```

  **Commit**: YES
  - Message: `feat(tui): add password entry and validation to restore view`
  - Files: `internal/tui/views/restore.go`
  - Pre-commit: `go build ./... && go test ./internal/tui/views/... -v`

---

- [x] 3. Implement file list and multi-select

  **What to do**:
  - After password validation, load file list via `restore.ListBackupContents(selectedBackup, password)`
  - Define `type filesLoadedMsg struct { files []restore.FileEntry }`
  - Create file list items with checkbox visual: `[ ] filename` / `[x] filename`
  - Implement multi-select: space key toggles `selectedFiles[path]`
  - Add "Select All" (a key) and "Select None" (n key) shortcuts
  - Show selected count: "3 of 10 files selected"
  - Disable restore if 0 files selected (show message)
  - On Enter: if files selected, transition to restore execution (phase 3)
  - On ESC: return to backup list (phase 0), clear state

  **Must NOT do**:
  - Don't implement diff preview yet
  - Don't implement actual restore yet
  - Don't allow proceeding with 0 files

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Adding file list phase, follows same patterns
  - **Skills**: [`git-master`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential
  - **Blocks**: Tasks 4, 5, 6
  - **Blocked By**: Task 2

  **References**:

  **Pattern References**:
  - `internal/tui/views/backuplist.go:68-78` - List item handling pattern
  - `internal/tui/views/setup.go:200-250` - Multi-step wizard navigation pattern

  **API/Type References**:
  - `internal/restore/restore.go:ListBackupContents` - Returns `[]FileEntry` with Path, Content, Mode, ModTime
  - `internal/restore/types.go:FileEntry` - Struct definition

  **WHY Each Reference Matters**:
  - `ListBackupContents`: Function to call after password validation
  - `FileEntry`: Data structure to display in file list

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: File list loads after password
    Tool: interactive_bash (tmux)
    Preconditions: Valid backup with multiple files
    Steps:
      1. Navigate to file selection phase (correct password entered)
      2. tmux capture-pane -t dotkeeper-test -p > /tmp/file-list.txt
      3. Assert: /tmp/file-list.txt contains "[ ]" (unchecked items)
      4. Assert: /tmp/file-list.txt shows file paths
    Expected Result: File list displayed with checkboxes
    Evidence: /tmp/file-list.txt

  Scenario: Space toggles file selection
    Tool: interactive_bash (tmux)
    Preconditions: In file selection phase
    Steps:
      1. tmux capture-pane -t dotkeeper-test -p (before toggle)
      2. tmux send-keys -t dotkeeper-test Space
      3. tmux capture-pane -t dotkeeper-test -p > /tmp/after-toggle.txt
      4. Assert: /tmp/after-toggle.txt contains "[x]" (checked item)
      5. tmux send-keys -t dotkeeper-test Space
      6. tmux capture-pane -t dotkeeper-test -p > /tmp/after-untoggle.txt
      7. Assert: /tmp/after-untoggle.txt contains "[ ]" (unchecked again)
    Expected Result: Space toggles checkbox state
    Evidence: /tmp/after-toggle.txt, /tmp/after-untoggle.txt

  Scenario: Select All with 'a' key
    Tool: interactive_bash (tmux)
    Preconditions: In file selection phase, no files selected
    Steps:
      1. tmux send-keys -t dotkeeper-test "a"
      2. tmux capture-pane -t dotkeeper-test -p > /tmp/select-all.txt
      3. Assert: /tmp/select-all.txt shows all items as "[x]"
      4. Assert: Count matches "N of N files selected"
    Expected Result: All files selected
    Evidence: /tmp/select-all.txt

  Scenario: Cannot proceed with 0 files selected
    Tool: interactive_bash (tmux)
    Preconditions: In file selection, 0 files selected
    Steps:
      1. tmux send-keys -t dotkeeper-test Enter
      2. tmux capture-pane -t dotkeeper-test -p > /tmp/zero-select.txt
      3. Assert: /tmp/zero-select.txt contains "Select at least one file"
      4. Assert: Still in file selection phase (not restore)
    Expected Result: Error message, stays in file selection
    Evidence: /tmp/zero-select.txt
  ```

  **Commit**: YES
  - Message: `feat(tui): add file list with multi-select to restore view`
  - Files: `internal/tui/views/restore.go`
  - Pre-commit: `go build ./... && go test ./internal/tui/views/... -v`

---

- [x] 4. Implement diff preview

  **What to do**:
  - Add phase 4: diff preview view
  - In file selection (phase 2), 'd' key triggers diff preview for selected file
  - Call `restore.PreviewRestore(selectedBackup, password, RestoreOptions{DryRun: true, ShowDiff: true, SelectedFiles: []string{currentFile}})`
  - Display diff in viewport.Model (scrollable)
  - Handle scrolling: j/k or arrow keys to scroll diff
  - ESC returns to file selection (phase 2)
  - Show message if file doesn't exist locally (will be created)
  - Show message for binary files (cannot diff)
  - Define `type diffLoadedMsg struct { diff string; file string }`

  **Must NOT do**:
  - Don't implement actual restore yet
  - Don't auto-show diff (only on 'd' key)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Adding diff phase, uses viewport component
  - **Skills**: [`git-master`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential
  - **Blocks**: Tasks 5, 6
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `github.com/charmbracelet/bubbles/viewport` - Viewport component for scrollable content

  **API/Type References**:
  - `internal/restore/restore.go:PreviewRestore` - Returns `*RestoreResult` with DiffResults map
  - `internal/restore/types.go:RestoreOptions` - DryRun, ShowDiff, SelectedFiles fields
  - `internal/restore/diff.go:GenerateDiff` - Unified diff format

  **WHY Each Reference Matters**:
  - `PreviewRestore`: Generates diffs without modifying files
  - `viewport`: BubbleTea component for scrollable text

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: 'd' key shows diff for selected file
    Tool: interactive_bash (tmux)
    Preconditions: In file selection, file exists locally with different content
    Steps:
      1. Select a file with space key
      2. tmux send-keys -t dotkeeper-test "d"
      3. Wait 1s
      4. tmux capture-pane -t dotkeeper-test -p > /tmp/diff-view.txt
      5. Assert: /tmp/diff-view.txt contains "---" (diff header)
      6. Assert: /tmp/diff-view.txt contains "+++" (diff header)
    Expected Result: Unified diff displayed
    Evidence: /tmp/diff-view.txt

  Scenario: Diff is scrollable
    Tool: interactive_bash (tmux)
    Preconditions: In diff preview with long diff
    Steps:
      1. tmux send-keys -t dotkeeper-test "j" (scroll down)
      2. tmux capture-pane -t dotkeeper-test -p > /tmp/scroll1.txt
      3. tmux send-keys -t dotkeeper-test "k" (scroll up)
      4. tmux capture-pane -t dotkeeper-test -p > /tmp/scroll2.txt
      5. Assert: Content different between scroll1 and scroll2
    Expected Result: Viewport scrolls
    Evidence: /tmp/scroll1.txt, /tmp/scroll2.txt

  Scenario: ESC returns to file selection from diff
    Tool: interactive_bash (tmux)
    Preconditions: In diff preview phase
    Steps:
      1. tmux send-keys -t dotkeeper-test Escape
      2. tmux capture-pane -t dotkeeper-test -p > /tmp/back-from-diff.txt
      3. Assert: /tmp/back-from-diff.txt shows file list with checkboxes
    Expected Result: Returns to file selection
    Evidence: /tmp/back-from-diff.txt

  Scenario: New file shows "will be created" message
    Tool: interactive_bash (tmux)
    Preconditions: Backup contains file that doesn't exist locally
    Steps:
      1. Select the non-existing file, press 'd'
      2. tmux capture-pane -t dotkeeper-test -p > /tmp/new-file-diff.txt
      3. Assert: /tmp/new-file-diff.txt contains "will be created" or "new file"
    Expected Result: Informative message instead of diff
    Evidence: /tmp/new-file-diff.txt
  ```

  **Commit**: YES
  - Message: `feat(tui): add diff preview to restore view`
  - Files: `internal/tui/views/restore.go`
  - Pre-commit: `go build ./... && go test ./internal/tui/views/... -v`

---

- [x] 5. Implement restore execution and results

  **What to do**:
  - Add phase 3: restoring (with progress)
  - When Enter pressed in file selection with files selected: transition to phase 3
  - Show "Restoring..." message
  - Call `restore.Restore(selectedBackup, password, RestoreOptions{SelectedFiles: selectedFilesList, ProgressCallback: progressFunc})`
  - ProgressCallback updates status: "Restoring: /path/to/file..."
  - Define `type restoreProgressMsg struct{ file, action string }`
  - Define `type restoreCompleteMsg struct{ result *restore.RestoreResult }`
  - Define `type restoreErrorMsg struct{ err error }`
  - On completion: show results (files restored, .bak created, errors)
  - Show summary: "Restored 5 files. 2 .bak files created."
  - Any key returns to backup list (phase 0), clear all state
  - On error: show error message, allow retry

  **Must NOT do**:
  - Don't allow cancel during restore (atomic writes)
  - Don't modify restore package

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Final phase implementation, follows patterns
  - **Skills**: [`git-master`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential
  - **Blocks**: Task 6
  - **Blocked By**: Task 4

  **References**:

  **Pattern References**:
  - `internal/tui/views/backuplist.go:122-136` - Async backup execution with result handling

  **API/Type References**:
  - `internal/restore/restore.go:Restore` - Main restore function
  - `internal/restore/types.go:RestoreOptions` - ProgressCallback field
  - `internal/restore/types.go:RestoreResult` - RestoredFiles, BackupFiles, SkippedFiles

  **WHY Each Reference Matters**:
  - `Restore`: The function to call for actual restoration
  - `RestoreResult`: Contains all info to display in results summary

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Restore executes successfully
    Tool: interactive_bash (tmux)
    Preconditions: Files selected, in file selection phase
    Steps:
      1. tmux send-keys -t dotkeeper-test Enter
      2. Wait 3s (restore operation)
      3. tmux capture-pane -t dotkeeper-test -p > /tmp/restore-result.txt
      4. Assert: /tmp/restore-result.txt contains "Restored" or "restored"
      5. Assert: /tmp/restore-result.txt contains file count
    Expected Result: Success message with restore summary
    Evidence: /tmp/restore-result.txt

  Scenario: Progress shows during restore
    Tool: interactive_bash (tmux)
    Preconditions: Multiple files selected
    Steps:
      1. Start restore
      2. Capture pane during restore (may need quick timing)
      3. Assert: Shows "Restoring:" or progress indicator
    Expected Result: Progress visible during operation
    Evidence: Captured output

  Scenario: .bak files created for conflicts
    Tool: Bash
    Preconditions: Restore completed for existing files
    Steps:
      1. ls -la ~/.config/dotkeeper/*.bak.* 2>/dev/null || echo "no bak files"
      2. Assert: .bak files exist (if files had conflicts)
    Expected Result: Backup files created
    Evidence: ls output

  Scenario: Any key returns to backup list after result
    Tool: interactive_bash (tmux)
    Preconditions: On restore result screen
    Steps:
      1. tmux send-keys -t dotkeeper-test "Enter"
      2. tmux capture-pane -t dotkeeper-test -p > /tmp/back-to-list.txt
      3. Assert: /tmp/back-to-list.txt shows backup list
    Expected Result: Returns to initial state
    Evidence: /tmp/back-to-list.txt

  Scenario: Restore error shows message
    Tool: interactive_bash (tmux)
    Preconditions: Restore fails (e.g., permission denied)
    Steps:
      1. Attempt restore to protected directory
      2. tmux capture-pane -t dotkeeper-test -p > /tmp/restore-error.txt
      3. Assert: /tmp/restore-error.txt contains error message
    Expected Result: Error displayed, can retry
    Evidence: /tmp/restore-error.txt
  ```

  **Commit**: YES
  - Message: `feat(tui): add restore execution with progress and results`
  - Files: `internal/tui/views/restore.go`
  - Pre-commit: `go build ./... && go test ./internal/tui/views/... -v`

---

- [x] 6. Add unit tests

  **What to do**:
  - Create `internal/tui/views/restore_test.go`
  - Follow patterns from `backuplist_test.go`
  - Test cases:
    - `TestNewRestore` - Model initializes correctly
    - `TestRestoreBackupListLoad` - Backups load from directory
    - `TestRestorePasswordValidation` - Correct/incorrect password handling
    - `TestRestorePasswordRetryLimit` - 3 attempts then back to list
    - `TestRestoreFileListLoad` - Files load after password
    - `TestRestoreMultiSelect` - Space toggles, 'a' selects all, 'n' selects none
    - `TestRestoreZeroFilesBlocked` - Can't proceed with 0 files
    - `TestRestoreESCNavigation` - ESC returns to previous phase
    - `TestRestoreExecution` - Mock restore call, verify result handling
  - Use table-driven tests where appropriate
  - Create test fixtures (mock backup files) in temp directory

  **Must NOT do**:
  - Don't test the restore package itself (already has tests)
  - Don't require real encryption (mock ValidateBackup, ListBackupContents)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Writing tests following existing patterns
  - **Skills**: [`git-master`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential (final task)
  - **Blocks**: None
  - **Blocked By**: Task 5

  **References**:

  **Test References**:
  - `internal/tui/views/backuplist_test.go` - Complete test pattern to follow
  - `internal/tui/views/setup_test.go` - Multi-step wizard test patterns

  **WHY Each Reference Matters**:
  - `backuplist_test.go`: Shows how to test BubbleTea models (send messages, assert state)
  - Shows fixture creation patterns with temp directories

  **Acceptance Criteria**:

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: All tests pass
    Tool: Bash
    Preconditions: restore_test.go created
    Steps:
      1. cd /home/diogo/dev/backup-dotfiles
      2. go test ./internal/tui/views/... -v -run TestRestore
    Expected Result: Exit code 0, all tests pass
    Evidence: Test output captured

  Scenario: Test coverage for restore.go
    Tool: Bash
    Preconditions: Tests implemented
    Steps:
      1. go test ./internal/tui/views/... -coverprofile=coverage.out
      2. go tool cover -func=coverage.out | grep restore.go
    Expected Result: Coverage > 60% for restore.go
    Evidence: Coverage report
  ```

  **Commit**: YES
  - Message: `test(tui): add unit tests for restore view`
  - Files: `internal/tui/views/restore_test.go`
  - Pre-commit: `go test ./internal/tui/views/... -v`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `feat(tui): add RestoreModel struct and backup list loading` | internal/tui/views/restore.go | `go build ./...` |
| 2 | `feat(tui): add password entry and validation to restore view` | internal/tui/views/restore.go | `go build ./...` |
| 3 | `feat(tui): add file list with multi-select to restore view` | internal/tui/views/restore.go | `go build ./...` |
| 4 | `feat(tui): add diff preview to restore view` | internal/tui/views/restore.go | `go build ./...` |
| 5 | `feat(tui): add restore execution with progress and results` | internal/tui/views/restore.go | `go build ./...` |
| 6 | `test(tui): add unit tests for restore view` | internal/tui/views/restore_test.go | `go test ./internal/tui/views/... -v` |

---

## Success Criteria

### Verification Commands
```bash
# Build succeeds
go build ./...
# Expected: exit 0

# All tests pass
go test ./internal/tui/views/... -v
# Expected: PASS for all test functions

# TUI launches and shows restore view
./bin/dotkeeper  # Tab to RestoreView
# Expected: Backup list visible, can select and restore
```

### Final Checklist
- [x] All "Must Have" features implemented
- [x] All "Must NOT Have" guardrails respected
- [x] All 6 tasks completed with commits
- [x] All unit tests pass
- [x] TUI restore flow works end-to-end

**PLAN COMPLETE** - Finalized 2026-02-05
