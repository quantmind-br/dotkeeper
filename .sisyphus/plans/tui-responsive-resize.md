# TUI Responsive Resize Optimization

## TL;DR

> **Quick Summary**: Fix 3 bugs and implement 6 improvements to ensure the TUI renders correctly across terminal sizes from 40x15 to arbitrarily large, with proper responsive breakpoints and a minimum-size warning.
> 
> **Deliverables**:
> - Unified chrome height constant (eliminates off-by-one rendering)
> - Working responsive dashboard layout for narrow terminals
> - Setup view components properly sized on resize
> - Responsive password input width
> - Constrained filepicker width via lipgloss wrapping
> - Named breakpoint constants replacing magic numbers
> - Minimum terminal size warning (40x15)
> - Consistent viewport padding in restore view
> - Comprehensive resize test matrix
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES - 3 waves
> **Critical Path**: Task 1 → Task 2 → Tasks 3-6 (parallel) → Task 7

---

## Context

### Original Request
Optimize the TUI for correct rendering in terminal windows of different sizes.

### Interview Summary
**Key Discussions**:
- Scope: Complete (all 9 items — 3 bugs + 6 improvements)
- Minimum terminal size: 40x15 — below this, show a warning message
- Tests-after strategy with existing Go testing infrastructure (10 test files under internal/tui/)

**Research Findings**:
- `mainChromeHeight=5` (model.go:30) vs `ViewChromeHeight=6` (styles.go:9) causes double-subtraction — views receive `height - 5` from propagateWindowSize, then subtract 6 more internally
- Dashboard responsive layout is broken — both branches of `width >= 80` produce identical output
- Setup view stores width/height but never applies them to filepicker or pathcompleter
- PasswordInput has hardcoded `Width = 40`
- bubbles filepicker has **no Width API** — must use lipgloss wrapping
- 6 magic number breakpoints (80, 60, 40) scattered without named constants
- Restore viewport has asymmetric padding between Update() and View()

### Metis Review
**Identified Gaps** (addressed):
- Chrome height is a **double-subtraction** bug, not just a mismatch — fixed by unifying into single-level subtraction
- ContentArea has MarginLeft(2) that eats into available width — sub-views should account for this
- Zero-width/height from terminal multiplexers needs defensive clamping
- FilePicker has no Width API — using lipgloss wrapping instead
- Tab bar can overflow at extreme narrow widths (< 30) — handled by minimum size enforcement
- Setup mode bypasses propagateWindowSize entirely — needs separate treatment

---

## Work Objectives

### Core Objective
Ensure the TUI renders correctly and without visual glitches across terminal sizes from 40x15 to arbitrarily large, with proper responsive breakpoints and graceful degradation.

### Concrete Deliverables
- Modified files: `internal/tui/styles/styles.go`, `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/view.go`, `internal/tui/views/dashboard.go`, `internal/tui/views/setup.go`, `internal/tui/views/restore.go`, `internal/tui/views/settings.go`, `internal/tui/views/backuplist.go`, `internal/tui/views/logs.go`, `internal/tui/components/passwordinput.go`
- New/modified test files: `internal/tui/views/dashboard_test.go`, `internal/tui/views/setup_test.go`, `internal/tui/views/restore_test.go`, `internal/tui/views/resize_test.go` (new)

### Definition of Done
- [x] `make test` passes with zero failures
- [x] `go vet ./internal/tui/...` reports zero issues
- [x] All 9 items addressed (3 bugs + 6 improvements)
- [x] Terminal at 40x15 shows content without overflow/truncation
- [x] Terminal below 40x15 shows a warning message
- [x] Terminal at 80x24 renders same as before (no regression)
- [x] Terminal at 200x100 renders without overflow

### Must Have
- Single source of truth for chrome height calculation
- Dashboard cards stack vertically when terminal width < 80
- Setup view filepicker resizes on WindowSizeMsg
- Minimum terminal size warning at < 40x15
- Named constants for all breakpoints
- Comprehensive resize tests

### Must NOT Have (Guardrails)
- **NO styling changes** — colors, padding, borders, fonts stay unchanged. Layout adjustments only.
- **NO tab bar icon mode** — TabBar already has `ShortLabel` abbreviation. Don't add icons.
- **NO base model refactor** — Don't create shared struct for views. Fix sizing only.
- **NO configurable minimum size** — Hardcode 40x15. Not a config option.
- **NO animation/transition** on resize events
- **NO help overlay changes** — Help overlay is out of scope unless broken by resize fixes
- **NO files outside `internal/tui/`** tree
- **NO changes to RenderStatusBar** — Already handles truncation correctly
- **NO unrelated bug fixes** — If you notice other bugs, log them but don't fix them

---

## Verification Strategy

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> ALL tasks in this plan MUST be verifiable WITHOUT any human action.
> This is NOT conditional — it applies to EVERY task.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: YES (tests-after)
- **Framework**: Go testing (standard library)

### Agent-Executed QA Scenarios (MANDATORY — ALL tasks)

**Verification Tool by Deliverable Type:**

| Type | Tool | How Agent Verifies |
|------|------|-------------------|
| **TUI resize behavior** | interactive_bash (tmux) | Launch TUI, send resize, validate output |
| **Unit test results** | Bash (go test) | Run test suite, parse results |
| **Code correctness** | Bash (go vet, go build) | Compile and vet check |

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately):
└── Task 1: Extract breakpoint constants & unify chrome height

