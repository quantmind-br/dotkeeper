package main

import (
	"flag"
	"fmt"
	"os"
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

	fmt.Println("dotkeeper - dotfiles backup manager")
	fmt.Println("Use --help for more information")
}

func printHelp() {
	fmt.Println(`dotkeeper - dotfiles backup manager

Usage:
  dotkeeper [command] [options]

Commands:
  backup      Create a backup of dotfiles
  restore     Restore dotfiles from backup
  list        List available backups
  config      Manage configuration
  help        Show this help message

Options:
  --help      Show this help message
  --version   Show version information

Examples:
  dotkeeper backup
  dotkeeper restore --backup-id <id>
  dotkeeper list
`)
}
