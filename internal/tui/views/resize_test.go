package views

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

// resizeTestSizes defines the terminal size matrix used across tests.
var resizeTestSizes = []tea.WindowSizeMsg{
	{Width: 0, Height: 0},     // Zero (terminal multiplexer init)
	{Width: 1, Height: 1},     // Absurdly small
	{Width: 39, Height: 14},   // Below minimum
	{Width: 40, Height: 15},   // Exact minimum
	{Width: 80, Height: 24},   // Standard
	{Width: 200, Height: 100}, // Very large
	{Width: 300, Height: 10},  // Ultra-wide
	{Width: 40, Height: 100},  // Ultra-tall
}

func testCfg() *config.Config {
	return &config.Config{
		BackupDir: "/tmp/dotkeeper-test-backup",
		Files:     []string{"~/.bashrc", "~/.zshrc"},
	}
}

// --- A: No-Panic Tests (ALL views, ALL sizes) ---

func TestAllViewsNoPanicOnResize_Dashboard(t *testing.T) {
	cfg := testCfg()
	for _, size := range resizeTestSizes {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			m := NewDashboard(cfg)
			model, _ := m.Update(size)
			dm := model.(DashboardModel)
			_ = dm.View()
		})
	}
}

func TestAllViewsNoPanicOnResize_BackupList(t *testing.T) {
	cfg := testCfg()
	for _, size := range resizeTestSizes {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			m := NewBackupList(cfg, nil)
			model, _ := m.Update(size)
			bm := model.(BackupListModel)
			_ = bm.View()
		})
	}
}

func TestAllViewsNoPanicOnResize_Restore(t *testing.T) {
	cfg := testCfg()
	for _, size := range resizeTestSizes {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			m := NewRestore(cfg, nil)
			model, _ := m.Update(size)
			rm := model.(RestoreModel)
			_ = rm.View()
		})
	}
}

func TestAllViewsNoPanicOnResize_Settings(t *testing.T) {
	cfg := testCfg()
	for _, size := range resizeTestSizes {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			m := NewSettings(cfg)
			model, _ := m.Update(size)
			sm := model.(SettingsModel)
			_ = sm.View()
		})
	}
}

func TestAllViewsNoPanicOnResize_Logs(t *testing.T) {
	cfg := testCfg()
	for _, size := range resizeTestSizes {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			m := NewLogs(cfg)
			model, _ := m.Update(size)
			lm := model.(LogsModel)
			_ = lm.View()
		})
	}
}

func TestAllViewsNoPanicOnResize_Setup(t *testing.T) {
	for _, size := range resizeTestSizes {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			m := NewSetup()
			model, _ := m.Update(size)
			sm := model.(SetupModel)
			_ = sm.View()
		})
	}
}

// --- B: Dashboard Responsive Layout ---

func TestDashboardResponsiveLayout(t *testing.T) {
	cfg := testCfg()

	// Wide layout (horizontal cards): width >= BreakpointWide (80)
	m := NewDashboard(cfg)
	m.fileCount = 2
	model, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	dm := model.(DashboardModel)
	wideView := stripANSI(dm.View())
	wideLines := strings.Split(strings.TrimRight(wideView, "\n"), "\n")

	// Narrow layout (vertical cards): width < BreakpointWide
	model, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 30})
	dm = model.(DashboardModel)
	narrowView := stripANSI(dm.View())
	narrowLines := strings.Split(strings.TrimRight(narrowView, "\n"), "\n")

	// Vertical layout should produce more lines than horizontal
	if len(narrowLines) <= len(wideLines) {
		t.Errorf("narrow layout (%d lines) should have more lines than wide layout (%d lines)",
			len(narrowLines), len(wideLines))
	}
}

func TestDashboardResponsiveButtons(t *testing.T) {
	cfg := testCfg()

	// Wide: buttons should be horizontal
	m := NewDashboard(cfg)
	model, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	dm := model.(DashboardModel)
	wideView := stripANSI(dm.View())

	// Narrow: buttons should stack vertically, producing more lines
	model, _ = m.Update(tea.WindowSizeMsg{Width: 50, Height: 30})
	dm = model.(DashboardModel)
	narrowView := stripANSI(dm.View())

	// Both should contain all buttons
	for _, label := range []string{"Backup", "Restore", "Settings"} {
		if !strings.Contains(wideView, label) {
			t.Errorf("wide view missing button %q", label)
		}
		if !strings.Contains(narrowView, label) {
			t.Errorf("narrow view missing button %q", label)
		}
	}
}

// --- C: Minimum Size Warning ---

