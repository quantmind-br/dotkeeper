# Code Quality & Refactoring Plan

**Generated:** 2026-02-06
**Branch:** master
**Analyzed:** 77 Go files (~12,700 lines of production + test code)
**Reviewed:** 2026-02-06 (critical analysis applied — 1 discarded, 2 adjusted, 16 kept)

---

## Executive Summary

The dotkeeper codebase is well-structured — clean package boundaries, consistent error wrapping, good test coverage. Issues concentrate in the TUI layer: code duplication and file size. Business logic packages (`backup/`, `restore/`, `crypto/`, `config/`) are clean.

**One functional bug found**: CQ-015 (`Validate()` requires `GitRemote` but setup allows it empty — produces broken configs).

---

## Phase 1 — Quick Wins (trivial effort, high impact)

### CQ-015: Validate() GitRemote Inconsistency [BUG]

**Category:** code_smells → **functional bug**
**Severity:** critical (upgraded from minor)

**Affected Files:**
- `internal/config/config.go` (lines 105-119)

**Current State:**
`Validate()` requires `GitRemote` to be non-empty, but the setup wizard allows it to be empty (Step 2 says "optional"). Configs created by setup may fail validation.

```go
func (c *Config) Validate() error {
    if c.GitRemote == "" {
        return fmt.Errorf("git_remote is required") // But setup allows empty!
    }
}
```

**Fix:** Make `GitRemote` optional in `Validate()`. Git operations should check for empty remote at call time, not at config validation time.

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** trivial

---

### CQ-008: Dead Code — `getFieldValue` in cli/config.go

**Category:** dead_code
**Severity:** minor

**Affected Files:**
- `internal/cli/config.go` (lines 207-214)

**Fix:** Remove `getFieldValue` function and unused `reflect` import.

**Estimated Effort:** trivial

---

### CQ-004: Magic Numbers for Restore Phases

**Category:** code_smells
**Severity:** major

**Affected Files:**
- `internal/tui/views/restore.go` (entire file, ~25 occurrences)

**Fix:** Replace raw integers with named constants:

```go
type restorePhase int

const (
    phaseBackupList  restorePhase = iota
    phasePassword
    phaseFileSelect
    phaseRestoring
    phaseDiffPreview
    phaseResults
)
```

**Estimated Effort:** small (trivial concept, but ~25 replacements)

---

### CQ-014: Variable Shadowing of `styles` Package

**Category:** naming
**Severity:** minor

**Affected Files:**
- `internal/tui/views/settings.go` (line 765)
- `internal/tui/views/setup.go` (line 343)

**Fix:** Rename local `styles` variable to `st` (which some views already use).

**Estimated Effort:** trivial

---

### CQ-011: HelpProvider Interface Cleanup

**Category:** code_smells
**Severity:** minor

**Affected Files:**
- `internal/tui/view.go` (lines 23-43)

**Fix:** Replace unnecessary `interface{}()` conversion with direct switch:

```go
func (m Model) currentViewHelp() []views.HelpEntry {
    switch m.state {
    case DashboardView:
        return m.dashboard.HelpBindings()
    case BackupListView:
        return m.backupList.HelpBindings()
    // ...
    }
    return nil
}
```

**Estimated Effort:** trivial

---

### CQ-019: `LoadOrDefault()` Uses `os.Getenv("HOME")`

**Category:** code_smells
**Severity:** minor

**Affected Files:**
- `internal/config/config.go` (line 163)

**Fix:** Replace `os.Getenv("HOME")` with `os.UserHomeDir()` for consistency with `GetConfigDir()` and proper error handling.

**Estimated Effort:** trivial

---

### CQ-017: Replace `filepath.Walk` with `filepath.WalkDir`

**Category:** performance
**Severity:** suggestion

**Affected Files:**
- `internal/pathutil/scanner.go` (lines 67, 129)

**Fix:** Replace `filepath.Walk` with `filepath.WalkDir` (Go 1.16+). Avoids redundant `os.Lstat` calls on every file.

**Estimated Effort:** trivial

---

### CQ-006: Password Input Factory

**Category:** duplication
**Severity:** minor

**Affected Files:**
- `internal/tui/views/backuplist.go` (lines 62-68)
- `internal/tui/views/restore.go` (lines 100-106)

**Fix:** Create `NewPasswordInput(placeholder string) textinput.Model` in `components/`.

**Estimated Effort:** trivial

---

### CQ-007: PathCompleter Styling

**Category:** duplication
**Severity:** minor

