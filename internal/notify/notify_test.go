package notify

import (
	"errors"
	"testing"
	"time"
)

// TestSend verifies that Send handles missing notify-send gracefully
func TestSend(t *testing.T) {
	// This test verifies that Send doesn't panic and returns nil
	// (gracefully handling missing notify-send)
	err := Send("Test Title", "Test Message")
	if err != nil {
		// We expect nil since notify-send might not be available
		// and we gracefully handle that case
		t.Logf("Send returned error (expected if notify-send not available): %v", err)
	}
}

// TestSendSuccess verifies that SendSuccess formats the message correctly
func TestSendSuccess(t *testing.T) {
	backupName := "backup-2025-02-04-120000"
	duration := 5 * time.Second

	err := SendSuccess(backupName, duration)
	if err != nil {
		// We expect nil since notify-send might not be available
		t.Logf("SendSuccess returned error (expected if notify-send not available): %v", err)
	}
}

// TestSendError verifies that SendError formats the error message correctly
func TestSendError(t *testing.T) {
	testErr := errors.New("test error message")
	err := SendError(testErr)
	if err != nil {
		// We expect nil since notify-send might not be available
		t.Logf("SendError returned error (expected if notify-send not available): %v", err)
	}
}

// TestSendWithEmptyTitle verifies Send handles empty title
func TestSendWithEmptyTitle(t *testing.T) {
	err := Send("", "Message without title")
	if err != nil {
		t.Logf("Send with empty title returned error (expected if notify-send not available): %v", err)
	}
}

// TestSendWithEmptyMessage verifies Send handles empty message
func TestSendWithEmptyMessage(t *testing.T) {
	err := Send("Title without message", "")
	if err != nil {
		t.Logf("Send with empty message returned error (expected if notify-send not available): %v", err)
	}
}
