package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

// ViewState represents the current view
type ViewState int

const (
	DashboardView ViewState = iota
	FileBrowserView
	BackupListView
	RestoreView
	SettingsView
	LogsView
	SetupView
)

// Model represents the main TUI model
type Model struct {
	state       ViewState
	width       int
	height      int
	quitting    bool
	err         error
	showingHelp bool
	history     *history.Store

	// Setup mode
	setupMode bool
	setup     views.SetupModel
	cfg       *config.Config

	// Sub-models for each view
	dashboard   views.DashboardModel
	fileBrowser views.FileBrowserModel
	backupList  views.BackupListModel
	restore     views.RestoreModel
	settings    views.SettingsModel
	logs        views.LogsModel
}

func NewModel() Model {
	cfg, err := config.Load()
	if err != nil {
		// Config doesn't exist, enter setup mode
		return Model{
			state:     SetupView,
			setupMode: true,
			setup:     views.NewSetup(),
			cfg:       nil,
		}
	}

	store, err := history.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: history store unavailable: %v\n", err)
		store = nil
	}

	return Model{
		state:       DashboardView,
		setupMode:   false,
		cfg:         cfg,
		history:     store,
		dashboard:   views.NewDashboard(cfg),
		fileBrowser: views.NewFileBrowser(cfg),
		backupList:  views.NewBackupList(cfg, store),
		restore:     views.NewRestore(cfg, store),
		settings:    views.NewSettings(cfg),
		logs:        views.NewLogs(cfg, store),
	}
}

func (m Model) Init() tea.Cmd {
	if m.setupMode {
		return m.setup.Init()
	}

	return tea.Batch(
		m.dashboard.Init(),
		m.fileBrowser.Init(),
		m.backupList.Init(),
		m.restore.Init(),
		m.settings.Init(),
		m.logs.Init(),
	)
}

func (m Model) GetConfig() *config.Config {
	return m.cfg
}
