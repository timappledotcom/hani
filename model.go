package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Configuration constants
const (
	DefaultWordWrap   = 80
	MaxWordWrap       = 120
	MinWordWrap       = 40
	WordWrapMargin    = 10
	CursorBlinkRate   = 500 * time.Millisecond
	StatusMsgDuration = 2 * time.Second
	ErrorMsgDuration  = 3 * time.Second
	MaxFileSize       = 10 * 1024 * 1024 // 10MB limit
	UIOverhead        = 5                // Lines used for UI elements
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
)

type Tab int

const (
	TabEditor Tab = iota
	TabPreview
)

type Model struct {
	filename         string
	content          []string
	cursor           Position
	mode             Mode
	activeTab        Tab
	width            int
	height           int
	viewport         Viewport
	previewOffset    int
	renderer         *glamour.TermRenderer
	highlighter      *SyntaxHighlighter
	saved            bool
	statusMsg        string
	statusMsgTimeout time.Time
	cursorBlink      bool
	codeBlocks       []CodeBlock
	codeBlocksDirty  bool
	config           Config
	lastError        error
}

type Position struct {
	row int
	col int
}

type Viewport struct {
	offsetRow int
	offsetCol int
}

type BlinkMsg struct{}

type CodeBlock struct {
	start int
	end   int
	lang  string
}

func NewModel(filename string) Model {
	content := []string{""}
	saved := false
	var statusMsg string
	var lastError error

	// Load configuration
	config := LoadConfig()

	// Load file if it exists
	if filename != "" {
		if info, err := os.Stat(filename); err == nil {
			// Check file size
			if info.Size() > MaxFileSize {
				statusMsg = fmt.Sprintf("File too large (%d MB). Maximum size is %d MB",
					info.Size()/(1024*1024), MaxFileSize/(1024*1024))
				lastError = fmt.Errorf("file too large: %d bytes", info.Size())
			} else if data, err := os.ReadFile(filename); err == nil {
				// Check if file is binary
				if isBinaryFile(data) {
					statusMsg = "Cannot edit binary file: " + filename
					lastError = fmt.Errorf("binary file detected")
				} else {
					content = strings.Split(string(data), "\n")
					if len(content) > 0 && content[len(content)-1] == "" {
						content = content[:len(content)-1]
					}
					saved = true
				}
			} else {
				statusMsg = "Error reading file: " + err.Error()
				lastError = err
				saved = false
			}
		} else {
			// File doesn't exist - this is okay for new files
			statusMsg = "New file: " + filename
			saved = false
		}
	} else {
		saved = true
	}

	// Initialize glamour renderer with configuration
	wordWrap := config.WordWrap
	if wordWrap == 0 {
		wordWrap = DefaultWordWrap
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(wordWrap),
	)
	if err != nil {
		// Fallback to nil renderer if initialization fails
		renderer = nil
		if lastError == nil {
			lastError = fmt.Errorf("failed to initialize markdown renderer: %w", err)
		}
	}

	// Initialize syntax highlighter
	highlighter := NewSyntaxHighlighter()
	if highlighter == nil && lastError == nil {
		lastError = fmt.Errorf("failed to initialize syntax highlighter")
	}

	m := Model{
		filename:         filename,
		content:          content,
		cursor:           Position{row: 0, col: 0},
		mode:             ModeNormal,
		activeTab:        TabEditor,
		viewport:         Viewport{offsetRow: 0, offsetCol: 0},
		renderer:         renderer,
		highlighter:      highlighter,
		saved:            saved,
		statusMsg:        statusMsg,
		statusMsgTimeout: time.Now().Add(StatusMsgDuration),
		cursorBlink:      true,
		codeBlocksDirty:  true,
		config:           config,
		lastError:        lastError,
	}

	// Initialize code blocks
	m.rebuildCodeBlocks()

	return m
}

