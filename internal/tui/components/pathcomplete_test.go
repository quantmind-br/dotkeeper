package components

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPathCompleter_CommonPrefix(t *testing.T) {
	tests := []struct {
		name       string
		candidates []CompletionCandidate
		expected   string
	}{
		{
			name: "single match",
			candidates: []CompletionCandidate{
				{Path: "/foo/bar"},
			},
			expected: "/foo/bar",
		},
		{
			name: "common prefix",
			candidates: []CompletionCandidate{
				{Path: "/foo/bar"},
				{Path: "/foo/baz"},
			},
			expected: "/foo/ba",
		},
		{
			name: "no common prefix",
			candidates: []CompletionCandidate{
				{Path: "/foo/bar"},
				{Path: "/baz/qux"},
			},
			expected: "/",
		},
		{
			name:       "empty candidates",
			candidates: []CompletionCandidate{},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := commonPrefix(tt.candidates)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestPathCompleter_Update(t *testing.T) {
	// Setup temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "pathcomplete_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create dummy files
	if err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	pc := NewPathCompleter()

	// Test 1: Trigger completion (Tab)
	pc.Input.SetValue(tmpDir + "/")
	// Simulate Tab key press
	pc, _ = pc.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Since completeCmd returns a tea.Cmd (function), we can't easily execute it here without the tea runtime,
	// but we can manually invoke the completion logic if we exposed it, or we can test the state transitions
	// assuming the Cmd returns the correct message.
	//
	// However, we can simulate the RESULT of the command being fed back into Update.

	// Construct a fake completion result
	candidates := []CompletionCandidate{
		{Path: filepath.Join(tmpDir, "file1.txt"), IsDir: false},
		{Path: filepath.Join(tmpDir, "file2.txt"), IsDir: false},
		{Path: filepath.Join(tmpDir, "subdir"), IsDir: true},
	}
	msg := CompletionResultMsg{
		Candidates: candidates,
		Prefix:     tmpDir + "/",
	}

	// Update with result
	pc, _ = pc.Update(msg)

	// Check if candidates are shown
	if !pc.showCandidates {
		t.Error("expected showCandidates to be true")
	}
	if len(pc.candidates) != 3 {
		t.Errorf("expected 3 candidates, got %d", len(pc.candidates))
	}
	// Common prefix should be set
	// Common prefix of .../file1.txt, .../file2.txt, .../subdir is .../ (tmpDir)
	// Actually common prefix of file1 and file2 is file, but subdir breaks it?
	// The paths are absolute. common prefix of /tmp/file1 and /tmp/file2 is /tmp/file.
	// But /tmp/subdir makes it /tmp/.
	// Wait, commonPrefix implementation checks strings.HasPrefix.

	// Test 2: Cycle through candidates with Tab
	// Initial selection is 0
	if pc.selectedCandidate != 0 {
		t.Errorf("expected selectedCandidate 0, got %d", pc.selectedCandidate)
	}

	// Hit Tab to cycle
	pc, _ = pc.Update(tea.KeyMsg{Type: tea.KeyTab})
	if pc.selectedCandidate != 1 {
		t.Errorf("expected selectedCandidate 1, got %d", pc.selectedCandidate)
	}

	// Check Input value update on cycle
	// Should match candidate 1 path
	expectedVal := candidates[1].Path
	if pc.Input.Value() != expectedVal {
		t.Errorf("expected input value %q, got %q", expectedVal, pc.Input.Value())
	}

	// Test 3: Auto-complete single match
	singleMsg := CompletionResultMsg{
		Candidates: []CompletionCandidate{{Path: "/unique/match", IsDir: false}},
		Prefix:     "/uni",
	}
	pc, _ = pc.Update(singleMsg)
	if pc.showCandidates {
		t.Error("expected showCandidates to be false for single match")
	}
	if pc.Input.Value() != "/unique/match" {
		t.Errorf("expected auto-complete to '/unique/match', got %q", pc.Input.Value())
	}
}
