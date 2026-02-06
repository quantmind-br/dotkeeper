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
	model := NewRestore(NewProgramContext(cfg, nil))
	view := stripANSI(model.View())

	// Verify phase 0 is rendered (backup list view)
	if !strings.Contains(view, "No items") {
		t.Errorf("Expected view to show empty backup list, but got:\n%s", view)
	}

	// StatusHelpText is separate from View()
	helpText := model.StatusHelpText()
	if !strings.Contains(helpText, "navigate") {
		t.Errorf("Expected StatusHelpText to contain 'navigate', got: %s", helpText)
	}
}

func TestNewRestore(t *testing.T) {
	cfg := &config.Config{
		BackupDir: "/tmp/test-backups",
	}

	model := NewRestore(NewProgramContext(cfg, nil))

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
	model := NewRestore(NewProgramContext(cfg, nil))

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
	model := NewRestore(NewProgramContext(cfg, nil))

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(RestoreModel)

	if m.ctx.Width != 100 {
		t.Errorf("Expected width 100, got %d", m.ctx.Width)
	}
	if m.ctx.Height != 50 {
		t.Errorf("Expected height 50, got %d", m.ctx.Height)
	}
}

func TestRestoreModel_Phase0_KeyHandling(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(NewProgramContext(cfg, nil))

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
	model := NewRestore(NewProgramContext(cfg, nil))

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
	model := NewRestore(NewProgramContext(cfg, nil))

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
	model := NewRestore(NewProgramContext(cfg, nil))
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
	model := NewRestore(NewProgramContext(cfg, nil))
	model.phase = 1
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updatedModel.(RestoreModel)
	if m.phase != 0 {
		t.Errorf("ESC in phase 1 should return to phase 0, got %d", m.phase)
	}

	// Test ESC in phase 2 returns to phase 0
	model = NewRestore(NewProgramContext(cfg, nil))
	model.phase = 2
	model.selectedFiles = make(map[string]bool)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(RestoreModel)
	if m.phase != 0 {
		t.Errorf("ESC in phase 2 should return to phase 0, got %d", m.phase)
	}

	// Test ESC in phase 4 returns to phase 2
	model = NewRestore(NewProgramContext(cfg, nil))
	model.phase = 4
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(RestoreModel)
	if m.phase != 2 {
		t.Errorf("ESC in phase 4 should return to phase 2, got %d", m.phase)
	}
}

func TestRestoreHelpBindings(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(NewProgramContext(cfg, nil))

	// Test phase 0 help bindings
	model.phase = 0
	help := model.HelpBindings()
	if help == nil {
		t.Error("HelpBindings should not return nil")
	}

	// Just verify we have some help entries
	if len(help) == 0 {
		t.Error("HelpBindings should return at least one entry")
	}
}

func TestRestoreIsInputActive(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(NewProgramContext(cfg, nil))

	// Only phasePassword (phase 1) should be input active
	const phasePassword = 1

	// Test phase 0 - not input active
	model.phase = 0
	if model.IsInputActive() {
		t.Error("Phase 0 should not be input active")
	}

	// Test phase 1 - password input IS active
	model.phase = phasePassword
	if !model.IsInputActive() {
		t.Error("Phase 1 should be input active (password)")
	}

	// Test phase 2 - file selection (not input)
	model.phase = 2
	if model.IsInputActive() {
		t.Error("Phase 2 should not be input active (file list navigation)")
	}

	// Test phase 3 - restoring
	model.phase = 3
	if model.IsInputActive() {
		t.Error("Phase 3 should not be input active (restoring)")
	}
}

func TestFileItem_Title(t *testing.T) {
	tests := []struct {
		name     string
		selected bool
		path     string
		want     string
	}{
		{
			name:     "not selected",
			selected: false,
			path:     "/home/user/.bashrc",
			want:     "[ ] /home/user/.bashrc",
		},
		{
			name:     "selected",
			selected: true,
			path:     "/home/user/.bashrc",
			want:     "[x] /home/user/.bashrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := fileItem{
				path:     tt.path,
				selected: tt.selected,
			}
			got := item.Title()
			if got != tt.want {
				t.Errorf("Title() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFileItem_Description(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{
			name: "zero bytes",
			size: 0,
			want: "0 bytes",
		},
		{
			name: "single byte",
			size: 1,
			want: "1 bytes",
		},
		{
			name: "kilobyte",
			size: 1024,
			want: "1024 bytes",
		},
		{
			name: "megabyte",
			size: 1024 * 1024,
			want: "1048576 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := fileItem{size: tt.size}
			got := item.Description()
			if got != tt.want {
				t.Errorf("Description() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFileItem_FilterValue(t *testing.T) {
	path := "/home/user/.config/nvim/init.lua"
	item := fileItem{path: path}

	got := item.FilterValue()
	if got != path {
		t.Errorf("FilterValue() = %q, want %q", got, path)
	}
}

func TestRestoreRefresh(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(NewProgramContext(cfg, nil))

	cmd := model.Refresh()
	if cmd == nil {
		t.Error("Refresh() should return a command")
	}
}

func TestRestoreModel_StatusHelpText_Phases(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(NewProgramContext(cfg, nil))

	// Test each phase's help text
	phases := []restorePhase{
		phaseBackupList,
		phasePassword,
		phaseFileSelect,
		phaseRestoring,
		phaseDiffPreview,
		phaseResults,
	}

	for _, phase := range phases {
		model.phase = phase
		helpText := model.StatusHelpText()
		if helpText == "" {
			t.Errorf("Phase %d should have help text", phase)
		}
	}
}

func TestRestoreHelpBindings_Phases(t *testing.T) {
	cfg := &config.Config{BackupDir: "."}
	model := NewRestore(NewProgramContext(cfg, nil))

	// Test each phase's help bindings
	phases := []restorePhase{
		phaseBackupList,
		phasePassword,
		phaseFileSelect,
		phaseDiffPreview,
		phaseResults,
	}

	for _, phase := range phases {
		model.phase = phase
		bindings := model.HelpBindings()
		if bindings == nil {
			t.Errorf("Phase %d HelpBindings should not be nil", phase)
		}
	}
}
