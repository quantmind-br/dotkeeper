package views

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/pathutil"
)

func (m *SettingsModel) pathsForType(lt pathListType) []string {
	if lt == pathListFiles {
		return m.ctx.Config.Files
	}
	return m.ctx.Config.Folders
}

func (m *SettingsModel) setPathsForType(lt pathListType, paths []string) {
	if lt == pathListFiles {
		m.ctx.Config.Files = paths
	} else {
		m.ctx.Config.Folders = paths
	}
}

func (m *SettingsModel) disabledPathsForType(lt pathListType) []string {
	if lt == pathListFiles {
		return m.ctx.Config.DisabledFiles
	}
	return m.ctx.Config.DisabledFolders
}

func (m *SettingsModel) setDisabledPathsForType(lt pathListType, paths []string) {
	if lt == pathListFiles {
		m.ctx.Config.DisabledFiles = paths
	} else {
		m.ctx.Config.DisabledFolders = paths
	}
}

func (m *SettingsModel) togglePathDisabled(lt pathListType, path string) {
	disabled := m.disabledPathsForType(lt)
	for i, d := range disabled {
		if d == path {
			m.setDisabledPathsForType(lt, append(disabled[:i], disabled[i+1:]...))
			return
		}
	}
	m.setDisabledPathsForType(lt, append(disabled, path))
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
	if m.inspecting {
		m.inspecting = false
		m.inspectInfo = ""
		return m, nil
	}

	paths := m.pathsForType(lt)
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
		m.togglePathDisabled(lt, path)
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

func (m SettingsModel) handleFilePickerInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.state = m.filePickerParent
		return m, nil
	}
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
			m.ctx.Config.Notifications = !m.ctx.Config.Notifications
			m.refreshMainList()
		}
		return m, nil

	case "a":
		selected, ok := m.mainList.SelectedItem().(settingItem)
		if !ok {
			return m, nil
		}
		if selected.index == 2 {
			m.startEditingSubItem(stateBrowsingFiles, len(m.ctx.Config.Files), "")
		} else if selected.index == 3 {
			m.startEditingSubItem(stateBrowsingFolders, len(m.ctx.Config.Folders), "")
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

func (m *SettingsModel) saveConfig() {
	if err := m.ctx.Config.Save(); err != nil {
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
	linfo, _ := os.Lstat(expanded)
	if linfo != nil && linfo.Mode()&os.ModeSymlink != 0 {
		target, _ := os.Readlink(expanded)
		fileType = fmt.Sprintf("symlink → %s", target)
	}
	return fmt.Sprintf("Path: %s\nSize: %s | Modified: %s | Permissions: %s\nType: %s",
		path, pathutil.FormatSize(info.Size()), modTime, perm, fileType)
}

func (m *SettingsModel) refreshMainList() {
	schedule := m.ctx.Config.Schedule
	if schedule == "" {
		schedule = "Not scheduled"
	}

	items := []list.Item{
		settingItem{label: "Backup Directory", value: m.ctx.Config.BackupDir, index: 0},
		settingItem{label: "Git Remote", value: m.ctx.Config.GitRemote, index: 1},
		settingItem{label: "Files", value: fmt.Sprintf("%d files", len(m.ctx.Config.Files)), index: 2},
		settingItem{label: "Folders", value: fmt.Sprintf("%d folders", len(m.ctx.Config.Folders)), index: 3},
		settingItem{label: "Schedule", value: schedule, index: 4},
		settingItem{label: "Notifications", value: fmt.Sprintf("%v", m.ctx.Config.Notifications), index: 5},
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

func (m SettingsModel) scanPathDescs() tea.Cmd {
	files := make([]string, len(m.ctx.Config.Files))
	copy(files, m.ctx.Config.Files)
	folders := make([]string, len(m.ctx.Config.Folders))
	copy(folders, m.ctx.Config.Folders)
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
	width := m.ctx.Width
	height := m.ctx.Height - settingsViewChromeHeight
	if m.state == stateEditingField || m.state == stateEditingSubItem {
		height -= 2
	}
	if height < 0 {
		height = 0
	}

	if m.state == stateFilePickerActive {
		m.filePicker.Height = height
	}

	m.mainList.SetSize(width, height)
	m.filesList.SetSize(width, height)
	m.foldersList.SetSize(width, height)
}
