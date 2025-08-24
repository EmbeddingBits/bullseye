package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pelletier/go-toml/v2"
)

var (
	textExtensions = []string{
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

type FileInfo struct {
	Entry    fs.DirEntry
	Size     int64
	ModTime  time.Time
	IsHidden bool
}

type model struct {
	currentDir    string
	parentDir     string
	files         []FileInfo
	parentFiles   []FileInfo
	selected      int
	parentSelected int
	listOffset    int
	preview       string
	previewOffset int
	width         int
	height        int
	err           error
	config        Config
	showHidden    bool
	sortBy        string // "name", "size", "modified"
	reverseSort   bool
	searchMode    bool
	searchQuery   string
}

type Config struct {
	BorderColor         string `toml:"border_color"`
	StatusBarBgColor    string `toml:"status_bar_bg_color"`
	StatusBarFgColor    string `toml:"status_bar_fg_color"`
	DirColor           string `toml:"dir_color"`
	SelectedItemColor   string `toml:"selected_item_color"`
	DefaultFgColor      string `toml:"default_fg_color"`
	PreviewBgColor      string `toml:"preview_bg_color"`
	HiddenFileColor     string `toml:"hidden_file_color"`
	ExecutableColor     string `toml:"executable_color"`
	SymlinkColor        string `toml:"symlink_color"`
	PreviewBorderColor  string `toml:"preview_border_color"`
}

func loadConfig() Config {
	defaultConfig := Config{
		BorderColor:        "240", // Gray
		StatusBarBgColor:   "235", // Dark gray
		StatusBarFgColor:   "255", // White
		DirColor:          "33",  // Blue
		SelectedItemColor:  "11",  // Yellow
		DefaultFgColor:     "252", // Light gray
		PreviewBgColor:     "234", // Very dark gray
		HiddenFileColor:    "244", // Dark gray
		ExecutableColor:    "46",  // Green
		SymlinkColor:       "14",  // Cyan
		PreviewBorderColor: "240", // Gray
	}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "yazi-go", "config.toml")
	
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
	if config.BorderColor == "" { config.BorderColor = defaultConfig.BorderColor }
	if config.StatusBarBgColor == "" { config.StatusBarBgColor = defaultConfig.StatusBarBgColor }
	if config.StatusBarFgColor == "" { config.StatusBarFgColor = defaultConfig.StatusBarFgColor }
	if config.DirColor == "" { config.DirColor = defaultConfig.DirColor }
	if config.SelectedItemColor == "" { config.SelectedItemColor = defaultConfig.SelectedItemColor }
	if config.DefaultFgColor == "" { config.DefaultFgColor = defaultConfig.DefaultFgColor }
	if config.PreviewBgColor == "" { config.PreviewBgColor = defaultConfig.PreviewBgColor }
	if config.HiddenFileColor == "" { config.HiddenFileColor = defaultConfig.HiddenFileColor }
	if config.ExecutableColor == "" { config.ExecutableColor = defaultConfig.ExecutableColor }
	if config.SymlinkColor == "" { config.SymlinkColor = defaultConfig.SymlinkColor }
	if config.PreviewBorderColor == "" { config.PreviewBorderColor = defaultConfig.PreviewBorderColor }

	return config
}

func getFileInfo(entry fs.DirEntry, dirPath string) FileInfo {
	info := FileInfo{
		Entry:    entry,
		IsHidden: strings.HasPrefix(entry.Name(), "."),
	}
	
	if fileInfo, err := entry.Info(); err == nil {
		info.Size = fileInfo.Size()
		info.ModTime = fileInfo.ModTime()
	}
	
	return info
}

func readDirWithInfo(dirPath string) ([]FileInfo, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	
	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		files = append(files, getFileInfo(entry, dirPath))
	}
	
	return files, nil
}

func (m *model) sortFiles(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		// Directories first
		if files[i].Entry.IsDir() != files[j].Entry.IsDir() {
			return files[i].Entry.IsDir()
		}
		
		var result bool
		switch m.sortBy {
		case "size":
			result = files[i].Size < files[j].Size
		case "modified":
			result = files[i].ModTime.Before(files[j].ModTime)
		default: // name
			result = strings.ToLower(files[i].Entry.Name()) < strings.ToLower(files[j].Entry.Name())
		}
		
		if m.reverseSort {
			return !result
		}
		return result
	})
}

