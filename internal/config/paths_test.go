package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir_UsesXDGConfigHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error: %v", err)
	}

	want := filepath.Join(tmp, "dotkeeper")
	if dir != want {
		t.Fatalf("GetConfigDir() = %q, want %q", dir, want)
	}
}

func TestGetConfigPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error: %v", err)
	}

	want := filepath.Join(tmp, "dotkeeper", "config.yaml")
	if path != want {
		t.Fatalf("GetConfigPath() = %q, want %q", path, want)
	}
}

func TestGetConfigDir_FallsBackToHome(t *testing.T) {
	// Clear XDG_CONFIG_HOME to test fallback to ~/.config
	tmpDir := t.TempDir()
	homeBackup := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		if homeBackup != "" {
			t.Setenv("HOME", homeBackup)
		}
	})
	t.Setenv("XDG_CONFIG_HOME", "")

	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() failed: %v", err)
	}

	want := filepath.Join(tmpDir, ".config", "dotkeeper")
	if dir != want {
		t.Errorf("GetConfigDir() = %q, want %q", dir, want)
	}
}
