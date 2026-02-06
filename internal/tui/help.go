package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

func globalHelp() []views.HelpEntry {
	return []views.HelpEntry{
		{Key: "Tab", Description: "Next view"},
		{Key: "Shift+Tab", Description: "Previous view"},
		{Key: "1-5", Description: "Go to view"},
		{Key: "q", Description: "Quit"},
		{Key: "?", Description: "Toggle help"},
	}
}

func renderHelpOverlay(global []views.HelpEntry, viewHelp []views.HelpEntry, width, height int) string {
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAAA"))

	var content strings.Builder

	content.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	content.WriteString("\n\n")

	content.WriteString(sectionStyle.Render("Global"))
	content.WriteString("\n")
	for _, entry := range global {
		content.WriteString(fmt.Sprintf("  %s  %s\n", keyStyle.Render(entry.Key), entry.Description))
	}

	if len(viewHelp) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("Current View"))
		content.WriteString("\n")
		for _, entry := range viewHelp {
			content.WriteString(fmt.Sprintf("  %s  %s\n", keyStyle.Render(entry.Key), entry.Description))
		}
	}

	if width < 40 || height < 15 {
		return content.String()
	}

	overlayW := width - 4
	if overlayW > 60 {
		overlayW = 60
	}
	overlayH := height - 4
	if overlayH > 20 {
		overlayH = 20
	}

	overlayStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(overlayW).
		Height(overlayH)

	overlay := overlayStyle.Render(content.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, overlay)
}
