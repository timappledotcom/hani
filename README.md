# Hani - A TUI Markdown Editor

A terminal-based markdown editor with vim-like bindings and live preview, built with Go, Bubbletea, and glamour.

## Features

- **Vim-like bindings**: Familiar navigation and editing commands
- **Live preview**: Real-time markdown rendering with glamour (glow)
- **Tabbed interface**: Clearly labeled tabs with visual indicators
- **Beautiful rendering**: Styled markdown preview using glamour
- **Visible cursor**: Clear cursor indication in editor mode
- **File management**: Save and load markdown files
- **Responsive design**: Works in any terminal size

## Installation

```bash
go build -o hani
```

## Usage

```bash
# Create a new file
hani

# Edit an existing file
hani README.md

# Edit a specific markdown file
hani document.md
```

## Key Bindings

### Global Commands
- `Tab` / `Shift+Tab` - Switch between editor and preview tabs
- `Ctrl+S` - Save file
- `Ctrl+Q` - Quit application

### Normal Mode (Vim-like)
- `h`, `j`, `k`, `l` - Move cursor left, down, up, right
- `w` - Move to next word
- `b` - Move to previous word
- `e` - Move to end of current word
- `0` - Move to beginning of line
- `$` - Move to end of line
- `gg` - Go to first line
- `G` - Go to last line
- `i` - Enter insert mode
- `a` - Enter insert mode (after cursor)
- `A` - Enter insert mode (end of line)
- `o` - Insert new line below and enter insert mode
- `O` - Insert new line above and enter insert mode
- `x` - Delete character under cursor
- `dd` - Delete current line

### Insert Mode
- `Esc` - Return to normal mode
- `Enter` - Create new line
- `Backspace` - Delete character before cursor
- `Delete` - Delete character at cursor
- Any printable character - Insert character

### Preview Mode
- `j` / `Down` - Scroll preview down
- `k` / `Up` - Scroll preview up
- `g` - Go to top of preview
- `G` - Go to bottom of preview

## Project Structure

```
hani/
├── main.go      # Application entry point
├── model.go     # Main application model and view logic
├── keys.go      # Key binding and input handling
├── README.md    # This file
└── go.mod       # Go module file
```

## Features Overview

### Editor Tab
- Full vim-like navigation and editing
- Syntax highlighting for markdown
- Line numbers and cursor position
- Real-time file modification tracking

### Preview Tab
- Live rendering of markdown content using glamour
- Beautiful terminal-native styling
- Syntax highlighting for code blocks
- Professional markdown rendering
- GitHub Flavored Markdown (GFM) support
- Responsive layout with borders

### Status Bar
- Current mode indicator (NORMAL/INSERT)
- File name and modification status
- Cursor position
- Temporary status messages

## Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling and layout
- [Glamour](https://github.com/charmbracelet/glamour) - Terminal markdown rendering

## Technical Details

### Architecture
- **Model-View-Update (MVU)**: Built with the Elm architecture pattern
- **Event-driven**: Reactive to user input and system events
- **Stateful**: Maintains editor state, cursor position, and file content
- **Modular**: Clean separation of concerns across files

### Rendering
- **Glamour**: Professional terminal markdown rendering
- **Lipgloss**: Styling and layout management
- **Live preview**: Real-time updates as you type

### Vim Bindings
- **Modal editing**: Separate normal and insert modes
- **Word movement**: w, b, e commands for word navigation
- **Line operations**: dd for line deletion, o/O for line insertion
- **Character operations**: x for character deletion

## Future Enhancements

- [ ] Improved HTML-to-terminal rendering
- [ ] Syntax highlighting in editor
- [ ] Search and replace functionality
- [ ] Multiple file support (tabs)
- [ ] Configuration file support
- [ ] Custom key bindings
- [ ] Export to different formats
- [ ] Plugin system

## Contributing

Feel free to submit issues and pull requests to improve Hani!

## License

MIT License - see LICENSE file for details

---

Made with ❤️ using Go and Bubbletea
