package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
)

func TestNewBackupList(t *testing.T) {
	// Create temporary directory for backups
	tempDir, err := os.MkdirTemp("", "dotkeeper-backups")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create dummy backup files
	backups := []string{
		"backup-20231026-100000.tar.gz.enc",
		"backup-20231027-110000.tar.gz.enc",
		"other-file.txt",
	}

	for _, b := range backups {
		path := filepath.Join(tempDir, b)
		err := os.WriteFile(path, []byte("dummy content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create dummy file %s: %v", b, err)
		}
		os.Chtimes(path, time.Now(), time.Now())
	}

	cfg := &config.Config{
		BackupDir: tempDir,
	}

	model := NewBackupList(cfg)

	// Send WindowSizeMsg to ensure list is rendered with some size
	// Note: Update returns a new model, we must capture it!
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel.(BackupListModel)

	// Get view again after resize
	view := model.View()

	if view == "" {
		t.Error("View returned empty string")
	}

	// We expect the view to contain the backup names (minus extension)
	expectedName1 := "backup-20231026-100000"
	expectedName2 := "backup-20231027-110000"

	if !strings.Contains(view, expectedName1) {
		t.Errorf("View does not contain %s", expectedName1)
	}
	if !strings.Contains(view, expectedName2) {
		t.Errorf("View does not contain %s", expectedName2)
	}

	// check that non-backup file is NOT present
	if strings.Contains(view, "other-file") {
		t.Error("View contains non-backup file 'other-file'")
	}
}

func TestBackupListModel_Update(t *testing.T) {
	cfg := &config.Config{
		BackupDir: ".",
	}
	model := NewBackupList(cfg)

	// Test WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, cmd := model.Update(msg)

	if cmd != nil {
		// bubbles/list update might return a command, which is fine.
	}

	m, ok := updatedModel.(BackupListModel)
	if !ok {
		t.Fatalf("Update did not return BackupListModel")
	}

	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("Expected height 50, got %d", m.height)
	}
}
