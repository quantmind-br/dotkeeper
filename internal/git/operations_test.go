package git

import (
	"os"
	"path/filepath"
	"strings"
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

func TestSetRemote(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test SetRemote - creates a remote configuration
	err = repo.SetRemote("origin", "https://github.com/test/repo.git")
	if err != nil {
		t.Errorf("SetRemote failed: %v", err)
	}

	// Verify remote was created by checking config
	cfg, err := repo.repo.Config()
	if err != nil {
		t.Fatalf("Failed to get repo config: %v", err)
	}

	remote := cfg.Remotes["origin"]
	if remote == nil {
		t.Fatal("Remote 'origin' was not created")
	}

	if len(remote.URLs) == 0 || remote.URLs[0] != "https://github.com/test/repo.git" {
		t.Errorf("Remote URL not set correctly, got: %v", remote.URLs)
	}
}

func TestSetRemote_UpdateExisting(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create initial remote
	err = repo.SetRemote("origin", "https://github.com/test/repo.git")
	if err != nil {
		t.Fatalf("SetRemote failed: %v", err)
	}

	// Try to SetRemote again with different URL - should fail because remote already exists
	err = repo.SetRemote("origin", "https://github.com/test/updated.git")
	if err == nil {
		t.Error("SetRemote should fail when remote already exists")
	}

	// Verify original URL is still in place
	cfg, err := repo.repo.Config()
	if err != nil {
		t.Fatalf("Failed to get repo config: %v", err)
	}

	remote := cfg.Remotes["origin"]
	if remote == nil {
		t.Fatal("Remote 'origin' was not found")
	}

	if len(remote.URLs) == 0 || remote.URLs[0] != "https://github.com/test/repo.git" {
		t.Errorf("Remote URL should not have changed, got: %v", remote.URLs)
	}
}

func TestPush_NoRemote(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create and commit a file
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := repo.Add("test.txt"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if err := repo.Commit("Test commit"); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Test Push without remote - should fail gracefully
	err = repo.Push()
	if err == nil {
		t.Error("Push should fail when no remote is configured")
	}
	// Error message should indicate the issue
	if err != nil && !strings.Contains(err.Error(), "push") && !strings.Contains(err.Error(), "remote") {
		t.Logf("Push error (expected): %v", err)
	}
}

func TestAdd_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test Add with non-existent file
	err = repo.Add("nonexistent.txt")
	if err == nil {
		t.Error("Add should fail for non-existent file")
	}
}
