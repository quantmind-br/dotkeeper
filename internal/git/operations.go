package git

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Add adds files to the staging area
func (r *Repository) Add(paths ...string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	for _, path := range paths {
		if _, err := w.Add(path); err != nil {
			return fmt.Errorf("failed to add %s: %w", path, err)
		}
	}

	return nil
}

// AddAll adds all changes to the staging area
func (r *Repository) AddAll() error {
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = w.Add(".")
	if err != nil {
		return fmt.Errorf("failed to add all: %w", err)
	}

	return nil
}

// Commit creates a new commit with the given message
func (r *Repository) Commit(message string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "dotkeeper",
			Email: "dotkeeper@localhost",
			When:  time.Now(),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// Push pushes commits to the remote
func (r *Repository) Push() error {
	err := r.repo.Push(&git.PushOptions{})
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	return nil
}

// SetRemote sets the remote URL
func (r *Repository) SetRemote(name, url string) error {
	_, err := r.repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
	if err != nil {
		return fmt.Errorf("failed to create remote: %w", err)
	}
	return nil
}
