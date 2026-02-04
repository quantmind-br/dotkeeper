package tui

import (
	tea "github.com/charmbracelet/bubbletea"
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
)

// Model represents the main TUI model
type Model struct {
	state    ViewState
	width    int
	height   int
	quitting bool
	err      error

	// Sub-models for each view (will be implemented in later tasks)
	// dashboard    dashboard.Model
	// fileBrowser  filebrowser.Model
	// backupList   backuplist.Model
	// restore      restore.Model
	// settings     settings.Model
	// logs         logs.Model
}

// NewModel creates a new TUI model
func NewModel() Model {
	return Model{
		state: DashboardView,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}
