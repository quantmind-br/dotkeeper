package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
)

// backupItem represents a backup in the list
type backupItem struct {
	name string
	size int64
	date string
}

func (i backupItem) Title() string       { return i.name }
func (i backupItem) Description() string { return fmt.Sprintf("%s - %d bytes", i.date, i.size) }
func (i backupItem) FilterValue() string { return i.name }

// BackupListModel represents the backup list view
type BackupListModel struct {
	config *config.Config
	list   list.Model
	width  int
	height int
}

// NewBackupList creates a new backup list model
func NewBackupList(cfg *config.Config) BackupListModel {
	items := []list.Item{}

	// Load backups from directory
	backups, _ := filepath.Glob(filepath.Join(cfg.BackupDir, "backup-*.tar.gz.enc"))
	for _, backup := range backups {
		info, _ := os.Stat(backup)
		if info != nil {
			name := strings.TrimSuffix(filepath.Base(backup), ".tar.gz.enc")
			items = append(items, backupItem{
				name: name,
				size: info.Size(),
				date: info.ModTime().Format("2006-01-02 15:04"),
			})
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Backups"

	return BackupListModel{
		config: cfg,
		list:   l,
	}
}

// Init initializes the backup list
func (m BackupListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m BackupListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the backup list
func (m BackupListModel) View() string {
	return m.list.View()
}
