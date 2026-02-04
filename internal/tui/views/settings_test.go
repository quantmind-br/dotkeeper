package views

import (
	"testing"

	"github.com/diogo/dotkeeper/internal/config"
)

func TestNewSettings(t *testing.T) {
	cfg := &config.Config{
		BackupDir:     "/tmp/backup",
		GitRemote:     "https://github.com/user/repo",
		Files:         []string{".zshrc", ".vimrc"},
		Folders:       []string{".config/nvim"},
		Schedule:      "daily",
		Notifications: true,
	}

	model := NewSettings(cfg)

	// Verify the model was initialized with the config
	// Since we can't access private fields directly in a different package if we were outside,
	// but here we are in 'views' package so we can test internal state if needed,
	// OR we can test the View output.

	// Let's test the View output contains key information
	viewOutput := model.View()

	expectedStrings := []string{
		"Settings",
		"/tmp/backup",
		"https://github.com/user/repo",
		"2 files",
		"1 folders",
		"daily",
		"true",
	}

	for _, s := range expectedStrings {
		if !contains(viewOutput, s) {
			t.Errorf("Expected view to contain %q, but it didn't. View:\n%s", s, viewOutput)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && search(s, substr)
}

func search(s, substr string) bool {
	// Simple containment check
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
