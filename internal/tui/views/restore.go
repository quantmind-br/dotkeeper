package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/history"
	"github.com/diogo/dotkeeper/internal/pathutil"
	"github.com/diogo/dotkeeper/internal/restore"
	"github.com/diogo/dotkeeper/internal/tui/components"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

// restorePhase represents the current phase of the restore workflow.
type restorePhase int

const (
	phaseBackupList  restorePhase = iota // 0: select backup
	phasePassword                        // 1: enter password
	phaseFileSelect                      // 2: select files
	phaseRestoring                       // 3: restoring in progress
	phaseDiffPreview                     // 4: diff preview
	phaseResults                         // 5: results display
)

const restoreViewChromeHeight = 5
const maxPasswordAttempts = 3

// RestoreModel represents the restore view
type RestoreModel struct {
	ctx              *ProgramContext
	backupList       list.Model
	phase            restorePhase
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
	spinner          spinner.Model
	loading          bool
}

type passwordValidMsg struct{}

// passwordInvalidMsg is consolidated to ErrorMsg with Source="restore-password".
type passwordInvalidMsg = ErrorMsg

type filesLoadedMsg struct {
	files []restore.FileEntry
}

type diffLoadedMsg struct {
	diff string
	file string
}

type restoreCompleteMsg struct {
	result *restore.RestoreResult
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
func NewRestore(ctx *ProgramContext) RestoreModel {
	l := styles.NewMinimalList()

	ti := components.NewPasswordInput("Enter password for decryption")
	ti.Width = 40 // Default width, will be adjusted on WindowSizeMsg

	fl := styles.NewMinimalList()

	vp := viewport.New(0, 0)

	s := spinner.New()
	s.Spinner = spinner.Dot

	return RestoreModel{
		ctx:           ensureProgramContext(ctx),
		backupList:    l,
		passwordInput: ti,
		fileList:      fl,
		selectedFiles: make(map[string]bool),
		viewport:      vp,
		phase:         phaseBackupList,
		spinner:       s,
	}
}

// Init initializes the restore view
func (m RestoreModel) Init() tea.Cmd {
	return tea.Batch(m.refreshBackups(), m.spinner.Tick)
}

// Refresh reloads the backup list
func (m RestoreModel) Refresh() tea.Cmd {
	return m.refreshBackups()
}

// refreshBackups scans the backup directory and loads available backups
func (m RestoreModel) refreshBackups() tea.Cmd {
	return func() tea.Msg {
		if m.ctx.Config == nil {
			return backupsLoadedMsg([]list.Item{})
		}
		return backupsLoadedMsg(LoadBackupItems(m.ctx.Config.BackupDir))
	}
}

func (m RestoreModel) validatePassword(backupPath, password string) tea.Cmd {
	return func() tea.Msg {
		err := restore.ValidateBackup(backupPath, password)
		if err != nil {
			return passwordInvalidMsg{Source: "restore-password", Err: err}
		}
		return passwordValidMsg{}
	}
}

func (m RestoreModel) loadFiles(backupPath, password string) tea.Cmd {
	return func() tea.Msg {
		entries, err := restore.ListBackupContents(backupPath, password)
		if err != nil {
			return passwordInvalidMsg{Source: "restore-password", Err: fmt.Errorf("failed to load files: %w", err)}
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
			return ErrorMsg{Source: "restore-diff", Err: err}
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
			return ErrorMsg{Source: "restore", Err: err}
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

func (m RestoreModel) handleBackupListKey(msg tea.KeyMsg) (RestoreModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if item := m.backupList.SelectedItem(); item != nil {
			selected := item.(backupItem)
			if m.ctx.Config == nil {
				m.restoreError = "Missing config"
				return m, nil
			}
			backupPath := filepath.Join(pathutil.ExpandHome(m.ctx.Config.BackupDir), selected.name+".tar.gz.enc")
			m.selectedBackup = backupPath
			m.passwordInput.SetValue("")
			m.restoreError = ""
			m.phase = phasePassword
			m.passwordInput.Focus()
			return m, textinput.Blink
		}
	case "r":
		m.loading = true
		return m, m.refreshBackups()
	default:
		var cmd tea.Cmd
		m.backupList, cmd = m.backupList.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m RestoreModel) handlePasswordKey(msg tea.KeyMsg) (RestoreModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.passwordInput.Value() != "" {
			m.loading = true
			m.restoreStatus = "Validating password..."
			m.restoreError = ""
			return m, m.validatePassword(m.selectedBackup, m.passwordInput.Value())
		}
	case "esc":
		m.phase = phaseBackupList
		m.passwordInput.SetValue("")
		m.passwordAttempts = 0
		m.restoreError = ""
		m.loading = false
		m.passwordInput.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		m.passwordInput, cmd = m.passwordInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m RestoreModel) handleFileSelectKey(msg tea.KeyMsg) (RestoreModel, tea.Cmd) {
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
			m.loading = true
			m.restoreStatus = "Loading diff..."
			m.restoreError = ""
			return m, m.loadDiff(fi.path)
		}
	case "enter":
		selectedCount := m.countSelectedFiles()
		if selectedCount == 0 {
			m.restoreError = "Select at least one file"
		} else {
			m.loading = true
			m.phase = phaseRestoring
			m.restoreStatus = fmt.Sprintf("Restoring %d files...", selectedCount)
			m.restoreError = ""
			return m, m.runRestore()
		}
	case "esc":
		m.phase = phaseBackupList
		m.selectedFiles = make(map[string]bool)
		m.password = ""
		m.restoreError = ""
		m.loading = false
		m.passwordInput.SetValue("")
		m.passwordInput.Blur()
	default:
		var cmd tea.Cmd
		m.fileList, cmd = m.fileList.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m RestoreModel) handleDiffPreviewKey(msg tea.KeyMsg) (RestoreModel, tea.Cmd) {
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
		m.phase = phaseFileSelect
		m.currentDiff = ""
		m.diffFile = ""
		m.restoreStatus = ""
		m.restoreError = ""
	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m RestoreModel) handleResultsKey() (RestoreModel, tea.Cmd) {
	m.phase = phaseBackupList
	m.restoreResult = nil
	m.selectedFiles = make(map[string]bool)
	m.password = ""
	m.restoreError = ""
	m.restoreStatus = ""
	m.passwordInput.SetValue("")
	m.passwordInput.Blur()
	m.loading = true
	return m, m.refreshBackups()
}

// Update handles messages
func (m RestoreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
		listHeight := msg.Height - restoreViewChromeHeight
		if listHeight < 0 {
			listHeight = 0
		}
		m.backupList.SetSize(msg.Width, listHeight)
		m.fileList.SetSize(msg.Width, listHeight)
		// Viewport needs to account for its border styling (RoundedBorder adds 1 char on each side)
		vpBorderW := 2 // left + right border
		vpBorderH := 2 // top + bottom border
		m.viewport.Width = msg.Width - vpBorderW
		m.viewport.Height = msg.Height - vpBorderH
		if m.viewport.Width < 0 {
			m.viewport.Width = 0
		}
		if m.viewport.Height < 0 {
			m.viewport.Height = 0
		}
		// Responsive password input width
		pw := msg.Width - 6
		if pw < 20 {
			pw = 20
		}
		if pw > 60 {
			pw = 60
		}
		m.passwordInput.Width = pw

	case backupsLoadedMsg:
		m.loading = false
		m.backupList.SetItems([]list.Item(msg))
		return m, nil

	case passwordValidMsg:
		m.password = m.passwordInput.Value()
		m.phase = phaseFileSelect
		m.restoreStatus = "Loading files..."
		m.restoreError = ""
		m.passwordAttempts = 0
		return m, m.loadFiles(m.selectedBackup, m.password)

	case passwordInvalidMsg:
		if msg.Source == "restore-password" {
			m.loading = false
			m.passwordAttempts++
			if m.passwordAttempts >= maxPasswordAttempts {
				m.restoreError = "Too many failed attempts"
				m.phase = phaseBackupList
				m.passwordAttempts = 0
				m.passwordInput.SetValue("")
				m.passwordInput.Blur()
			} else {
				m.restoreError = fmt.Sprintf("Invalid password (attempt %d/%d): %v", m.passwordAttempts, maxPasswordAttempts, msg.Err)
				m.passwordInput.SetValue("")
			}
			m.restoreStatus = ""
			return m, nil
		} else if msg.Source == "restore-diff" {
			m.loading = false
			m.restoreError = fmt.Sprintf("Failed to load diff: %v", msg.Err)
			m.restoreStatus = ""
			return m, nil
		} else if msg.Source == "restore" {
			m.loading = false
			m.restoreError = fmt.Sprintf("Restore failed: %v", msg.Err)
			m.restoreResult = nil
			m.phase = phaseResults
			m.restoreStatus = ""
			if m.ctx.Store != nil {
				_ = m.ctx.Store.Append(history.EntryFromRestoreError(msg.Err, m.selectedBackup))
			}
			return m, nil
		}

	case filesLoadedMsg:
		m.loading = false
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
		m.loading = false
		m.currentDiff = msg.diff
		m.diffFile = msg.file
		m.viewport.SetContent(msg.diff)
		m.viewport.GotoTop()
		m.phase = phaseDiffPreview
		m.restoreStatus = ""
		m.restoreError = ""
		return m, nil

	case restoreCompleteMsg:
		m.loading = false
		m.restoreResult = msg.result
		m.phase = phaseResults
		m.restoreStatus = ""
		m.restoreError = ""
		if m.ctx.Store != nil {
			_ = m.ctx.Store.Append(history.EntryFromRestoreResult(msg.result, m.selectedBackup))
		}
		return m, nil

	case tea.KeyMsg:
		var cmd tea.Cmd
		switch m.phase {
		case phaseBackupList:
			m, cmd = m.handleBackupListKey(msg)
			return m, cmd
		case phasePassword:
			m, cmd = m.handlePasswordKey(msg)
			return m, cmd
		case phaseFileSelect:
			m, cmd = m.handleFileSelectKey(msg)
			return m, cmd
		case phaseDiffPreview:
			m, cmd = m.handleDiffPreviewKey(msg)
			return m, cmd
		case phaseResults:
			m, cmd = m.handleResultsKey()
			return m, cmd
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the restore view
func (m RestoreModel) View() string {
	var s strings.Builder

	st := styles.DefaultStyles()

	// Phase 0: Backup list selection
	if m.phase == phaseBackupList {
		if m.loading {
			return lipgloss.JoinVertical(lipgloss.Center,
				"\n",
				m.spinner.View(),
				"\nLoading backups...",
			)
		}
		s.WriteString(m.backupList.View())
		s.WriteString("\n")
		s.WriteString(RenderStatusBar(m.ctx.Width, m.restoreStatus, m.restoreError, ""))
		return s.String()
	}

	// Phase 1: Password entry
	if m.phase == phasePassword {
		if m.loading {
			return lipgloss.JoinVertical(lipgloss.Center,
				"\n",
				m.spinner.View(),
				"\nValidating password...",
			)
		}
		s.WriteString(st.Title.Render("Enter Password") + "\n\n")
		s.WriteString(fmt.Sprintf("Backup: %s\n\n", filepath.Base(m.selectedBackup)))
		s.WriteString(m.passwordInput.View() + "\n\n")
		s.WriteString(RenderStatusBar(m.ctx.Width, m.restoreStatus, m.restoreError, ""))
		return s.String()
	}

	// Phase 2: File selection
	if m.phase == phaseFileSelect {
		if m.loading {
			return lipgloss.JoinVertical(lipgloss.Center,
				"\n",
				m.spinner.View(),
				"\n"+m.restoreStatus+"...",
			)
		}
		s.WriteString(st.Title.Render("Select Files to Restore") + "\n\n")

		selectedCount := m.countSelectedFiles()
		totalCount := len(m.selectedFiles)
		s.WriteString(st.Value.Render(fmt.Sprintf("%d of %d files selected", selectedCount, totalCount)) + "\n\n")

		s.WriteString(m.fileList.View())
		s.WriteString("\n")
		s.WriteString(RenderStatusBar(m.ctx.Width, m.restoreStatus, m.restoreError, ""))
		return s.String()
	}

	// Phase 3: Restoring
	if m.phase == phaseRestoring {
		return lipgloss.JoinVertical(lipgloss.Center,
			"\n",
			m.spinner.View(),
			"\n"+m.restoreStatus+"...",
		)
	}

	// Phase 4: Diff preview
	if m.phase == phaseDiffPreview {
		if m.loading {
			return lipgloss.JoinVertical(lipgloss.Center,
				"\n",
				m.spinner.View(),
				"\nLoading diff...",
			)
		}
		s.WriteString(st.Title.Render("Diff Preview") + "\n")
		s.WriteString(fmt.Sprintf("File: %s\n\n", m.diffFile))

		// Use viewport dimensions directly (already account for border in Update())
		viewportStyle := st.ViewportBorder.Copy().
			Width(m.viewport.Width).
			Height(m.viewport.Height)

		s.WriteString(viewportStyle.Render(m.viewport.View()) + "\n")
		s.WriteString(RenderStatusBar(m.ctx.Width, "", m.restoreError, ""))
		return s.String()
	}

	// Phase 5: Results
	if m.phase == phaseResults {
		s.WriteString(st.Title.Render("Restore Complete") + "\n\n")

		if m.restoreError != "" {
			s.WriteString(st.Error.Render(m.restoreError) + "\n\n")
		} else if m.restoreResult != nil {
			s.WriteString(st.Success.Render(fmt.Sprintf("✓ Restored %d files", m.restoreResult.FilesRestored)) + "\n")

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

		return s.String()
	}

	s.WriteString(st.Title.Render("Restore") + "\n\n")
	s.WriteString("Phase " + fmt.Sprintf("%d", m.phase) + " (implementation pending)")

	return s.String()
}

func (m RestoreModel) HelpBindings() []HelpEntry {
	switch m.phase {
	case phaseBackupList:
		return []HelpEntry{
			{"Enter", "Select backup"},
			{"r", "Refresh"},
			{"↑/↓", "Navigate"},
		}
	case phasePassword:
		return []HelpEntry{
			{"Enter", "Submit password"},
			{"Esc", "Back"},
		}
	case phaseFileSelect:
		return []HelpEntry{
			{"Space", "Toggle file"},
			{"a", "Select all"},
			{"n", "Select none"},
			{"d", "View diff"},
			{"Enter", "Restore"},
			{"Esc", "Back"},
		}
	case phaseDiffPreview:
		return []HelpEntry{
			{"j/k", "Scroll"},
			{"g/G", "Top/Bottom"},
			{"Esc", "Back"},
		}
	case phaseResults:
		return []HelpEntry{
			{"any key", "Continue"},
		}
	default:
		return nil
	}
}

func (m RestoreModel) StatusHelpText() string {
	switch m.phase {
	case phaseBackupList:
		return "↑/↓: navigate | Enter: select | r: refresh"
	case phasePassword:
		return "Enter: validate | Esc: back"
	case phaseFileSelect:
		return "Space: toggle | a: all | n: none | d: diff | Enter: restore | Esc: back"
	case phaseRestoring:
		return "Please wait..."
	case phaseDiffPreview:
		return "j/k or ↑/↓: scroll | g/G: top/bottom | Esc: back"
	case phaseResults:
		return "Press any key to continue"
	default:
		return ""
	}
}

func (m RestoreModel) IsInputActive() bool {
	return m.phase == phasePassword
}
