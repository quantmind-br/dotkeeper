package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchBar_NewNotActive(t *testing.T) {
	sb := NewSearchBar()
	if sb.IsActive() {
		t.Error("NewSearchBar should create inactive search bar")
	}
	if sb.Value() != "" {
		t.Error("NewSearchBar should have empty value")
	}
}

func TestSearchBar_Activate(t *testing.T) {
	sb := NewSearchBar()
	cmd := sb.Activate()

	if !sb.IsActive() {
		t.Error("Activate should set active to true")
	}
	if sb.Value() != "" {
		t.Error("Activate should clear value")
	}
	if cmd == nil {
		t.Error("Activate should return a command")
	}
}

func TestSearchBar_Deactivate(t *testing.T) {
	sb := NewSearchBar()
	sb.Activate()
	sb.input.SetValue("test search")

	sb.Deactivate()

	if sb.IsActive() {
		t.Error("Deactivate should set active to false")
	}
	if sb.Value() != "" {
		t.Error("Deactivate should clear value")
	}
}

func TestSearchBar_Value(t *testing.T) {
	sb := NewSearchBar()
	sb.Activate()
	sb.input.SetValue("hello world")

	if sb.Value() != "hello world" {
		t.Errorf("Value() should return 'hello world', got %q", sb.Value())
	}
}

func TestSearchBar_SetWidth(t *testing.T) {
	sb := NewSearchBar()
	sb.SetWidth(80)

	if sb.width != 80 {
		t.Errorf("SetWidth should set width to 80, got %d", sb.width)
	}
	if sb.input.Width != 76 {
		t.Errorf("SetWidth should set input.Width to 76, got %d", sb.input.Width)
	}
}

func TestSearchBar_SetWidthSmall(t *testing.T) {
	sb := NewSearchBar()
	sb.SetWidth(2)

	if sb.width != 2 {
		t.Errorf("SetWidth should set width to 2, got %d", sb.width)
	}
	if sb.input.Width < 0 {
		t.Errorf("SetWidth should not set negative input.Width, got %d", sb.input.Width)
	}
}

func TestSearchBar_ViewWhenInactive(t *testing.T) {
	sb := NewSearchBar()
	view := sb.View()

	if view != "" {
		t.Errorf("View() when inactive should return empty string, got %q", view)
	}
}

func TestSearchBar_ViewWhenActive(t *testing.T) {
	sb := NewSearchBar()
	sb.Activate()
	sb.input.SetValue("test")
	sb.SetWidth(80)

	view := sb.View()

	if view == "" {
		t.Error("View() when active should not return empty string")
	}
	if !strings.Contains(view, "/") {
		t.Error("View() should contain '/' prefix")
	}
	if !strings.Contains(view, "test") {
		t.Error("View() should contain search text 'test'")
	}
}

func TestSearchBar_UpdateWhenInactive(t *testing.T) {
	sb := NewSearchBar()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}

	sb2, cmd := sb.Update(msg)

	if sb2.Value() != "" {
		t.Error("Update when inactive should not change value")
	}
	if cmd != nil {
		t.Error("Update when inactive should return nil command")
	}
}

func TestSearchBar_UpdateWhenActive(t *testing.T) {
	sb := NewSearchBar()
	sb.Activate()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	sb2, cmd := sb.Update(msg)

	if cmd == nil {
		t.Error("Update when active should return a command")
	}
	// Note: We can't easily test the actual text input update without
	// running the full Bubble Tea program, but we verify the command is returned
}

func TestSearchBar_MultipleActivateDeactivate(t *testing.T) {
	sb := NewSearchBar()

	// Activate
	sb.Activate()
	if !sb.IsActive() {
		t.Error("First activate should set active to true")
	}

	// Deactivate
	sb.Deactivate()
	if sb.IsActive() {
		t.Error("First deactivate should set active to false")
	}

	// Activate again
	sb.Activate()
	if !sb.IsActive() {
		t.Error("Second activate should set active to true")
	}

	// Deactivate again
	sb.Deactivate()
	if sb.IsActive() {
		t.Error("Second deactivate should set active to false")
	}
}
