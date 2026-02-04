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

	files, err := CollectFiles([]string{file1, file2})
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

	// Create target file
	target := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(target, []byte("target content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create symlink
	link := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	files, err := CollectFiles([]string{link})
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	// Should follow symlink and get target content
	if files[0].Size != 14 {
		t.Errorf("Expected size 14 (target content), got %d", files[0].Size)
	}

	// Path should be the link path, not target
	if files[0].Path != link {
		t.Errorf("Expected path %s, got %s", link, files[0].Path)
	}
}

func TestCollectFiles_CircularSymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create circular symlinks
	link1 := filepath.Join(tmpDir, "link1")
	link2 := filepath.Join(tmpDir, "link2")

	if err := os.Symlink(link2, link1); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(link1, link2); err != nil {
		t.Fatal(err)
	}

	files, err := CollectFiles([]string{link1})

	// Should detect circular symlink and skip with warning
	if err != nil {
		t.Fatalf("CollectFiles should not fail on circular symlinks: %v", err)
	}

	// Should return empty list (skipped)
	if len(files) != 0 {
		t.Errorf("Expected 0 files (circular symlink skipped), got %d", len(files))
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

	files, err := CollectFiles([]string{unreadable})

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

	files, err := CollectFiles([]string{subDir})
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
	files, err := CollectFiles([]string{pathWithTilde})
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

func TestCollectFiles_MaxDepth(t *testing.T) {
	tmpDir := t.TempDir()

	// Create deep symlink chain (depth > 20)
	prev := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(prev, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 25; i++ {
		link := filepath.Join(tmpDir, filepath.Base(prev)+"_link")
		if err := os.Symlink(prev, link); err != nil {
			t.Fatal(err)
		}
		prev = link
	}

	files, err := CollectFiles([]string{prev})

	// Should detect max depth exceeded and skip
	if err != nil {
		t.Fatalf("CollectFiles should not fail on max depth: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files (max depth exceeded), got %d", len(files))
	}
}
