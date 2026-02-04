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
	"github.com/diogo/dotkeeper/internal/restore"
)

// RestoreModel represents the restore view
type RestoreModel struct {
	config           *config.Config
	width            int
	height           int
	backupList       list.Model
	phase            int // 0=backup list, 1=password, 2=file select, 3=restoring, 4=diff preview
	selectedBackup   string
	password         string // validated password for restore
	passwordInput    textinput.Model
	fileList         list.Model
	selectedFiles    map[string]bool
	restoreStatus    string
	restoreError     string
	passwordAttempts int
	viewport         viewport.Model
	currentDiff      string
}

type passwordValidMsg struct{}

type passwordInvalidMsg struct {
	err error
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

func (m RestoreModel) validatePassword(backupPath, password string) tea.Cmd {
	return func() tea.Msg {
		err := restore.ValidateBackup(backupPath, password)
		if err != nil {
			return passwordInvalidMsg{err: err}
		}
		return passwordValidMsg{}
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

	case passwordValidMsg:
		m.password = m.passwordInput.Value()
		m.phase = 2
		m.restoreStatus = ""
		m.restoreError = ""
		m.passwordAttempts = 0
		return m, nil

	case passwordInvalidMsg:
		m.passwordAttempts++
		if m.passwordAttempts >= 3 {
			m.restoreError = "Too many failed attempts"
			m.phase = 0
			m.passwordAttempts = 0
			m.passwordInput.SetValue("")
			m.passwordInput.Blur()
		} else {
			m.restoreError = fmt.Sprintf("Invalid password (attempt %d/3)", m.passwordAttempts)
			m.passwordInput.SetValue("")
		}
		m.restoreStatus = ""
		return m, nil

	case tea.KeyMsg:
		if m.phase == 0 {
			switch msg.String() {
			case "enter":
				if item := m.backupList.SelectedItem(); item != nil {
					selected := item.(backupItem)
					backupPath := filepath.Join(expandHome(m.config.BackupDir), selected.name+".tar.gz.enc")
					m.selectedBackup = backupPath
					m.passwordInput.SetValue("")
					m.restoreError = ""
					m.phase = 1
					m.passwordInput.Focus()
					return m, textinput.Blink
				}
			case "r":
				return m, m.refreshBackups()
			default:
				var cmd tea.Cmd
				m.backupList, cmd = m.backupList.Update(msg)
				return m, cmd
			}
		} else if m.phase == 1 {
			switch msg.String() {
			case "enter":
				if m.passwordInput.Value() != "" {
					m.restoreStatus = "Validating password..."
					m.restoreError = ""
					return m, m.validatePassword(m.selectedBackup, m.passwordInput.Value())
				}
			case "esc":
				m.phase = 0
				m.passwordInput.SetValue("")
				m.passwordAttempts = 0
				m.restoreError = ""
				m.passwordInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.passwordInput, cmd = m.passwordInput.Update(msg)
				return m, cmd
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the restore view
func (m RestoreModel) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))

	// Phase 0: Backup list selection
	if m.phase == 0 {
		s.WriteString(m.backupList.View())
		s.WriteString("\n")

		if m.restoreStatus != "" {
			s.WriteString(statusStyle.Render(m.restoreStatus) + "\n")
		}
		if m.restoreError != "" {
			s.WriteString(errorStyle.Render(m.restoreError) + "\n")
		}

		s.WriteString(helpStyle.Render("↑/↓: navigate | Enter: select | r: refresh"))
		return s.String()
	}

	// Phase 1: Password entry
	if m.phase == 1 {
		s.WriteString(titleStyle.Render("Enter Password") + "\n\n")
		s.WriteString(fmt.Sprintf("Backup: %s\n\n", filepath.Base(m.selectedBackup)))
		s.WriteString(m.passwordInput.View() + "\n\n")

		if m.restoreStatus != "" {
			s.WriteString(statusStyle.Render(m.restoreStatus) + "\n")
		}
		if m.restoreError != "" {
			s.WriteString(errorStyle.Render(m.restoreError) + "\n")
		}

		s.WriteString(helpStyle.Render("Enter: validate | Esc: back"))
		return s.String()
	}

	// Phase 2: File selection
	if m.phase == 2 {
		s.WriteString(titleStyle.Render("Select Files") + "\n\n")
		s.WriteString("File selection (implementation pending)")
		return s.String()
	}

	// Placeholder for other phases
	s.WriteString(titleStyle.Render("Restore") + "\n\n")
	s.WriteString("Phase " + fmt.Sprintf("%d", m.phase) + " (implementation pending)")

	return s.String()
}
