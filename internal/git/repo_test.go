package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if repo == nil {
		t.Fatal("Expected repository, got nil")
	}

	if repo.Path() != repoPath {
		t.Errorf("Expected path %s, got %s", repoPath, repo.Path())
	}

	// Verify .git directory exists
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory was not created")
	}
}

func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// First init a repo
	_, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Now open it
	repo, err := Open(repoPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if repo == nil {
		t.Fatal("Expected repository, got nil")
	}

	if repo.Path() != repoPath {
		t.Errorf("Expected path %s, got %s", repoPath, repo.Path())
	}
}

func TestOpenNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "nonexistent")

	_, err := Open(repoPath)
	if err == nil {
		t.Error("Expected error when opening non-existent repo, got nil")
	}
}

func TestInitOrOpen(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// First call should init
	repo1, err := InitOrOpen(repoPath)
	if err != nil {
		t.Fatalf("InitOrOpen (init) failed: %v", err)
	}

	if repo1 == nil {
		t.Fatal("Expected repository, got nil")
	}

	// Second call should open
	repo2, err := InitOrOpen(repoPath)
	if err != nil {
		t.Fatalf("InitOrOpen (open) failed: %v", err)
	}

	if repo2 == nil {
		t.Fatal("Expected repository, got nil")
	}

	if repo1.Path() != repo2.Path() {
		t.Error("Expected same path for both calls")
	}
}
