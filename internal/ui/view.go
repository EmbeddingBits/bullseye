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

		// Calculate the usable space for content inside the pane's borders.
		paneContentWidth := width - 2
		if paneContentWidth < 0 {
			paneContentWidth = 0
		}

		for i, file := range m.ParentFiles {
			if i >= height-2 {
				break
			}

			icon := GetFileIcon(file)
			name := file.Entry.Name()

			// Calculate the max width the filename can have.
			maxNameWidth := paneContentWidth - len(icon) - 1 // -1 for the space

			// Truncate the name ONLY if it exceeds the calculated max width.
			if len(name) > maxNameWidth {
				if maxNameWidth > 3 {
					name = name[:maxNameWidth-3] + "..."
				} else if maxNameWidth > 0 {
					name = name[:maxNameWidth] // Not enough space for "...", so just chop.
				} else {
					name = "" // No space for the name at all.
				}
			}

			style := GetFileStyle(file, i == m.ParentSelected, cfg)
			line := fmt.Sprintf("%s %s", icon, name)
			content.WriteString(style.Render(line) + "\n")
		}
	}

	borderStyle := GetBorderStyle(cfg)
	// Force the pane to the exact width and height.
	return borderStyle.Width(width).Height(height).Render(content.String())
}

// --- CORRECTED: Current Pane Renderer ---
// This version uses the same robust truncation logic as the parent pane.

func renderCurrentPane(m *models.Model, cfg config.Config, width, height int) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf(" %s (%d items)\n", filepath.Base(m.CurrentDir), len(m.Files)))
	content.WriteString(strings.Repeat("─", width-2) + "\n")

	if len(m.Files) == 0 {
		content.WriteString(" No Items")
	} else {
		start := m.ListOffset
		end := min(start+height-2, len(m.Files))

		paneContentWidth := width - 2
		if paneContentWidth < 0 {
			paneContentWidth = 0
		}

		for i := start; i < end; i++ {
			file := m.Files[i]
			icon := GetFileIcon(file)
			name := file.Entry.Name()

			maxNameWidth := paneContentWidth - len(icon) - 1

			if len(name) > maxNameWidth {
				if maxNameWidth > 3 {
					name = name[:maxNameWidth-3] + "..."
				} else if maxNameWidth > 0 {
					name = name[:maxNameWidth]
				} else {
					name = ""
				}
			}

			style := GetFileStyle(file, i == m.Selected, cfg)
			line := fmt.Sprintf("%s %s", icon, name)
			content.WriteString(style.Render(line) + "\n")
		}
	}

	borderStyle := GetBorderStyle(cfg)
	return borderStyle.Width(width).Height(height).Render(content.String())
}


// --- CORRECTED: Preview Pane Renderer ---
// This version truncates long lines within the text preview.

func renderPreviewPane(m *models.Model, cfg config.Config, width, height int) string {
	var content strings.Builder
	if m.Preview != "" {
		lines := strings.Split(m.Preview, "\n")
		start := m.PreviewOffset
		end := min(start+height-2, len(lines)) // -2 to account for border

		paneContentWidth := width - 2 // Space inside the borders
		if paneContentWidth < 0 {
			paneContentWidth = 0
		}
		
		for i := start; i < end; i++ {
			line := lines[i]
			// Truncate oversized preview lines.
			if len(line) > paneContentWidth {
				if paneContentWidth > 3 {
					line = line[:paneContentWidth-3] + "..."
				} else {
					line = line[:paneContentWidth]
				}
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

		selectedFile := m.Files[m.Selected]
		statusParts := []string{
			fmt.Sprintf("Dir: %s", selectedFile.Entry.Name()),
			fmt.Sprintf("Sort: %s%s", m.SortBy, sortIndicator),
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
