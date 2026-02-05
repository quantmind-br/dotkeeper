package views

import (
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
)

// FileBrowserModel represents the file browser view
type FileBrowserModel struct {
	config     *config.Config
	filepicker filepicker.Model
	selected   []string
	width      int
	height     int
}

// NewFileBrowser creates a new file browser model
func NewFileBrowser(cfg *config.Config) FileBrowserModel {
	fp := filepicker.New()
	fp.CurrentDirectory = "."
	fp.ShowHidden = true

	return FileBrowserModel{
		config:     cfg,
		filepicker: fp,
	}
}

// Init initializes the file browser
func (m FileBrowserModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

// Update handles messages
func (m FileBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.filepicker.SetHeight(m.height - 2) // Subtract header height
	}

	// Check if file was selected
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selected = append(m.selected, path)
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)
	return m, cmd
}

// View renders the file browser
func (m FileBrowserModel) View() string {
	return "File Browser\n\n" + m.filepicker.View()
}

func (m FileBrowserModel) HelpBindings() []HelpEntry {
	return []HelpEntry{
		{"Enter", "Select"},
		{"↑/↓", "Navigate"},
	}
}
