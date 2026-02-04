package restore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDiff_IdenticalContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("line1\nline2\nline3\n")

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := GenerateDiff(content, filePath)
	if err != nil {
		t.Fatalf("GenerateDiff failed: %v", err)
	}

	if result.HasDifference {
		t.Error("Expected no difference for identical content")
	}

	if result.Diff != "" {
		t.Errorf("Expected empty diff, got: %s", result.Diff)
	}
}

func TestGenerateDiff_DifferentContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	currentContent := []byte("line1\nline2\nline3\n")
	backupContent := []byte("line1\nmodified\nline3\n")

	if err := os.WriteFile(filePath, currentContent, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := GenerateDiff(backupContent, filePath)
	if err != nil {
		t.Fatalf("GenerateDiff failed: %v", err)
	}

	if !result.HasDifference {
		t.Error("Expected difference to be detected")
	}

	if result.Diff == "" {
		t.Error("Expected non-empty diff")
	}

	// Verify diff contains expected markers
	if !strings.Contains(result.Diff, "---") {
		t.Error("Expected diff to contain '---' marker")
	}
	if !strings.Contains(result.Diff, "+++") {
		t.Error("Expected diff to contain '+++' marker")
	}
}

func TestGenerateDiff_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.txt")
	backupContent := []byte("new file content\n")

	result, err := GenerateDiff(backupContent, filePath)
	if err != nil {
		t.Fatalf("GenerateDiff failed: %v", err)
	}

	if !result.HasDifference {
		t.Error("Expected difference for new file")
	}

	// Should show as new file
	if !strings.Contains(result.Diff, "/dev/null") {
		t.Error("Expected diff to show as new file")
	}
}

func TestGenerateDiffFromFiles(t *testing.T) {
	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "backup.txt")
	currentPath := filepath.Join(tmpDir, "current.txt")

	if err := os.WriteFile(backupPath, []byte("backup content\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(currentPath, []byte("current content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := GenerateDiffFromFiles(backupPath, currentPath)
	if err != nil {
		t.Fatalf("GenerateDiffFromFiles failed: %v", err)
	}

	if !result.HasDifference {
		t.Error("Expected difference between files")
	}
}

func TestIsBinaryFile(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		isBinary bool
	}{
		{
			name:     "text file",
			content:  []byte("Hello, World!\nThis is text."),
			isBinary: false,
		},
		{
			name:     "binary with null bytes",
			content:  []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00},
			isBinary: true,
		},
		{
			name:     "empty content",
			content:  []byte{},
			isBinary: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBinaryFile(tt.content)
			if result != tt.isBinary {
				t.Errorf("IsBinaryFile() = %v, want %v", result, tt.isBinary)
			}
		})
	}
}

func TestFormatDiffStats(t *testing.T) {
	diff := `--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 line1
-old line
+new line
+another new line
 line3`

	added, removed := FormatDiffStats(diff)

	if added != 2 {
		t.Errorf("Expected 2 added lines, got %d", added)
	}

	if removed != 1 {
		t.Errorf("Expected 1 removed line, got %d", removed)
	}
}

func TestComputeLCS(t *testing.T) {
	a := []string{"A", "B", "C", "D", "E"}
	b := []string{"A", "C", "E"}

	lcs := computeLCS(a, b)

	expected := []string{"A", "C", "E"}
	if len(lcs) != len(expected) {
		t.Errorf("LCS length = %d, want %d", len(lcs), len(expected))
	}

	for i, v := range lcs {
		if v != expected[i] {
			t.Errorf("LCS[%d] = %s, want %s", i, v, expected[i])
		}
	}
}

func TestComputeLCS_Empty(t *testing.T) {
	lcs := computeLCS([]string{}, []string{"A", "B"})
	if len(lcs) != 0 {
		t.Errorf("Expected empty LCS, got %v", lcs)
	}

	lcs = computeLCS([]string{"A", "B"}, []string{})
	if len(lcs) != 0 {
		t.Errorf("Expected empty LCS, got %v", lcs)
	}
}
