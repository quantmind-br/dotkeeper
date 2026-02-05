package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
)

// LogsModel represents the logs view
type LogsModel struct {
	config *config.Config
	width  int
	height int
}

// NewLogs creates a new logs model
func NewLogs(cfg *config.Config) LogsModel {
	return LogsModel{
		config: cfg,
	}
}

// Init initializes the logs view
func (m LogsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m LogsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the logs view
func (m LogsModel) View() string {
	styles := DefaultStyles()
	return styles.Title.Render("Logs") + "\n\nOperation history will be displayed here (implementation pending)"
}

func (m LogsModel) HelpBindings() []HelpEntry {
	return []HelpEntry{
		{"f", "Cycle filter"},
		{"r", "Refresh"},
		{"↑/↓", "Navigate"},
	}
}
