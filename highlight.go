package main

import (
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

// SyntaxHighlighter handles syntax highlighting for code blocks and markdown
type SyntaxHighlighter struct {
	formatter chroma.Formatter
	style     *chroma.Style
}

// NewSyntaxHighlighter creates a new syntax highlighter
func NewSyntaxHighlighter() *SyntaxHighlighter {
	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		formatter = formatters.Fallback
	}
	
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	return &SyntaxHighlighter{
		formatter: formatter,
		style:     style,
	}
}

// HighlightMarkdown applies syntax highlighting to markdown text
func (sh *SyntaxHighlighter) HighlightMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, len(lines))
	
	inCodeBlock := false
	codeBlockLang := ""
	codeBlockLines := []string{}
	
	for i, line := range lines {
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				// Start of code block
				inCodeBlock = true
				codeBlockLang = strings.TrimSpace(strings.TrimPrefix(line, "```"))
				codeBlockLines = []string{}
				result[i] = sh.styleCodeBlockMarker(line)
			} else {
				// End of code block
				inCodeBlock = false
				
				// Highlight the code block
				if len(codeBlockLines) > 0 {
					highlightedCode := sh.highlightCodeBlock(strings.Join(codeBlockLines, "\n"), codeBlockLang)
					highlightedLines := strings.Split(highlightedCode, "\n")
					
					// Insert highlighted lines
					for j, highlightedLine := range highlightedLines {
						if i-len(codeBlockLines)+j >= 0 && i-len(codeBlockLines)+j < len(result) {
							result[i-len(codeBlockLines)+j] = "  " + highlightedLine
						}
					}
				}
				
				result[i] = sh.styleCodeBlockMarker(line)
				codeBlockLines = []string{}
			}
		} else if inCodeBlock {
			// Inside code block, collect lines for highlighting
			codeBlockLines = append(codeBlockLines, line)
			result[i] = line // Will be replaced later
		} else {
			// Regular markdown line
			result[i] = sh.highlightMarkdownLine(line)
		}
	}
	
	return strings.Join(result, "\n")
}

// highlightCodeBlock highlights a code block with the specified language
func (sh *SyntaxHighlighter) highlightCodeBlock(code, lang string) string {
	if lang == "" {
		lang = "text"
	}
	
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}
	
	var buf strings.Builder
	err = sh.formatter.Format(&buf, sh.style, iterator)
	if err != nil {
		return code
	}
	
	return buf.String()
}

// highlightMarkdownLine highlights a single markdown line
func (sh *SyntaxHighlighter) highlightMarkdownLine(line string) string {
	// Headers
	if strings.HasPrefix(line, "# ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(CTPBlue)).
			Bold(true).
			Render(line)
	}
	if strings.HasPrefix(line, "## ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(CTPGreen)).
			Bold(true).
			Render(line)
	}
	if strings.HasPrefix(line, "### ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(CTPYellow)).
			Bold(true).
			Render(line)
	}
	if strings.HasPrefix(line, "#### ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(CTPPeach)).
			Bold(true).
			Render(line)
	}
	
	// Bold text
	boldRegex := regexp.MustCompile(`\*\*(.*?)\*\*`)
	line = boldRegex.ReplaceAllStringFunc(line, func(match string) string {
		content := strings.Trim(match, "*")
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(CTPText)).
			Bold(true).
			Render(content)
	})
	
	// Italic text
	italicRegex := regexp.MustCompile(`\*(.*?)\*`)
	line = italicRegex.ReplaceAllStringFunc(line, func(match string) string {
		content := strings.Trim(match, "*")
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(CTPText)).
			Italic(true).
			Render(content)
	})
	
	// Inline code
	codeRegex := regexp.MustCompile("`([^`]+)`")
	line = codeRegex.ReplaceAllStringFunc(line, func(match string) string {
		content := strings.Trim(match, "`")
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(CTPRed)).
			Background(lipgloss.Color(CTPSurface0)).
			Render(content)
	})
	
	// Links
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	line = linkRegex.ReplaceAllStringFunc(line, func(match string) string {
		parts := linkRegex.FindStringSubmatch(match)
		if len(parts) == 3 {
			linkText := parts[1]
			linkURL := parts[2]
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(CTPBlue)).
				Underline(true).
				Render(linkText) + lipgloss.NewStyle().
				Foreground(lipgloss.Color(CTPOverlay0)).
				Render(" ("+linkURL+")")
		}
		return match
	})
	
	// List items
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(CTPMauve)).
			Render("• ") + lipgloss.NewStyle().
			Foreground(lipgloss.Color(CTPText)).
			Render(line[2:])
	}
	
	// Numbered list items
	numberedListRegex := regexp.MustCompile(`^(\d+)\. (.*)$`)
	if numberedListRegex.MatchString(line) {
		parts := numberedListRegex.FindStringSubmatch(line)
		if len(parts) == 3 {
			number := parts[1]
			content := parts[2]
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(CTPMauve)).
				Render(number+". ") + lipgloss.NewStyle().
				Foreground(lipgloss.Color(CTPText)).
				Render(content)
		}
	}
	
	// Blockquotes
	if strings.HasPrefix(line, "> ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(CTPOverlay0)).
			Italic(true).
			Render(line)
	}
	
	// Default text color
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CTPText)).
		Render(line)
}

// styleCodeBlockMarker styles the ``` markers
func (sh *SyntaxHighlighter) styleCodeBlockMarker(line string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CTPOverlay0)).
		Render(line)
}
