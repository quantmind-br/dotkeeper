package restore

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/diogo/dotkeeper/internal/crypto"
)

// Restore restores files from an encrypted backup archive
func Restore(backupPath, password string, opts RestoreOptions) (*RestoreResult, error) {
	result := &RestoreResult{
		RestoredFiles: []string{},
		SkippedFiles:  []string{},
		BackupFiles:   []string{},
		DiffResults:   make(map[string]string),
	}

	// Read and decrypt the backup
	entries, err := decryptAndExtract(backupPath, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt and extract backup: %w", err)
	}

	result.TotalFiles = len(entries)

	// Filter files if SelectedFiles is specified
	if len(opts.SelectedFiles) > 0 {
		entries = filterEntries(entries, opts.SelectedFiles)
	}

	// Process each file
	for _, entry := range entries {
		targetPath := entry.Path
		if opts.TargetDir != "" {
			targetPath = filepath.Join(opts.TargetDir, filepath.Base(entry.Path))
		}

		// Generate diff if requested
		if opts.ShowDiff {
			diffResult, err := GenerateDiff(entry.Content, targetPath)
			if err != nil {
				// Log but continue
				if opts.ProgressCallback != nil {
					opts.ProgressCallback(targetPath, "diff-error")
				}
			} else if diffResult.HasDifference {
				result.DiffResults[targetPath] = diffResult.Diff
				if opts.DiffWriter != nil {
					fmt.Fprintf(opts.DiffWriter, "\n=== %s ===\n%s\n", targetPath, diffResult.Diff)
				}
			}
		}

		// In dry run mode, just report what would happen
		if opts.DryRun {
			if HasConflict(targetPath) {
				result.FilesConflict++
				if opts.ProgressCallback != nil {
					opts.ProgressCallback(targetPath, "would-backup")
				}
			}
			result.SkippedFiles = append(result.SkippedFiles, targetPath)
			result.FilesSkipped++
			continue
		}

		// Handle conflict
		action := ResolveConflict(targetPath, opts)
		var backupCreated string

		switch action {
		case ActionSkip:
			result.SkippedFiles = append(result.SkippedFiles, targetPath)
			result.FilesSkipped++
			if opts.ProgressCallback != nil {
				opts.ProgressCallback(targetPath, "skipped")
			}
			continue

		case ActionBackup:
			backupCreated, err = BackupExisting(targetPath)
			if err != nil {
				return nil, fmt.Errorf("failed to backup %s: %w", targetPath, err)
			}
			if backupCreated != "" {
				result.BackupFiles = append(result.BackupFiles, backupCreated)
				result.FilesConflict++
				if opts.ProgressCallback != nil {
					opts.ProgressCallback(targetPath, "backed-up")
				}
			}

		case ActionOverwrite:
			// No backup needed
		}

		if entry.LinkTarget != "" {
			if err := restoreSymlink(targetPath, entry.LinkTarget); err != nil {
				return nil, fmt.Errorf("failed to restore symlink %s: %w", targetPath, err)
			}
		} else if err := restoreFileAtomic(targetPath, entry.Content, entry.Mode); err != nil {
			return nil, fmt.Errorf("failed to restore %s: %w", targetPath, err)
		}

		result.RestoredFiles = append(result.RestoredFiles, targetPath)
		result.FilesRestored++
		if opts.ProgressCallback != nil {
			opts.ProgressCallback(targetPath, "restored")
		}
	}

	return result, nil
}

// decryptAndExtract decrypts the backup and extracts all files
func decryptAndExtract(backupPath, password string) ([]FileEntry, error) {
	// Read encrypted data
	encryptedData, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	// Read metadata to get salt
	metadataPath := backupPath + ".meta.json"
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata crypto.EncryptionMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Derive key and decrypt
	key := crypto.DeriveKey(password, metadata.Salt)
	decrypted, err := crypto.Decrypt(encryptedData, key)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password?): %w", err)
	}

	// Extract tar.gz archive
	return extractTarGz(decrypted)
}

