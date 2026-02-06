package views

import (
	"os"
	"strings"
	"testing"
)

func TestRenderStatusBar_StatusOnly(t *testing.T) {
	result := stripANSI(RenderStatusBar(80, "Backup created", "", "n: new | r: refresh"))
	if !strings.Contains(result, "Backup created") {
		t.Error("expected status text")
	}
	if !strings.Contains(result, "n: new | r: refresh") {
		t.Error("expected help text")
	}
}

func TestRenderStatusBar_ErrorOnly(t *testing.T) {
	result := stripANSI(RenderStatusBar(80, "", "Failed", "n: new"))
	if !strings.Contains(result, "Failed") {
		t.Error("expected error text")
	}
	if !strings.Contains(result, "n: new") {
		t.Error("expected help text")
	}
}

func TestRenderStatusBar_ErrorWins(t *testing.T) {
	result := stripANSI(RenderStatusBar(80, "Success", "Error occurred", "help"))
	if !strings.Contains(result, "Error occurred") {
		t.Error("expected error to win")
	}
	if strings.Contains(result, "Success") {
		t.Error("status should not appear when error present")
	}
}

func TestRenderStatusBar_EmptyAll(t *testing.T) {
	result := stripANSI(RenderStatusBar(80, "", "", "help text"))
	if !strings.Contains(result, "help text") {
		t.Error("expected help text")
	}
}

func TestRenderStatusBar_Truncation(t *testing.T) {
	longMsg := strings.Repeat("x", 200)
	result := RenderStatusBar(40, longMsg, "", "help")
	if !strings.Contains(stripANSI(result), "help") {
		t.Error("expected help text")
	}
}

func TestBackupItem_InterfaceMethods(t *testing.T) {
	item := backupItem{
		name: "backup-2025-01-01-120000",
		size: 1024 * 1024,
		date: "2025-01-01 12:00",
	}

	// Test Title
	title := item.Title()
	if title != "backup-2025-01-01-120000" {
		t.Errorf("Title() = %q, want %q", title, "backup-2025-01-01-120000")
	}

	// Test Description
	desc := item.Description()
	expectedDesc := "2025-01-01 12:00 - 1048576 bytes"
	if desc != expectedDesc {
		t.Errorf("Description() = %q, want %q", desc, expectedDesc)
	}

	// Test FilterValue
	filterValue := item.FilterValue()
	if filterValue != "backup-2025-01-01-120000" {
		t.Errorf("FilterValue() = %q, want %q", filterValue, "backup-2025-01-01-120000")
	}
}

func TestPlaceOverlay(t *testing.T) {
	content := "Centered Content"

	// Test various sizes
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "small", width: 20, height: 10},
		{name: "medium", width: 80, height: 24},
		{name: "large", width: 120, height: 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PlaceOverlay(tt.width, tt.height, content)
			if result == "" {
				t.Error("PlaceOverlay should not return empty string")
			}
			// The content should be in the result
			if !strings.Contains(result, content) {
				t.Errorf("Result should contain content %q", content)
			}
		})
	}
}

func TestValidateFolderPath(t *testing.T) {
	// Test with a directory
	t.Run("valid directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		result, err := ValidateFolderPath(tmpDir)
		if err != nil {
			t.Errorf("ValidateFolderPath(%q) should succeed, got: %v", tmpDir, err)
		}
		if result != tmpDir {
			t.Errorf("ValidateFolderPath(%q) = %q, want %q", tmpDir, result, tmpDir)
		}
	})

	// Test with a file instead of directory
	t.Run("file instead of directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create a file
		filePath := tmpDir + "/testfile.txt"
		_ = os.WriteFile(filePath, []byte("test"), 0644)

		_, validateErr := ValidateFolderPath(filePath)
		if validateErr == nil {
			t.Error("ValidateFolderPath should fail when path is a file")
		}
	})

	// Test with non-existent path
	t.Run("non-existent path", func(t *testing.T) {
		_, err := ValidateFolderPath("/nonexistent/path/that/does/not/exist")
		if err == nil {
			t.Error("ValidateFolderPath should fail for non-existent path")
		}
	})
}
