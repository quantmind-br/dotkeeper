# Code Quality & Refactoring Analysis Report

**Generated:** 2026-02-06
**Branch:** feat/settings-inline-actions
**Commit:** a0cc9df
**Total Files Analyzed:** 55 source files (excluding tests), 100 total Go files
**Total Lines:** ~32,700
**Reviewed:** 2026-02-06 (critical analysis applied, 1 item discarded, 4 adjusted, 2 merged)

---

## Executive Summary

The dotkeeper codebase is **generally well-structured** for a Go+BubbleTea project. Code is well-organized into packages with clear responsibilities, crypto is implemented correctly (AES-256-GCM + Argon2id), and the project follows atomic-write patterns consistently.

**Key areas needing attention:**
1. **3 oversized TUI view files** (settings.go 857 lines, restore.go 721 lines, setup.go 578 lines) with long functions that should be decomposed
2. **4 duplicate size-formatting implementations** across CLI and TUI layers
3. **11 repetitive type assertions** in the TUI update loop that could use a generic helper
4. **3 dead message types** defined but never used
5. **2 dead components** (`FileSelector`, `DiffViewer`) defined but never integrated

The codebase has **zero** `any` types, **zero** `@ts-ignore` equivalents, and **no globals** in the TUI layer — all strong signals of disciplined engineering.

---

## Phase 1: Quick Wins (< 1 hour)

High-value, trivial-effort items. Do these first.

### CQ-001: Size Formatting Duplicated 4 Times

**Category:** duplication
**Severity:** major

**Affected Files:**
- `internal/pathutil/scanner.go` (lines 105-121) — `FormatSize()` (exported, canonical)
- `internal/cli/list.go` (lines 170-183) — `formatSize()` (private duplicate)
- `internal/cli/history.go` (line 83) — calls `formatSize()` from list.go
- `internal/tui/views/logs.go` (lines 55-66) — `formatBytes()` (private duplicate)

**Current State:**
Three separate implementations of byte-to-human-readable conversion. `pathutil.FormatSize()` uses named constants (KB, MB, GB), while `cli/list.go:formatSize()` and `views/logs.go:formatBytes()` use an identical loop-based algorithm with `"KMGTPE"[exp]`. All produce **identical output**.

**Callers of each:**
- `pathutil.FormatSize()`: dashboard.go:128, settings.go:629, scanner.go:142/144
- `cli.formatSize()`: list.go:152/155, history.go:83
- `views.formatBytes()`: logs.go:44, setup.go:436/465

**Change:**
Remove all private implementations. Use `pathutil.FormatSize()` everywhere — it's already exported and used by `dashboard.go` and `settings.go`.

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** trivial

---

### CQ-005: Dead Message Types in messages.go

**Category:** dead_code
**Severity:** major

**Affected Files:**
- `internal/tui/views/messages.go` (lines 10-29)

**Current State:**
Three message types are defined but never instantiated anywhere in the codebase — zero instantiations, zero switch cases, zero references:

```go
type SuccessMsg struct { Source string; Message string }  // NEVER USED
type LoadingMsg struct { Source string; Message string }   // NEVER USED
type RefreshMsg struct { Source string }                   // NEVER USED
```

Each view manages its own success/loading/refresh patterns with local types instead.

**Change:**
Delete the three unused types.

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** trivial

---

### CQ-007: Dead Error Message Type Aliases (3 of 5)

**Category:** dead_code
**Severity:** minor

**Affected Files:**
- `internal/tui/views/backuplist.go` (line 31)
- `internal/tui/views/restore.go` (lines 72, 79)

**Current State:**
Three type aliases that resolve to `ErrorMsg` are defined but never appear in any switch case or type assertion:

```go
type backupDeleteErrorMsg = ErrorMsg // backuplist.go:31 — NEVER USED IN SWITCH
type diffErrorMsg = ErrorMsg         // restore.go:72 — NEVER USED IN SWITCH
type restoreErrorMsg = ErrorMsg      // restore.go:79 — NEVER USED IN SWITCH
```

Two other aliases (`BackupErrorMsg`, `passwordInvalidMsg`) ARE used in type switches and provide semantic clarity — keep those.

