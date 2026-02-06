# TUI BubbleTea Improvements — Complete Refactoring + Features

## TL;DR

> **Quick Summary**: Comprehensive TUI refactoring applying BubbleTea best practices: safety-net tests for untested framework layer, ProgramContext pattern to eliminate prop drilling, consolidated KeyMap/Help/View interfaces, loading spinners, AdaptiveColor theme support, restore.go sub-component extraction, and new features (toast notifications, auto-refresh, mouse support, search/filter).
> 
> **Deliverables**:
> - Framework-level test suite (model.go, update.go, view.go, help.go — currently 0% coverage)
> - ProgramContext shared state pattern across all views
> - Consolidated KeyMap system using `bubbles/key`
> - Loading spinners for async operations
> - AdaptiveColor theme support (light/dark terminals)
> - View interface standardization
> - Help system migration to `bubbles/help`
> - Message type consolidation
> - WindowSizeMsg propagation bug fix
> - Restore.go sub-component extraction (~50% line reduction)
> - New features: toast notifications, auto-refresh, mouse support, search/filter
> 
> **Estimated Effort**: XL
> **Parallel Execution**: YES - 5 waves
> **Critical Path**: Task 1 → Task 2 → Task 3 → Tasks 4-8 (parallel) → Tasks 9-12 (parallel) → Tasks 13-16 (parallel) → Task 17

---

## Context

### Original Request
Implement all TUI improvement suggestions from a comprehensive BubbleTea analysis: 8 core improvements, restore.go refactoring into sub-components, and future features (toast notifications, auto-refresh, mouse support, search/filter in lists).

### Interview Summary
**Key Discussions**:
- **Scope**: Full — all 8 improvements, restore extraction, AND future features
- **Test Strategy**: TDD — write framework safety-net tests first, then refactor
- **ProgramContext**: Accepted — refactor all view constructors to use shared context

**Research Findings**:
- Framework layer (model.go, update.go, view.go, help.go) has **ZERO tests** — 564 lines untested
- Views have test files but actual coverage is ~35-60%, not true 100%
- 11 unchecked type assertions in update.go — runtime panic risk on refactoring mistakes
- `backupsLoadedMsg` is shared between BackupList and Restore views via helpers.go
- `SetupModel` is asymmetric — takes zero constructor params, no StatusHelpText()
- `DefaultStyles()` called on every render cycle — must remain cheap
- `LogsModel` uses `LoadHistory()` instead of `Refresh()` — naming inconsistency
- Dashboard help text is hardcoded in view.go:80, not in DashboardModel
- `var keys = DefaultKeyMap()` is a package-level global in update.go:39

### Metis Review
**Identified Gaps** (addressed):
- **SetupModel asymmetry**: SetupModel will receive ProgramContext with nil Config (special case documented)
- **View interface scope creep**: Interface will be opt-in using Go type assertions, not mandatory
- **Message routing order**: Framework-level handlers BEFORE state-specific routing — order preserved
- **Style performance**: AdaptiveColor detection cached at startup, not per-render
- **backupsLoadedMsg sharing**: Routing behavior preserved — consumed by active view only
- **LoadHistory() naming**: Aliased as `Refresh()` while keeping original method
- **Restore sub-component location**: Stay in `views/` package, NOT moved to `components/`
- **Future feature isolation**: No future-feature fields added to ProgramContext or View interface

---

## Work Objectives

### Core Objective
Transform the DotKeeper TUI from a solid-but-ad-hoc BubbleTea application into a well-tested, pattern-compliant implementation following Charm ecosystem best practices, then extend with new user-facing features.

### Concrete Deliverables
- `internal/tui/tui_test.go` — Framework-level test suite
- `internal/tui/context.go` — ProgramContext struct
- `internal/tui/keys.go` — Consolidated KeyMap with `bubbles/key`
- Refactored `internal/tui/model.go`, `update.go`, `view.go`, `help.go`
- Refactored all 5 view constructors
- `internal/tui/views/restore_password.go`, `restore_fileselect.go`, `restore_diff.go`
- Updated `internal/tui/styles/styles.go` with AdaptiveColor
- New features: toast component, auto-refresh, mouse zones, search overlay

