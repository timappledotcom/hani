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
)

// Define reusable styles
var (
	activeTabStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	inactiveTabStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3C3C3C")).
				Foreground(lipgloss.Color("#CCCCCC")).
				Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E1E1E")).
			Foreground(lipgloss.Color("#CCCCCC")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2D2D2D")).
			Foreground(lipgloss.Color("#CCCCCC")).
			Padding(0, 1)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	contentStyle = lipgloss.NewStyle().
			Padding(0)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)
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

	// Initialize glamour renderer with configuration (lazy initialization for better startup performance)
	var renderer *glamour.TermRenderer
	wordWrap := config.WordWrap
	if wordWrap == 0 {
		wordWrap = DefaultWordWrap
	}

	// Only initialize renderer if we have a reasonable terminal size
	// This improves startup performance significantly
	if wordWrap > MinWordWrap && wordWrap < MaxWordWrap*2 {
		if r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(wordWrap),
		); err == nil {
			renderer = r
		} else if lastError == nil {
			lastError = fmt.Errorf("failed to initialize markdown renderer: %w", err)
		}
	}

	// Initialize syntax highlighter (lazy loading for better startup performance)
	var highlighter *SyntaxHighlighter
	// We'll initialize this on first use to improve startup time and memory usage

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
	for i := range checkLen {
		if data[i] == 0 {
			return true
		}
	}

	// Check for high ratio of non-printable characters
	nonPrintable := 0
	for i := range checkLen {
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

	// Lazy initialization of syntax highlighter for better performance
	if m.highlighter == nil && m.activeTab == TabEditor {
		m.highlighter = NewSyntaxHighlighter()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		oldWidth := m.width
		m.width = msg.Width
		m.height = msg.Height

		// Only update glamour renderer if width changed significantly (performance optimization)
		if m.width > 20 && m.renderer != nil && abs(m.width-oldWidth) > 10 {
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

		// Adjust viewport and cursor bounds after resize
		m.ensureCursorBounds()
		m.adjustViewport()

		// Reset preview offset if it's now out of bounds (only if in preview mode)
		if m.activeTab == TabPreview && m.previewOffset > 0 {
			contentHeight := m.height - 3 // tab + status + footer
			// Simple bounds check without expensive rendering
			if m.previewOffset > contentHeight {
				m.previewOffset = max(0, m.previewOffset-contentHeight)
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
	// Handle initialization state
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Handle small terminal size gracefully
	if m.height < 6 {
		return lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Terminal too small")
	}

	// Create UI elements
	tabBar := m.renderTabBar()
	statusBar := m.renderStatusBar()
	footer := m.renderFooter()

	// Calculate content area height more accurately
	// Account for tab bar, status bar, and footer
	contentHeight := m.height - 3 // tab + status + footer
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Create content based on active tab
	var content string
	if m.activeTab == TabEditor {
		content = m.renderEditor(contentHeight)
	} else {
		content = m.renderPreview(contentHeight)
	}

	// Use simple vertical join for better fullscreen handling
	// This avoids complex container styling that can cause layout issues
	return lipgloss.JoinVertical(lipgloss.Top,
		tabBar,
		content,
		statusBar,
		footer,
	)
}

func (m Model) renderEditor(height int) string {
	lines := make([]string, height)

	// Note: Since we're using a value receiver, we can't modify m.highlighter here
	// The proper initialization should happen in Update or a method with pointer receiver
	// We'll just use the highlighter if it's available

	for i := range height {
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

		// Use plain text without syntax highlighting for clean editing experience
		// We'll apply syntax highlighting only when needed for better performance
		displayLine := visibleLine

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

	// Simple cursor insertion for plain text
	displayRunes := []rune(displayLine)
	if cursorPos < len(displayRunes) {
		return string(displayRunes[:cursorPos]) + "█" + string(displayRunes[cursorPos:])
	}

	return displayLine + "█"
}

func (m Model) renderPreview(height int) string {
	// Lazy rendering: Only render when we're actually on the preview tab
	// This prevents expensive markdown rendering when on editor tab
	if m.activeTab != TabPreview {
		return "Preview not rendered (not active tab)"
	}

	// Only render if we have content and a renderer
	if m.renderer == nil || len(m.content) == 0 {
		return "Preview not available"
	}

	markdown := strings.Join(m.content, "\n")
	if strings.TrimSpace(markdown) == "" {
		return "No content to preview"
	}

	// Render markdown using glamour (with caching for performance)
	var rendered string
	if out, err := m.renderer.Render(markdown); err != nil {
		rendered = "Error rendering markdown: " + err.Error()
	} else {
		rendered = out
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
		style := statusBarStyle
		if m.lastError != nil {
			style = errorStyle.Background(lipgloss.Color("#1E1E1E")).Padding(0, 1)
		}
		return style.Width(m.width).Render(m.statusMsg)
	}

	// Mode indicator
	var modeStr string
	var modeStyle lipgloss.Style
	if m.mode == ModeInsert {
		modeStr = "INSERT"
		modeStyle = keyStyle.Background(lipgloss.Color("#7D56F4")).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)
	} else {
		modeStr = "NORMAL"
		modeStyle = lipgloss.NewStyle().Background(lipgloss.Color("#4A4A4A")).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)
	}

	// File status
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

	// Position
	position := fmt.Sprintf("(%d,%d)", m.cursor.row+1, m.cursor.col+1)

	// Error indicator
	errorIndicator := ""
	if m.lastError != nil {
		errorIndicator = errorStyle.Render(" ⚠️")
	}

	// Use Lipgloss to layout the status bar
	leftSection := lipgloss.JoinHorizontal(lipgloss.Left,
		modeStyle.Render(modeStr),
		statusBarStyle.Render(" "+fileStatus+" "),
	)

	rightSection := lipgloss.JoinHorizontal(lipgloss.Right,
		statusBarStyle.Render(position),
		errorIndicator,
	)

	// Calculate spacing
	usedWidth := lipgloss.Width(leftSection) + lipgloss.Width(rightSection)
	spacerWidth := max(0, m.width-usedWidth)
	spacer := strings.Repeat(" ", spacerWidth)

	return statusBarStyle.Width(m.width).Render(
		lipgloss.JoinHorizontal(lipgloss.Left, leftSection, spacer, rightSection),
	)
}

func (m Model) renderTabBar() string {
	// Use the global styles we defined
	var editorTab, previewTab string

	if m.activeTab == TabEditor {
		editorTab = activeTabStyle.Render("Editor")
		previewTab = inactiveTabStyle.Render("Preview")
	} else {
		editorTab = inactiveTabStyle.Render("Editor")
		previewTab = activeTabStyle.Render("Preview")
	}

	// Create tabs section
	tabsSection := lipgloss.JoinHorizontal(lipgloss.Left, editorTab, previewTab)

	// Instructions section
	instructions := separatorStyle.Render("Tab/Shift+Tab to switch")

	// Use Lipgloss to properly layout the tab bar with responsive spacing
	return lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#1E1E1E")).
		Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				tabsSection,
				lipgloss.NewStyle().Width(m.width-lipgloss.Width(tabsSection)-lipgloss.Width(instructions)).Render(""),
				instructions,
			),
		)
}

// renderFooter renders the footer with key command hints
func (m Model) renderFooter() string {
	var commands []string

	if m.activeTab == TabEditor {
		if m.mode == ModeNormal {
			// Normal mode commands
			commands = []string{
				keyStyle.Render("i") + " Insert",
				keyStyle.Render("Tab") + " Preview",
				keyStyle.Render("Ctrl+S") + " Save",
				keyStyle.Render("o") + " New Line",
				keyStyle.Render("dd") + " Delete Line",
				keyStyle.Render("Ctrl+Q") + " Quit",
			}
		} else {
			// Insert mode commands
			commands = []string{
				keyStyle.Render("Esc") + " Normal",
				keyStyle.Render("Tab") + " Preview",
				keyStyle.Render("Ctrl+S") + " Save",
				keyStyle.Render("Enter") + " New Line",
				keyStyle.Render("Ctrl+V") + " Paste",
				keyStyle.Render("Ctrl+Q") + " Quit",
			}
		}
	} else {
		// Preview mode commands
		commands = []string{
			keyStyle.Render("Tab") + " Editor",
			keyStyle.Render("j/k") + " Scroll",
			keyStyle.Render("g/G") + " Top/Bottom",
			keyStyle.Render("Ctrl+S") + " Save",
			keyStyle.Render("Ctrl+Q") + " Quit",
		}
	}

	// Use Lipgloss to join commands with separators
	separator := separatorStyle.Render(" │ ")
	commandText := strings.Join(commands, separator)

	// Use the global footerStyle and let Lipgloss handle width and truncation
	return footerStyle.
		Width(m.width).
		MaxWidth(m.width).
		Render(commandText)
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
		if lang, found := strings.CutPrefix(line, "```"); found {
			if !inCodeBlock {
				// Start of code block
				inCodeBlock = true
				currentBlock = CodeBlock{
					start: i,
					lang:  strings.TrimSpace(lang),
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
