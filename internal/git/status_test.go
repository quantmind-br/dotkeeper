package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetStatus(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Initially should be clean
	status, err := repo.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.IsClean {
		t.Error("Expected clean status for new repo")
	}

	// Create an untracked file
	testFile := filepath.Join(repoPath, "untracked.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	status, err = repo.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.IsClean {
		t.Error("Expected dirty status with untracked file")
	}

	if len(status.Untracked) != 1 {
		t.Errorf("Expected 1 untracked file, got %d", len(status.Untracked))
	}

	// Add and commit the file
	if err := repo.Add("untracked.txt"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if err := repo.Commit("Add file"); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Should be clean again
	status, err = repo.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.IsClean {
		t.Error("Expected clean status after commit")
	}

	// Modify the file
	if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	status, err = repo.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.IsClean {
		t.Error("Expected dirty status with modified file")
	}
}

func TestHasChanges(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Initially should have no changes
	hasChanges, err := repo.HasChanges()
	if err != nil {
		t.Fatalf("HasChanges failed: %v", err)
	}

	if hasChanges {
		t.Error("Expected no changes for new repo")
	}

	// Create a file
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should have changes now
	hasChanges, err = repo.HasChanges()
	if err != nil {
		t.Fatalf("HasChanges failed: %v", err)
	}

	if !hasChanges {
		t.Error("Expected changes after creating file")
	}

	// Add and commit
	if err := repo.Add("test.txt"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if err := repo.Commit("Add file"); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Should have no changes again
	hasChanges, err = repo.HasChanges()
	if err != nil {
		t.Fatalf("HasChanges failed: %v", err)
	}

	if hasChanges {
		t.Error("Expected no changes after commit")
	}
}
