package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

// SearchFilterMsg is sent when the user submits a search query
type SearchFilterMsg struct {
	Query string
}

// SearchBar is a search/filter overlay component with text input
type SearchBar struct {
	input  textinput.Model
	active bool
	width  int
}

// NewSearchBar creates a new SearchBar component
func NewSearchBar() SearchBar {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 100
	return SearchBar{
		input:  ti,
		active: false,
		width:  0,
	}
}

// Activate enables the search bar and focuses the input
func (s *SearchBar) Activate() tea.Cmd {
	s.active = true
	s.input.SetValue("")
	s.input.Focus()
	return textinput.Blink
}

// Deactivate disables the search bar and clears the input
func (s *SearchBar) Deactivate() {
	s.active = false
	s.input.SetValue("")
	s.input.Blur()
}

// IsActive returns whether the search bar is currently active
func (s SearchBar) IsActive() bool {
	return s.active
}

// Query returns the current search query
func (s SearchBar) Query() string {
	return s.input.Value()
}

// SetWidth sets the width of the search bar
func (s *SearchBar) SetWidth(w int) {
	s.width = w
	// Account for "/ " prefix (2 chars) and padding
	if w > 4 {
		s.input.Width = w - 4
	}
}

// Update handles messages for the search bar
func (s SearchBar) Update(msg tea.Msg) (SearchBar, tea.Cmd) {
	if !s.active {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.Deactivate()
			return s, nil
		case "enter":
			return s, func() tea.Msg {
				return SearchFilterMsg{Query: s.input.Value()}
			}
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

// View renders the search bar
func (s SearchBar) View() string {
	if !s.active {
		return ""
	}
	st := styles.DefaultStyles()
	prefix := st.Help.Render("/ ")

	// Apply border style
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	if s.width > 0 {
		style = style.Width(s.width - 4)
	}

	return style.Render(prefix + s.input.View())
}
