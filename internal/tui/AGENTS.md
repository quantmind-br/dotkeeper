# TUI CORE KNOWLEDGE BASE

**Scope:** `internal/tui/` (not views/)

## OVERVIEW

BubbleTea framework layer: model initialization, update loop, view switching, and help system. 4 files, ~430 lines.

## STRUCTURE

```
tui/
├── model.go      # Main Model struct, view state, NewModel()
├── update.go     # Update() loop, key bindings, message routing
├── view.go       # View() rendering, title bar, help overlay
├── help.go       # Help system, global bindings
└── views/        # [See views/AGENTS.md for view implementations]
```

## ARCHITECTURE

### Eager-Initialized Model

```go
type Model struct {
    state       ViewState          // DashboardView, BackupListView, etc.
    dashboard   views.DashboardModel  // All views created in NewModel()
    backupList  views.BackupListModel
    restore     views.RestoreModel
    // ... all views always in memory
}
```

**Why eager?** Fast view switching, no re-init overhead.

### View State Machine

```go
const (
    DashboardView ViewState = iota  // Tab cycles through
    FileBrowserView
    BackupListView
    RestoreView
    SettingsView
    LogsView
)
```

### Update Loop Flow

```
KeyMsg/WindowSizeMsg/CustomMsg
    ↓
[Global handlers] (quit, tab, help)
    ↓
[State-specific handler] → delegates to sub-view
    ↓
Type assertion back: m.dashboard = model.(views.DashboardModel)
```

### Type Assertions Required

After every `Update()` call, cast back to concrete type:

```go
var model tea.Model
model, cmd = m.dashboard.Update(msg)
m.dashboard = model.(views.DashboardModel)  // REQUIRED
```

## KEY BINDINGS

```go
var keys = KeyMap{
    Quit: key.NewBinding(key.WithKeys("q", "ctrl+c")),
    Tab:  key.NewBinding(key.WithKeys("tab")),      // Next view
    Help: key.NewBinding(key.WithKeys("?")),         // Toggle help
}
```

Dashboard shortcuts:
- `b` → BackupListView
- `r` → RestoreView
- `s` → SettingsView

## SETUP MODE

If config doesn't exist on startup:
- `setupMode = true`
- Shows SetupView only
- On `SetupCompleteMsg`: initialize all views, switch to Dashboard

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add global key binding | `update.go:DefaultKeyMap()` | Add to KeyMap struct |
| Add new view state | `model.go:ViewState` | Increment viewCount in update.go |
| Wire new view | `model.go` + `update.go` + `view.go` | Add to Model struct + all switches |
| Modify help overlay | `help.go` | Global + per-view help |
| Change title/header | `view.go` | Title style and rendering |

## ANTI-PATTERNS

- **Never** block in `Update()` - return `tea.Cmd` for async
- **Never** forget type assertion after sub-view Update()
- **Never** access globals - pass via Model constructor
- **Don't** lazy-initialize views - eager init in NewModel()

## MESSAGES

**Handled at framework level:**
- `tea.WindowSizeMsg` - propagated to all views
- `tea.KeyMsg` - global bindings first, then routed
- `views.SetupCompleteMsg` - exits setup mode
- `views.RefreshBackupListMsg` - triggers refresh

**Routed to active view only:**
- All other messages go to current state's view
