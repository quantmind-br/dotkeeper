package backup

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const maxSymlinkDepth = 20

// FileInfo represents metadata about a file to be backed up
type FileInfo struct {
	Path    string      // Original path (may be symlink)
	Size    int64       // File size in bytes
	Mode    fs.FileMode // File permissions
	ModTime int64       // Modification time (Unix timestamp)
}

// CollectFiles collects file information from the given paths.
// It follows symlinks and copies content (doesn't preserve as links).
// Detects and prevents circular symlinks (max depth 20).
// Skips unreadable files with warning.
func CollectFiles(paths []string) ([]FileInfo, error) {
	var files []FileInfo
	visited := make(map[string]bool)

	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		expanded := expandHome(trimmed)
		if err := collectPath(expanded, &files, visited, 0); err != nil {
			// Log error but continue with other paths
			log.Printf("Warning: skipping %s: %v", expanded, err)
		}
	}

	return files, nil
}

func collectPath(path string, files *[]FileInfo, visited map[string]bool, depth int) error {
	symlinkDepth := countSymlinkDepth(path)
	if symlinkDepth > maxSymlinkDepth {
		return fmt.Errorf("max symlink depth exceeded")
	}

	// Get file info (follows symlinks)
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}

	// Get real path to detect circular symlinks
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("eval symlinks failed: %w", err)
	}

	// Check if we've already visited this real path (circular symlink detection)
	if visited[realPath] {
		return fmt.Errorf("circular symlink detected")
	}

	// If it's a directory, collect all files recursively
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("read dir failed: %w", err)
		}

		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			if err := collectPath(entryPath, files, visited, depth); err != nil {
				log.Printf("Warning: skipping %s: %v", entryPath, err)
			}
		}
		return nil
	}

	// Skip special files (sockets, FIFOs, devices)
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	}

	// Check if file is readable
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("file not readable: %w", err)
	}
	file.Close()

	// Mark as visited
	visited[realPath] = true

	// Add file info
	*files = append(*files, FileInfo{
		Path:    path,
		Size:    info.Size(),
		Mode:    info.Mode().Perm(), // Get permission bits only
		ModTime: info.ModTime().Unix(),
	})

	return nil
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

func countSymlinkDepth(path string) int {
	depth := 0
	current := path

	for depth <= maxSymlinkDepth {
		info, err := os.Lstat(current)
		if err != nil {
			return depth
		}

		if info.Mode()&os.ModeSymlink == 0 {
			return depth
		}

		target, err := os.Readlink(current)
		if err != nil {
			return depth
		}

		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(current), target)
		}

		current = target
		depth++
	}

	return depth
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
