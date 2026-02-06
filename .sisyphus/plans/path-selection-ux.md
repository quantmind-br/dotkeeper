# Path Selection UX Improvements

## TL;DR

> **Quick Summary**: Implement 8 UI/UX improvements for dotkeeper's file/folder selection system, spanning config model, backup collector, setup wizard, settings view, dashboard, and file browser integration. The work modernizes the path selection experience from blind text input to interactive browsing, presets, previews, and exclusion patterns.
>
> **Deliverables**:
> - Shared `pathutil` package extracting duplicated `expandHome()`
> - Dotfile preset detection and checklist in Setup Wizard
> - Embedded filepicker in Settings and Setup for visual path browsing
> - Settings view without unnecessary read-only mode
> - Backup content preview with real file counts, sizes, and broken path detection
> - `Exclude`/`DisabledFiles`/`DisabledFolders` fields in config with collector support
> - Tab-completion style autocomplete during path input
> - Inline actions per path (inspect, delete, toggle enable/disable)
> - Bulk glob resolution for adding multiple paths at once
> - Tests for all new functionality following existing patterns
>
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 4 waves
> **Critical Path**: Task 0 → Task 1 → Task 3 → Task 6 → Task 9

---

## Context

### Original Request
Analyze the configuration system for selecting backup files/folders and provide UI/UX improvement suggestions, then plan their implementation.

### Interview Summary
**Key Discussions**:
- All 8 suggestions (A-H) included in scope
- Tests-after strategy following existing BubbleTea patterns (sendKey, stripANSI, type assertions)
- No backward compatibility concerns for adding new config fields
- File Browser architecture: embed `filepicker.Model` directly into Settings/Setup
- Enable/disable model: separate `DisabledFiles`/`DisabledFolders` lists in config
- Standalone FileBrowserView tab can be removed after integration

**Research Findings**:
- `filebrowser.go` exists with working `filepicker.Model` but disconnected from flows
- Settings has 6 states — stateReadOnly serves no purpose in TUI context
- Dashboard `fileCount` shows config entries not actual files
- `expandHome()` is duplicated in `views/helpers.go` and `backup/collector.go`
- Collector has no exclusion/filtering capability
- Go's `filepath.Glob` doesn't support `**` — need manual recursive glob or `doublestar` lib

### Metis Review
**Identified Gaps** (addressed):
- File Browser architecture decision → resolved: embed filepicker directly
- Enable/disable persistence model → resolved: separate DisabledPaths fields
- Duplicate `expandHome()` → resolved: extract to shared `internal/pathutil` package
- Exclusion pattern syntax → resolved: simple glob via `filepath.Match()` only
- Preset detection scope → resolved: curated hardcoded list, max 20, checked for existence
- Autocomplete scope → resolved: filesystem tab-completion only (next matching path)
- Settings state explosion risk → resolved: eliminate stateReadOnly first, cap at 8 total states
- Standalone File Browser tab → resolved: remove from tab bar after integration

---

## Work Objectives

### Core Objective
Transform dotkeeper's path selection from blind text input into an interactive, discoverable, and informative experience across all touchpoints (setup, settings, dashboard).

### Concrete Deliverables
- `internal/pathutil/pathutil.go` — shared path utilities
- Modified `internal/config/config.go` — new fields: `Exclude`, `DisabledFiles`, `DisabledFolders`
- Modified `internal/backup/collector.go` — exclusion pattern support
- Modified `internal/tui/views/setup.go` — dotfile presets checklist step
- Modified `internal/tui/views/settings.go` — embedded filepicker, eliminated read-only, inline actions, exclusion list, preview panel
- Modified `internal/tui/views/dashboard.go` — real file count via async scan
- Modified `internal/tui/views/helpers.go` — use shared pathutil, add autocomplete helper
- Removed standalone `FileBrowserView` from tab navigation
- Tests for all new features

### Definition of Done
- [x] `make test` passes with zero failures
- [x] `make lint` passes
- [x] All 8 features (A-H) functional in TUI
- [x] No blocking calls in any `Update()` method
- [x] Config backward-compatible (existing YAML loads without errors)

### Must Have
- Dotfile presets in setup wizard with toggle checklist
- Filepicker integration for browsing when adding paths
- Real file count and size preview
- Exclusion pattern support in config and collector
- All filesystem operations async via `tea.Cmd`
- Tests following existing patterns

### Must NOT Have (Guardrails)
- Must NOT add more than 2 new states to Settings — extract sub-components instead
- Must NOT use regex for exclusion patterns — glob only via `filepath.Match()`
- Must NOT scan `$HOME` recursively for preset detection — use hardcoded candidate list
- Must NOT add fuzzy search for autocomplete
- Must NOT add file content preview (only metadata: size, date, type)
- Must NOT change Files/Folders from `[]string` to `[]object`
- Must NOT add recursive glob `**` without a results cap (max 500 paths)
- Must NOT block in `Update()` — every filesystem operation uses `tea.Cmd`
- Must NOT add mouse interaction or drag-and-drop support
- Must NOT modify backup/restore flow logic (only collector input filtering)
- Must NOT add "select all" for presets — each must be individually toggled
- Must NOT persist undo history for settings changes

---

## Verification Strategy

> **UNIVERSAL RULE: ZERO HUMAN INTERVENTION**
>
> ALL tasks are verified by running `make test`, `make lint`, and specific `go test` commands.

### Test Decision
- **Infrastructure exists**: YES (30 test files, `go test` with `-race`)
- **Automated tests**: Tests-after (implement feature, then add tests)
- **Framework**: Go standard `testing` package + BubbleTea test patterns

### Agent-Executed QA Scenarios (MANDATORY — ALL tasks)

**Verification Tool by Deliverable Type:**

| Type | Tool | How Agent Verifies |
|------|------|-------------------|
| **TUI views** | interactive_bash (tmux) | Build binary, launch TUI, send keystrokes, validate output |
| **Config changes** | Bash (go test) | Unit tests with temp dirs and XDG override |
| **Collector logic** | Bash (go test) | Create temp files, run collector, assert results |
| **Build/Lint** | Bash | `make test && make lint` — exit code 0 |

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 0 (Foundation — Start Immediately):
└── Task 0: Extract shared pathutil package

Wave 1 (P0+P1 Core — After Wave 0):
├── Task 1: [F] Eliminate stateReadOnly in Settings
├── Task 2: [B] Dotfile preset detection + Setup Wizard integration
└── Task 3: Config struct: add Exclude, DisabledFiles, DisabledFolders

Wave 2 (P0+P1 Features — After Wave 1):
├── Task 4: [A] Embed filepicker in Settings path addition
├── Task 5: [A] Embed filepicker in Setup Wizard path addition
├── Task 6: [E] Collector exclusion pattern support
└── Task 7: [D] Backup content preview in Settings + Dashboard

Wave 3 (P2+P3 Enhancements — After Wave 2):
├── Task 8: [C] Tab-completion autocomplete for path input
├── Task 9: [H] Inline actions per path item
├── Task 10: [G] Bulk glob resolution
└── Task 11: Remove standalone FileBrowserView tab

Wave 4 (Finalization — After Wave 3):
└── Task 12: Full test suite + lint pass + cleanup

