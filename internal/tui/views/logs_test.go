package views

import (
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

	m := NewLogs(cfg, store)

	if m.store != store {
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
	m := NewLogs(cfg, store)

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
	m := NewLogs(cfg, store)

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
	m := NewLogs(cfg) // No store needed for this specific message test

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
	m := NewLogs(cfg, store)

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
	m := NewLogs(cfg, store)

	// Test Empty View
	view := stripANSI(m.View())
	if !strings.Contains(view, "Operation History [all]") {
		t.Error("expected title in view")
	}
	if !strings.Contains(view, "No operations recorded") {
		t.Error("expected empty state message")
	}
	if !strings.Contains(view, "f: filter (all)") {
		t.Errorf("expected filter help text in view, got: %s", view)
	}
}
