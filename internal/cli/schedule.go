package cli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	serviceName = "dotkeeper.service"
	timerName   = "dotkeeper.timer"
)

// EnableSchedule enables the systemd timer for automated backups
func EnableSchedule() error {
	// Check if systemd is available
	if !isSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	// Get the service and timer file paths
	serviceFile := filepath.Join("contrib", "systemd", serviceName)
	timerFile := filepath.Join("contrib", "systemd", timerName)

	// Check if files exist
	if _, err := os.Stat(serviceFile); os.IsNotExist(err) {
		return fmt.Errorf("service file not found: %s", serviceFile)
	}
	if _, err := os.Stat(timerFile); os.IsNotExist(err) {
		return fmt.Errorf("timer file not found: %s", timerFile)
	}

	// Get user systemd directory
	userSystemdDir, err := getUserSystemdDir()
	if err != nil {
		return fmt.Errorf("failed to get user systemd directory: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(userSystemdDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	// Copy service file
	destServiceFile := filepath.Join(userSystemdDir, serviceName)
	if err := copyFile(serviceFile, destServiceFile); err != nil {
		return fmt.Errorf("failed to copy service file: %w", err)
	}

	// Copy timer file
	destTimerFile := filepath.Join(userSystemdDir, timerName)
	if err := copyFile(timerFile, destTimerFile); err != nil {
		return fmt.Errorf("failed to copy timer file: %w", err)
	}

	// Reload systemd daemon
	if err := runSystemctl("daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}

	// Enable and start the timer
	if err := runSystemctl("enable", timerName); err != nil {
		return fmt.Errorf("failed to enable timer: %w", err)
	}

	if err := runSystemctl("start", timerName); err != nil {
		return fmt.Errorf("failed to start timer: %w", err)
	}

	fmt.Println("✓ Systemd timer enabled successfully")
	fmt.Printf("  Service file: %s\n", destServiceFile)
	fmt.Printf("  Timer file: %s\n", destTimerFile)
	fmt.Println("\nTo check timer status, run:")
	fmt.Printf("  systemctl --user status %s\n", timerName)
	fmt.Println("\nTo view timer schedule, run:")
	fmt.Printf("  systemctl --user list-timers %s\n", timerName)

	return nil
}

// DisableSchedule disables the systemd timer for automated backups
func DisableSchedule() error {
	// Check if systemd is available
	if !isSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	// Stop and disable the timer
	if err := runSystemctl("stop", timerName); err != nil {
		fmt.Printf("Warning: failed to stop timer: %v\n", err)
	}

	if err := runSystemctl("disable", timerName); err != nil {
		fmt.Printf("Warning: failed to disable timer: %v\n", err)
	}

	// Get user systemd directory
	userSystemdDir, err := getUserSystemdDir()
	if err != nil {
		return fmt.Errorf("failed to get user systemd directory: %w", err)
	}

	// Remove service and timer files
	destServiceFile := filepath.Join(userSystemdDir, serviceName)
	destTimerFile := filepath.Join(userSystemdDir, timerName)

	if err := os.Remove(destServiceFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove service file: %v\n", err)
	}

	if err := os.Remove(destTimerFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove timer file: %v\n", err)
	}

	// Reload systemd daemon
	if err := runSystemctl("daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}

	fmt.Println("✓ Systemd timer disabled successfully")

	return nil
}

// StatusSchedule shows the status of the systemd timer
func StatusSchedule() error {
	// Check if systemd is available
	if !isSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	// Show timer status
	cmd := exec.Command("systemctl", "--user", "status", timerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Non-zero exit code is expected if timer is not active
		fmt.Println("\nTimer is not active or not installed")
	}

	fmt.Println("\n--- Timer Schedule ---")

	// Show timer schedule
	cmd = exec.Command("systemctl", "--user", "list-timers", timerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to list timers: %w", err)
	}

	return nil
}

// ScheduleCommand handles the schedule CLI command for managing systemd timer
func ScheduleCommand(args []string) int {
	fs := flag.NewFlagSet("schedule", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dotkeeper schedule [enable|disable|status]\n\n")
		fmt.Fprintf(os.Stderr, "Manage automated backup scheduling.\n\n")
		fmt.Fprintf(os.Stderr, "Subcommands:\n")
		fmt.Fprintf(os.Stderr, "  enable   Enable the systemd timer for automated backups\n")
		fmt.Fprintf(os.Stderr, "  disable  Disable the systemd timer\n")
		fmt.Fprintf(os.Stderr, "  status   Show the status of the systemd timer\n")
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 1
	}

	if fs.NArg() < 1 {
		fs.Usage()
		return 1
	}

	subcommand := fs.Arg(0)

	switch subcommand {
	case "enable":
		if err := EnableSchedule(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	case "disable":
		if err := DisableSchedule(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	case "status":
		if err := StatusSchedule(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		fs.Usage()
		return 1
	}
}

// Helper functions

func isSystemdAvailable() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}

func getUserSystemdDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "systemd", "user"), nil
}

func runSystemctl(args ...string) error {
	fullArgs := append([]string{"--user"}, args...)
	cmd := exec.Command("systemctl", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