Critical Path: Task 0 → Task 1 → Task 4 → Task 9 → Task 12
Parallel Speedup: ~45% faster than sequential
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 0 | None | 1,2,3,4,5,6,7,8,9,10 | None (foundation) |
| 1 | 0 | 4, 9 | 2, 3 |
| 2 | 0 | 5 | 1, 3 |
| 3 | 0 | 6, 7, 9 | 1, 2 |
| 4 | 1 | 11 | 5, 6, 7 |
| 5 | 2 | 11 | 4, 6, 7 |
| 6 | 3 | 10 | 4, 5, 7 |
| 7 | 3 | None | 4, 5, 6 |
| 8 | 0 | 12 | 9, 10, 11 |
| 9 | 1, 3 | 12 | 8, 10, 11 |
| 10 | 6 | 12 | 8, 9, 11 |
| 11 | 4, 5 | 12 | 8, 9, 10 |
| 12 | 8, 9, 10, 11 | None | None (final) |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Agents |
|------|-------|-------------------|
| 0 | 0 | `category="quick", load_skills=["git-master"]` |
| 1 | 1, 2, 3 | 3 parallel agents: `category="unspecified-low"` |
| 2 | 4, 5, 6, 7 | 4 parallel agents: `category="visual-engineering", load_skills=["frontend-ui-ux"]` for 4,5,7; `category="unspecified-low"` for 6 |
| 3 | 8, 9, 10, 11 | 4 parallel agents |
| 4 | 12 | `category="quick"` |

---

## TODOs

- [x] 0. Extract shared `pathutil` package

  **What to do**:
  - Create `internal/pathutil/pathutil.go` with `ExpandHome(path string) string` function
  - Move the `expandHome()` logic from `views/helpers.go` (lines 13-24) to the new package
  - Update `views/helpers.go` to call `pathutil.ExpandHome()` instead of local `expandHome()`
  - Update `backup/collector.go` to call `pathutil.ExpandHome()` instead of local `expandHome()`
  - Remove the duplicate `expandHome()` from `backup/collector.go` (lines 109-120)
  - Make the function exported (`ExpandHome` with capital E)
  - Add basic unit test `internal/pathutil/pathutil_test.go`

  **Must NOT do**:
  - Do NOT change any behavior — pure extraction refactor
  - Do NOT rename the function signature pattern
  - Do NOT add extra functionality beyond `ExpandHome`

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`git-master`]
    - `git-master`: Atomic commit after extraction
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Not a UI task

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 0 (solo)
  - **Blocks**: All other tasks
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/helpers.go:13-24` — Current `expandHome()` implementation (source of truth)
  - `internal/backup/collector.go:109-120` — Duplicate `expandHome()` to remove

  **API/Type References**:
  - Both implementations are identical: handle `~/` prefix and `~` alone

  **Test References**:
  - `internal/tui/views/helpers_test.go` — May have tests for expandHome behavior
  - `internal/backup/collector_test.go` — May reference expandHome indirectly

  **Tool Recommendations**:
  - Use `lsp_find_references` on both `expandHome()` functions to find all callers
  - Use `ast_grep_search` pattern `expandHome($PATH)` in Go files to find all usages

  **Acceptance Criteria**:
  - [x] File exists: `internal/pathutil/pathutil.go` with exported `ExpandHome()`
  - [x] File exists: `internal/pathutil/pathutil_test.go` with basic tests
  - [x] `views/helpers.go` calls `pathutil.ExpandHome()` — no local `expandHome()`
  - [x] `backup/collector.go` calls `pathutil.ExpandHome()` — no local `expandHome()`
  - [x] `go test ./internal/pathutil/ -v` → PASS
  - [x] `go test ./internal/tui/views/ -v` → PASS (no regressions)
  - [x] `go test ./internal/backup/ -v` → PASS (no regressions)
  - [x] `make lint` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Shared pathutil works identically to old implementations
    Tool: Bash (go test)
    Preconditions: Code changes applied
    Steps:
      1. Run: go test ./internal/pathutil/ -v -run TestExpandHome
      2. Assert: Tests cover "~/path" → "/home/<user>/path" expansion
      3. Assert: Tests cover "~" → "/home/<user>" expansion
      4. Assert: Tests cover "/absolute/path" → unchanged
      5. Assert: Tests cover "relative/path" → unchanged
      6. Run: go test ./internal/tui/views/ -v
      7. Assert: All existing tests pass (zero regressions)
      8. Run: go test ./internal/backup/ -v
      9. Assert: All existing tests pass (zero regressions)
    Expected Result: All tests pass, behavior preserved
    Evidence: Terminal output captured

  Scenario: No duplicate expandHome functions remain
    Tool: Bash (grep)
    Preconditions: Extraction complete
    Steps:
      1. Run: grep -rn "func expandHome" internal/ --include="*.go"
      2. Assert: Only ONE result in `internal/pathutil/pathutil.go` (exported as ExpandHome)
      3. Assert: ZERO results in `views/helpers.go` or `backup/collector.go`
    Expected Result: Single source of truth for path expansion
    Evidence: grep output captured
  ```

  **Commit**: YES
  - Message: `refactor(pathutil): extract shared ExpandHome to internal/pathutil`
  - Files: `internal/pathutil/pathutil.go`, `internal/pathutil/pathutil_test.go`, `internal/tui/views/helpers.go`, `internal/backup/collector.go`
  - Pre-commit: `make test`

---