func (m *model) filterFiles(files []FileInfo) []FileInfo {
	if m.showHidden && m.searchQuery == "" {
		return files
	}
	
	filtered := make([]FileInfo, 0, len(files))
	for _, file := range files {
		// Filter hidden files
		if !m.showHidden && file.IsHidden {
			continue
		}
		
		// Filter by search query
		if m.searchQuery != "" {
			if !strings.Contains(strings.ToLower(file.Entry.Name()), strings.ToLower(m.searchQuery)) {
				continue
			}
		}
		
		filtered = append(filtered, file)
	}
	
	return filtered
}

func initialModel() model {
	dir, err := os.Getwd()
	if err != nil {
		return model{err: err}
	}
	
	m := model{
		currentDir:  dir,
		selected:    0,
		config:      loadConfig(),
		sortBy:      "name",
		showHidden:  false,
	}
	
	m.loadCurrentDir()
	return m
}

func (m *model) loadCurrentDir() {
	files, err := readDirWithInfo(m.currentDir)
	if err != nil {
		m.err = err
		return
	}
	
	m.files = m.filterFiles(files)
	m.sortFiles(m.files)
	
	// Load parent directory
	m.parentDir = filepath.Dir(m.currentDir)
	if m.parentDir != m.currentDir {
		parentFiles, err := readDirWithInfo(m.parentDir)
		if err == nil {
			m.parentFiles = m.filterFiles(parentFiles)
			m.sortFiles(m.parentFiles)
			
			// Find current directory in parent list
			currentDirName := filepath.Base(m.currentDir)
			for i, file := range m.parentFiles {
				if file.Entry.Name() == currentDirName {
					m.parentSelected = i
					break
				}
			}
		}
	} else {
		m.parentFiles = nil
	}
	
	// Reset selection if out of bounds
	if m.selected >= len(m.files) {
		m.selected = max(0, len(m.files)-1)
	}
	
	m.updatePreview()
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) updatePreview() {
	if len(m.files) == 0 {
		m.preview = "Empty directory"
		return
	}
	
	selectedFile := m.files[m.selected]
	fullPath := filepath.Join(m.currentDir, selectedFile.Entry.Name())
	
	if selectedFile.Entry.IsDir() {
		subFiles, err := readDirWithInfo(fullPath)
		if err != nil {
			m.preview = fmt.Sprintf("Error: %v", err)
			return
		}
		
		filtered := m.filterFiles(subFiles)
		m.sortFiles(filtered)
		
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(" %s\n", selectedFile.Entry.Name()))
		sb.WriteString(fmt.Sprintf("Items: %d\n\n", len(filtered)))
		
		for i, f := range filtered {
			if i >= 100 { // Show more items in preview
				sb.WriteString("... and more files")
				break
			}
			icon := m.getFileIcon(f)
			sb.WriteString(fmt.Sprintf("%s %s\n", icon, f.Entry.Name()))
		}
		m.preview = sb.String()
	} else {
		// Always try to read as text first
		content, err := os.ReadFile(fullPath)
		if err != nil {
			m.preview = fmt.Sprintf("Error reading file: %v", err)
			return
		}
		
		// Check if it's a text file by extension or content
		fileName := strings.ToLower(selectedFile.Entry.Name())
		ext := strings.ToLower(filepath.Ext(selectedFile.Entry.Name()))
		isText := false
		
		// Check by extension first
		for _, textExt := range textExtensions {
			if ext == textExt || strings.HasSuffix(fileName, strings.ToLower(textExt)) {
				isText = true
				break
			}
		}
		
		// If not recognized by extension, check if it's likely text by content
		if !isText {
			isText = isLikelyTextFile(content)
		}
		
		var sb strings.Builder
		
		// Always show file info header
		icon := m.getFileIcon(selectedFile)
		sb.WriteString(fmt.Sprintf("%s %s\n", icon, selectedFile.Entry.Name()))
		sb.WriteString(fmt.Sprintf("Size: %s\n", formatSize(selectedFile.Size)))
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
		
		m.preview = sb.String()
	}
}

