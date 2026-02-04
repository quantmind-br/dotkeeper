package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/config"
)

// RestoreModel represents the restore view
type RestoreModel struct {
	config           *config.Config
	width            int
	height           int
	backupList       list.Model
	phase            int // 0=backup list, 1=password, 2=file select, 3=restoring, 4=diff preview
	selectedBackup   string
	passwordInput    textinput.Model
	fileList         list.Model
	selectedFiles    map[string]bool
	restoreStatus    string
	restoreError     string
	passwordAttempts int
	viewport         viewport.Model
	currentDiff      string
}

// NewRestore creates a new restore model
func NewRestore(cfg *config.Config) RestoreModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Backups"
	l.SetShowHelp(false)

	ti := textinput.New()
	ti.Placeholder = "Enter password for decryption"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Width = 40

	fl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	fl.Title = "Files"
	fl.SetShowHelp(false)

	vp := viewport.New(0, 0)

	return RestoreModel{
		config:        cfg,
		backupList:    l,
		passwordInput: ti,
		fileList:      fl,
		selectedFiles: make(map[string]bool),
		viewport:      vp,
		phase:         0,
	}
}

// Init initializes the restore view
func (m RestoreModel) Init() tea.Cmd {
	return m.refreshBackups()
}

// refreshBackups scans the backup directory and loads available backups
func (m RestoreModel) refreshBackups() tea.Cmd {
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

// Update handles messages
func (m RestoreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.backupList.SetSize(msg.Width, msg.Height-6)
		m.fileList.SetSize(msg.Width, msg.Height-6)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6

	case backupsLoadedMsg:
		m.backupList.SetItems([]list.Item(msg))
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// View renders the restore view
func (m RestoreModel) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Phase 0: Backup list selection
	if m.phase == 0 {
		s.WriteString(m.backupList.View())
		s.WriteString("\n")

		if m.restoreStatus != "" {
			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
			s.WriteString(successStyle.Render(m.restoreStatus) + "\n")
		}
		if m.restoreError != "" {
			errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
			s.WriteString(errorStyle.Render(m.restoreError) + "\n")
		}

		s.WriteString(helpStyle.Render("↑/↓: navigate | Enter: select | r: refresh"))
		return s.String()
	}

	// Placeholder for other phases
	s.WriteString(titleStyle.Render("Restore") + "\n\n")
	s.WriteString("Phase " + fmt.Sprintf("%d", m.phase) + " (implementation pending)")

	return s.String()
}