**Affected Files:**
- `internal/tui/views/settings.go` (lines 83-85)
- `internal/tui/views/setup.go` (lines 70-72)

**Fix:** Move cursor/prompt styling into `NewPathCompleter()` constructor.

**Estimated Effort:** trivial

---

## Phase 2 — Small Refactors (low effort, solid value)

### CQ-012: Styles Re-created on Every Render

**Category:** performance
**Severity:** major (performance hot path)

**Affected Files:**
- `internal/tui/styles/styles.go` (`DefaultStyles()`)
- Every `View()` method

**Current State:**
`DefaultStyles()` creates 40+ `lipgloss.NewStyle()` objects on every call. BubbleTea re-renders on every keystroke, creating significant allocation pressure.

**Fix:** Cache as package-level var (lipgloss styles are immutable after creation):

```go
var defaultStyles = Styles{
    // ... all style definitions
}

func DefaultStyles() Styles {
    return defaultStyles
}
```

**Estimated Effort:** small

---

### CQ-002: Duplicated Backup List Refresh Logic

**Category:** duplication
**Severity:** major

**Affected Files:**
- `internal/tui/views/backuplist.go` (lines 82-104)
- `internal/tui/views/restore.go` (lines 138-160)

**Current State:**
100% copy-paste of 23 lines. Both scan backup directory, reverse sort, stat files, build item list.

**Fix:** Extract `LoadBackupItems(backupDir string) []list.Item` in `internal/tui/views/helpers.go`. Also deduplicate the `backupItem` type (currently defined in both files).

**Estimated Effort:** small

---

### CQ-003: Duplicated Tab/ShiftTab/Number-Key Navigation

**Category:** duplication
**Severity:** major

**Affected Files:**
- `internal/tui/update.go` (lines 136-203)

**Current State:**
Same 4-way refresh chain duplicated 3x (Tab, ShiftTab, number keys = ~60 lines of duplication).

**Fix:** Extract `refreshCmdForState()` and `switchToView()` helper:

```go
func (m *Model) refreshCmdForState(state ViewState) tea.Cmd {
    switch state {
    case DashboardView:
        return m.dashboard.Refresh()
    case BackupListView:
        return m.backupList.Refresh()
    case RestoreView:
        return m.restore.Refresh()
    case LogsView:
        return m.logs.LoadHistory()
    default:
        return nil
    }
}
```

**Estimated Effort:** small

---

### CQ-013: History Store Created Multiple Times

**Category:** code_smells
**Severity:** minor

**Affected Files:**
- `internal/cli/backup.go` (2 instantiations)
- `internal/cli/restore.go` (3 instantiations)

**Fix:** Create `history.NewStore()` once at function start, reuse throughout.

**Estimated Effort:** small

---

### CQ-005: List Initialization Factory

**Category:** duplication
**Severity:** minor

**Affected Files:**
- `internal/tui/views/settings.go` (3 blocks), `backuplist.go`, `restore.go`, `logs.go`

**Fix:** Create a simple `NewMinimalList()` with common defaults (no title, no help, no filtering). Skip functional options pattern — it's overkill for this. Callers set any additional options directly.

```go
func NewMinimalList() list.Model {
    l := list.New([]list.Item{}, NewListDelegate(), 0, 0)
    l.SetShowTitle(false)
    l.SetShowHelp(false)
    l.SetFilteringEnabled(false)
    return l
}
```

**Estimated Effort:** small

---

### CQ-009: Discarded Commands in `propagateWindowSize`

**Category:** code_smells
**Severity:** minor

**Affected Files:**
- `internal/tui/update.go` (lines 41-67)

**Fix:** Add a comment documenting intent rather than refactoring the function. The current pattern is standard BubbleTea — window resize never produces async work in these views. Over-engineering a fix for a hypothetical future issue adds noise.

```go
// Commands intentionally discarded — window resize produces no async work in our views.
tm, _ = m.dashboard.Update(viewMsg)
```

**Estimated Effort:** trivial

---

## Phase 3 — Medium Refactors (meaningful effort, high value)

### CQ-001: Settings View File/Folder Handler Deduplication

**Category:** large_files / duplication
**Severity:** major

**Affected Files:**
- `internal/tui/views/settings.go` (866 lines)

**Current State:**
`handleBrowsingFilesInput()` and `handleBrowsingFoldersInput()` are ~95% identical (~80 lines each). `refreshFilesList()` and `refreshFoldersList()` are ~98% identical (~30 lines each). Total: ~160 lines of near-duplicate code.

**Fix:** Extract generic handler with `pathListType` parameter:

