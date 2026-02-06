package tui

import (
	"strings"

	"github.com/diogo/dotkeeper/internal/tui/components"
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

	styles := views.DefaultStyles()

	b.WriteString(styles.AppTitle.Render("dotkeeper - Dotfiles Backup Manager"))
	b.WriteString("\n")

	tabBar := components.NewTabBar(styles)
	b.WriteString(tabBar.View(m.activeTabIndex(), m.width))
	b.WriteString("\n\n")

	switch m.state {
	case DashboardView:
		b.WriteString(styles.ContentArea.Render(m.dashboard.View()))
	case BackupListView:
		b.WriteString(styles.ContentArea.Render(m.backupList.View()))
	case RestoreView:
		b.WriteString(styles.ContentArea.Render(m.restore.View()))
	case SettingsView:
		b.WriteString(styles.ContentArea.Render(m.settings.View()))
	case LogsView:
		b.WriteString(styles.ContentArea.Render(m.logs.View()))
	default:
		// Fallback to dashboard for unreachable states
		b.WriteString(styles.ContentArea.Render(m.dashboard.View()))
	}

	b.WriteString("\n\n")

	b.WriteString(styles.GlobalHelp.Render("Tab/1-5: switch views | q: quit | ?: help"))
	b.WriteString("\n")

	return b.String()
}