**Change:**
Remove only the 3 dead aliases. Keep `BackupErrorMsg` and `passwordInvalidMsg` which are used in switch statements for semantic routing.

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** trivial

---

### CQ-010: Magic Number for Password Attempts

**Category:** code_smells
**Severity:** minor

**Affected Files:**
- `internal/tui/views/restore.go` (line 438)

**Current State:**
Magic `3` appears twice — in the condition and the error message format string:
```go
if m.passwordAttempts >= 3 {
    // ...
    m.restoreError = fmt.Sprintf("Invalid password (attempt %d/3): %v", ...)
```

**Change:**
```go
const maxPasswordAttempts = 3

if m.passwordAttempts >= maxPasswordAttempts {
    // ...
    m.restoreError = fmt.Sprintf("Invalid password (attempt %d/%d): %v", m.passwordAttempts, maxPasswordAttempts, ...)
```

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** trivial

---

### CQ-011: Manual String Trimming Instead of stdlib

**Category:** code_smells
**Severity:** minor

**Affected Files:**
- `internal/cli/backup.go` (lines 115-118)

**Current State:**
```go
password := string(data)
if len(password) > 0 && password[len(password)-1] == '\n' {
    password = password[:len(password)-1]
}
```

**Change:**
```go
password := strings.TrimSuffix(string(data), "\n")
```

> **Note:** Use `strings.TrimSuffix` (removes exactly one trailing `\n`), NOT `strings.TrimRight` (would strip ALL trailing newlines, changing behavior for edge cases).

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** trivial

---

## Phase 2: Structural Improvements (2-3 hours)

Medium-effort items that improve code organization.

### CQ-004: Type Assertion Boilerplate in Update Loop (11 repetitions)

**Category:** code_smells
**Severity:** major

**Affected Files:**
- `internal/tui/update.go` (lines 31-59, 111-116, 249-285)

**Current State:**
The same type-assertion-after-Update pattern is repeated **11 times** (5 in `propagateWindowSize()` + 6 in `Update()` including the setup model):

```go
// This exact pattern appears 11 times:
tm, cmd = m.dashboard.Update(viewMsg)
if d, ok := tm.(views.DashboardModel); ok {
    m.dashboard = d
}
cmds = append(cmds, cmd)
```

**Change:**
With Go 1.25's generics, create a helper:

```go
func updateView[T tea.Model](view T, msg tea.Msg) (T, tea.Cmd) {
    model, cmd := view.Update(msg)
    if v, ok := model.(T); ok {
        return v, cmd
    }
    return view, cmd
}

// Usage (reduces each 5-line block to 2 lines):
m.dashboard, cmd = updateView(m.dashboard, viewMsg)
cmds = append(cmds, cmd)
```

This reduces ~55 lines to ~22 lines — a 60% reduction in boilerplate.

**Breaking Change:** No
**Prerequisites:** Go 1.18+ (already on 1.25)
**Estimated Effort:** small

---

### CQ-003: Restore View Too Large (721 lines)

**Category:** large_files
**Severity:** major

**Affected Files:**
- `internal/tui/views/restore.go` (721 lines)

**Current State:**
The restore view manages a 6-phase workflow (backup selection -> password -> file selection -> restoring -> diff preview -> results) in a single file. The `View()` method uses 6 sequential if-blocks instead of a switch statement.

**Change:**
Extract `View()` rendering into phase-specific render methods and convert to switch:

```go
// Current: 6 sequential if-blocks
func (m RestoreModel) View() string {
    if m.phase == phaseBackupList { /* ... */ return s.String() }
    if m.phase == phasePassword { /* ... */ return s.String() }
    // ... 4 more
}

// Proposed: switch + extracted render methods
func (m RestoreModel) View() string {
    switch m.phase {
    case phaseBackupList:  return m.renderBackupList()
    case phasePassword:    return m.renderPassword()
    case phaseFileSelect:  return m.renderFileSelect()
    case phaseRestoring:   return m.renderRestoring()
    case phaseDiffPreview: return m.renderDiffPreview()
    case phaseResults:     return m.renderResults()
    }
}
```

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** small

---

### CQ-008: Deep Nesting in Path Toggle Logic

