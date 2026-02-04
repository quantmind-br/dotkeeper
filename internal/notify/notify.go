package notify

import (
	"fmt"
	"os/exec"
	"time"
)

// Send sends a desktop notification with the given title and message.
// It gracefully handles the case where notify-send is not available.
func Send(title, message string) error {
	cmd := exec.Command("notify-send", title, message)
	if err := cmd.Run(); err != nil {
		// Check if the error is due to notify-send not being found
		if _, ok := err.(*exec.ExitError); ok {
			// notify-send exists but returned an error
			return fmt.Errorf("notify-send failed: %w", err)
		}
		// notify-send not found or other error - gracefully ignore
		return nil
	}
	return nil
}

// SendSuccess sends a success notification for a completed backup.
func SendSuccess(backupName string, duration time.Duration) error {
	title := "Backup Successful"
	message := fmt.Sprintf("Backup '%s' completed in %v", backupName, duration)
	return Send(title, message)
}

// SendError sends an error notification for a failed backup.
func SendError(err error) error {
	title := "Backup Failed"
	message := fmt.Sprintf("Error: %v", err)
	return Send(title, message)
}
