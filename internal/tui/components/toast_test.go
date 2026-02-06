package components

import (
	"strings"
	"testing"
	"time"
)

func TestToast_ShowAndView(t *testing.T) {
	toast := NewToast()

	if toast.Visible {
		t.Error("New toast should not be visible")
	}

	if toast.View() != "" {
		t.Error("New toast should render empty string")
	}

	cmd := toast.Show("Test message", ToastInfo)
	if cmd == nil {
		t.Error("Show() should return a command")
	}

	if !toast.Visible {
		t.Error("Toast should be visible after Show()")
	}

	if toast.Message != "Test message" {
		t.Errorf("Toast message should be 'Test message', got %q", toast.Message)
	}

	view := toast.View()
	if view == "" {
		t.Error("Visible toast should render non-empty string")
	}

	if !strings.Contains(view, "Test message") {
		t.Errorf("Toast view should contain message, got %q", view)
	}
}

func TestToast_Dismiss(t *testing.T) {
	toast := NewToast()
	toast.Show("Test", ToastSuccess)

	if !toast.Visible {
		t.Error("Toast should be visible after Show()")
	}

	toast.Dismiss()

	if toast.Visible {
		t.Error("Toast should not be visible after Dismiss()")
	}

	if toast.Message != "" {
		t.Error("Toast message should be empty after Dismiss()")
	}

	if toast.View() != "" {
		t.Error("Dismissed toast should render empty string")
	}
}

func TestToast_EmptyWhenNotVisible(t *testing.T) {
	toast := NewToast()
	toast.Message = "Hidden message"
	toast.Visible = false

	if toast.View() != "" {
		t.Error("Invisible toast should render empty string even with message")
	}
}

func TestToast_Levels(t *testing.T) {
	tests := []struct {
		name  string
		level ToastLevel
	}{
		{"Info", ToastInfo},
		{"Success", ToastSuccess},
		{"Error", ToastError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toast := NewToast()
			cmd := toast.Show("Test "+tt.name, tt.level)

			if cmd == nil {
				t.Error("Show() should return a command")
			}

			if toast.Level != tt.level {
				t.Errorf("Toast level should be %v, got %v", tt.level, toast.Level)
			}

			view := toast.View()
			if view == "" {
				t.Error("Toast should render non-empty string")
			}

			if !strings.Contains(view, "Test "+tt.name) {
				t.Errorf("Toast view should contain message, got %q", view)
			}
		})
	}
}

func TestToast_DismissMsg(t *testing.T) {
	toast := NewToast()
	cmd := toast.Show("Test", ToastInfo)

	if cmd == nil {
		t.Fatal("Show() should return a command")
	}

	msg := cmd()
	if msg == nil {
		t.Fatal("Command should return a message")
	}

	_, ok := msg.(ToastDismissMsg)
	if !ok {
		t.Errorf("Command should return ToastDismissMsg, got %T", msg)
	}
}

func TestToast_DefaultDuration(t *testing.T) {
	toast := NewToast()

	if toast.duration != 3*time.Second {
		t.Errorf("Default duration should be 3 seconds, got %v", toast.duration)
	}
}