**Category:** complexity
**Severity:** minor

**Affected Files:**
- `internal/tui/views/settings.go` (lines 374-393)

**Current State:**
The space-key handler for toggling disabled paths has 4 nesting levels (switch -> if -> for -> if).

**Change:**
Extract into a dedicated method:

```go
func (m *SettingsModel) togglePathDisabled(lt pathListType, path string) {
    disabled := m.disabledPathsForType(lt)
    for i, d := range disabled {
        if d == path {
            m.setDisabledPathsForType(lt, append(disabled[:i], disabled[i+1:]...))
            return
        }
    }
    m.setDisabledPathsForType(lt, append(disabled, path))
}
```

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** trivial

---

### CQ-014: CLI History Logging Pattern Repeated

**Category:** duplication
**Severity:** minor

**Affected Files:**
- `internal/cli/backup.go` (lines 74-79, 96-100)
- `internal/cli/restore.go` (lines 91-97, 115-121, 128-132)

**Current State:**
The same "best-effort history logging" pattern appears 5 times:

```go
if storeErr == nil {
    store.Append(history.EntryFrom...(result))
} else {
    fmt.Fprintf(os.Stderr, "Warning: failed to log history: %v\n", storeErr)
}
```

**Change:**
Extract helper:

```go
func logHistory(store *history.Store, storeErr error, entry history.HistoryEntry) {
    if storeErr != nil {
        fmt.Fprintf(os.Stderr, "Warning: history unavailable: %v\n", storeErr)
        return
    }
    if err := store.Append(entry); err != nil {
        fmt.Fprintf(os.Stderr, "Warning: failed to log history: %v\n", err)
    }
}
```

**Breaking Change:** No
**Estimated Effort:** trivial

---

## Phase 3: Larger Refactors (5-7 hours)

Higher-effort items. Do after Phases 1-2 are complete.

### CQ-002: Settings View Too Large (857 lines)

**Category:** large_files
**Severity:** major

**Affected Files:**
- `internal/tui/views/settings.go` (857 lines)

**Current State:**
The settings view handles 6 distinct states (list navigation, field editing, file/folder browsing, sub-item editing, file picker) with all rendering logic, input handling, and path management in one file. The `Update()` method is 92 lines and the `View()` method is 60 lines.

**Change:**
Split into focused files following existing project conventions:

| New File | Contents | ~Lines |
|----------|----------|--------|
| `settings.go` | Model struct, `NewSettings()`, `Update()`, `Init()` | ~200 |
| `settings_editing.go` | `handleEditingFieldInput()`, `handleEditingSubItemInput()`, `startEditingField()`, `startEditingSubItem()`, `saveFieldValue()` | ~200 |
| `settings_paths.go` | `handleBrowsingPathsInput()`, path type helpers, `refreshPathList()`, `togglePathDisabled()` | ~200 |
| `settings_view.go` | `View()`, `HelpBindings()`, `StatusHelpText()`, rendering helpers | ~150 |

**Breaking Change:** No (all types stay in `views` package)
**Prerequisites:** CQ-008 (togglePathDisabled extraction makes the split cleaner)
**Estimated Effort:** medium

---

### CQ-006: Dead Components — Wire into Restore or Delete

**Category:** dead_code + structure
**Severity:** major

**Affected Files:**
- `internal/tui/views/restore.go`
- `internal/tui/views/restore_fileselect.go` (147 lines) — `FileSelector` component
- `internal/tui/views/restore_diff.go` (150 lines) — `DiffViewer` component

**Current State:**
Both components are fully implemented but **never referenced** by any file outside their own. The restore view duplicates their logic inline:

| Inline in restore.go | Exists in component |
|---|---|
| `selectedFiles map[string]bool` | `FileSelector.selected` |
| `updateFileListSelection()` | `FileSelector.updateFileListSelection()` |
| `countSelectedFiles()` | `FileSelector.SelectedCount()` |
| `getSelectedFilePaths()` | `FileSelector.SelectedFiles()` |
| `viewport` + `currentDiff` + `diffFile` | `DiffViewer` struct |

These appear to be extraction attempts that were never wired in, leaving ~300 lines of dead code.

