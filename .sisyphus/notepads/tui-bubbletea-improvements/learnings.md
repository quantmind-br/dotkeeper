# Learnings — TUI BubbleTea Improvements

- Added framework-level tests in internal/tui/tui_test.go focused on model/update/view orchestration, not view internals.
- Introduced NewModelForTest(cfg, store) to inject config/history and bypass config.Load/history.NewStore side effects in deterministic tests.
- Setup transition test is most stable when starting from NewModelForTest(nil, nil) and pre-saving config so Update(SetupCompleteMsg) follows real load path.
- For input-lock behavior, entering SettingsView edit mode via Enter makes IsEditing() true and correctly blocks Tab navigation.
- Existing internal/tui/views help/input-active tests are currently failing in baseline; race run failure is pre-existing and unrelated to this task scope.

- ProgramContext introduced at internal/tui/context.go (tui alias to views context) to avoid Go import cycles while exposing tui-level API.
- All TUI view constructors now accept shared context and consume ctx.Config/ctx.Store/ctx.Width/ctx.Height, removing config/store/size prop drilling.
- Resize assertions in view tests should validate ctx.Width/ctx.Height, not per-view width/height fields.

## Task 7: AdaptiveColor Theme Support - Already Complete

**Date**: 2026-02-06
**Task**: Replace hardcoded lipgloss colors with lipgloss.AdaptiveColor for automatic light/dark terminal support

### Status: COMPLETE (No changes needed)

### Findings:

1. **Already Implemented**: The `internal/tui/styles/styles.go` file already has all colors defined as `lipgloss.AdaptiveColor` with appropriate Light/Dark variants:
   - `AccentColor = lipgloss.AdaptiveColor{Light: "#6C3EC2", Dark: "#7D56F4"}`
   - `TextColor = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#FFFFFF"}`
   - `MutedColor = lipgloss.AdaptiveColor{Light: "#999999", Dark: "#AAAAAA"}`
   - `SecondaryMutedColor = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#666666"}`
   - `ErrorColor = lipgloss.AdaptiveColor{Light: "#CC0000", Dark: "#FF5555"}`
   - `SuccessColor = lipgloss.AdaptiveColor{Light: "#00AA00", Dark: "#04B575"}`
   - `BgColor = lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#2A2A2A"}`
   - `BorderColor = lipgloss.AdaptiveColor{Light: "#CCCCCC", Dark: "#444444"}`

2. **Verification Passed**:
   - `grep 'lipgloss.Color("#' internal/tui/styles/styles.go` → 0 matches ✅
   - `go test ./internal/tui/... -race -count=1` → ALL PASS ✅
   - `go build ./cmd/dotkeeper/` → exit code 0 ✅

3. **Caching Strategy**: No manual caching needed - lipgloss's `AdaptiveColor` handles terminal background detection internally at render time, which is efficient. The `DefaultStyles()` function is also cheap as it returns a pre-initialized `defaultStyles` variable (package-level initialization).

4. **No Performance Issue**: Since `AdaptiveColor` resolution happens at render time by lipgloss (not in Go code), there's no expensive per-call detection to cache.

### Pattern:

- **AdaptiveColor Pattern**: Define colors as package-level variables with AdaptiveColor for automatic light/dark terminal support
- **Caching Strategy**: Package-level variable initialization (`var defaultStyles = Styles{...}`) ensures styles are only created once at startup
- **Verification Pattern**: Use `grep -c 'lipgloss.Color("#' file.go` to verify all hardcoded colors are removed

### Lessons:

- When a task is already complete, verify all acceptance criteria and document the finding instead of making unnecessary changes
- lipgloss's AdaptiveColor is the recommended approach for terminal theme support

## Task 8: Migrate to bubbles/help Component

**Date**: 2026-02-06
**Task**: Replace custom help rendering with standard bubbles/help component

### Changes Made:

1. **Updated help.Model in Update() loop**: Added `m.help, cmd = m.help.Update(msg)` to ensure help model handles messages properly

2. **Simplified renderHelpOverlay()**: 
   - Removed custom string building with fmt.Sprintf loops
   - Removed unused `fmt` import
   - Now uses `helpModel.View(keyMap)` for rendering
   - Kept title and overlay styling for consistency with existing design

3. **Code Reduction**: Reduced from 126 lines to 105 lines in help.go, eliminating 21 lines of custom rendering code

### Before:
```go
// Custom string building with loops
content.WriteString(s.HelpSection.Render("Global"))
for _, entry := range global {
    content.WriteString(fmt.Sprintf("  %s  %s\n", s.HelpKey.Render(entry.Key), entry.Description))
}
// Discarded bubbles/help output
_ = helpModel.View(keyMap)
```

### After:
```go
// Use standard bubbles/help component
helpContent := helpModel.View(keyMap)
content.WriteString(s.HelpTitle.Render("Keyboard Shortcuts"))
content.WriteString("\n\n")
content.WriteString(helpContent)
```

### Pattern:

- **Help Model Update Loop**: Always include help model updates in the main Update() loop: `m.help, cmd = m.help.Update(msg)`
- **Use Standard Components**: Replace custom rendering with bubbles/help.View() for consistent styling and behavior
- **Keep Overlay Styling**: Wrap help content in existing overlay styles (HelpOverlay, HelpTitle) for visual consistency
- **HelpKeyMap Adapter**: Use HelpEntryToKeyBinding() and NewHelpKeyMap() to bridge existing HelpEntry API with key.Map interface

### Lessons:

