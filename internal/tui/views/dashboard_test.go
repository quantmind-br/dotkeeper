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
	m.width = 80
	m.height = 24
	m.fileCount = 2

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil command")
	}

	view := stripANSI(m.View())

	if !strings.Contains(view, "Last Backup") {
		t.Error("View() does not contain 'Last Backup'")
	}
	if !strings.Contains(view, "Files Tracked") {
		t.Error("View() does not contain 'Files Tracked'")
	}

	if !strings.Contains(view, "Backup") {
		t.Error("View() missing 'Backup' button")
	}
	if !strings.Contains(view, "Restore") {
		t.Error("View() missing 'Restore' button")
	}
	if !strings.Contains(view, "Settings") {
		t.Error("View() missing 'Settings' button")
	}

	m.width = 40
	viewNarrow := stripANSI(m.View())

	if !strings.Contains(viewNarrow, "Last Backup") {
		t.Error("Narrow View() missing 'Last Backup'")
	}
	if !strings.Contains(viewNarrow, "Files Tracked") {
		t.Error("Narrow View() missing 'Files Tracked'")
	}
}
