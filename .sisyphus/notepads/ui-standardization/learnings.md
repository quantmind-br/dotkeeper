# Learnings — UI Standardization

## Session: ses_3cf2a0e20ffe7uXFAoujkRJ0oE

## Unified Status Bar Rendering
- Standardizing the status bar rendering into a single helper `RenderStatusBar` ensures consistent behavior (truncation, styling) across different views.
- Truncating long status/error messages prevents terminal wrapping issues that can break TUI layouts.
- Preferring separate parameters for `status` and `error` avoids the "code smell" of checking string content (e.g., `strings.Contains(m.err, "success")`) to decide styling.