### Definition of Done
- [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS, 0 race conditions
- [ ] `go build ./cmd/dotkeeper/` → builds successfully
- [ ] `go vet ./...` → no issues
- [ ] Framework test coverage ≥10 new tests
- [ ] Zero package-level globals for keys or styles in TUI
- [ ] All views using ProgramContext
- [ ] TUI launches and all 5 views render correctly

### Must Have
- TDD: Safety-net tests BEFORE any refactoring
- Backward compatibility: All existing features work identically
- Race-free: `-race` flag passes on all tests
- No blocking in Update(): All I/O via `tea.Cmd`

### Must NOT Have (Guardrails)
- **DO NOT** change message routing order in update.go (framework handlers BEFORE state routing)
- **DO NOT** move restore sub-components to `components/` — keep in `views/` package
- **DO NOT** add future-feature fields to ProgramContext or View interface prematurely
- **DO NOT** replace `LoadHistory()` — alias it as `Refresh()` while keeping the original
- **DO NOT** make `DefaultStyles()` expensive — AdaptiveColor detection cached, not per-render
- **DO NOT** use unchecked type assertions — always use `value, ok` pattern
- **DO NOT** change the implicit `backupsLoadedMsg` routing behavior (active view only)
- **DO NOT** break the 27 resize tests in `resize_test.go` — run after EVERY task
- **DO NOT** add external test dependencies (testify, teatest) — use Go stdlib
- **DO NOT** block in `Update()` — async via `tea.Cmd` closures only

---

## Verification Strategy

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> ALL tasks in this plan MUST be verifiable WITHOUT any human action.

### Test Decision
- **Infrastructure exists**: YES — Go stdlib `testing` package
- **Automated tests**: TDD (write tests first, then refactor)
- **Framework**: `go test` with `-race` flag
- **Command**: `make test` → `go test -v -race -coverprofile=coverage.out ./...`

### TDD Workflow per Task

**Task Structure:**
1. **RED**: Write failing test first
   - Test file location specified per task
   - Run: `go test ./internal/tui/... -run TestName -v`
   - Expected: FAIL (test exists, implementation doesn't match)
2. **GREEN**: Implement minimum code to pass
   - Run: `go test ./internal/tui/... -run TestName -v`
   - Expected: PASS
3. **REFACTOR**: Clean up while keeping green
   - Run: `go test ./internal/tui/... -race -count=1`
   - Expected: ALL PASS (no regressions)

### Regression Gate (MANDATORY after EVERY task)

```bash
go test ./internal/tui/... -race -count=1 && go build ./cmd/dotkeeper/
```

### Agent-Executed QA Scenarios

| Type | Tool | How Agent Verifies |
|------|------|-------------------|
| **TUI Application** | interactive_bash (tmux) | Launch `./bin/dotkeeper`, send keystrokes, validate output |
| **Tests** | Bash (go test) | Run test suites, assert pass counts |
| **Build** | Bash (go build) | Compile binary, assert exit code 0 |

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 0 (Safety Net):
└── Task 1: Framework test suite + testable constructor

Wave 1 (Foundation):
├── Task 2: WindowSizeMsg propagation fix (independent, low risk)
└── Task 3: ProgramContext pattern (foundational)

Wave 2 (Core Improvements — after ProgramContext):
├── Task 4: Consolidated KeyMap system
├── Task 5: View interface standardization
├── Task 6: Message type consolidation
├── Task 7: AdaptiveColor theme support
└── Task 8: Help system migration to bubbles/help

Wave 3 (Loading & Restore Extraction):
├── Task 9: Spinner loading states
├── Task 10: Restore — PasswordValidator extraction
├── Task 11: Restore — FileSelector extraction
└── Task 12: Restore — DiffViewer extraction

Wave 4 (New Features):
├── Task 13: Toast notification system
├── Task 14: Auto-refresh dashboard
├── Task 15: Mouse support
└── Task 16: Search/filter overlay

Wave 5 (Final Integration):
└── Task 17: Integration verification + type assertion safety audit

Critical Path: Task 1 → Task 3 → Tasks 4-8 → Task 17
Parallel Speedup: ~55% faster than sequential
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2, 3 | None (must be first) |
| 2 | 1 | 17 | 3 |
| 3 | 1 | 4, 5, 6, 7, 8 | 2 |
| 4 | 3 | 17 | 5, 6, 7, 8 |
| 5 | 3 | 9, 10, 11, 12 | 4, 6, 7, 8 |
| 6 | 3 | 17 | 4, 5, 7, 8 |
| 7 | 3 | 17 | 4, 5, 6, 8 |
| 8 | 3 | 17 | 4, 5, 6, 7 |
| 9 | 5 | 17 | 10, 11, 12 |
| 10 | 5 | 17 | 9, 11, 12 |
| 11 | 5 | 17 | 9, 10, 12 |
| 12 | 5 | 17 | 9, 10, 11 |
| 13 | 5 | 17 | 14, 15, 16 |
| 14 | 5, 9 | 17 | 13, 15, 16 |
| 15 | 3 | 17 | 13, 14, 16 |
| 16 | 5 | 17 | 13, 14, 15 |
| 17 | ALL | None | None (must be last) |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Category |
|------|-------|---------------------|
| 0 | 1 | `deep` (testing requires understanding architecture) |
| 1 | 2, 3 | `quick` (bug fix) + `deep` (architectural change) |
| 2 | 4-8 | `unspecified-high` (multiple parallel) |
| 3 | 9-12 | `unspecified-high` (multiple parallel) |
| 4 | 13-16 | `unspecified-high` (feature development) |
| 5 | 17 | `deep` (integration verification) |

---

## TODOs

- [ ] 1. Framework Safety-Net Test Suite

  **What to do**:
  - Create `internal/tui/tui_test.go` with framework-level tests
  - Create `NewModelForTest(cfg *config.Config, store *history.Store) Model` constructor in `model.go` that accepts pre-built config (avoids filesystem dependency of `config.Load()`)
  - Create `testConfig()` helper that returns a minimal valid `*config.Config` with temp dir for BackupDir
  - Write tests covering:
    - `TestNewModel_InitializesAllViews` — verify all sub-models are non-zero
    - `TestUpdate_TabCyclesViews` — Tab cycles through all 5 views in order
    - `TestUpdate_ShiftTabCyclesReverse` — Shift+Tab cycles backwards
    - `TestUpdate_NumberKeysNavigate` — Keys 1-5 jump to correct views
    - `TestUpdate_QuitReturnsQuitCmd` — q/ctrl+c triggers tea.Quit
    - `TestUpdate_HelpToggles` — ? toggles showingHelp
    - `TestUpdate_WindowSizePropagates` — WindowSizeMsg reaches all views
    - `TestUpdate_DashboardShortcuts` — b/r/s navigate from dashboard
    - `TestUpdate_MessageRouting_RefreshBackupList` — RefreshBackupListMsg reaches backupList
    - `TestUpdate_MessageRouting_DashboardNavigate` — DashboardNavigateMsg switches view
    - `TestUpdate_InputActiveBlocksTab` — Tab ignored when view has active input
    - `TestView_MinTerminalSize` — Shows "too small" when below MinTerminalWidth/MinTerminalHeight
    - `TestView_HelpOverlay` — Help overlay renders when showingHelp is true
    - `TestSetupMode_CompleteTransitionsToDashboard` — SetupCompleteMsg exits setup, initializes views
  - Verify each test names a specific behavior and can fail independently

  **Must NOT do**:
  - Do NOT add testify, teatest, or other external dependencies
  - Do NOT modify existing view test patterns — follow `resize_test.go` conventions (stripANSI, sendKey helpers)
  - Do NOT test view-internal logic here — only framework routing and orchestration

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Requires deep understanding of BubbleTea message routing and the existing test patterns
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Needed for understanding tea.Cmd, message flow, and testing patterns

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 0 (solo)
  - **Blocks**: Tasks 2, 3
  - **Blocked By**: None (starts immediately)

  **References**:

  **Pattern References**:
  - `internal/tui/views/resize_test.go` — Follow this test structure: 8-terminal-size matrix, no-panic verification, stripANSI for assertions
  - `internal/tui/views/settings_test.go` — `sendKey()` helper pattern for simulating keyboard input
  - `internal/tui/views/testhelpers_test.go` — `stripANSI()` utility for removing color codes before string comparison
  - `internal/tui/views/dashboard_test.go` — Pattern for testing Init(), Update(), View() cycle

  **API/Type References**:
  - `internal/tui/model.go:33-54` — Model struct with all sub-model fields
  - `internal/tui/model.go:56-85` — `NewModel()` constructor (calls config.Load() — this is why we need NewModelForTest)
  - `internal/tui/update.go:90-263` — Full Update() loop to test
  - `internal/tui/model.go:16-23` — ViewState enum (DashboardView through SetupView)
  - `internal/tui/model.go:26` — `tabOrder` variable defining view cycling order

  **External References**:
  - BubbleTea testing: Use standard `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}` for key simulation

  **Acceptance Criteria**:

  - [ ] `internal/tui/tui_test.go` exists with ≥14 test functions
  - [ ] `NewModelForTest` constructor exists in model.go
  - [ ] `go test ./internal/tui/ -v -count=1` → ALL PASS (≥14 tests)
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS (no regressions)
  - [ ] `go build ./cmd/dotkeeper/` → exit code 0

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Framework tests pass
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. Run: go test ./internal/tui/ -v -count=1
      2. Count lines matching "--- PASS"
      3. Assert: ≥14 PASS lines
      4. Assert: 0 FAIL lines
    Expected Result: All 14+ framework tests pass
    Evidence: Test output captured

  Scenario: No regressions in existing tests
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. Run: go test ./internal/tui/... -race -count=1
      2. Assert: exit code 0
      3. Assert: 0 lines containing "FAIL"
    Expected Result: All existing + new tests pass with race detector
    Evidence: Test output captured

  Scenario: Binary builds successfully
    Tool: Bash (go build)
    Preconditions: None
    Steps:
      1. Run: go build ./cmd/dotkeeper/
      2. Assert: exit code 0
    Expected Result: Binary compiles without errors
    Evidence: Exit code captured
  ```

  **Commit**: YES
  - Message: `test(tui): add framework-level safety-net tests for model/update/view`
  - Files: `internal/tui/tui_test.go`, `internal/tui/model.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 2. Fix WindowSizeMsg Propagation Bug

  **What to do**:
  - **RED**: Write test `TestUpdate_WindowSizeMsg_PreservesViewCommands` in `tui_test.go`
    - Send WindowSizeMsg to Model
    - Verify returned commands include any commands from view Update() calls
    - Currently this test should FAIL because update.go:57-71 discards commands with `_`
  - **GREEN**: Fix `propagateWindowSize()` in `update.go` to collect and return all commands from view Updates
    - Change all `tm, _ = m.xxx.Update(viewMsg)` to `tm, cmd = m.xxx.Update(viewMsg)` and append to cmds
    - Return `tea.Batch(cmds...)` from the window size handler
  - **REFACTOR**: Verify no side effects from newly-collected commands

  **Must NOT do**:
  - Do NOT change the method signature of `propagateWindowSize()`
  - Do NOT change which views receive WindowSizeMsg (all 5 must still receive it)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small, focused bug fix in a single function
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding tea.Cmd and message propagation

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 3)
  - **Blocks**: Task 17
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `internal/tui/update.go:41-72` — `propagateWindowSize()` function with discarded commands (the bug)
  - `internal/tui/update.go:124-129` — Correct pattern: WindowSizeMsg handler in Update() stores dimensions

  **Acceptance Criteria**:

  - [ ] `propagateWindowSize()` collects all commands from view Updates
  - [ ] Zero `_, _` patterns for command returns in propagateWindowSize
  - [ ] `go test ./internal/tui/ -run TestUpdate_WindowSizeMsg -v` → PASS
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: WindowSizeMsg commands preserved
    Tool: Bash (go test)
    Preconditions: Task 1 complete (framework tests exist)
    Steps:
      1. Run: go test ./internal/tui/ -run TestUpdate_WindowSizeMsg -v
      2. Assert: PASS
    Expected Result: Test verifies commands are not discarded
    Evidence: Test output captured

  Scenario: No regressions after fix
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. Run: go test ./internal/tui/... -race -count=1
      2. Assert: exit code 0
    Expected Result: All tests pass
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `fix(tui): preserve view commands during window resize propagation`
  - Files: `internal/tui/update.go`, `internal/tui/tui_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 3. ProgramContext Pattern

  **What to do**:
  - **RED**: Write tests in `tui_test.go`:
    - `TestProgramContext_CarriesConfigAndStore` — verify ctx fields accessible
    - `TestProgramContext_DimensionsUpdatedOnResize` — verify width/height updated
    - `TestNewModelForTest_UsesProgramContext` — verify constructor works with ProgramContext
  - **GREEN**: Create `internal/tui/context.go`:
    ```go
    type ProgramContext struct {
        Config  *config.Config
        Store   *history.Store
        Width   int
        Height  int
    }
    ```
  - Refactor `Model` struct to embed `*ProgramContext`
  - Refactor ALL view constructors to accept `*ProgramContext`:
    - `NewDashboard(ctx *ProgramContext)`
    - `NewBackupList(ctx *ProgramContext)`
    - `NewRestore(ctx *ProgramContext)`
    - `NewSettings(ctx *ProgramContext)`
    - `NewLogs(ctx *ProgramContext)`
    - `NewSetup()` → `NewSetup(ctx *ProgramContext)` — ctx.Config may be nil during setup
  - Create `testProgramContext()` helper in test files for reuse
  - Update `NewModel()` to create ProgramContext and pass to views
  - Update `NewModelForTest()` to accept or create ProgramContext
  - Each view replaces `m.config` with `m.ctx.Config` and `m.store` with `m.ctx.Store` (where applicable)
  - Each view replaces `m.width`/`m.height` with `m.ctx.Width`/`m.ctx.Height` OR keeps local copies synced from ctx
  - Update ALL existing view tests to use `testProgramContext()` helper
  - Verify all 11 type assertions in update.go still compile

  **Must NOT do**:
  - Do NOT add future-feature fields to ProgramContext (no Notifications, Mouse, Search fields)
  - Do NOT remove `NewModelForTest` — it must remain for framework tests
  - Do NOT change view behavior — this is purely structural
  - Do NOT change message routing order

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Architectural change touching every view constructor and test file
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding of BubbleTea model composition and context patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 2)
  - **Blocks**: Tasks 4, 5, 6, 7, 8
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `internal/tui/model.go:34-54` — Current Model struct with individual fields
  - `internal/tui/model.go:56-85` — `NewModel()` constructor creating config/store/views
  - `internal/tui/views/dashboard.go:17-27` — View struct with config, width, height fields
  - `internal/tui/views/backuplist.go:31-43` — View struct with config AND store
  - `internal/tui/views/logs.go:69-83` — `NewLogs()` with variadic store parameter

  **API/Type References**:
  - `internal/config/config.go` — Config struct definition
  - `internal/history/history.go` — Store struct definition

  **Test References**:
  - `internal/tui/views/dashboard_test.go` — Current test pattern creating config directly
  - `internal/tui/views/settings_test.go` — Test pattern with helper setup functions
  - `internal/tui/views/testhelpers_test.go` — Shared test utilities

  **Acceptance Criteria**:

  - [ ] `internal/tui/context.go` exists with ProgramContext struct
  - [ ] `grep "config \*config.Config" internal/tui/views/*.go | grep -v _test.go` → 0 matches (views use ctx.Config)
  - [ ] `grep "func New" internal/tui/views/*.go | grep -v _test.go` — every constructor accepts `*ProgramContext` or `*tui.ProgramContext`
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS
  - [ ] `go build ./cmd/dotkeeper/` → exit code 0

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: ProgramContext tests pass
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Run: go test ./internal/tui/ -run TestProgramContext -v
      2. Assert: ≥3 PASS
    Expected Result: ProgramContext tests verify struct and integration
    Evidence: Test output captured

  Scenario: All view tests pass with refactored constructors
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. Run: go test ./internal/tui/views/ -v -count=1
      2. Assert: exit code 0
      3. Assert: 0 FAIL lines
    Expected Result: All existing view tests pass with new constructors
    Evidence: Test output captured

  Scenario: No direct config references remain in views
    Tool: Bash (grep)
    Preconditions: None
    Steps:
      1. Run: grep -rn "config \*config.Config" internal/tui/views/*.go | grep -v _test.go
      2. Assert: 0 matches
    Expected Result: All views use ctx.Config
    Evidence: Grep output captured
  ```

  **Commit**: YES
  - Message: `refactor(tui): introduce ProgramContext pattern for shared state`
  - Files: `internal/tui/context.go`, `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/tui_test.go`, `internal/tui/views/*.go`, `internal/tui/views/*_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 4. Consolidated KeyMap System

  **What to do**:
  - **RED**: Write `TestKeyMap_GlobalBindings` and `TestKeyMap_ViewSpecificBindings` in `tui_test.go`
  - **GREEN**: Create `internal/tui/keys.go`:
    - Define `AppKeyMap` struct with all global bindings using `bubbles/key`
    - Implement `key.Map` interface: `ShortHelp() []key.Binding` and `FullHelp() [][]key.Binding`
    - Move global key bindings from `update.go:11-37` to `keys.go`
    - Store AppKeyMap in ProgramContext (or Model struct)
    - Remove `var keys = DefaultKeyMap()` global from update.go
    - Update all `key.Matches(msg, keys.Xxx)` to use model/context-based keys
  - Define per-view KeyMaps for views that have complex keybindings (settings, restore, backuplist)
  - **REFACTOR**: Verify no duplicate key definitions

  **Must NOT do**:
  - Do NOT change the actual key bindings (same keys, same behavior)
  - Do NOT change key handling order in update.go

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Moderate refactoring across multiple files
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding `bubbles/key` patterns and KeyMap interface

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 5, 6, 7, 8)
  - **Blocks**: Task 17
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `internal/tui/update.go:11-37` — Current KeyMap struct and DefaultKeyMap()
  - `internal/tui/update.go:39` — `var keys = DefaultKeyMap()` global to eliminate
  - `internal/tui/update.go:154-192` — Global key matching: `key.Matches(msg, keys.Tab)`, etc.

  **External References**:
  - BubbleTea Skill: `key.NewBinding()` patterns, `key.Map` interface

  **Acceptance Criteria**:

  - [ ] `internal/tui/keys.go` exists with AppKeyMap struct
  - [ ] `grep "^var keys" internal/tui/update.go` → 0 matches (global removed)
  - [ ] `go test ./internal/tui/ -run TestKeyMap -v` → PASS
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: KeyMap tests pass
    Tool: Bash (go test)
    Steps:
      1. Run: go test ./internal/tui/ -run TestKeyMap -v
      2. Assert: PASS
    Expected Result: Key bindings verified
    Evidence: Test output captured

  Scenario: No global key variable
    Tool: Bash (grep)
    Steps:
      1. Run: grep "^var keys" internal/tui/update.go
      2. Assert: 0 matches
    Expected Result: Global eliminated
    Evidence: Grep output captured
  ```

  **Commit**: YES
  - Message: `refactor(tui): consolidate key bindings with bubbles/key AppKeyMap`
  - Files: `internal/tui/keys.go`, `internal/tui/update.go`, `internal/tui/tui_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 5. View Interface Standardization

  **What to do**:
  - **RED**: Write `TestViewInterface_AllViewsConform` in `tui_test.go` — verify each view satisfies the interface at compile time
  - **GREEN**: Define in `internal/tui/views/interfaces.go`:
    ```go
    type View interface {
        tea.Model
        HelpBindings() []HelpEntry
        StatusHelpText() string
    }

    type Refreshable interface {
        Refresh() tea.Cmd
    }

    type InputConsumer interface {
        IsInputActive() bool
    }
    ```
  - Migrate DashboardModel: Add `StatusHelpText()` method (move hardcoded text from view.go:80)
  - Alias `LogsModel.LoadHistory()` as `Refresh()`: `func (m LogsModel) Refresh() tea.Cmd { return m.LoadHistory() }`
  - Add `IsInputActive() bool` to ALL views (return false for Dashboard and Logs)
  - Update `model.go:isInputActive()` to use `InputConsumer` type assertion
  - Update `view.go:currentViewHelpText()` to use `View` interface

  **Must NOT do**:
  - Do NOT make the interface mandatory for compilation (use type assertions for optional methods)
  - Do NOT remove `LoadHistory()` from LogsModel — only alias
  - Do NOT add methods that future features need (search, mouse, toast)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Interface design + refactoring across all views
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding tea.Model interface composition

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 6, 7, 8)
  - **Blocks**: Tasks 9, 10, 11, 12, 13, 14, 16
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `internal/tui/views/helpers.go` — `HelpEntry` struct and `HelpProvider` interface
  - `internal/tui/view.go:77-92` — `currentViewHelpText()` with hardcoded Dashboard case
  - `internal/tui/model.go:116-128` — `isInputActive()` with per-view method name inconsistency
  - `internal/tui/views/logs.go:91` — `LoadHistory()` method to alias

  **Acceptance Criteria**:

  - [ ] `internal/tui/views/interfaces.go` exists with View, Refreshable, InputConsumer interfaces
  - [ ] All 5 views implement `StatusHelpText() string` (no hardcoded fallbacks in view.go)
  - [ ] All 5 views implement `IsInputActive() bool`
  - [ ] `LogsModel.Refresh()` exists and delegates to `LoadHistory()`
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Interface compliance verified at compile time
    Tool: Bash (go build)
    Steps:
      1. Run: go build ./cmd/dotkeeper/
      2. Assert: exit code 0
      3. Run: go vet ./internal/tui/...
      4. Assert: exit code 0
    Expected Result: All views satisfy interfaces
    Evidence: Build output captured

  Scenario: No hardcoded help text in view.go
    Tool: Bash (grep)
    Steps:
      1. Run: grep -n "select.*enter.*shortcuts" internal/tui/view.go
      2. Assert: 0 matches
    Expected Result: Dashboard help text moved to DashboardModel
    Evidence: Grep output captured
  ```

  **Commit**: YES
  - Message: `refactor(tui): standardize View/Refreshable/InputConsumer interfaces`
  - Files: `internal/tui/views/interfaces.go`, `internal/tui/views/dashboard.go`, `internal/tui/views/logs.go`, `internal/tui/views/backuplist.go`, `internal/tui/views/restore.go`, `internal/tui/views/settings.go`, `internal/tui/model.go`, `internal/tui/view.go`, `internal/tui/tui_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 6. Message Type Consolidation

  **What to do**:
  - **RED**: Write `TestMessageTypes_Consolidated` verifying message handling still works with consolidated types
  - **GREEN**: Create `internal/tui/messages/messages.go` (or define in `views/messages.go`):
    - Define shared message types: `ErrorMsg`, `SuccessMsg`, `LoadingMsg`, `RefreshMsg`
    - Consolidate duplicate patterns: `BackupErrorMsg`, `restoreErrorMsg`, `logsErrorMsg` → unified `ErrorMsg{Source string, Err error}`
    - Keep view-specific messages that carry unique data (e.g., `filesLoadedMsg`, `diffLoadedMsg`)
    - Update views to use consolidated types where appropriate
  - **REFACTOR**: Verify message routing order preserved (framework before state-specific)

  **Must NOT do**:
  - Do NOT change the routing of `backupsLoadedMsg` (shared between BackupList and Restore, consumed by active view only)
  - Do NOT change framework-level message handling order in update.go
  - Do NOT consolidate messages that carry fundamentally different data

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Cross-cutting change across message definitions and handlers
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding BubbleTea message patterns and routing

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5, 7, 8)
  - **Blocks**: Task 17
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `internal/tui/views/helpers.go:28` — `backupsLoadedMsg` (shared, DO NOT CHANGE routing)
  - `internal/tui/views/helpers.go:148-152` — `RefreshBackupListMsg`, `DashboardNavigateMsg`
  - `internal/tui/views/backuplist.go:20-29` — `BackupSuccessMsg`, `BackupErrorMsg`, `backupDeletedMsg`, `backupDeleteErrorMsg`
  - `internal/tui/views/restore.go:55-78` — 6 restore-specific message types
  - `internal/tui/views/logs.go:113-114` — `logsLoadedMsg`, `logsErrorMsg`
  - `internal/tui/views/dashboard.go:159-164` — `statusMsg`
  - `internal/tui/views/settings.go:59-61` — `pathDescsMsg`

  **Acceptance Criteria**:

  - [ ] Shared message types defined in a single location
  - [ ] No duplicate error message patterns across views
  - [ ] `backupsLoadedMsg` routing unchanged (consumed by active view only)
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Message consolidation doesn't break routing
    Tool: Bash (go test)
    Steps:
      1. Run: go test ./internal/tui/... -race -count=1
      2. Assert: exit code 0
    Expected Result: All tests pass with consolidated messages
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `refactor(tui): consolidate message types across views`
  - Files: `internal/tui/views/messages.go`, modified view files
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 7. AdaptiveColor Theme Support

  **What to do**:
  - **RED**: Write `TestStyles_AdaptiveColors` verifying styles use AdaptiveColor instead of hardcoded colors
  - **GREEN**: Refactor `internal/tui/styles/styles.go`:
    - Define color palette using `lipgloss.AdaptiveColor` for all colors:
      ```go
      var AccentColor = lipgloss.AdaptiveColor{Light: "#6C3EC2", Dark: "#7D56F4"}
      var TextColor = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#FFFFFF"}
      var MutedColor = lipgloss.AdaptiveColor{Light: "#999999", Dark: "#AAAAAA"}
      var ErrorColor = lipgloss.AdaptiveColor{Light: "#CC0000", Dark: "#FF5555"}
      var SuccessColor = lipgloss.AdaptiveColor{Light: "#00AA00", Dark: "#04B575"}
      var BgColor = lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#2A2A2A"}
      ```
    - Replace all `lipgloss.Color("#7D56F4")` → `AccentColor` (and similarly for all colors)
    - Ensure `DefaultStyles()` remains cheap (AdaptiveColor resolution happens at render time by lipgloss)
  - **REFACTOR**: Verify all views render correctly in both light and dark

  **Must NOT do**:
  - Do NOT make `DefaultStyles()` expensive (no per-call detection)
  - Do NOT change the visual appearance on dark terminals (same colors as before)
  - Do NOT change the Styles struct fields

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Cross-cutting style change across single file
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding Lip Gloss AdaptiveColor and style patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5, 6, 8)
  - **Blocks**: Task 17
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `internal/tui/styles/styles.go:50-137` — All hardcoded `lipgloss.Color()` calls to replace
  - `internal/tui/styles/styles.go:144-154` — `NewListDelegate()` with hardcoded colors

  **Acceptance Criteria**:

  - [ ] `grep 'lipgloss.Color("#' internal/tui/styles/styles.go` → 0 matches (all replaced with AdaptiveColor or palette variables)
  - [ ] Palette variables defined with Light/Dark variants
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS
  - [ ] `go build ./cmd/dotkeeper/` → exit code 0

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: No hardcoded colors remain
    Tool: Bash (grep)
    Steps:
      1. Run: grep -c 'lipgloss.Color("#' internal/tui/styles/styles.go
      2. Assert: 0
    Expected Result: All colors use AdaptiveColor palette
    Evidence: Grep output captured

  Scenario: Application still builds and tests pass
    Tool: Bash (go test + go build)
    Steps:
      1. Run: go test ./internal/tui/... -race -count=1
      2. Assert: exit code 0
      3. Run: go build ./cmd/dotkeeper/
      4. Assert: exit code 0
    Expected Result: No compilation errors from color changes
    Evidence: Build output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): add AdaptiveColor theme support for light/dark terminals`
  - Files: `internal/tui/styles/styles.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 8. Help System Migration to bubbles/help

  **What to do**:
  - **RED**: Write `TestHelp_UsesBubblesComponent` verifying help renders via bubbles/help
  - **GREEN**: Refactor `internal/tui/help.go`:
    - Add `help.Model` to main Model struct
    - Create adapter that converts `[]HelpEntry` → `[]key.Binding` for bubbles/help consumption
    - Replace custom `renderHelpOverlay()` with `help.Model.View()`
    - Keep `HelpEntry` as the view-facing API (adapter handles conversion)
  - Update `view.go` to use `help.Model.View()` for the inline help bar (bottom line)
  - Initialize `help.Model` in `NewModel()` and `NewModelForTest()`

  **Must NOT do**:
  - Do NOT change how views define their help bindings (keep `HelpBindings() []HelpEntry`)
  - Do NOT remove the `?` key toggle for full help overlay
  - Do NOT add `go get` for new packages — bubbles/help is part of bubbles (already imported)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Component migration with adapter pattern
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding bubbles/help component API and integration

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5, 6, 7)
  - **Blocks**: Task 17
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `internal/tui/help.go` — Current custom help system (65 lines)
  - `internal/tui/view.go:31-33` — Help overlay rendering
  - `internal/tui/view.go:65-71` — Inline help bar rendering
  - `internal/tui/views/helpers.go` — `HelpEntry` struct definition

  **External References**:
  - BubbleTea Skill: `bubbles/help` component — `help.New()`, `help.Model.View(keyMap)`
  - BubbleTea Skill: `key.Map` interface — `ShortHelp()`, `FullHelp()` for bubbles/help

  **Acceptance Criteria**:

  - [ ] `grep "help.Model" internal/tui/model.go` → ≥1 match
  - [ ] `grep "help.New" internal/tui/model.go` → ≥1 match
  - [ ] Views still return `[]HelpEntry` (adapter handles conversion)
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Help component integrated
    Tool: Bash (grep + go test)
    Steps:
      1. Run: grep "help.Model" internal/tui/model.go
      2. Assert: ≥1 match
      3. Run: go test ./internal/tui/ -run TestHelp -v
      4. Assert: PASS
    Expected Result: bubbles/help integrated with adapter
    Evidence: Output captured

  Scenario: TUI launches with help working
    Tool: interactive_bash (tmux)
    Preconditions: Binary built
    Steps:
      1. tmux new-session: ./bin/dotkeeper
      2. Wait for TUI render (timeout: 3s)
      3. Send keys: "?"
      4. Assert: Help overlay visible in output
      5. Send keys: "q"
    Expected Result: Help overlay appears and app exits cleanly
    Evidence: Terminal output captured
  ```

  **Commit**: YES
  - Message: `refactor(tui): migrate help system to bubbles/help component`
  - Files: `internal/tui/help.go`, `internal/tui/model.go`, `internal/tui/view.go`, `internal/tui/tui_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 9. Spinner Loading States

  **What to do**:
  - **RED**: Write `TestDashboard_ShowsSpinnerWhileLoading` and `TestBackupList_ShowsSpinnerWhileLoading`
  - **GREEN**: Add `bubbles/spinner` to views that perform async loading:
    - `DashboardModel`: spinner while `refreshStatus()` is running
    - `BackupListModel`: spinner while backup is creating or refreshing
    - `RestoreModel`: spinner during `phaseRestoring`
    - `LogsModel`: spinner while loading history
  - Each view adds:
    ```go
    spinner  spinner.Model
    loading  bool
    ```
  - Spinner initialized in constructor: `s := spinner.New(); s.Spinner = spinner.Dot`
  - `Init()` returns `tea.Batch(existingCmd, m.spinner.Tick)`
  - `Update()` handles `spinner.TickMsg` when loading
  - `View()` shows spinner when `m.loading` is true
  - Set `m.loading = true` before async command, `m.loading = false` on result message

  **Must NOT do**:
  - Do NOT block in Update() — spinner tick is already a tea.Cmd
  - Do NOT show spinner for instant operations (only for async I/O)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Adding component to 4 views
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding spinner component lifecycle (Tick, TickMsg)

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10, 11, 12)
  - **Blocks**: Task 14, 17
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `internal/tui/views/dashboard.go:49-51` — Current Init() returning refreshStatus (no loading state)
  - `internal/tui/views/dashboard.go:75-81` — statusMsg handler (sets data but no loading toggle)
  - `internal/tui/views/backuplist.go:118-136` — Backup success/error handlers

  **External References**:
  - BubbleTea Skill: Spinner component — `spinner.New()`, `spinner.Dot`, `spinner.TickMsg`

  **Acceptance Criteria**:

  - [ ] Dashboard, BackupList, Restore, Logs have spinner fields
  - [ ] `go test ./internal/tui/views/ -run TestSpinner -v` → PASS (or relevant test names)
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Dashboard shows spinner on init
    Tool: Bash (go test)
    Steps:
      1. Run: go test ./internal/tui/views/ -run TestDashboard_ShowsSpinner -v
      2. Assert: PASS
    Expected Result: Spinner renders during loading
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): add loading spinners to async views`
  - Files: `internal/tui/views/dashboard.go`, `internal/tui/views/backuplist.go`, `internal/tui/views/restore.go`, `internal/tui/views/logs.go`, view test files
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 10. Restore — PasswordValidator Extraction

  **What to do**:
  - **RED**: Write `TestPasswordValidator_ValidatesPassword` and `TestPasswordValidator_TracksAttempts` in `views/restore_password_test.go`
  - **GREEN**: Create `internal/tui/views/restore_password.go`:
    - Extract password validation logic from restore.go
    - Struct: `PasswordValidator` with `input textinput.Model`, `attempts int`, `maxAttempts int`, `status string`
    - Methods: `Init()`, `Update()`, `View()`, `Reset()`, `IsActive() bool`
    - Encapsulate: password input display, attempt tracking, validation command
  - Update `RestoreModel` to use `PasswordValidator` instead of direct passwordInput/passwordAttempts fields
  - Reuse `PasswordValidator` in `BackupListModel` (optional, if natural fit)

  **Must NOT do**:
  - Do NOT move to `components/` — keep in `views/` package
  - Do NOT break existing restore tests

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Sub-component extraction with test preservation
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding component composition patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9, 11, 12)
  - **Blocks**: Task 17
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `internal/tui/views/restore.go:34-53` — RestoreModel fields to extract (passwordInput, passwordAttempts)
  - `internal/tui/views/restore.go:145-153` — `validatePassword()` command
  - `internal/tui/views/restore.go:255-276` — `handlePasswordKey()` handler (approximate lines)
  - `internal/tui/components/passwordinput.go` — Existing password input wrapper (reuse)

  **Acceptance Criteria**:

  - [ ] `internal/tui/views/restore_password.go` exists with PasswordValidator struct
  - [ ] `go test ./internal/tui/views/ -run TestPasswordValidator -v` → PASS
  - [ ] `go test ./internal/tui/views/ -run TestRestore -v` → ALL existing tests still PASS
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Commit**: YES
  - Message: `refactor(tui): extract PasswordValidator sub-component from restore`
  - Files: `internal/tui/views/restore_password.go`, `internal/tui/views/restore_password_test.go`, `internal/tui/views/restore.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 11. Restore — FileSelector Extraction

  **What to do**:
  - **RED**: Write `TestFileSelector_ToggleSelection` and `TestFileSelector_SelectAll` in `views/restore_fileselect_test.go`
  - **GREEN**: Create `internal/tui/views/restore_fileselect.go`:
    - Extract file selection logic from restore.go
    - Struct: `FileSelector` with `list list.Model`, `selected map[string]bool`
    - Methods: `Init()`, `Update()`, `View()`, `ToggleItem()`, `SelectAll()`, `DeselectAll()`, `SelectedFiles() []string`, `IsActive() bool`
    - Encapsulate: file list display, selection toggling, select-all/deselect-all, diff preview trigger
  - Update `RestoreModel` to use `FileSelector`

  **Must NOT do**:
  - Do NOT move to `components/` — keep in `views/`
  - Do NOT break existing restore tests

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Sub-component extraction with selection logic
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding list component and multi-select patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9, 10, 12)
  - **Blocks**: Task 17
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `internal/tui/views/restore.go:34-53` — RestoreModel fields to extract (fileList, selectedFiles)
  - `internal/tui/views/restore.go:82-103` — fileItem struct with selection display
  - `internal/tui/views/restore.go:278-327` — `handleFileSelectKey()` handler (approximate lines)

  **Acceptance Criteria**:

  - [ ] `internal/tui/views/restore_fileselect.go` exists with FileSelector struct
  - [ ] `go test ./internal/tui/views/ -run TestFileSelector -v` → PASS
  - [ ] `go test ./internal/tui/views/ -run TestRestore -v` → ALL existing tests still PASS
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Commit**: YES
  - Message: `refactor(tui): extract FileSelector sub-component from restore`
  - Files: `internal/tui/views/restore_fileselect.go`, `internal/tui/views/restore_fileselect_test.go`, `internal/tui/views/restore.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 12. Restore — DiffViewer Extraction

  **What to do**:
  - **RED**: Write `TestDiffViewer_RendersContent` and `TestDiffViewer_ScrollNavigation` in `views/restore_diff_test.go`
  - **GREEN**: Create `internal/tui/views/restore_diff.go`:
    - Extract diff preview logic from restore.go
    - Struct: `DiffViewer` with `viewport viewport.Model`, `content string`, `file string`, `loading bool`
    - Methods: `Init()`, `Update()`, `View()`, `SetContent(diff, filename)`, `IsActive() bool`
    - Encapsulate: viewport management, j/k/g/G scrolling, diff content rendering
  - Update `RestoreModel` to use `DiffViewer`

  **Must NOT do**:
  - Do NOT move to `components/` — keep in `views/`
  - Do NOT break existing restore tests

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Sub-component extraction with viewport integration
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding viewport component and scroll patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9, 10, 11)
  - **Blocks**: Task 17
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `internal/tui/views/restore.go:34-53` — RestoreModel fields to extract (viewport, currentDiff, diffFile)
  - `internal/tui/views/restore.go:329-351` — `handleDiffPreviewKey()` handler (approximate lines)
  - `internal/tui/views/restore.go:165-185` — `loadDiff()` async command

  **Acceptance Criteria**:

  - [ ] `internal/tui/views/restore_diff.go` exists with DiffViewer struct
  - [ ] `go test ./internal/tui/views/ -run TestDiffViewer -v` → PASS
  - [ ] `go test ./internal/tui/views/ -run TestRestore -v` → ALL existing tests still PASS
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Commit**: YES
  - Message: `refactor(tui): extract DiffViewer sub-component from restore`
  - Files: `internal/tui/views/restore_diff.go`, `internal/tui/views/restore_diff_test.go`, `internal/tui/views/restore.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 13. Toast Notification System

  **What to do**:
  - **RED**: Write `TestToast_ShowsAndAutoDismisses` and `TestToast_RendersOverContent`
  - **GREEN**: Create `internal/tui/components/toast.go`:
    - Struct: `Toast` with `message string`, `style ToastStyle` (success/error/info), `timer time.Duration`, `visible bool`
    - Methods: `Show(msg, style, duration)`, `Update(msg)`, `View()`, `IsVisible() bool`
    - Uses `tea.Tick` for auto-dismiss after duration
    - Renders as a floating bar at bottom of screen using Lip Gloss positioning
  - Integrate into main Model: replace direct status/error strings with Toast
  - Wire BackupSuccessMsg, BackupErrorMsg, restoreCompleteMsg to Toast.Show()

  **Must NOT do**:
  - Do NOT block Update() — use tea.Tick for timers
  - Do NOT replace ALL status messages immediately — incremental adoption

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: New component with timer management
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding tea.Tick for timed events, component patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 14, 15, 16)
  - **Blocks**: Task 17
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `internal/tui/views/helpers.go` — `RenderStatusBar()` — current status display to evolve
  - `internal/tui/views/backuplist.go:119-136` — Success/error status handling
  - `internal/tui/components/` — Existing component directory

  **Acceptance Criteria**:

  - [ ] `internal/tui/components/toast.go` exists
  - [ ] `go test ./internal/tui/components/ -run TestToast -v` → PASS
  - [ ] Toast auto-dismisses after configured duration
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Toast shows on backup success
    Tool: interactive_bash (tmux)
    Preconditions: Binary built, config exists
    Steps:
      1. Launch TUI in tmux
      2. Navigate to Backups (Tab or "2")
      3. Trigger backup flow
      4. Observe toast notification on success/error
      5. Wait for auto-dismiss
      6. Send "q" to quit
    Expected Result: Toast appears and auto-dismisses
    Evidence: Terminal output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): add toast notification component with auto-dismiss`
  - Files: `internal/tui/components/toast.go`, `internal/tui/components/toast_test.go`, updated view files
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 14. Auto-Refresh Dashboard

  **What to do**:
  - **RED**: Write `TestDashboard_AutoRefreshOnInterval` verifying periodic refresh
  - **GREEN**: Add timer-based auto-refresh to DashboardModel:
    - Add `refreshInterval time.Duration` (default: 30 seconds)
    - Use `tea.Tick(m.refreshInterval, ...)` to trigger periodic refreshes
    - Only refresh when Dashboard is the active view (avoid unnecessary I/O)
    - Add `lastRefresh time.Time` to avoid redundant refreshes
    - Display last refresh timestamp in View()

  **Must NOT do**:
  - Do NOT refresh when another view is active
  - Do NOT use goroutines — only tea.Tick

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Timer management in BubbleTea
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding tea.Tick and timer patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 13, 15, 16)
  - **Blocks**: Task 17
  - **Blocked By**: Tasks 5, 9

  **References**:

  **Pattern References**:
  - `internal/tui/views/dashboard.go:166-191` — Current `refreshStatus()` and `Refresh()` commands
  - `internal/tui/views/dashboard.go:49-51` — Current Init() (one-time refresh)

  **Acceptance Criteria**:

  - [ ] Dashboard auto-refreshes every 30 seconds when active
  - [ ] Refresh stops when another view is active
  - [ ] `go test ./internal/tui/views/ -run TestDashboard_AutoRefresh -v` → PASS
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Commit**: YES
  - Message: `feat(tui): add auto-refresh timer to dashboard`
  - Files: `internal/tui/views/dashboard.go`, `internal/tui/views/dashboard_test.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 15. Mouse Support

  **What to do**:
  - **RED**: Write `TestModel_MouseSupportEnabled` verifying tea.WithMouseCellMotion is used
  - **GREEN**: Enable mouse support:
    - Add `tea.WithMouseCellMotion()` to `tea.NewProgram()` in `cmd/dotkeeper/main.go`
    - Handle `tea.MouseMsg` in update.go for tab bar clicks (map mouse position to tab index)
    - Handle `tea.MouseMsg` in Dashboard for button clicks
    - Optionally: use `bubblezone` for declarative click regions

  **Must NOT do**:
  - Do NOT require mouse — all keyboard shortcuts must continue working
  - Do NOT add mouse handling to complex views (settings, restore) — start with dashboard + tabs

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: New interaction paradigm with mouse event handling
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding tea.MouseMsg, WithMouseCellMotion, bubblezone

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 13, 14, 16)
  - **Blocks**: Task 17
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `cmd/dotkeeper/main.go` — Program creation (add WithMouseCellMotion)
  - `internal/tui/update.go:90` — Update() — add tea.MouseMsg handler
  - `internal/tui/components/tabbar.go` — Tab positions for click detection

  **External References**:
  - BubbleTea Skill: `tea.WithMouseCellMotion()`, `tea.MouseMsg` handling
  - BubbleTea Skill: `bubblezone` for declarative mouse regions

  **Acceptance Criteria**:

  - [ ] `grep "WithMouseCellMotion" cmd/dotkeeper/main.go` → ≥1 match
  - [ ] Tab bar clicks switch views
  - [ ] Keyboard navigation still works (no regression)
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Commit**: YES
  - Message: `feat(tui): add mouse support for tab bar and dashboard navigation`
  - Files: `cmd/dotkeeper/main.go`, `internal/tui/update.go`, `internal/tui/components/tabbar.go`, `internal/tui/views/dashboard.go`
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 16. Search/Filter Overlay

  **What to do**:
  - **RED**: Write `TestSearch_FiltersBackupList` and `TestSearch_ClearsOnEscape`
  - **GREEN**: Create search/filter capability:
    - Add search toggle key `/` (standard vim-like search trigger)
    - When `/` pressed: show textinput overlay at top of list views
    - Filter list items as user types (use `list.SetFilteringEnabled(true)` from bubbles/list or custom filter)
    - Esc clears search and shows all items
    - Apply to: BackupList, Restore backup selection, Logs
    - Add search indicator in status bar when filter is active

  **Must NOT do**:
  - Do NOT add search to Settings (already has navigation)
  - Do NOT add search to Dashboard (no list to search)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: New feature across 3 views
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding list.Model filtering and textinput integration

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 13, 14, 15)
  - **Blocks**: Task 17
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `internal/tui/views/backuplist.go:45-56` — BackupListModel with list.Model (currently filtering disabled)
  - `internal/tui/styles/styles.go:157-163` — `NewMinimalList()` creates list with `SetFilteringEnabled(false)`
  - `internal/tui/views/logs.go:141-155` — Current filter cycling (f key cycles all/backup/restore)

  **Acceptance Criteria**:

  - [ ] `/` key activates search in BackupList, Restore, and Logs views
  - [ ] Typing filters list items in real-time
  - [ ] Esc clears filter and shows all items
  - [ ] Status bar indicates when filter is active
  - [ ] `go test ./internal/tui/views/ -run TestSearch -v` → PASS
  - [ ] `go test ./internal/tui/... -race -count=1` → ALL PASS

  **Commit**: YES
  - Message: `feat(tui): add search/filter overlay for list views`
  - Files: `internal/tui/views/backuplist.go`, `internal/tui/views/restore.go`, `internal/tui/views/logs.go`, test files
  - Pre-commit: `go test ./internal/tui/... -race -count=1`

---

- [ ] 17. Integration Verification + Type Assertion Safety Audit

  **What to do**:
  - Audit ALL type assertions in `update.go` — replace unchecked `model.(Type)` with `model, ok := ...(Type)` pattern
  - Run full test suite with race detector: `go test ./... -race -count=1`
  - Run `go vet ./...` for static analysis
  - Build binary: `go build ./cmd/dotkeeper/`
  - Launch TUI manually via tmux and verify:
    - All 5 views render correctly
    - Tab cycling works through all views
    - Number keys (1-5) navigate correctly
    - Help overlay toggles with `?`
    - Spinner shows during loading
    - Search works in list views
    - Mouse clicks on tab bar work
    - Toast notifications appear on operations
  - Document any issues found in test output

  **Must NOT do**:
  - Do NOT add new features in this task — only verify and fix

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Comprehensive integration verification requires careful testing
  - **Skills**: [`bubbletea`]
    - `bubbletea`: Understanding of full BubbleTea application lifecycle

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 5 (solo, final)
  - **Blocks**: None (final task)
  - **Blocked By**: ALL previous tasks

  **References**:

  **Pattern References**:
  - `internal/tui/update.go:55-71` — 11 type assertion sites to audit
  - `internal/tui/update.go:234-260` — State-specific routing with type assertions

  **Acceptance Criteria**:

  - [ ] Zero unchecked type assertions in update.go (all use `, ok` pattern)
  - [ ] `go test ./... -race -count=1` → ALL PASS
  - [ ] `go vet ./...` → 0 issues
  - [ ] `go build ./cmd/dotkeeper/` → exit code 0
  - [ ] TUI launches and all views render without panic

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Full test suite passes with race detector
    Tool: Bash (go test)
    Steps:
      1. Run: go test ./... -race -count=1
      2. Assert: exit code 0
      3. Count "FAIL" lines
      4. Assert: 0 failures
    Expected Result: All tests pass, no race conditions
    Evidence: Full test output captured

  Scenario: Static analysis clean
    Tool: Bash (go vet)
    Steps:
      1. Run: go vet ./...
      2. Assert: exit code 0
    Expected Result: No issues found
    Evidence: Output captured

  Scenario: Type assertions are safe
    Tool: Bash (grep)
    Steps:
      1. Run: grep -n '\.(' internal/tui/update.go | grep -v ', ok' | grep -v 'switch' | grep -v '//'
      2. Assert: 0 unsafe assertions
    Expected Result: All type assertions use safety checks
    Evidence: Grep output captured

  Scenario: TUI full interaction test
    Tool: interactive_bash (tmux)
    Preconditions: Binary built
    Steps:
      1. tmux new-session: ./bin/dotkeeper
      2. Wait for TUI render (timeout: 3s)
      3. Send Tab key 5 times (cycle all views)
      4. Send "1" key (go to Dashboard)
      5. Send "?" (toggle help)
      6. Send any key (dismiss help)
      7. Send "/" (search — if in list view)
      8. Send Esc (clear search)
      9. Send "q" (quit)
      10. Assert: process exited with code 0
    Expected Result: Full interaction cycle completes without crash
    Evidence: Terminal output captured
  ```

  **Commit**: YES
  - Message: `fix(tui): add type assertion safety checks and verify integration`
  - Files: `internal/tui/update.go`
  - Pre-commit: `go test ./... -race -count=1`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `test(tui): add framework-level safety-net tests` | tui_test.go, model.go | `go test ./internal/tui/ -v` |
| 2 | `fix(tui): preserve view commands during window resize` | update.go, tui_test.go | `go test ./internal/tui/... -race` |
| 3 | `refactor(tui): introduce ProgramContext pattern` | context.go, model.go, update.go, all views | `go test ./internal/tui/... -race` |
| 4 | `refactor(tui): consolidate key bindings with AppKeyMap` | keys.go, update.go | `go test ./internal/tui/... -race` |
| 5 | `refactor(tui): standardize View/Refreshable/InputConsumer interfaces` | interfaces.go, all views | `go test ./internal/tui/... -race` |
| 6 | `refactor(tui): consolidate message types` | messages.go, view files | `go test ./internal/tui/... -race` |
| 7 | `feat(tui): add AdaptiveColor theme support` | styles.go | `go test ./internal/tui/... -race` |
| 8 | `refactor(tui): migrate help to bubbles/help` | help.go, model.go, view.go | `go test ./internal/tui/... -race` |
| 9 | `feat(tui): add loading spinners to async views` | 4 view files | `go test ./internal/tui/... -race` |
| 10 | `refactor(tui): extract PasswordValidator from restore` | restore_password.go, restore.go | `go test ./internal/tui/... -race` |
| 11 | `refactor(tui): extract FileSelector from restore` | restore_fileselect.go, restore.go | `go test ./internal/tui/... -race` |
| 12 | `refactor(tui): extract DiffViewer from restore` | restore_diff.go, restore.go | `go test ./internal/tui/... -race` |
| 13 | `feat(tui): add toast notification component` | toast.go, view files | `go test ./internal/tui/... -race` |
| 14 | `feat(tui): add auto-refresh timer to dashboard` | dashboard.go | `go test ./internal/tui/... -race` |
| 15 | `feat(tui): add mouse support for navigation` | main.go, update.go, tabbar.go | `go test ./internal/tui/... -race` |
| 16 | `feat(tui): add search/filter overlay for lists` | 3 view files | `go test ./internal/tui/... -race` |
| 17 | `fix(tui): type assertion safety + integration verify` | update.go | `go test ./... -race` |

---

## Success Criteria

### Verification Commands
```bash
# Full test suite with race detector
go test ./... -race -count=1  # Expected: ALL PASS

# Static analysis
go vet ./...  # Expected: 0 issues

# Build verification
go build ./cmd/dotkeeper/  # Expected: exit code 0

# Framework test count
go test ./internal/tui/ -v -count=1 2>&1 | grep -c "--- PASS"  # Expected: ≥14

# No globals for keys
grep "^var keys" internal/tui/update.go  # Expected: 0 matches

# No hardcoded colors
grep -c 'lipgloss.Color("#' internal/tui/styles/styles.go  # Expected: 0

# ProgramContext adopted
grep "config \*config.Config" internal/tui/views/*.go | grep -v _test.go  # Expected: 0 matches

# Type assertion safety
grep -n '\.(' internal/tui/update.go | grep -v ', ok' | grep -v 'switch' | grep -v '//'  # Expected: 0 matches
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass with -race flag
- [ ] Binary builds and TUI launches
- [ ] All 17 tasks have passing acceptance criteria
