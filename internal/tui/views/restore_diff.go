package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiffViewer displays diff content in a scrollable viewport
type DiffViewer struct {
	viewport viewport.Model
	content  string
	file     string
	loading  bool
	active   bool
}

// NewDiffViewer creates a new diff viewer component
func NewDiffViewer() DiffViewer {
	vp := viewport.New(0, 0)
	return DiffViewer{
		viewport: vp,
	}
}

// Init initializes the diff viewer
func (d DiffViewer) Init() tea.Cmd {
	return nil
}

// Update handles messages and keyboard input
func (d DiffViewer) Update(msg tea.Msg) (DiffViewer, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Account for border padding (2 chars each side)
		borderW := 2
		borderH := 2
		d.viewport.Width = msg.Width - borderW
		d.viewport.Height = msg.Height - borderH
		if d.viewport.Width < 0 {
			d.viewport.Width = 0
		}
		if d.viewport.Height < 0 {
			d.viewport.Height = 0
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			d.viewport.LineDown(1)
		case "k", "up":
			d.viewport.LineUp(1)
		case "g":
			d.viewport.GotoTop()
		case "G":
			d.viewport.GotoBottom()
		default:
			d.viewport, cmd = d.viewport.Update(msg)
		}
	}

	return d, cmd
}

// View renders the diff viewer
func (d DiffViewer) View() string {
	if !d.active {
		return ""
	}

	var s strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	s.WriteString(titleStyle.Render("Diff Preview") + "\n")

	// Filename
	if d.file != "" {
		s.WriteString("File: " + d.file + "\n\n")
	}

	// Loading state
	if d.loading {
		s.WriteString("Loading diff...\n")
		return s.String()
	}

	// Viewport with border
	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(d.viewport.Width).
		Height(d.viewport.Height)

	s.WriteString(viewportStyle.Render(d.viewport.View()))

	return s.String()
}

// SetContent sets the diff content and filename, and activates the viewer
func (d *DiffViewer) SetContent(diff, filename string) {
	d.content = diff
	d.file = filename
	d.loading = false
	d.active = true
	d.viewport.SetContent(diff)
	d.viewport.GotoTop()
}

// SetLoading sets the loading state
func (d *DiffViewer) SetLoading(loading bool) {
	d.loading = loading
	if loading {
		d.active = true
	}
}

// IsActive returns whether the diff viewer is currently active/visible
func (d DiffViewer) IsActive() bool {
	return d.active
}

// Deactivate hides the diff viewer
func (d *DiffViewer) Deactivate() {
	d.active = false
	d.content = ""
	d.file = ""
	d.loading = false
}

// GetContent returns the current diff content
func (d DiffViewer) GetContent() string {
	return d.content
}

// GetFile returns the current file being diffed
func (d DiffViewer) GetFile() string {
	return d.file
}

// GetViewport returns the underlying viewport model (for external sizing)
func (d *DiffViewer) GetViewport() *viewport.Model {
	return &d.viewport
}
