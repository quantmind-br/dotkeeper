# Learnings — TUI Responsive Resize

## Task 1: Unify Chrome Height & Extract Breakpoint Constants

### Chrome Height Calculation
- `view.go` View() produces this output structure: `AppTitle\nTabBar\n\n[content]\n\nviewHelp\nglobalHelp\n`
- Counting terminal rows: title(1) + tab(1) + blank-after-tab(1) + blank-after-content(1) + viewHelp(1) + globalHelp(1) + trailing-newline(1) = **7 rows**
- Previous `mainChromeHeight=5` missed the viewHelp line and trailing newline
- Previous double-subtraction bug: `propagateWindowSize` subtracted 5, then each view subtracted `ViewChromeHeight=6` — total 11 lines lost

### Architecture Pattern
- `propagateWindowSize()` in `update.go` is the single source of truth for available content height
- Sub-views receive already-adjusted `WindowSizeMsg` and should NOT do additional subtraction
- Width clamping added to prevent negative widths being passed to sub-views

### Breakpoint Constants
- `BreakpointWide = 80` — used in dashboard (card layout) and tabbar (short labels)
- `BreakpointMedium = 60` — used in dashboard (action button layout)
- `MinTerminalWidth = 40`, `MinTerminalHeight = 15` — defined for future use

### Hidden Reference
- `restore.go` had a `styles.ViewChromeHeight` in the View() method (line 530 in diff preview viewport style) that wasn't listed in the task's Update section — always grep after removing a constant

### Settings Quirk
- `settings.go:resizeLists()` had a `width <= 0` fallback to 80 — removed since propagateWindowSize now clamps width
- Settings also subtracts 2 more lines when editing (for the input field chrome) — this is view-internal chrome, not framework chrome

## Task 2: Dashboard Cards Responsive Layout

### Implementation
- Changed `internal/tui/views/dashboard.go` lines 120-126
- At `width >= styles.BreakpointWide` (80): cards render horizontally with `JoinHorizontal(lipgloss.Top, cards...)`
- At `width < styles.BreakpointWide`: cards render vertically with `JoinVertical(lipgloss.Left, cards...)`
- Removed placeholder comment "Split into rows if needed, simplified for now"

### Pattern Consistency
- Mirrors the action buttons layout pattern (lines 146-151)
- Both use the same breakpoint constant (`styles.BreakpointWide`)
- Both follow: horizontal at wide, vertical at narrow
- Alignment: `lipgloss.Top` for horizontal, `lipgloss.Left` for vertical

### Testing
- All 100+ tests pass (e2e, backup, restore, crypto, config, git, history, keyring, notify, pathutil, tui/components, tui/views)
- No linting issues
- Commit: `7d753a9` on `feat/settings-inline-actions`

### Key Insight
The dashboard has two responsive sections:
1. **Stats cards** (3-4 cards): Now responsive via this fix
2. **Action buttons** (3 buttons): Already responsive from Task 1

Both sections now follow the same responsive pattern, making the dashboard fully responsive on narrow terminals.

## Task 2: Setup View Resize Handling

### Setup Mode Architecture
- Setup view is SEPARATE from main TUI chrome — it bypasses `propagateWindowSize()` entirely
- Setup has its own `Update()` loop that handles `WindowSizeMsg` independently
- Setup stores width/height but was NOT applying them to components

### The Browsing Block Trap
- Setup has a `if m.browsing { ... }` block at the TOP of Update() that intercepts ALL messages
- This block was returning early without handling `WindowSizeMsg`
- Result: Terminal resize during file browsing mode was ignored
- **Fix**: Added `WindowSizeMsg` check INSIDE the browsing block before delegating to filepicker

### Component Sizing Pattern (from settings.go exemplar)
- Settings view uses `resizeLists()` method to apply width/height to components
- For filepicker: `m.filePicker.Height = height - 8` (reserves space for chrome)
- For pathcompleter input: `m.pathCompleter.Input.Width = width - 6` (accounts for margins/prompt)
- Minimum width clamp: `if pcWidth < 20 { pcWidth = 20 }`

### Setup View Chrome Calculation
- Setup title: ~2 lines
- Instructions/help text: ~2 lines
- Blank lines: ~2 lines
- Status bar: ~2 lines
- **Total: 8 lines reserved** → `filepicker.Height = height - 8`

