package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.setupMode {
		return m.setup.View()
	}

	if m.quitting {
		return "Goodbye!\n"
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
