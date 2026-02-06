package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// NewPasswordInput creates a pre-configured password text input.
func NewPasswordInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Width = 40
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	return ti
}
