package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
)

func TestDashboard(t *testing.T) {
	cfg := &config.Config{
		BackupDir: "/tmp/backup",
		Files:     []string{"file1", "file2"},
	}

	m := NewDashboard(cfg)
	m.width = 80
	m.height = 24
	m.fileCount = 2

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil command")
	}

	view := stripANSI(m.View())

	if !strings.Contains(view, "Last Backup") {
		t.Error("View() does not contain 'Last Backup'")
	}
	if !strings.Contains(view, "Files Tracked") {
		t.Error("View() does not contain 'Files Tracked'")
	}

	if !strings.Contains(view, "Backup") {
		t.Error("View() missing 'Backup' button")
	}
	if !strings.Contains(view, "Restore") {
		t.Error("View() missing 'Restore' button")
	}
	if !strings.Contains(view, "Settings") {
		t.Error("View() missing 'Settings' button")
	}

	m.width = 40
	viewNarrow := stripANSI(m.View())

	if !strings.Contains(viewNarrow, "Last Backup") {
		t.Error("Narrow View() missing 'Last Backup'")
	}
	if !strings.Contains(viewNarrow, "Files Tracked") {
		t.Error("Narrow View() missing 'Files Tracked'")
	}
}

func TestDashboardArrowSelectionAndEnter(t *testing.T) {
	cfg := &config.Config{BackupDir: "/tmp/backup"}
	m := NewDashboard(cfg)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(DashboardModel)
	if m.selected != 1 {
		t.Fatalf("selected after right: got %d, want %d", m.selected, 1)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(DashboardModel)
	if m.selected != 0 {
		t.Fatalf("selected after left: got %d, want %d", m.selected, 0)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(DashboardModel)
	if m.selected != 2 {
		t.Fatalf("selected after up wrap: got %d, want %d", m.selected, 2)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(DashboardModel)
	if cmd == nil {
		t.Fatal("enter should return a command")
	}

	msg := cmd()
	nav, ok := msg.(DashboardNavigateMsg)
	if !ok {
		t.Fatalf("enter command returned %T, want DashboardNavigateMsg", msg)
	}
	if nav.Target != "settings" {
		t.Fatalf("enter target: got %q, want %q", nav.Target, "settings")
	}
}