- [x] 1. [F] Eliminate `stateReadOnly` in Settings

  **What to do**:
  - Change `NewSettings()` to initialize with `state: stateListNavigating` instead of `stateReadOnly`
  - Remove the `stateReadOnly` constant from the `settingsState` enum (or keep it but never use it)
  - Remove `handleReadOnlyInput()` method entirely
  - Remove the `case stateReadOnly` in `View()` that renders "Press 'e' to edit"
  - Update `View()` to always show the navigable list with help text
  - Update `IsEditing()` to always return `true` (or refactor: settings is always "editing")
  - Update all existing tests that assert `stateReadOnly` initial state
  - The main list should be immediately navigable with arrow keys on entry

  **Must NOT do**:
  - Do NOT change the other 5 states' behavior
  - Do NOT change key bindings for save/escape/navigation
  - Do NOT change any other view's behavior

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Simple state removal, not design work

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3)
  - **Blocks**: Task 4 (filepicker needs simplified state machine)
  - **Blocked By**: Task 0 (pathutil extraction)

  **References**:

  **Pattern References**:
  - `internal/tui/views/settings.go:18-25` — Current state enum, `stateReadOnly` is first value
  - `internal/tui/views/settings.go:93` — `NewSettings()` sets `state: stateReadOnly`
  - `internal/tui/views/settings.go:143-152` — `handleReadOnlyInput()` to remove
  - `internal/tui/views/settings.go:517-522` — `View()` read-only rendering to remove
  - `internal/tui/views/settings.go:586-588` — `IsEditing()` checks `stateReadOnly`

  **Test References**:
  - `internal/tui/views/settings_test.go:34` — `TestSettingsNewSettings` asserts `stateReadOnly`
  - `internal/tui/views/settings_test.go:42-63` — `TestSettingsViewReadOnlyShowsValues` checks "Press 'e' to edit"
  - `internal/tui/views/settings_test.go:65-87` — `TestSettingsIsEditingByState` tests read-only returns false
  - `internal/tui/views/settings_test.go:89-111` — `TestSettingsTransitions` tests `e` key → stateListNavigating

  **Acceptance Criteria**:
  - [x] `NewSettings()` creates model with `state == stateListNavigating`
  - [x] View output does NOT contain "Press 'e' to edit"
  - [x] Arrow keys work immediately on settings entry (no `e` key required)
  - [x] `handleReadOnlyInput()` method removed from codebase
  - [x] All existing tests updated and passing
  - [x] `go test ./internal/tui/views/ -v -run TestSettings` → PASS
  - [x] `make test` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Settings opens in navigable mode immediately
    Tool: interactive_bash (tmux)
    Preconditions: Binary built with `make build`
    Steps:
      1. Create temp config: echo valid YAML to temp file
      2. Launch TUI: ./bin/dotkeeper
      3. Press 's' to navigate to Settings
      4. Assert: No "Press 'e' to edit" text visible
      5. Press down arrow 3 times
      6. Assert: Cursor moves through list items (no key rejected)
      7. Press 'q' to quit
    Expected Result: Settings is immediately interactive
    Evidence: Terminal output captured

  Scenario: Updated tests pass
    Tool: Bash (go test)
    Preconditions: Code changes applied
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestSettings -count=1
      2. Assert: TestSettingsNewSettings passes with new initial state
      3. Assert: TestSettingsTransitions passes without `e` key step
      4. Assert: TestSettingsIsEditingByState updated
      5. Run: make test
      6. Assert: Full suite passes
    Expected Result: All tests pass
    Evidence: Terminal output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): remove read-only mode from settings, enable direct navigation`
  - Files: `internal/tui/views/settings.go`, `internal/tui/views/settings_test.go`
  - Pre-commit: `make test`

---

- [x] 2. [B] Dotfile preset detection + Setup Wizard integration

  **What to do**:
  - Create `internal/pathutil/presets.go` with a curated list of common dotfiles and dotfolders (max 20 total):
    - Files (10): `~/.bashrc`, `~/.zshrc`, `~/.bash_profile`, `~/.profile`, `~/.gitconfig`, `~/.gitignore_global`, `~/.vimrc`, `~/.tmux.conf`, `~/.ssh/config`, `~/.gnupg/gpg.conf`
    - Folders (10): `~/.config/nvim`, `~/.config/kitty`, `~/.config/alacritty`, `~/.config/hypr`, `~/.config/fish`, `~/.config/starship`, `~/.config/waybar`, `~/.config/rofi`, `~/.config/wezterm`, `~/.config/zsh`
  - Create `DetectDotfiles(homeDir string) []DotfilePreset` that checks which presets exist on the filesystem
  - Each `DotfilePreset` has: `Path string`, `Exists bool`, `IsDir bool`, `Size int64`, `FileCount int` (for dirs)
  - Add a new setup step `StepPresetFiles` between `StepGitRemote` and `StepAddFiles`
  - `StepPresetFiles` shows a checklist of detected dotfiles with toggle (space to toggle, enter to continue)
  - Show file size or folder file count next to each entry: `[x] ~/.zshrc (2.1KB)` or `[x] ~/.config/nvim/ (45 files, 890KB)`
  - Pre-select shell config based on current `$SHELL` (if zsh → `.zshrc` pre-checked)
  - After presets, existing `StepAddFiles` and `StepAddFolders` still work for custom paths
  - Preset detection MUST run async via `tea.Cmd` (filesystem I/O)
  - Add a new step `StepPresetFolders` for folder presets, similarly between GitRemote and AddFolders

  **Must NOT do**:
  - Do NOT scan `$HOME` recursively — only check the curated list
  - Do NOT add more than 20 preset entries total
  - Do NOT add "select all" — each must be individually toggled
  - Do NOT block in `Update()` — preset detection uses `tea.Cmd`

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: Checklist UI design and interaction pattern
  - **Skills Evaluated but Omitted**:
    - `git-master`: Not a git task

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3)
  - **Blocks**: Task 5 (setup filepicker integration)
  - **Blocked By**: Task 0 (uses pathutil.ExpandHome)

  **References**:

  **Pattern References**:
  - `internal/tui/views/setup.go:14-24` — Current `SetupStep` enum — add `StepPresetFiles` and `StepPresetFolders`
  - `internal/tui/views/setup.go:97-173` — `handleEnter()` step progression — wire new steps
  - `internal/tui/views/setup.go:191-295` — `View()` rendering — add preset checklist views
  - `internal/tui/views/dashboard.go:149-168` — Async `tea.Cmd` pattern for filesystem I/O

  **Test References**:
  - `internal/tui/views/setup_test.go:41-90` — Step progression tests — update for new steps
  - `internal/tui/views/setup_test.go:129-186` — File adding tests — verify presets + custom still works

  **Acceptance Criteria**:
  - [x] `internal/pathutil/presets.go` exists with `DetectDotfiles()` function
  - [x] `internal/pathutil/presets_test.go` exists with detection tests using `t.TempDir()`
  - [x] Setup wizard shows `StepPresetFiles` after `StepGitRemote`
  - [x] Preset checklist shows only files that exist on the system
  - [x] Space toggles individual preset selection
  - [x] Enter continues to next step with selected presets added to config
  - [x] Custom path entry (`StepAddFiles`) still works after presets
  - [x] Preset detection runs async (no blocking in Update)
  - [x] `go test ./internal/pathutil/ -v -run TestDetectDotfiles` → PASS
  - [x] `go test ./internal/tui/views/ -v -run TestSetup` → PASS
  - [x] `make test` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Preset detection finds existing dotfiles
    Tool: Bash (go test)
    Preconditions: Test creates temp HOME with known dotfiles
    Steps:
      1. Run: go test ./internal/pathutil/ -v -run TestDetectDotfiles
      2. Assert: Creates temp dir with .bashrc, .zshrc, .config/nvim/
      3. Assert: DetectDotfiles returns 3 presets with Exists=true
      4. Assert: Missing presets return Exists=false
      5. Assert: Folder presets include FileCount
    Expected Result: Detection correctly identifies existing dotfiles
    Evidence: Test output captured

  Scenario: Setup wizard shows preset step and allows toggling
    Tool: Bash (go test)
    Preconditions: Test uses temp HOME
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestSetupPresets
      2. Assert: After StepGitRemote, next step is StepPresetFiles
      3. Assert: Space key toggles preset selection
      4. Assert: Enter advances with selected presets in config.Files
      5. Assert: StepAddFiles still accessible after presets
    Expected Result: Preset step integrates into wizard flow
    Evidence: Test output captured

  Scenario: Full setup flow with presets
    Tool: interactive_bash (tmux)
    Preconditions: Binary built, first-time launch (no config)
    Steps:
      1. Launch: XDG_CONFIG_HOME=/tmp/test-dk ./bin/dotkeeper
      2. Assert: Welcome screen appears
      3. Press Enter → backup dir step
      4. Press Enter (accept default) → git remote step
      5. Press Enter (skip) → preset files step
      6. Assert: Checklist of detected dotfiles visible
      7. Press space on first 2 items to toggle
      8. Press Enter → preset folders step (or add files step)
      9. Continue through to completion
      10. Assert: Config saved with selected presets
    Expected Result: Presets integrated into setup flow
    Evidence: Terminal output captured
  ```

  **Commit**: YES
  - Message: `feat(setup): add dotfile preset detection and checklist to setup wizard`
  - Files: `internal/pathutil/presets.go`, `internal/pathutil/presets_test.go`, `internal/tui/views/setup.go`, `internal/tui/views/setup_test.go`
  - Pre-commit: `make test`

---