// extractTarGz extracts files from a tar.gz archive in memory
func extractTarGz(data []byte) ([]FileEntry, error) {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var entries []FileEntry

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar read error: %w", err)
		}

		if header.Typeflag == tar.TypeDir {
			continue
		}

		if header.Typeflag == tar.TypeSymlink {
			entries = append(entries, FileEntry{
				Path:       header.Name,
				Mode:       header.Mode,
				ModTime:    header.ModTime.Unix(),
				LinkTarget: header.Linkname,
			})
			continue
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", header.Name, err)
		}

		entries = append(entries, FileEntry{
			Path:    header.Name,
			Content: content,
			Mode:    header.Mode,
			ModTime: header.ModTime.Unix(),
		})
	}

	return entries, nil
}

func restoreSymlink(path, target string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	os.Remove(path)

	if err := os.Symlink(target, path); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}
	return nil
}

// restoreFileAtomic writes a file atomically using temp file + rename
func restoreFileAtomic(path string, content []byte, mode int64) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create temp file in same directory (for atomic rename)
	tempFile, err := os.CreateTemp(dir, ".dotkeeper-restore-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Clean up temp file on error
	defer func() {
		if tempPath != "" {
			os.Remove(tempPath)
		}
	}()

	// Write content
	if _, err := tempFile.Write(content); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write content: %w", err)
	}

	// Set permissions
	if err := tempFile.Chmod(os.FileMode(mode)); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Close file before rename
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Clear temp path so defer doesn't remove the renamed file
	tempPath = ""

	return nil
}

// filterEntries filters file entries to only include selected files
func filterEntries(entries []FileEntry, selected []string) []FileEntry {
	if len(selected) == 0 {
		return entries
	}

	// Create a set for O(1) lookup
	selectedSet := make(map[string]bool)
	for _, s := range selected {
		selectedSet[s] = true
		// Also add normalized path
		selectedSet[filepath.Clean(s)] = true
	}

	var filtered []FileEntry
	for _, entry := range entries {
		if selectedSet[entry.Path] || selectedSet[filepath.Clean(entry.Path)] {
			filtered = append(filtered, entry)
			continue
		}
		// Check if base name matches
		if selectedSet[filepath.Base(entry.Path)] {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// ListBackupContents returns a list of files in the backup without restoring
func ListBackupContents(backupPath, password string) ([]FileEntry, error) {
	return decryptAndExtract(backupPath, password)
}

// PreviewRestore shows what would be restored without making changes
func PreviewRestore(backupPath, password string, opts RestoreOptions) (*RestoreResult, error) {
	opts.DryRun = true
	opts.ShowDiff = true
	return Restore(backupPath, password, opts)
}

// RestoreFile restores a single file from the backup
func RestoreFile(backupPath, password, filePath string, opts RestoreOptions) error {
	opts.SelectedFiles = []string{filePath}
	_, err := Restore(backupPath, password, opts)
	return err
}

// GetFileDiff returns the diff for a specific file without restoring
func GetFileDiff(backupPath, password, filePath string) (string, error) {
	entries, err := decryptAndExtract(backupPath, password)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.Path == filePath || filepath.Base(entry.Path) == filepath.Base(filePath) {
			if entry.LinkTarget != "" {
				return fmt.Sprintf("symlink → %s", entry.LinkTarget), nil
			}
			result, err := GenerateDiff(entry.Content, filePath)
			if err != nil {
				return "", err
			}
			return result.Diff, nil
		}
	}

	return "", fmt.Errorf("file %s not found in backup", filePath)
}

// ValidateBackup checks if a backup file is valid and decryptable
func ValidateBackup(backupPath, password string) error {
	// Check backup file exists
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup file not found: %w", err)
	}

	// Check metadata file exists
	metadataPath := backupPath + ".meta.json"
	if _, err := os.Stat(metadataPath); err != nil {
		return fmt.Errorf("metadata file not found: %w", err)
	}

	// Try to decrypt and extract (validates password)
	_, err := decryptAndExtract(backupPath, password)
	if err != nil {
		if strings.Contains(err.Error(), "decryption failed") {
			return fmt.Errorf("invalid password or corrupted backup")
		}
		return err
	}

	return nil
}
