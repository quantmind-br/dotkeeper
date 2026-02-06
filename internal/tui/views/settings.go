package views

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/pathutil"
	"github.com/diogo/dotkeeper/internal/tui/components"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

// Keep lipgloss import for cursor/prompt styles (component configuration)

type settingsState int

const (
	stateListNavigating settingsState = iota
	stateEditingField
	stateBrowsingFiles
	stateBrowsingFolders
	stateEditingSubItem
	stateFilePickerActive
)

type settingItem struct {
	label string
	value string
	index int
}

func (i settingItem) Title() string       { return i.label }
func (i settingItem) Description() string { return i.value }
func (i settingItem) FilterValue() string { return i.label }

type subSettingItem struct {
	title    string
	desc     string
	index    int
	isAdd    bool
	disabled bool
}

func (i subSettingItem) Title() string       { return i.title }
func (i subSettingItem) Description() string { return i.desc }
func (i subSettingItem) FilterValue() string { return i.title }

// SettingsModel represents the settings view
type SettingsModel struct {
	config        *config.Config
	width         int
	height        int
	state         settingsState
	mainList      list.Model
	filesList     list.Model
	foldersList   list.Model
	pathCompleter components.PathCompleter
	status        string
	errMsg        string

	inspecting  bool
	inspectInfo string

	editingFieldIndex int
	subEditParent     settingsState
	subEditIndex      int
	filePicker        filepicker.Model
	filePickerParent  settingsState
}

// NewSettings creates a new settings model
func NewSettings(cfg *config.Config) SettingsModel {
	pc := components.NewPathCompleter()
	pc.Input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	pc.Input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	mainList := list.New([]list.Item{}, styles.NewListDelegate(), 80, 18)
	mainList.SetShowTitle(false)
	mainList.SetShowHelp(false)
	mainList.SetShowStatusBar(false)
	mainList.SetShowPagination(false)
	mainList.SetFilteringEnabled(false)

	filesList := list.New([]list.Item{}, styles.NewListDelegate(), 80, 18)
	filesList.SetShowTitle(false)
	filesList.SetShowHelp(false)
	filesList.SetShowStatusBar(false)
	filesList.SetShowPagination(false)
	filesList.SetFilteringEnabled(false)

	foldersList := list.New([]list.Item{}, styles.NewListDelegate(), 80, 18)
	foldersList.SetShowTitle(false)
	foldersList.SetShowHelp(false)
	foldersList.SetShowStatusBar(false)
	foldersList.SetShowPagination(false)
	foldersList.SetFilteringEnabled(false)

	fp := filepicker.New()
	home, _ := os.UserHomeDir()
	if home != "" {
		fp.CurrentDirectory = home
	}
	fp.ShowHidden = true

	m := SettingsModel{
		config:        cfg,
		width:         80,
		height:        24,
		state:         stateListNavigating,
		mainList:      mainList,
		filesList:     filesList,
		foldersList:   foldersList,
		pathCompleter: pc,
		filePicker:    fp,
	}
	m.refreshMainList()
	m.refreshFilesList()
	m.refreshFoldersList()
	m.resizeLists()

	return m
}

// Init initializes the settings view
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Route messages to filepicker when active (for non-standard msgs)
	if m.state == stateFilePickerActive {
		switch msg.(type) {
		case tea.KeyMsg, tea.WindowSizeMsg:
			// Let the main switch handle these
		default:
			var cmd tea.Cmd
			m.filePicker, cmd = m.filePicker.Update(msg)
			// Check for file selection
			if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
				if m.filePickerParent == stateBrowsingFiles {
					m.config.Files = append(m.config.Files, path)
					m.refreshFilesList()
				} else {
					m.config.Folders = append(m.config.Folders, path)
					m.refreshFoldersList()
				}
				m.refreshMainList()
				m.state = m.filePickerParent
				return m, nil
			}
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case components.CompletionResultMsg:
		var cmd tea.Cmd
		m.pathCompleter, cmd = m.pathCompleter.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeLists()

	case tea.KeyMsg:
		switch m.state {
		case stateEditingField:
			return m.handleEditingFieldInput(msg)
		case stateEditingSubItem:
			return m.handleEditingSubItemInput(msg)
		case stateListNavigating:
			return m.handleEditModeInput(msg)
		case stateBrowsingFiles:
			return m.handleBrowsingFilesInput(msg)
		case stateBrowsingFolders:
			return m.handleBrowsingFoldersInput(msg)
		case stateFilePickerActive:
			return m.handleFilePickerInput(msg)
		}
		return m, nil
	}
	return m, nil
}

