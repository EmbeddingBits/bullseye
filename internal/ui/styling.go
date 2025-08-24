package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/embeddingbits/file_viewer/internal/config"
	"github.com/embeddingbits/file_viewer/pkg/models"
)

// GetFileStyle returns the appropriate style for a file or directory
func GetFileStyle(file models.FileInfo, isSelected bool, cfg config.Config) lipgloss.Style {
	var color string

	if file.IsHidden {
		color = cfg.HiddenFileColor
	} else if file.Entry.IsDir() {
		color = cfg.DirColor
	} else {
		// Check if executable
		if info, err := file.Entry.Info(); err == nil && info.Mode()&0111 != 0 {
			color = cfg.ExecutableColor
		} else {
			color = cfg.DefaultFgColor
		}
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	if isSelected {
		// Use foreground color with configured hover background instead of highlighting
		style = style.Foreground(lipgloss.Color(color)).Background(lipgloss.Color(cfg.HoverBgColor)).Bold(false)
	}

	return style
}

// GetBorderStyle returns the border style for panes
func GetBorderStyle(cfg config.Config) lipgloss.Style {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(cfg.BorderColor))
}

// GetPreviewBorderStyle returns the border style for the preview pane
func GetPreviewBorderStyle(cfg config.Config) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(cfg.PreviewBorderColor)).
		Background(lipgloss.Color(cfg.PreviewBgColor))
}

// GetStatusStyle returns the style for the status bar
func GetStatusStyle(cfg config.Config, width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color(cfg.StatusBarBgColor)).
		Foreground(lipgloss.Color(cfg.StatusBarFgColor)).
		Padding(0, 1)
}

// GetHelpStyle returns the style for the help bar
func GetHelpStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("248")).
		Padding(0, 1)
}

// TruncateString truncates a string to fit within the specified width
func TruncateString(s string, width int) string {
	if len(s) <= width {
		return s
	}
	return s[:width-3] + "..."
}

// FormatFileName formats a file name with size information
func FormatFileName(file models.FileInfo, maxWidth int, showSize bool) string {
	name := file.Entry.Name()

	if showSize && !file.Entry.IsDir() {
		// Import the fileutils package to use FormatSize
		// For now, we'll use a simple approach
		sizeStr := " (unknown)"
		if file.Size > 0 {
			const unit = 1024
			if file.Size < unit {
				sizeStr = fmt.Sprintf(" (%d B)", file.Size)
			} else {
				div, exp := int64(unit), 0
				for n := file.Size / unit; n >= unit; n /= unit {
					div *= unit
					exp++
				}
				sizeStr = fmt.Sprintf(" (%.1f %cB)", float64(file.Size)/float64(div), "KMGTPE"[exp])
			}
		}
		name += sizeStr
	}

	if len(name) > maxWidth {
		// Calculate how much space the size info takes
		sizeInfoLen := 0
		if showSize && !file.Entry.IsDir() {
			sizeInfoLen = len(" (xxx.x KB)")
		}

		// Truncate the name part, leaving space for size info
		truncateLen := maxWidth - sizeInfoLen - 3 // 3 for "..."
		if truncateLen < 1 {
			truncateLen = 1
		}
		name = name[:truncateLen] + "..."
	}

	return name
}
