package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha color palette
var (
	CTPRosewater = "#f5e0dc"
	CTPFlamingo = "#f2cdcd"
	CTPPink     = "#f5c2e7"
	CTPMauve    = "#cba6f7"
	CTPRed      = "#f38ba8"
	CTPMaroon   = "#eba0ac"
	CTPPeach    = "#fab387"
	CTPYellow   = "#f9e2af"
	CTPGreen    = "#a6e3a1"
	CTPTeal     = "#94e2d5"
	CTPSky      = "#89dceb"
	CTPSapphire = "#74c7ec"
	CTPBlue     = "#89b4fa"
	CTPLavender = "#b4befe"
	CTPText     = "#cdd6f4"
	CTPSubtext1 = "#bac2de"
	CTPSubtext0 = "#a6adc8"
	CTPOverlay2 = "#9399b2"
	CTPOverlay1 = "#7f849c"
	CTPOverlay0 = "#6c7086"
	CTPSurface2 = "#585b70"
	CTPSurface1 = "#45475a"
	CTPSurface0 = "#313244"
	CTPBase     = "#1e1e2e"
	CTPMantle   = "#181825"
	CTPCrust    = "#11111b"
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
	filename    string
	content     []string
	cursor      Position
	mode        Mode
	activeTab   Tab
	width       int
	height      int
	viewport    Viewport
	renderer    *glamour.TermRenderer
	highlighter *SyntaxHighlighter
	saved       bool
	statusMsg   string
	cursorBlink bool
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

func NewModel(filename string) Model {
	content := []string{""}
	saved := false // New files start as unsaved

	// Load file if it exists
	if filename != "" {
		if data, err := ioutil.ReadFile(filename); err == nil {
			content = strings.Split(string(data), "\n")
			// Remove empty line at the end if file ends with newline
			if len(content) > 0 && content[len(content)-1] == "" {
				content = content[:len(content)-1]
			}
			saved = true // Existing files start as saved
		}
	} else {
		// No filename provided, start as saved for empty content
		saved = true
	}

	// Initialize glamour renderer with dark theme
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithEnvironmentConfig(),
		glamour.WithWordWrap(80),
	)
	
	// Initialize syntax highlighter
	highlighter := NewSyntaxHighlighter()

	return Model{
		filename:    filename,
		content:     content,
		cursor:      Position{row: 0, col: 0},
		mode:        ModeNormal,
		activeTab:   TabEditor,
		viewport:    Viewport{offsetRow: 0, offsetCol: 0},
		renderer:    renderer,
		highlighter: highlighter,
		saved:       saved,
		cursorBlink: true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return BlinkMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case BlinkMsg:
		m.cursorBlink = !m.cursorBlink
		return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return BlinkMsg{}
		})
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Create main content area - account for window border (no tab bar)
	contentHeight := m.height - 8 // Leave space for status bar, help bar, and window border
	var content string

	if m.activeTab == TabEditor {
		content = m.renderEditor(contentHeight)
	} else {
		content = m.renderPreview(contentHeight)
	}

	// Create status bar
	statusBar := m.renderStatusBar()

	// Create help bar
	helpBar := m.renderHelpBar()

	// Join all components (no tab bar)
	mainContent := lipgloss.JoinVertical(
		lipgloss.Top,
		content,
		statusBar,
		helpBar,
	)

	// Add window border - let it size itself based on content
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(CTPBlue)).
		Background(lipgloss.Color(CTPBase)).
		Padding(0, 1)

	return windowStyle.Render(mainContent)
}

