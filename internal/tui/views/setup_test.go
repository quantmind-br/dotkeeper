package views

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
)

// TestNewSetup tests the NewSetup constructor
func TestNewSetup(t *testing.T) {
	m := NewSetup()

	// Verify initial step is StepWelcome
	if m.step != StepWelcome {
		t.Errorf("Expected initial step to be StepWelcome, got %d", m.step)
	}

	// Verify config is initialized
	if m.config == nil {
		t.Error("Expected config to be initialized, got nil")
	}

	// Verify input is initialized
	if m.input.Value() != "" {
		t.Errorf("Expected initial input value to be empty, got %q", m.input.Value())
	}

	// Verify addedFiles and addedFolders are empty
	if len(m.addedFiles) != 0 {
		t.Errorf("Expected addedFiles to be empty, got %d items", len(m.addedFiles))
	}
	if len(m.addedFolders) != 0 {
		t.Errorf("Expected addedFolders to be empty, got %d items", len(m.addedFolders))
	}
}

// TestSetupStepProgression tests advancing through setup steps with Enter key
func TestSetupStepProgression(t *testing.T) {
	model := NewSetup()

	// Step 1: Welcome -> BackupDir
	if model.step != StepWelcome {
		t.Fatalf("Expected to start at StepWelcome, got %d", model.step)
	}

	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepBackupDir {
		t.Errorf("Expected step to advance to StepBackupDir, got %d", model.step)
	}

	// Step 2: BackupDir -> GitRemote
	model.input.SetValue("/tmp/backup")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepGitRemote {
		t.Errorf("Expected step to advance to StepGitRemote, got %d", model.step)
	}
	if model.config.BackupDir != "/tmp/backup" {
		t.Errorf("Expected BackupDir to be '/tmp/backup', got %q", model.config.BackupDir)
	}

	// Step 3: GitRemote -> AddFiles
	model.input.SetValue("https://github.com/user/dotfiles.git")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepAddFiles {
		t.Errorf("Expected step to advance to StepAddFiles, got %d", model.step)
	}
	if model.config.GitRemote != "https://github.com/user/dotfiles.git" {
		t.Errorf("Expected GitRemote to be set, got %q", model.config.GitRemote)
	}

	// Step 4: AddFiles -> AddFolders (with empty input)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepAddFolders {
		t.Errorf("Expected step to advance to StepAddFolders, got %d", model.step)
	}

	// Step 5: AddFolders -> Confirm (with empty input)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepConfirm {
		t.Errorf("Expected step to advance to StepConfirm, got %d", model.step)
	}
}

// TestSetupStepRegression tests going back with Esc key
func TestSetupStepRegression(t *testing.T) {
	model := NewSetup()

	// Advance to StepBackupDir
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepBackupDir {
		t.Fatalf("Expected to be at StepBackupDir, got %d", model.step)
	}

	// Press Esc -> should go back to StepWelcome
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model = m.(SetupModel)
	if model.step != StepWelcome {
		t.Errorf("Expected step to regress to StepWelcome, got %d", model.step)
	}

	// Advance to StepGitRemote
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	model.input.SetValue("/tmp/backup")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if model.step != StepGitRemote {
		t.Fatalf("Expected to be at StepGitRemote, got %d", model.step)
	}

	// Press Esc -> should go back to StepBackupDir
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model = m.(SetupModel)
	if model.step != StepBackupDir {
		t.Errorf("Expected step to regress to StepBackupDir, got %d", model.step)
	}
}