**Change (recommended):**
Wire components into `RestoreModel` via composition:

```go
type RestoreModel struct {
    ctx            *ProgramContext
    backupList     list.Model
    phase          restorePhase
    selectedBackup string
    password       string
    passwordInput  textinput.Model
    fileSelector   FileSelector   // Replace inline file selection
    diffViewer     DiffViewer     // Replace inline diff viewing
    // ...
}
```

This reduces `restore.go` by ~100-150 lines and eliminates the dead code files.

**Fallback:** If wiring proves impractical, delete both files entirely.

**Breaking Change:** No
**Prerequisites:** CQ-003 (render method extraction makes wiring easier)
**Estimated Effort:** medium

---

## Suggestions (do when convenient)

### CQ-013: Setup View has Duplicated Rendering Pattern

**Category:** duplication

**Affected Files:**
- `internal/tui/views/setup.go` (lines 418-473)

**Current State:**
`StepPresetFiles` and `StepPresetFolders` View() code is nearly identical — same cursor/checkbox rendering, same list iteration, just different data source.

**Change:**
Extract shared preset rendering:

```go
func renderPresetList(presets []pathutil.DotfilePreset, cursor int, st styles.Styles) string {
    var s strings.Builder
    for i, p := range presets {
        // shared rendering logic
    }
    return s.String()
}
```

**Breaking Change:** No
**Estimated Effort:** trivial

---

### CQ-012: `DefaultStyles()` Called in Every View Render

**Category:** code_smells

**Affected Files:**
- 6 View() methods + 3 framework/helper files (9 total call sites)

**Current State:**
`styles.DefaultStyles()` is called on every `View()` invocation (every render frame). Currently it returns a package-level `var defaultStyles` so it's **cheap** — this is not a performance problem today.

**Change (only if theme switching is planned):**
Store styles reference in `ProgramContext`:

```go
type ProgramContext struct {
    Config *config.Config
    Store  *history.Store
    Styles styles.Styles  // Add styles to shared context
    Width  int
    Height int
}
```

> **Note:** This is a YAGNI candidate. Only worth doing if theme switching is actually on the roadmap. Current implementation has zero measurable performance impact.

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** small

---

## Discarded

### ~~CQ-009: Complex Conditional in Diff Algorithm~~ — DISCARDED

**Reason:** Over-engineering. The 5-condition check (`i < len(a) && j < len(b) && lcsIdx < len(lcs) && a[i] == lcs[lcsIdx] && b[j] == lcs[lcsIdx]`) is a standard LCS algorithm bounds-check pattern. Every diff implementation has this. Extracting it to a named function:
- Adds a function call in a potentially hot loop
- Obscures the algorithm for developers familiar with LCS/diff implementations
- The function name (`isLCSMatch`) doesn't add information the surrounding loop context doesn't already provide

---

## Code Metrics

| Metric | Value | Status |
|--------|-------|--------|
| Files > 500 lines (source) | 3 (settings 857, restore 721, setup 578) | Needs attention |
| Files > 500 lines (tests) | 5 (acceptable for test files) | Good |
| Functions > 50 lines | 7 | Needs attention |
| `any` types | 0 | Good |
| Duplicated formatting functions | 4 | Needs attention |
| Dead message types | 3 | Needs attention |
| Dead component files | 2 (~300 lines) | Needs attention |
| Dead error type aliases | 3 | Needs attention |
| `DefaultStyles()` calls per frame | 9 | Acceptable (cheap) |
| Type assertion repetitions | 11 | Needs attention |

## Summary

| Severity | Count |
|----------|-------|
| Critical | 0 |
| Major | 6 |
| Minor | 4 |
| Suggestion | 2 |
| Discarded | 1 |

**Total Actionable Issues:** 12 (down from 15)

| Category | Count |
|----------|-------|
| Dead Code | 3 (CQ-005, CQ-006, CQ-007) |
| Duplication | 3 (CQ-001, CQ-013, CQ-014) |
| Large Files | 2 (CQ-002, CQ-003) |
| Code Smells | 3 (CQ-004, CQ-010, CQ-011) |
| Complexity | 1 (CQ-008) |
| Suggestion | 1 (CQ-012) |
