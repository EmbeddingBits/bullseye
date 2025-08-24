package ui

import (
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/embeddingbits/file_viewer/internal/config"
	"github.com/embeddingbits/file_viewer/internal/fileutils"
	"github.com/embeddingbits/file_viewer/pkg/models"
)

// AppModel represents the main application model
type AppModel struct {
	*models.Model
	config config.Config
}

// NewAppModel creates a new application model
func NewAppModel() *AppModel {
	dir, err := os.Getwd()
	if err != nil {
		return &AppModel{
			Model: &models.Model{Err: err},
		}
	}

	cfg := config.LoadConfig()

	m := &AppModel{
		Model: &models.Model{
			CurrentDir: dir,
			Selected:   0,
			SortBy:     "name",
			ShowHidden: false,
		},
		config: cfg,
	}

	m.loadCurrentDir()
	return m
}

// Init initializes the model
func (m *AppModel) Init() tea.Cmd {
	return nil
}

// loadCurrentDir loads the current directory contents
func (m *AppModel) loadCurrentDir() {
	files, err := fileutils.ReadDirWithInfo(m.CurrentDir)
	if err != nil {
		m.Err = err
		return
	}

	m.Files = fileutils.FilterFiles(files, m.ShowHidden, m.SearchQuery)
	fileutils.SortFiles(m.Files, m.SortBy, m.ReverseSort)

	// Load parent directory
	m.ParentDir = filepath.Dir(m.CurrentDir)
	if m.ParentDir != m.CurrentDir {
		parentFiles, err := fileutils.ReadDirWithInfo(m.ParentDir)
		if err == nil {
			m.ParentFiles = fileutils.FilterFiles(parentFiles, m.ShowHidden, m.SearchQuery)
			fileutils.SortFiles(m.ParentFiles, m.SortBy, m.ReverseSort)

			// Find current directory in parent list
			currentDirName := filepath.Base(m.CurrentDir)
			for i, file := range m.ParentFiles {
				if file.Entry.Name() == currentDirName {
					m.ParentSelected = i
					break
				}
			}
		}
	} else {
		m.ParentFiles = nil
	}

	// Reset selection if out of bounds
	if m.Selected >= len(m.Files) {
		m.Selected = max(0, len(m.Files)-1)
	}

	UpdatePreview(m.Model)
}

// Update handles model updates
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.SearchMode {
			return m.handleSearchMode(msg)
		}

		return m.handleNormalMode(msg)
	}
	return m, nil
}

// handleSearchMode handles key events when in search mode
func (m *AppModel) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.SearchMode = false
		m.loadCurrentDir()
		return m, nil
	case "ctrl+c", "esc":
		m.SearchMode = false
		m.SearchQuery = ""
		m.loadCurrentDir()
		return m, nil
	case "backspace":
		if len(m.SearchQuery) > 0 {
			m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
			m.loadCurrentDir()
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			m.SearchQuery += msg.String()
			m.loadCurrentDir()
		}
		return m, nil
	}
}

// handleNormalMode handles key events when in normal mode
func (m *AppModel) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if m.Selected > 0 {
			m.Selected--
			if m.Selected < m.ListOffset {
				m.ListOffset = m.Selected
			}
			UpdatePreview(m.Model)
		}

	case "down", "j":
		if m.Selected < len(m.Files)-1 {
			m.Selected++
			visibleHeight := m.getVisibleHeight()
			if m.Selected >= m.ListOffset+visibleHeight {
				m.ListOffset = m.Selected - visibleHeight + 1
			}
			UpdatePreview(m.Model)
		}

	case "right", "l", "enter":
		if len(m.Files) == 0 {
			return m, nil
		}
		selectedFile := m.Files[m.Selected]
		fullPath := filepath.Join(m.CurrentDir, selectedFile.Entry.Name())
		if selectedFile.Entry.IsDir() {
			m.CurrentDir = fullPath
			m.Selected = 0
			m.ListOffset = 0
			m.PreviewOffset = 0
			m.loadCurrentDir()
		}

	case "left", "h":
		parent := filepath.Dir(m.CurrentDir)
		if parent != m.CurrentDir {
			m.CurrentDir = parent
			m.Selected = m.ParentSelected
			m.ListOffset = max(0, m.Selected-m.getVisibleHeight()/2)
			m.PreviewOffset = 0
			m.loadCurrentDir()
		}

	case "o": // Open file in editor
		if len(m.Files) == 0 {
			return m, nil
		}
		selectedFile := m.Files[m.Selected]
		if !selectedFile.Entry.IsDir() {
			fullPath := filepath.Join(m.CurrentDir, selectedFile.Entry.Name())
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
		m.Selected = 0
		m.ListOffset = 0
		UpdatePreview(m.Model)

	case "G": // Go to bottom
		if len(m.Files) > 0 {
			m.Selected = len(m.Files) - 1
			visibleHeight := m.getVisibleHeight()
			m.ListOffset = max(0, len(m.Files)-visibleHeight)
			UpdatePreview(m.Model)
		}

	case "~": // Go to home directory
		homeDir, err := os.UserHomeDir()
		if err == nil {
			m.CurrentDir = homeDir
			m.Selected = 0
			m.ListOffset = 0
			m.PreviewOffset = 0
			m.loadCurrentDir()
		}

	case "/": // Search mode
		m.SearchMode = true
		m.SearchQuery = ""

	case ".": // Toggle hidden files
		m.ShowHidden = !m.ShowHidden
		m.loadCurrentDir()

	case "s": // Sort by size
		if m.SortBy == "size" {
			m.ReverseSort = !m.ReverseSort
		} else {
			m.SortBy = "size"
			m.ReverseSort = false
		}
		m.loadCurrentDir()

	case "t": // Sort by time
		if m.SortBy == "modified" {
			m.ReverseSort = !m.ReverseSort
		} else {
			m.SortBy = "modified"
			m.ReverseSort = false
		}
		m.loadCurrentDir()

	case "n": // Sort by name
		if m.SortBy == "name" {
			m.ReverseSort = !m.ReverseSort
		} else {
			m.SortBy = "name"
			m.ReverseSort = false
		}
		m.loadCurrentDir()

	case "r": // Refresh
		m.loadCurrentDir()

	case "ctrl+u": // Page up
		visibleHeight := m.getVisibleHeight()
		m.Selected = max(0, m.Selected-visibleHeight/2)
		m.ListOffset = max(0, m.ListOffset-visibleHeight/2)
		UpdatePreview(m.Model)

	case "ctrl+d": // Page down
		visibleHeight := m.getVisibleHeight()
		m.Selected = min(len(m.Files)-1, m.Selected+visibleHeight/2)
		if m.Selected >= m.ListOffset+visibleHeight {
			m.ListOffset = m.Selected - visibleHeight + 1
		}
		UpdatePreview(m.Model)
	}
	return m, nil
}

// getVisibleHeight returns the visible height for the file list
func (m *AppModel) getVisibleHeight() int {
	return max(1, m.Height-4) // Account for borders and status bar
}

// View renders the application view
func (m *AppModel) View() string {
	return RenderView(m.Model, m.config)
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
