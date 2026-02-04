# Interactive TUI Configuration

## TL;DR

> **Quick Summary**: Implement interactive configuration wizard and enhanced settings view in the dotkeeper TUI application
> 
> **Deliverables**:
> - Setup wizard view for first-time configuration
> - Enhanced Settings view with editable fields
> - File/folder picker integration
> - Config save functionality
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: NO - sequential (views depend on each other)
> **Critical Path**: Setup View → Settings Enhancement → Model Integration → Testing

---

## Context

### Original Request
User requested interactive configuration through the TUI - when no config exists, show a setup wizard; allow editing settings through the TUI.

### Current State
- Settings view exists but is read-only (displays config values)
- No setup wizard when config is missing
- Uses fallback default config when config file doesn't exist
- TUI shows "No items" in Backups because no backups exist yet

---

## Work Objectives

### Core Objective
Enable users to configure dotkeeper entirely through the TUI interface without needing to manually edit YAML files.

### Concrete Deliverables
- `internal/tui/views/setup.go` - Setup wizard component
- Enhanced `internal/tui/views/settings.go` - Editable settings
- Updated `internal/tui/model.go` - Setup mode integration
- Updated `internal/tui/update.go` - Setup message handling

### Definition of Done
- [x] `dotkeeper` without config shows setup wizard
- [x] Setup wizard allows configuring: backup_dir, git_remote, files, folders
- [x] Settings view allows editing existing config
- [x] Config is saved to `~/.config/dotkeeper/config.yaml`
- [x] All tests pass

### Must Have
- Text input for backup directory path
- Text input for git remote URL (optional)
- Ability to add multiple files to track
- Ability to add multiple folders to track
- Save configuration to disk
- Return to normal TUI after setup complete

### Must NOT Have (Guardrails)
- No file browser for selecting files (use text input for paths)
- No schedule configuration in initial wizard (keep simple)
- No git integration in wizard (just save the URL)
- No encryption password setup in wizard

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: Tests-after
- **Framework**: go test

### Agent-Executed QA Scenarios (MANDATORY)

```
Scenario: Setup wizard appears on first run
  Tool: interactive_bash (tmux)
  Preconditions: No config file at ~/.config/dotkeeper/config.yaml
  Steps:
    1. rm -f ~/.config/dotkeeper/config.yaml
    2. tmux new-session: dotkeeper
    3. Wait for: "Welcome to dotkeeper" in output (timeout: 5s)
    4. Assert: Screen shows setup wizard, not dashboard
  Expected Result: Setup wizard is displayed
  Evidence: Terminal screenshot

Scenario: Complete setup wizard creates config
  Tool: interactive_bash (tmux)  
  Preconditions: No config file exists
  Steps:
    1. Start dotkeeper
    2. Press Enter (welcome screen)
    3. Type "~/.my-dotfiles" and Enter (backup dir)
    4. Press Enter (skip git remote)
    5. Type "~/.zshrc" and Enter (add file)
    6. Press Enter (done with files)
    7. Type "~/.config/hypr" and Enter (add folder)
    8. Press Enter (done with folders)
    9. Press Enter (confirm)
    10. Assert: Config file exists at ~/.config/dotkeeper/config.yaml
    11. Assert: Config contains backup_dir: ~/.my-dotfiles
  Expected Result: Config saved correctly
  Evidence: cat ~/.config/dotkeeper/config.yaml output

Scenario: Settings view shows edit options
  Tool: interactive_bash (tmux)
  Preconditions: Config file exists
  Steps:
    1. Start dotkeeper
    2. Press Tab until Settings view
    3. Assert: Edit options visible (e.g., "Press e to edit")
  Expected Result: Settings has edit capability
  Evidence: Terminal screenshot
```

---

## TODOs

- [x] 1. Create Setup Wizard View

  **What to do**:
  - Create `internal/tui/views/setup.go`
  - Implement multi-step wizard with states: Welcome, BackupDir, GitRemote, AddFiles, AddFolders, Confirm, Complete
  - Use `bubbles/textinput` for text entry
  - Track added files/folders in slice
  - Emit `SetupCompleteMsg` when done with config pointer

  **Must NOT do**:
  - Don't implement file picker (use text input)
  - Don't add schedule configuration

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Blocks**: Task 3, 4

  **References**:
  - `internal/tui/views/dashboard.go` - View pattern to follow
  - `github.com/charmbracelet/bubbles/textinput` - Text input component
  - `internal/config/config.go` - Config struct and Save method

  **Acceptance Criteria**:
  - [ ] SetupModel struct with step tracking
  - [ ] NewSetup() constructor
  - [ ] Init(), Update(), View() methods
  - [ ] SetupCompleteMsg message type
  - [ ] Handles Enter to advance steps
  - [ ] Handles Esc to go back
  - [ ] Displays current config values
  - [ ] Saves config on completion

  **Commit**: YES
  - Message: `feat(tui): add setup wizard for first-time configuration`
  - Files: `internal/tui/views/setup.go`

---