- [x] 3. Config struct: add `Exclude`, `DisabledFiles`, `DisabledFolders` fields

  **What to do**:
  - Add to `Config` struct in `config.go`:
    ```go
    Exclude         []string `yaml:"exclude,omitempty"`
    DisabledFiles   []string `yaml:"disabled_files,omitempty"`
    DisabledFolders []string `yaml:"disabled_folders,omitempty"`
    ```
  - Update `Validate()`: exclusion patterns are optional (no validation needed beyond non-empty strings)
  - Add `ActiveFiles() []string` method: returns `Files` minus entries in `DisabledFiles`
  - Add `ActiveFolders() []string` method: returns `Folders` minus entries in `DisabledFolders`
  - Update `LoadOrDefault()` default config: empty `Exclude`, `DisabledFiles`, `DisabledFolders`
  - Add unit tests for new fields and methods
  - Use `omitempty` yaml tags so existing configs without these fields load cleanly

  **Must NOT do**:
  - Do NOT change `Files` or `Folders` field types (stay `[]string`)
  - Do NOT add validation for exclusion pattern syntax at config level (that's collector's job)
  - Do NOT remove or rename any existing fields

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Backend config work

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2)
  - **Blocks**: Tasks 6, 7, 9
  - **Blocked By**: Task 0 (pathutil)

  **References**:

  **Pattern References**:
  - `internal/config/config.go:12-19` — Current Config struct to modify
  - `internal/config/config.go:101-115` — Validate() to potentially extend
  - `internal/config/config.go:117-132` — LoadOrDefault() to update

  **Test References**:
  - `internal/config/config_test.go` — Existing config tests to extend

  **Acceptance Criteria**:
  - [x] Config struct has `Exclude`, `DisabledFiles`, `DisabledFolders` fields with `omitempty` yaml tags
  - [x] `ActiveFiles()` returns Files minus DisabledFiles
  - [x] `ActiveFolders()` returns Folders minus DisabledFolders
  - [x] Existing config YAML without new fields loads without error
  - [x] Config YAML with new fields round-trips correctly (save → load)
  - [x] `go test ./internal/config/ -v` → PASS
  - [x] `make test` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Config backward compatibility
    Tool: Bash (go test)
    Preconditions: Test creates old-format YAML
    Steps:
      1. Create YAML: backup_dir + git_remote + files + folders (no exclude/disabled)
      2. LoadFromPath() → assert no error
      3. Assert: Exclude is nil or empty slice
      4. Assert: DisabledFiles is nil or empty slice
      5. Assert: DisabledFolders is nil or empty slice
      6. Assert: ActiveFiles() returns all Files
      7. Assert: ActiveFolders() returns all Folders
    Expected Result: Old configs load cleanly with zero-value new fields
    Evidence: Test output captured

  Scenario: ActiveFiles/ActiveFolders filtering
    Tool: Bash (go test)
    Preconditions: Config with Files, Folders, and disabled entries
    Steps:
      1. Create Config{Files: ["a", "b", "c"], DisabledFiles: ["b"]}
      2. Assert: ActiveFiles() returns ["a", "c"]
      3. Create Config{Folders: ["x", "y"], DisabledFolders: ["x"]}
      4. Assert: ActiveFolders() returns ["y"]
    Expected Result: Disabled entries correctly filtered
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(config): add Exclude, DisabledFiles, DisabledFolders fields with active helpers`
  - Files: `internal/config/config.go`, `internal/config/config_test.go`
  - Pre-commit: `make test`

---

- [x] 4. [A] Embed filepicker in Settings path addition

  **What to do**:
  - Add a `filepicker.Model` field to `SettingsModel` struct
  - Add new state `stateFilePickerActive` to `settingsState` enum
  - When user presses `a` (add) or a new key `b` (browse) in `stateBrowsingFiles` or `stateBrowsingFolders`, transition to `stateFilePickerActive`
  - Configure filepicker: `ShowHidden = true`, `CurrentDirectory = home dir`, `DirAllowed` based on context (files vs folders)
  - When filepicker completes selection (`DidSelectFile`), add the selected path to `config.Files` or `config.Folders`, then return to the previous browsing state
  - On Esc from filepicker, return to previous state
  - Keep text input as fallback: `a` = add via filepicker, `/` = type manually
  - Initialize filepicker in `NewSettings()` (eager init pattern)
  - Update `View()` to render filepicker when in `stateFilePickerActive`
  - Update help text to show `a: Browse | /: Type path`

  **Must NOT do**:
  - Do NOT add more than 1 new state (stateFilePickerActive)
  - Do NOT modify the filepicker library itself
  - Do NOT add file rename/delete/copy operations

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: Interaction pattern for modal filepicker in TUI
  - **Skills Evaluated but Omitted**:
    - `git-master`: Not a git task

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 5, 6, 7)
  - **Blocks**: Task 11 (remove standalone FileBrowser)
  - **Blocked By**: Task 1 (simplified settings state machine)

  **References**:

  **Pattern References**:
  - `internal/tui/views/filebrowser.go:19-28` — How filepicker.New() is configured
  - `internal/tui/views/filebrowser.go:44-47` — `DidSelectFile` pattern for detecting selection
  - `internal/tui/views/settings.go:205-244` — `handleBrowsingFilesInput()` — where to add browse trigger
  - `internal/tui/views/restore.go` — May have filepicker usage pattern for reference

  **API/Type References**:
  - `github.com/charmbracelet/bubbles/filepicker` — `New()`, `Init()`, `Update()`, `View()`, `DidSelectFile()`

  **Test References**:
  - `internal/tui/views/filebrowser_test.go` — Existing filepicker test patterns
  - `internal/tui/views/settings_test.go` — Settings test patterns to extend

  **Acceptance Criteria**:
  - [x] Pressing `a` in file/folder browsing opens filepicker
  - [x] Pressing `/` in file/folder browsing opens text input (manual entry)
  - [x] Selecting a file in filepicker adds to config.Files
  - [x] Selecting a folder adds to config.Folders
  - [x] Esc from filepicker returns to browsing state
  - [x] Filepicker shows hidden files (dotfiles visible)
  - [x] Help text shows `a: Browse | /: Type path`
  - [x] `go test ./internal/tui/views/ -v -run TestSettings` → PASS
  - [x] `make test` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Filepicker opens from settings file browsing
    Tool: interactive_bash (tmux)
    Preconditions: Binary built, config exists
    Steps:
      1. Launch TUI
      2. Navigate to Settings (s)
      3. Navigate to Files entry, press Enter
      4. Press 'a' to add via browse
      5. Assert: Filepicker visible with directory listing
      6. Navigate to a dotfile, press Enter to select
      7. Assert: Returns to file list with new entry added
      8. Press Esc to go back
    Expected Result: Filepicker integrates into settings flow
    Evidence: Terminal output captured

  Scenario: Manual path entry still works
    Tool: interactive_bash (tmux)
    Preconditions: Binary built
    Steps:
      1. Navigate to Settings → Files
      2. Press '/' for manual entry
      3. Type: ~/.bashrc
      4. Press Enter
      5. Assert: Path added to files list
    Expected Result: Text input fallback works alongside filepicker
    Evidence: Terminal output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): embed filepicker in settings for visual path browsing`
  - Files: `internal/tui/views/settings.go`, `internal/tui/views/settings_test.go`
  - Pre-commit: `make test`

---

