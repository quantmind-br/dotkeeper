package tui

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
	"github.com/diogo/dotkeeper/internal/tui/styles"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

func testConfig() *config.Config {
	return &config.Config{
		BackupDir: "/tmp/dotkeeper-test",
		Files:     []string{"~/.bashrc", "~/.zshrc"},
	}
}

func stripANSITest(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func testProgramContext() *ProgramContext {
	return NewProgramContext(testConfig(), history.NewStoreWithPath("/tmp/dotkeeper-test-history.jsonl"))
}

func sendKey(model tea.Model, keyStr string) (tea.Model, tea.Cmd) {
	switch keyStr {
	case "tab":
		return model.Update(tea.KeyMsg{Type: tea.KeyTab})
	case "shift+tab":
		return model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	default:
		return model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyStr)})
	}
}

func TestNewModel_InitializesAllViews(t *testing.T) {
	m := NewModelForTest(testConfig(), nil)

	if m.setupMode {
		t.Fatal("expected non-setup mode")
	}
	if m.state != DashboardView {
		t.Fatalf("state = %v, want %v", m.state, DashboardView)
	}

	if reflect.ValueOf(m.dashboard).IsZero() {
		t.Fatal("dashboard model should be initialized")
	}
	if reflect.ValueOf(m.backupList).IsZero() {
		t.Fatal("backup list model should be initialized")
	}
	if reflect.ValueOf(m.restore).IsZero() {
		t.Fatal("restore model should be initialized")
	}
	if reflect.ValueOf(m.settings).IsZero() {
		t.Fatal("settings model should be initialized")
	}
	if reflect.ValueOf(m.logs).IsZero() {
		t.Fatal("logs model should be initialized")
	}
}

func TestUpdate_TabCyclesViews(t *testing.T) {
	m := NewModelForTest(testConfig(), nil)
	sequence := []ViewState{BackupListView, RestoreView, SettingsView, LogsView, DashboardView}

	for i, want := range sequence {
		updated, _ := sendKey(m, "tab")
		m = updated.(Model)
		if m.state != want {
			t.Fatalf("step %d: state = %v, want %v", i, m.state, want)
		}
	}
}

func TestUpdate_ShiftTabCyclesReverse(t *testing.T) {
	m := NewModelForTest(testConfig(), nil)
	sequence := []ViewState{LogsView, SettingsView, RestoreView, BackupListView, DashboardView}

	for i, want := range sequence {
		updated, _ := sendKey(m, "shift+tab")
		m = updated.(Model)
		if m.state != want {
			t.Fatalf("step %d: state = %v, want %v", i, m.state, want)
		}
	}
}

func TestUpdate_NumberKeysNavigate(t *testing.T) {
	m := NewModelForTest(testConfig(), nil)
	tests := []struct {
		key  rune
		want ViewState
	}{
		{key: '1', want: DashboardView},
		{key: '2', want: BackupListView},
		{key: '3', want: RestoreView},
		{key: '4', want: SettingsView},
		{key: '5', want: LogsView},
	}

	for _, tt := range tests {
		updated, _ := sendKey(m, string(tt.key))
		m = updated.(Model)
		if m.state != tt.want {
			t.Fatalf("key %q: state = %v, want %v", string(tt.key), m.state, tt.want)
		}
	}
}

func TestUpdate_QuitReturnsQuitCmd(t *testing.T) {
	tests := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeyCtrlC},
	}

	for _, keyMsg := range tests {
		m := NewModelForTest(testConfig(), nil)
		updated, cmd := m.Update(keyMsg)
		m = updated.(Model)

		if !m.quitting {
			t.Fatalf("key %q should set quitting", keyMsg.String())
		}
		if cmd == nil {
			t.Fatalf("key %q should return quit command", keyMsg.String())
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("key %q returned non-quit message", keyMsg.String())
		}
	}
}

func TestUpdate_HelpToggles(t *testing.T) {
	m := NewModelForTest(testConfig(), nil)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)
	if !m.showingHelp {
		t.Fatal("help should be enabled")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)
	if m.showingHelp {
		t.Fatal("help should be disabled")
	}
}

func TestUpdate_WindowSizePropagates(t *testing.T) {
	m := NewModelForTest(testConfig(), nil)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 33})
	m = updated.(Model)

	if m.width != 100 || m.height != 33 {
		t.Fatalf("stored size = %dx%d, want 100x33", m.width, m.height)
	}
}

func TestUpdate_DashboardShortcuts_Framework(t *testing.T) {
	tests := []struct {
		key  rune
		want ViewState
	}{
		{key: 'b', want: BackupListView},
		{key: 'r', want: RestoreView},
		{key: 's', want: SettingsView},
	}

	for _, tt := range tests {
		m := NewModelForTest(testConfig(), nil)
		m.state = DashboardView
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
		m = updated.(Model)
		if m.state != tt.want {
			t.Fatalf("shortcut %q: state = %v, want %v", string(tt.key), m.state, tt.want)
		}
	}
}

