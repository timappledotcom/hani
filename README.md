# Hani - A TUI Markdown Editor

A terminal-based markdown editor with vim-like bindings and live preview, built with Go.

## Features

- **Vim-like bindings**: Familiar navigation and editing commands
- **Live preview**: Real-time markdown rendering with glamour (glow)
- **Tabbed interface**: Switch between editor and preview modes
- **Beautiful rendering**: Styled markdown preview with syntax highlighting
- **Fast performance**: Direct terminal control for responsive editing
- **File management**: Save and load markdown files
- **Responsive design**: Works in any terminal size

## Two Implementations

Hani provides two implementations:

### DIY Version (Recommended)
**File**: `diy_hani.go`
- Direct terminal control for maximum performance
- No framework overhead - pure Go with terminal manipulation
- Smooth paste operations for large code blocks
- Excellent responsiveness and stability
- Uses glamour for beautiful preview rendering

### Bubbletea Version (Legacy)
**Files**: `main.go`, `model.go`, `keys.go`
- Built with Charm's Bubbletea TUI framework
- Clean architecture using Model-View-Update pattern
- May experience performance issues with large content
- Maintained for reference and comparison

## Installation

```bash
# Build the recommended DIY version
go build -o hani diy_hani.go

# Or build the Bubbletea version
go build -o hani-bubbletea main.go model.go keys.go config.go highlight.go version.go
```

## Usage

```bash
# Create a new file
./hani

# Edit an existing file  
./hani README.md

# Edit a specific markdown file
./hani document.md
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
├── diy_hani.go    # DIY implementation (recommended)
├── main.go        # Bubbletea application entry point  
├── model.go       # Bubbletea application model and view logic
├── keys.go        # Bubbletea key binding and input handling
├── config.go      # Configuration and constants
├── highlight.go   # Syntax highlighting utilities
├── version.go     # Version information
├── README.md      # This file
├── go.mod         # Go module file
└── Makefile       # Build automation
```

## Performance Comparison

The DIY version offers significant advantages:
- **Faster paste operations**: Handles large code blocks without lag
- **Lower memory usage**: No framework overhead
- **More responsive**: Direct terminal control
- **Better stability**: Fewer moving parts, less complex state management

## Dependencies

- [Glamour](https://github.com/charmbracelet/glamour) - Terminal markdown rendering
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) - Terminal control (DIY version)

### Legacy Dependencies (Bubbletea version)
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling and layout

## Technical Details

### DIY Implementation Architecture
- **Direct terminal control**: Raw terminal I/O for maximum performance
- **Escape sequence handling**: Proper arrow key and special key support
- **Efficient rendering**: Minimal screen updates, optimized cursor movement
- **Pure Go**: No external TUI framework dependencies

### Bubbletea Implementation Architecture  
- **Model-View-Update (MVU)**: Built with the Elm architecture pattern
- **Event-driven**: Reactive to user input and system events
- **Stateful**: Maintains editor state, cursor position, and file content
- **Modular**: Clean separation of concerns across files

### Common Features
- **Glamour rendering**: Professional terminal markdown rendering
- **Live preview**: Real-time updates as you type
- **Vim bindings**: Modal editing with normal and insert modes
- **File operations**: Save, load, backup creation

## Future Enhancements

- [x] High-performance DIY implementation
- [x] Improved paste operations for large content
- [ ] Search and replace functionality  
- [ ] Multiple file support (tabs)
- [ ] Configuration file support
- [ ] Custom key bindings
- [ ] Export to different formats
- [ ] Syntax highlighting in editor mode

## Contributing

Feel free to submit issues and pull requests to improve Hani!

## License

MIT License - see LICENSE file for details

---

Made with ❤️ using Go and Bubbletea
