# File Viewer

A terminal-based file manager written in Go with a beautiful TUI interface.

## Features

- **Three-pane layout**: Parent directory, current directory, and file preview
- **File navigation**: Navigate through directories with keyboard shortcuts
- **File preview**: View text files and binary files with hex preview
- **Search functionality**: Search for files by name
- **Sorting options**: Sort by name, size, or modification time
- **Hidden files**: Toggle visibility of hidden files
- **File icons**: Visual indicators for different file types
- **Color themes**: Configurable color scheme via TOML configuration
- **Keyboard shortcuts**: Vim-like navigation and commands

## Project Structure

```
file_viewer/
├── cmd/
│   └── fileviewer/          # Main application entry point
│       └── main.go
├── internal/                 # Internal packages (not importable)
│   ├── config/              # Configuration management
│   │   └── config.go
│   ├── fileutils/           # File system utilities
│   │   └── fileutils.go
│   └── ui/                  # User interface components
│       ├── icons.go         # File icons and visual elements
│       ├── model.go         # Application model and business logic
│       ├── preview.go       # File preview generation
│       ├── styling.go       # UI styling and colors
│       └── view.go          # View rendering
├── pkg/                     # Public packages (importable)
│   └── models/              # Data models
│       └── fileinfo.go
├── config.toml              # Default configuration
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
└── README.md                # This file
```

## Architecture

The application follows a modular architecture with clear separation of concerns:

- **Models**: Data structures and business logic
- **Config**: Configuration management and defaults
- **FileUtils**: File system operations and utilities
- **UI**: User interface components and rendering
- **Main**: Application entry point and program setup

## Features to be added

- [ ] Open files with respective software (audio files - mpd, video files - mpv, code - nano.....)
- [ ] Syntax highlighting for Code/File preview

## Building

```bash
# Build the application
go build -o bullseye ./cmd/bullseye

# Run the application
./bullseye
```

## Configuration

The application uses TOML configuration files. Configuration can be placed in:

1. `~/.config/bullseye/config.toml` (user configuration)
2. `./config.toml` (local configuration)

### Configuration Options

```toml
border_color = "#EBDBB2"
status_bar_bg_color = "#458588"
status_bar_fg_color = "#fbf1c7"
selected_item_color = "#83a598"
dir_color = "#458588"
default_fg_color = "#ebdbb2"
preview_bg_color = "#282828"
hidden_file_color = "#928374"
executable_color = "#b8bb26"
symlink_color = "#83a598"
preview_border_color = "#504945"
hover_bg_color = "#000000"
```

## Keyboard Shortcuts

- **Navigation**:
  - `h` / `left`: Go to parent directory
  - `l` / `right`: Enter directory
  - `j` / `down`: Move down
  - `k` / `up`: Move up
  - `g`: Go to top
  - `G`: Go to bottom
  - `~`: Go to home directory

- **File Operations**:
  - `enter`: Open file in editor
  - `r`: Refresh directory

- **View Options**:
  - `.`: Toggle hidden files
  - `/`: Enter search mode
  - `s`: Sort by size
  - `t`: Sort by time
  - `n`: Sort by name

- **Search Mode**:
  - Type to search
  - `Enter`: Confirm search
  - `Esc` / `Ctrl+C`: Cancel search

- **Other**:
  - `q` / `Ctrl+C`: Quit
  - `Ctrl+U`: Page up
  - `Ctrl+D`: Page down

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea): TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss): Styling library
- [go-toml](https://github.com/pelletier/go-toml): TOML parsing

## Contribution

- **Find an Issue**: Look through our issues page for a bug or feature you'd like to work on. Comment on the issue to let others know you're working on it.
- **Fork the Repository**: Start by forking the repository to your own GitHub account.
- **Create a New Branch**: Before you start coding, create a new branch from main with a descriptive name (e.g., fix-preview-bug or feature-async-copy).
- **Write Your Code**: Add your code, making sure to follow the project's coding style and conventions.
- **Test Your Changes**: We highly recommend writing tests for new features or bug fixes. This ensures that your code works as expected and helps prevent future regressions.
- **Submit a Pull Request**: When you're ready, submit a pull request to the main branch. Provide a clear title and description of your changes.

## License

This project is open source and available under the MIT License.