// isBinaryFile checks if the file content appears to be binary
func isBinaryFile(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Check for null bytes in first 512 bytes (common binary indicator)
	checkLen := min(len(data), 512)
	for i := 0; i < checkLen; i++ {
		if data[i] == 0 {
			return true
		}
	}

	// Check for high ratio of non-printable characters
	nonPrintable := 0
	for i := 0; i < checkLen; i++ {
		if data[i] < 32 && data[i] != 9 && data[i] != 10 && data[i] != 13 {
			nonPrintable++
		}
	}

	return float64(nonPrintable)/float64(checkLen) > 0.3
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(CursorBlinkRate, func(t time.Time) tea.Msg {
		return BlinkMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Clear expired status messages
	if m.statusMsg != "" && time.Now().After(m.statusMsgTimeout) {
		m.statusMsg = ""
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update glamour renderer with new word wrap based on width
		if m.width > 20 && m.renderer != nil {
			wordWrap := m.width - WordWrapMargin
			if wordWrap > MaxWordWrap {
				wordWrap = MaxWordWrap
			} else if wordWrap < MinWordWrap {
				wordWrap = MinWordWrap
			}

			if renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(wordWrap),
			); err == nil {
				m.renderer = renderer
			} else {
				m.setStatusMsg("Warning: Failed to update renderer", false)
			}
		}

		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case BlinkMsg:
		m.cursorBlink = !m.cursorBlink
		return m, tea.Tick(CursorBlinkRate, func(t time.Time) tea.Msg {
			return BlinkMsg{}
		})
	}

	return m, nil
}

// setStatusMsg sets a status message with timeout
func (m *Model) setStatusMsg(msg string, isError bool) {
	m.statusMsg = msg
	if isError {
		m.statusMsgTimeout = time.Now().Add(ErrorMsgDuration)
	} else {
		m.statusMsgTimeout = time.Now().Add(StatusMsgDuration)
	}
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Create tab bar
	tabBar := m.renderTabBar()

	contentHeight := m.height - 5 // Account for tab bar and status bar
	var content string

	if m.activeTab == TabEditor {
		content = m.renderEditor(contentHeight)
	} else {
		content = m.renderPreview(contentHeight)
	}

	// Status bar
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Top, tabBar, content, statusBar)
}

func (m Model) renderEditor(height int) string {
	lines := make([]string, height)

	for i := 0; i < height; i++ {
		lineNum := m.viewport.offsetRow + i
		if lineNum >= len(m.content) {
			lines[i] = "~"
			continue
		}

		originalLine := m.content[lineNum]

		// Handle horizontal scrolling on original line
		visibleLine := originalLine
		if m.viewport.offsetCol > 0 {
			if m.viewport.offsetCol < len(originalLine) {
				visibleLine = originalLine[m.viewport.offsetCol:]
			} else {
				visibleLine = ""
			}
		}

		// Apply syntax highlighting to visible portion
		displayLine := visibleLine
		if m.highlighter != nil {
			if inCodeBlock, lang := m.isInCodeBlock(lineNum); inCodeBlock {
				displayLine = m.highlighter.HighlightCodeBlock(visibleLine, lang)
			} else {
				displayLine = m.highlighter.HighlightMarkdownLine(visibleLine)
			}
		}

		// Add cursor if this is the cursor line and cursor is visible
		if lineNum == m.cursor.row && m.cursorBlink {
			cursorPos := m.cursor.col - m.viewport.offsetCol
			if cursorPos >= 0 && cursorPos <= len(visibleLine) {
				// Insert cursor without breaking syntax highlighting
				displayLine = m.insertCursor(displayLine, visibleLine, cursorPos)
			}
		}

		lines[i] = displayLine
	}

	return strings.Join(lines, "\n")
}

// insertCursor safely inserts cursor into display line
func (m Model) insertCursor(displayLine, originalLine string, cursorPos int) string {
	if cursorPos >= len(originalLine) {
		// Cursor at end of line
		return displayLine + "█"
	}

	// For syntax highlighted text, we need to be careful about ANSI codes
	// Simple approach: convert to runes and insert cursor
	displayRunes := []rune(displayLine)
	originalRunes := []rune(originalLine)

	// If display line is longer due to ANSI codes, find the right position
	if len(displayRunes) > len(originalRunes) {
		// Count visible characters up to cursor position
		visibleChars := 0
		insertPos := 0
		inAnsiCode := false

		for i, r := range displayRunes {
			if r == '\x1b' {
				inAnsiCode = true
			} else if inAnsiCode && r == 'm' {
				inAnsiCode = false
				insertPos = i + 1
				continue
			}

			if !inAnsiCode {
				if visibleChars == cursorPos {
					insertPos = i
					break
				}
				visibleChars++
				insertPos = i + 1
			}
		}

		if insertPos <= len(displayRunes) {
			return string(displayRunes[:insertPos]) + "█" + string(displayRunes[insertPos:])
		}
	}

	// Fallback: simple insertion
	if cursorPos < len(displayRunes) {
		return string(displayRunes[:cursorPos]) + "█" + string(displayRunes[cursorPos:])
	}

	return displayLine + "█"
}