// Helper function to detect if content is likely text
func isLikelyTextFile(content []byte) bool {
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

func formatSize(size int64) string {
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		if m.searchMode {
			switch msg.String() {
			case "enter":
				m.searchMode = false
				m.loadCurrentDir()
				return m, nil
			case "ctrl+c", "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.loadCurrentDir()
				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.loadCurrentDir()
				}
				return m, nil
			default:
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
					m.loadCurrentDir()
				}
				return m, nil
			}
		}
		
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
			
		case "up", "k":
			if m.selected > 0 {
				m.selected--
				if m.selected < m.listOffset {
					m.listOffset = m.selected
				}
				m.updatePreview()
			}
			
		case "down", "j":
			if m.selected < len(m.files)-1 {
				m.selected++
				visibleHeight := m.getVisibleHeight()
				if m.selected >= m.listOffset+visibleHeight {
					m.listOffset = m.selected - visibleHeight + 1
				}
				m.updatePreview()
			}
			
		case "right", "l", "enter":
			if len(m.files) == 0 {
				return m, nil
			}
			selectedFile := m.files[m.selected]
			fullPath := filepath.Join(m.currentDir, selectedFile.Entry.Name())
			if selectedFile.Entry.IsDir() {
				m.currentDir = fullPath
				m.selected = 0
				m.listOffset = 0
				m.previewOffset = 0
				m.loadCurrentDir()
			}
			
		case "left", "h":
			parent := filepath.Dir(m.currentDir)
			if parent != m.currentDir {
				m.currentDir = parent
				m.selected = m.parentSelected
				m.listOffset = max(0, m.selected-m.getVisibleHeight()/2)
				m.previewOffset = 0
				m.loadCurrentDir()
			}
			
		case "o": // Open file in editor
			if len(m.files) == 0 {
				return m, nil
			}
			selectedFile := m.files[m.selected]
			if !selectedFile.Entry.IsDir() {
				fullPath := filepath.Join(m.currentDir, selectedFile.Entry.Name())
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = "nvim"
				}
				cmd := exec.Command(editor, fullPath)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					if err != nil {
						return err
					}
					return nil
				})
			}
			
		case "g": // Go to top
			m.selected = 0
			m.listOffset = 0
			m.updatePreview()
			
		case "G": // Go to bottom
			if len(m.files) > 0 {
				m.selected = len(m.files) - 1
				visibleHeight := m.getVisibleHeight()
				m.listOffset = max(0, len(m.files)-visibleHeight)
				m.updatePreview()
			}
			
		case "~": // Go to home directory
			homeDir, err := os.UserHomeDir()
			if err == nil {
				m.currentDir = homeDir
				m.selected = 0
				m.listOffset = 0
				m.previewOffset = 0
				m.loadCurrentDir()
			}
			
		case "/": // Search mode
			m.searchMode = true
			m.searchQuery = ""
			
		case ".": // Toggle hidden files
			m.showHidden = !m.showHidden
			m.loadCurrentDir()
			
		case "s": // Sort by size
			if m.sortBy == "size" {
				m.reverseSort = !m.reverseSort
			} else {
				m.sortBy = "size"
				m.reverseSort = false
			}
			m.loadCurrentDir()
			
		case "t": // Sort by time
			if m.sortBy == "modified" {
				m.reverseSort = !m.reverseSort
			} else {
				m.sortBy = "modified"
				m.reverseSort = false
			}
			m.loadCurrentDir()
			
		case "n": // Sort by name
			if m.sortBy == "name" {
				m.reverseSort = !m.reverseSort
			} else {
				m.sortBy = "name"
				m.reverseSort = false
			}
			m.loadCurrentDir()
			
		case "r": // Refresh
			m.loadCurrentDir()
			
		case "ctrl+u": // Page up
			visibleHeight := m.getVisibleHeight()
			m.selected = max(0, m.selected-visibleHeight/2)
			m.listOffset = max(0, m.listOffset-visibleHeight/2)
			m.updatePreview()
			
		case "ctrl+d": // Page down
			visibleHeight := m.getVisibleHeight()
			m.selected = min(len(m.files)-1, m.selected+visibleHeight/2)
			if m.selected >= m.listOffset+visibleHeight {
				m.listOffset = m.selected - visibleHeight + 1
			}
			m.updatePreview()
		}
	}
	return m, nil
}

func (m model) getVisibleHeight() int {
	return max(1, m.height-4) // Account for borders and status bar
}

