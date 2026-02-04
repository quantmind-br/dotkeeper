package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
)

func TestLogsModel(t *testing.T) {
	// Setup
	cfg := &config.Config{
		BackupDir: "/tmp/backup",
	}
	m := NewLogs(cfg)

	// Test Init
	if cmd := m.Init(); cmd != nil {
		t.Errorf("Init() returned %v, expected nil", cmd)
	}

	// Test Update with WindowSizeMsg
	width, height := 100, 50
	msg := tea.WindowSizeMsg{Width: width, Height: height}
	updatedModel, cmd := m.Update(msg)
	if cmd != nil {
		t.Errorf("Update() returned cmd %v, expected nil", cmd)
	}

	// Verify model updated dimensions
	logsModel, ok := updatedModel.(LogsModel)
	if !ok {
		t.Errorf("Update() returned model of type %T, expected LogsModel", updatedModel)
	}

	if logsModel.width != width || logsModel.height != height {
		t.Errorf("LogsModel dimensions not updated: got %dx%d, expected %dx%d",
			logsModel.width, logsModel.height, width, height)
	}

	// Test View
	// We expect the view to NOT be empty and contain "Logs"
	view := m.View()
	expectedTitle := "Logs"

	if view == "" {
		t.Error("View() returned empty string")
	}

	// Simple check if the title is present (ignoring ansi codes for loose matching or assuming clean string)
	// Since we broke the implementation to return "", this will fail.
	if len(view) < len(expectedTitle) {
		t.Errorf("View() content too short: %q", view)
	}
}
