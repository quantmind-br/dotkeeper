package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

const viewCount = 6

type KeyMap struct {
	Quit key.Binding
	Tab  key.Binding
	Help key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next view"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

var keys = DefaultKeyMap()

func (m *Model) propagateWindowSize(msg tea.WindowSizeMsg) {
	var tm tea.Model

	tm, _ = m.dashboard.Update(msg)
	m.dashboard = tm.(views.DashboardModel)

	tm, _ = m.fileBrowser.Update(msg)
	m.fileBrowser = tm.(views.FileBrowserModel)

	tm, _ = m.backupList.Update(msg)
	m.backupList = tm.(views.BackupListModel)

	tm, _ = m.restore.Update(msg)
	m.restore = tm.(views.RestoreModel)

	tm, _ = m.settings.Update(msg)
	m.settings = tm.(views.SettingsModel)

	tm, _ = m.logs.Update(msg)
	m.logs = tm.(views.LogsModel)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.setupMode {
		switch msg := msg.(type) {
		case views.SetupCompleteMsg:
			m.setupMode = false
			cfg, _ := config.Load()
			m.cfg = cfg
			m.dashboard = views.NewDashboard(cfg)
			m.fileBrowser = views.NewFileBrowser(cfg)
			m.backupList = views.NewBackupList(cfg)
			m.restore = views.NewRestore(cfg)
			m.settings = views.NewSettings(cfg)
			m.logs = views.NewLogs(cfg)
			m.state = DashboardView
			return m, nil
		default:
			var model tea.Model
			model, cmd = m.setup.Update(msg)
			m.setup = model.(views.SetupModel)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.propagateWindowSize(msg)
		return m, nil

	case views.RefreshBackupListMsg:
		cmd = m.backupList.Refresh()
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.showingHelp {
			m.showingHelp = false
			return m, nil
		}

		if key.Matches(msg, keys.Help) {
			m.showingHelp = !m.showingHelp
			return m, nil
		}

		if key.Matches(msg, keys.Quit) {
			m.quitting = true
			return m, tea.Quit
		}

		if key.Matches(msg, keys.Tab) {
			prevState := m.state
			m.state = (m.state + 1) % viewCount
			if m.state == BackupListView && prevState != BackupListView {
				cmds = append(cmds, m.backupList.Refresh())
			}
			if m.state == RestoreView && prevState != RestoreView {
				cmds = append(cmds, m.restore.Refresh())
			}
			return m, tea.Batch(cmds...)
		}

		if m.state == DashboardView {
			switch msg.String() {
			case "b":
				m.state = BackupListView
				cmds = append(cmds, m.backupList.Refresh())
				return m, tea.Batch(cmds...)
			case "r":
				m.state = RestoreView
				return m, m.restore.Refresh()
			case "s":
				m.state = SettingsView
				return m, nil
			}
		}
	}

	switch m.state {
	case DashboardView:
		var model tea.Model
		model, cmd = m.dashboard.Update(msg)
		m.dashboard = model.(views.DashboardModel)
		cmds = append(cmds, cmd)
	case FileBrowserView:
		var model tea.Model
		model, cmd = m.fileBrowser.Update(msg)
		m.fileBrowser = model.(views.FileBrowserModel)
		cmds = append(cmds, cmd)
	case BackupListView:
		var model tea.Model
		model, cmd = m.backupList.Update(msg)
		m.backupList = model.(views.BackupListModel)
		cmds = append(cmds, cmd)
	case RestoreView:
		var model tea.Model
		model, cmd = m.restore.Update(msg)
		m.restore = model.(views.RestoreModel)
		cmds = append(cmds, cmd)
	case SettingsView:
		var model tea.Model
		model, cmd = m.settings.Update(msg)
		m.settings = model.(views.SettingsModel)
		cmds = append(cmds, cmd)
	case LogsView:
		var model tea.Model
		model, cmd = m.logs.Update(msg)
		m.logs = model.(views.LogsModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