func (m model) getFileIcon(file FileInfo) string {
	name := strings.ToLower(file.Entry.Name())
	ext := strings.ToLower(filepath.Ext(file.Entry.Name()))
	
	if file.Entry.IsDir() {
		// Special directory icons
		switch filepath.Base(name) {
		case ".git":
			return ""
		case ".config":
			return ""
		case "node_modules":
			return ""
		case "downloads":
			return "󰉍"
		case "documents":
			return "󰲂"
		case "pictures", "images":
			return ""
		case "music", "audio":
			return "󱍙"
		case "videos", "movies":
			return ""
		case "desktop":
			return ""
		case "home":
			return "󱂵"
		default:
			return ""
		}
	}
	
	// Special file names
	switch name {
	case "dockerfile":
		return ""
	case "docker-compose.yml", "docker-compose.yaml":
		return ""
	case "makefile":
		return ""
	case "cmakelists.txt":
		return ""
	case "readme.md", "readme.txt", "readme":
		return ""
	case "license", "licence":
		return ""
	case ".gitignore":
		return ""
	case ".gitconfig":
		return ""
	case "package.json":
		return ""
	case "cargo.toml":
		return ""
	case "go.mod":
		return ""
	case "requirements.txt":
		return ""
	}
	
	// File extensions
	switch ext {
	// Programming languages
	case ".go":
		return ""
	case ".py":
		return ""
	case ".js", ".mjs":
		return ""
	case ".ts":
		return ""
	case ".jsx":
		return ""
	case ".tsx":
		return ""
	case ".html", ".htm":
		return ""
	case ".css":
		return ""
	case ".scss", ".sass":
		return ""
	case ".less":
		return ""
	case ".vue":
		return "﵂"
	case ".php":
		return ""
	case ".rb":
		return ""
	case ".java":
		return ""
	case ".c":
		return ""
	case ".cpp", ".cc", ".cxx":
		return ""
	case ".h", ".hpp":
		return ""
	case ".cs":
		return ""
	case ".rs":
		return ""
	case ".swift":
		return ""
	case ".kt":
		return ""
	case ".scala":
		return ""
	case ".clj", ".cljs":
		return ""
	case ".hs":
		return ""
	case ".elm":
		return ""
	case ".lua":
		return ""
	case ".r":
		return "ﳒ"
	case ".sql":
		return ""
	case ".sh", ".bash", ".zsh", ".fish":
		return ""
	case ".ps1":
		return ""
	case ".bat", ".cmd":
		return ""
	
	// Markup and data
	case ".md", ".markdown":
		return ""
	case ".json":
		return ""
	case ".yaml", ".yml":
		return ""
	case ".toml":
		return ""
	case ".xml":
		return ""
	case ".csv":
		return ""
	case ".ini", ".cfg", ".conf":
		return ""
	case ".env":
		return ""
	
	// Images
	case ".jpg", ".jpeg":
		return ""
	case ".png":
		return ""
	case ".gif":
		return ""
	case ".svg":
		return "ﰟ"
	case ".ico":
		return ""
	case ".bmp":
		return ""
	case ".webp":
		return ""
	case ".tiff", ".tif":
		return ""
	
	// Audio
	case ".mp3":
		return ""
	case ".wav":
		return ""
	case ".flac":
		return ""
	case ".ogg":
		return ""
	case ".m4a":
		return ""
	case ".aac":
		return ""
	
	// Video
	case ".mp4":
		return ""
	case ".avi":
		return ""
	case ".mkv":
		return ""
	case ".mov":
		return ""
	case ".wmv":
		return ""
	case ".flv":
		return ""
	case ".webm":
		return ""
	
	// Archives
	case ".zip":
		return ""
	case ".tar", ".tgz", ".tar.gz":
		return ""
	case ".gz":
		return ""
	case ".rar":
		return ""
	case ".7z":
		return ""
	// Documents
	case ".pdf":
		return ""
	case ".doc", ".docx":
		return ""
	case ".xls", ".xlsx":
		return ""
	case ".ppt", ".pptx":
		return ""
	case ".odt":
		return ""
	case ".ods":
		return ""
	case ".odp":
		return ""
	case ".rtf":
		return ""
	
	// Fonts
	case ".ttf", ".otf", ".woff", ".woff2":
		return ""
	
	// Others
	case ".txt":
		return ""
	case ".log":
		return ""
	case ".lock":
		return ""
	case ".tmp":
		return ""
	case ".bak":
		return ""
	case ".iso":
		return ""
	case ".dmg":
		return ""
	case ".exe":
		return ""
	case ".msi":
		return ""
	case ".app":
		return ""
	case ".deb":
		return ""
	case ".rpm":
		return ""
	
	default:
		return ""
	}
}

