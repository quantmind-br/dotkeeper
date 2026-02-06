package views

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/pathutil"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

// backupItem represents a backup entry in lists
type backupItem struct {
	name string
	size int64
	date string
}

func (i backupItem) Title() string       { return i.name }
func (i backupItem) Description() string { return fmt.Sprintf("%s - %d bytes", i.date, i.size) }
func (i backupItem) FilterValue() string { return i.name }

// backupsLoadedMsg carries loaded backup items to the view.
type backupsLoadedMsg []list.Item

// LoadBackupItems scans a backup directory and returns backup items sorted newest-first.
func LoadBackupItems(backupDir string) []list.Item {
	dir := pathutil.ExpandHome(backupDir)
	paths, _ := filepath.Glob(filepath.Join(dir, "backup-*.tar.gz.enc"))

	for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
		paths[i], paths[j] = paths[j], paths[i]
	}

	items := make([]list.Item, 0, len(paths))
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info != nil {
			name := strings.TrimSuffix(filepath.Base(p), ".tar.gz.enc")
			items = append(items, backupItem{
				name: name,
				size: info.Size(),
				date: info.ModTime().Format("2006-01-02 15:04"),
			})
		}
	}
	return items
}

type PathValidationResult struct {
	Valid        bool
	Exists       bool
	IsDir        bool
	IsFile       bool
	Readable     bool
	Error        string
	ExpandedPath string
}

func ValidatePath(path string) PathValidationResult {
	result := PathValidationResult{
		ExpandedPath: pathutil.ExpandHome(path),
	}

	// Check if path exists
	info, err := os.Stat(result.ExpandedPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Error = fmt.Sprintf("path does not exist: %s", result.ExpandedPath)
		} else if os.IsPermission(err) {
			result.Error = fmt.Sprintf("permission denied: %s", result.ExpandedPath)
		} else {
			result.Error = fmt.Sprintf("error accessing path: %v", err)
		}
		return result
	}

	result.Exists = true
	result.IsDir = info.IsDir()
	result.IsFile = !info.IsDir()

	// Check readability
	if result.IsDir {
		// For directories, try to read the directory
		_, err := os.ReadDir(result.ExpandedPath)
		if err != nil {
			result.Error = fmt.Sprintf("directory not readable: %s", result.ExpandedPath)
			return result
		}
		result.Readable = true
	} else {
		// For files, try to open them
		file, err := os.Open(result.ExpandedPath)
		if err != nil {
			result.Error = fmt.Sprintf("file not readable: %s", result.ExpandedPath)
			return result
		}
		file.Close()
		result.Readable = true
	}

	result.Valid = true
	return result
}

func ValidateFilePath(path string) (string, error) {
	result := ValidatePath(path)

	if !result.Valid {
		return "", errors.New(result.Error)
	}

	if result.IsDir {
		return "", fmt.Errorf("path is a directory, not a file: %s", result.ExpandedPath)
	}

	return result.ExpandedPath, nil
}

func ValidateFolderPath(path string) (string, error) {
	result := ValidatePath(path)

	if !result.Valid {
		return "", errors.New(result.Error)
	}

	if result.IsFile {
		return "", fmt.Errorf("path is a file, not a directory: %s", result.ExpandedPath)
	}

	return result.ExpandedPath, nil
}

// HelpEntry represents a single keyboard shortcut entry
type HelpEntry struct {
	Key         string
	Description string
}

// HelpProvider is implemented by views that expose keyboard shortcuts
type HelpProvider interface {
	HelpBindings() []HelpEntry
}

type RefreshBackupListMsg struct{}

// DashboardNavigateMsg requests navigation from dashboard action buttons.
type DashboardNavigateMsg struct {
	Target string
}

func RenderStatusBar(width int, status string, errMsg string, helpText string) string {
	st := styles.DefaultStyles()
	var s strings.Builder

	msg := ""
	style := st.Success
	if errMsg != "" {
		msg = errMsg
		style = st.Error
	} else if status != "" {
		msg = status
		style = st.Success
	}

	if msg != "" {
		if width > 7 && len(msg) > width-4 {
			msg = msg[:width-7] + "..."
		}
		s.WriteString(style.Render(msg) + "\n")
	}

	s.WriteString(st.Help.Render(helpText))
	return st.StatusBar.Render(s.String())
}

func PlaceOverlay(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
