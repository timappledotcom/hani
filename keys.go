package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "ctrl+s":
		return m.saveFile()

	case "tab":
		if m.activeTab == TabEditor {
			m.activeTab = TabPreview
		} else {
			m.activeTab = TabEditor
		}
		return m, nil

	case "shift+tab":
		if m.activeTab == TabEditor {
			m.activeTab = TabPreview
		} else {
			m.activeTab = TabEditor
		}
		return m, nil
	}

	// Only handle editor keys when on editor tab
	if m.activeTab == TabEditor {
		switch m.mode {
		case ModeNormal:
			return m.handleNormalMode(msg)
		case ModeInsert:
			return m.handleInsertMode(msg)
		}
	} else if m.activeTab == TabPreview {
		// Handle scrolling in preview mode
		return m.handlePreviewMode(msg)
	}

	return m, nil
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ensure cursor is within bounds before any operation
	m.ensureCursorBounds()

	switch msg.String() {
	case "h", "left":
		if m.cursor.col > 0 {
			m.cursor.col--
		}
		m.adjustViewport()
		return m, nil

	case "j", "down":
		if m.cursor.row < len(m.content)-1 {
			m.cursor.row++
			// Adjust column if the new line is shorter
			if m.cursor.col > len(m.content[m.cursor.row]) {
				m.cursor.col = len(m.content[m.cursor.row])
			}
		}
		m.adjustViewport()
		return m, nil

	case "k", "up":
		if m.cursor.row > 0 {
			m.cursor.row--
			// Adjust column if the new line is shorter
			if m.cursor.col > len(m.content[m.cursor.row]) {
				m.cursor.col = len(m.content[m.cursor.row])
			}
		}
		m.adjustViewport()
		return m, nil

	case "l", "right":
		if m.cursor.row < len(m.content) && m.cursor.col < len(m.content[m.cursor.row]) {
			m.cursor.col++
		}
		m.adjustViewport()
		return m, nil

	case "0":
		m.cursor.col = 0
		m.adjustViewport()
		return m, nil

	case "$":
		m.cursor.col = len(m.content[m.cursor.row])
		m.adjustViewport()
		return m, nil

	case "gg":
		m.cursor.row = 0
		m.cursor.col = 0
		m.adjustViewport()
		return m, nil

	case "G":
		m.cursor.row = len(m.content) - 1
		m.cursor.col = len(m.content[m.cursor.row])
		m.adjustViewport()
		return m, nil

	case "i":
		m.mode = ModeInsert
		return m, nil

	case "a":
		m.mode = ModeInsert
		if m.cursor.col < len(m.content[m.cursor.row]) {
			m.cursor.col++
		}
		return m, nil

	case "A":
		m.mode = ModeInsert
		m.cursor.col = len(m.content[m.cursor.row])
		return m, nil

	case "o":
		m.mode = ModeInsert
		// Insert new line after current line
		newLine := ""
		m.content = append(m.content[:m.cursor.row+1], append([]string{newLine}, m.content[m.cursor.row+1:]...)...)
		m.cursor.row++
		m.cursor.col = 0
		m.saved = false
		m.codeBlocksDirty = true
		m.adjustViewport()
		return m, nil

	case "O":
		m.mode = ModeInsert
		// Insert new line before current line
		newLine := ""
		m.content = append(m.content[:m.cursor.row], append([]string{newLine}, m.content[m.cursor.row:]...)...)
		m.cursor.col = 0
		m.saved = false
		m.codeBlocksDirty = true
		m.adjustViewport()
		return m, nil

	case "x":
		// Delete character under cursor (vim-style, continues across lines)
		if m.cursor.col < len(m.content[m.cursor.row]) {
			line := m.content[m.cursor.row]
			m.content[m.cursor.row] = line[:m.cursor.col] + line[m.cursor.col+1:]
			m.saved = false
			m.codeBlocksDirty = true
		} else if m.cursor.row < len(m.content)-1 {
			// At end of line, join with next line
			currentLine := m.content[m.cursor.row]
			nextLine := m.content[m.cursor.row+1]
			m.content[m.cursor.row] = currentLine + nextLine
			m.content = append(m.content[:m.cursor.row+1], m.content[m.cursor.row+2:]...)
			m.saved = false
			m.codeBlocksDirty = true
		}
		return m, nil

	case "dd":
		// Delete current line
		if len(m.content) > 1 {
			m.content = append(m.content[:m.cursor.row], m.content[m.cursor.row+1:]...)
			if m.cursor.row >= len(m.content) {
				m.cursor.row = len(m.content) - 1
			}
			if m.cursor.col > len(m.content[m.cursor.row]) {
				m.cursor.col = len(m.content[m.cursor.row])
			}
			m.saved = false
			m.codeBlocksDirty = true
		} else {
			m.content[0] = ""
			m.cursor.col = 0
			m.saved = false
			m.codeBlocksDirty = true
		}
		m.adjustViewport()
		return m, nil

	case "w":
		// Move to next word
		m.cursor = m.nextWord()
		m.adjustViewport()
		return m, nil

	case "b":
		// Move to previous word
		m.cursor = m.prevWord()
		m.adjustViewport()
		return m, nil

	case "e":
		// Move to end of current word
		m.cursor = m.endOfWord()
		m.adjustViewport()
		return m, nil
	}

	return m, nil
}