func (m SettingsModel) handleFilePickerInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.state = m.filePickerParent
		return m, nil
	}
	// Forward to filepicker
	var cmd tea.Cmd
	m.filePicker, cmd = m.filePicker.Update(msg)
	return m, cmd
}

// handleEditModeInput handles input when in edit mode
func (m SettingsModel) handleEditModeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, nil

	case "enter":
		selected, ok := m.mainList.SelectedItem().(settingItem)
		if !ok {
			return m, nil
		}

		switch selected.index {
		case 0, 1, 4:
			m.editingFieldIndex = selected.index
			m.startEditingField()
			m.state = stateEditingField
			m.resizeLists()
		case 2:
			m.state = stateBrowsingFiles
		case 3:
			m.state = stateBrowsingFolders
		case 5:
			m.config.Notifications = !m.config.Notifications
			m.refreshMainList()
		}
		return m, nil

	case "a":
		selected, ok := m.mainList.SelectedItem().(settingItem)
		if !ok {
			return m, nil
		}
		if selected.index == 2 {
			m.startEditingSubItem(stateBrowsingFiles, len(m.config.Files), "")
		} else if selected.index == 3 {
			m.startEditingSubItem(stateBrowsingFolders, len(m.config.Folders), "")
		}
		return m, nil

	case "s":
		m.saveConfig()
		return m, nil
	}

	var cmd tea.Cmd
	m.mainList, cmd = m.mainList.Update(msg)
	return m, cmd
}

func (m SettingsModel) handleBrowsingFilesInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateListNavigating
		return m, nil
	case " ":
		selected, ok := m.filesList.SelectedItem().(subSettingItem)
		if !ok || selected.isAdd {
			return m, nil
		}
		path := m.config.Files[selected.index]
		// Toggle in DisabledFiles
		found := false
		for i, d := range m.config.DisabledFiles {
			if d == path {
				m.config.DisabledFiles = append(m.config.DisabledFiles[:i], m.config.DisabledFiles[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			m.config.DisabledFiles = append(m.config.DisabledFiles, path)
		}
		m.refreshFilesList()
		return m, nil
	case "i":
		selected, ok := m.filesList.SelectedItem().(subSettingItem)
		if !ok || selected.isAdd {
			return m, nil
		}
		path := m.config.Files[selected.index]
		m.inspecting = true
		m.inspectInfo = getInspectInfo(path)
		return m, nil
	case "enter":
		selected, ok := m.filesList.SelectedItem().(subSettingItem)
		if !ok {
			return m, nil
		}
		if selected.isAdd {
			m.startEditingSubItem(stateBrowsingFiles, len(m.config.Files), "")
		} else if selected.index >= 0 && selected.index < len(m.config.Files) {
			m.startEditingSubItem(stateBrowsingFiles, selected.index, m.config.Files[selected.index])
		}
		return m, nil
	case "a":
		m.startEditingSubItem(stateBrowsingFiles, len(m.config.Files), "")
		return m, nil
	case "b":
		m.filePickerParent = stateBrowsingFiles
		m.state = stateFilePickerActive
		return m, m.filePicker.Init()
	case "d":
		selected, ok := m.filesList.SelectedItem().(subSettingItem)
		if !ok || selected.isAdd || selected.index < 0 || selected.index >= len(m.config.Files) {
			return m, nil
		}
		m.config.Files = append(m.config.Files[:selected.index], m.config.Files[selected.index+1:]...)
		m.refreshFilesList()
		m.refreshMainList()
		if selected.index < len(m.filesList.Items()) {
			m.filesList.Select(selected.index)
		}
		return m, nil
	case "s":
		m.saveConfig()
		return m, nil
	}

	// If inspecting, any key dismisses
	if m.inspecting {
		m.inspecting = false
		return m, nil
	}

	var cmd tea.Cmd
	m.filesList, cmd = m.filesList.Update(msg)
	return m, cmd
}

func (m SettingsModel) handleBrowsingFoldersInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateListNavigating
		return m, nil
	case "enter":
		selected, ok := m.foldersList.SelectedItem().(subSettingItem)
		if !ok {
			return m, nil
		}
		if selected.isAdd {
			m.startEditingSubItem(stateBrowsingFolders, len(m.config.Folders), "")
		} else if selected.index >= 0 && selected.index < len(m.config.Folders) {
			m.startEditingSubItem(stateBrowsingFolders, selected.index, m.config.Folders[selected.index])
		}
		return m, nil
	case "a":
		m.startEditingSubItem(stateBrowsingFolders, len(m.config.Folders), "")
		return m, nil
	case "b":
		m.filePickerParent = stateBrowsingFolders
		m.state = stateFilePickerActive
		return m, m.filePicker.Init()
	case "d":
		selected, ok := m.foldersList.SelectedItem().(subSettingItem)
		if !ok || selected.isAdd || selected.index < 0 || selected.index >= len(m.config.Folders) {
			return m, nil
		}
		m.config.Folders = append(m.config.Folders[:selected.index], m.config.Folders[selected.index+1:]...)
		m.refreshFoldersList()
		m.refreshMainList()
		if selected.index < len(m.foldersList.Items()) {
			m.foldersList.Select(selected.index)
		}
		return m, nil
	case "s":
		m.saveConfig()
		return m, nil
	}

	var cmd tea.Cmd
	m.foldersList, cmd = m.foldersList.Update(msg)
	return m, cmd
}

