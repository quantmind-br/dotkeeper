package styles

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// AdaptiveColor palette for light/dark terminal support
var (
	AccentColor         = lipgloss.AdaptiveColor{Light: "#6C3EC2", Dark: "#7D56F4"}
	TextColor           = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#FFFFFF"}
	MutedColor          = lipgloss.AdaptiveColor{Light: "#999999", Dark: "#AAAAAA"}
	SecondaryMutedColor = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#666666"}
	ErrorColor          = lipgloss.AdaptiveColor{Light: "#CC0000", Dark: "#FF5555"}
	SuccessColor        = lipgloss.AdaptiveColor{Light: "#00AA00", Dark: "#04B575"}
	BgColor             = lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#2A2A2A"}
	BorderColor         = lipgloss.AdaptiveColor{Light: "#CCCCCC", Dark: "#444444"}
)

// Responsive breakpoint constants
const (
	BreakpointWide    = 80 // Full horizontal layout (cards, tabs)
	BreakpointMedium  = 60 // Medium layout (action buttons horizontal)
	MinTerminalWidth  = 40 // Minimum supported terminal width
	MinTerminalHeight = 15 // Minimum supported terminal height
)

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
		Foreground(AccentColor).
		MarginLeft(2),
	ContentArea: lipgloss.NewStyle().MarginLeft(2),
	GlobalHelp: lipgloss.NewStyle().
		Foreground(SecondaryMutedColor).
		MarginLeft(2),
	HelpKey:     lipgloss.NewStyle().Bold(true).Foreground(AccentColor),
	HelpTitle:   lipgloss.NewStyle().Bold(true).Foreground(AccentColor),
	HelpSection: lipgloss.NewStyle().Bold(true).Foreground(MutedColor),
	HelpOverlay: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(1, 2),
	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		MarginLeft(2),
	Subtitle: lipgloss.NewStyle().
		Foreground(AccentColor).
		MarginLeft(2),
	Normal: lipgloss.NewStyle().
		MarginLeft(2),
	Selected: lipgloss.NewStyle().
		Foreground(AccentColor).
		Background(BgColor).
		Bold(true),
	Help: lipgloss.NewStyle().
		Foreground(SecondaryMutedColor).
		MarginLeft(2),
	Error: lipgloss.NewStyle().
		Foreground(ErrorColor),
	Success: lipgloss.NewStyle().
		Foreground(SuccessColor),
	Label: lipgloss.NewStyle().
		Bold(true),
	Value: lipgloss.NewStyle().
		Foreground(MutedColor),
	Hint: lipgloss.NewStyle().
		Foreground(SecondaryMutedColor).
		Italic(true),
	TabActive: lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		Underline(true),
	TabInactive: lipgloss.NewStyle().
		Foreground(SecondaryMutedColor),
	TabSeparator: lipgloss.NewStyle().
		Foreground(BorderColor),
	ViewContainer: lipgloss.NewStyle().MarginLeft(2),
	StatusBar:     lipgloss.NewStyle().MarginTop(1),
	Card: lipgloss.NewStyle().
		Background(BgColor).
		Padding(1, 2).
		MarginRight(2),
	CardTitle: lipgloss.NewStyle().
		Bold(true).
		Foreground(TextColor),
	CardLabel: lipgloss.NewStyle().
		Foreground(MutedColor),
	ActionButton: lipgloss.NewStyle().
		Background(BgColor).
		Padding(0, 2).
		MarginRight(1),
	ActionButtonKey: lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true),
	ButtonSelected: lipgloss.NewStyle().
		Bold(true).
		Foreground(TextColor).
		Background(AccentColor).
		Padding(0, 2).
		MarginRight(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor),
	ButtonNormal: lipgloss.NewStyle().
		Foreground(MutedColor).
		Background(BgColor).
		Padding(0, 2).
		MarginRight(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor),
	ViewportBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor),
}

// DefaultStyles returns the default styles
func DefaultStyles() Styles {
	return defaultStyles
}

// NewListDelegate returns a new list delegate with common styling
func NewListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle = lipgloss.NewStyle().Foreground(TextColor).Padding(0, 0, 0, 2)
	d.Styles.NormalDesc = lipgloss.NewStyle().Foreground(MutedColor).Padding(0, 0, 0, 2)
	d.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(AccentColor).Bold(true).Padding(0, 0, 0, 2)
	d.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(AccentColor).Padding(0, 0, 0, 2)
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(AccentColor)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(AccentColor)
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
