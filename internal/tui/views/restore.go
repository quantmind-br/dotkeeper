package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/config"
)

// RestoreModel represents the restore view
type RestoreModel struct {
	config *config.Config
	width  int
	height int
}

// NewRestore creates a new restore model
func NewRestore(cfg *config.Config) RestoreModel {
	return RestoreModel{
		config: cfg,
	}
}

// Init initializes the restore view
func (m RestoreModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m RestoreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the restore view
func (m RestoreModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	return titleStyle.Render("Restore") + "\n\nSelect a backup to restore (implementation pending)"
}
