package views

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") || p == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return p
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
		ExpandedPath: expandHome(path),
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

func RenderStatusBar(width int, status string, errMsg string, helpText string) string {
	styles := DefaultStyles()
	var s strings.Builder

	msg := ""
	style := styles.Success
	if errMsg != "" {
		msg = errMsg
		style = styles.Error
	} else if status != "" {
		msg = status
		style = styles.Success
	}

	if msg != "" {
		if width > 7 && len(msg) > width-4 {
			msg = msg[:width-7] + "..."
		}
		s.WriteString(style.Render(msg) + "\n")
	}

	s.WriteString(styles.Help.Render(helpText))
	return s.String()
}

func PlaceOverlay(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
