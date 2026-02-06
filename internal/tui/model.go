package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
	"github.com/diogo/dotkeeper/internal/tui/components"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

// ViewState represents the current view
type ViewState int

const (
	DashboardView ViewState = iota
	BackupListView
	RestoreView
	SettingsView
	LogsView
	SetupView
)

// tabOrder defines the views accessible via tabs (excludes FileBrowser and Setup)
var tabOrder = []ViewState{DashboardView, BackupListView, RestoreView, SettingsView, LogsView}

// mainChromeHeight is the total terminal rows consumed by the main view frame:
// app title (1) + tab bar (1) + blank after tabs (1) + blank after content (1)
// + view help (1) + global help (1) + trailing newline (1) = 7.
const mainChromeHeight = 7

// Model represents the main TUI model
type Model struct {
	state       ViewState
	width       int
	height      int
	quitting    bool
	err         error
	showingHelp bool
	history     *history.Store
	ctx         *ProgramContext
	keys        AppKeyMap
	help        help.Model

	// Setup mode
	setupMode bool
	setup     views.SetupModel
	cfg       *config.Config

	// Sub-models for each view
	dashboard  views.DashboardModel
	backupList views.BackupListModel
	restore    views.RestoreModel
	settings   views.SettingsModel
	logs       views.LogsModel

	// Toast notification
	toast components.Toast
}

func NewModel() Model {
	cfg, err := config.Load()
	if err != nil {
		// Config doesn't exist, enter setup mode
		ctx := NewProgramContext(nil, nil)
		return Model{
			state:     SetupView,
			setupMode: true,
			setup:     views.NewSetup(ctx),
			ctx:       ctx,
			cfg:       nil,
			keys:      DefaultKeyMap(),
			help:      help.New(),
		}
	}

	store, err := history.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: history store unavailable: %v\n", err)
		store = nil
	}

	ctx := NewProgramContext(cfg, store)

	return Model{
		state:      DashboardView,
		setupMode:  false,
		cfg:        cfg,
		history:    store,
		ctx:        ctx,
		keys:       DefaultKeyMap(),
		help:       help.New(),
		dashboard:  views.NewDashboard(ctx),
		backupList: views.NewBackupList(ctx),
		restore:    views.NewRestore(ctx),
		settings:   views.NewSettings(ctx),
		logs:       views.NewLogs(ctx),
		toast:      components.NewToast(),
	}
}

// NewModelForTest creates a Model with pre-built config/store for testing.
// This bypasses config/history loading side effects.
func NewModelForTest(cfg *config.Config, store *history.Store) Model {
	if cfg == nil {
		ctx := NewProgramContext(nil, store)
		return Model{
			state:     SetupView,
			setupMode: true,
			setup:     views.NewSetup(ctx),
			ctx:       ctx,
			cfg:       nil,
			keys:      DefaultKeyMap(),
			help:      help.New(),
			toast:     components.NewToast(),
		}
	}

	ctx := NewProgramContext(cfg, store)

	return Model{
		state:      DashboardView,
		setupMode:  false,
		cfg:        cfg,
		history:    store,
		ctx:        ctx,
		keys:       DefaultKeyMap(),
		help:       help.New(),
		dashboard:  views.NewDashboard(ctx),
		backupList: views.NewBackupList(ctx),
		restore:    views.NewRestore(ctx),
		settings:   views.NewSettings(ctx),
		logs:       views.NewLogs(ctx),
		toast:      components.NewToast(),
	}
}

func (m Model) Init() tea.Cmd {
	if m.setupMode {
		return m.setup.Init()
	}

	return tea.Batch(
		m.dashboard.Init(),
		m.backupList.Init(),
		m.restore.Init(),
		m.settings.Init(),
		m.logs.Init(),
	)
}

func (m Model) GetConfig() *config.Config {
	if m.ctx != nil {
		return m.ctx.Config
	}
	return m.cfg
}

// activeTabIndex returns the index (0-4) of the current view in tabOrder.
// Returns 0 (Dashboard) as fallback for views not in tabOrder.
func (m Model) activeTabIndex() int {
	for i, v := range tabOrder {
		if v == m.state {
			return i
		}
	}
	return 0
}

// isInputActive returns true when the current view is consuming keyboard input
// (e.g., text fields, password prompts, edit mode).
func (m Model) isInputActive() bool {
	switch m.state {
	case SettingsView:
		return m.settings.IsEditing()
	case BackupListView:
		return m.backupList.IsCreating()
	case RestoreView:
		return m.restore.IsInputActive()
	}
	return false
}

// calculateTabFromX determines which tab was clicked based on X coordinate.
// Returns the tab index (0-4) or -1 if no tab was hit.
// Tab bar layout: "1 Dashboard │ 2 Backups │ 3 Restore │ 4 Settings │ 5 Logs"
// with 2-char left margin.
func (m Model) calculateTabFromX(x int) int {
	// Tab names and their widths (including key + space + label)
	tabNames := []string{"Dashboard", "Backups", "Restore", "Settings", "Logs"}

	// Start position after 2-char left margin
	pos := 2

	for i, name := range tabNames {
		// Tab width: "K " (key + space) + label + padding
		tabWidth := 2 + len(name)

		if x >= pos && x < pos+tabWidth {
			return i
		}

		// Move to next tab: current tab + separator " │ " (3 chars)
		pos += tabWidth + 3
	}

	return -1
}
