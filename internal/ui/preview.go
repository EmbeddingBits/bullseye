package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/embeddingbits/file_viewer/internal/fileutils"
	"github.com/embeddingbits/file_viewer/pkg/models"
)

// UpdatePreview updates the preview content for the selected file
func UpdatePreview(m *models.Model) {
	if len(m.Files) == 0 {
		m.Preview = "Empty directory"
		return
	}

	selectedFile := m.Files[m.Selected]
	fullPath := filepath.Join(m.CurrentDir, selectedFile.Entry.Name())

	if selectedFile.Entry.IsDir() {
		updateDirectoryPreview(m, selectedFile, fullPath)
	} else {
		updateFilePreview(m, selectedFile, fullPath)
	}
}

// updateDirectoryPreview updates preview for directory selection
func updateDirectoryPreview(m *models.Model, selectedFile models.FileInfo, fullPath string) {
	subFiles, err := fileutils.ReadDirWithInfo(fullPath)
	if err != nil {
		m.Preview = fmt.Sprintf("Error: %v", err)
		return
	}

	filtered := fileutils.FilterFiles(subFiles, m.ShowHidden, m.SearchQuery)
	fileutils.SortFiles(filtered, m.SortBy, m.ReverseSort)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(" %s\n", selectedFile.Entry.Name()))
	sb.WriteString(fmt.Sprintf("Items: %d\n\n", len(filtered)))

	for i, f := range filtered {
		if i >= 100 { // Show more items in preview
			sb.WriteString("... and more files")
			break
		}
		icon := GetFileIcon(f)
		sb.WriteString(fmt.Sprintf("%s %s\n", icon, f.Entry.Name()))
	}
	m.Preview = sb.String()
}

// updateFilePreview updates preview for file selection
func updateFilePreview(m *models.Model, selectedFile models.FileInfo, fullPath string) {
	// Always try to read as text first
	content, err := os.ReadFile(fullPath)
	if err != nil {
		m.Preview = fmt.Sprintf("Error reading file: %v", err)
		return
	}

	// Check if it's a text file by extension or content
	fileName := strings.ToLower(selectedFile.Entry.Name())
	isText := fileutils.IsTextFileByExtension(fileName)

	// If not recognized by extension, check if it's likely text by content
	if !isText {
		isText = fileutils.IsLikelyTextFile(content)
	}

	var sb strings.Builder

	// Always show file info header
	icon := GetFileIcon(selectedFile)
	sb.WriteString(fmt.Sprintf("%s %s\n", icon, selectedFile.Entry.Name()))
	sb.WriteString(fmt.Sprintf("Size: %s\n", fileutils.FormatSize(selectedFile.Size)))
	sb.WriteString(fmt.Sprintf("Modified: %s\n", selectedFile.ModTime.Format("2006-01-02 15:04:05")))

	// Show file permissions/mode
	if fileInfo, err := os.Stat(fullPath); err == nil {
		sb.WriteString(fmt.Sprintf("Mode: %s\n", fileInfo.Mode().String()))
	}

	sb.WriteString("\n")

	if isText && len(content) > 0 {
		// Display text content
		contentStr := string(content)

		// Limit preview size for very large files
		if len(contentStr) > 50000 {
			lines := strings.Split(contentStr, "\n")
			if len(lines) > 500 {
				contentStr = strings.Join(lines[:500], "\n") + "\n\n... (file truncated for preview)"
			}
		}

		sb.WriteString(contentStr)
	} else if len(content) == 0 {
		sb.WriteString("(empty file)")
	} else {
		// For binary files, show hex preview and file type info
		sb.WriteString("Binary file - hex preview:\n\n")

		// Show first 256 bytes as hex
		hexBytes := content
		if len(hexBytes) > 256 {
			hexBytes = hexBytes[:256]
		}

		for i := 0; i < len(hexBytes); i += 16 {
			// Address
			sb.WriteString(fmt.Sprintf("%08x: ", i))

			// Hex bytes
			end := i + 16
			if end > len(hexBytes) {
				end = len(hexBytes)
			}

			for j := i; j < end; j++ {
				sb.WriteString(fmt.Sprintf("%02x ", hexBytes[j]))
			}

			// Padding for incomplete lines
			for j := end; j < i+16; j++ {
				sb.WriteString("   ")
			}

			// ASCII representation
			sb.WriteString(" |")
			for j := i; j < end; j++ {
				if hexBytes[j] >= 32 && hexBytes[j] <= 126 {
					sb.WriteByte(hexBytes[j])
				} else {
					sb.WriteString(".")
				}
			}
			sb.WriteString("|\n")
		}

		if len(content) > 256 {
			sb.WriteString(fmt.Sprintf("\n... (%d more bytes)", len(content)-256))
		}
	}

	m.Preview = sb.String()
}

