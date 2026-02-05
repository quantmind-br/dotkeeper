# Learnings

## Help Overlay Implementation (2026-02-05)

- **Import cycle prevention**: HelpEntry and HelpProvider MUST live in `views` package since `tui` imports `views` (one-way). Putting them in `tui/help.go` would create a cycle when views need to implement HelpProvider.
- **view.go needed views import**: The original view.go did not import the `views` package directly (it accessed views via model fields). Adding the HelpProvider type check required adding the import.
- **Help is overlay, not ViewState**: showingHelp is a bool on Model, not a ViewState entry. viewCount stays at 6. This keeps the tab cycling logic clean.
- **Key handling order matters**: showingHelp dismiss → Help toggle → quit → tab → dashboard shortcuts. Dismissal must come first so any key (including q, tab) dismisses the overlay instead of performing its normal action.
- **WindowSizeMsg unaffected**: It's in a separate case branch from KeyMsg, so it propagates normally regardless of showingHelp state.
- **interface{} type assertion for HelpProvider**: Used `interface{}(m.dashboard).(views.HelpProvider)` pattern since Go requires the value to be interface{} before asserting a different interface. Views don't implement HelpProvider yet (separate task), so the ok check gracefully handles this.

## History Wiring to TUI (2026-02-05)

- **Signature changes cascade to tests**: When changing `NewBackupList` and `NewRestore` signatures to accept `*history.Store`, all test calls must also be updated (pass `nil` for tests that don't need history).
- **Store error handling in NewModel()**: `history.NewStore()` can fail (e.g., can't create state dir). Use `fmt.Fprintf(os.Stderr, ...)` to warn and set store to nil — never crash the TUI.
- **SetupCompleteMsg handler**: Mirror the same store creation pattern as `NewModel()` — create store, pass to all views that need it, store on model.
- **Tab refresh pattern**: Exact same pattern as BackupListView/RestoreView: `if m.state == LogsView && prevState != LogsView { cmds = append(cmds, m.logs.LoadHistory()) }`. Goes after existing refresh checks, before `return m, tea.Batch(cmds...)`.
- **Best-effort history logging**: Always `if m.store != nil { _ = m.store.Append(...) }` — nil check + ignore error. TUI operations must never fail due to history logging.
- **Go `:=` reuse**: In `NewModel()`, `err` can be reused with `:=` since `store` is a new variable in the same assignment, which is valid Go.
