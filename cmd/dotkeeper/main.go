package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo/dotkeeper/internal/cli"
	"github.com/diogo/dotkeeper/internal/tui"
)

const version = "0.1.0"

func main() {
	helpFlag := flag.Bool("help", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Printf("dotkeeper version %s\n", version)
		os.Exit(0)
	}

	// If no commands provided, launch TUI
	if flag.NArg() == 0 {
		p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Route CLI commands
	command := flag.Arg(0)
	args := flag.Args()[1:]

	var exitCode int
	switch command {
	case "backup":
		exitCode = cli.BackupCommand(args)
	case "restore":
		exitCode = cli.RestoreCommand(args)
	case "list":
		exitCode = cli.ListCommand(args)
	case "delete":
		exitCode = cli.DeleteCommand(args)
	case "config":
		exitCode = cli.ConfigCommand(args)
	case "history":
		exitCode = cli.HistoryCommand(args)
	case "schedule":
		exitCode = cli.ScheduleCommand(args)
	case "help":
		printHelp()
		exitCode = 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Use 'dotkeeper --help' for usage information\n")
		exitCode = 1
	}

	os.Exit(exitCode)
}

func printHelp() {
	help := `dotkeeper - dotfiles backup manager

Usage:
  dotkeeper [command] [options]

Commands:
  backup      Create a backup of dotfiles
  restore     Restore dotfiles from backup
  list        List available backups
  delete      Delete a backup
  config      Manage configuration
  history     Show operation history
  schedule    Manage automated backup scheduling
  help        Show this help message

Options:
  --help      Show this help message
  --version   Show version information

Examples:
  dotkeeper backup
  dotkeeper restore --backup-id <id>
  dotkeeper list
  dotkeeper schedule enable`
	fmt.Println(help)
}
