// Package main implements Hani, a terminal-based markdown editor with vim-like bindings
// and live preview capabilities. Built with Go, Bubbletea, and Glamour.
//
// Key Features:
// - Vim-like navigation and editing commands
// - Real-time markdown preview with Glamour
// - Tabbed interface with visual indicators
// - Syntax highlighting for markdown and code blocks
// - File management with save/load operations
// - Responsive design that adapts to terminal size
//
// Usage:
//   hani [filename]     # Start with optional file
//
// Key Bindings:
//   Tab/Shift+Tab - Switch between editor and preview tabs
//   Ctrl+S        - Save file
//   Ctrl+Q        - Quit application
//   i             - Enter insert mode (vim-like)
//   Esc           - Return to normal mode
//   h,j,k,l       - Navigate (left, down, up, right)
//   w,b,e         - Word movements
//   0,$           - Line beginning/end
//   gg,G          - File beginning/end
package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	m := NewModel(filename)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