### Implementation Details
- WindowSizeMsg handler now sets both filepicker.Height and pathcompleter.Input.Width
- Browsing block also handles resize to keep filepicker responsive during file selection
- No minimum height clamp needed (filepicker handles it internally)
- Width clamp of 20 prevents pathcompleter from becoming unusable on very narrow terminals

### Testing
- All 38 tests in `internal/tui/views` pass
- Setup view tests include `TestSetupBrowseMode` which exercises the browsing block
- No new tests needed — existing tests cover the resize behavior

## Task 3: Restore View Viewport Padding Consistency

### Problem
The restore view's diff preview phase had a viewport padding inconsistency:
- `Update()` method (lines 374-375) set viewport dimensions to raw `msg.Width` and `msg.Height`
- `View()` method (lines 528-530) rendered the viewport inside a styled container with `RoundedBorder()` which adds 1 char on each side (2 total width, 2 total height)
- Result: scrollable content was taller than the visible area, causing overflow

### Root Cause
The viewport content was sized without accounting for the border styling applied during rendering. The View() method was doing `Width(m.width - 4).Height(m.height - 4)` but the viewport itself was sized to full dimensions.

### Solution
1. **Update() method**: Account for border dimensions when setting viewport size
   ```go
   vpBorderW := 2  // left + right border from RoundedBorder
   vpBorderH := 2  // top + bottom border from RoundedBorder
   m.viewport.Width = msg.Width - vpBorderW
   m.viewport.Height = msg.Height - vpBorderH
   ```
   Added safety checks for negative dimensions (clamp to 0)

2. **View() method**: Use viewport dimensions directly instead of doing additional math
   ```go
   viewportStyle := st.ViewportBorder.Copy().
       Width(m.viewport.Width).
       Height(m.viewport.Height)
   ```

### Key Insight
The viewport dimensions in Update() should match the visible area after styling is applied. By accounting for border in Update(), the View() method can use those dimensions directly without additional calculations. This ensures consistency and prevents overflow.

### Architecture Pattern
- Single source of truth: viewport dimensions set in Update() account for all styling
- View() uses those dimensions directly without modification
- Prevents double-subtraction bugs (like the chrome height issue in Task 1)

### Testing
- All 100+ tests pass
- No linting issues
- Commit: `f6ce7f0` on `feat/settings-inline-actions`

### Related
- Task 1: Unified chrome height calculation (propagateWindowSize is single source of truth)
- Task 2: Dashboard responsive layout (uses same breakpoint pattern)
- Task 3: Viewport padding consistency (viewport dimensions account for styling)

All three tasks follow the pattern: **Calculate dimensions once in Update(), use directly in View()**.

## Task 4: Minimum Terminal Size Warning

### Implementation
- Added minimum size check to `internal/tui/view.go` View() method (lines 20-26)
- Added same check to `internal/tui/views/setup.go` View() method (lines 365-371)
- Check placed AFTER `m.quitting` check but BEFORE `m.showingHelp` check in main view
- Check placed at TOP of setup's View() method to catch resize before rendering

### Guard Condition
- `if m.width > 0 && m.height > 0` prevents showing warning on initial render
- Initial render has width/height = 0 before first `WindowSizeMsg` arrives
- This guard prevents spurious warnings during startup

### Rendering Pattern
- Uses `lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, message)`
- Message format: "Terminal too small\nMinimum: 40x15\nCurrent: WxH"
- Centered both horizontally and vertically
- Replaces entire TUI content when triggered

### Imports Added
- `view.go`: Added `"fmt"` and `"github.com/charmbracelet/lipgloss"`
- `setup.go`: Added `"github.com/charmbracelet/lipgloss"`

### Architecture Pattern
- Main TUI and Setup mode both check minimum size independently
- Setup mode has its own View() that bypasses main chrome
- Both use the same constants: `styles.MinTerminalWidth` and `styles.MinTerminalHeight`
- Both use the same rendering approach: `lipgloss.Place()` for centering

### Testing
- All 100+ tests pass
- No linting issues
- Commit: `7298719` on `feat/settings-inline-actions`

