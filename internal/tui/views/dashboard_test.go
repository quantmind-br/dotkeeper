package views

import (
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
)

func TestDashboard(t *testing.T) {
	cfg := &config.Config{
		BackupDir: "/tmp/backup",
		Files:     []string{"file1", "file2"},
	}

	m := NewDashboard(cfg)

	// Test Init
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil command")
	}

	// Test View initial state
	view := m.View()
	if !strings.Contains(view, "Dashboard") {
		t.Error("View() does not contain 'Dashboard'")
	}
	if !strings.Contains(view, "Last Backup:") {
		t.Error("View() does not contain 'Last Backup:'")
	}
}
