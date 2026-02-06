package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

	if m.width > 0 && m.height > 0 &&
		(m.width < styles.MinTerminalWidth || m.height < styles.MinTerminalHeight) {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			"Terminal too small\n"+
				fmt.Sprintf("Minimum: %dx%d", styles.MinTerminalWidth, styles.MinTerminalHeight)+"\n"+
				fmt.Sprintf("Current: %dx%d", m.width, m.height),
		)
	}

	if m.showingHelp {
		viewHelp := m.currentViewHelp()
		return renderHelpOverlay(m.help, globalHelp(), viewHelp, m.width, m.height, m.ctx.Styles)
	}

	var b strings.Builder

	s := m.ctx.Styles

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

	// Render toast notification if visible
	if toastView := m.toast.View(); toastView != "" {
		b.WriteString(toastView)
		b.WriteString("\n")
	}

	// Use bubbles/help for the inline help bar
	viewHelp := m.currentViewHelpText()
	if viewHelp != "" {
		b.WriteString(s.GlobalHelp.Render(viewHelp))
		b.WriteString("\n")
	}

	// Build global help bar using bubbles/help adapter
	helpBar := RenderHelpBar(m.help, globalHelp())
	b.WriteString(s.GlobalHelp.Render(helpBar))
	b.WriteString("\n")

	return b.String()
}

// currentViewHelpText returns the inline help text for the active view's status bar.
func (m Model) currentViewHelpText() string {
	var view interface {
		StatusHelpText() string
	}
	switch m.state {
	case DashboardView:
		view = m.dashboard
	case BackupListView:
		view = m.backupList
	case RestoreView:
		view = m.restore
	case SettingsView:
		view = m.settings
	case LogsView:
		view = m.logs
	default:
		return ""
	}
	return view.StatusHelpText()
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
