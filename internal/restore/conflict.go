package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ConflictAction defines what action to take when a file conflict occurs
type ConflictAction int

const (
	// ActionSkip skips restoring the file
	ActionSkip ConflictAction = iota
	// ActionOverwrite overwrites without backup
	ActionOverwrite
	// ActionBackup creates .bak and then overwrites
	ActionBackup
)

// BackupExisting renames the existing file to .bak before restore
// Returns the backup path if created, empty string if file didn't exist
func BackupExisting(path string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", nil // No file to backup
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	backupPath := filepath.Join(dir, fmt.Sprintf("%s.bak.%s", base, timestamp))

	// Ensure unique backup name
	backupPath = ensureUniquePath(backupPath)

	// Rename existing file to backup
	if err := os.Rename(path, backupPath); err != nil {
		return "", fmt.Errorf("failed to backup existing file: %w", err)
	}

	return backupPath, nil
}

// ensureUniquePath adds a suffix if the path already exists
func ensureUniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	// File exists, add numeric suffix
	for i := 1; i < 1000; i++ {
		newPath := fmt.Sprintf("%s.%d", path, i)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}

	// Fallback: use nanoseconds
	return fmt.Sprintf("%s.%d", path, time.Now().UnixNano())
}

// ResolveConflict determines the action to take for a file conflict
func ResolveConflict(path string, opts RestoreOptions) ConflictAction {
	// Check if file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ActionOverwrite // No conflict
	}

	if opts.Force {
		return ActionOverwrite
	}

	return ActionBackup
}

// HasConflict checks if restoring would overwrite an existing file
func HasConflict(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CleanupBackups removes .bak files for a given original path
func CleanupBackups(originalPath string) error {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	pattern := filepath.Join(dir, base+".bak.*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find backups: %w", err)
	}

	for _, match := range matches {
		if err := os.Remove(match); err != nil {
			return fmt.Errorf("failed to remove backup %s: %w", match, err)
		}
	}

	return nil
}

// RestoreFromBackup restores a file from its .bak version
func RestoreFromBackup(originalPath, backupPath string) error {
	// Remove the current file if it exists
	if _, err := os.Stat(originalPath); err == nil {
		if err := os.Remove(originalPath); err != nil {
			return fmt.Errorf("failed to remove current file: %w", err)
		}
	}

	// Rename backup back to original
	if err := os.Rename(backupPath, originalPath); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	return nil
}

// ListBackups returns all .bak files for a given original path
func ListBackups(originalPath string) ([]string, error) {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	pattern := filepath.Join(dir, base+".bak.*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find backups: %w", err)
	}

	return matches, nil
}
