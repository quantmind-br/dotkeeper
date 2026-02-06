package views

import (
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/config"
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

const settingsViewChromeHeight = 5

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
	ctx           *ProgramContext
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
	spinner           spinner.Model
	loading           bool
}

// NewSettings creates a new settings model
func NewSettings(ctx *ProgramContext) SettingsModel {
	ctx = ensureProgramContext(ctx)
	if ctx.Config == nil {
		ctx.Config = &config.Config{}
	}
	if ctx.Width == 0 {
		ctx.Width = 80
	}
	if ctx.Height == 0 {
		ctx.Height = 24
	}

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

	s := spinner.New()
	s.Spinner = spinner.Dot

	m := SettingsModel{
		ctx:           ctx,
		state:         stateListNavigating,
		mainList:      mainList,
		filesList:     filesList,
		foldersList:   foldersList,
		pathCompleter: pc,
		filePicker:    fp,
		spinner:       s,
	}
	m.refreshMainList()
	m.refreshPathList(pathListFiles)
	m.refreshPathList(pathListFolders)
	m.resizeLists()

	return m
}

// Init initializes the settings view
func (m SettingsModel) Init() tea.Cmd {
	return tea.Batch(m.scanPathDescs(), m.spinner.Tick)
}

// Update handles messages
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if descs, ok := msg.(pathDescsMsg); ok {
		m.loading = false
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
					m.ctx.Config.Files = append(m.ctx.Config.Files, path)
					m.refreshPathList(pathListFiles)
				} else {
					if statErr == nil && !info.IsDir() {
						m.errMsg = "Selected path is a file, not a directory"
						m.state = m.filePickerParent
						return m, nil
					}
					m.ctx.Config.Folders = append(m.ctx.Config.Folders, path)
					m.refreshPathList(pathListFolders)
				}
				m.errMsg = ""
				m.refreshMainList()
				m.state = m.filePickerParent
				m.loading = true
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
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
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
