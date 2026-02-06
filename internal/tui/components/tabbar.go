package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/diogo/dotkeeper/internal/tui/views"
)

// TabItem represents a single tab
type TabItem struct {
	Key        string
	Label      string
	ShortLabel string
}

// TabBar is a stateless tab bar renderer
type TabBar struct {
	items  []TabItem
	styles views.Styles
}

// NewTabBar creates a new TabBar with 5 hardcoded tabs
func NewTabBar(styles views.Styles) TabBar {
	return TabBar{
		items: []TabItem{
			{Key: "1", Label: "Dashboard", ShortLabel: "Dash"},
			{Key: "2", Label: "Backups", ShortLabel: "Bkps"},
			{Key: "3", Label: "Restore", ShortLabel: "Rest"},
			{Key: "4", Label: "Settings", ShortLabel: "Sett"},
			{Key: "5", Label: "Logs", ShortLabel: "Logs"},
		},
		styles: styles,
	}
}

// View renders the tab bar with the given active index and width
func (tb TabBar) View(activeIndex int, width int) string {
	// Validate activeIndex
	if activeIndex < 0 || activeIndex >= len(tb.items) {
		activeIndex = 0
	}

	// Determine if we should use abbreviated labels
	useShort := width < 80

	// Build tab strings
	var tabs []string
	for i, item := range tb.items {
		label := item.Label
		if useShort {
			label = item.ShortLabel
		}

		tabText := item.Key + " " + label

		// Apply style based on active state
		if i == activeIndex {
			tabText = tb.styles.TabActive.Render(tabText)
		} else {
			tabText = tb.styles.TabInactive.Render(tabText)
		}

		tabs = append(tabs, tabText)
	}

	// Join tabs with separator
	separator := tb.styles.TabSeparator.Render(" │ ")
	result := strings.Join(tabs, separator)

	// Apply margin left
	result = lipgloss.NewStyle().MarginLeft(2).Render(result)

	return result
}