- [x] 5. [A] Embed filepicker in Setup Wizard path addition

  **What to do**:
  - Add a `filepicker.Model` field to `SetupModel` struct
  - Add a `browsing bool` flag to track when filepicker is active during StepAddFiles/StepAddFolders
  - In StepAddFiles and StepAddFolders, pressing `b` opens filepicker instead of text input
  - When filepicker completes selection, add to `addedFiles` or `addedFolders`, return to text input mode
  - Keep existing text input for manual entry (Enter with typed path)
  - Update View() to render filepicker when `browsing` is true
  - Update help text: `Enter: add typed path | b: Browse | Empty Enter: continue`
  - Initialize filepicker in NewSetup() (eager init)
  - Filepicker `ShowHidden = true`, starts at `$HOME`

  **Must NOT do**:
  - Do NOT change the preset step flow (Task 2)
  - Do NOT remove text input — it must coexist with filepicker
  - Do NOT modify wizard step progression

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: Wizard UX with dual input modes
  - **Skills Evaluated but Omitted**:
    - `git-master`: Not a git task

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 6, 7)
  - **Blocks**: Task 11 (remove standalone FileBrowser)
  - **Blocked By**: Task 2 (setup wizard changes from presets)

  **References**:

  **Pattern References**:
  - `internal/tui/views/setup.go:30-55` — SetupModel struct to extend
  - `internal/tui/views/setup.go:87-91` — Input handling for StepAddFiles/StepAddFolders
  - `internal/tui/views/setup.go:214-252` — View rendering for add files/folders steps
  - `internal/tui/views/filebrowser.go:19-28` — Filepicker config pattern

  **Test References**:
  - `internal/tui/views/setup_test.go:129-186` — TestSetupAddFiles — extend for browse mode
  - `internal/tui/views/filebrowser_test.go` — Filepicker test pattern

  **Acceptance Criteria**:
  - [x] Pressing `b` in StepAddFiles/StepAddFolders opens filepicker
  - [x] Selecting a file/folder in filepicker adds to respective list
  - [x] Esc from filepicker returns to text input mode
  - [x] Manual text entry still works (type path + Enter)
  - [x] Help text shows both options
  - [x] `go test ./internal/tui/views/ -v -run TestSetup` → PASS
  - [x] `make test` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Browse mode in setup wizard
    Tool: Bash (go test)
    Preconditions: Test with temp HOME containing known dotfiles
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestSetupBrowseMode
      2. Assert: Pressing 'b' at StepAddFiles sets browsing=true
      3. Assert: Filepicker rendered in View()
      4. Assert: File selection adds to addedFiles
      5. Assert: Esc returns browsing=false
      6. Assert: Text input still works after browse
    Expected Result: Dual input modes work in setup wizard
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(setup): add file browser support for visual path selection during setup`
  - Files: `internal/tui/views/setup.go`, `internal/tui/views/setup_test.go`
  - Pre-commit: `make test`

---

- [x] 6. [E] Collector exclusion pattern support

  **What to do**:
  - Update `CollectFiles()` signature to: `CollectFiles(paths []string, excludePatterns []string) ([]FileInfo, error)`
  - Add `shouldExclude(path string, patterns []string) bool` helper using `filepath.Match()`
  - In `collectPath()`, before adding a file to results, check against exclude patterns
  - Match patterns against both the full path and the base name (e.g., `*.log` matches `/path/to/file.log`)
  - Support directory exclusion patterns ending in `/`: `node_modules/` skips entire directories
  - Update `backup.go:42` to pass `cfg.Exclude` to `CollectFiles()`
  - Use `cfg.ActiveFiles()` and `cfg.ActiveFolders()` instead of raw `cfg.Files`/`cfg.Folders`
  - Add comprehensive collector tests with temp directories containing files that should/shouldn't be excluded

  **Must NOT do**:
  - Do NOT add regex support — `filepath.Match()` glob patterns only
  - Do NOT add gitignore negation patterns (`!pattern`)
  - Do NOT modify the archive or encryption logic
  - Do NOT add recursive glob `**` support here (that's Task 10)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Backend logic, not UI

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5, 7)
  - **Blocks**: Task 10 (glob resolution depends on exclusion being in place)
  - **Blocked By**: Task 3 (config Exclude field)

  **References**:

  **Pattern References**:
  - `internal/backup/collector.go:25-42` — `CollectFiles()` to modify
  - `internal/backup/collector.go:44-107` — `collectPath()` — add exclude check
  - `internal/backup/backup.go:42` — Caller of CollectFiles — update to pass excludes

  **API/Type References**:
  - `filepath.Match(pattern, name)` — Go stdlib glob matching
  - `internal/config/config.go` — `ActiveFiles()`, `ActiveFolders()`, `Exclude` field (from Task 3)

  **Test References**:
  - `internal/backup/collector_test.go` — Extend with exclusion tests
  - `internal/backup/backup_test.go` — Update backup test to pass excludes

  **Acceptance Criteria**:
  - [x] `CollectFiles(paths, excludes)` filters out files matching exclude patterns
  - [x] `*.log` pattern excludes all `.log` files
  - [x] `node_modules/` pattern skips entire `node_modules` directories
  - [x] Pattern matching works on both full path and base name
  - [x] `backup.go` uses `cfg.ActiveFiles()` + `cfg.ActiveFolders()` and passes `cfg.Exclude`
  - [x] `go test ./internal/backup/ -v -run TestCollectFiles` → PASS
  - [x] `make test` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Exclusion patterns filter files correctly
    Tool: Bash (go test)
    Preconditions: Tests create temp dirs with known file structures
    Steps:
      1. Run: go test ./internal/backup/ -v -run TestCollectFilesWithExclusions
      2. Assert: Test creates: dir/a.txt, dir/b.log, dir/c.txt, dir/sub/d.log
      3. Assert: CollectFiles(["dir"], ["*.log"]) returns only a.txt, c.txt
      4. Assert: CollectFiles(["dir"], ["sub/"]) returns a.txt, b.log, c.txt
      5. Assert: CollectFiles(["dir"], []) returns all 4 files
    Expected Result: Exclusion patterns work correctly
    Evidence: Test output captured

  Scenario: Backup uses active files and excludes
    Tool: Bash (go test)
    Preconditions: Config with disabled files and exclude patterns
    Steps:
      1. Run: go test ./internal/backup/ -v -run TestBackupRespectsExclusions
      2. Assert: Disabled files not collected
      3. Assert: Excluded patterns not in archive
    Expected Result: Full backup pipeline respects exclusions
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(backup): add exclusion pattern support to file collector`
  - Files: `internal/backup/collector.go`, `internal/backup/collector_test.go`, `internal/backup/backup.go`, `internal/backup/backup_test.go`
  - Pre-commit: `make test`

---

