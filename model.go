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
	DefaultWordWrap     = 80
	MaxWordWrap        = 120
	MinWordWrap        = 40
	WordWrapMargin     = 10
	CursorBlinkRate    = 500 * time.Millisecond
	StatusMsgDuration  = 2 * time.Second
	ErrorMsgDuration   = 3 * time.Second
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
	filename      string
	content       []string
	cursor        Position
	mode          Mode
	activeTab     Tab
	width         int
	height        int
	viewport      Viewport
	previewOffset int
	renderer      *glamour.TermRenderer
	highlighter   *SyntaxHighlighter
	saved         bool
	statusMsg     string
	cursorBlink   bool
	codeBlocks    []CodeBlock
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

	// Load file if it exists
	if filename != "" {
		if data, err := os.ReadFile(filename); err == nil {
			content = strings.Split(string(data), "\n")
			if len(content) > 0 && content[len(content)-1] == "" {
				content = content[:len(content)-1]
			}
			saved = true
		} else {
			// File doesn't exist or can't be read - show status message
			statusMsg = "Could not load file: " + filename + " (starting with new file)"
			saved = false
		}
	} else {
		saved = true
	}

	// Initialize glamour renderer with defaults
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(DefaultWordWrap), // Will be updated when window size is known
	)
	if err != nil {
		// Fallback to nil renderer if initialization fails
		renderer = nil
	}

	// Initialize syntax highlighter
	highlighter := NewSyntaxHighlighter()

	m := Model{
		filename:    filename,
		content:     content,
		cursor:      Position{row: 0, col: 0},
		mode:        ModeNormal,
		activeTab:   TabEditor,
		viewport:    Viewport{offsetRow: 0, offsetCol: 0},
		renderer:    renderer,
		highlighter: highlighter,
		saved:       saved,
		statusMsg:   statusMsg,
		cursorBlink: true,
	}

	// Initialize code blocks
	m.rebuildCodeBlocks()

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(CursorBlinkRate, func(t time.Time) tea.Msg {
		return BlinkMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update glamour renderer with new word wrap based on width
		if m.width > 20 {
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

		line := m.content[lineNum]

		// Handle horizontal scrolling
		if m.viewport.offsetCol > 0 {
			if m.viewport.offsetCol < len(line) {
				line = line[m.viewport.offsetCol:]
			} else {
				line = ""
			}
		}

		// Apply syntax highlighting first
		displayLine := line
		if m.highlighter != nil {
			if inCodeBlock, lang := m.isInCodeBlock(lineNum); inCodeBlock {
				displayLine = m.highlighter.HighlightCodeBlock(line, lang)
			} else {
				displayLine = m.highlighter.HighlightMarkdownLine(line)
			}
		}

		// Show cursor if this is the cursor line and we're in the visible area
		if lineNum == m.cursor.row && m.cursorBlink {
			// Calculate cursor position considering horizontal scrolling
			cursorPos := m.cursor.col - m.viewport.offsetCol
			if cursorPos >= 0 && cursorPos <= len(line)-m.viewport.offsetCol {
				// Insert cursor at the correct position
				if cursorPos == len(line)-m.viewport.offsetCol {
					displayLine += "█"
				} else if cursorPos < len(displayLine) {
					// Insert cursor in the middle of the line
					runes := []rune(displayLine)
					if cursorPos < len(runes) {
						displayLine = string(runes[:cursorPos]) + "█" + string(runes[cursorPos:])
					} else {
						displayLine += "█"
					}
				}
			}
		}

		lines[i] = displayLine
	}

	return strings.Join(lines, "\n")
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

	return rendered
}

func (m Model) renderStatusBar() string {
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

	if m.statusMsg != "" {
		return m.statusMsg
	}

	return fmt.Sprintf(" %s | %s | %s ", modeStr, fileStatus, position)
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