func (m Model) renderPreview(height int) string {
	markdown := strings.Join(m.content, "\n")

	if strings.TrimSpace(markdown) == "" {
		return "No content to preview"
	}

	// Render markdown using glamour
	var rendered string
	if m.renderer != nil {
		if out, err := m.renderer.Render(markdown); err != nil {
			rendered = "Error rendering markdown: " + err.Error()
		} else {
			rendered = out
		}
	} else {
		rendered = "Renderer not initialized"
	}

	// Apply scrolling by splitting into lines and applying offset
	lines := strings.Split(rendered, "\n")

	// Calculate safe offset bounds
	offset := m.previewOffset
	if offset < 0 {
		offset = 0
	}
	maxOffset := max(0, len(lines)-height)
	if offset > maxOffset {
		offset = maxOffset
	}

	// Get the visible portion based on offset and height
	startLine := offset
	endLine := min(startLine+height, len(lines))

	if startLine >= len(lines) {
		return "End of preview"
	}

	visibleLines := lines[startLine:endLine]

	// Pad with empty lines if needed to fill the height
	for len(visibleLines) < height {
		visibleLines = append(visibleLines, "")
	}

	return strings.Join(visibleLines, "\n")
}

func (m Model) renderStatusBar() string {
	// Show status message if active and not expired
	if m.statusMsg != "" && time.Now().Before(m.statusMsgTimeout) {
		return m.statusMsg
	}

	var modeStr string
	if m.mode == ModeInsert {
		modeStr = "INSERT"
	} else {
		modeStr = "NORMAL"
	}

	var fileStatus string
	if m.filename != "" {
		fileStatus = m.filename
		if !m.saved {
			fileStatus += " [modified]"
		}
	} else {
		fileStatus = "[New File]"
		if !m.saved {
			fileStatus += " [modified]"
		}
	}

	position := fmt.Sprintf("(%d,%d)", m.cursor.row+1, m.cursor.col+1)

	// Show error indicator if there's a last error
	errorIndicator := ""
	if m.lastError != nil {
		errorIndicator = " ⚠️"
	}

	return fmt.Sprintf(" %s | %s | %s%s ", modeStr, fileStatus, position, errorIndicator)
}

func (m Model) renderTabBar() string {
	activeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Bold(true)

	inactiveStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#3C3C3C")).
		Foreground(lipgloss.Color("#CCCCCC")).
		Padding(0, 1)

	var editorTab, previewTab string

	if m.activeTab == TabEditor {
		editorTab = activeStyle.Render("Editor")
		previewTab = inactiveStyle.Render("Preview")
	} else {
		editorTab = inactiveStyle.Render("Editor")
		previewTab = activeStyle.Render("Preview")
	}

	// Add spacing and instructions
	spacer := strings.Repeat(" ", max(0, m.width-len("Editor")-len("Preview")-20))
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render("Tab/Shift+Tab to switch")

	return editorTab + previewTab + spacer + instructions
}

// rebuildCodeBlocks analyzes the content and identifies code blocks
func (m *Model) rebuildCodeBlocks() {
	// Only rebuild if dirty
	if !m.codeBlocksDirty {
		return
	}

	m.codeBlocks = []CodeBlock{}
	inCodeBlock := false
	var currentBlock CodeBlock

	for i, line := range m.content {
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				// Start of code block
				inCodeBlock = true
				currentBlock = CodeBlock{
					start: i,
					lang:  strings.TrimSpace(strings.TrimPrefix(line, "```")),
				}
			} else {
				// End of code block
				inCodeBlock = false
				currentBlock.end = i
				m.codeBlocks = append(m.codeBlocks, currentBlock)
			}
		}
	}

	// Handle unclosed code block
	if inCodeBlock {
		currentBlock.end = len(m.content) - 1
		m.codeBlocks = append(m.codeBlocks, currentBlock)
	}

	m.codeBlocksDirty = false
}

// isInCodeBlock checks if a line is inside a code block
func (m *Model) isInCodeBlock(lineNum int) (bool, string) {
	for _, block := range m.codeBlocks {
		// Include lines within the code block content (not the fences)
		if lineNum > block.start && lineNum < block.end {
			return true, block.lang
		}
	}
	return false, ""
}

// isCodeFence checks if a line is a code fence (``` line)
func (m *Model) isCodeFence(lineNum int) bool {
	if lineNum >= len(m.content) {
		return false
	}
	return strings.HasPrefix(m.content[lineNum], "```")
}
