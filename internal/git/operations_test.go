package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitOperations(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Init repository
	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test Add
	if err := repo.Add("test.txt"); err != nil {
		t.Errorf("Add failed: %v", err)
	}

	// Test Commit
	if err := repo.Commit("Initial commit"); err != nil {
		t.Errorf("Commit failed: %v", err)
	}

	// Verify commit was created
	status, err := repo.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.IsClean {
		t.Error("Expected clean status after commit")
	}
}

func TestAddAll(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create multiple test files
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, file := range files {
		path := filepath.Join(repoPath, file)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test AddAll
	if err := repo.AddAll(); err != nil {
		t.Errorf("AddAll failed: %v", err)
	}

	// Commit to verify all were added
	if err := repo.Commit("Add all files"); err != nil {
		t.Errorf("Commit failed: %v", err)
	}

	status, err := repo.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.IsClean {
		t.Error("Expected clean status after commit")
	}
}

func TestAddMultiplePaths(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create test files
	file1 := filepath.Join(repoPath, "file1.txt")
	file2 := filepath.Join(repoPath, "file2.txt")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test Add with multiple paths
	if err := repo.Add("file1.txt", "file2.txt"); err != nil {
		t.Errorf("Add multiple paths failed: %v", err)
	}

	// Commit to verify both were added
	if err := repo.Commit("Add multiple files"); err != nil {
		t.Errorf("Commit failed: %v", err)
	}

	status, err := repo.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.IsClean {
		t.Error("Expected clean status after commit")
	}
}