// handleEditingFieldInput handles input when actively editing a field
func (m SettingsModel) handleEditingFieldInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateListNavigating
		m.pathCompleter.Input.Blur()
		m.pathCompleter.Input.SetValue("")
		m.resizeLists()
		return m, nil

	case "enter":
		value := strings.TrimSpace(m.pathCompleter.Input.Value())
		m.saveFieldValue(value)
		m.refreshMainList()
		m.state = stateListNavigating
		m.pathCompleter.Input.Blur()
		m.pathCompleter.Input.SetValue("")
		m.resizeLists()
		return m, nil

	default:
		var cmd tea.Cmd
		m.pathCompleter, cmd = m.pathCompleter.Update(msg)
		return m, cmd
	}
}

func (m SettingsModel) handleEditingSubItemInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = m.subEditParent
		m.pathCompleter.Input.Blur()
		m.pathCompleter.Input.SetValue("")
		m.resizeLists()
		return m, nil

	case "enter":
		value := strings.TrimSpace(m.pathCompleter.Input.Value())
		if pathutil.IsGlobPattern(value) {
			results, err := pathutil.ResolveGlob(value, m.config.Exclude)
			if err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			if m.subEditParent == stateBrowsingFiles {
				m.config.Files = append(m.config.Files, results...)
			} else {
				m.config.Folders = append(m.config.Folders, results...)
			}
			m.status = fmt.Sprintf("Added %d paths from glob", len(results))
			m.refreshFilesList()
			m.refreshFoldersList()
			m.refreshMainList()
			m.state = m.subEditParent
			m.pathCompleter.Input.Blur()
			m.pathCompleter.Input.SetValue("")
			m.resizeLists()
			return m, nil
		}
		m.saveFieldValue(value)
		m.refreshFilesList()
		m.refreshFoldersList()
		m.refreshMainList()
		m.state = m.subEditParent
		m.pathCompleter.Input.Blur()
		m.pathCompleter.Input.SetValue("")
		m.resizeLists()
		return m, nil

	default:
		var cmd tea.Cmd
		m.pathCompleter, cmd = m.pathCompleter.Update(msg)
		return m, cmd
	}
}

// startEditingField initializes field editing
func (m *SettingsModel) startEditingField() {
	m.pathCompleter.Input.Focus()
	switch m.editingFieldIndex {
	case 0:
		m.pathCompleter.Input.SetValue(m.config.BackupDir)
	case 1:
		m.pathCompleter.Input.SetValue(m.config.GitRemote)
	case 4:
		m.pathCompleter.Input.SetValue(m.config.Schedule)
	default:
		m.pathCompleter.Input.SetValue("")
	}
}

func (m *SettingsModel) startEditingSubItem(parent settingsState, index int, value string) {
	m.subEditParent = parent
	m.subEditIndex = index
	m.state = stateEditingSubItem
	m.pathCompleter.Input.Focus()
	m.pathCompleter.Input.SetValue(value)
	m.resizeLists()
}

