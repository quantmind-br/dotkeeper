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
			walkErrors := 0
			_ = filepath.WalkDir(expanded, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					walkErrors++
					return nil
				}
				if !d.IsDir() {
					baseName := filepath.Base(path)
					excluded := false
					for _, pat := range exclude {
						if matched, _ := filepath.Match(pat, baseName); matched {
							excluded = true
							break
						}
					}
					if !excluded {
						if info, err := d.Info(); err == nil {
							stat.FileCount++
							stat.Size += info.Size()
							result.TotalFiles++
							result.TotalSize += info.Size()
						} else {
							walkErrors++
						}
					}
				}
				return nil
			})
			if walkErrors > 0 {
				result.BrokenPaths = append(result.BrokenPaths, f)
			}
		}
		result.PathStats = append(result.PathStats, stat)
	}

	return result
}

// FormatSize returns a human-readable size string.
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
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
		_ = filepath.WalkDir(expanded, func(_ string, d os.DirEntry, _ error) error {
			if d != nil && !d.IsDir() {
				if info, err := d.Info(); err == nil {
					count++
					size += info.Size()
				}
			}
			return nil
		})
		return fmt.Sprintf("%d files, %s", count, FormatSize(size))
	}
	return FormatSize(info.Size())
}
