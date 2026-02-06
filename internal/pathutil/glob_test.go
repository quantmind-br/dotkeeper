package pathutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGlobPattern(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"*.txt", true},
		{"foo?bar", true},
		{"[abc]", true},
		{"normal.txt", false},
		{"/path/to/file", false},
		{"~/.config/nvim", false},
		{"dir/**/*.lua", true},
	}
	for _, tt := range tests {
		if got := IsGlobPattern(tt.input); got != tt.want {
			t.Errorf("IsGlobPattern(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestResolveGlob_StandardPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt", "c.log", "d.txt"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	results, err := ResolveGlob(filepath.Join(tmpDir, "*.txt"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 .txt files, got %d: %v", len(results), results)
	}
}

func TestResolveGlob_QuestionMark(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"config-1.yaml", "config-2.yaml", "config-AB.yaml"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	results, err := ResolveGlob(filepath.Join(tmpDir, "config-?.yaml"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 single-char matches, got %d: %v", len(results), results)
	}
}

func TestResolveGlob_Recursive(t *testing.T) {
	tmpDir := t.TempDir()
	sub := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "root.lua"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "nested.lua"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := ResolveGlob(filepath.Join(tmpDir, "**/*.lua"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 .lua files recursively, got %d: %v", len(results), results)
	}
}

func TestResolveGlob_Exclude(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt", "c.log"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	results, err := ResolveGlob(filepath.Join(tmpDir, "*"), []string{"*.log"})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if filepath.Ext(r) == ".log" {
			t.Errorf("excluded .log file present: %s", r)
		}
	}
	if len(results) != 2 {
		t.Errorf("expected 2 after excluding .log, got %d: %v", len(results), results)
	}
}

func TestResolveGlob_Cap(t *testing.T) {
	tmpDir := t.TempDir()
	for i := 0; i < MaxGlobResults+1; i++ {
		if err := os.WriteFile(filepath.Join(tmpDir, filepath.Base(t.TempDir())+".txt"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	_, err := ResolveGlob(filepath.Join(tmpDir, "*.txt"), nil)
	if err == nil {
		t.Error("expected error for exceeding MaxGlobResults")
	}
}

func TestResolveGlob_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()

	results, err := ResolveGlob(filepath.Join(tmpDir, "*.nonexistent"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 matches, got %d", len(results))
	}
}

func TestResolveGlob_InvalidPattern(t *testing.T) {
	_, err := ResolveGlob("[invalid", nil)
	if err == nil {
		t.Error("expected error for invalid pattern")
	}
}