func TestSetupMinTerminalSizeWarning(t *testing.T) {
	// Below minimum — should show warning
	m := NewSetup()
	model, _ := m.Update(tea.WindowSizeMsg{Width: 35, Height: 10})
	sm := model.(SetupModel)
	view := sm.View()
	if !strings.Contains(view, "Terminal too small") {
		t.Error("expected 'Terminal too small' warning in setup view at 35x10")
	}
	expected := fmt.Sprintf("%dx%d", styles.MinTerminalWidth, styles.MinTerminalHeight)
	if !strings.Contains(view, expected) {
		t.Errorf("expected minimum dimensions '%s' in warning", expected)
	}

	// At minimum — should NOT show warning
	model, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 15})
	sm = model.(SetupModel)
	view = sm.View()
	if strings.Contains(view, "Terminal too small") {
		t.Error("should not show warning at 40x15 (exactly minimum)")
	}

	// Above minimum — should NOT show warning
	model, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sm = model.(SetupModel)
	view = sm.View()
	if strings.Contains(view, "Terminal too small") {
		t.Error("should not show warning at 80x24 (above minimum)")
	}
}

func TestSetupMinSizeGuardOnZeroDimensions(t *testing.T) {
	// At init (0x0), guard should prevent showing warning
	m := NewSetup()
	view := m.View()
	if strings.Contains(view, "Terminal too small") {
		t.Error("should not show 'Terminal too small' at initial 0x0 (guard condition)")
	}
}

func TestSetupMinSizeShowsCurrentDimensions(t *testing.T) {
	m := NewSetup()
	model, _ := m.Update(tea.WindowSizeMsg{Width: 30, Height: 12})
	sm := model.(SetupModel)
	view := sm.View()
	if !strings.Contains(view, "30x12") || !strings.Contains(view, fmt.Sprintf("%dx%d", 30, 12)) {
		t.Errorf("warning should show current dimensions 30x12, got:\n%s", view)
	}
}

// --- D: Rapid Resize Sequence ---

func TestRapidResizeSequence_Dashboard(t *testing.T) {
	cfg := testCfg()
	m := NewDashboard(cfg)

	sizes := []tea.WindowSizeMsg{
		{Width: 80, Height: 24},
		{Width: 120, Height: 40},
		{Width: 40, Height: 15},
		{Width: 200, Height: 100},
		{Width: 60, Height: 20},
		{Width: 80, Height: 24},
		{Width: 100, Height: 50},
		{Width: 45, Height: 18},
		{Width: 80, Height: 24},
		{Width: 150, Height: 60},
	}

	var model tea.Model = m
	for _, size := range sizes {
		model, _ = model.Update(size)
	}

	dm := model.(DashboardModel)
	view := dm.View()
	if view == "" {
		t.Error("view should not be empty after rapid resize")
	}

	// Final dimensions should match last size
	if dm.width != 150 {
		t.Errorf("expected final width 150, got %d", dm.width)
	}
	if dm.height != 60 {
		t.Errorf("expected final height 60, got %d", dm.height)
	}
}

func TestRapidResizeSequence_Restore(t *testing.T) {
	cfg := testCfg()
	m := NewRestore(cfg, nil)

	sizes := []tea.WindowSizeMsg{
		{Width: 80, Height: 24},
		{Width: 120, Height: 40},
		{Width: 40, Height: 15},
		{Width: 200, Height: 100},
		{Width: 60, Height: 20},
		{Width: 80, Height: 24},
		{Width: 100, Height: 50},
		{Width: 45, Height: 18},
		{Width: 80, Height: 24},
		{Width: 150, Height: 60},
	}

	var model tea.Model = m
	for _, size := range sizes {
		model, _ = model.Update(size)
	}

	rm := model.(RestoreModel)
	view := rm.View()
	if view == "" {
		t.Error("view should not be empty after rapid resize")
	}
	if rm.width != 150 {
		t.Errorf("expected final width 150, got %d", rm.width)
	}
	if rm.height != 60 {
		t.Errorf("expected final height 60, got %d", rm.height)
	}
}

func TestRapidResizeSequence_Settings(t *testing.T) {
	cfg := testCfg()
	m := NewSettings(cfg)

	sizes := []tea.WindowSizeMsg{
		{Width: 80, Height: 24},
		{Width: 120, Height: 40},
		{Width: 40, Height: 15},
		{Width: 200, Height: 100},
		{Width: 60, Height: 20},
		{Width: 80, Height: 24},
		{Width: 100, Height: 50},
		{Width: 45, Height: 18},
		{Width: 80, Height: 24},
		{Width: 150, Height: 60},
	}

	var model tea.Model = m
	for _, size := range sizes {
		model, _ = model.Update(size)
	}

	sm := model.(SettingsModel)
	view := sm.View()
	if view == "" {
		t.Error("view should not be empty after rapid resize")
	}
}

