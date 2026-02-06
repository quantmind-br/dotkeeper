package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Test Quit binding
	if !km.Quit.Enabled() {
		t.Error("Quit binding should be enabled")
	}
	if len(km.Quit.Keys()) == 0 {
		t.Error("Quit binding has no keys")
	}
	quitKeys := km.Quit.Keys()
	if len(quitKeys) < 2 {
		t.Error("Quit binding should have at least 2 keys")
	}

	// Test Tab binding
	if len(km.Tab.Keys()) == 0 {
		t.Error("Tab binding has no keys")
	}

	// Test ShiftTab binding
	if len(km.ShiftTab.Keys()) == 0 {
		t.Error("ShiftTab binding has no keys")
	}

	// Test Help binding
	helpKey := km.Help.Keys()
	if len(helpKey) == 0 || helpKey[0] != "?" {
		t.Errorf("Help key should be '?', got %v", helpKey)
	}
}

func TestPropagateWindowSize(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	tests := []struct {
		name          string
		width         int
		height        int
		expectNoPanic bool
	}{
		{name: "normal size", width: 80, height: 24, expectNoPanic: true},
		{name: "small size", width: 40, height: 15, expectNoPanic: true},
		{name: "zero width", width: 0, height: 24, expectNoPanic: true},
		{name: "zero height", width: 80, height: 0, expectNoPanic: true},
		{name: "negative values", width: -10, height: -10, expectNoPanic: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil && tt.expectNoPanic {
					t.Errorf("propagateWindowSize panicked with %v", r)
				}
			}()

			m.propagateWindowSize(tea.WindowSizeMsg{Width: tt.width, Height: tt.height})
		})
	}
}

func TestRefreshCmdForState(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	tests := []struct {
		name      string
		state     ViewState
		wantNil   bool
	}{
		{name: "dashboard", state: DashboardView, wantNil: false},
		{name: "backup list", state: BackupListView, wantNil: false},
		{name: "restore", state: RestoreView, wantNil: false},
		{name: "logs", state: LogsView, wantNil: false},
		{name: "settings", state: SettingsView, wantNil: true},
		{name: "setup", state: SetupView, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := m.refreshCmdForState(tt.state)
			if (cmd == nil) != tt.wantNil {
				t.Errorf("refreshCmdForState(%v) = %v, want nil %v", tt.state, cmd == nil, tt.wantNil)
			}
		})
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	// Normal mode window size
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	model, cmd := m.Update(msg)
	m = model.(Model)

	if m.width != 80 {
		t.Errorf("width = %d, want 80", m.width)
	}
	if m.height != 24 {
		t.Errorf("height = %d, want 24", m.height)
	}
	if cmd != nil {
		t.Error("expected nil command for WindowSizeMsg")
	}
}

func TestUpdate_SetupModeWindowAndComplete(t *testing.T) {
	t.Run("setup mode window size", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)

		m := NewModel()
		if !m.setupMode {
			t.Fatal("expected setup mode")
		}

		msg := tea.WindowSizeMsg{Width: 80, Height: 24}
		model, _ := m.Update(msg)
		m = model.(Model)

		if m.width != 80 {
			t.Errorf("width = %d, want 80", m.width)
		}
		if m.height != 24 {
			t.Errorf("height = %d, want 24", m.height)
		}
	})

	t.Run("setup complete transitions to normal mode", func(t *testing.T) {
		tmp := t.TempDir()
		// Start with empty XDG_CONFIG_HOME to trigger setup mode
		t.Setenv("XDG_CONFIG_HOME", tmp+"/nonexistent")
		t.Setenv("XDG_STATE_HOME", tmp+"/state")

		m := NewModel()
		if !m.setupMode {
			t.Fatal("expected setup mode")
		}

		// Now create a config file that will be loaded after setup completes
		cfgDir := tmp + "/dotkeeper"
		if err := createConfigDir(cfgDir); err != nil {
			t.Fatal(err)
		}
		cfg := &config.Config{
			BackupDir: tmp + "/backups",
			Files:     []string{".zshrc"},
		}
		if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
			t.Fatal(err)
		}
		// Set XDG_CONFIG_HOME to point to the config directory for loading
		t.Setenv("XDG_CONFIG_HOME", tmp)

		// Send setup complete message
		msg := views.SetupCompleteMsg{}
		model, cmd := m.Update(msg)
		m = model.(Model)

		if m.setupMode {
			t.Error("should have exited setup mode")
		}
		if m.state != DashboardView {
			t.Errorf("state = %v, want DashboardView", m.state)
		}
		if m.cfg == nil {
			t.Error("cfg should be loaded")
		}
		if cmd != nil {
			t.Error("expected nil command after setup complete")
		}
	})
}

