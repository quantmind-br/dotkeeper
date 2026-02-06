package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchBar_NewIsInactive(t *testing.T) {
	sb := NewSearchBar()
	if sb.IsActive() {
		t.Error("expected new SearchBar to be inactive")
	}
	if sb.Query() != "" {
		t.Errorf("expected empty query, got %q", sb.Query())
	}
}

func TestSearchBar_Activate(t *testing.T) {
	sb := NewSearchBar()
	cmd := sb.Activate()
	if !sb.IsActive() {
		t.Error("expected SearchBar to be active after Activate()")
	}
	if cmd == nil {
		t.Error("expected Activate() to return a command")
	}
}

func TestSearchBar_Deactivate(t *testing.T) {
	sb := NewSearchBar()
	sb.Activate()
	sb.Deactivate()
	if sb.IsActive() {
		t.Error("expected SearchBar to be inactive after Deactivate()")
	}
	if sb.Query() != "" {
		t.Errorf("expected empty query after deactivate, got %q", sb.Query())
	}
}

func TestSearchBar_Query(t *testing.T) {
	sb := NewSearchBar()
	sb.Activate()
	sb.input.SetValue("test query")
	if sb.Query() != "test query" {
		t.Errorf("expected query 'test query', got %q", sb.Query())
	}
}

func TestSearchBar_EscDismisses(t *testing.T) {
	sb := NewSearchBar()
	sb.Activate()
	sb.input.SetValue("test")

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	sb, cmd := sb.Update(msg)

	if sb.IsActive() {
		t.Error("expected SearchBar to be inactive after Esc")
	}
	if sb.Query() != "" {
		t.Errorf("expected empty query after Esc, got %q", sb.Query())
	}
	if cmd != nil {
		t.Error("expected no command after Esc")
	}
}

func TestSearchBar_ViewWhenInactive(t *testing.T) {
	sb := NewSearchBar()
	view := sb.View()
	if view != "" {
		t.Errorf("expected empty view when inactive, got %q", view)
	}
}

func TestSearchBar_ViewWhenActive(t *testing.T) {
	sb := NewSearchBar()
	sb.Activate()
	sb.input.SetValue("test")
	sb.SetWidth(80)

	view := sb.View()
	if view == "" {
		t.Error("expected non-empty view when active")
	}
	if !strings.Contains(view, "/") {
		t.Error("expected view to contain '/' prefix")
	}
	if !strings.Contains(view, "test") {
		t.Error("expected view to contain search text")
	}
}

func TestSearchBar_SetWidth(t *testing.T) {
	sb := NewSearchBar()
	sb.SetWidth(80)
	if sb.width != 80 {
		t.Errorf("expected width 80, got %d", sb.width)
	}
	if sb.input.Width <= 0 {
		t.Error("expected input width to be set")
	}
}

func TestSearchBar_EnterSendsMessage(t *testing.T) {
	sb := NewSearchBar()
	sb.Activate()
	sb.input.SetValue("search term")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	sb, cmd := sb.Update(msg)

	if cmd == nil {
		t.Error("expected command after Enter")
	}

	resultMsg := cmd()
	filterMsg, ok := resultMsg.(SearchFilterMsg)
	if !ok {
		t.Errorf("expected SearchFilterMsg, got %T", resultMsg)
	}
	if filterMsg.Query != "search term" {
		t.Errorf("expected query 'search term', got %q", filterMsg.Query)
	}
}

func TestSearchBar_UpdateWhenInactive(t *testing.T) {
	sb := NewSearchBar()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	sb, cmd := sb.Update(msg)
	if cmd != nil {
		t.Error("expected nil command when inactive")
	}
}
