package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

// Terminal modes
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
)

// Tab types
type Tab int

const (
	TabEditor Tab = iota
	TabPreview
)

// Position represents cursor/viewport position
type Position struct {
	row, col int
}

// Viewport represents the visible area
type Viewport struct {
	offsetRow, offsetCol int
}

// DIYEditor represents our custom terminal editor
type DIYEditor struct {
	// Content and state
	content  []string
	cursor   Position
	viewport Viewport
	mode     Mode
	activeTab Tab
	saved    bool
	filename string

	// Terminal control
	width     int
	height    int
	oldState  *term.State

	// Preview (using Charm's glamour)
	renderer      *glamour.TermRenderer
	previewOffset int

	// Status
	statusMsg     string
	statusExpiry  time.Time
}

// NewDIYEditor creates a new DIY editor
func NewDIYEditor(filename string) (*DIYEditor, error) {
	// Get terminal size
	width, height, err := getTerminalSize()
	if err != nil {
		return nil, fmt.Errorf("failed to get terminal size: %w", err)
	}

	// Set up terminal raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to set raw mode: %w", err)
	}

	// Load content
	var content []string
	if filename != "" {
		if file, err := os.Open(filename); err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				content = append(content, scanner.Text())
			}
		}
	}

	if len(content) == 0 {
		content = []string{""}
	}

	// Create glamour renderer for preview with syntax highlighting
	// Try different styles for best syntax highlighting
	var renderer *glamour.TermRenderer
	var err error

	// Try auto style first (adapts to terminal)
	renderer, err = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-4), // Leave some margin
		glamour.WithPreserveNewLines(),
	)

	if err != nil {
		// Try dark style as fallback
		renderer, err = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(width-4),
		)
	}

	if err != nil {
		// Try dracula style (known for good syntax highlighting)
		renderer, err = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dracula"),
			glamour.WithWordWrap(width-4),
		)
	}

	if err != nil {
		// Final fallback - basic renderer
		renderer = nil
	}

	editor := &DIYEditor{
		content:   content,
		cursor:    Position{0, 0},
		viewport:  Viewport{0, 0},
		mode:      ModeNormal,
		activeTab: TabEditor,
		saved:     true,
		filename:  filename,
		width:     width,
		height:    height,
		oldState:  oldState,
		renderer:  renderer,
	}

	// Set up signal handling for cleanup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		editor.Cleanup()
		os.Exit(0)
	}()

	return editor, nil
}

// Cleanup restores terminal state
func (e *DIYEditor) Cleanup() {
	if e.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), e.oldState)
	}
	fmt.Print("\033[?25h") // Show cursor
	fmt.Print("\033[2J\033[H") // Clear screen
}

// getTerminalSize gets current terminal dimensions
func getTerminalSize() (int, int, error) {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return 0, 0, errno
	}
	return int(ws.Col), int(ws.Row), nil
}

// Terminal control functions
func (e *DIYEditor) clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func (e *DIYEditor) moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func (e *DIYEditor) clearLine() {
	fmt.Print("\033[K")
}

func (e *DIYEditor) hideCursor() {
	fmt.Print("\033[?25l")
}

func (e *DIYEditor) showCursor() {
	fmt.Print("\033[?25h")
}

// setStatus sets a temporary status message
func (e *DIYEditor) setStatus(msg string) {
	e.statusMsg = msg
	e.statusExpiry = time.Now().Add(3 * time.Second)
}

// Render performs a full screen redraw
func (e *DIYEditor) Render() {
	e.hideCursor()
	e.clearScreen()

	// Draw tab bar
	e.renderTabBar()

	// Draw content based on active tab
	contentHeight := e.height - 3 // tab + status + footer
	if e.activeTab == TabEditor {
		e.renderEditor(contentHeight)
	} else {
		e.renderPreview(contentHeight)
	}

	// Draw status bar
	e.renderStatusBar()

	// Draw footer
	e.renderFooter()

	// Position cursor if on editor tab
	if e.activeTab == TabEditor {
		cursorRow := e.cursor.row - e.viewport.offsetRow + 2 // +2 for tab bar
		cursorCol := e.cursor.col - e.viewport.offsetCol + 1
		if cursorRow > 1 && cursorRow <= contentHeight+1 && cursorCol > 0 {
			e.moveCursor(cursorRow, cursorCol)
			e.showCursor()
		} else {
			// Cursor is off-screen, hide it
			e.hideCursor()
		}
	} else {
		e.hideCursor()
	}
}

