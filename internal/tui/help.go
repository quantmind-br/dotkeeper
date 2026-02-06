package tui

import (
	"fmt"
	"strings"

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
	styles := views.DefaultStyles()

	var content strings.Builder

	content.WriteString(styles.HelpTitle.Render("Keyboard Shortcuts"))
	content.WriteString("\n\n")

	content.WriteString(styles.HelpSection.Render("Global"))
	content.WriteString("\n")
	for _, entry := range global {
		content.WriteString(fmt.Sprintf("  %s  %s\n", styles.HelpKey.Render(entry.Key), entry.Description))
	}

	if len(viewHelp) > 0 {
		content.WriteString("\n")
		content.WriteString(styles.HelpSection.Render("Current View"))
		content.WriteString("\n")
		for _, entry := range viewHelp {
			content.WriteString(fmt.Sprintf("  %s  %s\n", styles.HelpKey.Render(entry.Key), entry.Description))
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

	overlayStyle := styles.HelpOverlay.
		Width(overlayW).
		Height(overlayH)

	overlay := overlayStyle.Render(content.String())

	return views.PlaceOverlay(width, height, overlay)
}
