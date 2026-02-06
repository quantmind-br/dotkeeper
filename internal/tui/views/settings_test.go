package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
)

func testSettingsConfig() *config.Config {
	return &config.Config{
		BackupDir:     "/tmp/backup",
		GitRemote:     "https://github.com/user/repo",
		Files:         []string{".zshrc", ".vimrc"},
		Folders:       []string{".config/nvim"},
		Schedule:      "daily",
		Notifications: true,
	}
}

func sendKey(t *testing.T, model SettingsModel, key tea.KeyMsg) SettingsModel {
	t.Helper()
	updated, _ := model.Update(key)
	return updated.(SettingsModel)
}

func TestSettingsNewSettings(t *testing.T) {
	cfg := testSettingsConfig()
	model := NewSettings(cfg)

	if model.config != cfg {
		t.Fatalf("expected model to keep provided config pointer")
	}
	if model.state != stateListNavigating {
		t.Fatalf("expected initial stateListNavigating, got %v", model.state)
	}
	if len(model.mainList.Items()) != 6 {
		t.Fatalf("expected 6 main list items, got %d", len(model.mainList.Items()))
	}
}

func TestSettingsViewShowsValues(t *testing.T) {
	cfg := testSettingsConfig()
	model := NewSettings(cfg)
	viewOutput := stripANSI(model.View())

	expectedStrings := []string{
		"Settings",
		"/tmp/backup",
		"https://github.com/user/repo",
		"2 files",
		"1 folders",
		"daily",
		"true",
	}

	for _, s := range expectedStrings {
		if !contains(viewOutput, s) {
			t.Errorf("Expected view to contain %q, but it didn't. View:\n%s", s, viewOutput)
		}
	}

	if contains(viewOutput, "Press 'e' to edit") {
		t.Errorf("View should not contain 'Press 'e' to edit' hint")
	}
}

func TestSettingsIsEditingByState(t *testing.T) {
	model := NewSettings(testSettingsConfig())

	states := []settingsState{
		stateListNavigating,
		stateEditingField,
		stateBrowsingFiles,
		stateBrowsingFolders,
		stateEditingSubItem,
	}

	for _, st := range states {
		model.state = st
		if !model.IsEditing() {
			t.Fatalf("expected IsEditing true in state %v", st)
		}
	}
}

func TestSettingsTransitions(t *testing.T) {
	model := NewSettings(testSettingsConfig())

	if model.state != stateListNavigating {
		t.Fatalf("expected stateListNavigating at start, got %v", model.state)
	}

	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if model.state != stateEditingField {
		t.Fatalf("expected stateEditingField after enter on first field, got %v", model.state)
	}

	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyEscape})
	if model.state != stateListNavigating {
		t.Fatalf("expected stateListNavigating after esc from field edit, got %v", model.state)
	}
}

func TestSettingsSaveWithS(t *testing.T) {
	model := NewSettings(testSettingsConfig())
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	if model.status != "Config saved successfully!" {
		t.Fatalf("expected success status after save, got status=%q err=%q", model.status, model.errMsg)
	}
	if model.errMsg != "" {
		t.Fatalf("expected empty errMsg after successful save, got %q", model.errMsg)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("expected saved config to load, got error: %v", err)
	}
	if loaded.BackupDir != model.config.BackupDir {
		t.Fatalf("saved config backup_dir mismatch: got %q want %q", loaded.BackupDir, model.config.BackupDir)
	}
}

func contains(s, substr string) bool {
	s = stripANSI(s)
	return len(s) >= len(substr) && search(s, substr)
}

func search(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSettingsFilePickerState(t *testing.T) {
	model := NewSettings(testSettingsConfig())
	// Navigate to files browsing
	model.state = stateBrowsingFiles
	// Press 'b' to open file picker
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if model.state != stateFilePickerActive {
		t.Fatalf("expected stateFilePickerActive, got %v", model.state)
	}
	if model.filePickerParent != stateBrowsingFiles {
		t.Fatalf("expected filePickerParent to be stateBrowsingFiles")
	}
	// Press Esc to return
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyEscape})
	if model.state != stateBrowsingFiles {
		t.Fatalf("expected return to stateBrowsingFiles, got %v", model.state)
	}
}