- [x] 2. Enhance Settings View with Edit Mode

  **What to do**:
  - Modify `internal/tui/views/settings.go`
  - Add edit mode toggle (press 'e' to enter edit mode)
  - Add cursor to navigate between fields
  - Use textinput for editing values
  - Add ability to add/remove files and folders
  - Save config on exit from edit mode

  **Must NOT do**:
  - Don't remove existing read-only display
  - Don't auto-save on every keystroke

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Blocks**: Task 4

  **References**:
  - `internal/tui/views/settings.go` - Current implementation
  - `internal/tui/views/setup.go` - Text input pattern (after Task 1)
  - `internal/config/config.go:Save()` - Save method

  **Acceptance Criteria**:
  - [ ] Press 'e' enters edit mode
  - [ ] Arrow keys navigate fields
  - [ ] Enter edits selected field
  - [ ] 'a' adds new file/folder
  - [ ] 'd' deletes selected file/folder
  - [ ] Esc exits edit mode
  - [ ] 's' saves config
  - [ ] Visual indication of edit mode

  **Commit**: YES
  - Message: `feat(tui): add edit mode to settings view`
  - Files: `internal/tui/views/settings.go`

---

- [x] 3. Update Main Model for Setup Mode

  **What to do**:
  - Modify `internal/tui/model.go`
  - Add `setupMode bool` and `setup SetupModel` fields
  - In NewModel(), check if config exists; if not, set setupMode=true
  - Add SetupView to ViewState enum

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Depends On**: Task 1
  - **Blocks**: Task 4

  **References**:
  - `internal/tui/model.go` - Current model
  - `internal/config/config.go:Load()` - Check if config exists

  **Acceptance Criteria**:
  - [ ] Model has setupMode field
  - [ ] Model has setup SetupModel field
  - [ ] NewModel checks for config existence
  - [ ] If no config, setupMode=true and setup initialized

  **Commit**: NO (group with Task 4)

---

- [x] 4. Update Update/View for Setup Mode

  **What to do**:
  - Modify `internal/tui/update.go`
  - If setupMode, forward messages to setup model
  - Handle SetupCompleteMsg to exit setup mode and reload config
  - Modify `internal/tui/view.go`
  - If setupMode, render setup view instead of normal views

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Depends On**: Task 1, 3

  **References**:
  - `internal/tui/update.go` - Current update logic
  - `internal/tui/view.go` - Current view logic
  - `internal/tui/views/setup.go:SetupCompleteMsg` - Message type

  **Acceptance Criteria**:
  - [ ] Update forwards to setup when setupMode
  - [ ] SetupCompleteMsg transitions to normal mode
  - [ ] Config reloaded after setup complete
  - [ ] All views reinitialized with new config
  - [ ] View renders setup when setupMode

  **Commit**: YES
  - Message: `feat(tui): integrate setup wizard into main TUI flow`
  - Files: `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/view.go`

---

- [x] 5. Add Tests for Setup Wizard

  **What to do**:
  - Create `internal/tui/views/setup_test.go`
  - Test step transitions
  - Test config population
  - Test save functionality

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Depends On**: Task 1

  **References**:
  - `internal/tui/views/dashboard_test.go` - Test pattern
  - `internal/tui/views/setup.go` - Implementation to test

  **Acceptance Criteria**:
  - [ ] Test NewSetup creates model
  - [ ] Test step progression with Enter
  - [ ] Test step regression with Esc
  - [ ] Test config values are set
  - [ ] All tests pass

  **Commit**: YES
  - Message: `test(tui): add tests for setup wizard`
  - Files: `internal/tui/views/setup_test.go`

---

- [x] 6. Build and Manual Verification

  **What to do**:
  - Run `go build ./...`
  - Run `go test ./...`
  - Delete config and run dotkeeper manually
  - Complete setup wizard
  - Verify config file created
  - Verify TUI works after setup

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`

  **Acceptance Criteria**:
  - [ ] Build succeeds
  - [ ] All tests pass
  - [ ] Manual verification complete

  **Commit**: NO

---

## Commit Strategy

| After Task | Message | Files |
|------------|---------|-------|
| 1 | `feat(tui): add setup wizard for first-time configuration` | setup.go |
| 2 | `feat(tui): add edit mode to settings view` | settings.go |
| 4 | `feat(tui): integrate setup wizard into main TUI flow` | model.go, update.go, view.go |
| 5 | `test(tui): add tests for setup wizard` | setup_test.go |

---

## Success Criteria

### Verification Commands
```bash
# Remove config and test
rm -f ~/.config/dotkeeper/config.yaml
go build -o dotkeeper ./cmd/dotkeeper
./dotkeeper
# Should show setup wizard

# After completing wizard
cat ~/.config/dotkeeper/config.yaml
# Should show saved config
```

### Final Checklist
- [x] Setup wizard appears when no config
- [x] All setup steps work correctly
- [x] Config saved after wizard completion
- [x] Normal TUI loads after setup
- [x] Settings can be edited
- [x] All tests pass
