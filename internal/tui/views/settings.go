package views

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
	"github.com/diogo/dotkeeper/internal/pathutil"
	"github.com/diogo/dotkeeper/internal/tui/components"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

type settingsState int

const (
	stateListNavigating settingsState = iota
	stateEditingField
	stateBrowsingFiles
	stateBrowsingFolders
	stateEditingSubItem
	stateFilePickerActive
)

// pathListType distinguishes between file and folder path lists.
type pathListType int

const (
	pathListFiles pathListType = iota
	pathListFolders
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

type pathDescsMsg struct {
	descs map[string]string
}

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
	pathDescs     map[string]string

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

	mainList := styles.NewMinimalList()
	mainList.SetShowStatusBar(false)
	mainList.SetShowPagination(false)

	filesList := styles.NewMinimalList()
	filesList.SetShowStatusBar(false)
	filesList.SetShowPagination(false)

	foldersList := styles.NewMinimalList()
	foldersList.SetShowStatusBar(false)
	foldersList.SetShowPagination(false)

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
	m.refreshPathList(pathListFiles)
	m.refreshPathList(pathListFolders)
	m.resizeLists()

	return m
}

// Init initializes the settings view
func (m SettingsModel) Init() tea.Cmd {
	return m.scanPathDescs()
}

// Update handles messages
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if descs, ok := msg.(pathDescsMsg); ok {
		if m.pathDescs == nil {
			m.pathDescs = make(map[string]string)
		}
		for k, v := range descs.descs {
			m.pathDescs[k] = v
		}
		m.refreshPathList(pathListFiles)
		m.refreshPathList(pathListFolders)
		return m, nil
	}

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
				info, statErr := os.Stat(path)
				if m.filePickerParent == stateBrowsingFiles {
					if statErr == nil && info.IsDir() {
						m.errMsg = "Selected path is a directory, not a file"
						m.state = m.filePickerParent
						return m, nil
					}
					m.config.Files = append(m.config.Files, path)
					m.refreshPathList(pathListFiles)
				} else {
					if statErr == nil && !info.IsDir() {
						m.errMsg = "Selected path is a file, not a directory"
						m.state = m.filePickerParent
						return m, nil
					}
					m.config.Folders = append(m.config.Folders, path)
					m.refreshPathList(pathListFolders)
				}
				m.errMsg = ""
				m.refreshMainList()
				m.state = m.filePickerParent
				return m, m.scanPathDescs()
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
			return m.handleBrowsingPathsInput(msg, pathListFiles)
		case stateBrowsingFolders:
			return m.handleBrowsingPathsInput(msg, pathListFolders)
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

func (m *SettingsModel) pathsForType(lt pathListType) []string {
	if lt == pathListFiles {
		return m.config.Files
	}
	return m.config.Folders
}

func (m *SettingsModel) setPathsForType(lt pathListType, paths []string) {
	if lt == pathListFiles {
		m.config.Files = paths
	} else {
		m.config.Folders = paths
	}
}

func (m *SettingsModel) disabledPathsForType(lt pathListType) []string {
	if lt == pathListFiles {
		return m.config.DisabledFiles
	}
	return m.config.DisabledFolders
}

func (m *SettingsModel) setDisabledPathsForType(lt pathListType, paths []string) {
	if lt == pathListFiles {
		m.config.DisabledFiles = paths
	} else {
		m.config.DisabledFolders = paths
	}
}

func (m *SettingsModel) listForType(lt pathListType) *list.Model {
	if lt == pathListFiles {
		return &m.filesList
	}
	return &m.foldersList
}

func (m *SettingsModel) browsingStateForType(lt pathListType) settingsState {
	if lt == pathListFiles {
		return stateBrowsingFiles
	}
	return stateBrowsingFolders
}

func (m *SettingsModel) addLabel(lt pathListType) string {
	if lt == pathListFiles {
		return "[+] Add new file"
	}
	return "[+] Add new folder"
}

