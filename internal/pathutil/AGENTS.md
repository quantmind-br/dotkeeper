# PATHUTIL PACKAGE KNOWLEDGE BASE

**Scope:** `internal/pathutil/`

## OVERVIEW

Path utilities for dotfile discovery, glob matching, and filesystem scanning. 8 files, ~340 lines. Pure utility package with no external dependencies except `doublestar`.

## STRUCTURE

```
pathutil/
├── pathutil.go        # Core: ExpandHome() for ~ expansion
├── scanner.go         # Path scanning with stats (ScanPaths, FormatSize)
├── glob.go            # Glob pattern resolution with doublestar
├── presets.go         # Dotfile preset detection (shell configs, etc.)
└── *_test.go          # Unit tests
```

## API

```go
// Expand ~ to home directory
func ExpandHome(p string) string

// Scan files/folders with size stats
func ScanPaths(files, folders, exclude []string) ScanResult

// Format bytes to human-readable
func FormatSize(bytes int64) string

// Resolve glob patterns (supports **)
func ResolveGlob(pattern string, exclude []string) ([]string, error)

// Detect common dotfile presets
func DetectDotfiles(homeDir string) ([]DotfilePreset, []DotfilePreset)
```

## GLOB PATTERNS

Supports standard glob plus `**` for recursive matching:
- `~/.config/**/*.conf` - all .conf files recursively
- `~/.ssh/*` - direct children only

Capped at `MaxGlobResults = 500` to prevent runaway patterns.

## PRESETS

Auto-detected dotfiles for setup wizard:
- **Shells**: `.bashrc`, `.zshrc`, `.bash_profile`, `.profile`
- **Git**: `.gitconfig`, `.gitignore_global`
- **Tools**: `.vimrc`, `.tmux.conf`, `.ssh/config`
- **Folders**: `~/.config/nvim`, `~/.config/fish`, etc.

Selected based on `$SHELL` environment variable.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add preset paths | `presets.go` | PresetFiles or PresetFolders slice |
| Change glob limit | `glob.go:MaxGlobResults` | Currently 500 |
| Add path utility | New file or `pathutil.go` | Keep focused |
| Modify exclude logic | `scanner.go`, `glob.go` | Both use filepath.Match |

## ANTI-PATTERNS

- **Never** use os/exec for path operations
- **Don't** add circular dependencies - this is a leaf package