// renderTabBar draws the tab bar
func (e *DIYEditor) renderTabBar() {
	e.moveCursor(1, 1)
	e.clearLine()

	editorStyle := " Editor "
	previewStyle := " Preview "

	if e.activeTab == TabEditor {
		editorStyle = "\033[7m Editor \033[0m" // Inverse video
	} else {
		previewStyle = "\033[7m Preview \033[0m" // Inverse video
	}

	fmt.Printf("%s│%s", editorStyle, previewStyle)
}

// renderEditor draws the editor content
func (e *DIYEditor) renderEditor(height int) {
	for i := 0; i < height; i++ {
		row := i + 2 // Start after tab bar
		e.moveCursor(row, 1)
		e.clearLine()

		lineNum := e.viewport.offsetRow + i
		if lineNum >= len(e.content) {
			fmt.Print("\033[34m~\033[0m") // Blue tilde like vim
			continue
		}

		line := e.content[lineNum]

		// Handle horizontal scrolling
		visibleLine := line
		if e.viewport.offsetCol > 0 {
			if e.viewport.offsetCol < len(line) {
				visibleLine = line[e.viewport.offsetCol:]
			} else {
				visibleLine = ""
			}
		}

		// Truncate if too long
		if len(visibleLine) > e.width {
			visibleLine = visibleLine[:e.width]
		}

		fmt.Print(visibleLine)
	}
}

// renderPreview draws the markdown preview using Charm's glamour
func (e *DIYEditor) renderPreview(height int) {
	if e.renderer == nil {
		e.moveCursor(2, 1)
		fmt.Print("Preview not available (glamour renderer failed)")
		return
	}

	markdown := strings.Join(e.content, "\n")
	if strings.TrimSpace(markdown) == "" {
		e.moveCursor(2, 1)
		fmt.Print("No content to preview")
		return
	}

	// Render markdown using glamour
	rendered, err := e.renderer.Render(markdown)
	if err != nil {
		e.moveCursor(2, 1)
		fmt.Printf("Error rendering markdown: %s", err.Error())
		return
	}

	// Split into lines and apply scrolling
	lines := strings.Split(rendered, "\n")

	// Calculate safe offset
	offset := e.previewOffset
	if offset < 0 {
		offset = 0
	}
	maxOffset := max(0, len(lines)-height)
	if offset > maxOffset {
		offset = maxOffset
		e.previewOffset = offset
	}

	// Draw visible lines
	startLine := offset
	endLine := min(startLine+height, len(lines))

	for i := 0; i < height; i++ {
		row := i + 2 // Start after tab bar
		e.moveCursor(row, 1)
		e.clearLine()

		lineIdx := startLine + i
		if lineIdx < len(lines) && lineIdx < endLine {
			line := lines[lineIdx]
			if len(line) > e.width {
				line = line[:e.width]
			}
			fmt.Print(line)
		}
	}
}

// renderStatusBar draws the status bar
func (e *DIYEditor) renderStatusBar() {
	row := e.height - 1
	e.moveCursor(row, 1)
	e.clearLine()

	// Check if status message has expired
	if time.Now().After(e.statusExpiry) {
		e.statusMsg = ""
	}

	if e.statusMsg != "" {
		fmt.Printf("\033[7m %s \033[0m", e.statusMsg)
	} else {
		modeStr := "NORMAL"
		if e.mode == ModeInsert {
			modeStr = "INSERT"
		}

		saveStatus := ""
		if !e.saved {
			saveStatus = " [+]"
		}

		fmt.Printf("\033[7m %s   %s%s \033[0m", modeStr, e.filename, saveStatus)

		if e.activeTab == TabEditor {
			fmt.Printf("\033[7m (%d,%d) \033[0m", e.cursor.row+1, e.cursor.col+1)
		}
	}
}