```go
type pathListType int

const (
    pathListFiles   pathListType = iota
    pathListFolders
)

func (m SettingsModel) handleBrowsingPathsInput(msg tea.KeyMsg, listType pathListType) (tea.Model, tea.Cmd) {
    paths := m.pathsForType(listType)
    disabledPaths := m.disabledPathsForType(listType)
    // ... shared logic
}

func (m *SettingsModel) refreshPathList(listType pathListType) {
    // ... shared logic
}
```

**Estimated Effort:** medium

---

### CQ-010: Restore Update() Complexity

**Category:** complexity
**Severity:** minor

**Affected Files:**
- `internal/tui/views/restore.go` (lines 250-464, ~215 lines)

**Current State:**
Single method handles all 6 phases with nested if-else + switch. ~215 lines, ~8 levels of branching.

**Fix:** Extract phase-specific handlers (follows settings.go pattern):

```go
func (m RestoreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch m.phase {
        case phaseBackupList:
            return m.handleBackupListInput(msg)
        case phasePassword:
            return m.handlePasswordInput(msg)
        // ...
        }
    }
}
```

**Prerequisites:** CQ-004 (named phase constants)
**Estimated Effort:** medium

---

### CQ-018: Setup handleEnter() Step Handlers

**Category:** complexity
**Severity:** suggestion

**Affected Files:**
- `internal/tui/views/setup.go` (lines 209-331, 122 lines)

**Fix:** Extract only the larger cases (>15 lines) into separate methods. Cases that are 5-8 lines (StepWelcome, StepConfirm) should stay inline. Expect 3-4 extractions, not all 8.

**Estimated Effort:** small

---

## Discarded

### ~~CQ-016: ViewModel Interface~~ — REMOVED

**Reason:** Adding abstraction to save a few switch cases is a net negative. BubbleTea views return concrete types (`tea.Model`), requiring type assertions after `Update()` regardless. The current switch statements are clear, explicit, debuggable, and follow standard BubbleTea patterns. The main model's dispatch logic is already small (~20 lines). An interface would add indirection without meaningful simplification.

---

## Code Metrics

| Metric | Value | Status |
|--------|-------|--------|
| Files > 500 lines (non-test) | 3 (settings: 866, restore: 610, setup: 524) | Needs attention |
| Functions > 50 lines | 5 (Update methods, handleEnter, View) | Acceptable with extraction |
| Duplicated blocks | 7 significant instances | Needs attention |
| Dead code | 1 function + 1 import | Minor |
| Magic numbers | 1 file (restore phases) | Needs attention |
| `_ =` error discards (non-test) | 12 | Mostly acceptable |
| Total packages | 12 | Good |
| Test coverage | Good (co-located tests) | Good |

## Summary

| Severity | Count |
|----------|-------|
| Critical (bug) | 1 |
| Major | 5 |
| Minor | 8 |
| Suggestion | 3 |
| Discarded | 1 |

**Remaining: 17 items** (down from 19 — 1 discarded, 1 adjusted to comment-only)

## Implementation Order

| # | ID | Effort | Phase | Description |
|---|-----|--------|-------|-------------|
| 1 | CQ-015 | trivial | 1 | **[BUG]** Fix GitRemote validation |
| 2 | CQ-008 | trivial | 1 | Remove dead code |
| 3 | CQ-004 | small | 1 | Named restore phase constants |
| 4 | CQ-014 | trivial | 1 | Fix styles variable shadowing |
| 5 | CQ-011 | trivial | 1 | HelpProvider cleanup |
| 6 | CQ-019 | trivial | 1 | os.UserHomeDir() consistency |
| 7 | CQ-017 | trivial | 1 | filepath.WalkDir upgrade |
| 8 | CQ-006 | trivial | 1 | Password input factory |
| 9 | CQ-007 | trivial | 1 | PathCompleter default styling |
| 10 | CQ-012 | small | 2 | Styles singleton (perf) |
| 11 | CQ-002 | small | 2 | Shared backup list loading |
| 12 | CQ-003 | small | 2 | switchToView() helper |
| 13 | CQ-013 | small | 2 | History store single init |
| 14 | CQ-005 | small | 2 | NewMinimalList() factory |
| 15 | CQ-009 | trivial | 2 | Document propagateWindowSize intent |
| 16 | CQ-001 | medium | 3 | Settings file/folder dedup |
| 17 | CQ-010 | medium | 3 | Restore phase handler extraction |
| 18 | CQ-018 | small | 3 | Setup handleEnter() partial extraction |
