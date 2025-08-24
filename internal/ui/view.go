package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/embeddingbits/file_viewer/internal/config"
	"github.com/embeddingbits/file_viewer/pkg/models"
)

// RenderView renders the complete application view
func RenderView(m *models.Model, cfg config.Config) string {
	if m.Err != nil {
		return fmt.Sprintf("Error: %v\nPress 'q' to quit.", m.Err)
	}

	if m.Width == 0 || m.Height == 0 {
		return "Initializing..."
	}

	// Calculate pane widths
	parentWidth := max(m.Width/4, 15)
	currentWidth := max(m.Width/3, 20)
	previewWidth := max(m.Width-parentWidth-currentWidth-4, 20)

	visibleHeight := getVisibleHeight(m.Height)

	// Parent directory pane
	parentPane := renderParentPane(m, cfg, parentWidth, visibleHeight)

	// Current directory pane
	currentPane := renderCurrentPane(m, cfg, currentWidth, visibleHeight)

	// Preview pane
	previewPane := renderPreviewPane(m, cfg, previewWidth, visibleHeight)

	// Join panes horizontally
	panes := lipgloss.JoinHorizontal(lipgloss.Top, parentPane, currentPane, previewPane)

	// Status bar
	status := renderStatusBar(m, cfg)

	// Help bar
	help := renderHelpBar(m, cfg)

	// Full view
	return lipgloss.JoinVertical(lipgloss.Left, panes, status, help)
}

// renderParentPane renders the parent directory pane
func renderParentPane(m *models.Model, cfg config.Config, width, height int) string {
	var content strings.Builder
	if m.ParentFiles != nil && len(m.ParentFiles) > 0 {
		content.WriteString(fmt.Sprintf(" %s\n", filepath.Base(m.ParentDir)))
		content.WriteString(strings.Repeat("─", width-2) + "\n")

		for i, file := range m.ParentFiles {
			if i >= height-2 {
				break
			}

			icon := GetFileIcon(file)
			name := file.Entry.Name()
			if len(name) > width-6 {
				name = name[:width-9] + "..."
			}

			style := GetFileStyle(file, i == m.ParentSelected, cfg)
			line := fmt.Sprintf("%s %s", icon, name)
			content.WriteString(style.Render(line) + "\n")
		}
	}

	borderStyle := GetBorderStyle(cfg)
	return borderStyle.Width(width).Height(height).Render(content.String())
}

// renderCurrentPane renders the current directory pane
func renderCurrentPane(m *models.Model, cfg config.Config, width, height int) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf(" %s (%d items)\n", filepath.Base(m.CurrentDir), len(m.Files)))
	content.WriteString(strings.Repeat("─", width-2) + "\n")

	if len(m.Files) == 0 {
		content.WriteString("Empty directory")
	} else {
		start := m.ListOffset
		end := min(start+height-2, len(m.Files))

		for i := start; i < end; i++ {
			file := m.Files[i]
			icon := GetFileIcon(file)
			name := file.Entry.Name()

			// Add size info for files
			sizeInfo := ""
			if !file.Entry.IsDir() {
				sizeInfo = fmt.Sprintf(" (%s)", FormatSize(file.Size))
			}

			fullName := name + sizeInfo
			if len(fullName) > width-6 {
				name = name[:max(1, width-9-len(sizeInfo))] + "..."
				fullName = name + sizeInfo
			}

			style := GetFileStyle(file, i == m.Selected, cfg)
			line := fmt.Sprintf("%s %s", icon, fullName)
			content.WriteString(style.Render(line) + "\n")
		}
	}

	borderStyle := GetBorderStyle(cfg)
	return borderStyle.Width(width).Height(height).Render(content.String())
}

// renderPreviewPane renders the preview pane
func renderPreviewPane(m *models.Model, cfg config.Config, width, height int) string {
	var content strings.Builder
	if m.Preview != "" {
		lines := strings.Split(m.Preview, "\n")
		start := m.PreviewOffset
		end := min(start+height, len(lines))

		for i := start; i < end; i++ {
			line := lines[i]
			if len(line) > width-2 {
				line = line[:width-5] + "..."
			}
			content.WriteString(line + "\n")
		}
	}

	previewBorderStyle := GetPreviewBorderStyle(cfg)
	return previewBorderStyle.Width(width).Height(height).Render(content.String())
}

// renderStatusBar renders the status bar
func renderStatusBar(m *models.Model, cfg config.Config) string {
	var statusText string
	if m.SearchMode {
		statusText = fmt.Sprintf("Search: %s", m.SearchQuery)
	} else {
		sortIndicator := "↑"
		if m.ReverseSort {
			sortIndicator = "↓"
		}

		statusParts := []string{
			fmt.Sprintf("Dir: %s", m.CurrentDir),
			fmt.Sprintf("Sort: %s%s", m.SortBy, sortIndicator),
		}

		if m.ShowHidden {
			statusParts = append(statusParts, "Hidden: ON")
		}

		if len(m.Files) > 0 {
			statusParts = append(statusParts, fmt.Sprintf("%d/%d", m.Selected+1, len(m.Files)))
		}

		statusText = strings.Join(statusParts, " | ")
	}

	statusStyle := GetStatusStyle(cfg, m.Width)
	return statusStyle.Render(statusText)
}

// renderHelpBar renders the help bar
func renderHelpBar(m *models.Model, cfg config.Config) string {
	helpText := "q:quit | h/l:nav | j/k:up/down | o:open | .:hidden | s:size | t:time | n:name | /:search | r:refresh"
	if m.SearchMode {
		helpText = "Type to search | Enter:confirm | Esc:cancel"
	}

	helpStyle := GetHelpStyle(m.Width)
	return helpStyle.Render(helpText)
}

// Helper functions
func getVisibleHeight(height int) int {
	return max(1, height-4) // Account for borders and status bar
}

// FormatSize formats file size in human-readable format
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
