package views

import tea "github.com/charmbracelet/bubbletea"

// View is the standard interface for all TUI views.
// It extends tea.Model with help-related methods.
type View interface {
	tea.Model
	HelpBindings() []HelpEntry
	StatusHelpText() string
}

// Refreshable is implemented by views that can refresh their data.
type Refreshable interface {
	Refresh() tea.Cmd
}

// InputConsumer is implemented by views that consume keyboard input
// (e.g., text fields, password prompts, edit mode).
type InputConsumer interface {
	IsInputActive() bool
}
