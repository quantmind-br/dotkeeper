package views

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
)

func TestNewLogs(t *testing.T) {
	tmpDir := t.TempDir()
	store := history.NewStoreWithPath(tmpDir + "/history.jsonl")
	cfg := &config.Config{}

	m := NewLogs(NewProgramContext(cfg, store))

	if m.ctx.Store != store {
		t.Error("expected store to be set")
	}
	if m.filter != "all" {
		t.Errorf("expected initial filter 'all', got '%s'", m.filter)
	}
}

func TestLogsInit(t *testing.T) {
	tmpDir := t.TempDir()
	store := history.NewStoreWithPath(tmpDir + "/history.jsonl")
	cfg := &config.Config{}
	m := NewLogs(NewProgramContext(cfg, store))

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a command")
	}
}

func TestLogsUpdate_Loaded(t *testing.T) {
	tmpDir := t.TempDir()
	store := history.NewStoreWithPath(tmpDir + "/history.jsonl")

	// Add some entries
	entry := history.HistoryEntry{
		Timestamp: time.Now(),
		Operation: "backup",
		Status:    "success",
		FileCount: 5,
		TotalSize: 1024,
	}
	store.Append(entry)

	cfg := &config.Config{}
	m := NewLogs(NewProgramContext(cfg, store))

	// Simulate loading
	entries, _ := store.Read(100)
	msg := logsLoadedMsg(entries)

	model, _ := m.Update(msg)
	m = model.(LogsModel)

	if len(m.list.Items()) != 1 {
		t.Errorf("expected 1 item, got %d", len(m.list.Items()))
	}

	item := m.list.Items()[0].(logItem)
	if item.entry.Operation != "backup" {
		t.Errorf("expected operation 'backup', got '%s'", item.entry.Operation)
	}
}

func TestLogsUpdate_Empty(t *testing.T) {
	cfg := &config.Config{}
	m := NewLogs(NewProgramContext(cfg, nil)) // No store needed for this specific message test

	msg := logsLoadedMsg([]history.HistoryEntry{})
	model, _ := m.Update(msg)
	m = model.(LogsModel)

	if len(m.list.Items()) != 0 {
		t.Errorf("expected 0 items, got %d", len(m.list.Items()))
	}
	if m.err != "" {
		t.Error("expected no error")
	}
}

func TestLogsFilterCycle(t *testing.T) {
	tmpDir := t.TempDir()
	store := history.NewStoreWithPath(tmpDir + "/history.jsonl")
	cfg := &config.Config{}
	m := NewLogs(NewProgramContext(cfg, store))

	// all -> backup
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	m = model.(LogsModel)
	if m.filter != "backup" {
		t.Errorf("expected filter 'backup', got '%s'", m.filter)
	}
	if cmd == nil {
		t.Error("expected cmd to reload history")
	}

	// backup -> restore
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	m = model.(LogsModel)
	if m.filter != "restore" {
		t.Errorf("expected filter 'restore', got '%s'", m.filter)
	}

	// restore -> all
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	m = model.(LogsModel)
	if m.filter != "all" {
		t.Errorf("expected filter 'all', got '%s'", m.filter)
	}
}

func TestLogsView(t *testing.T) {
	tmpDir := t.TempDir()
	store := history.NewStoreWithPath(tmpDir + "/history.jsonl")
	cfg := &config.Config{}
	m := NewLogs(NewProgramContext(cfg, store))

	// Test Empty View
	view := stripANSI(m.View())
	if !strings.Contains(view, "Operation History [all]") {
		t.Error("expected title in view")
	}
	if !strings.Contains(view, "No operations recorded") {
		t.Error("expected empty state message")
	}

	// StatusHelpText is separate from View()
	helpText := m.StatusHelpText()
	if !strings.Contains(helpText, "f: filter (all)") {
		t.Errorf("expected filter help in StatusHelpText(), got: %s", helpText)
	}
}

func TestLogsView_WithFilter(t *testing.T) {
	tmpDir := t.TempDir()
	store := history.NewStoreWithPath(tmpDir + "/history.jsonl")
	cfg := &config.Config{}

	// Test with different filters
	filters := []string{"all", "backup", "restore"}
	for _, filter := range filters {
		m := NewLogs(NewProgramContext(cfg, store))
		m.filter = filter

		view := stripANSI(m.View())
		expectedTitle := "Operation History [" + filter + "]"
		if !strings.Contains(view, expectedTitle) {
			t.Errorf("expected title '%s' for filter '%s'", expectedTitle, filter)
		}
	}
}

