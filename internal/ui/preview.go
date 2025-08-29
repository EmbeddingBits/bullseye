package ui

import (
	"fmt"
	"image"
	// Import decoders for desired image formats
	_ "image/jpeg"
	_ "image/png"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"os"
	"path/filepath"
	"strings"

	"github.com/embeddingbits/file_viewer/internal/fileutils"
	"github.com/embeddingbits/file_viewer/pkg/models"
	"github.com/qeesung/image2ascii/convert"
)

// isImageFileByExtension helper detects a wide range of common image formats.
func isImageFileByExtension(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".bmp", ".tif", ".tiff", ".webp":
		return true
	default:
		return false
	}
}

// UpdatePreview is the main entry point to update the preview pane content.
func UpdatePreview(m *models.Model) {
	if len(m.Files) == 0 {
		m.Preview = "No Items"
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

// updateDirectoryPreview shows the contents of a selected directory.
func updateDirectoryPreview(m *models.Model, selectedFile models.FileInfo, fullPath string) {
	// ... (This function is unchanged)
	subFiles, err := fileutils.ReadDirWithInfo(fullPath)
	if err != nil {
		m.Preview = fmt.Sprintf("Error: %v", err)
		return
	}
	filtered := fileutils.FilterFiles(subFiles, m.ShowHidden, m.SearchQuery)
	fileutils.SortFiles(filtered, m.SortBy, m.ReverseSort)

	var sb strings.Builder
	for i, f := range filtered {
		if i >= 100 {
			sb.WriteString("... and more files")
			break
		}
		icon := GetFileIcon(f)
		sb.WriteString(fmt.Sprintf("%s %s\n", icon, f.Entry.Name()))
	}
	m.Preview = sb.String()
}

// updateFilePreview handles rendering for image, text, and binary files.
func updateFilePreview(m *models.Model, selectedFile models.FileInfo, fullPath string) {
	fileName := selectedFile.Entry.Name()

	// --- ASPECT-RATIO-PRESERVING IMAGE RENDERING LOGIC ---
	if isImageFileByExtension(fileName) {
		file, err := os.Open(fullPath)
		if err != nil {
			m.Preview = fmt.Sprintf("Error opening image: %v", err)
			return
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			renderBinaryPreview(m, selectedFile, fullPath)
			return
		}

		// --- NEW: SOPHISTICATED SIZING LOGIC ---

		// 1. Calculate available content space within the pane's borders.
		parentWidth := max(m.Width/4, 15)
		currentWidth := max(m.Width/3, 20)
		paneWidth := max(m.Width-parentWidth-currentWidth-4, 20)
		paneHeight := max(1, m.Height-4)
		contentWidth := max(1, paneWidth-2)
		contentHeight := max(1, paneHeight-2)

		// 2. Get original image dimensions.
		imageWidth := img.Bounds().Dx()
		imageHeight := img.Bounds().Dy()

		// 3. Define the aspect ratio of a terminal character (they are taller than wide).
		//    The value 0.55 is a good approximation.
		charRatio := 0.55

		// 4. Calculate the visual aspect ratio of the image and the pane.
		//    We adjust the image's ratio to account for the non-square character cells.
		imageAspect := (float64(imageWidth) / float64(imageHeight)) / charRatio
		paneAspect := float64(contentWidth) / float64(contentHeight)

		var finalWidth, finalHeight int

		// 5. Compare ratios to decide whether to fit to width or height.
		if imageAspect > paneAspect {
			// The image is "wider" than the pane, so we're limited by the pane's width.
			finalWidth = contentWidth
			finalHeight = int(float64(finalWidth) / imageAspect)
		} else {
			// The image is "taller" than the pane, so we're limited by the pane's height.
			finalHeight = contentHeight
			finalWidth = int(float64(finalHeight) * imageAspect)
		}

		// 6. Set converter options with our perfectly calculated dimensions.
		converter := convert.NewImageConverter()
		options := convert.DefaultOptions
		options.Colored = false // Still rendering as monochrome per last request
		options.FixedWidth = max(1, finalWidth)   // Ensure width is at least 1
		options.FixedHeight = max(1, finalHeight) // Ensure height is at least 1

		asciiStr := converter.Image2ASCIIString(img, &options)
		m.Preview = asciiStr
		return
	}

	// Fallback for non-image files.
	renderBinaryPreview(m, selectedFile, fullPath)
}

// renderBinaryPreview shows file info and a hex dump.
func renderBinaryPreview(m *models.Model, selectedFile models.FileInfo, fullPath string) {
	// ... (This function is unchanged)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		m.Preview = fmt.Sprintf("Error reading file: %v", err)
		return
	}

	fileName := selectedFile.Entry.Name()
	isText := fileutils.IsTextFileByExtension(fileName)
	if !isText {
		isText = fileutils.IsLikelyTextFile(content)
	}

	var sb strings.Builder
	icon := GetFileIcon(selectedFile)
	sb.WriteString(fmt.Sprintf("%s %s\n", icon, selectedFile.Entry.Name()))
	sb.WriteString(fmt.Sprintf("Size: %s\n", fileutils.FormatSize(selectedFile.Size)))
	sb.WriteString(fmt.Sprintf("Modified: %s\n", selectedFile.ModTime.Format("2006-01-02 15:04:05")))
	if fileInfo, err := os.Stat(fullPath); err == nil {
		sb.WriteString(fmt.Sprintf("Mode: %s\n", fileInfo.Mode().String()))
	}
	sb.WriteString("\n")

	if isText && len(content) > 0 {
		contentStr := string(content)
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
		sb.WriteString("Binary file - hex preview:\n\n")
		hexBytes := content
		if len(hexBytes) > 256 {
			hexBytes = hexBytes[:256]
		}
		for i := 0; i < len(hexBytes); i += 16 {
			sb.WriteString(fmt.Sprintf("%08x: ", i))
			end := min(i+16, len(hexBytes))
			for j := i; j < end; j++ {
				sb.WriteString(fmt.Sprintf("%02x ", hexBytes[j]))
			}
			sb.WriteString(strings.Repeat("   ", i+16-end))
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
