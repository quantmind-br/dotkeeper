package components

import (
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/tui/views"
)

func TestTabBarActiveHighlight(t *testing.T) {
	tb := NewTabBar(views.DefaultStyles())

	for i := 0; i < 5; i++ {
		output := tb.View(i, 100)
		if output == "" {
			t.Errorf("View(%d, 100) returned empty string", i)
		}
		if !strings.Contains(output, "Dashboard") {
			t.Errorf("View(%d, 100) missing 'Dashboard'", i)
		}
	}

	// Verify activeIndex is clamped: out-of-bounds defaults to 0, same as explicit 0
	outOfBounds := tb.View(99, 100)
	explicit0 := tb.View(0, 100)
	if outOfBounds != explicit0 {
		t.Error("Out-of-bounds activeIndex should default to 0")
	}
}

func TestTabBarResponsive(t *testing.T) {
	tb := NewTabBar(views.DefaultStyles())

	// Wide terminal (>= 80): full labels
	wide := tb.View(0, 100)
	if !strings.Contains(wide, "Dashboard") {
		t.Error("Wide view should contain 'Dashboard'")
	}
	if !strings.Contains(wide, "Backups") {
		t.Error("Wide view should contain 'Backups'")
	}
	if !strings.Contains(wide, "Settings") {
		t.Error("Wide view should contain 'Settings'")
	}

	// Narrow terminal (< 80): abbreviated labels
	narrow := tb.View(0, 60)
	if !strings.Contains(narrow, "Dash") {
		t.Error("Narrow view should contain 'Dash'")
	}
	if !strings.Contains(narrow, "Bkps") {
		t.Error("Narrow view should contain 'Bkps'")
	}
	if !strings.Contains(narrow, "Sett") {
		t.Error("Narrow view should contain 'Sett'")
	}
	// Full name should NOT appear in narrow view
	if strings.Contains(narrow, "Dashboard") {
		t.Error("Narrow view should NOT contain 'Dashboard' (should be 'Dash')")
	}
}

func TestTabBarSeparators(t *testing.T) {
	tb := NewTabBar(views.DefaultStyles())
	output := tb.View(0, 100)

	// Should have 4 separators (between 5 tabs)
	count := strings.Count(output, "│")
	if count != 4 {
		t.Errorf("Expected 4 separators (│), got %d", count)
	}
}

func TestTabBarNumberLabels(t *testing.T) {
	tb := NewTabBar(views.DefaultStyles())
	output := tb.View(0, 100)

	for _, num := range []string{"1", "2", "3", "4", "5"} {
		if !strings.Contains(output, num) {
			t.Errorf("View() should contain number '%s'", num)
		}
	}
}

func TestTabBarAllTabs(t *testing.T) {
	tb := NewTabBar(views.DefaultStyles())
	output := tb.View(0, 100)

	tabs := []string{"Dashboard", "Backups", "Restore", "Settings", "Logs"}
	for _, tab := range tabs {
		if !strings.Contains(output, tab) {
			t.Errorf("View() should contain tab '%s'", tab)
		}
	}
}

func TestTabBarOutOfBoundsIndex(t *testing.T) {
	tb := NewTabBar(views.DefaultStyles())

	// Should not panic with out-of-bounds index
	_ = tb.View(-1, 100) // negative
	_ = tb.View(10, 100) // too high
	_ = tb.View(5, 100)  // exactly out of bounds
}