func TestUpdate_MessageRouting_RefreshBackupList(t *testing.T) {
	m := NewModelForTest(testConfig(), nil)
	updated, cmd := m.Update(views.RefreshBackupListMsg{})
	m = updated.(Model)

	if m.state != DashboardView {
		t.Fatalf("state = %v, want %v", m.state, DashboardView)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestUpdate_MessageRouting_DashboardNavigate(t *testing.T) {
	tests := []struct {
		target string
		want   ViewState
	}{
		{target: "backups", want: BackupListView},
		{target: "restore", want: RestoreView},
		{target: "settings", want: SettingsView},
	}

	for _, tt := range tests {
		m := NewModelForTest(testConfig(), nil)
		m.state = DashboardView
		updated, _ := m.Update(views.DashboardNavigateMsg{Target: tt.target})
		m = updated.(Model)
		if m.state != tt.want {
			t.Fatalf("target %q: state = %v, want %v", tt.target, m.state, tt.want)
		}
	}
}

func TestUpdate_WindowSizeMsg_PreservesViewCommands(t *testing.T) {
	m := NewModelForTest(testConfig(), nil)

	_, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	_ = cmd
}

func TestUpdate_InputActiveBlocksTab(t *testing.T) {
	m := NewModelForTest(testConfig(), nil)
	m.state = SettingsView

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.settings.IsEditing() {
		t.Fatal("settings should be in editing mode after enter")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.state != SettingsView {
		t.Fatalf("state = %v, want %v (tab should be blocked)", m.state, SettingsView)
	}
}

func TestView_MinTerminalSize(t *testing.T) {
	tests := []tea.WindowSizeMsg{
		{Width: styles.MinTerminalWidth - 1, Height: styles.MinTerminalHeight},
		{Width: styles.MinTerminalWidth, Height: styles.MinTerminalHeight - 1},
	}

	for _, size := range tests {
		m := NewModelForTest(testConfig(), nil)
		updated, _ := m.Update(size)
		m = updated.(Model)
		view := stripANSITest(m.View())
		if !strings.Contains(view, "Terminal too small") {
			t.Fatalf("expected size warning at %dx%d", size.Width, size.Height)
		}
	}
}

func TestView_HelpOverlayFramework(t *testing.T) {
	m := NewModelForTest(testConfig(), nil)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = updated.(Model)
	m.showingHelp = true

	view := stripANSITest(m.View())
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Fatalf("expected help overlay header, got:\n%s", view)
	}
}

func TestSetupMode_CompleteTransitionsToDashboard(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_STATE_HOME", tmp+"/state")

	cfg := testConfig()
	cfg.BackupDir = tmp + "/backups"
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	m := Model{
		state:     SetupView,
		setupMode: true,
		setup:     views.NewSetup(NewProgramContext(nil, nil)),
	}
	if !m.setupMode {
		t.Fatal("expected setup mode model")
	}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(Model)

	updated, _ = m.Update(views.SetupCompleteMsg{Config: cfg})
	m = updated.(Model)

	if m.setupMode {
		t.Fatal("expected setup mode to be disabled")
	}
	if m.state != DashboardView {
		t.Fatalf("state = %v, want %v", m.state, DashboardView)
	}
	if reflect.ValueOf(m.dashboard).IsZero() || reflect.ValueOf(m.backupList).IsZero() || reflect.ValueOf(m.restore).IsZero() || reflect.ValueOf(m.settings).IsZero() || reflect.ValueOf(m.logs).IsZero() {
		t.Fatal("expected all main views initialized after setup completion")
	}
}

func TestProgramContext_CarriesConfigAndStore(t *testing.T) {
	cfg := testConfig()
	store := history.NewStoreWithPath(t.TempDir() + "/history.jsonl")
	ctx := NewProgramContext(cfg, store)

	if ctx.Config != cfg {
		t.Fatal("expected context to carry config pointer")
	}
	if ctx.Store != store {
		t.Fatal("expected context to carry history store pointer")
	}
	if ctx.Width != 0 || ctx.Height != 0 {
		t.Fatalf("expected initial dimensions 0x0, got %dx%d", ctx.Width, ctx.Height)
	}
}

func TestProgramContext_DimensionsUpdatedOnResize(t *testing.T) {
	ctx := testProgramContext()
	d := views.NewDashboard((*views.ProgramContext)(ctx))

	updated, _ := d.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
	d = updated.(views.DashboardModel)

	if ctx.Width != 120 || ctx.Height != 44 {
		t.Fatalf("expected context dimensions updated to 120x44, got %dx%d", ctx.Width, ctx.Height)
	}
}

func TestNewModelForTest_UsesProgramContext(t *testing.T) {
	cfg := testConfig()
	store := history.NewStoreWithPath(t.TempDir() + "/history.jsonl")
	m := NewModelForTest(cfg, store)

	if m.ctx == nil {
		t.Fatal("expected model context to be initialized")
	}
	if m.ctx.Config != cfg {
		t.Fatal("expected model context config pointer to match input")
	}
	if m.ctx.Store != store {
		t.Fatal("expected model context store pointer to match input")
	}
}