// TestSetupAddFiles tests adding files in StepAddFiles
func TestSetupAddFiles(t *testing.T) {
	model := NewSetup()

	// Navigate to StepAddFiles
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	model.input.SetValue("/tmp/backup")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	model.input.SetValue("https://github.com/user/dotfiles.git")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if model.step != StepAddFiles {
		t.Fatalf("Expected to be at StepAddFiles, got %d", model.step)
	}

	// Add first file
	model.input.SetValue("~/.bashrc")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if len(model.addedFiles) != 1 {
		t.Errorf("Expected 1 added file, got %d", len(model.addedFiles))
	}
	expectedPath := expandHome("~/.bashrc")
	if model.addedFiles[0] != expectedPath {
		t.Errorf("Expected first file to be %q, got %q", expectedPath, model.addedFiles[0])
	}
	if model.step != StepAddFiles {
		t.Errorf("Expected to still be at StepAddFiles after adding file, got %d", model.step)
	}

	// Add second file
	model.input.SetValue("~/.zshrc")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if len(model.addedFiles) != 2 {
		t.Errorf("Expected 2 added files, got %d", len(model.addedFiles))
	}
	expectedPath = expandHome("~/.zshrc")
	if model.addedFiles[1] != expectedPath {
		t.Errorf("Expected second file to be %q, got %q", expectedPath, model.addedFiles[1])
	}

	// Press Enter with empty input -> should advance to StepAddFolders
	model.input.SetValue("")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if model.step != StepAddFolders {
		t.Errorf("Expected step to advance to StepAddFolders, got %d", model.step)
	}
	if len(model.addedFiles) != 2 {
		t.Errorf("Expected addedFiles to still have 2 items, got %d", len(model.addedFiles))
	}
}

// TestSetupAddFolders tests adding folders in StepAddFolders
func TestSetupAddFolders(t *testing.T) {
	model := NewSetup()

	// Navigate to StepAddFolders
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	model.input.SetValue("/tmp/backup")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	model.input.SetValue("https://github.com/user/dotfiles.git")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if model.step != StepAddFolders {
		t.Fatalf("Expected to be at StepAddFolders, got %d", model.step)
	}

	// Add first folder
	model.input.SetValue("~/.config")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if len(model.addedFolders) != 1 {
		t.Errorf("Expected 1 added folder, got %d", len(model.addedFolders))
	}
	expectedFolderPath := expandHome("~/.config")
	if model.addedFolders[0] != expectedFolderPath {
		t.Errorf("Expected first folder to be %q, got %q", expectedFolderPath, model.addedFolders[0])
	}
	if model.step != StepAddFolders {
		t.Errorf("Expected to still be at StepAddFolders after adding folder, got %d", model.step)
	}

	// Add second folder
	model.input.SetValue("~/.local/share")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if len(model.addedFolders) != 2 {
		t.Errorf("Expected 2 added folders, got %d", len(model.addedFolders))
	}

	// Press Enter with empty input -> should advance to StepConfirm
	model.input.SetValue("")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if model.step != StepConfirm {
		t.Errorf("Expected step to advance to StepConfirm, got %d", model.step)
	}
	if len(model.addedFolders) != 2 {
		t.Errorf("Expected addedFolders to still have 2 items, got %d", len(model.addedFolders))
	}
}

