package backup

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/diogo/dotkeeper/internal/pathutil"
)

// FileInfo represents metadata about a file to be backed up
type FileInfo struct {
	Path       string      // Original path (may be symlink)
	Size       int64       // File size in bytes
	Mode       fs.FileMode // File permissions
	ModTime    int64       // Modification time (Unix timestamp)
	LinkTarget string      // Symlink target (empty for regular files)
}

func shouldExclude(path string, baseName string, isDir bool, patterns []string) bool {
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}

		if strings.HasSuffix(pattern, "/") && isDir {
			dirPattern := strings.TrimSuffix(pattern, "/")
			if matched, _ := filepath.Match(dirPattern, baseName); matched {
				return true
			}
		}

		if matched, _ := filepath.Match(pattern, baseName); matched {
			return true
		}

		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
}

// CollectFiles collects file information from the given paths.
// It follows symlinks and copies content (doesn't preserve as links).
// Detects and prevents circular symlinks (max depth 20).
// Skips unreadable files with warning.
// Files matching any excludePatterns are skipped.
func CollectFiles(paths []string, excludePatterns []string) ([]FileInfo, error) {
	var files []FileInfo
	visited := make(map[string]bool)

	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		expanded := pathutil.ExpandHome(trimmed)
		if err := collectPath(expanded, &files, visited, 0, excludePatterns); err != nil {
			// Log error but continue with other paths
			log.Printf("Warning: skipping %s: %v", expanded, err)
		}
	}

	return files, nil
}

func collectPath(path string, files *[]FileInfo, visited map[string]bool, depth int, excludePatterns []string) error {
	linfo, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("lstat failed: %w", err)
	}

	baseName := filepath.Base(path)
	if shouldExclude(path, baseName, linfo.IsDir(), excludePatterns) {
		return nil
	}

	if linfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return fmt.Errorf("readlink failed: %w", err)
		}

		*files = append(*files, FileInfo{
			Path:       path,
			Mode:       linfo.Mode().Perm(),
			ModTime:    linfo.ModTime().Unix(),
			LinkTarget: target,
		})
		return nil
	}

	if linfo.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("read dir failed: %w", err)
		}

		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			if err := collectPath(entryPath, files, visited, depth, excludePatterns); err != nil {
				log.Printf("Warning: skipping %s: %v", entryPath, err)
			}
		}
		return nil
	}

	if !linfo.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("file not readable: %w", err)
	}
	file.Close()

	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		realPath = path
	}
	if visited[realPath] {
		return nil
	}
	visited[realPath] = true

	*files = append(*files, FileInfo{
		Path:    path,
		Size:    linfo.Size(),
		Mode:    linfo.Mode().Perm(),
		ModTime: linfo.ModTime().Unix(),
	})

	return nil
}
