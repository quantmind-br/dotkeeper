package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/diogo/dotkeeper/internal/history"
	"github.com/diogo/dotkeeper/internal/pathutil"
)

// HistoryCommand handles the history subcommand
func HistoryCommand(args []string) int {
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	opType := fs.String("type", "", "Filter by operation type (backup or restore)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dotkeeper history [--json] [--type TYPE]\n\n")
		fmt.Fprintf(os.Stderr, "Show operation history.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 1
	}

	// Create history store
	store, err := history.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing history: %v\n", err)
		return 1
	}

	// Read history entries
	var entries []history.HistoryEntry
	if *opType != "" {
		entries, err = store.ReadByType(*opType, 50)
	} else {
		entries, err = store.Read(50)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading history: %v\n", err)
		return 1
	}

	// Output
	if *jsonOutput {
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			return 1
		}
		fmt.Println(string(data))
	} else {
		if len(entries) == 0 {
			fmt.Println("No operations recorded yet.")
		} else {
			printHistoryTable(entries)
		}
	}

	return 0
}

// printHistoryTable prints history entries in a formatted table
func printHistoryTable(entries []history.HistoryEntry) {
	fmt.Printf("%-20s %-10s %-10s %-8s %-12s %-10s\n",
		"TIMESTAMP", "OPERATION", "STATUS", "FILES", "SIZE", "DURATION")
	fmt.Println(strings.Repeat("-", 80))

	for _, entry := range entries {
		timestamp := entry.Timestamp.Local().Format("2006-01-02 15:04:05")
		operation := entry.Operation
		status := entry.Status
		fileCount := fmt.Sprintf("%d", entry.FileCount)
		if entry.FileCount == 0 && entry.Status == "error" {
			fileCount = "-"
		}
		size := pathutil.FormatSize(entry.TotalSize)
		if entry.TotalSize == 0 && entry.Status == "error" {
			size = "-"
		}
		duration := formatDuration(entry.DurationMs)
		if entry.DurationMs == 0 && entry.Status == "error" {
			duration = "-"
		}

		fmt.Printf("%-20s %-10s %-10s %-8s %-12s %-10s\n",
			timestamp,
			operation,
			status,
			fileCount,
			size,
			duration,
		)
	}

	fmt.Printf("\nTotal: %d operation(s)\n", len(entries))
}

// formatDuration formats milliseconds to a human-readable string
func formatDuration(ms int64) string {
	if ms == 0 {
		return "0ms"
	}
	d := time.Duration(ms) * time.Millisecond
	if d < time.Second {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