func (m Model) renderTabBar() string {
	var tabs []string

	// Inactive tab style
	inactiveStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Background(lipgloss.Color(CTPSurface0)).
		Foreground(lipgloss.Color(CTPSubtext0)).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(CTPSurface2))

	// Active tab style
	activeStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Background(lipgloss.Color(CTPBlue)).
		Foreground(lipgloss.Color(CTPCrust)).
		Bold(true).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(CTPBlue))

	// Render tabs
	if m.activeTab == TabEditor {
		tabs = append(tabs, activeStyle.Render("📝 Editor"))
		tabs = append(tabs, inactiveStyle.Render("👁  Preview"))
	} else {
		tabs = append(tabs, inactiveStyle.Render("📝 Editor"))
		tabs = append(tabs, activeStyle.Render("👁  Preview"))
	}

	// Join tabs horizontally
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Fill the rest of the line with background
	remaining := (m.width - 6) - lipgloss.Width(tabBar)
	if remaining > 0 {
		filler := lipgloss.NewStyle().
			Width(remaining).
			Background(lipgloss.Color(CTPMantle)).
			Render("")
		tabBar = lipgloss.JoinHorizontal(lipgloss.Top, tabBar, filler)
	}

	// Add top and bottom borders to the entire tab bar
	tabBarStyle := lipgloss.NewStyle().
		Width(m.width - 6). // Account for window border and padding
		Background(lipgloss.Color(CTPMantle)).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(CTPSurface2))

	return tabBarStyle.Render(tabBar)
}

func (m Model) renderEditor(height int) string {
	lines := make([]string, height)

	for i := 0; i < height; i++ {
		lineNum := m.viewport.offsetRow + i
		if lineNum >= len(m.content) {
			// Style the tilde markers
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(CTPOverlay0)).Render("~")
			continue
		}

		line := m.content[lineNum]
		
		// Apply syntax highlighting to the line
		if m.highlighter != nil {
			line = m.highlighter.highlightMarkdownLine(line)
		}
		
		// Show cursor in current line (after highlighting)
		if lineNum == m.cursor.row && m.cursor.col <= len(m.content[lineNum]) {
			cursorChar := " "
			if m.cursorBlink {
				cursorChar = lipgloss.NewStyle().Background(lipgloss.Color(CTPText)).Foreground(lipgloss.Color(CTPBase)).Render("█")
			}
			
			// Insert cursor at the right position
			originalLine := m.content[lineNum]
			if m.cursor.col == len(originalLine) {
				line += cursorChar
			} else {
				// For cursor insertion with syntax highlighting, we need to be more careful
				// This is a simplified approach - in a more complex editor, you'd want
				// to handle this with proper cursor positioning after highlighting
				plainLine := m.content[lineNum]
				if m.cursor.col < len(plainLine) {
					// Re-highlight with cursor inserted
					lineWithCursor := plainLine[:m.cursor.col] + "█" + plainLine[m.cursor.col+1:]
					if m.highlighter != nil {
						line = m.highlighter.highlightMarkdownLine(lineWithCursor)
					}
				}
			}
		}

		lines[i] = line
	}

	content := strings.Join(lines, "\n")

	style := lipgloss.NewStyle().
		Width(m.width - 6). // Account for window border and padding
		Height(height).
		Padding(1).
		Foreground(lipgloss.Color(CTPText)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(CTPSurface2))

	return style.Render(content)
}

func (m Model) renderPreview(height int) string {
	// Join all content lines
	markdown := strings.Join(m.content, "\n")

	// If content is empty, show placeholder
	if strings.TrimSpace(markdown) == "" {
		placeholder := "No content to preview\n\nStart typing in the editor to see a live preview here!"
		style := lipgloss.NewStyle().
			Width(m.width - 6). // Account for window border and padding
			Height(height).
			Padding(1).
			Foreground(lipgloss.Color(CTPSubtext0)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(CTPBlue)).
			Align(lipgloss.Left)
		return style.Render(placeholder)
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

	// Style the preview
	style := lipgloss.NewStyle().
		Width(m.width - 6). // Account for window border and padding
		Height(height).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(CTPBlue))

	return style.Render(rendered)
}

