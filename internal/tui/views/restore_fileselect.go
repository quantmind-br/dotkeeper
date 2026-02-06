package views

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

// FileSelector handles file selection UI for restore operations
type FileSelector struct {
	list     list.Model
	selected map[string]bool
}

// NewFileSelector creates a new FileSelector with default settings
func NewFileSelector() FileSelector {
	return FileSelector{
		list:     styles.NewMinimalList(),
		selected: make(map[string]bool),
	}
}

// Init returns the initial command for the file selector
func (fs FileSelector) Init() tea.Cmd {
	return nil
}

// Update handles messages for the file selector
func (fs FileSelector) Update(msg tea.Msg) (FileSelector, tea.Cmd) {
	var cmd tea.Cmd
	fs.list, cmd = fs.list.Update(msg)
	return fs, cmd
}

// View renders the file selector
func (fs FileSelector) View() string {
	return fs.list.View()
}

// SetFiles populates the file selector with files
func (fs *FileSelector) SetFiles(files []FileEntry) {
	items := make([]list.Item, len(files))
	fs.selected = make(map[string]bool)
	for i, entry := range files {
		items[i] = fileItem{
			path:     entry.Path,
			size:     int64(len(entry.Content)),
			selected: false,
		}
		fs.selected[entry.Path] = false
	}
	fs.list.SetItems(items)
}

// SetSize updates the size of the list
func (fs *FileSelector) SetSize(width, height int) {
	fs.list.SetSize(width, height)
}

// ToggleItem toggles the selection state of the currently selected item
func (fs *FileSelector) ToggleItem() {
	if item := fs.list.SelectedItem(); item != nil {
		fi := item.(fileItem)
		fs.selected[fi.path] = !fs.selected[fi.path]
		fs.updateFileListSelection()
	}
}

// SelectAll selects all files
func (fs *FileSelector) SelectAll() {
	for path := range fs.selected {
		fs.selected[path] = true
	}
	fs.updateFileListSelection()
}

// DeselectAll deselects all files
func (fs *FileSelector) DeselectAll() {
	for path := range fs.selected {
		fs.selected[path] = false
	}
	fs.updateFileListSelection()
}

// SelectedFiles returns the list of selected file paths
func (fs FileSelector) SelectedFiles() []string {
	var paths []string
	for path, selected := range fs.selected {
		if selected {
			paths = append(paths, path)
		}
	}
	return paths
}

// SelectedCount returns the number of selected files
func (fs FileSelector) SelectedCount() int {
	count := 0
	for _, selected := range fs.selected {
		if selected {
			count++
		}
	}
	return count
}

// TotalCount returns the total number of files
func (fs FileSelector) TotalCount() int {
	return len(fs.selected)
}

// IsActive returns true if the file selector is active
func (fs FileSelector) IsActive() bool {
	return true
}

// Reset clears the file selector state
func (fs *FileSelector) Reset() {
	fs.selected = make(map[string]bool)
	fs.list.SetItems([]list.Item{})
}

// updateFileListSelection updates the UI to reflect current selections
func (fs *FileSelector) updateFileListSelection() {
	items := fs.list.Items()
	newItems := make([]list.Item, len(items))
	for i, item := range items {
		fi := item.(fileItem)
		fi.selected = fs.selected[fi.path]
		newItems[i] = fi
	}
	fs.list.SetItems(newItems)
}

// FileEntry represents a file entry for the file selector
type FileEntry struct {
	Path    string
	Content []byte
}

// SelectedItemPath returns the path of the currently selected item
func (fs FileSelector) SelectedItemPath() string {
	if item := fs.list.SelectedItem(); item != nil {
		return item.(fileItem).path
	}
	return ""
}