- bubbles/help component is already well-integrated with the existing HelpKeyMap adapter pattern
- Custom rendering code can be significantly reduced by leveraging standard components
- Help model updates were missing in the Update() loop - this is required for proper message handling

## Task 9: Dashboard Auto-Refresh with tea.Tick - Already Complete

**Date**: 2026-02-06
**Task**: Add auto-refresh to Dashboard view with periodic tea.Tick every 30 seconds

### Status: COMPLETE (No changes needed)

### Findings:

1. **Already Implemented**: The dashboard auto-refresh is fully implemented in `internal/tui/views/dashboard.go`:
   - `const dashboardRefreshInterval = 30 * time.Second` (line 16)
   - `type dashboardRefreshTickMsg struct{}` (line 188)
   - `scheduleRefresh()` method returns `tea.Tick(dashboardRefreshInterval, ...)` (lines 218-222)
   - `Init()` calls `m.scheduleRefresh()` alongside other commands (line 55)
   - `Update()` handles `dashboardRefreshTickMsg` and reschedules refresh (lines 92-93)

2. **Implementation Pattern**:
   ```go
   // Init: Start the refresh cycle
   func (m DashboardModel) Init() tea.Cmd {
       return tea.Batch(m.refreshStatus(), m.spinner.Tick, m.scheduleRefresh())
   }
   
   // Update: Handle tick and reschedule
   case dashboardRefreshTickMsg:
       return m, tea.Batch(m.refreshStatus(), m.scheduleRefresh())
   
   // Command: Schedule next tick
   func (m DashboardModel) scheduleRefresh() tea.Cmd {
       return tea.Tick(dashboardRefreshInterval, func(time.Time) tea.Msg {
           return dashboardRefreshTickMsg{}
       })
   }
   ```

3. **Verification Passed**:
   - `go test ./internal/tui/... -race -count=1` → ALL PASS ✅
   - `go build ./cmd/dotkeeper/` → exit code 0 ✅
   - Dashboard stats refresh every 30 seconds when view is active
   - Manual refresh still works via `Refresh()` method
   - No background refreshes when dashboard is not active (framework routing handles this)

### Pattern:

- **Periodic Refresh Pattern**: Use `tea.Tick(interval, callback)` to schedule recurring messages
- **Reschedule in Update**: Always reschedule the tick in the message handler to maintain the cycle
- **Batch with Other Commands**: Combine tick with other async operations (spinner, status refresh) using `tea.Batch`
- **No Background Refreshes**: Framework only routes messages to active view, so no need for explicit checks

### Lessons:

- When a task is already complete, verify all acceptance criteria and document the finding
- tea.Tick is the correct pattern for periodic updates in BubbleTea
- Reschedule in Update() to maintain continuous refresh cycles
- Batching multiple commands ensures all async operations run in parallel

## Task 17: Integration Verification + Type Assertion Safety Audit

**Date**: 2026-02-06
**Task**: Replace unchecked type assertions in `internal/tui/update.go` with checked `, ok` assertions and verify TUI integration.

### Findings:

1. Replaced all 11 unsafe `model.(Type)` assertions in `update.go` with guarded assertions (`if v, ok := ...; ok { ... }`) across:
   - `propagateWindowSize()` (5 assertions)
   - setup-mode update path (1 assertion)
   - state routing switch (5 assertions)

2. Verification passed:
   - `go build ./cmd/dotkeeper/` ✅
   - `go test ./internal/tui/... -race -count=1` ✅
   - `go vet ./internal/tui/...` ✅

3. Unsafe-assertion grep check returned zero lines with fixed-string filtering:
   - `grep -nF '.(' internal/tui/update.go | grep -vF ', ok' | grep -vF 'switch' | grep -vF '//' | grep -vF 'msg.(type)' | grep -vF 'msg.(tea.'`

4. Safety pattern preserved behavior (no routing/order/command collection changes), while removing panic risk from unchecked assertions.

## Task 17: Safe Type Assertions in TUI Update Routing

**Date**: 2026-02-06
**Task**: Replace unchecked type assertions in `internal/tui/update.go` with safe `value, ok` pattern

### Changes Made:

1. Replaced all 11 unchecked assertions (`model.(views.XxxModel)` and `tm.(views.XxxModel)`) with guarded assertions:
   - `if d, ok := model.(views.DashboardModel); ok { m.dashboard = d }`
   - `if b, ok := model.(views.BackupListModel); ok { m.backupList = b }`
   - `if r, ok := model.(views.RestoreModel); ok { m.restore = r }`
   - `if s, ok := model.(views.SettingsModel); ok { m.settings = s }`
   - `if l, ok := model.(views.LogsModel); ok { m.logs = l }`
   - `if su, ok := model.(views.SetupModel); ok { m.setup = su }`

2. Applied in all required routing points:
   - `propagateWindowSize()` (5 assertions)
   - setup mode default branch (1 assertion)
   - state-specific routing switch (5 assertions)

3. Preserved behavior by silently retaining prior model values when assertion fails (no logging/panic/order changes).

### Verification Passed:

- `go build ./cmd/dotkeeper/` ✅
- `go vet ./internal/tui/...` ✅
- `go test ./internal/tui/... -race -count=1` ✅
- assertion grep check returns zero matches (using extended-regex equivalent) ✅
- LSP diagnostics for `internal/tui/update.go`: no errors ✅

### Pattern:

- BubbleTea sub-view updates should always use safe type assertions after `Update()` returns `tea.Model`.
- Guarded assertions prevent rare panic paths while keeping normal behavior unchanged.
