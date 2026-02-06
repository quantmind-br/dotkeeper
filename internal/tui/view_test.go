package tui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
)

// stripANSI removes ANSI escape codes from strings for easier testing
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func setupModelWithConfig(t *testing.T) Model {
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
	// Set window size to avoid terminal too small warning
	m.width = 100
	m.height = 40

	return m
}

func TestView_SetupMode(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp+"/nonexistent")

	m := NewModel()
	if !m.setupMode {
		t.Fatal("expected setup mode")
	}

	view := m.View()
	if view == "" {
		t.Error("view should not be empty in setup mode")
	}
	// The view should be from the SetupModel, which we don't test in detail here
	// Just verify it doesn't panic and returns something
}

func TestView_Quitting(t *testing.T) {
	m := setupModelWithConfig(t)
	m.quitting = true

	view := m.View()
	if view != "Goodbye!\n" {
		t.Errorf("view = %q, want 'Goodbye!\\n'", view)
	}
}

func TestView_TerminalTooSmall(t *testing.T) {
	m := setupModelWithConfig(t)
	m.width = 30
	m.height = 10

	view := stripANSI(m.View())
	if !strings.Contains(view, "Terminal too small") {
		t.Errorf("view should contain 'Terminal too small', got: %s", view)
	}
	if !strings.Contains(view, "40x15") {
		t.Errorf("view should contain minimum size '40x15', got: %s", view)
	}
	if !strings.Contains(view, "30x10") {
		t.Errorf("view should contain current size '30x10', got: %s", view)
	}
}

func TestView_HelpOverlay(t *testing.T) {
	m := setupModelWithConfig(t)
	m.showingHelp = true

	view := stripANSI(m.View())
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Errorf("view should contain 'Keyboard Shortcuts', got: %s", view)
	}
}

func TestView_ContainsTitle(t *testing.T) {
	m := setupModelWithConfig(t)

	view := stripANSI(m.View())
	if !strings.Contains(view, "DotKeeper") {
		t.Errorf("view should contain 'DotKeeper', got: %s", view)
	}
	if !strings.Contains(view, "Dotfiles Backup Manager") {
		t.Errorf("view should contain app title, got: %s", view)
	}
}

func TestView_ContainsGlobalHelp(t *testing.T) {
	m := setupModelWithConfig(t)

	view := stripANSI(m.View())
	// Check for global help elements
	if !strings.Contains(view, "Tab") {
		t.Errorf("view should contain Tab help, got: %s", view)
	}
	if !strings.Contains(view, "quit") {
		t.Errorf("view should contain quit help, got: %s", view)
	}
	if !strings.Contains(view, "help") {
		t.Errorf("view should contain help reference, got: %s", view)
	}
}

func TestView_DifferentStates(t *testing.T) {
	tests := []struct {
		name  string
		state ViewState
	}{
		{name: "dashboard", state: DashboardView},
		{name: "backup list", state: BackupListView},
		{name: "restore", state: RestoreView},
		{name: "settings", state: SettingsView},
		{name: "logs", state: LogsView},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setupModelWithConfig(t)
			m.state = tt.state

			view := m.View()
			view = stripANSI(view)

			// Should contain the title
			if !strings.Contains(view, "DotKeeper") {
				t.Errorf("%s: view should contain 'DotKeeper'", tt.name)
			}

			// Should contain global help
			if !strings.Contains(view, "Tab") {
				t.Errorf("%s: view should contain Tab help", tt.name)
			}

			// View should not be empty
			if view == "" {
				t.Errorf("%s: view should not be empty", tt.name)
			}
		})
	}
}

func TestCurrentViewHelpText(t *testing.T) {
	tests := []struct {
		name        string
		state       ViewState
		wantContains string
	}{
		{name: "dashboard", state: DashboardView, wantContains: "select"},
		{name: "backup list", state: BackupListView, wantContains: ""},
		{name: "restore", state: RestoreView, wantContains: ""},
		{name: "settings", state: SettingsView, wantContains: ""},
		{name: "logs", state: LogsView, wantContains: ""},
		{name: "setup", state: SetupView, wantContains: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setupModelWithConfig(t)
			m.state = tt.state

			help := m.currentViewHelpText()

			// Dashboard should have specific help
			if tt.state == DashboardView && !strings.Contains(help, tt.wantContains) {
				t.Errorf("dashboard help should contain %q, got %q", tt.wantContains, help)
			}

			// Other views delegate to sub-models, just verify no panic
			if tt.state != DashboardView && tt.state != SetupView {
				// These views delegate to their models, which may return empty or specific text
				_ = help // Just verify it doesn't panic
			}
		})
	}
}

func TestCurrentViewHelp(t *testing.T) {
	tests := []struct {
		name  string
		state ViewState
	}{
		{name: "dashboard", state: DashboardView},
		{name: "backup list", state: BackupListView},
		{name: "restore", state: RestoreView},
		{name: "settings", state: SettingsView},
		{name: "logs", state: LogsView},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setupModelWithConfig(t)
			m.state = tt.state

			help := m.currentViewHelp()

			// Help should not panic
			// Most views return help entries, nil is valid for some cases
			_ = help
		})
	}

	t.Run("setup mode", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp+"/nonexistent")

		m := NewModel()
		if !m.setupMode {
			t.Fatal("expected setup mode")
		}

		help := m.currentViewHelp()
		if help != nil {
			t.Errorf("setup mode should return nil help, got %v", help)
		}
	})
}

func TestView_TabBarContent(t *testing.T) {
	m := setupModelWithConfig(t)

	// Test that the tab bar is rendered for different views
	states := []ViewState{DashboardView, BackupListView, RestoreView, SettingsView, LogsView}

	for _, state := range states {
		t.Run(state.String(), func(t *testing.T) {
			m.state = state
			view := stripANSI(m.View())

			// Should contain visual indication of tabs
			// The actual rendering may vary, but we should see something
			if view == "" {
				t.Error("view should not be empty")
			}
		})
	}
}

// Helper to get string representation of ViewState for test names
func (v ViewState) String() string {
	switch v {
	case DashboardView:
		return "dashboard"
	case BackupListView:
		return "backup_list"
	case RestoreView:
		return "restore"
	case SettingsView:
		return "settings"
	case LogsView:
		return "logs"
	case SetupView:
		return "setup"
	default:
		return "unknown"
	}
}
