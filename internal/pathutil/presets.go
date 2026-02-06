package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

// DotfilePreset represents a pre-defined dotfile or folder configuration
type DotfilePreset struct {
	Path      string // Display path like "~/.zshrc"
	FullPath  string // Expanded path like "/home/user/.zshrc"
	Exists    bool
	IsDir     bool
	Size      int64
	FileCount int  // For directories only
	Selected  bool // For checklist toggling
}

// PresetFiles contains common dotfiles to look for
var PresetFiles = []string{
	"~/.bashrc", "~/.zshrc", "~/.bash_profile", "~/.profile",
	"~/.gitconfig", "~/.gitignore_global", "~/.vimrc",
	"~/.tmux.conf", "~/.ssh/config", "~/.gnupg/gpg.conf",
}

// PresetFolders contains common configuration folders to look for
var PresetFolders = []string{
	"~/.config/nvim", "~/.config/kitty", "~/.config/alacritty",
	"~/.config/hypr", "~/.config/fish", "~/.config/starship",
	"~/.config/waybar", "~/.config/rofi", "~/.config/wezterm",
	"~/.config/zsh",
}

// DetectDotfiles checks which presets exist on the filesystem.
// homeDir parameter allows testing with temp dirs.
// Returns (filePresets, folderPresets) — only presets that EXIST are returned.
func DetectDotfiles(homeDir string) ([]DotfilePreset, []DotfilePreset) {
	var files []DotfilePreset
	var folders []DotfilePreset

	shell := os.Getenv("SHELL")

	// Process files
	for _, p := range PresetFiles {
		fullPath := replaceHome(p, homeDir)
		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() {
			preset := DotfilePreset{
				Path:     p,
				FullPath: fullPath,
				Exists:   true,
				IsDir:    false,
				Size:     info.Size(),
			}

			// Auto-select based on shell
			if strings.Contains(shell, "zsh") && strings.Contains(p, ".zshrc") {
				preset.Selected = true
			}
			if strings.Contains(shell, "bash") && (strings.Contains(p, ".bashrc") || strings.Contains(p, ".bash_profile")) {
				preset.Selected = true
			}

			files = append(files, preset)
		}
	}

	// Process folders
	for _, p := range PresetFolders {
		fullPath := replaceHome(p, homeDir)
		info, err := os.Stat(fullPath)
		if err == nil && info.IsDir() {
			preset := DotfilePreset{
				Path:     p,
				FullPath: fullPath,
				Exists:   true,
				IsDir:    true,
			}

			// Calculate size and count files
			entries, err := os.ReadDir(fullPath)
			if err == nil {
				var size int64
				var count int
				for _, entry := range entries {
					if !entry.IsDir() {
						count++
						if info, err := entry.Info(); err == nil {
							size += info.Size()
						}
					}
				}
				preset.FileCount = count
				preset.Size = size
			}

			folders = append(folders, preset)
		}
	}

	return files, folders
}

// replaceHome replaces "~/" prefix with the provided home directory
func replaceHome(path, homeDir string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	return path
}
