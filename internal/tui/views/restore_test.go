package views

import (
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
)

func TestRestore_View(t *testing.T) {
	cfg := &config.Config{}
	model := NewRestore(cfg)
	view := model.View()

	expectedHelp := "↑/↓: navigate"

	if !strings.Contains(view, expectedHelp) {
		t.Errorf("Expected view to contain help text %q, but got:\n%s", expectedHelp, view)
	}

	// Verify phase 0 is rendered (backup list view)
	if !strings.Contains(view, "No items") {
		t.Errorf("Expected view to show empty backup list, but got:\n%s", view)
	}
}