func (m Model) htmlToText(html string) string {
	// This is a very basic HTML to text converter
	// In a real implementation, you'd want to use a proper HTML parser
	// and convert to terminal formatting

	text := html

	// Remove HTML tags (basic)
	text = strings.ReplaceAll(text, "<p>", "")
	text = strings.ReplaceAll(text, "</p>", "\n\n")
	text = strings.ReplaceAll(text, "<h1>", "# ")
	text = strings.ReplaceAll(text, "</h1>", "\n\n")
	text = strings.ReplaceAll(text, "<h2>", "## ")
	text = strings.ReplaceAll(text, "</h2>", "\n\n")
	text = strings.ReplaceAll(text, "<h3>", "### ")
	text = strings.ReplaceAll(text, "</h3>", "\n\n")
	text = strings.ReplaceAll(text, "<strong>", "**")
	text = strings.ReplaceAll(text, "</strong>", "**")
	text = strings.ReplaceAll(text, "<em>", "*")
	text = strings.ReplaceAll(text, "</em>", "*")
	text = strings.ReplaceAll(text, "<li>", "• ")
	text = strings.ReplaceAll(text, "</li>", "\n")
	text = strings.ReplaceAll(text, "<ul>", "")
	text = strings.ReplaceAll(text, "</ul>", "\n")
	text = strings.ReplaceAll(text, "<ol>", "")
	text = strings.ReplaceAll(text, "</ol>", "\n")
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<hr>", "---\n")

	return text
}

func (m *Model) toggleTab() {
	if m.activeTab == TabEditor {
		m.activeTab = TabPreview
	} else {
		m.activeTab = TabEditor
	}
}

func (m Model) renderHelpBar() string {
	// Define help text based on current mode and tab
	var helpText string
	if m.activeTab == TabEditor {
		if m.mode == ModeInsert {
			helpText = "ESC: Normal | TAB: Preview | CTRL+V: Paste | CTRL+S: Save | CTRL+Q: Quit"
		} else {
			helpText = "I: Insert | X: Delete | TAB: Preview | CTRL+S: Save | CTRL+Q: Quit"
		}
	} else {
		helpText = "TAB: Editor | CTRL+S: Save | CTRL+Q: Quit"
	}

	// Style the help bar
	helpStyle := lipgloss.NewStyle().
		Width(m.width - 6). // Account for window border and padding
		Background(lipgloss.Color(CTPMantle)).
		Foreground(lipgloss.Color(CTPSubtext0)).
		Align(lipgloss.Center).
		Padding(0, 1)

	return helpStyle.Render(helpText)
}

func (m Model) renderStatusBar() string {
	var status string

	// Colorful mode indicators
	var modeStr string
	if m.mode == ModeInsert {
		modeStr = lipgloss.NewStyle().Foreground(lipgloss.Color(CTPPink)).Bold(true).Render("INSERT")
	} else {
		modeStr = lipgloss.NewStyle().Foreground(lipgloss.Color(CTPBlue)).Bold(true).Render("NORMAL")
	}

	// File status with colors
	var fileStatus string
	if m.filename != "" {
		fileStatus = lipgloss.NewStyle().Foreground(lipgloss.Color(CTPGreen)).Render(m.filename)
		if !m.saved {
			fileStatus += lipgloss.NewStyle().Foreground(lipgloss.Color(CTPRed)).Render(" [modified]")
		}
	} else {
		fileStatus = lipgloss.NewStyle().Foreground(lipgloss.Color(CTPPeach)).Render("[New File]")
		if !m.saved {
			fileStatus += lipgloss.NewStyle().Foreground(lipgloss.Color(CTPRed)).Render(" [modified]")
		}
	}

	// Position with color
	position := lipgloss.NewStyle().Foreground(lipgloss.Color(CTPTeal)).Render(fmt.Sprintf("(%d,%d)", m.cursor.row+1, m.cursor.col+1))

	leftSide := fmt.Sprintf(" %s | %s", modeStr, fileStatus)
	rightSide := fmt.Sprintf("%s ", position)

	if m.statusMsg != "" {
		leftSide = fmt.Sprintf(" %s", m.statusMsg)
	}

	// Calculate spacing
	totalWidth := m.width - 6 // Account for window border and padding
	usedWidth := len(leftSide) + len(rightSide)
	spacing := totalWidth - usedWidth
	if spacing < 0 {
		spacing = 0
	}

	status = leftSide + strings.Repeat(" ", spacing) + rightSide

	style := lipgloss.NewStyle().
		Width(m.width - 6). // Account for window border and padding
		Background(lipgloss.Color(CTPSurface0)).
		Foreground(lipgloss.Color(CTPText))

	return style.Render(status)
}
