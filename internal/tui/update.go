package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

func (m *Model) propagateWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	viewWidth := msg.Width
	if viewWidth < 0 {
		viewWidth = 0
	}
	viewHeight := msg.Height - mainChromeHeight
	if viewHeight < 0 {
		viewHeight = 0
	}
	viewMsg := tea.WindowSizeMsg{
		Width:  viewWidth,
		Height: viewHeight,
	}

	var tm tea.Model
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Type assertion required after Update() - all views implement views.View interface
	tm, cmd = m.dashboard.Update(viewMsg)
	m.dashboard = tm.(views.DashboardModel)
	cmds = append(cmds, cmd)

	tm, cmd = m.backupList.Update(viewMsg)
	m.backupList = tm.(views.BackupListModel)
	cmds = append(cmds, cmd)

	tm, cmd = m.restore.Update(viewMsg)
	m.restore = tm.(views.RestoreModel)
	cmds = append(cmds, cmd)

	tm, cmd = m.settings.Update(viewMsg)
	m.settings = tm.(views.SettingsModel)
	cmds = append(cmds, cmd)

	tm, cmd = m.logs.Update(viewMsg)
	m.logs = tm.(views.LogsModel)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// refreshCmdForState returns the refresh command for a given view state.
func (m *Model) refreshCmdForState(state ViewState) tea.Cmd {
	switch state {
	case DashboardView:
		return m.dashboard.Refresh()
	case BackupListView:
		return m.backupList.Refresh()
	case RestoreView:
		return m.restore.Refresh()
	case LogsView:
		return m.logs.LoadHistory()
	default:
		return nil
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	if m.setupMode {
		if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = wsMsg.Width
			m.height = wsMsg.Height
		}
		switch msg := msg.(type) {
		case views.SetupCompleteMsg:
			m.setupMode = false
			cfg, _ := config.Load()
			m.cfg = cfg
			store, _ := history.NewStore()
			m.history = store
			m.ctx = NewProgramContext(cfg, store)
			m.dashboard = views.NewDashboard(m.ctx)
			m.backupList = views.NewBackupList(m.ctx)
			m.restore = views.NewRestore(m.ctx)
			m.settings = views.NewSettings(m.ctx)
			m.logs = views.NewLogs(m.ctx)
			m.state = DashboardView
			if m.width > 0 && m.height > 0 {
				return m, m.propagateWindowSize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			}
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
		return m, m.propagateWindowSize(msg)

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

		if key.Matches(msg, m.keys.Help) {
			m.showingHelp = !m.showingHelp
			return m, nil
		}

		if key.Matches(msg, m.keys.Quit) {
			m.quitting = true
			return m, tea.Quit
		}

		if key.Matches(msg, m.keys.Tab) {
			if !m.isInputActive() {
				currentIdx := m.activeTabIndex()
				nextIdx := (currentIdx + 1) % len(tabOrder)
				prevState := m.state
				m.state = tabOrder[nextIdx]
				if m.state != prevState {
					if cmd := m.refreshCmdForState(m.state); cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
			}
			return m, tea.Batch(cmds...)
		}

		if key.Matches(msg, m.keys.ShiftTab) {
			if !m.isInputActive() {
				currentIdx := m.activeTabIndex()
				prevIdx := (currentIdx - 1 + len(tabOrder)) % len(tabOrder)
				prevState := m.state
				m.state = tabOrder[prevIdx]
				if m.state != prevState {
					if cmd := m.refreshCmdForState(m.state); cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
			}
			return m, tea.Batch(cmds...)
		}

		// Number key navigation (only when not in input-consuming state)
		if !m.isInputActive() {
			var targetIdx int = -1
			switch msg.String() {
			case "1":
				targetIdx = 0
			case "2":
				targetIdx = 1
			case "3":
				targetIdx = 2
			case "4":
				targetIdx = 3
			case "5":
				targetIdx = 4
			}
			if targetIdx >= 0 && targetIdx < len(tabOrder) {
				m.state = tabOrder[targetIdx]
				if cmd := m.refreshCmdForState(m.state); cmd != nil {
					cmds = append(cmds, cmd)
				}
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