- [x] 7. [D] Backup content preview in Settings + Dashboard

  **What to do**:
  - Create `internal/pathutil/scanner.go` with `ScanPaths(files, folders []string, exclude []string) ScanResult`
  - `ScanResult` struct: `TotalFiles int`, `TotalSize int64`, `BrokenPaths []string`, `PathStats []PathStat`
  - `PathStat` struct: `Path string`, `Exists bool`, `IsDir bool`, `FileCount int`, `Size int64`
  - All scanning runs async via `tea.Cmd` closures
  - **Dashboard**: Replace `fileCount = len(Files) + len(Folders)` with actual file count from async scan
  - Add new card to dashboard: "Total Size" showing human-readable size (e.g., "1.2 MB")
  - Show warning indicator on dashboard if broken paths detected
  - **Settings**: When browsing Files or Folders list, show per-path stats in the description field:
    - Files: `~/.zshrc (2.1 KB, modified 2024-01-15)` or `~/.old (NOT FOUND)`
    - Folders: `~/.config/nvim (45 files, 890 KB)` or `~/.missing/ (NOT FOUND)`
  - Broken paths highlighted with error styling (red)
  - Add spinner or "Scanning..." while async scan is in progress

  **Must NOT do**:
  - Do NOT show file contents (only metadata)
  - Do NOT block in Update() — async scanning only
  - Do NOT add file content diff or comparison

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: Dashboard cards and info display design
  - **Skills Evaluated but Omitted**:
    - `git-master`: Not a git task

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5, 6)
  - **Blocks**: None
  - **Blocked By**: Task 3 (needs ActiveFiles/ActiveFolders and Exclude)

  **References**:

  **Pattern References**:
  - `internal/tui/views/dashboard.go:149-168` — `refreshStatus()` async pattern — enhance this
  - `internal/tui/views/dashboard.go:79-142` — Dashboard View() — add size card and warning
  - `internal/tui/views/settings.go:456-472` — `refreshFilesList()` — enhance with path stats
  - `internal/tui/views/settings.go:37-46` — `subSettingItem` — use desc field for stats

  **Test References**:
  - `internal/tui/views/dashboard_test.go` — Extend for new cards
  - `internal/pathutil/pathutil_test.go` — Add scanner tests

  **Acceptance Criteria**:
  - [x] `internal/pathutil/scanner.go` exists with `ScanPaths()` function
  - [x] Dashboard shows actual file count (not config entry count)
  - [x] Dashboard shows total backup size in human-readable format
  - [x] Dashboard shows warning if any broken paths detected
  - [x] Settings file/folder list shows per-path size and status
  - [x] Broken paths shown in red/error styling
  - [x] Scanning is async (non-blocking)
  - [x] `go test ./internal/pathutil/ -v -run TestScanPaths` → PASS
  - [x] `go test ./internal/tui/views/ -v -run TestDashboard` → PASS
  - [x] `make test` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Scanner reports correct stats
    Tool: Bash (go test)
    Preconditions: Temp dir with known files and one broken path
    Steps:
      1. Run: go test ./internal/pathutil/ -v -run TestScanPaths
      2. Assert: Creates temp with 3 files (100B, 200B, 300B) and 1 broken path
      3. Assert: TotalFiles == 3, TotalSize == 600
      4. Assert: BrokenPaths contains the broken path
      5. Assert: Per-path stats have correct Exists, Size, FileCount
    Expected Result: Scanner accurately reports filesystem state
    Evidence: Test output captured

  Scenario: Dashboard shows real file counts
    Tool: interactive_bash (tmux)
    Preconditions: Binary built, config with valid paths
    Steps:
      1. Launch TUI
      2. Assert: "Files Tracked" card shows actual file count (not config entries)
      3. Assert: Total size card visible
      4. Press 'q' to quit
    Expected Result: Dashboard displays real metrics
    Evidence: Terminal output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): add backup content preview with real file counts and broken path detection`
  - Files: `internal/pathutil/scanner.go`, `internal/pathutil/scanner_test.go`, `internal/tui/views/dashboard.go`, `internal/tui/views/dashboard_test.go`, `internal/tui/views/settings.go`, `internal/tui/views/settings_test.go`
  - Pre-commit: `make test`

---

- [x] 8. [C] Tab-completion autocomplete for path input

  **What to do**:
  - Create `internal/tui/components/pathcomplete.go` with a `PathCompleter` component
  - `PathCompleter` wraps `textinput.Model` and adds Tab key handling
  - On Tab press: read current input value, find the directory portion, list entries matching prefix
  - If single match: complete the full path. If multiple: complete common prefix + show options below input
  - Show completion candidates as a dimmed list below the text input (max 10 visible)
  - Pressing Tab cycles through candidates if multiple matches
  - Integrate into Settings `stateEditingSubItem` and Setup `StepAddFiles`/`StepAddFolders`
  - Filesystem listing runs via `tea.Cmd` to avoid blocking
  - Handle `~/` expansion in completion context
  - Show path type indicator: `[F]` for file, `[D]` for directory

  **Must NOT do**:
  - Do NOT add fuzzy matching — prefix match only
  - Do NOT show more than 10 candidates at once
  - Do NOT block in Update() — all fs reads via `tea.Cmd`
  - Do NOT add mouse support for candidate selection

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: Custom TUI component design
  - **Skills Evaluated but Omitted**:
    - None significant

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9, 10, 11)
  - **Blocks**: Task 12
  - **Blocked By**: Task 0 (uses pathutil.ExpandHome)

  **References**:

  **Pattern References**:
  - `internal/tui/views/settings.go:287-312` — `handleEditingFieldInput()` — where Tab key is intercepted
  - `internal/tui/views/settings.go:314-340` — `handleEditingSubItemInput()` — integrate autocomplete
  - `internal/tui/views/setup.go:87-91` — Setup input handling — integrate autocomplete
  - `internal/tui/components/tabbar_test.go` — Existing component test pattern

  **API/Type References**:
  - `github.com/charmbracelet/bubbles/textinput` — Wrapping this component
  - `os.ReadDir()` — For listing directory contents

  **Test References**:
  - `internal/tui/components/tabbar_test.go` — Component test structure

  **Acceptance Criteria**:
  - [x] `internal/tui/components/pathcomplete.go` exists
  - [x] Tab key triggers path completion in settings and setup path inputs
  - [x] Single match auto-completes full path
  - [x] Multiple matches show candidate list (max 10)
  - [x] `~/` is handled correctly in completion context
  - [x] Completion runs async (no blocking)
  - [x] `go test ./internal/tui/components/ -v -run TestPathComplete` → PASS
  - [x] `make test` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Tab completion works with single match
    Tool: Bash (go test)
    Preconditions: Temp dir with unique prefix files
    Steps:
      1. Run: go test ./internal/tui/components/ -v -run TestPathCompleteSingle
      2. Assert: Input "~/temp/uni" + Tab → completes to "~/temp/unique-file.txt"
    Expected Result: Single match auto-completes
    Evidence: Test output captured

  Scenario: Tab completion shows candidates for multiple matches
    Tool: Bash (go test)
    Preconditions: Temp dir with files sharing prefix
    Steps:
      1. Run: go test ./internal/tui/components/ -v -run TestPathCompleteMultiple
      2. Assert: Input "~/temp/f" + Tab → shows "file1.txt", "file2.txt", "folder/"
      3. Assert: Common prefix completed: "~/temp/f" → "~/temp/fi" if common
    Expected Result: Multiple matches display candidates
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): add tab-completion autocomplete for path inputs`
  - Files: `internal/tui/components/pathcomplete.go`, `internal/tui/components/pathcomplete_test.go`, `internal/tui/views/settings.go`, `internal/tui/views/setup.go`
  - Pre-commit: `make test`

---

