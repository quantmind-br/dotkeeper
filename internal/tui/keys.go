package tui

import "github.com/charmbracelet/bubbles/key"

// AppKeyMap holds all global key bindings for the application
type AppKeyMap struct {
	Quit     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Help     key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() AppKeyMap {
	return AppKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next view"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous view"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// ShortHelp returns the short help bindings for the help component.
// Implements the key.Map interface.
func (k AppKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns the full help bindings for the help component.
// Implements the key.Map interface.
func (k AppKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.ShiftTab},
		{k.Help, k.Quit},
	}
}
