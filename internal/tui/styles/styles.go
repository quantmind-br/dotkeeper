package styles

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// ViewChromeHeight accounts for title, tabbar, help, and margins
const ViewChromeHeight = 6

// Styles holds common styles for the TUI
type Styles struct {
	Title           lipgloss.Style
	Subtitle        lipgloss.Style
	Normal          lipgloss.Style
	Selected        lipgloss.Style
	Help            lipgloss.Style
	Error           lipgloss.Style
	Success         lipgloss.Style
	Label           lipgloss.Style
	Value           lipgloss.Style
	Hint            lipgloss.Style
	TabActive       lipgloss.Style
	TabInactive     lipgloss.Style
	TabSeparator    lipgloss.Style
	ViewContainer   lipgloss.Style
	StatusBar       lipgloss.Style
	Card            lipgloss.Style
	CardTitle       lipgloss.Style
	CardLabel       lipgloss.Style
	ActionButton    lipgloss.Style
	ActionButtonKey lipgloss.Style
	ButtonSelected  lipgloss.Style
	ButtonNormal    lipgloss.Style
	ViewportBorder  lipgloss.Style
	AppTitle        lipgloss.Style
	ContentArea     lipgloss.Style
	GlobalHelp      lipgloss.Style
	HelpKey         lipgloss.Style
	HelpTitle       lipgloss.Style
	HelpSection     lipgloss.Style
	HelpOverlay     lipgloss.Style
}

var defaultStyles = Styles{
	AppTitle: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginLeft(2),
	ContentArea: lipgloss.NewStyle().MarginLeft(2),
	GlobalHelp: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		MarginLeft(2),
	HelpKey:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")),
	HelpTitle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")),
	HelpSection: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAAA")),
	HelpOverlay: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2),
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
	ViewContainer: lipgloss.NewStyle().MarginLeft(2),
	StatusBar:     lipgloss.NewStyle().MarginTop(1),
	Card: lipgloss.NewStyle().
		Background(lipgloss.Color("#2A2A2A")).
		Padding(1, 2).
		MarginRight(2),
	CardTitle: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")),
	CardLabel: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")),
	ActionButton: lipgloss.NewStyle().
		Background(lipgloss.Color("#2A2A2A")).
		Padding(0, 2).
		MarginRight(1),
	ActionButtonKey: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true),
	ButtonSelected: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 2).
		MarginRight(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")),
	ButtonNormal: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Background(lipgloss.Color("#2A2A2A")).
		Padding(0, 2).
		MarginRight(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")),
	ViewportBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")),
}

// DefaultStyles returns the default styles
func DefaultStyles() Styles {
	return defaultStyles
}

// NewListDelegate returns a new list delegate with common styling
func NewListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 0, 0, 2)
	d.Styles.NormalDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Padding(0, 0, 0, 2)
	d.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true).Padding(0, 0, 0, 2)
	d.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Padding(0, 0, 0, 2)
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("#7D56F4"))
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("#7D56F4"))
	return d
}

// NewMinimalList creates a list.Model with common defaults (no title, no help, no filtering).
func NewMinimalList() list.Model {
	l := list.New([]list.Item{}, NewListDelegate(), 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	return l
}
