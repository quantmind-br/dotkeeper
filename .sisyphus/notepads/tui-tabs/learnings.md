# Learnings — tui-tabs

## 2026-02-05 Session Start
- ViewState enum: DashboardView=0, FileBrowserView=1, BackupListView=2, RestoreView=3, SettingsView=4, LogsView=5, SetupView=6
- viewCount = 6 in update.go:11
- Tab cycling: (m.state + 1) % viewCount
- Views use magic numbers for height (msg.Height - 6, m.height - 10)
- propagateWindowSize passes raw WindowSizeMsg
- Lipgloss v1.1.0, bubbletea v1.3.10, bubbles v0.21.1
- Module path: github.com/diogo/dotkeeper
- Components dir exists but is empty (has .gitkeep)
## Lipgloss Test Patterns
When testing TUI components styled with lipgloss, use a dual-layer verification:
1. Strip ANSI codes to verify structural correctness (content consistency).
2. Compare RAW outputs to verify styling changes between different states, but only if ANSI codes are detected in the output.