func TestLogsView_WithError(t *testing.T) {
	cfg := &config.Config{}
	m := NewLogs(NewProgramContext(cfg, nil))
	m.err = "failed to load history"

	view := stripANSI(m.View())
	if !strings.Contains(view, "failed to load history") {
		t.Error("expected error message in view")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name string
		b    int64
		want string
	}{
		{name: "zero", b: 0, want: "0 B"},
		{name: "bytes", b: 512, want: "512 B"},
		{name: "kilobytes", b: 1024, want: "1.0 KB"},
		{name: "megabytes", b: 1024 * 1024, want: "1.0 MB"},
		{name: "gigabytes", b: 1024 * 1024 * 1024, want: "1.0 GB"},
		{name: "terabytes", b: 1024 * 1024 * 1024 * 1024, want: "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.b)
			if got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.b, got, tt.want)
			}
		})
	}
}

func TestLogItem_InterfaceMethods(t *testing.T) {
	entry := history.HistoryEntry{
		Timestamp:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Operation:  "backup",
		Status:     "success",
		FileCount:  42,
		TotalSize:  1024 * 1024,
		DurationMs: 5000,
		BackupPath: "/backups/backup.tar.gz.enc",
	}

	item := logItem{entry: entry}

	// Test Title
	title := item.Title()
	if title == "" {
		t.Error("Title should not be empty")
	}
	// Should contain operation and status
	if !strings.Contains(title, "backup") {
		t.Errorf("Title should contain operation, got: %s", title)
	}
	// Success status should have checkmark
	if !strings.Contains(title, "✓") {
		t.Errorf("Success entry should have ✓, got: %s", title)
	}

	// Test Description
	desc := item.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	// Should contain file count
	if !strings.Contains(desc, "42") {
		t.Errorf("Description should contain file count, got: %s", desc)
	}

	// Test FilterValue
	filterValue := item.FilterValue()
	if filterValue != "backup" {
		t.Errorf("FilterValue should be operation type, got: %s", filterValue)
	}
}

func TestLogItem_ErrorStatus(t *testing.T) {
	entry := history.HistoryEntry{
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Operation: "backup",
		Status:    "error",
		Error:     "disk full",
	}

	item := logItem{entry: entry}
	title := item.Title()

	// Error status should have X mark
	if !strings.Contains(title, "✗") {
		t.Errorf("Error entry should have ✗, got: %s", title)
	}
}

func TestLogsHelpBindings(t *testing.T) {
	cfg := &config.Config{}
	m := NewLogs(NewProgramContext(cfg, nil))

	bindings := m.HelpBindings()
	if bindings == nil {
		t.Fatal("HelpBindings should not be nil")
	}

	if len(bindings) == 0 {
		t.Error("HelpBindings should have entries")
	}

	// Check expected bindings exist
	bindingMap := make(map[string]string)
	for _, b := range bindings {
		bindingMap[b.Key] = b.Description
	}

	if bindingMap["f"] != "filter" {
		t.Error("Expected 'f' -> 'filter' binding")
	}
	if bindingMap["r"] != "refresh" {
		t.Error("Expected 'r' -> 'refresh' binding")
	}
}

func TestLogsUpdate_WithRefreshKey(t *testing.T) {
	tmpDir := t.TempDir()
	store := history.NewStoreWithPath(tmpDir + "/history.jsonl")
	cfg := &config.Config{}
	m := NewLogs(NewProgramContext(cfg, store))

	// Press 'r' to refresh
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = model.(LogsModel)

	if cmd == nil {
		t.Error("Expected command to refresh history")
	}
}

func TestLogsUpdate_WithWindowSize(t *testing.T) {
	tmpDir := t.TempDir()
	store := history.NewStoreWithPath(tmpDir + "/history.jsonl")
	cfg := &config.Config{}
	m := NewLogs(NewProgramContext(cfg, store))

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	model, _ := m.Update(msg)
	m = model.(LogsModel)

	if m.ctx.Width != 80 {
		t.Errorf("Expected width 80, got %d", m.ctx.Width)
	}
	if m.ctx.Height != 24 {
		t.Errorf("Expected height 24, got %d", m.ctx.Height)
	}
}

func TestLogsUpdate_WithErrorMessage(t *testing.T) {
	cfg := &config.Config{}
	m := NewLogs(NewProgramContext(cfg, nil))

	errMsg := ErrorMsg{Source: "logs", Err: fmt.Errorf("test error")}
	model, _ := m.Update(errMsg)
	m = model.(LogsModel)

	if m.err == "" {
		t.Error("Expected error message to be set")
	}
	if !strings.Contains(m.err, "test error") {
		t.Errorf("Expected error message to contain 'test error', got: %s", m.err)
	}
}
