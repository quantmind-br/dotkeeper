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

	expectedTitle := "Restore"
	expectedContent := "Select a backup to restore (implementation pending)"

	if !strings.Contains(view, expectedTitle) {
		t.Errorf("Expected view to contain title %q, but got:\n%s", expectedTitle, view)
	}

	if !strings.Contains(view, expectedContent) {
		t.Errorf("Expected view to contain content %q, but got:\n%s", expectedContent, view)
	}
}