- [x] 9. [H] Inline actions per path item

  **What to do**:
  - In Settings `stateBrowsingFiles` and `stateBrowsingFolders`, add new key bindings:
    - `Space` — toggle enable/disable: adds/removes path from `config.DisabledFiles`/`DisabledFolders`
    - `i` — inspect: show path details panel (size, mod time, permissions, is symlink?, target if symlink)
    - `d` — delete (already exists, keep as-is)
    - `Enter` — edit (already exists, keep as-is)
  - Update `subSettingItem` to include `disabled bool` field
  - Disabled items rendered with strikethrough or dimmed styling + `[disabled]` indicator
  - Inspect shows a temporary overlay or inline expansion below the selected item:
    ```
    ~/.zshrc
      Size: 2.1 KB | Modified: 2024-01-15 | Permissions: 644
      Type: regular file
    ```
  - Inspect panel dismisses on any key press
  - Update `refreshFilesList()`/`refreshFoldersList()` to check DisabledFiles/DisabledFolders
  - After toggling or deleting, auto-save is NOT triggered (user must still press `s`)
  - Update help text: `Space: toggle | i: inspect | d: delete | Enter: edit | a: browse | /: type`

  **Must NOT do**:
  - Do NOT add drag-and-drop or reordering
  - Do NOT auto-save on toggle (user must explicitly save)
  - Do NOT add undo/redo for actions
  - Do NOT show file contents in inspect panel

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: Inline action UX patterns
  - **Skills Evaluated but Omitted**:
    - None significant

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 10, 11)
  - **Blocks**: Task 12
  - **Blocked By**: Tasks 1 (simplified settings states), 3 (DisabledFiles/DisabledFolders config fields)

  **References**:

  **Pattern References**:
  - `internal/tui/views/settings.go:205-244` — `handleBrowsingFilesInput()` — add Space, i bindings
  - `internal/tui/views/settings.go:246-285` — `handleBrowsingFoldersInput()` — same
  - `internal/tui/views/settings.go:456-472` — `refreshFilesList()` — add disabled flag
  - `internal/tui/views/styles.go` — Styling for disabled items

  **API/Type References**:
  - `internal/config/config.go` — `DisabledFiles`, `DisabledFolders`, `ActiveFiles()`, `ActiveFolders()` (from Task 3)

  **Test References**:
  - `internal/tui/views/settings_test.go` — Extend for toggle and inspect actions

  **Acceptance Criteria**:
  - [x] Space toggles path between enabled/disabled state
  - [x] Disabled paths shown with dimmed styling and `[disabled]` marker
  - [x] `i` key shows inspect panel with file metadata
  - [x] Inspect panel dismisses on next key press
  - [x] Toggling does NOT auto-save (requires `s` key)
  - [x] Help text updated with all available actions
  - [x] `go test ./internal/tui/views/ -v -run TestSettings` → PASS
  - [x] `make test` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Toggle enable/disable path
    Tool: Bash (go test)
    Preconditions: Settings with Files list containing 3 items
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestSettingsToggleDisable
      2. Assert: Space on item adds to config.DisabledFiles
      3. Assert: View shows [disabled] marker on item
      4. Assert: Space again removes from DisabledFiles
      5. Assert: Save required for persistence (no auto-save)
    Expected Result: Toggle correctly manages disabled state
    Evidence: Test output captured

  Scenario: Inspect panel shows file metadata
    Tool: Bash (go test)
    Preconditions: Settings with Files containing real paths
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestSettingsInspect
      2. Assert: 'i' key on file item shows size, mod time, permissions
      3. Assert: Any subsequent key dismisses inspect panel
      4. Assert: State returns to normal browsing
    Expected Result: Inspect shows metadata and dismisses cleanly
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(tui): add inline actions (toggle, inspect, delete) for path items in settings`
  - Files: `internal/tui/views/settings.go`, `internal/tui/views/settings_test.go`
  - Pre-commit: `make test`

---

- [x] 10. [G] Bulk glob resolution

  **What to do**:
  - Create `internal/pathutil/glob.go` with `ResolveGlob(pattern string, exclude []string) ([]string, error)`
  - Support standard `filepath.Glob` patterns: `*`, `?`, `[...]`
  - Add `doublestar` library for `**` recursive glob support: `go get github.com/bmatcuk/doublestar/v4`
  - Cap results at 500 paths with clear error message if exceeded
  - In Settings/Setup text input: detect if typed value contains glob characters (`*`, `?`, `[`)
  - If glob detected, resolve and show confirmation: `"~/.config/*.lua" → 12 files matched. Add all? [y/n]`
  - Add new sub-state or confirmation step for glob review
  - Show resolved paths in a scrollable list before confirming
  - On confirm: add all resolved paths individually to config
  - Pass exclude patterns from config to glob resolution

  **Must NOT do**:
  - Do NOT add results beyond 500 path cap
  - Do NOT auto-add without confirmation
  - Do NOT modify the collector's file collection logic (that handles flat paths)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Mostly logic work with minimal UI (confirmation dialog)

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 9, 11)
  - **Blocks**: Task 12
  - **Blocked By**: Task 6 (exclusion patterns needed for glob filtering)

  **References**:

  **Pattern References**:
  - `internal/tui/views/settings.go:314-340` — Sub-item editing where glob detection happens
  - `internal/tui/views/setup.go:121-154` — Setup add steps where glob detection happens

  **External References**:
  - `github.com/bmatcuk/doublestar/v4` — Recursive glob library for Go

  **Test References**:
  - `internal/pathutil/pathutil_test.go` — Extend for glob tests

  **Acceptance Criteria**:
  - [x] `internal/pathutil/glob.go` exists with `ResolveGlob()` function
  - [x] Standard glob patterns work: `*.lua`, `config-?.yaml`
  - [x] Recursive `**` patterns work via doublestar
  - [x] Results capped at 500 with clear warning
  - [x] Glob in text input triggers confirmation dialog showing resolved paths
  - [x] Confirmed paths added individually to config
  - [x] Exclude patterns respected during resolution
  - [x] `go test ./internal/pathutil/ -v -run TestResolveGlob` → PASS
  - [x] `make test` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Glob resolves and caps correctly
    Tool: Bash (go test)
    Preconditions: Temp dir with known file structure
    Steps:
      1. Run: go test ./internal/pathutil/ -v -run TestResolveGlob
      2. Assert: "*.txt" in dir with 3 .txt files → returns 3 paths
      3. Assert: Exclude ["b.txt"] → returns 2 paths
      4. Assert: Pattern matching 600 files → error about 500 cap
    Expected Result: Glob resolution works with cap and exclusions
    Evidence: Test output captured

  Scenario: Glob in Settings triggers confirmation
    Tool: Bash (go test)
    Preconditions: Settings model in editing state
    Steps:
      1. Run: go test ./internal/tui/views/ -v -run TestSettingsGlobConfirmation
      2. Assert: Typing "*.lua" + Enter → confirmation view showing matched files
      3. Assert: Confirming adds all matched paths to config
      4. Assert: Canceling returns to input without adding
    Expected Result: Glob UI flow works correctly
    Evidence: Test output captured
  ```

  **Commit**: YES
  - Message: `feat(pathutil): add bulk glob resolution with confirmation UI`
  - Files: `internal/pathutil/glob.go`, `internal/pathutil/glob_test.go`, `internal/tui/views/settings.go`, `internal/tui/views/setup.go`, `go.mod`, `go.sum`
  - Pre-commit: `make test`

---

- [x] 11. Remove standalone FileBrowserView tab

  **What to do**:
  - Remove `FileBrowserView` from `ViewState` enum in `internal/tui/model.go`
  - Remove `fileBrowser views.FileBrowserModel` field from main `Model` struct
  - Remove `FileBrowserView` case from `Update()` routing in `internal/tui/update.go`
  - Remove `FileBrowserView` case from `View()` rendering in `internal/tui/view.go`
  - Remove `FileBrowserView` from tab bar configuration
  - Decrement `viewCount` if used for tab cycling
  - DELETE `filebrowser.go` and `filebrowser_test.go` — the embedded filepicker in Settings/Setup (Tasks 4, 5) uses `charmbracelet/bubbles/filepicker` directly, not `FileBrowserModel`. Verify with `lsp_find_references` on `FileBrowserModel` before deletion to confirm zero remaining references.
  - Update any tests that reference `FileBrowserView`

  **Must NOT do**:
  - Do NOT remove `filepicker` usage from Settings/Setup (those use embedded filepicker, not FileBrowserModel)
  - Do NOT break tab cycling

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`git-master`]
    - `git-master`: Clean removal commit
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Simple removal, not design

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 9, 10)
  - **Blocks**: Task 12
  - **Blocked By**: Tasks 4, 5 (filepicker embedded before removing standalone)

  **References**:

  **Pattern References**:
  - `internal/tui/model.go` — ViewState enum, Model struct
  - `internal/tui/update.go` — Update routing switch
  - `internal/tui/view.go` — View rendering switch
  - `internal/tui/views/filebrowser.go` — The file to potentially delete

  **Tool Recommendations**:
  - Use `lsp_find_references` on `FileBrowserView` and `FileBrowserModel` before removal
  - Use `grep` to find any remaining references

  **Acceptance Criteria**:
  - [x] `FileBrowserView` not in ViewState enum
  - [x] Tab cycling skips old FileBrowser position
  - [x] No compilation errors
  - [x] `make test` → PASS
  - [x] `make lint` → PASS

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Tab cycling works without FileBrowser
    Tool: interactive_bash (tmux)
    Preconditions: Binary built
    Steps:
      1. Launch TUI
      2. Press Tab repeatedly through all views
      3. Assert: Dashboard → BackupList → Restore → Settings → Logs → Dashboard
      4. Assert: No FileBrowser tab appears
    Expected Result: Tab cycling is clean without standalone browser
    Evidence: Terminal output captured

  Scenario: No references to removed view
    Tool: Bash (grep)
    Steps:
      1. Run: grep -rn "FileBrowserView\|FileBrowserModel" internal/ --include="*.go"
      2. Assert: Zero results (or only in deleted file)
    Expected Result: Clean removal with no dangling references
    Evidence: grep output captured
  ```

  **Commit**: YES
  - Message: `refactor(tui): remove standalone FileBrowserView tab (superseded by embedded filepicker)`
  - Files: `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/view.go`, `internal/tui/views/filebrowser.go` (deleted), `internal/tui/views/filebrowser_test.go` (deleted)
  - Pre-commit: `make test`

---

- [x] 12. Full test suite + lint pass + cleanup

  **What to do**:
  - Run `make test` — fix any failures
  - Run `make lint` — fix any lint issues
  - Run `go vet ./...` — fix any vet issues
  - Verify all new files have proper package declarations and imports
  - Verify no unused imports or variables
  - Run full E2E test suite: `go test ./e2e/ -v`
  - Verify that `make build` produces a working binary
  - Quick TUI smoke test: launch binary, navigate all tabs, verify no panics

  **Must NOT do**:
  - Do NOT delete or skip failing tests to make build pass
  - Do NOT add `//nolint` without justification

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`git-master`]
    - `git-master`: Final cleanup commit
  - **Skills Evaluated but Omitted**:
    - None needed for verification

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4 (solo, final)
  - **Blocks**: None (final task)
  - **Blocked By**: Tasks 8, 9, 10, 11

  **References**:

  **Pattern References**:
  - `Makefile` — `make test`, `make lint`, `make build` targets

  **Acceptance Criteria**:
  - [x] `make test` → PASS (zero failures)
  - [x] `make lint` → PASS (zero issues)
  - [x] `go vet ./...` → PASS
  - [x] `make build` → produces `./bin/dotkeeper`
  - [x] `go test ./e2e/ -v` → PASS
  - [x] TUI launches without panic

  **Agent-Executed QA Scenarios:**

  ```
  Scenario: Full verification suite
    Tool: Bash
    Preconditions: All previous tasks completed
    Steps:
      1. Run: make test
      2. Assert: exit code 0, zero FAIL lines
      3. Run: make lint
      4. Assert: exit code 0
      5. Run: go vet ./...
      6. Assert: no output (clean)
      7. Run: make build
      8. Assert: ./bin/dotkeeper exists and is executable
      9. Run: go test ./e2e/ -v -timeout 120s
      10. Assert: all E2E tests pass
    Expected Result: Project is fully green
    Evidence: Terminal output captured for each command

  Scenario: TUI smoke test
    Tool: interactive_bash (tmux)
    Preconditions: Binary built
    Steps:
      1. Launch: XDG_CONFIG_HOME=/tmp/dk-smoke ./bin/dotkeeper
      2. Complete setup wizard (accept defaults)
      3. Tab through all views: Dashboard → BackupList → Restore → Settings → Logs
      4. Navigate to Settings, browse files list
      5. Press 'q' to quit
      6. Assert: No panics in terminal output
    Expected Result: TUI is stable through all views
    Evidence: Terminal output captured
  ```

  **Commit**: YES (if any fixes needed)
  - Message: `chore: fix test/lint issues from path selection UX changes`
  - Pre-commit: `make test && make lint`