func (m *Model) handlePreviewMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only process preview keys if we're actually on the preview tab
	if m.activeTab != TabPreview {
		return m, nil
	}

	switch msg.String() {
	case "j", "down":
		// Calculate max scroll based on rendered content
		markdown := strings.Join(m.content, "\n")
		if strings.TrimSpace(markdown) != "" && m.renderer != nil {
			if rendered, err := m.renderer.Render(markdown); err == nil {
				lines := strings.Split(rendered, "\n")
				contentHeight := m.height - 3 // tab + status + footer
				maxOffset := max(0, len(lines)-contentHeight)
				if m.previewOffset < maxOffset {
					m.previewOffset++
				}
			}
		}
		return m, nil
	case "k", "up":
		if m.previewOffset > 0 {
			m.previewOffset--
		}
		return m, nil
	case "g":
		// Go to top
		m.previewOffset = 0
		return m, nil
	case "G":
		// Go to bottom
		markdown := strings.Join(m.content, "\n")
		if strings.TrimSpace(markdown) != "" && m.renderer != nil {
			if rendered, err := m.renderer.Render(markdown); err == nil {
				lines := strings.Split(rendered, "\n")
				contentHeight := m.height - 3 // tab + status + footer
				m.previewOffset = max(0, len(lines)-contentHeight)
			}
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		if m.cursor.col > 0 {
			m.cursor.col--
		}
		return m, nil

	case "left":
		m.cursor.col = max(0, m.cursor.col-1)
		m.adjustViewport()
		return m, nil

	case "right":
		if m.cursor.col < len(m.content[m.cursor.row]) {
			m.cursor.col++
		}
		m.adjustViewport()
		return m, nil

	case "up":
		if m.cursor.row > 0 {
			m.cursor.row--
			// Adjust column if the new line is shorter
			if m.cursor.col > len(m.content[m.cursor.row]) {
				m.cursor.col = len(m.content[m.cursor.row])
			}
		}
		m.adjustViewport()
		return m, nil

	case "down":
		if m.cursor.row < len(m.content)-1 {
			m.cursor.row++
			// Adjust column if the new line is shorter
			if m.cursor.col > len(m.content[m.cursor.row]) {
				m.cursor.col = len(m.content[m.cursor.row])
			}
		}
		m.adjustViewport()
		return m, nil

	case "enter":
		// Split line at cursor position
		currentLine := m.content[m.cursor.row]
		beforeCursor := currentLine[:m.cursor.col]
		afterCursor := currentLine[m.cursor.col:]

		m.content[m.cursor.row] = beforeCursor
		m.content = append(m.content[:m.cursor.row+1], append([]string{afterCursor}, m.content[m.cursor.row+1:]...)...)

		m.cursor.row++
		m.cursor.col = 0
		m.saved = false
		m.codeBlocksDirty = true
		m.adjustViewport()
		return m, nil

	case "backspace":
		if m.cursor.col > 0 {
			// Delete character before cursor
			line := m.content[m.cursor.row]
			m.content[m.cursor.row] = line[:m.cursor.col-1] + line[m.cursor.col:]
			m.cursor.col--
			m.saved = false
			m.codeBlocksDirty = true
		} else if m.cursor.row > 0 {
			// Join with previous line
			prevLine := m.content[m.cursor.row-1]
			currentLine := m.content[m.cursor.row]
			m.content[m.cursor.row-1] = prevLine + currentLine
			m.content = append(m.content[:m.cursor.row], m.content[m.cursor.row+1:]...)
			m.cursor.row--
			m.cursor.col = len(prevLine)
			m.saved = false
			m.codeBlocksDirty = true
		}
		m.adjustViewport()
		return m, nil

	case "delete":
		if m.cursor.col < len(m.content[m.cursor.row]) {
			// Delete character at cursor
			line := m.content[m.cursor.row]
			m.content[m.cursor.row] = line[:m.cursor.col] + line[m.cursor.col+1:]
			m.saved = false
			m.codeBlocksDirty = true
		} else if m.cursor.row < len(m.content)-1 {
			// At end of line, join with next line
			currentLine := m.content[m.cursor.row]
			nextLine := m.content[m.cursor.row+1]
			m.content[m.cursor.row] = currentLine + nextLine
			m.content = append(m.content[:m.cursor.row+1], m.content[m.cursor.row+2:]...)
			m.saved = false
			m.codeBlocksDirty = true
		}
		return m, nil

	case "ctrl+v", "ctrl+p", "shift+insert":
		// Special paste handler to completely avoid render loops with code blocks
		clipboard := getClipboard()
		if clipboard == "" {
			return m, nil
		}

		// Debug: Allow pasting code blocks but track what happens
		containsCodeBlocks := strings.Contains(clipboard, "```")
		if containsCodeBlocks {
			// Log the issue for debugging
			fmt.Fprintf(os.Stderr, "DEBUG: Pasting code block content, lines=%d\n", len(strings.Split(clipboard, "\n")))
			m.setStatusMsg("Pasting code block (chunked approach)", false)

			// Try a different approach: paste line by line to avoid overwhelming Bubbletea
			lines := strings.Split(clipboard, "\n")
			if len(lines) > 10 { // Only use chunked approach for large pastes
				// Insert first line normally
				line := m.content[m.cursor.row]
				m.content[m.cursor.row] = line[:m.cursor.col] + lines[0]

				// Insert middle lines
				for i := 1; i < len(lines)-1; i++ {
					m.content = append(m.content[:m.cursor.row+i], append([]string{lines[i]}, m.content[m.cursor.row+i:]...)...)
				}

				// Insert last line
				if len(lines) > 1 {
					finalLine := lines[len(lines)-1] + line[m.cursor.col:]
					m.content = append(m.content[:m.cursor.row+len(lines)-1], append([]string{finalLine}, m.content[m.cursor.row+len(lines)-1:]...)...)
				}

				m.cursor.row += len(lines) - 1
				m.cursor.col = len(lines[len(lines)-1])
				m.saved = false
				// Completely disable code block tracking for chunked paste operations
				// m.codeBlocksDirty = true  // DISABLED FOR PASTE
				fmt.Fprintf(os.Stderr, "DEBUG: Chunked paste complete, content_lines=%d, NO codeBlocksDirty set\n", len(m.content))
				return m, nil
			}
		}

		// Ensure cursor bounds
		m.ensureCursorBounds()

		lines := strings.Split(clipboard, "\n")

		if len(lines) == 1 {
			// Single line paste
			line := m.content[m.cursor.row]
			m.content[m.cursor.row] = line[:m.cursor.col] + clipboard + line[m.cursor.col:]
			m.cursor.col += len(clipboard)
		} else {
			// Multi-line paste
			currentLine := m.content[m.cursor.row]
			beforeCursor := currentLine[:m.cursor.col]
			afterCursor := currentLine[m.cursor.col:]

			// Build new content
			newContent := make([]string, 0, len(m.content)+len(lines)-1)
			newContent = append(newContent, m.content[:m.cursor.row]...)
			newContent = append(newContent, beforeCursor+lines[0])
			if len(lines) > 2 {
				newContent = append(newContent, lines[1:len(lines)-1]...)
			}
			newContent = append(newContent, lines[len(lines)-1]+afterCursor)
			newContent = append(newContent, m.content[m.cursor.row+1:]...)

			m.content = newContent
			m.cursor.row += len(lines) - 1
			m.cursor.col = len(lines[len(lines)-1])
		}

		m.saved = false
		// Completely disable code block tracking for paste operations to prevent render loops
		// m.codeBlocksDirty = true  // DISABLED FOR PASTE
		fmt.Fprintf(os.Stderr, "DEBUG: Paste complete, content_lines=%d, NO codeBlocksDirty set\n", len(m.content))
		return m, nil

	default:
		// Insert character
		if len(msg.String()) == 1 {
			char := msg.String()
			line := m.content[m.cursor.row]
			m.content[m.cursor.row] = line[:m.cursor.col] + char + line[m.cursor.col:]
			m.cursor.col++
			m.saved = false
			m.codeBlocksDirty = true
		}
		return m, nil
	}
}

