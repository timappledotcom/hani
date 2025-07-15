# Hani - TUI Markdown Editor

## Getting Started

Run the editor with:
```bash
./hani [filename]
```

If no filename is provided, it will create a new file.

## Interface Overview

Hani has a tabbed interface with two main areas:
- **📝 Editor**: Where you edit your markdown
- **👁 Preview**: Live preview of your rendered markdown

## Key Bindings

### Navigation
- **Tab**: Switch between Editor and Preview tabs
- **Shift+Tab**: Switch between tabs (same as Tab)

### Editor Mode Commands

#### Normal Mode (Default)
- **h/j/k/l**: Move cursor left/down/up/right
- **w**: Move to next word
- **b**: Move to previous word  
- **e**: Move to end of current word
- **0**: Move to beginning of line
- **$**: Move to end of line
- **gg**: Go to top of file
- **G**: Go to bottom of file

#### Editing Commands
- **i**: Enter Insert mode at cursor
- **a**: Enter Insert mode after cursor
- **A**: Enter Insert mode at end of line
- **o**: Open new line below and enter Insert mode
- **O**: Open new line above and enter Insert mode
- **x**: Delete character under cursor
- **dd**: Delete current line

#### Insert Mode
- **Esc**: Return to Normal mode
- **Enter**: Create new line
- **Backspace**: Delete character before cursor
- **Delete**: Delete character at cursor
- Type normally to insert text

### File Operations
- **Ctrl+S**: Save file
- **Ctrl+C** or **Ctrl+Q**: Quit editor

## Status Bar

The bottom status bar shows:
- Current mode (NORMAL or INSERT)
- Filename and modification status
- Cursor position (row, column)

## Live Preview

Switch to the Preview tab to see your markdown rendered in real-time using the glamour renderer. The preview updates automatically as you edit.

## Tips

1. Start in Normal mode - use **i** to begin editing
2. Use **Tab** to quickly switch between editing and preview
3. The cursor is visible as a █ block in the editor
4. Files are automatically saved with **Ctrl+S**
5. The editor supports standard Vim-like navigation

Enjoy writing markdown with Hani! 🎉
