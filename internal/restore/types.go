package restore

import "io"

// RestoreOptions configures the restore operation
type RestoreOptions struct {
	// DryRun only shows what would be restored without making changes
	DryRun bool

	// ShowDiff displays diff between backup and current files
	ShowDiff bool

	// Force overwrites files without .bak backup
	Force bool

	// TargetDir overrides the original paths (useful for restore to different location)
	TargetDir string

	// SelectedFiles limits restore to specific files (empty = all)
	SelectedFiles []string

	// DiffWriter is where diff output is written (defaults to os.Stdout)
	DiffWriter io.Writer

	// ProgressCallback is called for each file restored
	ProgressCallback func(file string, action string)
}

// RestoreResult contains information about a completed restore
type RestoreResult struct {
	RestoredFiles []string          // Files that were restored
	SkippedFiles  []string          // Files that were skipped
	BackupFiles   []string          // .bak files created
	DiffResults   map[string]string // Diffs for each file
	TotalFiles    int
	FilesRestored int
	FilesSkipped  int
	FilesConflict int
}

// FileEntry represents a file extracted from backup
type FileEntry struct {
	Path    string
	Content []byte
	Mode    int64
	ModTime int64
}

// DiffResult represents the diff between two files
type DiffResult struct {
	HasDifference bool
	Diff          string
	BackupPath    string
	CurrentPath   string
}
