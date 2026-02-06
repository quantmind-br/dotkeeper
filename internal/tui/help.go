package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/diogo/dotkeeper/internal/tui/styles"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

// HelpEntryToKeyBinding converts a views.HelpEntry to a key.Binding for use with bubbles/help.
// This adapter keeps HelpEntry as the view-facing API while enabling use of the bubbles/help component.
func HelpEntryToKeyBinding(entry views.HelpEntry) key.Binding {
	return key.NewBinding(
		key.WithKeys(entry.Key),
		key.WithHelp(entry.Key, entry.Description),
	)
}

// HelpEntriesToKeyBindings converts a slice of HelpEntry to a slice of key.Binding.
func HelpEntriesToKeyBindings(entries []views.HelpEntry) []key.Binding {
	bindings := make([]key.Binding, len(entries))
	for i, entry := range entries {
		bindings[i] = HelpEntryToKeyBinding(entry)
	}
	return bindings
}

// HelpKeyMap implements help.KeyMap interface for use with bubbles/help.Model.
// It wraps global and view-specific help bindings.
type HelpKeyMap struct {
	global []key.Binding
	view   []key.Binding
}

// ShortHelp returns a slice of bindings to be displayed in the short help bar.
func (h HelpKeyMap) ShortHelp() []key.Binding {
	return append(h.global, h.view...)
}

// FullHelp returns an extended group of help items, grouped by columns.
func (h HelpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{h.global, h.view}
}

// NewHelpKeyMap creates a HelpKeyMap from global and view HelpEntry slices.
func NewHelpKeyMap(globalEntries, viewEntries []views.HelpEntry) HelpKeyMap {
	return HelpKeyMap{
		global: HelpEntriesToKeyBindings(globalEntries),
		view:   HelpEntriesToKeyBindings(viewEntries),
	}
}

// globalHelp returns the global keyboard shortcuts for the TUI.
func globalHelp() []views.HelpEntry {
	return []views.HelpEntry{
		{Key: "Tab", Description: "Next view"},
		{Key: "Shift+Tab", Description: "Previous view"},
		{Key: "1-5", Description: "Go to view"},
		{Key: "q", Description: "Quit"},
		{Key: "?", Description: "Toggle help"},
	}
}

// renderHelpOverlay renders the help overlay using bubbles/help.Model.
// It combines global and view-specific help bindings into a styled overlay.
func renderHelpOverlay(helpModel help.Model, global []views.HelpEntry, viewHelp []views.HelpEntry, width, height int) string {
	s := styles.DefaultStyles()

	keyMap := NewHelpKeyMap(global, viewHelp)
	helpContent := helpModel.View(keyMap)

	var content strings.Builder
	content.WriteString(s.HelpTitle.Render("Keyboard Shortcuts"))
	content.WriteString("\n\n")
	content.WriteString(helpContent)

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

	overlayStyle := s.HelpOverlay.
		Width(overlayW).
		Height(overlayH)

	overlay := overlayStyle.Render(content.String())

	return views.PlaceOverlay(width, height, overlay)
}

// RenderHelpBar renders the inline help bar using bubbles/help.Model.
// This provides a compact help view suitable for the bottom status bar.
func RenderHelpBar(helpModel help.Model, globalEntries []views.HelpEntry) string {
	keyMap := NewHelpKeyMap(globalEntries, nil)
	return helpModel.View(keyMap)
}
