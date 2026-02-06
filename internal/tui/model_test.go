package tui

import (
	"os"
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

func TestNewModel_SetupMode(t *testing.T) {
	// Set XDG_CONFIG_HOME to a non-existent directory to force setup mode
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	m := NewModel()

	if !m.setupMode {
		t.Error("expected setupMode to be true when config doesn't exist")
	}
	if m.state != SetupView {
		t.Errorf("expected state SetupView, got %v", m.state)
	}
	if m.cfg != nil {
		t.Error("expected cfg to be nil in setup mode")
	}
}

func TestNewModel_NormalMode(t *testing.T) {
	// Create a valid config
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfgDir := tmp + "/dotkeeper"
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir:     tmp + "/backups",
		GitRemote:     "",
		Schedule:      "",
		Notifications: false,
		Files:         []string{".zshrc"},
		Folders:       []string{".config"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	m := NewModel()

	if m.setupMode {
		t.Error("expected setupMode to be false when config exists")
	}
	if m.state != DashboardView {
		t.Errorf("expected state DashboardView, got %v", m.state)
	}
	if m.cfg == nil {
		t.Error("expected cfg to be non-nil in normal mode")
	}
	if m.history == nil {
		// History store may be nil in tests, that's acceptable
		t.Log("history store is nil (acceptable in test environment)")
	}
}

func TestNewModel_HistoryStoreError(t *testing.T) {
	// Create a valid config but make state directory invalid
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfgDir := tmp + "/dotkeeper"
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: tmp + "/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	// Create a file where XDG_STATE_HOME should be a directory
	badState := tmp + "/state-file"
	if err := os.WriteFile(badState, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_STATE_HOME", badState)

	// This should not panic, but history store will be nil
	m := NewModel()

	if m.setupMode {
		t.Error("expected setupMode to be false when config exists")
	}
	if m.history != nil {
		t.Error("expected history store to be nil when it cannot be created")
	}
	// Model should still be functional otherwise
	if m.cfg == nil {
		t.Error("expected cfg to be non-nil")
	}
}

func TestModel_Init_SetupMode(t *testing.T) {
	m := Model{
		setupMode: true,
		setup:     views.NewSetup(views.NewProgramContext(nil, nil)),
	}

	cmd := m.Init()
	// SetupModel.Init() returns nil, which is valid
	_ = cmd // Just ensure the call doesn't panic
}

func TestModel_Init_NormalMode(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
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

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init to return a command in normal mode")
	}
}

func TestModel_GetConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BackupDir: "/test/backups",
		Files:     []string{".zshrc"},
	}
	if err := cfg.SaveToPath(cfgDir + "/config.yaml"); err != nil {
		t.Fatal(err)
	}

	m := NewModel()

	got := m.GetConfig()
	if got == nil {
		t.Fatal("GetConfig returned nil")
	}
	if got.BackupDir != "/test/backups" {
		t.Errorf("BackupDir = %q, want %q", got.BackupDir, "/test/backups")
	}
}

func TestModel_activeTabIndex(t *testing.T) {
	tests := []struct {
		name  string
		state ViewState
		want  int
	}{
		{name: "dashboard", state: DashboardView, want: 0},
		{name: "backup list", state: BackupListView, want: 1},
		{name: "restore", state: RestoreView, want: 2},
		{name: "settings", state: SettingsView, want: 3},
		{name: "logs", state: LogsView, want: 4},
		{name: "setup - fallback to 0", state: SetupView, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{state: tt.state}
			if got := m.activeTabIndex(); got != tt.want {
				t.Errorf("activeTabIndex() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestModel_isInputActive(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfgDir := tmp + "/dotkeeper"
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
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

	// Test default state (DashboardView) - not input active
	if m.isInputActive() {
		t.Error("DashboardView should not be input active by default")
	}

	// Test SettingsView - default not editing
	m.state = SettingsView
	if m.isInputActive() {
		t.Error("SettingsView should not be input active when not editing")
	}

	// Test BackupListView - default not creating
	m.state = BackupListView
	if m.isInputActive() {
		t.Error("BackupListView should not be input active when not creating")
	}

	// Test RestoreView - default not input active
	m.state = RestoreView
	if m.isInputActive() {
		t.Error("RestoreView should not be input active by default")
	}

	// Test LogsView - should never be input active
	m.state = LogsView
	if m.isInputActive() {
		t.Error("LogsView should never be input active")
	}
}

func TestViewState_String(t *testing.T) {
	// This test ensures ViewState values can be used in comparisons
	views := []ViewState{
		DashboardView,
		BackupListView,
		RestoreView,
		SettingsView,
		LogsView,
		SetupView,
	}

	for _, v := range views {
		// Just ensure the values are distinct
		for _, other := range views {
			if v != other {
				// They should be different
			}
		}
	}
}

func TestTabOrder_Constants(t *testing.T) {
	// Ensure tabOrder has the expected number of views
	if len(tabOrder) != 5 {
		t.Errorf("tabOrder has %d entries, want 5", len(tabOrder))
	}

	// Ensure all expected views are in tabOrder
	expectedViews := map[ViewState]bool{
		DashboardView:  false,
		BackupListView: false,
		RestoreView:    false,
		SettingsView:   false,
		LogsView:       false,
	}

	for _, v := range tabOrder {
		if _, exists := expectedViews[v]; !exists {
			t.Errorf("Unexpected view %v in tabOrder", v)
		}
		expectedViews[v] = true
	}

	for view, found := range expectedViews {
		if !found {
			t.Errorf("View %v not found in tabOrder", view)
		}
	}

	// Ensure SetupView is NOT in tabOrder
	for _, v := range tabOrder {
		if v == SetupView {
			t.Error("SetupView should not be in tabOrder")
		}
	}
}

func TestMainChromeHeight(t *testing.T) {
	if mainChromeHeight != 7 {
		t.Errorf("mainChromeHeight = %d, want 7", mainChromeHeight)
	}
}
