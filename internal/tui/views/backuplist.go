package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/backup"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/history"
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

type BackupSuccessMsg struct {
	Result *backup.BackupResult
}

type BackupErrorMsg struct {
	Error error
}

type BackupListModel struct {
	config         *config.Config
	store          *history.Store
	list           list.Model
	width          int
	height         int
	creatingBackup bool
	passwordInput  textinput.Model
	backupStatus   string
	backupError    string
}

func NewBackupList(cfg *config.Config, store *history.Store) BackupListModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Backups"
	l.SetShowHelp(false)

	ti := textinput.New()
	ti.Placeholder = "Enter password for encryption"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Width = 40

	return BackupListModel{
		config:        cfg,
		store:         store,
		list:          l,
		passwordInput: ti,
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

func (m BackupListModel) runBackup(password string) tea.Cmd {
	return func() tea.Msg {
		cfg := m.config
		cfg.BackupDir = expandHome(cfg.BackupDir)

		result, err := backup.Backup(cfg, password)
		if err != nil {
			return BackupErrorMsg{Error: err}
		}
		return BackupSuccessMsg{Result: result}
	}
}

func (m BackupListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)

	case backupsLoadedMsg:
		m.list.SetItems([]list.Item(msg))
		return m, nil

	case BackupSuccessMsg:
		m.creatingBackup = false
		m.backupStatus = fmt.Sprintf("✓ Backup created: %s (%d files)", msg.Result.BackupName, msg.Result.FileCount)
		m.backupError = ""
		m.passwordInput.SetValue("")
		if m.store != nil {
			_ = m.store.Append(history.EntryFromBackupResult(msg.Result))
		}
		return m, m.Refresh()

	case BackupErrorMsg:
		m.creatingBackup = false
		m.backupStatus = ""
		m.backupError = fmt.Sprintf("✗ Backup failed: %v", msg.Error)
		m.passwordInput.SetValue("")
		if m.store != nil {
			_ = m.store.Append(history.EntryFromBackupError(msg.Error))
		}
		return m, nil

	case tea.KeyMsg:
		if m.creatingBackup {
			switch msg.String() {
			case "enter":
				password := m.passwordInput.Value()
				if password != "" {
					m.backupStatus = "Creating backup..."
					m.backupError = ""
					return m, m.runBackup(password)
				}
			case "esc":
				m.creatingBackup = false
				m.passwordInput.SetValue("")
				m.passwordInput.Blur()
				return m, nil
			}

			var cmd tea.Cmd
			m.passwordInput, cmd = m.passwordInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "n", "c":
			m.creatingBackup = true
			m.backupStatus = ""
			m.backupError = ""
			m.passwordInput.Focus()
			return m, textinput.Blink
		case "r":
			return m, m.Refresh()
		}
	}

	if !m.creatingBackup {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m BackupListModel) View() string {
	var s strings.Builder

	styles := DefaultStyles()

	if m.creatingBackup {
		s.WriteString(styles.Title.Render("Create New Backup") + "\n\n")
		s.WriteString("Enter encryption password:\n\n")
		s.WriteString(m.passwordInput.View() + "\n\n")
		s.WriteString(styles.Help.Render("Press Enter to create backup, Esc to cancel"))
		return s.String()
	}

	s.WriteString(m.list.View())
	s.WriteString("\n")

	if m.backupStatus != "" {
		s.WriteString(styles.Success.Render(m.backupStatus) + "\n")
	}
	if m.backupError != "" {
		s.WriteString(styles.Error.Render(m.backupError) + "\n")
	}

	s.WriteString(styles.Help.Render("n: new backup | r: refresh | ↑/↓: navigate"))

	return s.String()
}

func (m BackupListModel) HelpBindings() []HelpEntry {
	if m.creatingBackup {
		return []HelpEntry{
			{"Enter", "Create backup"},
			{"Esc", "Cancel"},
		}
	}
	return []HelpEntry{
		{"n/c", "New backup"},
		{"r", "Refresh list"},
		{"↑/↓", "Navigate"},
	}
}

// IsCreating returns true when the backup list is in backup creation mode (password input).
func (m BackupListModel) IsCreating() bool {
	return m.creatingBackup
}