Wave 2 (After Wave 1):
├── Task 2: Fix dashboard responsive layout
├── Task 3: Fix setup view resize handling
├── Task 4: Fix restore viewport padding
├── Task 5: Make PasswordInput responsive + wrap FilePicker width
└── Task 6: Add minimum terminal size warning

Wave 3 (After Wave 2):
└── Task 7: Comprehensive resize test matrix
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2, 3, 4, 5, 6 | None (foundational) |
| 2 | 1 | 7 | 3, 4, 5, 6 |
| 3 | 1 | 7 | 2, 4, 5, 6 |
| 4 | 1 | 7 | 2, 3, 5, 6 |
| 5 | 1 | 7 | 2, 3, 4, 6 |
| 6 | 1 | 7 | 2, 3, 4, 5 |
| 7 | 2, 3, 4, 5, 6 | None | None (final) |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Agents |
|------|-------|-------------------|
| 1 | 1 | delegate_task(category="unspecified-high") |
| 2 | 2, 3, 4, 5, 6 | dispatch in parallel, each delegate_task(category="quick") |
| 3 | 7 | delegate_task(category="unspecified-high") |

---

## TODOs

- [x] 1. Extract breakpoint constants & unify chrome height calculation

  **What to do**:
  
  **Part A — Breakpoint constants:**
  - Add named constants to `internal/tui/styles/styles.go` (in the existing file, NOT a new file):
    ```go
    // Responsive breakpoint constants
    const (
        BreakpointWide    = 80  // Full horizontal layout (cards, tabs)
        BreakpointMedium  = 60  // Medium layout (action buttons horizontal)
        MinTerminalWidth  = 40  // Minimum supported terminal width
        MinTerminalHeight = 15  // Minimum supported terminal height
    )
    ```
  - Replace all magic numbers across the codebase with these constants:
    - `dashboard.go:121` → `m.width >= styles.BreakpointWide`
    - `dashboard.go:147` → `m.width >= styles.BreakpointMedium`
    - `tabbar.go:45` → `width < styles.BreakpointWide`

  **Part B — Unify chrome height:**
  - The current system has a **double-subtraction bug**: `propagateWindowSize()` subtracts `mainChromeHeight=5`, then each sub-view subtracts `ViewChromeHeight=6` again via `list.SetSize()`.
  - **Fix**: Make `propagateWindowSize()` the single source of truth. Sub-views should treat the height they receive as their full usable height and NOT subtract `ViewChromeHeight` again.
  - Count the actual chrome lines rendered in `view.go`:
    - AppTitle (1 line)
    - `\n` after title (1 line)
    - TabBar (1 line)
    - `\n\n` after tab bar (includes 1 blank line = 2 chars but 1 additional line)
    - [content area]
    - `\n\n` after content (1 blank line)
    - viewHelp line (1 line, always present for all current views)
    - `\n` after viewHelp
    - globalHelp line (1 line)
    - `\n` after globalHelp
  - Set `mainChromeHeight` to the correct value based on the actual count above
  - Remove `ViewChromeHeight` constant from `styles/styles.go`
  - Update ALL sub-views to stop subtracting `ViewChromeHeight`:
    - `backuplist.go:102`: Change `m.list.SetSize(msg.Width, msg.Height-styles.ViewChromeHeight)` → `m.list.SetSize(msg.Width, msg.Height)`
    - `restore.go:372-375`: Remove `- styles.ViewChromeHeight` from all SetSize calls
    - `settings.go:686`: Remove `- styles.ViewChromeHeight` from height calculation
    - `logs.go:123`: Remove `- styles.ViewChromeHeight` from SetSize call
  - **Also add width=0/height=0 defensive clamping** in `propagateWindowSize()`:
    ```go
    viewWidth := msg.Width
    if viewWidth < 0 {
        viewWidth = 0
    }
    viewHeight := msg.Height - mainChromeHeight
    if viewHeight < 0 {
        viewHeight = 0
    }
    viewMsg := tea.WindowSizeMsg{
        Width:  viewWidth,
        Height: viewHeight,
    }
    ```
  - Remove the defensive `if width <= 0 { width = 80 }` fallback from `settings.go:682-683` (no longer needed since propagateWindowSize clamps properly)

  **Must NOT do**:
  - Don't change any lipgloss styles (colors, padding, borders)
  - Don't refactor view architecture
  - Don't change how views are initialized in NewModel()
  - Don't modify the view rendering logic in view.go (just count lines accurately)
  - Don't change ContentArea or GlobalHelp margins

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Foundational task that touches many files and requires careful coordination across the chrome height chain
  - **Skills**: [`git-master`]
    - `git-master`: Atomic commit of coordinated cross-file changes
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Not applicable — Go TUI, not web frontend

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1 (solo)
  - **Blocks**: Tasks 2, 3, 4, 5, 6
  - **Blocked By**: None

  **References**:

  **Pattern References:**
  - `internal/tui/update.go:41-68` — `propagateWindowSize()` is the function to modify. Currently subtracts `mainChromeHeight=5` and broadcasts to all views. This is the single place where chrome height is applied.
  - `internal/tui/views/settings.go:680-701` — `resizeLists()` is the EXEMPLAR for how views should handle sizing. Follow this pattern for defensive clamping but remove the ViewChromeHeight subtraction.
  - `internal/tui/model.go:28-30` — `mainChromeHeight` constant definition with comment explaining the chrome composition.

  **API/Type References:**
  - `internal/tui/styles/styles.go:9` — `ViewChromeHeight = 6` — the constant to REMOVE.
  - `internal/tui/view.go:11-63` — `View()` method that renders the actual chrome. Count lines here to determine correct `mainChromeHeight` value.

  **File References (all consumers of ViewChromeHeight to update):**
  - `internal/tui/views/backuplist.go:102` — `m.list.SetSize(msg.Width, msg.Height-styles.ViewChromeHeight)`
  - `internal/tui/views/restore.go:372-375` — 3 SetSize calls using `styles.ViewChromeHeight`
  - `internal/tui/views/settings.go:686` — `height := m.height - styles.ViewChromeHeight`
  - `internal/tui/views/logs.go:123` — `m.list.SetSize(msg.Width, msg.Height-styles.ViewChromeHeight)`

  **Test References:**
  - `internal/tui/views/dashboard_test.go` — Existing resize tests to verify no regression
  - `internal/tui/views/restore_test.go:100` — Uses WindowSizeMsg{100, 50}

  **WHY Each Reference Matters:**
  - `update.go:41-68` — This is THE function being modified. Must understand current flow before changing.
  - `view.go:11-63` — Must count actual rendered lines to set correct mainChromeHeight.
  - `styles.go:9` — ViewChromeHeight is being REMOVED. Must find and update all consumers.
  - Each sub-view file — These are the consumers that currently double-subtract. Must remove their subtraction.

  **Acceptance Criteria**:

  - [x] `ViewChromeHeight` constant removed from `styles/styles.go`
  - [x] `mainChromeHeight` set to correct value matching actual chrome line count in `view.go`
  - [x] No file in `internal/tui/views/` references `styles.ViewChromeHeight`
  - [x] Breakpoint constants (`BreakpointWide`, `BreakpointMedium`, `MinTerminalWidth`, `MinTerminalHeight`) added to `styles/styles.go`
  - [x] No magic number `80` or `60` in `dashboard.go` or `tabbar.go` — replaced with named constants
  - [x] `propagateWindowSize()` clamps both width and height to >= 0
  - [x] `go build ./...` compiles successfully
  - [x] `make test` passes (all existing tests)
  - [x] `go vet ./internal/tui/...` reports zero issues

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Build succeeds after chrome height unification
    Tool: Bash
    Preconditions: Repository at latest state
    Steps:
      1. Run: go build ./...
      2. Assert: exit code 0
      3. Run: go vet ./internal/tui/...
      4. Assert: exit code 0
    Expected Result: No compilation or vet errors
    Evidence: Command output captured

  Scenario: All existing tests pass after changes
    Tool: Bash
    Preconditions: Repository at latest state
    Steps:
      1. Run: make test
      2. Assert: exit code 0
      3. Assert: output contains "PASS" for all test packages
      4. Assert: output does NOT contain "FAIL"
    Expected Result: Zero test regressions
    Evidence: Test output captured

  Scenario: ViewChromeHeight fully removed
    Tool: Bash (grep)
    Preconditions: Changes applied
    Steps:
      1. Run: grep -rn "ViewChromeHeight" internal/tui/
      2. Assert: zero matches (exit code 1 = no matches)
    Expected Result: No remaining references to ViewChromeHeight
    Evidence: grep output captured

  Scenario: Magic numbers replaced
    Tool: Bash (grep)
    Preconditions: Changes applied
    Steps:
      1. Run: grep -n "width < 80\|width >= 80\|width >= 60" internal/tui/views/dashboard.go internal/tui/components/tabbar.go
      2. Assert: zero matches (all replaced with named constants)
    Expected Result: No magic breakpoint numbers in dashboard or tabbar
    Evidence: grep output captured
  ```

  **Commit**: YES
  - Message: `refactor(tui): unify chrome height and extract breakpoint constants`
  - Files: `internal/tui/styles/styles.go`, `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/views/backuplist.go`, `internal/tui/views/restore.go`, `internal/tui/views/settings.go`, `internal/tui/views/logs.go`, `internal/tui/views/dashboard.go`, `internal/tui/components/tabbar.go`
  - Pre-commit: `make test`

---

- [x] 2. Fix dashboard responsive layout for narrow terminals

  **What to do**:
  - In `dashboard.go`, fix the `width >= 80` conditional so the `else` branch actually renders cards vertically:
    ```go
    if m.width >= styles.BreakpointWide {
        statsBlock = lipgloss.JoinHorizontal(lipgloss.Top, cards...)
    } else {
        statsBlock = lipgloss.JoinVertical(lipgloss.Left, cards...)
    }
    ```
  - This is the only change needed — the action buttons responsive layout at `width >= 60` already works correctly (horizontal vs vertical).
  - Verify that `styles.BreakpointWide` is used (from Task 1) instead of magic `80`.

  **Must NOT do**:
  - Don't change card styling (colors, padding, borders)
  - Don't add new breakpoints or layouts (2+1, 2+2, etc.)
  - Don't modify the action buttons responsive logic (already works)
  - Don't add width constraints to individual cards

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single-line fix in one file
  - **Skills**: [`git-master`]
    - `git-master`: Clean atomic commit

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 4, 5, 6)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References:**
  - `internal/tui/views/dashboard.go:120-126` — The broken responsive conditional. Both branches currently call `JoinHorizontal`. The `else` branch should call `JoinVertical`.
  - `internal/tui/views/dashboard.go:146-151` — The action buttons responsive logic (working example). `width >= 60`: horizontal, else: vertical. Follow same pattern for cards.

  **API/Type References:**
  - `internal/tui/styles/styles.go` — `BreakpointWide` constant (created in Task 1)

  **Test References:**
  - `internal/tui/views/dashboard_test.go:46` — Existing narrow terminal test at width=40. Verify this test still passes.

  **WHY Each Reference Matters:**
  - `dashboard.go:120-126` — This is THE code being fixed. Must change `else` branch from `JoinHorizontal` to `JoinVertical`.
  - `dashboard.go:146-151` — Working reference for how responsive layout should behave.

  **Acceptance Criteria**:

  - [x] At `width >= 80`: cards render horizontally (same as before)
  - [x] At `width < 80`: cards render vertically (stacked)
  - [x] No magic number `80` in the conditional — uses `styles.BreakpointWide`
  - [x] `make test` passes

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Dashboard renders cards vertically at narrow width
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestDashboard
      2. Assert: all tests pass
      3. In test or manual verification: create DashboardModel with width=60, height=20
      4. Call View() and count lines
      5. Assert: vertical layout produces more lines than horizontal (cards are stacked)
    Expected Result: Narrow terminal shows cards stacked vertically
    Evidence: Test output captured

  Scenario: Dashboard renders cards horizontally at wide width
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Create DashboardModel with width=100, height=30
      2. Call View() and verify cards appear on same logical row
      3. Assert: line count is less than narrow layout
    Expected Result: Wide terminal shows cards side by side
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `fix(tui): dashboard cards stack vertically on narrow terminals`
  - Files: `internal/tui/views/dashboard.go`
  - Pre-commit: `make test`

---

- [x] 3. Fix setup view resize handling

  **What to do**:
  - In `setup.go`, update the `tea.WindowSizeMsg` handler to resize components:
    ```go
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.filePicker.Height = msg.Height - 8  // Reserve space for title, instructions, status
        m.pathCompleter.Input.Width = msg.Width - 6  // Account for margins and prompt
        if m.pathCompleter.Input.Width < 20 {
            m.pathCompleter.Input.Width = 20
        }
    ```
  - The `- 8` for filepicker accounts for: step title (2 lines) + instruction text (2 lines) + blank lines (2) + status bar (2 lines). This is approximate — count the actual lines in the View() method for each step that uses filepicker.
  - The `- 6` for pathcompleter accounts for: prompt prefix (`> `) + left margin (2) + right safety (2).
  - Also handle the case when `m.browsing == true` — filepicker resize events during browsing should be forwarded. Currently the `Update()` method enters the `m.browsing` block before reaching `WindowSizeMsg` handling. Fix this by moving the `WindowSizeMsg` case inside the browsing block too:
    ```go
    if m.browsing {
        switch msg := msg.(type) {
        case tea.WindowSizeMsg:
            m.width = msg.Width
            m.height = msg.Height
            m.filePicker.Height = msg.Height - 8
        }
        // ... rest of browsing logic
    }
    ```

  **Must NOT do**:
  - Don't change the setup wizard steps or flow
  - Don't modify setup view styling
  - Don't add the setup view to `propagateWindowSize()` (it's a separate mode)
  - Don't add minimum size warning to setup (handled in Task 6 at model level)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small focused change in one file
  - **Skills**: [`git-master`]
    - `git-master`: Clean atomic commit

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 4, 5, 6)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References:**
  - `internal/tui/views/settings.go:190-193` — How Settings handles WindowSizeMsg → calls `resizeLists()`. Setup should follow similar pattern.
  - `internal/tui/views/settings.go:694-696` — How Settings sizes filepicker: `m.filePicker.Height = height`. Setup should do the same.

  **API/Type References:**
  - `internal/tui/views/setup.go:48-65` — SetupModel struct with `filePicker filepicker.Model`, `pathCompleter components.PathCompleter`, `width int`, `height int`
  - `internal/tui/views/setup.go:92-124` — Current Update() method. Note the `m.browsing` block at top that swallows all messages during file browsing mode.
  - `internal/tui/views/setup.go:126-129` — Current WindowSizeMsg handler (only stores, doesn't resize)

  **Test References:**
  - `internal/tui/views/setup_test.go` — Existing setup tests to verify no regression

  **WHY Each Reference Matters:**
  - `settings.go:190-196` — Exemplar for how to properly handle resize in a view with filepicker.
  - `setup.go:92-124` — The browsing block must be understood because it intercepts ALL messages. WindowSizeMsg during browsing needs special handling.
  - `setup.go:126-129` — The exact code being modified.

  **Acceptance Criteria**:

  - [x] `setup.go` WindowSizeMsg handler sets `m.filePicker.Height`
  - [x] `setup.go` WindowSizeMsg handler sets `m.pathCompleter.Input.Width`
  - [x] WindowSizeMsg during browsing mode (`m.browsing == true`) still updates filepicker height
  - [x] PathCompleter width has minimum clamp (>= 20)
  - [x] `make test` passes

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Setup filepicker resizes on WindowSizeMsg
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestSetup
      2. Assert: all existing setup tests pass
      3. Write test: create SetupModel, send WindowSizeMsg{Width:120, Height:50}
      4. Assert: m.filePicker.Height > 0 and reflects new height
      5. Assert: m.pathCompleter.Input.Width > 0 and reflects new width
    Expected Result: Components sized after resize event
    Evidence: Test output captured

  Scenario: Setup filepicker resizes during browsing
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Create SetupModel, set m.browsing = true
      2. Send WindowSizeMsg{Width:80, Height:30}
      3. Assert: m.filePicker.Height is updated
      4. Assert: m.width and m.height are updated
    Expected Result: Resize during browsing works
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `fix(tui): setup view resizes filepicker and input on terminal resize`
  - Files: `internal/tui/views/setup.go`
  - Pre-commit: `make test`

---

- [x] 4. Fix restore viewport padding consistency

  **What to do**:
  - In `restore.go`, the viewport is sized in two places with different calculations:
    - `Update()` line 374-375: `m.viewport.Width = msg.Width`, `m.viewport.Height = msg.Height - styles.ViewChromeHeight` — After Task 1, this becomes `m.viewport.Height = msg.Height` (no subtraction since propagateWindowSize already adjusted).
    - `View()` lines 529-530: Uses `Width(m.width - 4).Height(m.height - styles.ViewChromeHeight - 4)` for styling — The `-4` accounts for border/padding of the viewport border style.
  - **The issue**: viewport content is sized to one dimension in Update(), but rendered inside a smaller container in View(). This causes scrollable content to be taller than the visible area.
  - **Fix**: Make the viewport dimensions in Update() match the actual visible area:
    ```go
    // In Update(), WindowSizeMsg handler:
    borderH := 2  // top + bottom border
    borderW := 2  // left + right border
    padding := 2  // ViewportBorder padding (1 each side from Padding in style - verify)
    m.viewport.Width = msg.Width - borderW - padding
    m.viewport.Height = msg.Height - borderH - padding
    ```
  - Verify the actual border/padding values from `styles.go:129-131` — `ViewportBorder` has `Border(lipgloss.RoundedBorder())` which adds 2 chars (1 each side) for border. Check if there's additional padding.
  - Apply the same calculation for the viewport border style in View() to ensure consistency.

  **Must NOT do**:
  - Don't change the ViewportBorder style itself
  - Don't modify other phases of the restore workflow (only the diff preview phase)
  - Don't change list sizing (backupList, fileList) — those were fixed in Task 1

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Focused fix in one section of one file
  - **Skills**: [`git-master`]
    - `git-master`: Clean atomic commit

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3, 5, 6)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References:**
  - `internal/tui/views/restore.go:369-375` — WindowSizeMsg handler with viewport sizing (Update level)
  - `internal/tui/views/restore.go:525-535` — Viewport rendering in View() with different sizing

  **API/Type References:**
  - `internal/tui/styles/styles.go:129-131` — `ViewportBorder` style definition. Has `Border(lipgloss.RoundedBorder())` + `BorderForeground`. Check for Padding().
  - `internal/tui/views/restore.go:37-38` — width/height fields stored in RestoreModel

  **Test References:**
  - `internal/tui/views/restore_test.go:100` — Uses WindowSizeMsg{100, 50}

  **WHY Each Reference Matters:**
  - `restore.go:369-375` — THE code being fixed. Must make viewport sizing match View() rendering.
  - `restore.go:525-535` — The View() rendering that uses different calculations. Must verify what border/padding values are applied.
  - `styles.go:129-131` — Source of truth for viewport border dimensions.

  **Acceptance Criteria**:

  - [x] Viewport width in Update() accounts for border/padding (same as View())
  - [x] Viewport height in Update() accounts for border/padding (same as View())
  - [x] No ViewChromeHeight references remain in restore.go
  - [x] `make test` passes

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Viewport dimensions are consistent between Update and View
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestRestore
      2. Assert: all existing restore tests pass
      3. Write test: create RestoreModel, send WindowSizeMsg{Width:80, Height:24}
      4. Verify: m.viewport.Width and m.viewport.Height match the visible area
         (accounting for border/padding consistently)
    Expected Result: Viewport content fits exactly in visible area
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `fix(tui): consistent viewport padding in restore diff preview`
  - Files: `internal/tui/views/restore.go`
  - Pre-commit: `make test`

---

- [x] 5. Make PasswordInput responsive & wrap FilePicker width

  **What to do**:

  **Part A — Responsive PasswordInput:**
  - In `components/passwordinput.go`, change `NewPasswordInput()` to NOT set a fixed width. Instead, let the caller set width after creation:
    ```go
    func NewPasswordInput(placeholder string) textinput.Model {
        ti := textinput.New()
        ti.Placeholder = placeholder
        ti.EchoMode = textinput.EchoPassword
        ti.EchoCharacter = '•'
        // Width NOT set here — callers set it based on terminal width
        ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
        ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
        return ti
    }
    ```
  - In `restore.go` WindowSizeMsg handler, add password input width update:
    ```go
    pw := msg.Width - 6  // Account for margins and prompt
    if pw < 20 { pw = 20 }
    if pw > 60 { pw = 60 }  // Cap max width for readability
    m.passwordInput.Width = pw
    ```
  - In `backuplist.go` WindowSizeMsg handler, add the same for its password input (if it has one — verify).
  - Set a reasonable initial width (40) in NewRestore/NewBackupList constructors as default until first WindowSizeMsg.

  **Part B — FilePicker width wrapping:**
  - The bubbles `filepicker.Model` does NOT have a Width API. Width must be constrained via lipgloss wrapping.
  - In `settings.go` View(), when rendering the filepicker, wrap it:
    ```go
    case stateFilePickerActive:
        // ...
        fpView := m.filePicker.View()
        if m.width > 0 {
            fpView = lipgloss.NewStyle().MaxWidth(m.width - 4).Render(fpView)
        }
        b.WriteString(fpView)
    ```
  - In `setup.go` View(), when rendering the filepicker (in StepAddFiles and StepAddFolders), apply the same wrapping:
    ```go
    fpView := m.filePicker.View()
    if m.width > 0 {
        fpView = lipgloss.NewStyle().MaxWidth(m.width - 4).Render(fpView)
    }
    s.WriteString(fpView)
    ```

  **Must NOT do**:
  - Don't modify filepicker.Model internals (it's a library)
  - Don't set a global default width on the textinput (let callers control it)
  - Don't change PasswordInput styling (only width management)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small changes across 4-5 files, all straightforward
  - **Skills**: [`git-master`]
    - `git-master`: Clean atomic commit

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3, 4, 6)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References:**
  - `internal/tui/components/passwordinput.go:9-18` — Current `NewPasswordInput()` with hardcoded `ti.Width = 40`
  - `internal/tui/views/settings.go:733-739` — How settings renders filepicker in View() (stateFilePickerActive)
  - `internal/tui/views/setup.go:443-445` — How setup renders filepicker during StepAddFiles browsing
  - `internal/tui/views/setup.go:470-472` — How setup renders filepicker during StepAddFolders browsing

  **API/Type References:**
  - `internal/tui/views/restore.go:369-375` — Restore WindowSizeMsg handler (where to add password width update)
  - `internal/tui/views/backuplist.go:99-102` — BackupList WindowSizeMsg handler

  **External References:**
  - bubbles filepicker has NO Width field — confirmed via source code analysis. Must use lipgloss MaxWidth wrapping.

  **WHY Each Reference Matters:**
  - `passwordinput.go` — THE file being modified. Remove hardcoded width.
  - `restore.go` and `backuplist.go` — These are the consumers that need to set password input width dynamically.
  - `settings.go:733-739` and `setup.go:443-445` — These are where filepicker View() output is rendered, where MaxWidth wrapping needs to be added.

  **Acceptance Criteria**:

  - [x] `passwordinput.go` no longer has `ti.Width = 40`
  - [x] Restore view updates password input width on WindowSizeMsg
  - [x] Password input width is capped at max 60 and min 20
  - [x] FilePicker output wrapped with `MaxWidth` in settings and setup views
  - [x] `make test` passes

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Password input width adapts to terminal size
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestRestore
      2. Assert: all existing tests pass
      3. Create RestoreModel, send WindowSizeMsg{Width:120, Height:30}
      4. Assert: m.passwordInput.Width > 40 and <= 60
      5. Send WindowSizeMsg{Width:45, Height:20}
      6. Assert: m.passwordInput.Width < 45 and >= 20
    Expected Result: Password input width scales with terminal
    Evidence: Test output captured

  Scenario: FilePicker output is width-constrained
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestSettings
      2. Assert: all existing tests pass
      3. Verify: filepicker rendering in View() is wrapped with MaxWidth
    Expected Result: No horizontal overflow from filepicker
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `fix(tui): responsive password input width and filepicker width constraint`
  - Files: `internal/tui/components/passwordinput.go`, `internal/tui/views/restore.go`, `internal/tui/views/backuplist.go`, `internal/tui/views/settings.go`, `internal/tui/views/setup.go`
  - Pre-commit: `make test`

---

- [x] 6. Add minimum terminal size warning

  **What to do**:
  - In `view.go`, at the top of the `View()` method (after the `m.setupMode` and `m.quitting` checks), add a minimum size check:
    ```go
    if m.width < styles.MinTerminalWidth || m.height < styles.MinTerminalHeight {
        return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
            "Terminal too small\n"+
            fmt.Sprintf("Minimum: %dx%d", styles.MinTerminalWidth, styles.MinTerminalHeight)+"\n"+
            fmt.Sprintf("Current: %dx%d", m.width, m.height),
        )
    }
    ```
  - This replaces the normal TUI content with a centered warning when the terminal is below 40x15.
  - Import `fmt` in `view.go` if not already imported.
  - The check uses `MinTerminalWidth` and `MinTerminalHeight` constants from Task 1.
  - NOTE: This applies only to the main TUI mode, NOT to setup mode (setup has its own View rendering).
  - Also add the same check in setup mode's View() method in `setup.go` (since setup renders independently):
    ```go
    func (m SetupModel) View() string {
        if m.width > 0 && m.height > 0 &&
           (m.width < styles.MinTerminalWidth || m.height < styles.MinTerminalHeight) {
            return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
                "Terminal too small\n"+
                fmt.Sprintf("Minimum: %dx%d", styles.MinTerminalWidth, styles.MinTerminalHeight)+"\n"+
                fmt.Sprintf("Current: %dx%d", m.width, m.height),
            )
        }
        // ... rest of existing View()
    ```
  - The `m.width > 0 && m.height > 0` guard prevents showing the warning on initial render before first WindowSizeMsg.

  **Must NOT do**:
  - Don't make the minimum size configurable
  - Don't add resize animation
  - Don't prevent the TUI from starting at small sizes (just show warning, allow resize to fix it)
  - Don't log a warning or exit — just render the message

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple conditional added to two View() methods
  - **Skills**: [`git-master`]
    - `git-master`: Clean atomic commit

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3, 4, 5)
  - **Blocks**: Task 7
  - **Blocked By**: Task 1

  **References**:

  **Pattern References:**
  - `internal/tui/help.go:44` — Existing small terminal fallback: `if width < 40 || height < 15` returns plain text. Our minimum size warning follows the same pattern.
  - `internal/tui/views/helpers.go:180-182` — `PlaceOverlay()` using `lipgloss.Place()` for centering. Use the same approach for the warning message.

  **API/Type References:**
  - `internal/tui/view.go:11-63` — Main View() method where the check is added (after quitting check on line 16)
  - `internal/tui/views/setup.go:351-535` — Setup View() method where the check is added (at top)
  - `internal/tui/styles/styles.go` — `MinTerminalWidth` and `MinTerminalHeight` constants (from Task 1)

  **WHY Each Reference Matters:**
  - `view.go:11-63` — THE function being modified. Warning goes before any content rendering.
  - `setup.go:351` — Setup has independent View() that needs separate minimum check.
  - `help.go:44` — Demonstrates the existing pattern for small terminal handling.

  **Acceptance Criteria**:

  - [x] Terminal at 39x14 shows "Terminal too small" warning with current and minimum dimensions
  - [x] Terminal at 40x15 shows normal TUI content (no warning)
  - [x] Warning is centered in the available space
  - [x] Setup mode also shows the warning below minimum size
  - [x] Warning disappears when terminal is resized above minimum
  - [x] `make test` passes

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Minimum size warning shown at 39x14
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Create Model, send WindowSizeMsg{Width:39, Height:14}
      2. Call View()
      3. Assert: output contains "Terminal too small"
      4. Assert: output contains "Minimum: 40x15"
      5. Assert: output contains "Current: 39x14"
    Expected Result: Warning message displayed
    Evidence: Test output captured

  Scenario: Normal rendering at exactly minimum size 40x15
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Create Model, send WindowSizeMsg{Width:40, Height:15}
      2. Call View()
      3. Assert: output does NOT contain "Terminal too small"
      4. Assert: output contains "DotKeeper" (normal title)
    Expected Result: Normal TUI content shown
    Evidence: Test output captured

  Scenario: Warning clears when terminal resized above minimum
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Create Model, send WindowSizeMsg{Width:30, Height:10}
      2. Call View() → assert contains "Terminal too small"
      3. Send WindowSizeMsg{Width:80, Height:24}
      4. Call View() → assert does NOT contain "Terminal too small"
    Expected Result: Warning disappears after resize
    Evidence: Test output captured

  Scenario: Setup mode also shows warning
    Tool: Bash (go test)
    Preconditions: Task 1 complete
    Steps:
      1. Create SetupModel, send WindowSizeMsg{Width:35, Height:10}
      2. Call View()
      3. Assert: output contains "Terminal too small"
    Expected Result: Setup mode has minimum size warning
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): show warning when terminal is below minimum size (40x15)`
  - Files: `internal/tui/view.go`, `internal/tui/views/setup.go`
  - Pre-commit: `make test`

