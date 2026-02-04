package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds common styles for the TUI
type Styles struct {
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Normal   lipgloss.Style
	Selected lipgloss.Style
	Help     lipgloss.Style
	Error    lipgloss.Style
	Success  lipgloss.Style
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
			Foreground(lipgloss.Color("#FF0000")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")),
	}
}
