# Decisions — tui-tabs

## 2026-02-05 Planning
- 5 tabs only: Dashboard, Backups, Restore, Settings, Logs
- FileBrowser excluded from cycling (kept in code)
- Setup wizard: no tab bar shown
- Underline + bold + purple for active tab
- Gray (#666666) for inactive tabs
- Separator: │ with color #444444
- Number keys 1-5 for direct navigation
- tabOrder slice decouples from enum values
- isInputActive() guards number keys and Tab
- tabBarHeight = 2 (1 line tabs + 1 line spacing)
- Dashboard shortcuts b/r/s kept for backward compat
- View titles kept (not removed)