func TestUpdate_HelpToggle(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	// Toggle help on
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	model, _ := m.Update(msg)
	m = model.(Model)

	if !m.showingHelp {
		t.Error("expected showingHelp to be true")
	}

	// Toggle help off
	model, _ = m.Update(msg)
	m = model.(Model)

	if m.showingHelp {
		t.Error("expected showingHelp to be false")
	}

	// Press any key while help is showing should close it
	m.showingHelp = true
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	model, _ = m.Update(msg)
	m = model.(Model)

	if m.showingHelp {
		t.Error("expected showingHelp to be false after pressing key")
	}
}

func TestUpdate_Quit(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	// Test 'q' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	model, cmd := m.Update(msg)
	m = model.(Model)

	if !m.quitting {
		t.Error("expected quitting to be true")
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}

	// Reset and test ctrl+c
	m.quitting = false
	m = NewModel()
	msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	model, cmd = m.Update(msg)
	m = model.(Model)

	if !m.quitting {
		t.Error("expected quitting to be true for ctrl+c")
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestUpdate_TabNavigation(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	// Start at DashboardView (index 0)
	if m.state != DashboardView {
		t.Fatalf("initial state = %v, want DashboardView", m.state)
	}

	// Press tab to go to next view
	msg := tea.KeyMsg{Type: tea.KeyTab}
	model, _ := m.Update(msg)
	m = model.(Model)

	if m.state != BackupListView {
		t.Errorf("state = %v, want BackupListView", m.state)
	}

	// Press tab again
	model, _ = m.Update(msg)
	m = model.(Model)

	if m.state != RestoreView {
		t.Errorf("state = %v, want RestoreView", m.state)
	}
}

func TestUpdate_ShiftTabNavigation(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	// Move to LogsView (last in tabOrder)
	m.state = LogsView

	// Press shift+tab to go to previous view
	msg := key.NewBinding(key.WithKeys("shift+tab"))
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(msg.Keys()[0])}
	// Actually, shift+tab is a special key type
	keyMsg = tea.KeyMsg{Type: tea.KeyShiftTab}

	model, _ := m.Update(keyMsg)
	m = model.(Model)

	if m.state != SettingsView {
		t.Errorf("state = %v, want SettingsView", m.state)
	}
}

func TestUpdate_NumberKeyNavigation(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	tests := []struct {
		key   string
		state ViewState
	}{
		{key: "1", state: DashboardView},
		{key: "2", state: BackupListView},
		{key: "3", state: RestoreView},
		{key: "4", state: SettingsView},
		{key: "5", state: LogsView},
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			m.state = DashboardView // Reset to dashboard

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			model, _ := m.Update(msg)
			m = model.(Model)

			if m.state != tt.state {
				t.Errorf("state = %v, want %v", m.state, tt.state)
			}
		})
	}
}

func TestUpdate_DashboardShortcuts(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	// Ensure we're on dashboard
	m.state = DashboardView

	// Test 'b' for backups
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")}
	model, _ := m.Update(msg)
	m = model.(Model)

	if m.state != BackupListView {
		t.Errorf("state = %v, want BackupListView", m.state)
	}

	// Reset and test 'r' for restore
	m.state = DashboardView
	m = NewModel()
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	model, _ = m.Update(msg)
	m = model.(Model)

	if m.state != RestoreView {
		t.Errorf("state = %v, want RestoreView", m.state)
	}

	// Reset and test 's' for settings
	m.state = DashboardView
	m = NewModel()
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	model, _ = m.Update(msg)
	m = model.(Model)

	if m.state != SettingsView {
		t.Errorf("state = %v, want SettingsView", m.state)
	}
}

func TestUpdate_RefreshBackupListMsg(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	// Send RefreshBackupListMsg
	msg := views.RefreshBackupListMsg{}
	model, cmd := m.Update(msg)
	m = model.(Model)

	if cmd == nil {
		t.Error("expected command from RefreshBackupListMsg")
	}
}

func TestUpdate_DashboardNavigateMsg(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := createConfigDir(cfgDir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	tests := []struct {
		name  string
		target string
		state ViewState
	}{
		{name: "backups", target: "backups", state: BackupListView},
		{name: "restore", target: "restore", state: RestoreView},
		{name: "settings", target: "settings", state: SettingsView},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.state = DashboardView // Reset

			msg := views.DashboardNavigateMsg{Target: tt.target}
			model, _ := m.Update(msg)
			m = model.(Model)

			if m.state != tt.state {
				t.Errorf("state = %v, want %v", m.state, tt.state)
			}
		})
	}
}

// Helper function to create config directory
func createConfigDir(cfgDir string) error {
	// This is handled by the config package's SaveToPath
	return nil
}