// saveFieldValue saves the edited field value
func (m *SettingsModel) saveFieldValue(value string) {
	if m.state == stateEditingSubItem && m.subEditParent == stateBrowsingFiles {
		// Check if empty
		if value == "" {
			return
		}
		// Validate file path
		expandedPath, err := ValidateFilePath(value)
		if err != nil {
			m.errMsg = err.Error()
			m.status = ""
			return
		}
		// Clear error and save expanded path
		m.errMsg = ""
		if m.subEditIndex < len(m.config.Files) {
			m.config.Files[m.subEditIndex] = expandedPath
		} else {
			m.config.Files = append(m.config.Files, expandedPath)
		}
	} else if m.state == stateEditingSubItem && m.subEditParent == stateBrowsingFolders {
		// Check if empty
		if value == "" {
			return
		}
		// Validate folder path
		expandedPath, err := ValidateFolderPath(value)
		if err != nil {
			m.errMsg = err.Error()
			m.status = ""
			return
		}
		// Clear error and save expanded path
		m.errMsg = ""
		if m.subEditIndex < len(m.config.Folders) {
			m.config.Folders[m.subEditIndex] = expandedPath
		} else {
			m.config.Folders = append(m.config.Folders, expandedPath)
		}
	} else {
		switch m.editingFieldIndex {
		case 0:
			m.config.BackupDir = pathutil.ExpandHome(value)
		case 1:
			m.config.GitRemote = value
		case 4:
			m.config.Schedule = value
		case 5:
			m.config.Notifications = value == "true"
		}
	}
}

func (m *SettingsModel) saveConfig() {
	if err := m.config.Save(); err != nil {
		m.errMsg = err.Error()
		m.status = ""
		return
	}
	m.status = "Config saved successfully!"
	m.errMsg = ""
}

func getInspectInfo(path string) string {
	expanded := pathutil.ExpandHome(path)
	info, err := os.Stat(expanded)
	if err != nil {
		return fmt.Sprintf("Path: %s\nStatus: NOT FOUND", path)
	}
	modTime := info.ModTime().Format("2006-01-02 15:04:05")
	perm := fmt.Sprintf("%o", info.Mode().Perm())
	fileType := "regular file"
	if info.IsDir() {
		fileType = "directory"
	}
	// Check if symlink
	linfo, _ := os.Lstat(expanded)
	if linfo != nil && linfo.Mode()&os.ModeSymlink != 0 {
		target, _ := os.Readlink(expanded)
		fileType = fmt.Sprintf("symlink → %s", target)
	}
	return fmt.Sprintf("Path: %s\nSize: %s | Modified: %s | Permissions: %s\nType: %s",
		path, pathutil.FormatSize(info.Size()), modTime, perm, fileType)
}

func (m *SettingsModel) refreshMainList() {
	schedule := m.config.Schedule
	if schedule == "" {
		schedule = "Not scheduled"
	}

	items := []list.Item{
		settingItem{label: "Backup Directory", value: m.config.BackupDir, index: 0},
		settingItem{label: "Git Remote", value: m.config.GitRemote, index: 1},
		settingItem{label: "Files", value: fmt.Sprintf("%d files", len(m.config.Files)), index: 2},
		settingItem{label: "Folders", value: fmt.Sprintf("%d folders", len(m.config.Folders)), index: 3},
		settingItem{label: "Schedule", value: schedule, index: 4},
		settingItem{label: "Notifications", value: fmt.Sprintf("%v", m.config.Notifications), index: 5},
	}

	selected := m.mainList.Index()
	m.mainList.SetItems(items)
	if selected >= len(items) {
		selected = len(items) - 1
	}
	if selected < 0 {
		selected = 0
	}
	m.mainList.Select(selected)
}

func (m *SettingsModel) refreshFilesList() {
	disabledSet := make(map[string]bool)
	for _, d := range m.config.DisabledFiles {
		disabledSet[d] = true
	}

	items := make([]list.Item, 0, len(m.config.Files)+1)
	for i, filePath := range m.config.Files {
		desc := pathutil.GetPathDesc(filePath)
		isDisabled := disabledSet[filePath]
		if isDisabled {
			desc = "[disabled] " + desc
		}
		items = append(items, subSettingItem{title: filePath, desc: desc, index: i, disabled: isDisabled})
	}
	items = append(items, subSettingItem{title: "[+] Add new file", desc: "", index: len(m.config.Files), isAdd: true})

	selected := m.filesList.Index()
	m.filesList.SetItems(items)
	if selected >= len(items) {
		selected = len(items) - 1
	}
	if selected < 0 {
		selected = 0
	}
	m.filesList.Select(selected)
}

