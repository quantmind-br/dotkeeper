package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

func (m Model) View() string {
	if m.setupMode {
		return m.setup.View()
	}

	if m.quitting {
		return "Goodbye!\n"
	}

	if m.showingHelp {
		var viewHelp []views.HelpEntry
		switch m.state {
		case DashboardView:
			if hp, ok := interface{}(m.dashboard).(views.HelpProvider); ok {
				viewHelp = hp.HelpBindings()
			}
		case FileBrowserView:
			if hp, ok := interface{}(m.fileBrowser).(views.HelpProvider); ok {
				viewHelp = hp.HelpBindings()
			}
		case BackupListView:
			if hp, ok := interface{}(m.backupList).(views.HelpProvider); ok {
				viewHelp = hp.HelpBindings()
			}
		case RestoreView:
			if hp, ok := interface{}(m.restore).(views.HelpProvider); ok {
				viewHelp = hp.HelpBindings()
			}
		case SettingsView:
			if hp, ok := interface{}(m.settings).(views.HelpProvider); ok {
				viewHelp = hp.HelpBindings()
			}
		case LogsView:
			if hp, ok := interface{}(m.logs).(views.HelpProvider); ok {
				viewHelp = hp.HelpBindings()
			}
		}
		return renderHelpOverlay(globalHelp(), viewHelp, m.width, m.height)
	}

	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginLeft(2)
	b.WriteString(titleStyle.Render("dotkeeper - Dotfiles Backup Manager"))
	b.WriteString("\n\n")

	contentStyle := lipgloss.NewStyle().MarginLeft(2)

	switch m.state {
	case DashboardView:
		b.WriteString(contentStyle.Render(m.dashboard.View()))
	case FileBrowserView:
		b.WriteString(contentStyle.Render(m.fileBrowser.View()))
	case BackupListView:
		b.WriteString(contentStyle.Render(m.backupList.View()))
	case RestoreView:
		b.WriteString(contentStyle.Render(m.restore.View()))
	case SettingsView:
		b.WriteString(contentStyle.Render(m.settings.View()))
	case LogsView:
		b.WriteString(contentStyle.Render(m.logs.View()))
	}

	b.WriteString("\n\n")

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		MarginLeft(2)
	b.WriteString(helpStyle.Render("Tab: switch views | q: quit | ?: help"))
	b.WriteString("\n")

	return b.String()
}
