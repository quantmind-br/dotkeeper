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
	phase            int // 0=backup list, 1=password, 2=file select, 3=restoring, 4=diff preview, 5=results
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
	diffFile         string                 // path of file being diffed
	restoreResult    *restore.RestoreResult // result of restore operation
}

type passwordValidMsg struct{}

type passwordInvalidMsg struct {
	err error
}

type filesLoadedMsg struct {
	files []restore.FileEntry
}

type diffLoadedMsg struct {
	diff string
	file string
}

type diffErrorMsg struct {
	err error
}

type restoreCompleteMsg struct {
	result *restore.RestoreResult
}

type restoreErrorMsg struct {
	err error
}

// fileItem represents a file in the restore list with selection state
type fileItem struct {
	path     string
	size     int64
	selected bool
}

func (i fileItem) Title() string {
	checkbox := "[ ]"
	if i.selected {
		checkbox = "[x]"
	}
	return checkbox + " " + i.path
}

func (i fileItem) Description() string {
	return fmt.Sprintf("%d bytes", i.size)
}

func (i fileItem) FilterValue() string {
	return i.path
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

// Refresh reloads the backup list
func (m RestoreModel) Refresh() tea.Cmd {
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

func (m RestoreModel) loadFiles(backupPath, password string) tea.Cmd {
	return func() tea.Msg {
		entries, err := restore.ListBackupContents(backupPath, password)
		if err != nil {
			return passwordInvalidMsg{err: fmt.Errorf("failed to load files: %w", err)}
		}
		return filesLoadedMsg{files: entries}
	}
}

func (m RestoreModel) loadDiff(filePath string) tea.Cmd {
	return func() tea.Msg {
		diff, err := restore.GetFileDiff(m.selectedBackup, m.password, filePath)
		if err != nil {
			if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
				return diffLoadedMsg{
					diff: "[New file - will be created during restore]",
					file: filePath,
				}
			}
			return diffErrorMsg{err: err}
		}
		if diff == "" {
			return diffLoadedMsg{
				diff: "[No differences - file is identical]",
				file: filePath,
			}
		}
		return diffLoadedMsg{diff: diff, file: filePath}
	}
}

func (m RestoreModel) runRestore() tea.Cmd {
	return func() tea.Msg {
		opts := restore.RestoreOptions{
			SelectedFiles: m.getSelectedFilePaths(),
		}

		result, err := restore.Restore(m.selectedBackup, m.password, opts)
		if err != nil {
			return restoreErrorMsg{err: err}
		}
		return restoreCompleteMsg{result: result}
	}
}

func (m *RestoreModel) updateFileListSelection() {
	items := m.fileList.Items()
	newItems := make([]list.Item, len(items))
	for i, item := range items {
		fi := item.(fileItem)
		fi.selected = m.selectedFiles[fi.path]
		newItems[i] = fi
	}
	m.fileList.SetItems(newItems)
}

func (m RestoreModel) countSelectedFiles() int {
	count := 0
	for _, selected := range m.selectedFiles {
		if selected {
			count++
		}
	}
	return count
}

func (m RestoreModel) getSelectedFilePaths() []string {
	var paths []string
	for path, selected := range m.selectedFiles {
		if selected {
			paths = append(paths, path)
		}
	}
	return paths
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
		m.restoreStatus = "Loading files..."
		m.restoreError = ""
		m.passwordAttempts = 0
		return m, m.loadFiles(m.selectedBackup, m.password)

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

	case filesLoadedMsg:
		items := make([]list.Item, len(msg.files))
		m.selectedFiles = make(map[string]bool)
		for i, entry := range msg.files {
			items[i] = fileItem{
				path:     entry.Path,
				size:     int64(len(entry.Content)),
				selected: false,
			}
			m.selectedFiles[entry.Path] = false
		}
		m.fileList.SetItems(items)
		m.restoreStatus = fmt.Sprintf("Loaded %d files", len(msg.files))
		m.restoreError = ""
		return m, nil

	case diffLoadedMsg:
		m.currentDiff = msg.diff
		m.diffFile = msg.file
		m.viewport.SetContent(msg.diff)
		m.viewport.GotoTop()
		m.phase = 4
		m.restoreStatus = ""
		m.restoreError = ""
		return m, nil

	case diffErrorMsg:
		m.restoreError = fmt.Sprintf("Failed to load diff: %v", msg.err)
		m.restoreStatus = ""
		return m, nil

	case restoreCompleteMsg:
		m.restoreResult = msg.result
		m.phase = 5
		m.restoreStatus = ""
		m.restoreError = ""
		return m, nil

	case restoreErrorMsg:
		m.restoreError = fmt.Sprintf("Restore failed: %v", msg.err)
		m.restoreResult = nil
		m.phase = 5
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
		} else if m.phase == 2 {
			switch msg.String() {
			case " ":
				if item := m.fileList.SelectedItem(); item != nil {
					fi := item.(fileItem)
					m.selectedFiles[fi.path] = !m.selectedFiles[fi.path]
					m.updateFileListSelection()
				}
			case "a":
				for path := range m.selectedFiles {
					m.selectedFiles[path] = true
				}
				m.updateFileListSelection()
			case "n":
				for path := range m.selectedFiles {
					m.selectedFiles[path] = false
				}
				m.updateFileListSelection()
			case "d":
				if item := m.fileList.SelectedItem(); item != nil {
					fi := item.(fileItem)
					m.restoreStatus = "Loading diff..."
					m.restoreError = ""
					return m, m.loadDiff(fi.path)
				}
			case "enter":
				selectedCount := m.countSelectedFiles()
				if selectedCount == 0 {
					m.restoreError = "Select at least one file"
				} else {
					m.phase = 3
					m.restoreStatus = fmt.Sprintf("Restoring %d files...", selectedCount)
					m.restoreError = ""
					return m, m.runRestore()
				}
			case "esc":
				m.phase = 0
				m.selectedFiles = make(map[string]bool)
				m.password = ""
				m.restoreError = ""
				m.restoreStatus = ""
				m.passwordInput.SetValue("")
				m.passwordInput.Blur()
			default:
				var cmd tea.Cmd
				m.fileList, cmd = m.fileList.Update(msg)
				return m, cmd
			}
		} else if m.phase == 4 {
			switch msg.String() {
			case "j", "down":
				m.viewport.LineDown(1)
			case "k", "up":
				m.viewport.LineUp(1)
			case "g":
				m.viewport.GotoTop()
			case "G":
				m.viewport.GotoBottom()
			case "esc":
				m.phase = 2
				m.currentDiff = ""
				m.diffFile = ""
				m.restoreStatus = ""
				m.restoreError = ""
			default:
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		} else if m.phase == 5 {
			m.phase = 0
			m.restoreResult = nil
			m.selectedFiles = make(map[string]bool)
			m.password = ""
			m.restoreError = ""
			m.restoreStatus = ""
			m.passwordInput.SetValue("")
			m.passwordInput.Blur()
			return m, m.refreshBackups()
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the restore view
func (m RestoreModel) View() string {
	var s strings.Builder

	styles := DefaultStyles()

	// Phase 0: Backup list selection
	if m.phase == 0 {
		s.WriteString(m.backupList.View())
		s.WriteString("\n")

		if m.restoreStatus != "" {
			s.WriteString(styles.Success.Render(m.restoreStatus) + "\n")
		}
		if m.restoreError != "" {
			s.WriteString(styles.Error.Render(m.restoreError) + "\n")
		}

		s.WriteString(styles.Help.Render("↑/↓: navigate | Enter: select | r: refresh"))
		return s.String()
	}

	// Phase 1: Password entry
	if m.phase == 1 {
		s.WriteString(styles.Title.Render("Enter Password") + "\n\n")
		s.WriteString(fmt.Sprintf("Backup: %s\n\n", filepath.Base(m.selectedBackup)))
		s.WriteString(m.passwordInput.View() + "\n\n")

		if m.restoreStatus != "" {
			s.WriteString(styles.Success.Render(m.restoreStatus) + "\n")
		}
		if m.restoreError != "" {
			s.WriteString(styles.Error.Render(m.restoreError) + "\n")
		}

		s.WriteString(styles.Help.Render("Enter: validate | Esc: back"))
		return s.String()
	}

	// Phase 2: File selection
	if m.phase == 2 {
		s.WriteString(styles.Title.Render("Select Files to Restore") + "\n\n")

		selectedCount := m.countSelectedFiles()
		totalCount := len(m.selectedFiles)
		s.WriteString(styles.Value.Render(fmt.Sprintf("%d of %d files selected", selectedCount, totalCount)) + "\n\n")

		s.WriteString(m.fileList.View())
		s.WriteString("\n")

		if m.restoreStatus != "" {
			s.WriteString(styles.Success.Render(m.restoreStatus) + "\n")
		}
		if m.restoreError != "" {
			s.WriteString(styles.Error.Render(m.restoreError) + "\n")
		}

		s.WriteString(styles.Help.Render("Space: toggle | a: all | n: none | d: diff | Enter: restore | Esc: back"))
		return s.String()
	}

	// Phase 3: Restoring
	if m.phase == 3 {
		s.WriteString(styles.Title.Render("Restoring...") + "\n\n")
		s.WriteString(m.restoreStatus)
		return s.String()
	}

	// Phase 4: Diff preview
	if m.phase == 4 {
		s.WriteString(styles.Title.Render("Diff Preview") + "\n")
		s.WriteString(fmt.Sprintf("File: %s\n\n", m.diffFile))

		viewportStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Width(m.width - 4).
			Height(m.height - 10)

		s.WriteString(viewportStyle.Render(m.viewport.View()) + "\n")

		if m.restoreError != "" {
			s.WriteString(styles.Error.Render(m.restoreError) + "\n")
		}

		s.WriteString(styles.Help.Render("j/k or ↑/↓: scroll | g/G: top/bottom | Esc: back"))
		return s.String()
	}

	// Phase 5: Results
	if m.phase == 5 {
		s.WriteString(styles.Title.Render("Restore Complete") + "\n\n")

		if m.restoreError != "" {
			s.WriteString(styles.Error.Render(m.restoreError) + "\n\n")
		} else if m.restoreResult != nil {
			s.WriteString(styles.Success.Render(fmt.Sprintf("✓ Restored %d files", m.restoreResult.FilesRestored)) + "\n")

			if len(m.restoreResult.BackupFiles) > 0 {
				s.WriteString(fmt.Sprintf("  %d .bak files created\n", len(m.restoreResult.BackupFiles)))
			}
			if m.restoreResult.FilesSkipped > 0 {
				s.WriteString(fmt.Sprintf("  %d files skipped\n", m.restoreResult.FilesSkipped))
			}

			s.WriteString("\n")

			if len(m.restoreResult.RestoredFiles) > 0 {
				s.WriteString("Restored files:\n")
				for _, f := range m.restoreResult.RestoredFiles {
					s.WriteString(fmt.Sprintf("  • %s\n", f))
				}
				s.WriteString("\n")
			}

			if len(m.restoreResult.BackupFiles) > 0 {
				s.WriteString("Backup files created:\n")
				for _, f := range m.restoreResult.BackupFiles {
					s.WriteString(fmt.Sprintf("  • %s\n", f))
				}
				s.WriteString("\n")
			}
		}

		s.WriteString(styles.Help.Render("Press any key to continue"))
		return s.String()
	}

	s.WriteString(styles.Title.Render("Restore") + "\n\n")
	s.WriteString("Phase " + fmt.Sprintf("%d", m.phase) + " (implementation pending)")

	return s.String()
}

func (m RestoreModel) HelpBindings() []HelpEntry {
	switch m.phase {
	case 0:
		return []HelpEntry{
			{"Enter", "Select backup"},
			{"r", "Refresh"},
			{"↑/↓", "Navigate"},
		}
	case 1:
		return []HelpEntry{
			{"Enter", "Submit password"},
			{"Esc", "Back"},
		}
	case 2:
		return []HelpEntry{
			{"Space", "Toggle file"},
			{"a", "Select all"},
			{"n", "Select none"},
			{"d", "View diff"},
			{"Enter", "Restore"},
			{"Esc", "Back"},
		}
	case 4:
		return []HelpEntry{
			{"j/k", "Scroll"},
			{"g/G", "Top/Bottom"},
			{"Esc", "Back"},
		}
	case 5:
		return []HelpEntry{
			{"any key", "Continue"},
		}
	default:
		return nil
	}
}