func (m *SettingsModel) refreshFoldersList() {
	disabledSet := make(map[string]bool)
	for _, d := range m.config.DisabledFolders {
		disabledSet[d] = true
	}

	items := make([]list.Item, 0, len(m.config.Folders)+1)
	for i, folderPath := range m.config.Folders {
		desc := pathutil.GetPathDesc(folderPath)
		isDisabled := disabledSet[folderPath]
		if isDisabled {
			desc = "[disabled] " + desc
		}
		items = append(items, subSettingItem{title: folderPath, desc: desc, index: i, disabled: isDisabled})
	}
	items = append(items, subSettingItem{title: "[+] Add new folder", desc: "", index: len(m.config.Folders), isAdd: true})

	selected := m.foldersList.Index()
	m.foldersList.SetItems(items)
	if selected >= len(items) {
		selected = len(items) - 1
	}
	if selected < 0 {
		selected = 0
	}
	m.foldersList.Select(selected)
}

func (m *SettingsModel) resizeLists() {
	width := m.width
	if width <= 0 {
		width = 80
	}

	height := m.height - styles.ViewChromeHeight
	if m.state == stateEditingField || m.state == stateEditingSubItem {
		height -= 2
	}
	if height < 6 {
		height = 6
	}

	if m.state == stateFilePickerActive {
		m.filePicker.Height = height
	}

	m.mainList.SetSize(width, height)
	m.filesList.SetSize(width, height)
	m.foldersList.SetSize(width, height)
}

// View renders the settings view
func (m SettingsModel) View() string {
	var b strings.Builder

	styles := styles.DefaultStyles()

	b.WriteString(styles.Title.Render("Settings") + "\n\n")

	helpText := ""
	switch m.state {
	case stateListNavigating:
		b.WriteString(m.mainList.View())
		helpText = "↑/↓: Navigate | Enter: Edit | a: Add | s: Save | Esc: Exit"
	case stateEditingField:
		b.WriteString("Editing: " + m.pathCompleter.View() + "\n\n")
		b.WriteString(m.mainList.View())
		helpText = "Enter: Save field | Esc: Cancel"
	case stateBrowsingFiles:
		b.WriteString(styles.Subtitle.Render("Files") + "\n")
		b.WriteString(m.filesList.View())
		helpText = "↑/↓: Navigate | Enter: Edit | a: Type path | b: Browse | d: Delete | s: Save | Esc: Back"
	case stateBrowsingFolders:
		b.WriteString(styles.Subtitle.Render("Folders") + "\n")
		b.WriteString(m.foldersList.View())
		helpText = "↑/↓: Navigate | Enter: Edit | a: Type path | b: Browse | d: Delete | s: Save | Esc: Back"
	case stateEditingSubItem:
		title := "Editing File"
		listView := m.filesList.View()
		if m.subEditParent == stateBrowsingFolders {
			title = "Editing Folder"
			listView = m.foldersList.View()
		}
		b.WriteString(styles.Subtitle.Render(title) + "\n")
		b.WriteString("Value: " + m.pathCompleter.View() + "\n\n")
		b.WriteString(listView)
		helpText = "Enter: Save item | Esc: Cancel"
	case stateFilePickerActive:
		title := "Browse Files"
		if m.filePickerParent == stateBrowsingFolders {
			title = "Browse Folders"
		}
		b.WriteString(styles.Subtitle.Render(title) + "\n")
		b.WriteString(m.filePicker.View())
		helpText = "Enter: Select | ↑/↓: Navigate | Esc: Cancel"
	}

	b.WriteString("\n" + RenderStatusBar(m.width, m.status, m.errMsg, helpText))

	return b.String()
}

func (m SettingsModel) HelpBindings() []HelpEntry {
	switch m.state {
	case stateEditingField, stateEditingSubItem:
		return []HelpEntry{
			{"Enter", "Save field"},
			{"Esc", "Cancel edit"},
		}
	case stateListNavigating, stateBrowsingFiles, stateBrowsingFolders:
		return []HelpEntry{
			{"↑/↓", "Navigate"},
			{"Enter", "Edit field"},
			{"a", "Type path"},
			{"b", "Browse"},
			{"d", "Delete item"},
			{"s", "Save config"},
			{"Esc", "Exit edit"},
		}
	case stateFilePickerActive:
		return []HelpEntry{
			{"Enter", "Select"},
			{"Esc", "Cancel"},
		}
	default:
		return []HelpEntry{
			{"↑/↓", "Navigate"},
			{"Enter", "Edit field"},
			{"a", "Add item"},
			{"d", "Delete item"},
			{"s", "Save config"},
		}
	}
}

// IsEditing returns true — settings is always in an interactive mode.
func (m SettingsModel) IsEditing() bool {
	return true
}
