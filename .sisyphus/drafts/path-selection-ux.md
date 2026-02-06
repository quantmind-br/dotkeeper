# Draft: Path Selection UX Improvements

## Requirements (confirmed)
- User wants ALL 8 suggestions planned (A through H)
- Prioritization defined: P0 (B, A), P1 (F, D), P2 (E, C, H), P3 (G)

## Technical Decisions
- Test infrastructure exists: 30 test files, Go standard `testing` + BubbleTea test pattern
- Config struct: `Files []string` + `Folders []string` (flat lists)
- TUI pattern: BubbleTea with eager-initialized sub-models, type assertions after Update
- File Browser: `charmbracelet/bubbles/filepicker` already imported and implemented but NOT wired to settings/setup
- Validation: `ValidateFilePath()` / `ValidateFolderPath()` in helpers.go — validates on submit only

## Research Findings
- `filebrowser.go` exists with working filepicker but disconnected from add-path flows
- Settings has 6 navigation states — overly deep for simple CRUD
- Dashboard `fileCount` counts config entries not actual files
- Setup wizard is linear text-input-only — no interactive selection
- Collector merges Files+Folders into single `CollectFiles(paths)` — no exclusion support
- No glob/pattern support anywhere in config or collector

## Suggestions Summary
- **A**: Integrate File Browser into path addition flow
- **B**: Dotfile presets in Setup Wizard (auto-detect common dotfiles)
- **C**: Real-time validation + autocomplete during typing
- **D**: Preview of backup content (real file count, sizes, broken paths)
- **E**: Exclusion patterns in config + collector + TUI
- **F**: Eliminate stateReadOnly in Settings (direct edit mode)
- **G**: Bulk add with glob resolution
- **H**: Inline actions per path item (inspect, toggle, etc.)

## Open Questions
- Should all 8 be in one plan or only selected ones?
- Test strategy: TDD or tests-after for TUI changes?
- Backward compatibility for config struct changes (E)?

## Scope Boundaries
- INCLUDE: All 8 suggestions
- EXCLUDE: TBD based on user preference
