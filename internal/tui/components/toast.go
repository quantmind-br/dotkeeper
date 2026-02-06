package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToastLevel represents the severity level of a toast notification
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastError
)

// ToastDismissMsg is sent when a toast should be dismissed
type ToastDismissMsg struct{}

// Toast represents a temporary notification that auto-dismisses
type Toast struct {
	Message  string
	Level    ToastLevel
	Visible  bool
	duration time.Duration
}

// NewToast creates a new Toast with default 3-second duration
func NewToast() Toast {
	return Toast{
		duration: 3 * time.Second,
	}
}

// Show displays a toast message and returns a command to auto-dismiss it
func (t *Toast) Show(msg string, level ToastLevel) tea.Cmd {
	t.Message = msg
	t.Level = level
	t.Visible = true
	return t.dismissAfter()
}

// dismissAfter returns a command that sends ToastDismissMsg after the duration
func (t Toast) dismissAfter() tea.Cmd {
	return tea.Tick(t.duration, func(time.Time) tea.Msg {
		return ToastDismissMsg{}
	})
}

// Dismiss hides the toast
func (t *Toast) Dismiss() {
	t.Visible = false
	t.Message = ""
}

// View renders the toast message with appropriate styling
func (t Toast) View() string {
	if !t.Visible || t.Message == "" {
		return ""
	}

	var style lipgloss.Style
	switch t.Level {
	case ToastSuccess:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#00AA00", Dark: "#04B575"})
	case ToastError:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#CC0000", Dark: "#FF5555"})
	default: // ToastInfo
		style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#FFFFFF"})
	}

	return style.Render(t.Message)
}
