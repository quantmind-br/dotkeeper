# Learnings

## 2026-02-06 Session Start
- Tests pass, lint has pre-existing issues (unused field, deprecated methods, gosimple)
- These lint issues are NOT from our work — they existed before
- Project uses `make test`, `make lint`, `make build`
- BubbleTea TUI with eager-initialized sub-models
- Go standard `testing` package + BubbleTea test patterns (sendKey, stripANSI)

## Task 3: Remove stateReadOnly from Settings
- Removed `stateReadOnly` from iota enum entirely (made `stateListNavigating` the new `0` value)
- No other code referenced `stateReadOnly` outside settings.go/settings_test.go — safe removal
- `handleReadOnlyInput` was the only method that handled the `e` key to enter edit mode — removed entirely
- `IsEditing()` now always returns `true` — settings is always interactive
- Esc in `stateListNavigating` now does nothing (parent tab system handles view exit)
- `TestSettingsSaveWithS` previously needed `e` key press first to enter edit mode — no longer needed since model starts in `stateListNavigating`

## Task 4: Dotfile Preset Detection
- Implemented async preset detection using `tea.Cmd` in BubbleTea setup wizard.
- Created `internal/pathutil/presets.go` to scan for common dotfiles (bashrc, zshrc, vimrc, etc.) and folders.
- Logic detects existing files/folders relative to user home.
- Added two new Setup Wizard steps: `StepPresetFiles` and `StepPresetFolders`.
- Integrated preset selection into `m.addedFiles` and `m.addedFolders` before the manual add steps.
- Updated setup tests to handle the new step progression and verify preset toggling.
- `SetupModel` now manages a `presetCursor` state for checklist navigation.
- Encountered and fixed a flaky test in `internal/backup` (unrelated cleanup issue).

## Task: Embed filepicker in Settings
- Embedded `filepicker.Model` directly into `SettingsModel` (eager initialization).
- Routed messages to filepicker *before* the main `tea.KeyMsg` switch because filepicker relies on its own internal message types (like `filepicker.FileSelectedMsg`) which need to be intercepted.
- Handled `filepicker.DidSelectFile(msg)` in the general message handler (top of Update) to capture file selection events.
- Added `stateFilePickerActive` to manage the modal state, allowing return to previous browsing state (`filePickerParent`).
- Preserved existing `a` key functionality (text input) while adding `b` key for browsing.