// renderFooter draws the footer with key bindings
func (e *DIYEditor) renderFooter() {
	row := e.height
	e.moveCursor(row, 1)
	e.clearLine()

	if e.activeTab == TabEditor {
		if e.mode == ModeInsert {
			fmt.Print(" Ctrl+V Paste │ Esc Normal │ Tab Preview │ Ctrl+S Save │ Ctrl+Q Quit")
		} else {
			fmt.Print(" i Insert │ Tab Preview │ Ctrl+S Save │ o New Line │ d Delete Line │ Ctrl+Q Quit")
		}
	} else {
		fmt.Print(" j/k Scroll │ Tab Editor │ g Top │ G Bottom │ Ctrl+Q Quit")
	}
}

// Main editor loop
func (e *DIYEditor) Run() error {
	defer e.Cleanup()

	e.Render()

	// Use a buffer to handle escape sequences properly
	buffer := make([]byte, 4) // Increased buffer size
	for {
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			return err
		}

		if n > 0 {
			// Handle escape sequences (like Delete key and arrow keys)
			if buffer[0] == 27 { // ESC sequence start
				if n >= 3 && buffer[1] == '[' { // CSI sequence [\033[
					switch buffer[2] {
					case '3': // Delete key (\033[3~) - need to read the ~ if not already read
						if n == 3 {
							// Read the trailing ~
							extraBuf := make([]byte, 1)
							os.Stdin.Read(extraBuf)
						}
						if e.activeTab == TabEditor && e.mode == ModeInsert {
							e.handleDeleteKey()
						}
						e.Render()
						continue
					case 'A': // Up arrow (\033[A)
						if e.activeTab == TabEditor {
							if e.mode == ModeInsert {
								e.handleArrowKey('k')
							} else {
								e.handleNormalKey('k')
							}
						}
						e.Render()
						continue
					case 'B': // Down arrow (\033[B)
						if e.activeTab == TabEditor {
							if e.mode == ModeInsert {
								e.handleArrowKey('j')
							} else {
								e.handleNormalKey('j')
							}
						}
						e.Render()
						continue
					case 'C': // Right arrow (\033[C)
						if e.activeTab == TabEditor {
							if e.mode == ModeInsert {
								e.handleArrowKey('l')
							} else {
								e.handleNormalKey('l')
							}
						}
						e.Render()
						continue
					case 'D': // Left arrow (\033[D)
						if e.activeTab == TabEditor {
							if e.mode == ModeInsert {
								e.handleArrowKey('h')
							} else {
								e.handleNormalKey('h')
							}
						}
						e.Render()
						continue
					}
				} else if n == 1 {
					// Just ESC key - switch to normal mode
					if e.activeTab == TabEditor && e.mode == ModeInsert {
						e.mode = ModeNormal
						if e.cursor.col > 0 {
							e.cursor.col--
						}
						e.adjustViewport()
					}
					e.Render()
					continue
				}
			}

			if e.handleKey(buffer[0]) {
				return nil // Exit requested
			}
			e.Render()
		}
	}
}

// handleKey processes a single key press
func (e *DIYEditor) handleKey(key byte) bool {
	// Global keys
	switch key {
	case 17: // Ctrl+Q
		return true
	case 19: // Ctrl+S
		e.saveFile()
		return false
	case 9: // Tab
		if e.activeTab == TabEditor {
			e.activeTab = TabPreview
		} else {
			e.activeTab = TabEditor
		}
		return false
	}

	if e.activeTab == TabEditor {
		return e.handleEditorKey(key)
	} else {
		return e.handlePreviewKey(key)
	}
}

// handleEditorKey handles keys in editor mode
func (e *DIYEditor) handleEditorKey(key byte) bool {
	if e.mode == ModeNormal {
		return e.handleNormalKey(key)
	} else {
		return e.handleInsertKey(key)
	}
}

