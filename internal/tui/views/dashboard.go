package views

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	styles := DefaultStyles()

	var lastBackupVal string
	if !m.lastBackup.IsZero() {
		lastBackupVal = m.lastBackup.Format("2006-01-02 15:04")
	} else {
		lastBackupVal = "Never"
	}

	card1 := styles.Card.Render(
		styles.CardTitle.Render(lastBackupVal) + "\n" +
			styles.CardLabel.Render("Last Backup"),
	)

	card2 := styles.Card.Render(
		styles.CardTitle.Render(fmt.Sprintf("%d", m.fileCount)) + "\n" +
			styles.CardLabel.Render("Files Tracked"),
	)

	var statsBlock string
	if m.width >= 60 {
		statsBlock = lipgloss.JoinHorizontal(lipgloss.Top, card1, card2)
	} else {
		statsBlock = lipgloss.JoinVertical(lipgloss.Left, card1, card2)
	}

	btnBackup := styles.ActionButton.Render(styles.ActionButtonKey.Render("b") + " Backup")
	btnRestore := styles.ActionButton.Render(styles.ActionButtonKey.Render("r") + " Restore")
	btnSettings := styles.ActionButton.Render(styles.ActionButtonKey.Render("s") + " Settings")

	var actionsBlock string
	if m.width >= 60 {
		actionsBlock = lipgloss.JoinHorizontal(lipgloss.Top, btnBackup, btnRestore, btnSettings)
	} else {
		actionsBlock = lipgloss.JoinVertical(lipgloss.Left, btnBackup, btnRestore, btnSettings)
	}

	statusBar := RenderStatusBar(m.width, "", "", "b: backup | r: restore | s: settings")

	return lipgloss.JoinVertical(lipgloss.Left,
		statsBlock,
		"\n",
		actionsBlock,
		"\n",
		statusBar,
	)
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
