package components

import (
	"regexp"
	"strings"
	"testing"

	"github.com/diogo/dotkeeper/internal/tui/views"
)

func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func TestTabBarActiveHighlight(t *testing.T) {
	tb := NewTabBar(views.DefaultStyles())
	outputs := make([]string, 5)
	plainOutputs := make([]string, 5)

	tabs := []string{"Dashboard", "Backups", "Restore", "Settings", "Logs"}

	for i := 0; i < 5; i++ {
		outputs[i] = tb.View(i, 100)
		if outputs[i] == "" {
			t.Errorf("View(%d, 100) returned empty string", i)
		}

		plainOutputs[i] = stripANSI(outputs[i])

		// Verify structural correctness: all labels must be present
		for _, tab := range tabs {
			if !strings.Contains(plainOutputs[i], tab) {
				t.Errorf("View(%d, 100) plain text missing tab label '%s'", i, tab)
			}
		}
	}

	// Verify that different active indices produce visually different outputs (if ANSI present)
	hasANSI := strings.Contains(outputs[0], "\x1b")
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			// Plain text should be identical for all (structural consistency)
			if plainOutputs[i] != plainOutputs[j] {
				t.Errorf("Structural mismatch: plain text of View(%d) != View(%d)", i, j)
			}

			// RAW output should differ if ANSI is present
			if hasANSI && outputs[i] == outputs[j] {
				t.Errorf("Visual mismatch: RAW output of View(%d) == View(%d) despite ANSI being enabled", i, j)
			}
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
