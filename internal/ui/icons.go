package ui

import (
	"path/filepath"
	"strings"

	"github.com/embeddingbits/file_viewer/pkg/models"
)

// GetFileIcon returns the appropriate icon for a file or directory.
// It uses Nerd Font icons for graphical representation.
func GetFileIcon(file models.FileInfo) string {
	name := strings.ToLower(file.Entry.Name())
	ext := strings.ToLower(filepath.Ext(file.Entry.Name()))

	if file.Entry.IsDir() {
		// Special directory icons
		switch filepath.Base(name) {
		case ".git":
			return "" // nf-dev-git
		case ".config":
			return "" // nf-fa-cogs
		case "node_modules":
			return "" // nf-dev-nodejs_small
		case "downloads":
			return "" // nf-fa-download
		case "documents":
			return "" // nf-fa-folder_open
		case "pictures", "images":
			return "" // nf-fa-picture_o
		case "music", "audio":
			return "🎵" // nf-fa-music
		case "videos", "movies":
			return "" // nf-fa-video_camera
		case "desktop":
			return "" // nf-fa-desktop
		case "home":
			return "" // nf-fa-home
		default:
			return "" // nf-fa-folder (Default directory)
		}
	}

	// Special file names (exact match)
	switch name {
	case "dockerfile":
		return "" // nf-dev-docker
	case "docker-compose.yml", "docker-compose.yaml":
		return "" // nf-dev-docker
	case "makefile":
		return "" // nf-fa-terminal
	case "cmakelists.txt":
		return "" // nf-fa-upload (Represents build/compile)
	case "readme.md", "readme.txt", "readme":
		return "" // nf-fa-book
	case "license", "licence":
		return "" // nf-fa-balance_scale
	case ".gitignore":
		return "" // nf-dev-git
	case ".gitconfig":
		return "" // nf-fa-cog
	case "package.json":
		return "" // nf-dev-npm
	case "cargo.toml":
		return "" // nf-dev-rust
	case "go.mod":
		return "" // nf-dev-go
	case "requirements.txt":
		return "" // nf-dev-python
	}

	// File extensions
	switch ext {
	// Programming languages
	case ".go":
		return "" // nf-dev-go
	case ".py":
		return "" // nf-dev-python
	case ".js", ".mjs":
		return "" // nf-dev-javascript
	case ".ts":
		return "" // nf-dev-typescript
	case ".jsx", ".tsx":
		return "" // nf-dev-react
	case ".html", ".htm":
		return "" // nf-dev-html5
	case ".css":
		return "" // nf-dev-css3
	case ".scss", ".sass":
		return "" // nf-dev-sass
	case ".less":
		return "" // nf-dev-less
	case ".vue":
		return "﵂" // nf-dev-vue
	case ".php":
		return "" // nf-dev-php
	case ".rb":
		return "" // nf-dev-ruby
	case ".java":
		return "" // nf-fae-java
	case ".c":
		return "" // nf-custom-c
	case ".cpp", ".cc", ".cxx":
		return "" // nf-custom-cpp
	case ".h", ".hpp":
		return "" // nf-fa-header
	case ".cs":
		return "" // nf-dev-csharp
	case ".rs":
		return "" // nf-dev-rust
	case ".swift":
		return "" // nf-dev-swift
	case ".kt":
		return "" // nf-dev-kotlin
	case ".scala":
		return "" // nf-dev-scala
	case ".clj", ".cljs":
		return "" // nf-dev-clojure
	case ".hs":
		return "" // nf-dev-haskell
	case ".elm":
		return "" // nf-dev-elm
	case ".lua":
		return "" // nf-dev-lua
	case ".r":
		return "" // nf-mdi-language_r
	case ".sql":
		return "" // nf-fa-database
	case ".sh", ".bash", ".zsh", ".fish":
		return "" // nf-fa-terminal
	case ".ps1":
		return "" // nf-mdi-powershell
	case ".bat", ".cmd":
		return "" // nf-fa-windows

	// Markup and data
	case ".md", ".markdown":
		return "" // nf-dev-markdown
	case ".json":
		return "ﬥ" // nf-mdi-json
	case ".yaml", ".yml":
		return "" // nf-fa-file_code_o
	case ".toml":
		return "" // nf-fa-file_code_o
	case ".xml":
		return "" // nf-fa-code
	case ".csv":
		return "" // nf-fa-table
	case ".ini", ".cfg", ".conf":
		return "" // nf-fa-cogs
	case ".env":
		return "" // nf-fa-key

	// Images
	case ".jpg", ".jpeg", ".png", ".gif", ".svg", ".ico", ".bmp", ".webp", ".tiff", ".tif":
		return "" // nf-fa-file_image_o

	// Audio
	case ".mp3", ".wav", ".flac", ".ogg", ".m4a", ".aac":
		return "" // nf-fa-file_audio_o

	// Video
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm":
		return "" // nf-fa-file_video_o

	// Archives
	case ".zip", ".tar", ".tgz", ".tar.gz", ".gz", ".rar", ".7z":
		return "" // nf-fa-file_archive_o

	// Documents
	case ".pdf":
		return "" // nf-fa-file_pdf_o
	case ".doc", ".docx":
		return "" // nf-fa-file_word_o
	case ".xls", ".xlsx":
		return "" // nf-fa-file_excel_o
	case ".ppt", ".pptx":
		return "" // nf-fa-file_powerpoint_o
	case ".odt", ".ods", ".odp":
		return "" // Using Word icon as a generic document
	case ".rtf":
		return "" // nf-fa-file_text_o

	// Fonts
	case ".ttf", ".otf", ".woff", ".woff2":
		return "" // nf-fa-font

	// Others
	case ".txt":
		return "" // nf-fa-file_text_o
	case ".log":
		return "" // nf-fa-list_alt
	case ".lock":
		return "" // nf-fa-lock
	case ".tmp":
		return "" // nf-fa-trash
	case ".bak":
		return "" // nf-fa-save
	case ".iso", ".dmg":
		return "" // nf-fa-hdd_o
	case ".exe", ".msi":
		return "" // nf-fa-windows
	case ".app":
		return "" // nf-fa-apple
	case ".deb":
		return "" // nf-dev-debian
	case ".rpm":
		return "" // nf-dev-redhat

	default:
		return "" // nf-fa-file_o (Default file)
	}
}