// ensureCursorBounds ensures the cursor is within valid bounds
func (m *Model) ensureCursorBounds() {
	// Ensure we have content
	if len(m.content) == 0 {
		m.content = []string{""}
	}

	// Ensure row is within bounds
	if m.cursor.row < 0 {
		m.cursor.row = 0
	} else if m.cursor.row >= len(m.content) {
		m.cursor.row = len(m.content) - 1
	}

	// Ensure column is within bounds for current row
	if m.cursor.row < len(m.content) {
		maxCol := len(m.content[m.cursor.row])
		if m.cursor.col < 0 {
			m.cursor.col = 0
		} else if m.cursor.col > maxCol {
			m.cursor.col = maxCol
		}
	}
}

func (m *Model) adjustViewport() {
	// Ensure cursor is within bounds first
	m.ensureCursorBounds()

	// Calculate the actual content height available for editor text
	contentHeight := m.height - 3 // tab + status + footer
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Vertical scrolling with improved logic
	if m.cursor.row < m.viewport.offsetRow {
		// Cursor moved above visible area, scroll up
		m.viewport.offsetRow = m.cursor.row
	} else if m.cursor.row >= m.viewport.offsetRow+contentHeight {
		// Cursor moved below visible area, scroll down
		m.viewport.offsetRow = m.cursor.row - contentHeight + 1
	}

	// Ensure viewport doesn't go negative or beyond content
	if m.viewport.offsetRow < 0 {
		m.viewport.offsetRow = 0
	}
	maxOffsetRow := max(0, len(m.content)-contentHeight)
	if m.viewport.offsetRow > maxOffsetRow {
		m.viewport.offsetRow = maxOffsetRow
	}

	// Horizontal scrolling with improved logic
	contentWidth := m.width - 3 // account for UI elements
	if contentWidth < 1 {
		contentWidth = 1
	}

	if m.cursor.col < m.viewport.offsetCol {
		// Cursor moved left of visible area, scroll left
		m.viewport.offsetCol = m.cursor.col
	} else if m.cursor.col >= m.viewport.offsetCol+contentWidth {
		// Cursor moved right of visible area, scroll right
		m.viewport.offsetCol = m.cursor.col - contentWidth + 1
	}

	// Ensure horizontal viewport doesn't go negative
	if m.viewport.offsetCol < 0 {
		m.viewport.offsetCol = 0
	}
}

