package main

import (
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/pelletier/go-toml/v2"
)

var (
    textExtensions = []string{
        ".txt", ".go", ".md", ".json", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf",
        ".sh", ".py", ".js", ".html", ".css", ".c", ".cpp", ".h", ".java",
    }
)

// Config holds color settings for the TUI
type Config struct {
    BorderColor       string `toml:"border_color"`
    StatusBarBgColor  string `toml:"status_bar_bg_color"`
    StatusBarFgColor  string `toml:"status_bar_fg_color"`
    SelectedItemColor string `toml:"selected_item_color"`
}

type model struct {
    currentDir string
    files      []fs.DirEntry
    selected   int
    listOffset int
    preview    string
    width      int
    height     int
    err        error
    config     Config
}

func loadConfig() Config {
    defaultConfig := Config{
        BorderColor:       "63",  // Purple
        StatusBarBgColor:  "62",  // Dark purple
        StatusBarFgColor:  "255", // White
        SelectedItemColor: "208", // Orange
    }

    data, err := os.ReadFile("config.toml")
    if err != nil {
        return defaultConfig
    }

    var config Config
    if err := toml.Unmarshal(data, &config); err != nil {
        return defaultConfig
    }

    // Ensure non-empty values, fallback to defaults if empty
    if config.BorderColor == "" {
        config.BorderColor = defaultConfig.BorderColor
    }
    if config.StatusBarBgColor == "" {
        config.StatusBarBgColor = defaultConfig.StatusBarBgColor
    }
    if config.StatusBarFgColor == "" {
        config.StatusBarFgColor = defaultConfig.StatusBarFgColor
    }
    if config.SelectedItemColor == "" {
        config.SelectedItemColor = defaultConfig.SelectedItemColor
    }

    return config
}

func initialModel() model {
    dir, err := os.Getwd()
    if err != nil {
        return model{err: err}
    }
    files, err := os.ReadDir(dir)
    if err != nil {
        return model{err: err}
    }
    m := model{
        currentDir: dir,
        files:      files,
        selected:   0,
        config:     loadConfig(),
    }
    m.updatePreview()
    return m
}

func (m model) Init() tea.Cmd {
    return tea.WindowSize()
}

func (m *model) updatePreview() {
    if len(m.files) == 0 {
        m.preview = ""
        return
    }
    selectedFile := m.files[m.selected]
    fullPath := filepath.Join(m.currentDir, selectedFile.Name())
    if selectedFile.IsDir() {
        subFiles, err := os.ReadDir(fullPath)
        if err != nil {
            m.preview = "Error listing directory"
            return
        }
        var sb strings.Builder
        for _, f := range subFiles {
            sb.WriteString(f.Name() + "\n")
        }
        m.preview = sb.String()
    } else {
        ext := filepath.Ext(selectedFile.Name())
        isText := false
        for _, textExt := range textExtensions {
            if ext == textExt {
                isText = true
                break
            }
        }
        if !isText {
            m.preview = "Not a text or code file"
            return
        }
        content, err := os.ReadFile(fullPath)
        if err != nil {
            m.preview = "Error reading file"
            return
        }
        m.preview = string(content)
    }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        return m, nil
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            if m.selected > 0 {
                m.selected--
                if m.selected < m.listOffset {
                    m.listOffset--
                }
                m.updatePreview()
            }
        case "down", "j":
            if m.selected < len(m.files)-1 {
                m.selected++
                visibleHeight := m.height - 3 // account for borders/status
                if m.selected >= m.listOffset+visibleHeight {
                    m.listOffset++
                }
                m.updatePreview()
            }
        case "enter", "l":
            selectedFile := m.files[m.selected]
            fullPath := filepath.Join(m.currentDir, selectedFile.Name())
            if selectedFile.IsDir() {
                files, err := os.ReadDir(fullPath)
                if err != nil {
                    m.err = err
                    return m, nil
                }
                m.currentDir = fullPath
                m.files = files
                m.selected = 0
                m.listOffset = 0
                m.updatePreview()
            }
        case "backspace", "h":
            parent := filepath.Dir(m.currentDir)
            if parent != m.currentDir {
                files, err := os.ReadDir(parent)
                if err != nil {
                    m.err = err
                    return m, nil
                }
                m.currentDir = parent
                m.files = files
                m.selected = 0
                m.listOffset = 0
                m.updatePreview()
            }
        }
    }
    return m, nil
}

func (m model) View() string {
    if m.err != nil {
        return fmt.Sprintf("Error: %v", m.err)
    }
    if m.width == 0 || m.height == 0 {
        return "Loading..."
    }

    // Calculate widths
    leftWidth := m.width / 3
    rightWidth := m.width - leftWidth - 2 // account for border

    // Left pane: file list
    var leftSB strings.Builder
    visibleHeight := m.height - 3 // borders and status
    start := m.listOffset
    end := start + visibleHeight
    if end > len(m.files) {
        end = len(m.files)
    }
    for i := start; i < end; i++ {
        cursor := "  "
        name := m.files[i].Name()
        if i == m.selected {
            cursor = "> "
            name = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.SelectedItemColor)).Render(name)
        }
        leftSB.WriteString(cursor + name + "\n")
    }
    leftStr := leftSB.String()

    // Right pane: preview (truncate to height)
    previewLines := strings.Split(m.preview, "\n")
    var rightSB strings.Builder
    for i := 0; i < len(previewLines) && i < visibleHeight; i++ {
        line := previewLines[i]
        if len(line) > rightWidth-2 {
            line = line[:rightWidth-2] // truncate wide lines
        }
        rightSB.WriteString(line + "\n")
    }
    rightStr := rightSB.String()

    // Styles
    border := lipgloss.NormalBorder()
    leftStyle := lipgloss.NewStyle().
        Width(leftWidth).
        Height(visibleHeight).
        Border(border, false, true, false, false).
        BorderForeground(lipgloss.Color(m.config.BorderColor))
    rightStyle := lipgloss.NewStyle().
        Width(rightWidth).
        Height(visibleHeight).
        Border(border, false, false, false, false).
        BorderForeground(lipgloss.Color(m.config.BorderColor))

    // Join panes
    panes := lipgloss.JoinHorizontal(
        lipgloss.Top,
        leftStyle.Render(leftStr),
        rightStyle.Render(rightStr),
    )

    // Status bar
    statusStyle := lipgloss.NewStyle().
        Width(m.width - 2).
        Background(lipgloss.Color(m.config.StatusBarBgColor)).
        Foreground(lipgloss.Color(m.config.StatusBarFgColor)).
        Padding(0, 1)
    status := statusStyle.Render(m.currentDir)

    // Full view
    return lipgloss.JoinVertical(lipgloss.Left, panes, status)
}

func main() {
    p := tea.NewProgram(initialModel(), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}
