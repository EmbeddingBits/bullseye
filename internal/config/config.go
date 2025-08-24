package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration
type Config struct {
	BorderColor        string `toml:"border_color"`
	StatusBarBgColor   string `toml:"status_bar_bg_color"`
	StatusBarFgColor   string `toml:"status_bar_fg_color"`
	DirColor           string `toml:"dir_color"`
	SelectedItemColor  string `toml:"selected_item_color"`
	DefaultFgColor     string `toml:"default_fg_color"`
	PreviewBgColor     string `toml:"preview_bg_color"`
	HiddenFileColor    string `toml:"hidden_file_color"`
	ExecutableColor    string `toml:"executable_color"`
	SymlinkColor       string `toml:"symlink_color"`
	PreviewBorderColor string `toml:"preview_border_color"`
	HoverBgColor       string `toml:"hover_bg_color"`
}

// LoadConfig loads configuration from file or returns default configuration
func LoadConfig() Config {
	defaultConfig := Config{
		BorderColor:        "240", // Gray
		StatusBarBgColor:   "235", // Dark gray
		StatusBarFgColor:   "255", // White
		DirColor:           "33",  // Blue
		SelectedItemColor:  "11",  // Yellow
		DefaultFgColor:     "252", // Light gray
		PreviewBgColor:     "234", // Very dark gray
		HiddenFileColor:    "244", // Dark gray
		ExecutableColor:    "46",  // Green
		SymlinkColor:       "14",  // Cyan
		PreviewBorderColor: "240", // Gray
		HoverBgColor:       "0",   // Black
	}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "bullseye", "config.toml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Try local config
		data, err = os.ReadFile("config.toml")
		if err != nil {
			return defaultConfig
		}
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return defaultConfig
	}

	// Set defaults for empty values
	if config.BorderColor == "" {
		config.BorderColor = defaultConfig.BorderColor
	}
	if config.StatusBarBgColor == "" {
		config.StatusBarBgColor = defaultConfig.StatusBarBgColor
	}
	if config.StatusBarFgColor == "" {
		config.StatusBarFgColor = defaultConfig.StatusBarFgColor
	}
	if config.DirColor == "" {
		config.DirColor = defaultConfig.DirColor
	}
	if config.SelectedItemColor == "" {
		config.SelectedItemColor = defaultConfig.SelectedItemColor
	}
	if config.DefaultFgColor == "" {
		config.DefaultFgColor = defaultConfig.DefaultFgColor
	}
	if config.PreviewBgColor == "" {
		config.PreviewBgColor = defaultConfig.PreviewBgColor
	}
	if config.HiddenFileColor == "" {
		config.HiddenFileColor = defaultConfig.HiddenFileColor
	}
	if config.ExecutableColor == "" {
		config.ExecutableColor = defaultConfig.ExecutableColor
	}
	if config.SymlinkColor == "" {
		config.SymlinkColor = defaultConfig.SymlinkColor
	}
	if config.PreviewBorderColor == "" {
		config.PreviewBorderColor = defaultConfig.PreviewBorderColor
	}
	if config.HoverBgColor == "" {
		config.HoverBgColor = defaultConfig.HoverBgColor
	}

	return config
}