---

## Commit Strategy

| After Task | Message | Key Files | Verification |
|------------|---------|-----------|--------------|
| 0 | `refactor(pathutil): extract shared ExpandHome to internal/pathutil` | pathutil.go, helpers.go, collector.go | `make test` |
| 1 | `feat(tui): remove read-only mode from settings` | settings.go | `make test` |
| 2 | `feat(setup): add dotfile preset detection and checklist` | presets.go, setup.go | `make test` |
| 3 | `feat(config): add Exclude, DisabledFiles, DisabledFolders` | config.go | `make test` |
| 4 | `feat(tui): embed filepicker in settings` | settings.go | `make test` |
| 5 | `feat(setup): add file browser to setup wizard` | setup.go | `make test` |
| 6 | `feat(backup): add exclusion pattern support` | collector.go, backup.go | `make test` |
| 7 | `feat(tui): add backup content preview` | scanner.go, dashboard.go, settings.go | `make test` |
| 8 | `feat(tui): add tab-completion for path inputs` | pathcomplete.go, settings.go | `make test` |
| 9 | `feat(tui): add inline actions for path items` | settings.go | `make test` |
| 10 | `feat(pathutil): add bulk glob resolution` | glob.go, settings.go | `make test` |
| 11 | `refactor(tui): remove standalone FileBrowserView` | model.go, update.go, view.go | `make test` |
| 12 | `chore: fix test/lint issues` | various | `make test && make lint` |

---

## Success Criteria

### Verification Commands
```bash
make test          # Expected: PASS, zero failures
make lint          # Expected: PASS, zero issues
make build         # Expected: ./bin/dotkeeper produced
go test ./e2e/ -v  # Expected: all E2E tests pass
go vet ./...       # Expected: clean
```

### Final Checklist
- [x] All "Must Have" features present and functional
- [x] All "Must NOT Have" guardrails respected
- [x] All 13 tasks completed with passing tests
- [x] No blocking calls in any Update() method
- [x] Config backward-compatible
- [x] `expandHome` duplication eliminated
- [x] Standalone FileBrowser tab removed
- [x] All new files have tests