// handleNormalKey handles keys in normal mode
func (e *DIYEditor) handleNormalKey(key byte) bool {
	switch key {
	case 'h': // Left
		if e.cursor.col > 0 {
			e.cursor.col--
		}
		e.adjustViewport()
	case 'j': // Down
		if e.cursor.row < len(e.content)-1 {
			e.cursor.row++
			if e.cursor.col > len(e.content[e.cursor.row]) {
				e.cursor.col = len(e.content[e.cursor.row])
			}
		}
		e.adjustViewport()
	case 'k': // Up
		if e.cursor.row > 0 {
			e.cursor.row--
			if e.cursor.col > len(e.content[e.cursor.row]) {
				e.cursor.col = len(e.content[e.cursor.row])
			}
		}
		e.adjustViewport()
	case 'l': // Right
		if e.cursor.col < len(e.content[e.cursor.row]) {
			e.cursor.col++
		}
		e.adjustViewport()
	case '0': // Beginning of line
		e.cursor.col = 0
		e.adjustViewport()
	case '$': // End of line
		e.cursor.col = len(e.content[e.cursor.row])
		e.adjustViewport()
	case 'g': // Handle gg (go to top) - need to read next key
		// For simplicity, just go to top on single 'g'
		e.cursor.row = 0
		e.cursor.col = 0
		e.adjustViewport()
	case 'G': // Go to bottom
		e.cursor.row = len(e.content) - 1
		e.cursor.col = len(e.content[e.cursor.row])
		e.adjustViewport()
	case 'i': // Insert mode
		e.mode = ModeInsert
	case 'a': // Append
		e.mode = ModeInsert
		if e.cursor.col < len(e.content[e.cursor.row]) {
			e.cursor.col++
		}
	case 'A': // Append at end of line
		e.mode = ModeInsert
		e.cursor.col = len(e.content[e.cursor.row])
	case 'o': // Open line below
		e.mode = ModeInsert
		newLine := ""
		e.content = append(e.content[:e.cursor.row+1], append([]string{newLine}, e.content[e.cursor.row+1:]...)...)
		e.cursor.row++
		e.cursor.col = 0
		e.saved = false
		e.adjustViewport()
	case 'O': // Open line above
		e.mode = ModeInsert
		newLine := ""
		e.content = append(e.content[:e.cursor.row], append([]string{newLine}, e.content[e.cursor.row:]...)...)
		e.cursor.col = 0
		e.saved = false
		e.adjustViewport()
	case 'x': // Delete character under cursor
		if e.cursor.col < len(e.content[e.cursor.row]) {
			line := e.content[e.cursor.row]
			e.content[e.cursor.row] = line[:e.cursor.col] + line[e.cursor.col+1:]
			e.saved = false
		} else if e.cursor.row < len(e.content)-1 {
			// At end of line, join with next line
			currentLine := e.content[e.cursor.row]
			nextLine := e.content[e.cursor.row+1]
			e.content[e.cursor.row] = currentLine + nextLine
			e.content = append(e.content[:e.cursor.row+1], e.content[e.cursor.row+2:]...)
			e.saved = false
		}
	case 'd': // Handle dd (delete line) - simplified for now
		// For full vim compatibility, this would need proper command parsing
		// For now, let's implement dd as a single 'd'
		e.deleteLine()
	case 'w': // Next word
		e.cursor = e.nextWord()
		e.adjustViewport()
	case 'b': // Previous word
		e.cursor = e.prevWord()
		e.adjustViewport()
	case 'e': // End of word
		e.cursor = e.endOfWord()
		e.adjustViewport()
	}
	return false
}

// handleInsertKey handles keys in insert mode
func (e *DIYEditor) handleInsertKey(key byte) bool {
	switch key {
	case 27: // Escape
		e.mode = ModeNormal
		if e.cursor.col > 0 {
			e.cursor.col--
		}
	case 22: // Ctrl+V - Paste
		e.pasteFromClipboard()
	case 127, 8: // Backspace
		if e.cursor.col > 0 {
			line := e.content[e.cursor.row]
			e.content[e.cursor.row] = line[:e.cursor.col-1] + line[e.cursor.col:]
			e.cursor.col--
			e.saved = false
		} else if e.cursor.row > 0 {
			// Join with previous line
			prevLine := e.content[e.cursor.row-1]
			currentLine := e.content[e.cursor.row]
			e.content[e.cursor.row-1] = prevLine + currentLine
			e.content = append(e.content[:e.cursor.row], e.content[e.cursor.row+1:]...)
			e.cursor.row--
			e.cursor.col = len(prevLine)
			e.saved = false
		}
		e.adjustViewport()
	case 13: // Enter
		currentLine := e.content[e.cursor.row]
		beforeCursor := currentLine[:e.cursor.col]
		afterCursor := currentLine[e.cursor.col:]

		e.content[e.cursor.row] = beforeCursor
		e.content = append(e.content[:e.cursor.row+1], append([]string{afterCursor}, e.content[e.cursor.row+1:]...)...)

		e.cursor.row++
		e.cursor.col = 0
		e.saved = false
		e.adjustViewport()
	default:
		// Handle arrow keys and other special keys
		if key == 27 { // Escape sequence start - would need more complex parsing for full arrow key support
			return false
		}

		// Regular character input
		if key >= 32 && key <= 126 { // Printable ASCII
			line := e.content[e.cursor.row]
			char := string(rune(key))
			e.content[e.cursor.row] = line[:e.cursor.col] + char + line[e.cursor.col:]
			e.cursor.col++
			e.saved = false
		}
	}
	return false
}

