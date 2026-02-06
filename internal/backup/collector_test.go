package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectFiles_Basic(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0755); err != nil {
		t.Fatal(err)
	}

	files, err := CollectFiles([]string{file1, file2}, nil)
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	// Verify first file
	if files[0].Path != file1 {
		t.Errorf("Expected path %s, got %s", file1, files[0].Path)
	}
	if files[0].Size != 8 {
		t.Errorf("Expected size 8, got %d", files[0].Size)
	}
	if files[0].Mode != 0644 {
		t.Errorf("Expected mode 0644, got %o", files[0].Mode)
	}
}

func TestCollectFiles_WithSymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	target := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(target, []byte("target content"), 0644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	files, err := CollectFiles([]string{link}, nil)
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	if files[0].Path != link {
		t.Errorf("Expected path %s, got %s", link, files[0].Path)
	}

	if files[0].LinkTarget != target {
		t.Errorf("Expected LinkTarget %s, got %s", target, files[0].LinkTarget)
	}
}

func TestCollectFiles_CircularSymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	link1 := filepath.Join(tmpDir, "link1")
	link2 := filepath.Join(tmpDir, "link2")

	if err := os.Symlink(link2, link1); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(link1, link2); err != nil {
		t.Fatal(err)
	}

	files, err := CollectFiles([]string{link1}, nil)
	if err != nil {
		t.Fatalf("CollectFiles should not fail on circular symlinks: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file (symlink stored as-is), got %d", len(files))
	}

	if files[0].LinkTarget != link2 {
		t.Errorf("Expected LinkTarget %s, got %s", link2, files[0].LinkTarget)
	}
}

func TestCollectFiles_UnreadableFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with no read permissions
	unreadable := filepath.Join(tmpDir, "unreadable.txt")
	if err := os.WriteFile(unreadable, []byte("secret"), 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(unreadable, 0644) // Cleanup

	files, err := CollectFiles([]string{unreadable}, nil)

	// Should not fail, just skip unreadable file
	if err != nil {
		t.Fatalf("CollectFiles should not fail on unreadable files: %v", err)
	}

	// Should return empty list (skipped)
	if len(files) != 0 {
		t.Errorf("Expected 0 files (unreadable file skipped), got %d", len(files))
	}
}

func TestCollectFiles_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory with files
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	file1 := filepath.Join(subDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := CollectFiles([]string{subDir}, nil)
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}

	// Should collect all files in directory
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestCollectFiles_ExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tmpDir, err := os.MkdirTemp(home, "dotkeeper-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir in home: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	relPath := strings.TrimPrefix(filePath, home+string(filepath.Separator))
	if relPath == filePath {
		t.Fatalf("expected file to be under home dir: %s", filePath)
	}

	pathWithTilde := filepath.Join("~", relPath)
	files, err := CollectFiles([]string{pathWithTilde}, nil)
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if files[0].Path != filePath {
		t.Errorf("expected path %s, got %s", filePath, files[0].Path)
	}
}

func TestCollectFiles_SymlinkChain(t *testing.T) {
	tmpDir := t.TempDir()

	target := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(target, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	prev := target
	for i := 0; i < 5; i++ {
		link := filepath.Join(tmpDir, filepath.Base(prev)+"_link")
		if err := os.Symlink(prev, link); err != nil {
			t.Fatal(err)
		}
		prev = link
	}

	files, err := CollectFiles([]string{prev}, nil)
	if err != nil {
		t.Fatalf("CollectFiles should not fail: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file (symlink stored as-is), got %d", len(files))
	}

	if files[0].LinkTarget == "" {
		t.Error("Expected LinkTarget to be set for symlink")
	}
}

func TestCollectFiles_DuplicatePathsViaSymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	realDir := filepath.Join(tmpDir, "real")
	os.MkdirAll(realDir, 0755)
	os.WriteFile(filepath.Join(realDir, "file.txt"), []byte("content"), 0644)

	link1 := filepath.Join(tmpDir, "link1")
	link2 := filepath.Join(tmpDir, "link2")
	os.Symlink(realDir, link1)
	os.Symlink(realDir, link2)

	files, err := CollectFiles([]string{link1, link2}, nil)
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 symlink entries, got %d", len(files))
	}

	for _, f := range files {
		if f.LinkTarget == "" {
			t.Errorf("Expected LinkTarget for %s", f.Path)
		}
	}
}

func TestCollectFilesWithExclusions(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.log"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "c.txt"), []byte("c"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "sub", "d.log"), []byte("d"), 0644)

	files, err := CollectFiles([]string{tmpDir}, []string{"*.log"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Path, ".log") {
			t.Errorf("expected .log files to be excluded, found %s", f.Path)
		}
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestCollectFilesExcludeDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "node_modules", "b.js"), []byte("b"), 0644)

	files, err := CollectFiles([]string{tmpDir}, []string{"node_modules/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

func TestCollectFilesNoExclusions(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("b"), 0644)

	files, err := CollectFiles([]string{tmpDir}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestShouldExclude(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		baseName string
		isDir    bool
		patterns []string
		want     bool
	}{
		{"empty patterns", "/foo/bar.log", "bar.log", false, nil, false},
		{"match base name", "/foo/bar.log", "bar.log", false, []string{"*.log"}, true},
		{"no match", "/foo/bar.txt", "bar.txt", false, []string{"*.log"}, false},
		{"dir pattern on dir", "/foo/node_modules", "node_modules", true, []string{"node_modules/"}, true},
		{"dir pattern on file (should fail if match)", "/foo/node_modules", "node_modules", false, []string{"node_modules/"}, false},
		{"empty pattern skipped", "/foo/bar.log", "bar.log", false, []string{""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldExclude(tt.path, tt.baseName, tt.isDir, tt.patterns)
			if got != tt.want {
				t.Errorf("shouldExclude(%q, %q, %v, %v) = %v, want %v",
					tt.path, tt.baseName, tt.isDir, tt.patterns, got, tt.want)
			}
		})
	}
}