---

- [x] 7. Comprehensive resize test matrix

  **What to do**:
  - Create a new test file `internal/tui/views/resize_test.go` with comprehensive resize tests.
  - Also update existing test files where resize-specific assertions are needed.

  **Test Categories:**

  **A — No-Panic Tests (all views, all sizes):**
  ```go
  func TestAllViewsNoPanicOnResize(t *testing.T) {
      sizes := []tea.WindowSizeMsg{
          {Width: 0, Height: 0},       // Zero (terminal multiplexer init)
          {Width: 1, Height: 1},       // Absurdly small
          {Width: 39, Height: 14},     // Below minimum
          {Width: 40, Height: 15},     // Exact minimum
          {Width: 80, Height: 24},     // Standard
          {Width: 200, Height: 100},   // Very large
          {Width: 300, Height: 10},    // Ultra-wide
          {Width: 40, Height: 100},    // Ultra-tall
      }
      // For each view (Dashboard, BackupList, Restore, Settings, Logs, Setup):
      //   For each size: send WindowSizeMsg, call View(), assert no panic
  }
  ```

  **B — Chrome Height Consistency Test:**
  ```go
  func TestViewHeightConsistency(t *testing.T) {
      // Send WindowSizeMsg{80, 24} to each view through propagateWindowSize
      // Verify all views receive the same usable height
      // Verify no double-subtraction occurs
  }
  ```

  **C — Dashboard Responsive Layout Tests:**
  ```go
  func TestDashboardResponsiveLayout(t *testing.T) {
      // width=100: cards horizontal (fewer lines)
      // width=60: cards vertical (more lines)
      // width=40: cards vertical + buttons vertical
  }
  ```

  **D — Minimum Size Warning Tests:**
  ```go
  func TestMinTerminalSizeWarning(t *testing.T) {
      // width=39: shows warning
      // width=40: normal content
      // height=14: shows warning
      // height=15: normal content
  }
  ```

  **E — Rapid Resize Sequence Test:**
  ```go
  func TestRapidResizeSequence(t *testing.T) {
      // Send 10 resize events in succession
      // Verify final state matches last resize dimensions
      // Verify no accumulated drift
  }
  ```

  **F — Setup Filepicker Resize Test:**
  ```go
  func TestSetupFilepickerResize(t *testing.T) {
      // Send WindowSizeMsg, verify filepicker.Height updated
      // Send during browsing mode, verify still updated
  }
  ```

  **G — Update existing tests:**
  - Add resize assertions to `dashboard_test.go`, `setup_test.go`, `restore_test.go` where relevant
  - Ensure existing tests at 80x24, 100x50, 40x24 still pass

  **Must NOT do**:
  - Don't create tests that require human interaction
  - Don't test visual appearance (pixel-perfect rendering) — test structural properties (line count, contains/not-contains)
  - Don't add tests for features outside scope (help overlay, status bar truncation)
  - Don't mock terminal — use direct model instantiation and message sending

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Comprehensive test file covering all views and many size combinations. Requires understanding of all views.
  - **Skills**: [`git-master`]
    - `git-master`: Clean atomic commit
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Not applicable — Go tests, not frontend

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (solo, after all fixes)
  - **Blocks**: None (final task)
  - **Blocked By**: Tasks 2, 3, 4, 5, 6

  **References**:

  **Pattern References:**
  - `internal/tui/views/dashboard_test.go` — Existing test patterns. Uses `views.NewDashboard(cfg)`, sends `tea.WindowSizeMsg`, calls `View()`, uses `stripANSI()` and `strings.Contains()`.
  - `internal/tui/views/testhelpers_test.go` — Shared test helpers. Contains `stripANSI()`, `testConfig()`, mock types. Use these in new tests.
  - `internal/tui/views/restore_test.go` — Restore test patterns with WindowSizeMsg

  **Test References (existing to not regress):**
  - `internal/tui/views/dashboard_test.go:18-19` — Test at 80x24
  - `internal/tui/views/dashboard_test.go:46` — Test at 40x24
  - `internal/tui/views/restore_test.go:100` — Test at 100x50

  **WHY Each Reference Matters:**
  - `dashboard_test.go` — Pattern to follow for test structure, helpers used, assertion style.
  - `testhelpers_test.go` — Must use existing helpers (stripANSI, testConfig) instead of creating duplicates.
  - All existing test files — Must verify no regressions from changes in Tasks 1-6.

  **Acceptance Criteria**:

  - [x] New file `internal/tui/views/resize_test.go` created
  - [x] All no-panic tests pass (8 sizes × 6 views = 48 sub-tests)
  - [x] Chrome height consistency test passes
  - [x] Dashboard responsive layout tests pass for 3+ width breakpoints
  - [x] Minimum size warning tests pass (below/at/above minimum)
  - [x] Rapid resize sequence test passes (10 events, no drift)
  - [x] Setup filepicker resize test passes
  - [x] `make test` passes (all new + existing tests)
  - [x] `go test ./internal/tui/... -v -count=1` — all PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Full resize test suite passes
    Tool: Bash
    Preconditions: Tasks 1-6 complete
    Steps:
      1. Run: go test ./internal/tui/... -v -count=1
      2. Assert: exit code 0
      3. Assert: output contains "PASS" for resize_test.go
      4. Assert: output does NOT contain "FAIL"
      5. Count test cases: grep -c "=== RUN" output
      6. Assert: at least 48 test sub-cases run (no-panic matrix)
    Expected Result: All tests pass, >= 48 sub-tests run
    Evidence: Full test output captured

  Scenario: No regression in existing tests
    Tool: Bash
    Preconditions: Tasks 1-6 complete
    Steps:
      1. Run: make test
      2. Assert: exit code 0
      3. Assert: coverage report generated
    Expected Result: Zero regressions
    Evidence: Test output and coverage report captured
  ```

  **Commit**: YES
  - Message: `test(tui): comprehensive resize test matrix covering all views and sizes`
  - Files: `internal/tui/views/resize_test.go`, `internal/tui/views/dashboard_test.go` (if updated), `internal/tui/views/setup_test.go` (if updated), `internal/tui/views/restore_test.go` (if updated)
  - Pre-commit: `make test`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `refactor(tui): unify chrome height and extract breakpoint constants` | styles.go, model.go, update.go, backuplist.go, restore.go, settings.go, logs.go, dashboard.go, tabbar.go | `make test` |
| 2 | `fix(tui): dashboard cards stack vertically on narrow terminals` | dashboard.go | `make test` |
| 3 | `fix(tui): setup view resizes filepicker and input on terminal resize` | setup.go | `make test` |
| 4 | `fix(tui): consistent viewport padding in restore diff preview` | restore.go | `make test` |
| 5 | `fix(tui): responsive password input width and filepicker width constraint` | passwordinput.go, restore.go, backuplist.go, settings.go, setup.go | `make test` |
| 6 | `feat(tui): show warning when terminal is below minimum size (40x15)` | view.go, setup.go | `make test` |
| 7 | `test(tui): comprehensive resize test matrix covering all views and sizes` | resize_test.go + updated test files | `make test` |

---

## Success Criteria

### Verification Commands
```bash
make test                                        # Expected: PASS, 0 failures
go test ./internal/tui/... -v -count=1           # Expected: all PASS, >= 48 resize sub-tests
go vet ./internal/tui/...                        # Expected: 0 issues
go build ./...                                   # Expected: builds successfully
grep -rn "ViewChromeHeight" internal/tui/        # Expected: 0 matches (removed)
grep -n "width < 80\|width >= 80\|width >= 60" internal/tui/views/dashboard.go internal/tui/components/tabbar.go  # Expected: 0 matches
```

### Final Checklist
- [x] All "Must Have" present (unified chrome, responsive dashboard, setup resize, min size warning, constants, tests)
- [x] All "Must NOT Have" absent (no styling changes, no config option for min size, no tab bar icons, no files outside internal/tui/)
- [x] All tests pass (existing + new)
- [x] 40x15 terminal renders correctly
- [x] Below 40x15 shows centered warning
- [x] 80x24 terminal renders same as before (no regression)