// handlePreviewKey handles keys in preview mode
func (e *DIYEditor) handlePreviewKey(key byte) bool {
	switch key {
	case 'j': // Scroll down
		markdown := strings.Join(e.content, "\n")
		if strings.TrimSpace(markdown) != "" && e.renderer != nil {
			if rendered, err := e.renderer.Render(markdown); err == nil {
				lines := strings.Split(rendered, "\n")
				contentHeight := e.height - 3
				maxOffset := max(0, len(lines)-contentHeight)
				if e.previewOffset < maxOffset {
					e.previewOffset++
				}
			}
		}
	case 'k': // Scroll up
		if e.previewOffset > 0 {
			e.previewOffset--
		}
	case 'g': // Go to top
		e.previewOffset = 0
	case 'G': // Go to bottom
		markdown := strings.Join(e.content, "\n")
		if strings.TrimSpace(markdown) != "" && e.renderer != nil {
			if rendered, err := e.renderer.Render(markdown); err == nil {
				lines := strings.Split(rendered, "\n")
				contentHeight := e.height - 3
				e.previewOffset = max(0, len(lines)-contentHeight)
			}
		}
	}
	return false
}

// adjustViewport ensures cursor is visible
func (e *DIYEditor) adjustViewport() {
	// Ensure we have content
	if len(e.content) == 0 {
		e.content = []string{""}
		e.cursor.row = 0
		e.cursor.col = 0
		return
	}

	// Bounds checking - be more lenient
	if e.cursor.row < 0 {
		e.cursor.row = 0
	}
	if e.cursor.row >= len(e.content) {
		e.cursor.row = len(e.content) - 1
	}

	// Column bounds checking
	if e.cursor.row < len(e.content) && e.cursor.row >= 0 {
		lineLen := len(e.content[e.cursor.row])
		if e.cursor.col < 0 {
			e.cursor.col = 0
		} else if e.cursor.col > lineLen {
			e.cursor.col = lineLen
		}
	}

	contentHeight := e.height - 3
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Vertical scrolling - ensure cursor row is visible
	if e.cursor.row < e.viewport.offsetRow {
		e.viewport.offsetRow = e.cursor.row
	} else if e.cursor.row >= e.viewport.offsetRow+contentHeight {
		e.viewport.offsetRow = e.cursor.row - contentHeight + 1
		if e.viewport.offsetRow < 0 {
			e.viewport.offsetRow = 0
		}
	}

	// Horizontal scrolling
	contentWidth := e.width - 3
	if contentWidth < 1 {
		contentWidth = 1
	}
	if e.cursor.col < e.viewport.offsetCol {
		e.viewport.offsetCol = e.cursor.col
	} else if e.cursor.col >= e.viewport.offsetCol+contentWidth {
		e.viewport.offsetCol = e.cursor.col - contentWidth + 1
		if e.viewport.offsetCol < 0 {
			e.viewport.offsetCol = 0
		}
	}
}

// saveFile saves the current content
func (e *DIYEditor) saveFile() {
	filename := e.filename
	if filename == "" {
		filename = "untitled.md"
		e.filename = filename
	}

	content := strings.Join(e.content, "\n")

	// Create backup
	if _, err := os.Stat(filename); err == nil {
		backupName := filename + ".bak"
		if backupData, err := os.ReadFile(filename); err == nil {
			os.WriteFile(backupName, backupData, 0644)
		}
	}

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		e.setStatus("Error saving file: " + err.Error())
	} else {
		e.saved = true
		e.setStatus("File saved: " + filename)
	}
}

