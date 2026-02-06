package pathutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create structure:
	// tmpDir/
	//   file1.txt (10 bytes)
	//   folder1/
	//     file2.txt (20 bytes)
	//     file3.log (30 bytes)
	//   folder2/ (empty)
	//   file4.tmp (excluded)

	file1 := filepath.Join(tmpDir, "file1.txt")
	os.WriteFile(file1, []byte("0123456789"), 0644)

	folder1 := filepath.Join(tmpDir, "folder1")
	os.Mkdir(folder1, 0755)
	file2 := filepath.Join(folder1, "file2.txt")
	os.WriteFile(file2, make([]byte, 20), 0644)
	file3 := filepath.Join(folder1, "file3.log")
	os.WriteFile(file3, make([]byte, 30), 0644)

	folder2 := filepath.Join(tmpDir, "folder2")
	os.Mkdir(folder2, 0755)

	file4 := filepath.Join(folder1, "file4.tmp")
	os.WriteFile(file4, make([]byte, 100), 0644)

	files := []string{file1, filepath.Join(tmpDir, "nonexistent.txt")}
	folders := []string{folder1, folder2, filepath.Join(tmpDir, "nonexistent_folder")}
	exclude := []string{"*.tmp"}

	result := ScanPaths(files, folders, exclude)

	// file1(10) + file2(20) + file3(30) = 60 bytes
	expectedSize := int64(60)
	expectedFiles := 3
	expectedBroken := 2 // nonexistent.txt + nonexistent_folder

	if result.TotalSize != expectedSize {
		t.Errorf("TotalSize: want %d, got %d", expectedSize, result.TotalSize)
	}
	if result.TotalFiles != expectedFiles {
		t.Errorf("TotalFiles: want %d, got %d", expectedFiles, result.TotalFiles)
	}
	if len(result.BrokenPaths) != expectedBroken {
		t.Errorf("BrokenPaths: want %d, got %d", expectedBroken, len(result.BrokenPaths))
	}

	// Verify stats for folder1
	foundFolder1 := false
	for _, stat := range result.PathStats {
		if stat.Path == folder1 {
			foundFolder1 = true
			if stat.FileCount != 2 {
				t.Errorf("Folder1 FileCount: want 2, got %d", stat.FileCount)
			}
			if stat.Size != 50 {
				t.Errorf("Folder1 Size: want 50, got %d", stat.Size)
			}
		}
	}
	if !foundFolder1 {
		t.Error("Folder1 stats not found")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		got := FormatSize(tt.input)
		if got != tt.expected {
			t.Errorf("FormatSize(%d): want %s, got %s", tt.input, tt.expected, got)
		}
	}
}

func TestGetPathDesc(t *testing.T) {
	tmpDir := t.TempDir()

	// File case
	f := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(f, make([]byte, 100), 0644)
	if desc := GetPathDesc(f); desc != "100 B" {
		t.Errorf("GetPathDesc(file): want '100 B', got '%s'", desc)
	}

	// Folder case
	d := filepath.Join(tmpDir, "dir")
	os.Mkdir(d, 0755)
	os.WriteFile(filepath.Join(d, "a"), make([]byte, 10), 0644)
	os.WriteFile(filepath.Join(d, "b"), make([]byte, 20), 0644)

	// Expect "2 files, 30 B"
	expected := "2 files, 30 B"
	if desc := GetPathDesc(d); desc != expected {
		t.Errorf("GetPathDesc(dir): want '%s', got '%s'", expected, desc)
	}

	// Non-existent
	if desc := GetPathDesc(filepath.Join(tmpDir, "nope")); desc != "NOT FOUND" {
		t.Errorf("GetPathDesc(missing): want 'NOT FOUND', got '%s'", desc)
	}
}