func TestRapidResizeSequence_Setup(t *testing.T) {
	m := NewSetup()

	sizes := []tea.WindowSizeMsg{
		{Width: 80, Height: 24},
		{Width: 120, Height: 40},
		{Width: 40, Height: 15},
		{Width: 200, Height: 100},
		{Width: 60, Height: 20},
		{Width: 80, Height: 24},
		{Width: 100, Height: 50},
		{Width: 45, Height: 18},
		{Width: 80, Height: 24},
		{Width: 150, Height: 60},
	}

	var model tea.Model = m
	for _, size := range sizes {
		model, _ = model.Update(size)
	}

	sm := model.(SetupModel)
	view := sm.View()
	if view == "" {
		t.Error("view should not be empty after rapid resize")
	}
	if sm.width != 150 {
		t.Errorf("expected final width 150, got %d", sm.width)
	}
	if sm.height != 60 {
		t.Errorf("expected final height 60, got %d", sm.height)
	}
}

// --- E: Setup Filepicker Resize ---

func TestSetupFilepickerResize(t *testing.T) {
	m := NewSetup()

	// Send initial resize
	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	sm := model.(SetupModel)

	view := sm.View()
	if view == "" {
		t.Error("setup view should render after resize")
	}

	// Resize to small
	model, _ = sm.Update(tea.WindowSizeMsg{Width: 40, Height: 15})
	sm = model.(SetupModel)
	view = sm.View()
	if view == "" {
		t.Error("setup view should render at minimum size")
	}

	// Navigate to a file-adding step and resize
	// Advance: Welcome -> BackupDir
	model, _ = sm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	sm = model.(SetupModel)

	// Resize while on BackupDir step
	model, _ = sm.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	sm = model.(SetupModel)
	view = sm.View()
	if view == "" {
		t.Error("setup view should render at BackupDir step after resize")
	}
}

func TestSetupFilepickerResizeInBrowsingMode(t *testing.T) {
	m := NewSetup()

	// Navigate to AddFiles step
	m = navigateToAddFiles(m)
	if m.step != StepAddFiles {
		t.Fatalf("expected StepAddFiles, got %d", m.step)
	}

	// Enter browsing mode
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	sm := model.(SetupModel)
	if !sm.browsing {
		t.Fatal("expected browsing to be true")
	}

	// Resize while browsing — should not panic
	model, _ = sm.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	sm = model.(SetupModel)
	view := sm.View()
	if view == "" {
		t.Error("setup view should render in browsing mode after resize")
	}

	// Resize to small while browsing
	model, _ = sm.Update(tea.WindowSizeMsg{Width: 40, Height: 15})
	sm = model.(SetupModel)
	view = sm.View()
	if view == "" {
		t.Error("setup view should render in browsing mode at small size")
	}
}

// --- F: Extreme Size Edge Cases ---

func TestDashboardZeroSizeDoesNotPanic(t *testing.T) {
	cfg := testCfg()
	m := NewDashboard(cfg)

	// Resize to zero
	model, _ := m.Update(tea.WindowSizeMsg{Width: 0, Height: 0})
	dm := model.(DashboardModel)
	view := dm.View()

	// Should produce some output (even if empty/minimal)
	_ = view
}

func TestRestoreViewportAtExtremeSize(t *testing.T) {
	cfg := testCfg()
	m := NewRestore(cfg, nil)

	// Set to diff preview phase with some content
	m.phase = phaseDiffPreview
	m.currentDiff = "--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-old\n+new"
	m.diffFile = "file.txt"

	// Resize to standard
	model, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	rm := model.(RestoreModel)
	view := rm.View()
	if view == "" {
		t.Error("restore diff preview should render at standard size")
	}

	// Resize to very small — should not panic
	model, _ = rm.Update(tea.WindowSizeMsg{Width: 1, Height: 1})
	rm = model.(RestoreModel)
	_ = rm.View()

	// Resize to ultra-wide
	model, _ = rm.Update(tea.WindowSizeMsg{Width: 300, Height: 10})
	rm = model.(RestoreModel)
	_ = rm.View()
}

func TestSettingsResizePreservesState(t *testing.T) {
	cfg := testCfg()
	m := NewSettings(cfg)

	// Set an initial size
	model, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sm := model.(SettingsModel)

	// Remember state
	initialState := sm.state

	// Resize should not change the settings state
	model, _ = sm.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	sm = model.(SettingsModel)
	if sm.state != initialState {
		t.Errorf("resize changed settings state from %d to %d", initialState, sm.state)
	}
}

func TestLogsResizeUpdatesListDimensions(t *testing.T) {
	cfg := testCfg()
	m := NewLogs(cfg)

	// Set to standard size
	model, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	lm := model.(LogsModel)
	if lm.width != 80 {
		t.Errorf("expected width 80, got %d", lm.width)
	}
	if lm.height != 24 {
		t.Errorf("expected height 24, got %d", lm.height)
	}

	// Resize to large
	model, _ = lm.Update(tea.WindowSizeMsg{Width: 200, Height: 100})
	lm = model.(LogsModel)
	if lm.width != 200 {
		t.Errorf("expected width 200, got %d", lm.width)
	}
	if lm.height != 100 {
		t.Errorf("expected height 100, got %d", lm.height)
	}
}
