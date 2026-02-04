package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/config"
)

// SettingsModel represents the settings view
type SettingsModel struct {
	config *config.Config
	width  int
	height int
}

// NewSettings creates a new settings model
func NewSettings(cfg *config.Config) SettingsModel {
	return SettingsModel{
		config: cfg,
	}
}

// Init initializes the settings view
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the settings view
func (m SettingsModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	b.WriteString(titleStyle.Render("Settings") + "\n\n")

	labelStyle := lipgloss.NewStyle().Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

	b.WriteString(labelStyle.Render("Backup Directory: "))
	b.WriteString(valueStyle.Render(m.config.BackupDir) + "\n")

	b.WriteString(labelStyle.Render("Git Remote: "))
	b.WriteString(valueStyle.Render(m.config.GitRemote) + "\n")

	b.WriteString(labelStyle.Render("Files: "))
	b.WriteString(valueStyle.Render(fmt.Sprintf("%d files, %d folders", len(m.config.Files), len(m.config.Folders))) + "\n")

	b.WriteString(labelStyle.Render("Schedule: "))
	if m.config.Schedule != "" {
		b.WriteString(valueStyle.Render(m.config.Schedule) + "\n")
	} else {
		b.WriteString(valueStyle.Render("Not scheduled") + "\n")
	}

	b.WriteString(labelStyle.Render("Notifications: "))
	b.WriteString(valueStyle.Render(fmt.Sprintf("%v", m.config.Notifications)) + "\n")

	return b.String()
}
