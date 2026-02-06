package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

type KeyMap struct {
	Quit     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Help     key.Binding
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
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous view"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

var keys = DefaultKeyMap()

func (m *Model) propagateWindowSize(msg tea.WindowSizeMsg) {
	viewHeight := msg.Height - mainChromeHeight
	if viewHeight < 0 {
		viewHeight = 0
	}
	viewMsg := tea.WindowSizeMsg{
		Width:  msg.Width,
		Height: viewHeight,
	}

	var tm tea.Model

	tm, _ = m.dashboard.Update(viewMsg)
	m.dashboard = tm.(views.DashboardModel)

	tm, _ = m.backupList.Update(viewMsg)
	m.backupList = tm.(views.BackupListModel)

	tm, _ = m.restore.Update(viewMsg)
	m.restore = tm.(views.RestoreModel)

	tm, _ = m.settings.Update(viewMsg)
	m.settings = tm.(views.SettingsModel)

	tm, _ = m.logs.Update(viewMsg)
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
			store, _ := history.NewStore()
			m.history = store
			m.dashboard = views.NewDashboard(cfg)
			m.backupList = views.NewBackupList(cfg, store)
			m.restore = views.NewRestore(cfg, store)
			m.settings = views.NewSettings(cfg)
			m.logs = views.NewLogs(cfg, store)
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

	case views.DashboardNavigateMsg:
		switch msg.Target {
		case "backups":
			m.state = BackupListView
			cmds = append(cmds, m.backupList.Refresh())
		case "restore":
			m.state = RestoreView
			cmds = append(cmds, m.restore.Refresh())
		case "settings":
			m.state = SettingsView
		}
		return m, tea.Batch(cmds...)

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
			if !m.isInputActive() {
				currentIdx := m.activeTabIndex()
				nextIdx := (currentIdx + 1) % len(tabOrder)
				prevState := m.state
				m.state = tabOrder[nextIdx]
				if m.state == BackupListView && prevState != BackupListView {
					cmds = append(cmds, m.backupList.Refresh())
				}
				if m.state == RestoreView && prevState != RestoreView {
					cmds = append(cmds, m.restore.Refresh())
				}
				if m.state == LogsView && prevState != LogsView {
					cmds = append(cmds, m.logs.LoadHistory())
				}
			}
			return m, tea.Batch(cmds...)
		}

		if key.Matches(msg, keys.ShiftTab) {
			if !m.isInputActive() {
				currentIdx := m.activeTabIndex()
				prevIdx := (currentIdx - 1 + len(tabOrder)) % len(tabOrder)
				prevState := m.state
				m.state = tabOrder[prevIdx]
				if m.state == BackupListView && prevState != BackupListView {
					cmds = append(cmds, m.backupList.Refresh())
				}
				if m.state == RestoreView && prevState != RestoreView {
					cmds = append(cmds, m.restore.Refresh())
				}
				if m.state == LogsView && prevState != LogsView {
					cmds = append(cmds, m.logs.LoadHistory())
				}
			}
			return m, tea.Batch(cmds...)
		}

		// Number key navigation (only when not in input-consuming state)
		if !m.isInputActive() {
			switch msg.String() {
			case "1":
				m.state = tabOrder[0]
				return m, nil
			case "2":
				m.state = tabOrder[1]
				cmds = append(cmds, m.backupList.Refresh())
				return m, tea.Batch(cmds...)
			case "3":
				m.state = tabOrder[2]
				cmds = append(cmds, m.restore.Refresh())
				return m, tea.Batch(cmds...)
			case "4":
				m.state = tabOrder[3]
				return m, nil
			case "5":
				m.state = tabOrder[4]
				cmds = append(cmds, m.logs.LoadHistory())
				return m, tea.Batch(cmds...)
			}
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
