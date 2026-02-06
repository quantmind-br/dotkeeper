package tui

import (
	"strings"

	"github.com/diogo/dotkeeper/internal/tui/components"
	"github.com/diogo/dotkeeper/internal/tui/styles"
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
		viewHelp := m.currentViewHelp()
		return renderHelpOverlay(globalHelp(), viewHelp, m.width, m.height)
	}

	var b strings.Builder

	s := styles.DefaultStyles()

	b.WriteString(s.AppTitle.Render("DotKeeper - Dotfiles Backup Manager"))
	b.WriteString("\n")

	tabBar := components.NewTabBar(s)
	b.WriteString(tabBar.View(m.activeTabIndex(), m.width))
	b.WriteString("\n\n")

	switch m.state {
	case DashboardView:
		b.WriteString(s.ContentArea.Render(m.dashboard.View()))
	case BackupListView:
		b.WriteString(s.ContentArea.Render(m.backupList.View()))
	case RestoreView:
		b.WriteString(s.ContentArea.Render(m.restore.View()))
	case SettingsView:
		b.WriteString(s.ContentArea.Render(m.settings.View()))
	case LogsView:
		b.WriteString(s.ContentArea.Render(m.logs.View()))
	default:
		// Fallback to dashboard for unreachable states
		b.WriteString(s.ContentArea.Render(m.dashboard.View()))
	}

	b.WriteString("\n\n")

	b.WriteString(s.GlobalHelp.Render("Tab/1-5: switch views | q: quit | ?: help"))
	b.WriteString("\n")

	return b.String()
}

// currentViewHelp returns help bindings for the currently active view.
func (m Model) currentViewHelp() []views.HelpEntry {
	switch m.state {
	case DashboardView:
		return m.dashboard.HelpBindings()
	case BackupListView:
		return m.backupList.HelpBindings()
	case RestoreView:
		return m.restore.HelpBindings()
	case SettingsView:
		return m.settings.HelpBindings()
	case LogsView:
		return m.logs.HelpBindings()
	}
	return nil
}
