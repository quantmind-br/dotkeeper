# TUI VIEWS KNOWLEDGE BASE

**Scope:** `internal/tui/views/`

## OVERVIEW

BubbleTea view implementations. Each file = one view model. 15 files, ~3,500 lines.

## STRUCTURE

```
views/
├── dashboard.go        # Main dashboard with stats/actions
├── backuplist.go       # List of backups with selection
├── restore.go          # Restore workflow (largest: 583 lines)
├── settings.go         # Configuration editor
├── setup.go            # First-time setup wizard
├── logs.go             # Backup/restore log viewer
├── filebrowser.go      # File selection component
└── helpers.go          # Shared view utilities
```

## VIEW PATTERN

Each view implements `tea.Model`:

```go
type XxxModel struct {
    config *config.Config
    width, height int
    // view-specific state
}

func (m XxxModel) Init() tea.Cmd { return nil }
func (m XxxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { ... }
func (m XxxModel) View() string { ... }
```

## MESSAGES

**Private** (internal to view):
- `statusMsg` - async operation status
- `backupsLoadedMsg` - data loaded

**Public** (cross-view):
- `BackupSuccessMsg` / `BackupErrorMsg`
- `RefreshBackupListMsg`
- `SetupCompleteMsg`

## CONVENTIONS

- **Eager init**: All views created in `NewModel()`, not on-demand
- **Type assertion**: After `Update()`, cast back: `m.dashboard = model.(views.DashboardModel)`
- **Bubbles components**: Use `list`, `textinput`, `viewport` from charmbracelet/bubbles
- **Async I/O**: Return closures: `return func() tea.Msg { /* blocking */ }`

## WHERE TO LOOK

| Task | File | Notes |
|------|------|-------|
| Add new view | Create `views/xxx.go` | Add to `ViewState`, wire in `update.go`/`view.go` |
| Modify backup list | `backuplist.go` | Uses `list.Model` from bubbles |
| Fix restore UI | `restore.go` | Complex: file browser + conflict resolution |
| Change styling | `../styles.go` | Lipgloss styles centralized |

## ANTI-PATTERNS

- **Never** block in `Update()` - always use `tea.Cmd`
- **Never** access globals - pass config via constructor
- **Never** modify other views' state directly - use messages