func (m SettingsModel) handleBrowsingPathsInput(msg tea.KeyMsg, lt pathListType) (tea.Model, tea.Cmd) {
	// If inspecting, any key dismisses the inspect overlay
	if m.inspecting {
		m.inspecting = false
		m.inspectInfo = ""
		return m, nil
	}

	paths := m.pathsForType(lt)
	disabledPaths := m.disabledPathsForType(lt)
	pathList := m.listForType(lt)
	browsingState := m.browsingStateForType(lt)

	switch msg.String() {
	case "esc":
		m.state = stateListNavigating
		m.inspecting = false
		m.inspectInfo = ""
		return m, nil
	case " ":
		selected, ok := pathList.SelectedItem().(subSettingItem)
		if !ok || selected.isAdd {
			return m, nil
		}
		path := paths[selected.index]
		found := false
		for i, d := range disabledPaths {
			if d == path {
				newDisabled := append(disabledPaths[:i], disabledPaths[i+1:]...)
				m.setDisabledPathsForType(lt, newDisabled)
				found = true
				break
			}
		}
		if !found {
			m.setDisabledPathsForType(lt, append(disabledPaths, path))
		}
		m.refreshPathList(lt)
		return m, nil
	case "i":
		selected, ok := pathList.SelectedItem().(subSettingItem)
		if !ok || selected.isAdd {
			return m, nil
		}
		path := paths[selected.index]
		m.inspecting = true
		m.inspectInfo = getInspectInfo(path)
		return m, nil
	case "enter":
		selected, ok := pathList.SelectedItem().(subSettingItem)
		if !ok {
			return m, nil
		}
		if selected.isAdd {
			m.startEditingSubItem(browsingState, len(paths), "")
		} else if selected.index >= 0 && selected.index < len(paths) {
			m.startEditingSubItem(browsingState, selected.index, paths[selected.index])
		}
		return m, nil
	case "a":
		m.startEditingSubItem(browsingState, len(paths), "")
		return m, nil
	case "b":
		m.filePickerParent = browsingState
		m.state = stateFilePickerActive
		return m, m.filePicker.Init()
	case "d":
		selected, ok := pathList.SelectedItem().(subSettingItem)
		if !ok || selected.isAdd || selected.index < 0 || selected.index >= len(paths) {
			return m, nil
		}
		newPaths := append(paths[:selected.index], paths[selected.index+1:]...)
		m.setPathsForType(lt, newPaths)
		m.refreshPathList(lt)
		m.refreshMainList()
		if selected.index < len(pathList.Items()) {
			pathList.Select(selected.index)
		}
		return m, nil
	case "s":
		m.saveConfig()
		return m, nil
	}

	var cmd tea.Cmd
	*pathList, cmd = pathList.Update(msg)
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
			m.refreshPathList(pathListFiles)
			m.refreshPathList(pathListFolders)
			m.refreshMainList()
			m.state = m.subEditParent
			m.pathCompleter.Input.Blur()
			m.pathCompleter.Input.SetValue("")
			m.resizeLists()
			return m, m.scanPathDescs()
		}
		m.saveFieldValue(value)
		m.refreshPathList(pathListFiles)
		m.refreshPathList(pathListFolders)
		m.refreshMainList()
		m.state = m.subEditParent
		m.pathCompleter.Input.Blur()
		m.pathCompleter.Input.SetValue("")
		m.resizeLists()
		return m, m.scanPathDescs()

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

func (m *SettingsModel) refreshPathList(lt pathListType) {
	paths := m.pathsForType(lt)
	disabledPaths := m.disabledPathsForType(lt)
	pathList := m.listForType(lt)

	disabledSet := make(map[string]bool)
	for _, d := range disabledPaths {
		disabledSet[d] = true
	}

	items := make([]list.Item, 0, len(paths)+1)
	for i, p := range paths {
		desc := "scanning..."
		if cached, ok := m.pathDescs[p]; ok {
			desc = cached
		}
		isDisabled := disabledSet[p]
		if isDisabled {
			desc = "[disabled] " + desc
		}
		items = append(items, subSettingItem{title: p, desc: desc, index: i, disabled: isDisabled})
	}
	items = append(items, subSettingItem{title: m.addLabel(lt), desc: "", index: len(paths), isAdd: true})

	selected := pathList.Index()
	pathList.SetItems(items)
	if selected >= len(items) {
		selected = len(items) - 1
	}
	if selected < 0 {
		selected = 0
	}
	pathList.Select(selected)
}

