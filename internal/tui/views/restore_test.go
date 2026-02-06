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

func TestRestore_View(t *testing.T) {
	cfg := &config.Config{}
	model := NewRestore(cfg, nil)
	view := stripANSI(model.View())

	expectedHelp := "↑/↓: navigate"

	if !strings.Contains(view, expectedHelp) {
		t.Errorf("Expected view to contain help text %q, but got:\n%s", expectedHelp, view)
	}

	// Verify phase 0 is rendered (backup list view)
	if !strings.Contains(view, "No items") {
		t.Errorf("Expected view to show empty backup list, but got:\n%s", view)
	}
}

func TestNewRestore(t *testing.T) {
	cfg := &config.Config{
		BackupDir: "/tmp/test-backups",
	}

	model := NewRestore(cfg, nil)

	if model.phase != 0 {
		t.Errorf("Expected initial phase 0, got %d", model.phase)
	}

	if model.selectedFiles == nil {
		t.Error("Expected selectedFiles map to be initialized")
	}
}

func TestRestoreBackupListLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dotkeeper-restore-test")
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
		if err := os.WriteFile(path, []byte("dummy"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		os.Chtimes(path, time.Now(), time.Now())
	}

	cfg := &config.Config{BackupDir: tempDir}
	model := NewRestore(cfg, nil)

	initCmd := model.Init()
	if initCmd == nil {
		t.Fatal("Init() should return a command")
	}

	msg := initCmd()
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(RestoreModel)

	updatedModel, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updatedModel.(RestoreModel)

	view := stripANSI(model.View())

	if !strings.Contains(view, "backup-20231026-100000") {
		t.Error("View should contain first backup")
	}
	if !strings.Contains(view, "backup-20231027-110000") {
		t.Error("View should contain second backup")
	}
	if strings.Contains(view, "other-file") {
		t.Error("View should not contain non-backup files")
	}
}

func TestRestoreModel_Update_WindowSize(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(cfg, nil)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(RestoreModel)

	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("Expected height 50, got %d", m.height)
	}
}

func TestRestoreModel_Phase0_KeyHandling(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(cfg, nil)

	// Test 'r' key triggers refresh
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m := updatedModel.(RestoreModel)

	if m.phase != 0 {
		t.Errorf("Phase should remain 0, got %d", m.phase)
	}
	if cmd == nil {
		t.Error("'r' key should return a refresh command")
	}
}

func TestRestoreModel_SelectedFilesCount(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(cfg, nil)

	model.selectedFiles = map[string]bool{
		"/path/file1": true,
		"/path/file2": false,
		"/path/file3": true,
	}

	count := model.countSelectedFiles()
	if count != 2 {
		t.Errorf("Expected 2 selected files, got %d", count)
	}
}

func TestRestoreModel_GetSelectedFilePaths(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(cfg, nil)

	model.selectedFiles = map[string]bool{
		"/path/file1": true,
		"/path/file2": false,
		"/path/file3": true,
	}

	paths := model.getSelectedFilePaths()
	if len(paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(paths))
	}

	// Check both selected files are in the result
	found1, found3 := false, false
	for _, p := range paths {
		if p == "/path/file1" {
			found1 = true
		}
		if p == "/path/file3" {
			found3 = true
		}
	}
	if !found1 || !found3 {
		t.Error("Expected both selected files to be in paths")
	}
}

func TestRestoreModel_Phase2_ZeroFilesBlocked(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(cfg, nil)
	model.phase = 2
	model.selectedFiles = map[string]bool{
		"/path/file1": false,
		"/path/file2": false,
	}

	// Press enter with no files selected
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(RestoreModel)

	if m.phase != 2 {
		t.Errorf("Phase should remain 2 when no files selected, got %d", m.phase)
	}
	if m.restoreError == "" {
		t.Error("Should show error when trying to restore with no files selected")
	}
	if !strings.Contains(m.restoreError, "Select at least one file") {
		t.Errorf("Error should mention selecting files, got: %s", m.restoreError)
	}
}

func TestRestoreModel_ESCNavigation(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}

	// Test ESC in phase 1 returns to phase 0
	model := NewRestore(cfg, nil)
	model.phase = 1
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updatedModel.(RestoreModel)
	if m.phase != 0 {
		t.Errorf("ESC in phase 1 should return to phase 0, got %d", m.phase)
	}

	// Test ESC in phase 2 returns to phase 0
	model = NewRestore(cfg, nil)
	model.phase = 2
	model.selectedFiles = make(map[string]bool)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(RestoreModel)
	if m.phase != 0 {
		t.Errorf("ESC in phase 2 should return to phase 0, got %d", m.phase)
	}

	// Test ESC in phase 4 returns to phase 2
	model = NewRestore(cfg, nil)
	model.phase = 4
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(RestoreModel)
	if m.phase != 2 {
		t.Errorf("ESC in phase 4 should return to phase 2, got %d", m.phase)
	}
}