// pasteFromClipboard handles clipboard paste operations efficiently
func (e *DIYEditor) pasteFromClipboard() {
	clipboard := e.getClipboard()
	if clipboard == "" {
		e.setStatus("Clipboard empty")
		return
	}

	// This is the key difference - no framework overhead!
	// Direct content modification with immediate rendering
	lines := strings.Split(clipboard, "\n")

	if len(lines) == 1 {
		// Single line paste
		line := e.content[e.cursor.row]
		e.content[e.cursor.row] = line[:e.cursor.col] + clipboard + line[e.cursor.col:]
		e.cursor.col += len(clipboard)
	} else {
		// Multi-line paste - handle efficiently
		currentLine := e.content[e.cursor.row]
		beforeCursor := currentLine[:e.cursor.col]
		afterCursor := currentLine[e.cursor.col:]

		// Build new content directly
		newContent := make([]string, 0, len(e.content)+len(lines)-1)
		newContent = append(newContent, e.content[:e.cursor.row]...)

		// First line: combine with text before cursor
		newContent = append(newContent, beforeCursor+lines[0])

		// Middle lines (if any)
		if len(lines) > 2 {
			newContent = append(newContent, lines[1:len(lines)-1]...)
		}

		// Last line: combine with text after cursor
		lastLineContent := lines[len(lines)-1] + afterCursor
		newContent = append(newContent, lastLineContent)

		// Add remaining content
		newContent = append(newContent, e.content[e.cursor.row+1:]...)

		e.content = newContent
		e.cursor.row += len(lines) - 1
		e.cursor.col = len(lines[len(lines)-1])

		// Ensure cursor is still valid after paste
		if e.cursor.row >= len(e.content) {
			e.cursor.row = len(e.content) - 1
		}
		if e.cursor.row >= 0 && e.cursor.col > len(e.content[e.cursor.row]) {
			e.cursor.col = len(e.content[e.cursor.row])
		}
	}

	e.saved = false
	e.adjustViewport()

	// Show status for large pastes
	if len(lines) > 10 {
		e.setStatus(fmt.Sprintf("Pasted %d lines", len(lines)))
	} else if len(lines) > 1 {
		e.setStatus(fmt.Sprintf("Pasted %d lines", len(lines)))
	}
}
}

// getClipboard gets clipboard content using various clipboard tools
func (e *DIYEditor) getClipboard() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try wl-paste (Wayland)
	cmd := exec.CommandContext(ctx, "wl-paste")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimRight(string(output), "\n")
	}

	// Try xclip (X11)
	cmd = exec.CommandContext(ctx, "xclip", "-o", "-selection", "clipboard")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimRight(string(output), "\n")
	}

	// Try pbpaste (macOS)
	cmd = exec.CommandContext(ctx, "pbpaste")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimRight(string(output), "\n")
	}

	return ""
}

// deleteLine deletes the current line (dd command)
func (e *DIYEditor) deleteLine() {
	if len(e.content) > 1 {
		e.content = append(e.content[:e.cursor.row], e.content[e.cursor.row+1:]...)
		if e.cursor.row >= len(e.content) {
			e.cursor.row = len(e.content) - 1
		}
		if e.cursor.col > len(e.content[e.cursor.row]) {
			e.cursor.col = len(e.content[e.cursor.row])
		}
		e.saved = false
	} else {
		e.content[0] = ""
		e.cursor.col = 0
		e.saved = false
	}
	e.adjustViewport()
}

// Word movement functions (ported from original)
func (e *DIYEditor) nextWord() Position {
	row := e.cursor.row
	col := e.cursor.col

	// Bounds checking
	if row >= len(e.content) {
		if len(e.content) > 0 {
			return Position{row: len(e.content) - 1, col: len(e.content[len(e.content)-1])}
		}
		return Position{row: 0, col: 0}
	}

	line := e.content[row]

	// Skip current word (non-whitespace characters)
	for col < len(line) && !e.isWhitespace(line[col]) {
		col++
	}

	// Skip whitespace
	for col < len(line) && e.isWhitespace(line[col]) {
		col++
	}

	// If we're at the end of the line, move to next line
	if col >= len(line) && row < len(e.content)-1 {
		row++
		col = 0
		// Skip leading whitespace on next line
		if row < len(e.content) {
			line = e.content[row]
			for col < len(line) && e.isWhitespace(line[col]) {
				col++
			}
		}
	}

	return Position{row: row, col: col}
}