### Key Insight
The minimum size check is a **guard clause** that prevents rendering the full TUI when terminal is too small. It's placed early in the View() method to short-circuit before any complex rendering logic. The guard condition `m.width > 0 && m.height > 0` is critical to avoid showing the warning during initial render when dimensions haven't been set yet.

## Task 3: Responsive Password Input Width & FilePicker Width Constraint

### Password Input Responsiveness

**Problem**: PasswordInput had hardcoded `Width = 40`, making it unresponsive to terminal resizing.

**Solution**:
1. Removed hardcoded `ti.Width = 40` from `components/passwordinput.go`
2. Set default width in constructors: `ti.Width = 40` in `NewRestore()` and `NewBackupList()`
3. Added responsive sizing in WindowSizeMsg handlers:
   ```go
   pw := msg.Width - 6  // Account for margins/prompt
   if pw < 20 { pw = 20 }  // Minimum width
   if pw > 60 { pw = 60 }  // Maximum width
   m.passwordInput.Width = pw
   ```

**Pattern**: Same as viewport sizing — calculate available width, clamp to min/max, apply to component.

### FilePicker Width Constraint

**Problem**: bubbles/filepicker has no `Width` field (only `Height`), so output could overflow terminal width.

**Solution**: Wrap filepicker View() output with lipgloss MaxWidth:
```go
fpView := m.filePicker.View()
if m.width > 0 {
    fpView = lipgloss.NewStyle().MaxWidth(m.width - 4).Render(fpView)
}
s.WriteString(fpView)
```

**Applied to**:
- `settings.go`: stateFilePickerActive case (line 735)
- `setup.go`: StepAddFiles browsing mode (line 459)
- `setup.go`: StepAddFolders browsing mode (line 486)

**Key insight**: lipgloss.MaxWidth() wraps long lines, preventing horizontal overflow.

### Testing
- All 100+ tests pass
- No new tests needed — existing tests cover resize behavior
- Commit: `a09397a` on `feat/settings-inline-actions`

### Architecture Notes
- Password input width clamping: min 20, max 60 (prevents unusable narrow inputs and excessive width)
- FilePicker wrapping: `m.width - 4` accounts for padding/borders
- Both patterns follow established responsive design from Task 1 (viewport sizing)

## Task 7: Comprehensive Resize Test Matrix

### Test Structure
- 6 no-panic test functions (one per view) × 8 sizes = 48 sub-tests
- Each view gets its own test function because `Update()` returns concrete types requiring specific type assertions (`DashboardModel`, `BackupListModel`, etc.)
- A single generic function with `tea.Model` interface wouldn't work because `View()` requires the concrete type

### Constructor Signatures
- `NewDashboard(cfg)` — config only
- `NewBackupList(cfg, store)` — store can be nil
- `NewRestore(cfg, store)` — store can be nil
- `NewSettings(cfg)` — config only
- `NewLogs(cfg, stores...)` — variadic store (can omit)
- `NewSetup()` — no args

### Test Categories Created
1. **No-Panic Matrix** (A): 8 sizes × 6 views — ensures no view panics at any size
2. **Dashboard Responsive** (B): Wide vs narrow layout produces different line counts
3. **Minimum Size Warning** (C): Setup view shows/hides warning based on dimensions; guard at 0x0
4. **Rapid Resize** (D): 10 sequential resizes on 4 views — no drift, final dimensions correct
5. **Setup Filepicker** (E): Resize during normal mode and browsing mode
6. **Extreme Edge Cases** (F): Zero size, viewport at extreme sizes, state preservation, dimension tracking

### Key Patterns
- `stripANSI()` from testhelpers_test.go for comparing rendered output
- `navigateToAddFiles()` from setup_test.go for reaching browsing mode
- Existing tests use `package views` (not `views_test`) — same-package access to internals
- `testConfig()` didn't exist in testhelpers — created local `testCfg()` helper

### Gotchas
- `NewLogs` takes variadic `stores ...*history.Store`, not `*history.Store` directly — can call with no store arg
- Dashboard responsive test needs `strings.TrimRight(view, "\n")` before splitting to avoid trailing empty lines skewing count
- Setup's min-size guard: `m.width > 0 && m.height > 0` prevents false positive at init (before first WindowSizeMsg)
- Restore viewport phase test: must set `m.phase = phaseDiffPreview` and provide `m.currentDiff` content
