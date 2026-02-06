package views

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/pathutil"
)

// TestNewSetup tests the NewSetup constructor
func TestNewSetup(t *testing.T) {
	m := NewSetup()

	if m.step != StepWelcome {
		t.Errorf("Expected initial step to be StepWelcome, got %d", m.step)
	}
	if m.config == nil {
		t.Error("Expected config to be initialized, got nil")
	}
	if m.pathCompleter.Input.Value() != "" {
		t.Errorf("Expected initial input value to be empty, got %q", m.pathCompleter.Input.Value())
	}
}

// TestSetupStepProgression tests advancing through setup steps with Enter key
func TestSetupStepProgression(t *testing.T) {
	model := NewSetup()

	// Step 1: Welcome -> BackupDir
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepBackupDir {
		t.Errorf("Expected step StepBackupDir, got %d", model.step)
	}

	// Step 2: BackupDir -> GitRemote
	model.pathCompleter.Input.SetValue("/tmp/backup")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepGitRemote {
		t.Errorf("Expected step StepGitRemote, got %d", model.step)
	}

	// Step 3: GitRemote -> PresetFiles
	model.pathCompleter.Input.SetValue("https://github.com/user/dotfiles.git")
	m, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepPresetFiles {
		t.Errorf("Expected step StepPresetFiles, got %d", model.step)
	}
	// Should return a command for detection
	if cmd == nil {
		t.Error("Expected detection command when entering StepPresetFiles")
	}

	// Simulate detection completion
	m, _ = model.Update(presetsDetectedMsg{
		files:   []pathutil.DotfilePreset{{Path: "~/.bashrc", Selected: true}},
		folders: []pathutil.DotfilePreset{},
	})
	model = m.(SetupModel)
	if !model.presetsLoaded {
		t.Error("Expected presetsLoaded to be true")
	}

	// Step 4: PresetFiles -> PresetFolders
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepPresetFolders {
		t.Errorf("Expected step StepPresetFolders, got %d", model.step)
	}

	// Step 5: PresetFolders -> AddFiles
	// This transition adds selected presets to addedFiles
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepAddFiles {
		t.Errorf("Expected step StepAddFiles, got %d", model.step)
	}
	if len(model.addedFiles) != 1 {
		t.Errorf("Expected 1 added file from presets, got %d", len(model.addedFiles))
	}

	// Step 6: AddFiles -> AddFolders
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepAddFolders {
		t.Errorf("Expected step StepAddFolders, got %d", model.step)
	}

	// Step 7: AddFolders -> Confirm
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepConfirm {
		t.Errorf("Expected step StepConfirm, got %d", model.step)
	}
}

// TestSetupPresetToggle tests toggling presets with Space key
func TestSetupPresetToggle(t *testing.T) {
	model := NewSetup()
	model.step = StepPresetFiles
	model.presetFiles = []pathutil.DotfilePreset{
		{Path: "~/.bashrc", Selected: false},
		{Path: "~/.zshrc", Selected: true},
	}
	model.presetCursor = 0

	// Toggle first item ON
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	model = m.(SetupModel)
	if !model.presetFiles[0].Selected {
		t.Error("Expected first item to be selected")
	}

	// Move down
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = m.(SetupModel)
	if model.presetCursor != 1 {
		t.Error("Expected cursor to move to index 1")
	}

	// Toggle second item OFF
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	model = m.(SetupModel)
	if model.presetFiles[1].Selected {
		t.Error("Expected second item to be deselected")
	}
}

// Helper to fast-forward to AddFiles step
func navigateToAddFiles(model SetupModel) SetupModel {
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	model.pathCompleter.Input.SetValue("/tmp")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return m.(SetupModel)
}

func TestSetupAddFiles(t *testing.T) {
	model := NewSetup()
	model = navigateToAddFiles(model)

	if model.step != StepAddFiles {
		t.Fatalf("Expected to be at StepAddFiles, got %d", model.step)
	}

	// Add file
	model.pathCompleter.Input.SetValue("~/.bashrc")
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if len(model.addedFiles) != 1 {
		t.Errorf("Expected 1 added file, got %d", len(model.addedFiles))
	}
}

func TestSetupComplete(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "dotkeeper")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))

	model := NewSetup()
	model = navigateToAddFiles(model)

	// Skip AddFiles
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Skip AddFolders
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if model.step != StepConfirm {
		t.Fatalf("Expected to be at StepConfirm, got %d", model.step)
	}

	// Save
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if model.step != StepComplete {
		t.Errorf("Expected StepComplete, got %d", model.step)
	}

	// Verify config file
	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Config file not created: %v", err)
	}
}

func TestSetupStepRegression(t *testing.T) {
	model := NewSetup()

	// Advance
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Regress with Esc
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model = m.(SetupModel)
	if model.step != StepWelcome {
		t.Errorf("Expected regression to StepWelcome, got %d", model.step)
	}
}

func TestSetupBrowseMode(t *testing.T) {
	model := NewSetup()
	model = navigateToAddFiles(model)

	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = m.(SetupModel)
	if !model.browsing {
		t.Fatal("expected browsing to be true after 'b' key")
	}

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model = m.(SetupModel)
	if model.browsing {
		t.Fatal("expected browsing to be false after Esc")
	}
}
