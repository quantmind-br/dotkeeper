package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/backup"
	"github.com/diogo/dotkeeper/internal/history"
	"github.com/diogo/dotkeeper/internal/pathutil"
	"github.com/diogo/dotkeeper/internal/tui/components"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

type BackupSuccessMsg struct {
	Result *backup.BackupResult
}

// BackupErrorMsg is consolidated to ErrorMsg with Source="backup".
type BackupErrorMsg = ErrorMsg

type backupDeletedMsg struct{ name string }

// backupDeleteErrorMsg is consolidated to ErrorMsg with Source="backup-delete".
type backupDeleteErrorMsg = ErrorMsg

type BackupListModel struct {
	ctx              *ProgramContext
	list             list.Model
	creatingBackup   bool
	confirmingDelete bool
	deleteTarget     string
	passwordInput    textinput.Model
	backupStatus     string
	backupError      string
	spinner          spinner.Model
	loading          bool
}

func NewBackupList(ctx *ProgramContext) BackupListModel {
	l := styles.NewMinimalList()

	ti := components.NewPasswordInput("Enter password for encryption")
	ti.Width = 40

	s := spinner.New()
	s.Spinner = spinner.Dot

	return BackupListModel{
		ctx:           ensureProgramContext(ctx),
		list:          l,
		passwordInput: ti,
		spinner:       s,
	}
}

func (m BackupListModel) Init() tea.Cmd {
	return m.Refresh()
}

func (m *BackupListModel) Refresh() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		if m.ctx.Config == nil {
			return backupsLoadedMsg([]list.Item{})
		}
		return backupsLoadedMsg(LoadBackupItems(m.ctx.Config.BackupDir))
	}
}

func (m *BackupListModel) runBackup(password string) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		if m.ctx.Config == nil {
			return BackupErrorMsg{Source: "backup", Err: fmt.Errorf("missing config")}
		}
		cfg := m.ctx.Config
		cfg.BackupDir = pathutil.ExpandHome(cfg.BackupDir)

		result, err := backup.Backup(cfg, password)
		if err != nil {
			return BackupErrorMsg{Source: "backup", Err: err}
		}
		return BackupSuccessMsg{Result: result}
	}
}

func (m *BackupListModel) deleteBackup(name string) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		if m.ctx.Config == nil {
			return backupDeleteErrorMsg{Source: "backup-delete", Err: fmt.Errorf("missing config")}
		}
		dir := pathutil.ExpandHome(m.ctx.Config.BackupDir)
		encPath := filepath.Join(dir, name+".tar.gz.enc")
		metaPath := encPath + ".meta.json"

		if err := os.Remove(encPath); err != nil {
			return backupDeleteErrorMsg{Source: "backup-delete", Err: fmt.Errorf("delete %s: %w", filepath.Base(encPath), err)}
		}
		os.Remove(metaPath)
		return backupDeletedMsg{name: name}
	}
}

func (m BackupListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		// Responsive password input width
		pw := msg.Width - 6
		if pw < 20 {
			pw = 20
		}
		if pw > 60 {
			pw = 60
		}
		m.passwordInput.Width = pw

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case backupsLoadedMsg:
		m.loading = false
		m.list.SetItems([]list.Item(msg))
		return m, nil

	case BackupSuccessMsg:
		m.creatingBackup = false
		m.loading = false
		m.backupStatus = fmt.Sprintf("✓ Backup created: %s (%d files)", msg.Result.BackupName, msg.Result.FileCount)
		m.backupError = ""
		m.passwordInput.SetValue("")
		if m.ctx.Store != nil {
			_ = m.ctx.Store.Append(history.EntryFromBackupResult(msg.Result))
		}
		return m, m.Refresh()

	case BackupErrorMsg:
		if msg.Source == "backup" {
			m.creatingBackup = false
			m.loading = false
			m.backupStatus = ""
			m.backupError = fmt.Sprintf("✗ Backup failed: %v", msg.Err)
			m.passwordInput.SetValue("")
			if m.ctx.Store != nil {
				_ = m.ctx.Store.Append(history.EntryFromBackupError(msg.Err))
			}
			return m, nil
		} else if msg.Source == "backup-delete" {
			m.confirmingDelete = false
			m.loading = false
			m.backupStatus = ""
			m.backupError = fmt.Sprintf("✗ Delete failed: %v", msg.Err)
			m.deleteTarget = ""
			return m, nil
		}

	case backupDeletedMsg:
		m.confirmingDelete = false
		m.loading = false
		m.backupStatus = fmt.Sprintf("✓ Deleted: %s", msg.name)
		m.backupError = ""
		m.deleteTarget = ""
		return m, m.Refresh()

	case tea.KeyMsg:
		if m.confirmingDelete {
			switch msg.String() {
			case "y", "Y":
				return m, m.deleteBackup(m.deleteTarget)
			default:
				m.confirmingDelete = false
				m.deleteTarget = ""
				return m, nil
			}
		}

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
		case "d":
			if item := m.list.SelectedItem(); item != nil {
				selected := item.(backupItem)
				m.confirmingDelete = true
				m.deleteTarget = selected.name
				m.backupStatus = ""
				m.backupError = ""
				return m, nil
			}
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

	st := styles.DefaultStyles()

	if m.loading && !m.creatingBackup && !m.confirmingDelete {
		return lipgloss.JoinVertical(lipgloss.Center,
			"\n",
			m.spinner.View(),
			"\nLoading backups...",
		)
	}

	if m.confirmingDelete {
		s.WriteString(st.Title.Render("Delete Backup") + "\n\n")
		s.WriteString(fmt.Sprintf("Are you sure you want to delete %s?\n\n", st.Value.Render(m.deleteTarget)))
		return s.String()
	}

	if m.creatingBackup {
		s.WriteString(st.Title.Render("Create New Backup") + "\n\n")
		s.WriteString("Enter encryption password:\n\n")
		s.WriteString(m.passwordInput.View() + "\n\n")
		s.WriteString(RenderStatusBar(m.ctx.Width, m.backupStatus, m.backupError, ""))
		return s.String()
	}

	s.WriteString(st.Title.Render("Backups") + "\n\n")
	s.WriteString(m.list.View())
	s.WriteString("\n")

	s.WriteString(RenderStatusBar(m.ctx.Width, m.backupStatus, m.backupError, ""))

	return s.String()
}

func (m BackupListModel) HelpBindings() []HelpEntry {
	if m.confirmingDelete {
		return []HelpEntry{
			{"y", "Confirm delete"},
			{"any", "Cancel"},
		}
	}
	if m.creatingBackup {
		return []HelpEntry{
			{"Enter", "Create backup"},
			{"Esc", "Cancel"},
		}
	}
	return []HelpEntry{
		{"n/c", "New backup"},
		{"d", "Delete backup"},
		{"r", "Refresh list"},
		{"↑/↓", "Navigate"},
	}
}

func (m BackupListModel) StatusHelpText() string {
	if m.confirmingDelete {
		return "y: confirm | any other key: cancel"
	}
	if m.creatingBackup {
		return "Press Enter to create backup, Esc to cancel"
	}
	return "n: new backup | d: delete | r: refresh | ↑/↓: navigate"
}

func (m BackupListModel) IsCreating() bool {
	return m.creatingBackup || m.confirmingDelete
}

func (m BackupListModel) IsInputActive() bool {
	return m.IsCreating()
}
