package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginLeft(2)
	b.WriteString(titleStyle.Render("dotkeeper - Dotfiles Backup Manager"))
	b.WriteString("\n\n")

	// Current view indicator
	viewNames := []string{"Dashboard", "File Browser", "Backup List", "Restore", "Settings", "Logs"}
	viewStyle := lipgloss.NewStyle().MarginLeft(2)
	b.WriteString(viewStyle.Render(fmt.Sprintf("Current View: %s", viewNames[m.state])))
	b.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		MarginLeft(2)
	b.WriteString(helpStyle.Render("Press Tab to switch views, q to quit, ? for help"))
	b.WriteString("\n")

	return b.String()
}
