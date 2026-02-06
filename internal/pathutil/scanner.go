package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// PathStat holds metadata about a single path.
type PathStat struct {
	Path      string
	Exists    bool
	IsDir     bool
	FileCount int   // Files in directory (0 for files)
	Size      int64 // Total bytes
}

// ScanResult holds aggregated scan results.
type ScanResult struct {
	TotalFiles  int
	TotalSize   int64
	BrokenPaths []string
	PathStats   []PathStat
}

// ScanPaths scans the given files and folders, computing stats.
// exclude patterns are applied using filepath.Match on base names.
func ScanPaths(files, folders, exclude []string) ScanResult {
	result := ScanResult{}

	for _, f := range files {
		expanded := ExpandHome(f)
		stat := PathStat{Path: f}
		info, err := os.Stat(expanded)
		if err != nil {
			stat.Exists = false
			result.BrokenPaths = append(result.BrokenPaths, f)
		} else {
			stat.Exists = true
			stat.IsDir = info.IsDir()
			stat.Size = info.Size()
			if !stat.IsDir {
				result.TotalFiles++
				result.TotalSize += info.Size()
			}
		}
		result.PathStats = append(result.PathStats, stat)
	}

	for _, f := range folders {
		expanded := ExpandHome(f)
		stat := PathStat{Path: f, IsDir: true}
		info, err := os.Stat(expanded)
		if err != nil {
			stat.Exists = false
			result.BrokenPaths = append(result.BrokenPaths, f)
		} else if !info.IsDir() {
			// It's a file but listed as folder
			stat.Exists = true
			stat.IsDir = false
			stat.Size = info.Size()
			result.TotalFiles++
			result.TotalSize += info.Size()
		} else {
			stat.Exists = true
			filepath.Walk(expanded, func(path string, fi os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !fi.IsDir() {
					baseName := filepath.Base(path)
					excluded := false
					for _, pat := range exclude {
						if matched, _ := filepath.Match(pat, baseName); matched {
							excluded = true
							break
						}
					}
					if !excluded {
						stat.FileCount++
						stat.Size += fi.Size()
						result.TotalFiles++
						result.TotalSize += fi.Size()
					}
				}
				return nil
			})
		}
		result.PathStats = append(result.PathStats, stat)
	}

	return result
}

// FormatSize returns a human-readable size string.
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// GetPathDesc returns a brief description of a path's status.
func GetPathDesc(path string) string {
	expanded := ExpandHome(path)
	info, err := os.Stat(expanded)
	if err != nil {
		return "NOT FOUND"
	}
	if info.IsDir() {
		count := 0
		var size int64
		filepath.Walk(expanded, func(_ string, fi os.FileInfo, _ error) error {
			if fi != nil && !fi.IsDir() {
				count++
				size += fi.Size()
			}
			return nil
		})
		return fmt.Sprintf("%d files, %s", count, FormatSize(size))
	}
	return FormatSize(info.Size())
}
