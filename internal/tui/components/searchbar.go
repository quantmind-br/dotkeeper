package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

// SearchFilterMsg is sent when the user submits a search query
type SearchFilterMsg struct {
	Query string
}

type SearchBar struct {
	input  textinput.Model
	active bool
	width  int
}

func NewSearchBar() SearchBar {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 100
	return SearchBar{input: ti}
}

func (s *SearchBar) Activate() tea.Cmd {
	s.active = true
	s.input.SetValue("")
	s.input.Focus()
	return textinput.Blink
}

func (s *SearchBar) Deactivate() {
	s.active = false
	s.input.SetValue("")
	s.input.Blur()
}

func (s SearchBar) IsActive() bool {
	return s.active
}

func (s SearchBar) Query() string {
	return s.input.Value()
}

func (s *SearchBar) SetWidth(w int) {
	s.width = w
	if w > 4 {
		s.input.Width = w - 4
	}
}

func (s SearchBar) Update(msg tea.Msg) (SearchBar, tea.Cmd) {
	if !s.active {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			s.Deactivate()
			return s, nil
		case tea.KeyEnter:
			return s, func() tea.Msg {
				return SearchFilterMsg{Query: s.input.Value()}
			}
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

func (s SearchBar) View() string {
	if !s.active {
		return ""
	}
	st := styles.DefaultStyles()
	prefix := st.Help.Render("/ ")
	return prefix + s.input.View()
}
