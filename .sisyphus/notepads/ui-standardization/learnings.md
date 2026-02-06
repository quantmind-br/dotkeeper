# Learnings — UI Standardization

## Session: ses_3cf2a0e20ffe7uXFAoujkRJ0oE

## Unified Status Bar Rendering
- Standardizing the status bar rendering into a single helper `RenderStatusBar` ensures consistent behavior (truncation, styling) across different views.
- Truncating long status/error messages prevents terminal wrapping issues that can break TUI layouts.
- Preferring separate parameters for `status` and `error` avoids the "code smell" of checking string content (e.g., `strings.Contains(m.err, "success")`) to decide styling.

## 2026-02-06 Execution Complete

### Architecture Patterns
- `NewListDelegate()` in `views/styles.go` avoids import cycle with `components/tabbar.go`
- `ViewChromeHeight = 6` constant replaces magic `-6` across all views
- `RenderStatusBar(width, status, errMsg, helpText)` unifies all footer rendering
- `PlaceOverlay()` helper wraps `lipgloss.Place` to avoid lipgloss imports outside styles.go
- Settings state machine: typed enum `settingsState` replaces 5 interacting booleans

### Testing Patterns
- `stripANSI()` test helper must be created BEFORE any view styling changes
- All View() output assertions should use `stripANSI()` wrapper for resilience
- Tests in same package as code — `package views` — can access private fields

### Style Migration
- Component configuration (textinput cursor/prompt) stays inline — not view styling
- Dynamic styles (viewport with computed width/height) stay inline — static base goes to styles.go
- `lipgloss` import can be removed from files that only use styles via `views.DefaultStyles()`

### Execution Insights
- 9 tasks completed in 6 waves with ~40% speedup from parallelization
- Wave 3 (3 parallel list standardizations) was most efficient
- Settings refactor (Task 8) was highest risk but completed successfully with typed state machine
- All 44 TUI tests pass with race detection
