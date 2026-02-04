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

type backupItem struct {
	name string
	size int64
	date string
}

func (i backupItem) Title() string       { return i.name }
func (i backupItem) Description() string { return fmt.Sprintf("%s - %d bytes", i.date, i.size) }
func (i backupItem) FilterValue() string { return i.name }

type backupsLoadedMsg []list.Item

type BackupListModel struct {
	config *config.Config
	list   list.Model
	width  int
	height int
}

func NewBackupList(cfg *config.Config) BackupListModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Backups"

	return BackupListModel{
		config: cfg,
		list:   l,
	}
}

func (m BackupListModel) Init() tea.Cmd {
	return m.Refresh()
}

func (m BackupListModel) Refresh() tea.Cmd {
	return func() tea.Msg {
		dir := expandHome(m.config.BackupDir)
		paths, _ := filepath.Glob(filepath.Join(dir, "backup-*.tar.gz.enc"))

		items := make([]list.Item, 0, len(paths))
		for _, p := range paths {
			if info, err := os.Stat(p); err == nil && info != nil {
				name := strings.TrimSuffix(filepath.Base(p), ".tar.gz.enc")
				items = append(items, backupItem{
					name: name,
					size: info.Size(),
					date: info.ModTime().Format("2006-01-02 15:04"),
				})
			}
		}
		return backupsLoadedMsg(items)
	}
}

func (m BackupListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
	case backupsLoadedMsg:
		m.list.SetItems([]list.Item(msg))
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m BackupListModel) View() string {
	return m.list.View()
}
