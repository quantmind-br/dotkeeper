package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/history"
	"github.com/diogo/dotkeeper/internal/tui/styles"
)

// LogsModel represents the logs view
type LogsModel struct {
	ctx     *ProgramContext
	list    list.Model
	filter  string // "all", "backup", "restore"
	err     string
	spinner spinner.Model
	loading bool
}

// logItem adapts history.HistoryEntry to list.Item
type logItem struct {
	entry history.HistoryEntry
}

func (i logItem) Title() string {
	status := "✓"
	if i.entry.Status == "error" {
		status = "✗"
	}
	// Local time "2006-01-02 15:04"
	ts := i.entry.Timestamp.Local().Format("2006-01-02 15:04")
	return fmt.Sprintf("%s | %s | %s", ts, i.entry.Operation, status)
}

func (i logItem) Description() string {
	size := formatBytes(i.entry.TotalSize)
	// Duration
	duration := time.Duration(i.entry.DurationMs) * time.Millisecond
	return fmt.Sprintf("%d files, %s, %s", i.entry.FileCount, size, duration.String())
}

func (i logItem) FilterValue() string {
	return i.entry.Operation
}

// formatBytes converts bytes to human readable string
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// NewLogs creates a new logs model.
func NewLogs(ctx *ProgramContext) LogsModel {
	l := styles.NewMinimalList()

	s := spinner.New()
	s.Spinner = spinner.Dot

	return LogsModel{
		ctx:     ensureProgramContext(ctx),
		list:    l,
		filter:  "all",
		spinner: s,
	}
}

// Init initializes the logs view
func (m LogsModel) Init() tea.Cmd {
	return tea.Batch(m.LoadHistory(), m.spinner.Tick)
}

// LoadHistory loads history from the store. Exported for cross-view refresh.
func (m *LogsModel) LoadHistory() tea.Cmd {
	if m.ctx.Store == nil {
		return nil
	}

	m.loading = true
	return func() tea.Msg {
		var entries []history.HistoryEntry
		var err error

		if m.filter == "all" {
			entries, err = m.ctx.Store.Read(100)
		} else {
			entries, err = m.ctx.Store.ReadByType(m.filter, 100)
		}

		if err != nil {
			return ErrorMsg{Source: "logs", Err: err}
		}
		return logsLoadedMsg(entries)
	}
}

type logsLoadedMsg []history.HistoryEntry

// Update handles messages
func (m LogsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
		// Adjust list size. Reserve space for chrome
		m.list.SetSize(msg.Width, msg.Height)

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case logsLoadedMsg:
		m.loading = false
		items := make([]list.Item, len(msg))
		for i, entry := range msg {
			items[i] = logItem{entry: entry}
		}
		m.list.SetItems(items)
		if len(items) == 0 {
			m.err = ""
		}
		return m, nil

	case ErrorMsg:
		if msg.Source == "logs" {
			m.loading = false
			m.err = msg.Err.Error()
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "f":
			// Cycle filter
			switch m.filter {
			case "all":
				m.filter = "backup"
			case "backup":
				m.filter = "restore"
			case "restore":
				m.filter = "all"
			default:
				m.filter = "all"
			}
			m.list.ResetSelected()
			return m, m.LoadHistory()

		case "r":
			return m, m.LoadHistory()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the logs view
func (m LogsModel) View() string {
	var s strings.Builder
	st := styles.DefaultStyles()

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Center,
			"\n",
			m.spinner.View(),
			"\nLoading logs...",
		)
	}

	// Title
	title := "Operation History"
	switch m.filter {
	case "all":
		title += " [all]"
	case "backup":
		title += " [backup]"
	case "restore":
		title += " [restore]"
	}
	s.WriteString(st.Title.Render(title) + "\n\n")

	// Content
	if m.err != "" {
		s.WriteString("\n")
	} else if len(m.list.Items()) == 0 {
		s.WriteString("No operations recorded yet. Run a backup to get started!\n")
	} else {
		s.WriteString(m.list.View())
	}

	s.WriteString(RenderStatusBar(m.ctx.Width, "", m.err, ""))

	return s.String()
}

func (m LogsModel) StatusHelpText() string {
	return fmt.Sprintf("f: filter (%s) | r: refresh | ↑/↓: navigate", m.filter)
}

// Refresh reloads the history. Alias for LoadHistory to implement Refreshable.
func (m LogsModel) Refresh() tea.Cmd {
	return m.LoadHistory()
}

func (m LogsModel) IsInputActive() bool {
	return false
}

func (m LogsModel) HelpBindings() []HelpEntry {
	return []HelpEntry{
		{"f", "filter"},
		{"r", "refresh"},
		{"↑/↓", "navigate"},
	}
}
