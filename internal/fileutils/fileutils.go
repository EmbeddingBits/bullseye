package fileutils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/embeddingbits/file_viewer/pkg/models"
)

var (
	// TextExtensions contains file extensions that are typically text files
	TextExtensions = []string{
		// Programming languages
		".txt", ".go", ".py", ".js", ".ts", ".jsx", ".tsx", ".html", ".htm", ".css", ".scss", ".sass", ".less",
		".php", ".rb", ".java", ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".cs", ".rs", ".swift", ".kt",
		".scala", ".clj", ".cljs", ".hs", ".elm", ".lua", ".r", ".sql", ".sh", ".bash", ".zsh", ".fish",
		".ps1", ".bat", ".cmd", ".vim", ".lua", ".pl", ".pm", ".awk", ".sed",

		// Markup and configuration
		".md", ".markdown", ".json", ".yaml", ".yml", ".toml", ".xml", ".csv", ".ini", ".cfg", ".conf",
		".env", ".gitignore", ".gitconfig", ".gitattributes", ".gitmodules", ".editorconfig",
		".prettierrc", ".eslintrc", ".babelrc", ".npmrc", ".yarnrc",

		// Documentation and text
		".rst", ".org", ".tex", ".bib", ".man", ".1", ".2", ".3", ".4", ".5", ".6", ".7", ".8", ".9",
		".readme", ".changelog", ".authors", ".contributors", ".copying", ".license", ".licence",
		".todo", ".fixme", ".bugs", ".news", ".thanks", ".install",

		// Web and styles
		".vue", ".svelte", ".astro", ".styl", ".stylus", ".postcss",

		// Data formats
		".tsv", ".psv", ".dsv", ".ndjson", ".jsonl", ".geojson", ".topojson",

		// Configuration files (no extension)
		"dockerfile", "makefile", "cmakelists.txt", "vagrantfile", "gemfile", "rakefile",
		"package.json", "composer.json", "cargo.toml", "pyproject.toml", "poetry.lock",
		"requirements.txt", "pipfile", "pipfile.lock", "go.mod", "go.sum",

		// Log and temporary files
		".log", ".out", ".err", ".tmp", ".temp", ".bak", ".backup", ".orig", ".swp", ".swo",

		// Others
		".pub", ".pem", ".key", ".crt", ".cer", ".p12", ".pfx", ".jks",
	}
)

// GetFileInfo creates a FileInfo struct from a directory entry
func GetFileInfo(entry fs.DirEntry, dirPath string) models.FileInfo {
	info := models.FileInfo{
		Entry:    entry,
		IsHidden: strings.HasPrefix(entry.Name(), "."),
	}

	if fileInfo, err := entry.Info(); err == nil {
		info.Size = fileInfo.Size()
		info.ModTime = fileInfo.ModTime()
	}

	return info
}

// ReadDirWithInfo reads a directory and returns FileInfo for each entry
func ReadDirWithInfo(dirPath string) ([]models.FileInfo, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	files := make([]models.FileInfo, 0, len(entries))
	for _, entry := range entries {
		files = append(files, GetFileInfo(entry, dirPath))
	}

	return files, nil
}

// SortFiles sorts files based on the specified criteria
func SortFiles(files []models.FileInfo, sortBy string, reverseSort bool) {
	sort.Slice(files, func(i, j int) bool {
		// Directories first
		if files[i].Entry.IsDir() != files[j].Entry.IsDir() {
			return files[i].Entry.IsDir()
		}

		var result bool
		switch sortBy {
		case "size":
			result = files[i].Size < files[j].Size
		case "modified":
			result = files[i].ModTime.Before(files[j].ModTime)
		default: // name
			result = strings.ToLower(files[i].Entry.Name()) < strings.ToLower(files[j].Entry.Name())
		}

		if reverseSort {
			return !result
		}
		return result
	})
}

// FilterFiles filters files based on hidden status and search query
func FilterFiles(files []models.FileInfo, showHidden bool, searchQuery string) []models.FileInfo {
	if showHidden && searchQuery == "" {
		return files
	}

	filtered := make([]models.FileInfo, 0, len(files))
	for _, file := range files {
		// Filter hidden files
		if !showHidden && file.IsHidden {
			continue
		}

		// Filter by search query
		if searchQuery != "" {
			if !strings.Contains(strings.ToLower(file.Entry.Name()), strings.ToLower(searchQuery)) {
				continue
			}
		}

		filtered = append(filtered, file)
	}

	return filtered
}

// IsLikelyTextFile detects if content is likely text based on binary analysis
func IsLikelyTextFile(content []byte) bool {
	if len(content) == 0 {
		return true
	}

	// Check first 512 bytes for null bytes (common in binary files)
	checkBytes := content
	if len(checkBytes) > 512 {
		checkBytes = checkBytes[:512]
	}

	nullCount := 0
	for _, b := range checkBytes {
		if b == 0 {
			nullCount++
		}
	}

	// If more than 1% null bytes, likely binary
	if float64(nullCount)/float64(len(checkBytes)) > 0.01 {
		return false
	}

	// Check for mostly printable characters
	printableCount := 0
	for _, b := range checkBytes {
		if (b >= 32 && b <= 126) || b == '\t' || b == '\n' || b == '\r' {
			printableCount++
		}
	}

	// If more than 95% printable, likely text
	return float64(printableCount)/float64(len(checkBytes)) > 0.95
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

// IsTextFileByExtension checks if a file is text based on its extension
func IsTextFileByExtension(fileName string) bool {
	fileName = strings.ToLower(fileName)
	ext := strings.ToLower(filepath.Ext(fileName))

	// Check by extension first
	for _, textExt := range TextExtensions {
		if ext == textExt || strings.HasSuffix(fileName, strings.ToLower(textExt)) {
			return true
		}
	}

	return false
}
