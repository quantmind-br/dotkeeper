package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// Repository wraps a git repository
type Repository struct {
	repo *git.Repository
	path string
}

// Init initializes a new git repository at the given path
func Init(path string) (*Repository, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	repo, err := git.PlainInit(path, false)
	if err != nil {
		return nil, fmt.Errorf("failed to init repository: %w", err)
	}

	return &Repository{repo: repo, path: path}, nil
}

// Open opens an existing git repository at the given path
func Open(path string) (*Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return &Repository{repo: repo, path: path}, nil
}

// InitOrOpen initializes a new repository or opens an existing one
func InitOrOpen(path string) (*Repository, error) {
	if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
		return Init(path)
	}
	return Open(path)
}

// Path returns the repository path
func (r *Repository) Path() string {
	return r.path
}