func (m SettingsModel) scanPathDescs() tea.Cmd {
	files := make([]string, len(m.config.Files))
	copy(files, m.config.Files)
	folders := make([]string, len(m.config.Folders))
	copy(folders, m.config.Folders)
	return func() tea.Msg {
		descs := make(map[string]string, len(files)+len(folders))
		for _, p := range files {
			descs[p] = pathutil.GetPathDesc(p)
		}
		for _, p := range folders {
			descs[p] = pathutil.GetPathDesc(p)
		}
		return pathDescsMsg{descs: descs}
	}
}

func (m *SettingsModel) resizeLists() {
	width := m.width
	height := m.height
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

	st := styles.DefaultStyles()

	b.WriteString(st.Title.Render("Settings") + "\n\n")

	switch m.state {
	case stateListNavigating:
		b.WriteString(m.mainList.View())
	case stateEditingField:
		b.WriteString("Editing: " + m.pathCompleter.View() + "\n\n")
		b.WriteString(m.mainList.View())
	case stateBrowsingFiles:
		b.WriteString(st.Subtitle.Render("Files") + "\n")
		b.WriteString(m.filesList.View())
	case stateBrowsingFolders:
		b.WriteString(st.Subtitle.Render("Folders") + "\n")
		b.WriteString(m.foldersList.View())
	case stateEditingSubItem:
		title := "Editing File"
		listView := m.filesList.View()
		if m.subEditParent == stateBrowsingFolders {
			title = "Editing Folder"
			listView = m.foldersList.View()
		}
		b.WriteString(st.Subtitle.Render(title) + "\n")
		b.WriteString("Value: " + m.pathCompleter.View() + "\n\n")
		b.WriteString(listView)
	case stateFilePickerActive:
		title := "Browse Files"
		if m.filePickerParent == stateBrowsingFolders {
			title = "Browse Folders"
		}
		b.WriteString(st.Subtitle.Render(title) + "\n")
		b.WriteString(m.filePicker.View())
	}

	if m.inspecting && m.inspectInfo != "" {
		b.WriteString("\n")
		b.WriteString(st.Card.Render(m.inspectInfo))
		b.WriteString("\n")
	}

	b.WriteString("\n" + RenderStatusBar(m.width, m.status, m.errMsg, ""))

	return b.String()
}

func (m SettingsModel) HelpBindings() []HelpEntry {
	switch m.state {
	case stateEditingField, stateEditingSubItem:
		return []HelpEntry{
			{"Enter", "Save field"},
			{"Esc", "Cancel edit"},
		}
	case stateListNavigating:
		return []HelpEntry{
			{"↑/↓", "Navigate"},
			{"Enter", "Edit"},
			{"a", "Add (on lists)"},
			{"s", "Save"},
			{"Esc", "Back"},
		}
	case stateBrowsingFiles, stateBrowsingFolders:
		return []HelpEntry{
			{"↑/↓", "Navigate"},
			{"Space", "Toggle"},
			{"i", "Inspect"},
			{"Enter", "Edit"},
			{"a", "Type path"},
			{"b", "Browse"},
			{"d", "Delete"},
			{"s", "Save"},
			{"Esc", "Back"},
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

func (m SettingsModel) StatusHelpText() string {
	switch m.state {
	case stateListNavigating:
		return "↑/↓: Navigate | Enter: Edit | a: Add | s: Save | Esc: Exit"
	case stateEditingField:
		return "Enter: Save field | Esc: Cancel"
	case stateBrowsingFiles, stateBrowsingFolders:
		return "↑/↓: Navigate | Enter: Edit | a: Type path | b: Browse | d: Delete | s: Save | Esc: Back"
	case stateEditingSubItem:
		return "Enter: Save item | Esc: Cancel"
	case stateFilePickerActive:
		return "Enter: Select | ↑/↓: Navigate | Esc: Cancel"
	default:
		return ""
	}
}

func (m SettingsModel) IsEditing() bool {
	return m.state != stateListNavigating
}