func (e *DIYEditor) prevWord() Position {
	row := e.cursor.row
	col := e.cursor.col

	// Bounds checking
	if row >= len(e.content) || row < 0 {
		return Position{row: 0, col: 0}
	}

	if col > 0 {
		col--
	} else if row > 0 {
		row--
		if row < len(e.content) {
			col = len(e.content[row])
		}
	}

	if row < 0 {
		return Position{row: 0, col: 0}
	}
	if row >= len(e.content) {
		return Position{row: len(e.content) - 1, col: 0}
	}

	line := e.content[row]

	// Skip whitespace backwards
	for col > 0 && col < len(line) && e.isWhitespace(line[col]) {
		col--
	}

	// Skip word backwards
	for col > 0 && col < len(line) && !e.isWhitespace(line[col]) {
		col--
	}

	// Move to start of word
	if col > 0 && col < len(line) && e.isWhitespace(line[col]) {
		col++
	}

	return Position{row: row, col: col}
}

func (e *DIYEditor) endOfWord() Position {
	row := e.cursor.row
	col := e.cursor.col

	// Bounds checking
	if row >= len(e.content) {
		if len(e.content) > 0 {
			return Position{row: len(e.content) - 1, col: len(e.content[len(e.content)-1])}
		}
		return Position{row: 0, col: 0}
	}

	line := e.content[row]

	// If we're at the end of a word, move to next word first
	if col < len(line) && !e.isWhitespace(line[col]) {
		// Move to end of current word
		for col < len(line) && !e.isWhitespace(line[col]) {
			col++
		}
		if col > 0 {
			col--
		}
		return Position{row: row, col: col}
	}

	// Skip whitespace to find next word
	for col < len(line) && e.isWhitespace(line[col]) {
		col++
	}

	// Move to end of next word
	for col < len(line) && !e.isWhitespace(line[col]) {
		col++
	}

	if col > 0 {
		col--
	}

	return Position{row: row, col: col}
}

// isWhitespace checks if a character is whitespace
func (e *DIYEditor) isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// handleDeleteKey handles the Delete key in insert mode
func (e *DIYEditor) handleDeleteKey() {
	if e.cursor.col < len(e.content[e.cursor.row]) {
		// Delete character at cursor
		line := e.content[e.cursor.row]
		e.content[e.cursor.row] = line[:e.cursor.col] + line[e.cursor.col+1:]
		e.saved = false
	} else if e.cursor.row < len(e.content)-1 {
		// At end of line, join with next line
		currentLine := e.content[e.cursor.row]
		nextLine := e.content[e.cursor.row+1]
		e.content[e.cursor.row] = currentLine + nextLine
		e.content = append(e.content[:e.cursor.row+1], e.content[e.cursor.row+2:]...)
		e.saved = false
	}
	e.adjustViewport()
}

// handleArrowKey handles arrow keys in insert mode
func (e *DIYEditor) handleArrowKey(direction byte) {
	switch direction {
	case 'h': // Left
		if e.cursor.col > 0 {
			e.cursor.col--
		}
	case 'j': // Down
		// Make sure we can move down if there are more lines
		if e.cursor.row < len(e.content)-1 {
			e.cursor.row++
			// Adjust column if the new line is shorter, but allow navigation
			lineLen := len(e.content[e.cursor.row])
			if e.cursor.col > lineLen {
				e.cursor.col = lineLen
			}
		}
	case 'k': // Up
		if e.cursor.row > 0 {
			e.cursor.row--
			// Adjust column if the new line is shorter
			lineLen := len(e.content[e.cursor.row])
			if e.cursor.col > lineLen {
				e.cursor.col = lineLen
			}
		}
	case 'l': // Right
		if e.cursor.row < len(e.content) && e.cursor.col < len(e.content[e.cursor.row]) {
			e.cursor.col++
		}
	}
	e.adjustViewport()
}

// Utility functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	editor, err := NewDIYEditor(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating editor: %v\n", err)
		os.Exit(1)
	}

	if err := editor.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Editor error: %v\n", err)
		os.Exit(1)
	}
}