func (m model) getFileStyle(file FileInfo, isSelected bool) lipgloss.Style {
	var color string
	
	if file.IsHidden {
		color = m.config.HiddenFileColor
	} else if file.Entry.IsDir() {
		color = m.config.DirColor
	} else {
		// Check if executable
		if info, err := file.Entry.Info(); err == nil && info.Mode()&0111 != 0 {
			color = m.config.ExecutableColor
		} else {
			color = m.config.DefaultFgColor
		}
	}
	
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	
	if isSelected {
		style = style.Foreground(lipgloss.Color(m.config.SelectedItemColor)).Bold(true)
	}
	
	return style
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'q' to quit.", m.err)
	}
	
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}
	
	// Calculate pane widths
	parentWidth := max(m.width/4, 15)
	currentWidth := max(m.width/3, 20)
	previewWidth := max(m.width-parentWidth-currentWidth-4, 20)
	
	visibleHeight := m.getVisibleHeight()
	
	// Parent directory pane
	var parentContent strings.Builder
	if m.parentFiles != nil && len(m.parentFiles) > 0 {
		parentContent.WriteString(fmt.Sprintf(" %s\n", filepath.Base(m.parentDir)))
		parentContent.WriteString(strings.Repeat("─", parentWidth-2) + "\n")
		
		for i, file := range m.parentFiles {
			if i >= visibleHeight-2 {
				break
			}
			
			icon := m.getFileIcon(file)
			name := file.Entry.Name()
			if len(name) > parentWidth-6 {
				name = name[:parentWidth-9] + "..."
			}
			
			style := m.getFileStyle(file, i == m.parentSelected)
			line := fmt.Sprintf("%s %s", icon, name)
			parentContent.WriteString(style.Render(line) + "\n")
		}
	}
	
	// Current directory pane
	var currentContent strings.Builder
	currentContent.WriteString(fmt.Sprintf(" %s (%d items)\n", filepath.Base(m.currentDir), len(m.files)))
	currentContent.WriteString(strings.Repeat("─", currentWidth-2) + "\n")
	
	if len(m.files) == 0 {
		currentContent.WriteString("Empty directory")
	} else {
		start := m.listOffset
		end := min(start+visibleHeight-2, len(m.files))
		
		for i := start; i < end; i++ {
			file := m.files[i]
			icon := m.getFileIcon(file)
			name := file.Entry.Name()
			
			// Add size info for files
			sizeInfo := ""
			if !file.Entry.IsDir() {
				sizeInfo = fmt.Sprintf(" (%s)", formatSize(file.Size))
			}
			
			fullName := name + sizeInfo
			if len(fullName) > currentWidth-6 {
				name = name[:max(1, currentWidth-9-len(sizeInfo))] + "..."
				fullName = name + sizeInfo
			}
			
			style := m.getFileStyle(file, i == m.selected)
			line := fmt.Sprintf("%s %s", icon, fullName)
			currentContent.WriteString(style.Render(line) + "\n")
		}
	}
	
	// Preview pane
	var previewContent strings.Builder
	if m.preview != "" {
		lines := strings.Split(m.preview, "\n")
		start := m.previewOffset
		end := min(start+visibleHeight, len(lines))
		
		for i := start; i < end; i++ {
			line := lines[i]
			if len(line) > previewWidth-2 {
				line = line[:previewWidth-5] + "..."
			}
			previewContent.WriteString(line + "\n")
		}
	}
	
	// Styles
	borderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(m.config.BorderColor))
	previewBorderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(m.config.PreviewBorderColor)).Background(lipgloss.Color(m.config.PreviewBgColor))
	
	parentPane := borderStyle.Width(parentWidth).Height(visibleHeight).Render(parentContent.String())
	currentPane := borderStyle.Width(currentWidth).Height(visibleHeight).Render(currentContent.String())
	previewPane := previewBorderStyle.Width(previewWidth).Height(visibleHeight).Render(previewContent.String())
	
	// Join panes horizontally
	panes := lipgloss.JoinHorizontal(lipgloss.Top, parentPane, currentPane, previewPane)
	
	// Status bar
	var statusText string
	if m.searchMode {
		statusText = fmt.Sprintf("Search: %s", m.searchQuery)
	} else {
		sortIndicator := "↑"
		if m.reverseSort {
			sortIndicator = "↓"
		}
		
		statusParts := []string{
			fmt.Sprintf("Dir: %s", m.currentDir),
			fmt.Sprintf("Sort: %s%s", m.sortBy, sortIndicator),
		}
		
		if m.showHidden {
			statusParts = append(statusParts, "Hidden: ON")
		}
		
		if len(m.files) > 0 {
			statusParts = append(statusParts, fmt.Sprintf("%d/%d", m.selected+1, len(m.files)))
		}
		
		statusText = strings.Join(statusParts, " | ")
	}
	
	statusStyle := lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color(m.config.StatusBarBgColor)).
		Foreground(lipgloss.Color(m.config.StatusBarFgColor)).
		Padding(0, 1)
	
	status := statusStyle.Render(statusText)
	
	// Help bar
	helpText := "q:quit | h/l:nav | j/k:up/down | o:open | .:hidden | s:size | t:time | n:name | /:search | r:refresh"
	if m.searchMode {
		helpText = "Type to search | Enter:confirm | Esc:cancel"
	}
	
	helpStyle := lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("248")).
		Padding(0, 1)
	
	help := helpStyle.Render(helpText)
	
	// Full view
	return lipgloss.JoinVertical(lipgloss.Left, panes, status, help)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
