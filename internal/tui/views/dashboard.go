package views

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/pathutil"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

const dashboardRefreshInterval = 30 * time.Second

// DashboardModel represents the dashboard view
type DashboardModel struct {
	ctx         *ProgramContext
	lastBackup  time.Time
	fileCount   int
	totalSize   int64
	brokenPaths int
	selected    int
	err         error
	spinner     spinner.Model
	loading     bool
}

type dashboardAction struct {
	key    string
	label  string
	target string
}

var dashboardActions = []dashboardAction{
	{key: "b", label: "Backup", target: "backups"},
	{key: "r", label: "Restore", target: "restore"},
	{key: "s", label: "Settings", target: "settings"},
}

// NewDashboard creates a new dashboard model
func NewDashboard(ctx *ProgramContext) DashboardModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return DashboardModel{
		ctx:     ensureProgramContext(ctx),
		spinner: s,
	}
}

// Init initializes the dashboard
func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(m.refreshStatus(), m.spinner.Tick, m.scheduleRefresh())
}

// Update handles messages
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "up":
			if m.selected == 0 {
				m.selected = len(dashboardActions) - 1
			} else {
				m.selected--
			}
		case "right", "down":
			m.selected = (m.selected + 1) % len(dashboardActions)
		case "enter":
			target := dashboardActions[m.selected].target
			return m, func() tea.Msg {
				return DashboardNavigateMsg{Target: target}
			}
		}
	case statusMsg:
		m.lastBackup = msg.lastBackup
		m.fileCount = msg.fileCount
		m.totalSize = msg.totalSize
		m.brokenPaths = msg.brokenPaths
		m.loading = false
	case dashboardRefreshTickMsg:
		return m, tea.Batch(m.refreshStatus(), m.scheduleRefresh())
	}
	return m, nil
}

// View renders the dashboard
func (m DashboardModel) View() string {
	st := styles.DefaultStyles()

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Center,
			"\n",
			m.spinner.View(),
			"\nLoading dashboard...",
		)
	}

	var lastBackupVal string
	if !m.lastBackup.IsZero() {
		lastBackupVal = m.lastBackup.Format("2006-01-02 15:04")
	} else {
		lastBackupVal = "Never"
	}

	card1 := st.Card.Render(
		st.CardTitle.Render(lastBackupVal) + "\n" +
			st.CardLabel.Render("Last Backup"),
	)

	card2 := st.Card.Render(
		st.CardTitle.Render(fmt.Sprintf("%d", m.fileCount)) + "\n" +
			st.CardLabel.Render("Files Tracked"),
	)

	card3 := st.Card.Render(
		st.CardTitle.Render(pathutil.FormatSize(m.totalSize)) + "\n" +
			st.CardLabel.Render("Total Size"),
	)

	cards := []string{card1, card2, card3}

	if m.brokenPaths > 0 {
		warningCard := st.Card.Render(
			st.Error.Render(fmt.Sprintf("⚠ %d", m.brokenPaths)) + "\n" +
				st.CardLabel.Render("Broken Paths"),
		)
		cards = append(cards, warningCard)
	}

	var statsBlock string
	if m.ctx.Width >= styles.BreakpointWide {
		statsBlock = lipgloss.JoinHorizontal(lipgloss.Top, cards...)
	} else {
		statsBlock = lipgloss.JoinVertical(lipgloss.Left, cards...)
	}

	buttonIcons := map[string]string{
		"b": "󰁯",
		"r": "󰦛",
		"s": "",
	}

	actionButtons := make([]string, 0, len(dashboardActions))
	for i, action := range dashboardActions {
		icon := buttonIcons[action.key]
		label := icon + "  " + action.label

		btnStyle := st.ButtonNormal
		if i == m.selected {
			btnStyle = st.ButtonSelected
		}
		actionButtons = append(actionButtons, btnStyle.Render(label))
	}

	var actionsBlock string
	if m.ctx.Width >= styles.BreakpointMedium {
		actionsBlock = lipgloss.JoinHorizontal(lipgloss.Top, actionButtons...)
	} else {
		actionsBlock = lipgloss.JoinVertical(lipgloss.Left, actionButtons...)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		statsBlock,
		"\n",
		actionsBlock,
	)
}

type statusMsg struct {
	lastBackup  time.Time
	fileCount   int
	totalSize   int64
	brokenPaths int
}

type dashboardRefreshTickMsg struct{}

func (m *DashboardModel) refreshStatus() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		if m.ctx.Config == nil {
			return statusMsg{}
		}

		result := pathutil.ScanPaths(m.ctx.Config.ActiveFiles(), m.ctx.Config.ActiveFolders(), m.ctx.Config.Exclude)

		var lastBackup time.Time
		dir := pathutil.ExpandHome(m.ctx.Config.BackupDir)
		backups, _ := filepath.Glob(filepath.Join(dir, "backup-*.tar.gz.enc"))
		if len(backups) > 0 {
			info, _ := os.Stat(backups[len(backups)-1])
			if info != nil {
				lastBackup = info.ModTime()
			}
		}

		return statusMsg{
			lastBackup:  lastBackup,
			fileCount:   result.TotalFiles,
			totalSize:   result.TotalSize,
			brokenPaths: len(result.BrokenPaths),
		}
	}
}

func (m DashboardModel) scheduleRefresh() tea.Cmd {
	return tea.Tick(dashboardRefreshInterval, func(time.Time) tea.Msg {
		return dashboardRefreshTickMsg{}
	})
}

func (m *DashboardModel) Refresh() tea.Cmd {
	return m.refreshStatus()
}

func (m DashboardModel) HelpBindings() []HelpEntry {
	return []HelpEntry{
		{"b", "Go to backups"},
		{"r", "Go to restore"},
		{"s", "Go to settings"},
		{"arrow keys", "Select action"},
		{"enter", "Open selected action"},
	}
}

func (m DashboardModel) StatusHelpText() string {
	return "←/→: select | enter: open | b/r/s: shortcuts"
}

func (m DashboardModel) IsInputActive() bool {
	return false
}
