# Learnings — TUI BubbleTea Improvements

- Added framework-level tests in internal/tui/tui_test.go focused on model/update/view orchestration, not view internals.
- Introduced NewModelForTest(cfg, store) to inject config/history and bypass config.Load/history.NewStore side effects in deterministic tests.
- Setup transition test is most stable when starting from NewModelForTest(nil, nil) and pre-saving config so Update(SetupCompleteMsg) follows real load path.
- For input-lock behavior, entering SettingsView edit mode via Enter makes IsEditing() true and correctly blocks Tab navigation.
- Existing internal/tui/views help/input-active tests are currently failing in baseline; race run failure is pre-existing and unrelated to this task scope.
