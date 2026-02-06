package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/pathutil"
)

// Keep lipgloss import for cursor/prompt styles (component configuration)

type settingsState int

const (
	stateReadOnly settingsState = iota
	stateListNavigating
	stateEditingField
	stateBrowsingFiles
	stateBrowsingFolders
	stateEditingSubItem
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
	title string
	desc  string
	index int
	isAdd bool
}

func (i subSettingItem) Title() string       { return i.title }
func (i subSettingItem) Description() string { return i.desc }
func (i subSettingItem) FilterValue() string { return i.title }

// SettingsModel represents the settings view
type SettingsModel struct {
	config      *config.Config
	width       int
	height      int
	state       settingsState
	mainList    list.Model
	filesList   list.Model
	foldersList list.Model
	textInput   textinput.Model
	status      string
	errMsg      string

	editingFieldIndex int
	subEditParent     settingsState
	subEditIndex      int
}

// NewSettings creates a new settings model
func NewSettings(cfg *config.Config) SettingsModel {
	ti := textinput.New()
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	mainList := list.New([]list.Item{}, NewListDelegate(), 80, 18)
	mainList.SetShowTitle(false)
	mainList.SetShowHelp(false)
	mainList.SetShowStatusBar(false)
	mainList.SetShowPagination(false)
	mainList.SetFilteringEnabled(false)

	filesList := list.New([]list.Item{}, NewListDelegate(), 80, 18)
	filesList.SetShowTitle(false)
	filesList.SetShowHelp(false)
	filesList.SetShowStatusBar(false)
	filesList.SetShowPagination(false)
	filesList.SetFilteringEnabled(false)

	foldersList := list.New([]list.Item{}, NewListDelegate(), 80, 18)
	foldersList.SetShowTitle(false)
	foldersList.SetShowHelp(false)
	foldersList.SetShowStatusBar(false)
	foldersList.SetShowPagination(false)
	foldersList.SetFilteringEnabled(false)

	m := SettingsModel{
		config:      cfg,
		width:       80,
		height:      24,
		state:       stateReadOnly,
		mainList:    mainList,
		filesList:   filesList,
		foldersList: foldersList,
		textInput:   ti,
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
	switch msg := msg.(type) {
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
		}
		return m.handleReadOnlyInput(msg)
	}
	return m, nil
}

// handleReadOnlyInput handles input when not in edit mode
func (m SettingsModel) handleReadOnlyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "e":
		m.state = stateListNavigating
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// handleEditModeInput handles input when in edit mode
func (m SettingsModel) handleEditModeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateReadOnly
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
		m.textInput.Blur()
		m.textInput.SetValue("")
		m.resizeLists()
		return m, nil

	case "enter":
		value := strings.TrimSpace(m.textInput.Value())
		m.saveFieldValue(value)
		m.refreshMainList()
		m.state = stateListNavigating
		m.textInput.Blur()
		m.textInput.SetValue("")
		m.resizeLists()
		return m, nil

	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m SettingsModel) handleEditingSubItemInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = m.subEditParent
		m.textInput.Blur()
		m.textInput.SetValue("")
		m.resizeLists()
		return m, nil

	case "enter":
		value := strings.TrimSpace(m.textInput.Value())
		m.saveFieldValue(value)
		m.refreshFilesList()
		m.refreshFoldersList()
		m.refreshMainList()
		m.state = m.subEditParent
		m.textInput.Blur()
		m.textInput.SetValue("")
		m.resizeLists()
		return m, nil

	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

// startEditingField initializes field editing
func (m *SettingsModel) startEditingField() {
	m.textInput.Focus()
	switch m.editingFieldIndex {
	case 0:
		m.textInput.SetValue(m.config.BackupDir)
	case 1:
		m.textInput.SetValue(m.config.GitRemote)
	case 4:
		m.textInput.SetValue(m.config.Schedule)
	default:
		m.textInput.SetValue("")
	}
}

func (m *SettingsModel) startEditingSubItem(parent settingsState, index int, value string) {
	m.subEditParent = parent
	m.subEditIndex = index
	m.state = stateEditingSubItem
	m.textInput.Focus()
	m.textInput.SetValue(value)
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
	items := make([]list.Item, 0, len(m.config.Files)+1)
	for i, filePath := range m.config.Files {
		items = append(items, subSettingItem{title: filePath, desc: "", index: i})
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
	items := make([]list.Item, 0, len(m.config.Folders)+1)
	for i, folderPath := range m.config.Folders {
		items = append(items, subSettingItem{title: folderPath, desc: "", index: i})
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

	height := m.height - ViewChromeHeight
	if m.state == stateEditingField || m.state == stateEditingSubItem {
		height -= 2
	}
	if height < 6 {
		height = 6
	}

	m.mainList.SetSize(width, height)
	m.filesList.SetSize(width, height)
	m.foldersList.SetSize(width, height)
}

// View renders the settings view
func (m SettingsModel) View() string {
	var b strings.Builder

	styles := DefaultStyles()

	if m.state == stateReadOnly {
		b.WriteString(styles.Title.Render("Settings") + "\n")
		b.WriteString(styles.Hint.Render("Press 'e' to edit") + "\n")
	} else {
		b.WriteString(styles.Title.Render("Settings [EDIT MODE]") + "\n\n")
	}

	helpText := ""
	switch m.state {
	case stateReadOnly:
		b.WriteString(m.mainList.View())
		helpText = "e: Edit mode"
	case stateListNavigating:
		b.WriteString(m.mainList.View())
		helpText = "↑/↓: Navigate | Enter: Edit | a: Add | s: Save | Esc: Exit"
	case stateEditingField:
		b.WriteString("Editing: " + m.textInput.View() + "\n\n")
		b.WriteString(m.mainList.View())
		helpText = "Enter: Save field | Esc: Cancel"
	case stateBrowsingFiles:
		b.WriteString(styles.Subtitle.Render("Files") + "\n")
		b.WriteString(m.filesList.View())
		helpText = "↑/↓: Navigate | Enter: Edit | a: Add | d: Delete | s: Save | Esc: Back"
	case stateBrowsingFolders:
		b.WriteString(styles.Subtitle.Render("Folders") + "\n")
		b.WriteString(m.foldersList.View())
		helpText = "↑/↓: Navigate | Enter: Edit | a: Add | d: Delete | s: Save | Esc: Back"
	case stateEditingSubItem:
		title := "Editing File"
		listView := m.filesList.View()
		if m.subEditParent == stateBrowsingFolders {
			title = "Editing Folder"
			listView = m.foldersList.View()
		}
		b.WriteString(styles.Subtitle.Render(title) + "\n")
		b.WriteString("Value: " + m.textInput.View() + "\n\n")
		b.WriteString(listView)
		helpText = "Enter: Save item | Esc: Cancel"
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
			{"a", "Add item"},
			{"d", "Delete item"},
			{"s", "Save config"},
			{"Esc", "Exit edit"},
		}
	default:
		return []HelpEntry{
			{"e", "Edit mode"},
		}
	}
}

// IsEditing returns true when the settings view is in edit mode or editing a field.
func (m SettingsModel) IsEditing() bool {
	return m.state != stateReadOnly
}
