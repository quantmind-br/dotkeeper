package views

import (
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
)

func TestFileBrowser(t *testing.T) {
	cfg := &config.Config{
		BackupDir: "/tmp/backup",
	}

	model := NewFileBrowser(cfg)

	// Verify basic initialization
	if model.config != cfg {
		t.Errorf("Expected config to be set")
	}

	// Verify Init
	cmd := model.Init()
	if cmd == nil {
		t.Errorf("Expected Init to return a command")
	}

	// Verify View
	view := model.View()
	if view == "" {
		t.Errorf("Expected View to return content")
	}
}
