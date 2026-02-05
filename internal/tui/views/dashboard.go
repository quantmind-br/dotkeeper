package views

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
)

// DashboardModel represents the dashboard view
type DashboardModel struct {
	config     *config.Config
	width      int
	height     int
	lastBackup time.Time
	fileCount  int
	err        error
}

// NewDashboard creates a new dashboard model
func NewDashboard(cfg *config.Config) DashboardModel {
	return DashboardModel{
		config: cfg,
	}
}

// Init initializes the dashboard
func (m DashboardModel) Init() tea.Cmd {
	return m.refreshStatus()
}

// Update handles messages
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case statusMsg:
		m.lastBackup = msg.lastBackup
		m.fileCount = msg.fileCount
	}
	return m, nil
}

// View renders the dashboard
func (m DashboardModel) View() string {
	var s string

	styles := DefaultStyles()
	s += styles.Title.Render("Dashboard") + "\n\n"

	// Status section
	if !m.lastBackup.IsZero() {
		s += fmt.Sprintf("Last Backup: %s\n", m.lastBackup.Format("2006-01-02 15:04"))
	} else {
		s += "Last Backup: Never\n"
	}

	s += fmt.Sprintf("Files Tracked: %d\n\n", m.fileCount)

	// Quick actions
	s += "Quick Actions:\n"
	s += "  [b] Backup now\n"
	s += "  [r] Restore\n"
	s += "  [s] Settings\n"

	return s
}

type statusMsg struct {
	lastBackup time.Time
	fileCount  int
}

func (m DashboardModel) refreshStatus() tea.Cmd {
	return func() tea.Msg {
		count := len(m.config.Files) + len(m.config.Folders)

		var lastBackup time.Time
		dir := expandHome(m.config.BackupDir)
		backups, _ := filepath.Glob(filepath.Join(dir, "backup-*.tar.gz.enc"))
		if len(backups) > 0 {
			info, _ := os.Stat(backups[len(backups)-1])
			if info != nil {
				lastBackup = info.ModTime()
			}
		}

		return statusMsg{
			lastBackup: lastBackup,
			fileCount:  count,
		}
	}
}

func (m DashboardModel) HelpBindings() []HelpEntry {
	return []HelpEntry{
		{"b", "Go to backups"},
		{"r", "Go to restore"},
		{"s", "Go to settings"},
	}
}