// TestSetupComplete tests completing setup and saving config
func TestSetupComplete(t *testing.T) {
	// Create temporary directory for config
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "dotkeeper")

	// Override XDG_CONFIG_HOME for this test
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))

	model := NewSetup()

	// Navigate through all steps
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	model.input.SetValue("/tmp/backup")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	model.input.SetValue("https://github.com/user/dotfiles.git")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Add a file
	model.input.SetValue("~/.bashrc")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Move to AddFolders
	model.input.SetValue("")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Add a folder
	model.input.SetValue("~/.config")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Move to Confirm
	model.input.SetValue("")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if model.step != StepConfirm {
		t.Fatalf("Expected to be at StepConfirm, got %d", model.step)
	}

	// Verify config values before saving
	if model.config.BackupDir != "/tmp/backup" {
		t.Errorf("Expected BackupDir to be '/tmp/backup', got %q", model.config.BackupDir)
	}
	if model.config.GitRemote != "https://github.com/user/dotfiles.git" {
		t.Errorf("Expected GitRemote to be set, got %q", model.config.GitRemote)
	}

	// Press Enter to save and complete
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if model.step != StepComplete {
		t.Errorf("Expected step to be StepComplete, got %d", model.step)
	}

	// Verify config file was created
	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Expected config file to exist at %s, got error: %v", configPath, err)
	}

	// Verify config file contents
	savedCfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Errorf("Failed to load saved config: %v", err)
	}

	if savedCfg.BackupDir != "/tmp/backup" {
		t.Errorf("Expected saved BackupDir to be '/tmp/backup', got %q", savedCfg.BackupDir)
	}
	if savedCfg.GitRemote != "https://github.com/user/dotfiles.git" {
		t.Errorf("Expected saved GitRemote to be set, got %q", savedCfg.GitRemote)
	}
	expectedFile := expandHome("~/.bashrc")
	if len(savedCfg.Files) != 1 || savedCfg.Files[0] != expectedFile {
		t.Errorf("Expected saved Files to contain %q, got %v", expectedFile, savedCfg.Files)
	}
	expectedFolder := expandHome("~/.config")
	if len(savedCfg.Folders) != 1 || savedCfg.Folders[0] != expectedFolder {
		t.Errorf("Expected saved Folders to contain %q, got %v", expectedFolder, savedCfg.Folders)
	}
}

// TestSetupCtrlC tests that Ctrl+C quits the application
func TestSetupCtrlC(t *testing.T) {
	model := NewSetup()

	m, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	_ = m.(SetupModel)

	// Verify we're still in the model (Ctrl+C returns tea.Quit command)
	if cmd == nil {
		t.Error("Expected Ctrl+C to return a command (tea.Quit)")
	}
}

// TestSetupDefaultBackupDir tests that default backup dir is used when empty
func TestSetupDefaultBackupDir(t *testing.T) {
	model := NewSetup()

	// Navigate to StepBackupDir
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Press Enter with empty input (should use default)
	model.input.SetValue("")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Verify default backup dir is set
	home, _ := os.UserHomeDir()
	expectedDefault := filepath.Join(home, ".dotfiles")
	if model.config.BackupDir != expectedDefault {
		t.Errorf("Expected default BackupDir to be %q, got %q", expectedDefault, model.config.BackupDir)
	}
}

// TestSetupInputFocus tests that input is focused at appropriate steps
func TestSetupInputFocus(t *testing.T) {
	model := NewSetup()

	// At Welcome, input should not be focused
	if model.input.Focused() {
		t.Error("Expected input to not be focused at StepWelcome")
	}

	// After advancing to BackupDir, input should be focused
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	if !model.input.Focused() {
		t.Error("Expected input to be focused at StepBackupDir")
	}
}

// TestSetupEmptyGitRemote tests that empty git remote is allowed
func TestSetupEmptyGitRemote(t *testing.T) {
	model := NewSetup()

	// Navigate to StepGitRemote
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)
	model.input.SetValue("/tmp/backup")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Press Enter with empty git remote
	model.input.SetValue("")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	if model.step != StepAddFiles {
		t.Errorf("Expected step to advance to StepAddFiles, got %d", model.step)
	}
	if model.config.GitRemote != "" {
		t.Errorf("Expected GitRemote to be empty, got %q", model.config.GitRemote)
	}
}

// TestSetupNotificationsDefault tests that notifications default to true
func TestSetupNotificationsDefault(t *testing.T) {
	tempDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))

	model := NewSetup()

	// Navigate through all steps without adding files/folders
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	model.input.SetValue("/tmp/backup")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	model.input.SetValue("https://github.com/user/dotfiles.git")
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Skip files
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Skip folders
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Save config
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = m.(SetupModel)

	// Verify notifications is set to true
	if !model.config.Notifications {
		t.Error("Expected Notifications to default to true when no files/folders added")
	}
}
