package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/embeddingbits/file_viewer/internal/config"
	"github.com/embeddingbits/file_viewer/pkg/models"
)

type StatusBarContent struct {
	IsSearchMode bool
	SearchQuery  string
	Directory    string
	SortInfo     string
	FileCount    string
	Permissions  string // To hold file mode like "-rwxr-xr-x"
}

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

	// Panes
	parentPane := renderParentPane(m, cfg, parentWidth, visibleHeight)
	currentPane := renderCurrentPane(m, cfg, currentWidth, visibleHeight)
	previewPane := renderPreviewPane(m, cfg, previewWidth, visibleHeight)
	panes := lipgloss.JoinHorizontal(lipgloss.Top, parentPane, currentPane, previewPane)

	// --- MODIFIED: Status Bar Rendering Layout ---
	statusBarContent := getStatusBarContent(m, cfg)
	statusStyle := GetStatusStyle(cfg, m.Width)

	var status string
	if statusBarContent.IsSearchMode {
		status = statusStyle.Render(statusBarContent.SearchQuery)
	} else {
		// Left side of the status bar contains Directory and Sort info.
		leftStatus := strings.Join([]string{statusBarContent.Directory, statusBarContent.SortInfo}, "")
		
		// Right side now contains Permissions and File Count.
		var rightItems []string
		if statusBarContent.Permissions != "" {
			rightItems = append(rightItems, statusBarContent.Permissions)
		}
		if statusBarContent.FileCount != "" {
			rightItems = append(rightItems, statusBarContent.FileCount)
		}
		rightStatus := strings.Join(rightItems, " | ")
		
		// Create the flexible gap in between
		gapWidth := m.Width - lipgloss.Width(leftStatus) - lipgloss.Width(rightStatus) - 2 // -2 for style padding
		if gapWidth < 0 {
			gapWidth = 0
		}
		gap := strings.Repeat(" ", gapWidth)
		
		finalStatusText := lipgloss.JoinHorizontal(lipgloss.Top, leftStatus, gap, rightStatus)
		status = statusStyle.Render(finalStatusText)
	}

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
		paneContentWidth := max(0, width-2)

		for i, file := range m.ParentFiles {
			if i >= height-2 {
				break
			}
			icon := GetFileIcon(file)
			name := file.Entry.Name()
			maxNameWidth := paneContentWidth - len(icon) - 1
			if len(name) > maxNameWidth {
				if maxNameWidth > 3 {
					name = name[:maxNameWidth-3] + "..."
				} else {
					name = name[:max(0, maxNameWidth)]
				}
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
		content.WriteString(" No Items")
	} else {
		start := m.ListOffset
		end := min(start+height-2, len(m.Files))
		paneContentWidth := max(0, width-2)

		for i := start; i < end; i++ {
			file := m.Files[i]
			icon := GetFileIcon(file)
			name := file.Entry.Name()
			maxNameWidth := paneContentWidth - len(icon) - 1
			if len(name) > maxNameWidth {
				if maxNameWidth > 3 {
					name = name[:maxNameWidth-3] + "..."
				} else {
					name = name[:max(0, maxNameWidth)]
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

// renderPreviewPane renders the preview pane
func renderPreviewPane(m *models.Model, cfg config.Config, width, height int) string {
	var content strings.Builder
	if m.Preview != "" {
		lines := strings.Split(m.Preview, "\n")
		start := m.PreviewOffset
		end := min(start+height-2, len(lines))
		paneContentWidth := max(0, width-2)
		
		for i := start; i < end; i++ {
			line := lines[i]
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


func getStatusBarContent(m *models.Model, cfg config.Config) StatusBarContent {
	if m.SearchMode {
		return StatusBarContent{
			IsSearchMode: true,
			SearchQuery:  fmt.Sprintf("Search: %s", m.SearchQuery),
		}
	}
	
	var dir, fileCount, permissions string
	
	if len(m.Files) > 0 && m.Selected < len(m.Files) {
		selectedFile := m.Files[m.Selected]
		dir = fmt.Sprintf("Dir: %s", selectedFile.Entry.Name())
		fileCount = fmt.Sprintf("%d/%d", m.Selected+1, len(m.Files))
		
		if info, err := selectedFile.Entry.Info(); err == nil {
			permissions = info.Mode().String()
		}

	} else {
		dir = fmt.Sprintf("Dir: %s", filepath.Base(m.CurrentDir))
	}

	return StatusBarContent{
		IsSearchMode: false,
		Directory:    dir,
		FileCount:    fileCount,
		Permissions:  permissions,
	}
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
