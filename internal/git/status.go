package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

// Status represents the repository status
type Status struct {
	IsClean   bool
	Modified  []string
	Added     []string
	Deleted   []string
	Untracked []string
}

// GetStatus returns the current repository status
func (r *Repository) GetStatus() (*Status, error) {
	w, err := r.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	gitStatus, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	status := &Status{
		IsClean: gitStatus.IsClean(),
	}

	for file, s := range gitStatus {
		// Check for untracked files first
		if s.Worktree == git.Untracked {
			status.Untracked = append(status.Untracked, file)
			continue
		}

		// Check staging area
		switch s.Staging {
		case git.Modified:
			status.Modified = append(status.Modified, file)
		case git.Added:
			status.Added = append(status.Added, file)
		case git.Deleted:
			status.Deleted = append(status.Deleted, file)
		}
	}

	return status, nil
}

// HasChanges returns true if there are uncommitted changes
func (r *Repository) HasChanges() (bool, error) {
	status, err := r.GetStatus()
	if err != nil {
		return false, err
	}
	return !status.IsClean, nil
}