func (m Model) saveFile() (tea.Model, tea.Cmd) {
	filename := m.filename
	if filename == "" {
		filename = "untitled.md"
		m.filename = filename
	}

	content := strings.Join(m.content, "\n")

	// Create backup if file exists
	if _, err := os.Stat(filename); err == nil {
		backupName := filename + ".bak"
		if backupData, err := os.ReadFile(filename); err == nil {
			os.WriteFile(backupName, backupData, 0644)
		}
	}

	err := os.WriteFile(filename, []byte(content), 0644)

	if err != nil {
		m.setStatusMsg("Error saving file: "+err.Error(), true)
		return m, nil
	}

	m.saved = true
	m.codeBlocksDirty = true // Mark for rebuild since content changed
	m.setStatusMsg("File saved: "+filename, false)
	return m, nil
}

// Word movement functions
func (m Model) nextWord() Position {
	row := m.cursor.row
	col := m.cursor.col

	// Bounds checking
	if row >= len(m.content) {
		if len(m.content) > 0 {
			return Position{row: len(m.content) - 1, col: len(m.content[len(m.content)-1])}
		}
		return Position{row: 0, col: 0}
	}

	line := m.content[row]

	// Skip current word (non-whitespace characters)
	for col < len(line) && !isWhitespace(line[col]) {
		col++
	}

	// Skip whitespace
	for col < len(line) && isWhitespace(line[col]) {
		col++
	}

	// If we're at the end of the line, move to next line
	if col >= len(line) && row < len(m.content)-1 {
		row++
		col = 0
		// Skip leading whitespace on next line
		if row < len(m.content) {
			line = m.content[row]
			for col < len(line) && isWhitespace(line[col]) {
				col++
			}
		}
	}

	return Position{row: row, col: col}
}

