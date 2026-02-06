package views

import "github.com/charmbracelet/lipgloss"

// Styles holds common styles for the TUI
type Styles struct {
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	Normal       lipgloss.Style
	Selected     lipgloss.Style
	Help         lipgloss.Style
	Error        lipgloss.Style
	Success      lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Hint         lipgloss.Style
	TabActive    lipgloss.Style
	TabInactive  lipgloss.Style
	TabSeparator lipgloss.Style
}

// DefaultStyles returns the default styles
func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginLeft(2),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			MarginLeft(2),
		Normal: lipgloss.NewStyle().
			MarginLeft(2),
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Background(lipgloss.Color("#2A2A2A")).
			Bold(true),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			MarginLeft(2),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
		Label: lipgloss.NewStyle().
			Bold(true),
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")),
		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Underline(true),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),
		TabSeparator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")),
	}
}
