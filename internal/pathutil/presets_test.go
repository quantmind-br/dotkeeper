package pathutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDotfiles(t *testing.T) {
	// Create a temporary home directory
	tempHome := t.TempDir()

	// Create some dummy dotfiles
	createFile(t, tempHome, ".bashrc", "alias foo=bar")
	createFile(t, tempHome, ".vimrc", "set number")
	createFile(t, tempHome, ".ssh/config", "Host *") // Nested file

	// Create some dummy folders
	createDir(t, tempHome, ".config/nvim")
	createFile(t, tempHome, ".config/nvim/init.lua", "print('hello')")
	createFile(t, tempHome, ".config/nvim/plugin.lua", "require('plugin')")

	createDir(t, tempHome, ".config/kitty")
	createFile(t, tempHome, ".config/kitty/kitty.conf", "font_size 12")

	// Set SHELL environment variable for testing auto-selection
	originalShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", originalShell)
	os.Setenv("SHELL", "/bin/bash")

	// Run detection
	files, folders := DetectDotfiles(tempHome)

	// Verify files
	if len(files) != 3 {
		t.Errorf("Expected 3 detected files, got %d", len(files))
	}

	foundBashrc := false
	for _, f := range files {
		if f.Path == "~/.bashrc" {
			foundBashrc = true
			if !f.Selected {
				t.Error("Expected ~/.bashrc to be selected due to SHELL=/bin/bash")
			}
			if f.Size == 0 {
				t.Error("Expected ~/.bashrc size > 0")
			}
		}
		if f.Path == "~/.zshrc" {
			t.Error("Did not expect ~/.zshrc to be detected")
		}
	}
	if !foundBashrc {
		t.Error("Expected ~/.bashrc to be detected")
	}

	// Verify folders
	if len(folders) != 2 {
		t.Errorf("Expected 2 detected folders, got %d", len(folders))
	}

	foundNvim := false
	for _, f := range folders {
		if f.Path == "~/.config/nvim" {
			foundNvim = true
			if !f.IsDir {
				t.Error("Expected ~/.config/nvim to be a directory")
			}
			if f.FileCount != 2 {
				t.Errorf("Expected ~/.config/nvim to have 2 files, got %d", f.FileCount)
			}
		}
	}
	if !foundNvim {
		t.Error("Expected ~/.config/nvim to be detected")
	}
}

func TestDetectDotfiles_ZshSelection(t *testing.T) {
	tempHome := t.TempDir()
	createFile(t, tempHome, ".zshrc", "# zsh config")

	originalShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", originalShell)
	os.Setenv("SHELL", "/usr/bin/zsh")

	files, _ := DetectDotfiles(tempHome)

	foundZshrc := false
	for _, f := range files {
		if f.Path == "~/.zshrc" {
			foundZshrc = true
			if !f.Selected {
				t.Error("Expected ~/.zshrc to be selected due to SHELL=/usr/bin/zsh")
			}
		}
	}
	if !foundZshrc {
		t.Error("Expected ~/.zshrc to be detected")
	}
}

func createFile(t *testing.T, dir, path, content string) {
	fullPath := filepath.Join(dir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
}

func createDir(t *testing.T, dir, path string) {
	fullPath := filepath.Join(dir, path)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
}