// isWhitespace checks if a character is whitespace
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func (m Model) prevWord() Position {
	row := m.cursor.row
	col := m.cursor.col

	// Bounds checking
	if row >= len(m.content) || row < 0 {
		return Position{row: 0, col: 0}
	}

	if col > 0 {
		col--
	} else if row > 0 {
		row--
		if row < len(m.content) {
			col = len(m.content[row])
		}
	}

	if row < 0 {
		return Position{row: 0, col: 0}
	}
	if row >= len(m.content) {
		return Position{row: len(m.content) - 1, col: 0}
	}

	line := m.content[row]

	// Skip whitespace backwards
	for col > 0 && col < len(line) && isWhitespace(line[col]) {
		col--
	}

	// Skip word backwards
	for col > 0 && col < len(line) && !isWhitespace(line[col]) {
		col--
	}

	// Move to start of word
	if col > 0 && col < len(line) && isWhitespace(line[col]) {
		col++
	}

	return Position{row: row, col: col}
}

func (m Model) endOfWord() Position {
	row := m.cursor.row
	col := m.cursor.col

	// Bounds checking
	if row >= len(m.content) {
		if len(m.content) > 0 {
			return Position{row: len(m.content) - 1, col: len(m.content[len(m.content)-1])}
		}
		return Position{row: 0, col: 0}
	}

	line := m.content[row]

	// If we're at the end of a word, move to next word first
	if col < len(line) && !isWhitespace(line[col]) {
		// Move to end of current word
		for col < len(line) && !isWhitespace(line[col]) {
			col++
		}
		if col > 0 {
			col--
		}
		return Position{row: row, col: col}
	}

	// Skip whitespace to find next word
	for col < len(line) && isWhitespace(line[col]) {
		col++
	}

	// Move to end of next word
	for col < len(line) && !isWhitespace(line[col]) {
		col++
	}

	if col > 0 {
		col--
	}

	return Position{row: row, col: col}
}

// getClipboard attempts to get clipboard content using various clipboard tools
// Returns empty string if no clipboard tool is available or clipboard is empty
func getClipboard() string {
	// Set a reasonable timeout for clipboard operations
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try xclip first (X11)
	cmd := exec.CommandContext(ctx, "xclip", "-o", "-selection", "clipboard")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimRight(string(output), "\n")
	}

	// Try wl-paste (Wayland)
	cmd = exec.CommandContext(ctx, "wl-paste")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimRight(string(output), "\n")
	}

	// Try pbpaste (macOS)
	cmd = exec.CommandContext(ctx, "pbpaste")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimRight(string(output), "\n")
	}

	// No clipboard tool available or all failed
	return ""
}

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

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
