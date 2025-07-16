package main

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// SyntaxHighlighter handles syntax highlighting using Chroma
type SyntaxHighlighter struct {
	formatter chroma.Formatter
	style     *chroma.Style
}

// NewSyntaxHighlighter creates a new syntax highlighter
func NewSyntaxHighlighter() *SyntaxHighlighter {
	// Use the terminal256 formatter which works well with terminals
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Get("terminal")
		if formatter == nil {
			formatter = formatters.Fallback
		}
	}

	// Use a dark theme that works well in terminals
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Get("github-dark")
		if style == nil {
			style = styles.Fallback
		}
	}

	highlighter := &SyntaxHighlighter{
		formatter: formatter,
		style:     style,
	}

	// Test the highlighter to ensure it works
	if highlighter == nil || highlighter.formatter == nil || highlighter.style == nil {
		return nil
	}

	return highlighter
}

// HighlightCodeBlock highlights a code block using Chroma
func (sh *SyntaxHighlighter) HighlightCodeBlock(code, lang string) string {
	// Handle empty code or language
	if code == "" {
		return code
	}

	// Get the lexer for the language
	lexer := lexers.Get(lang)
	if lexer == nil {
		// Try to guess the lexer from the content
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		// Fall back to plain text
		lexer = lexers.Get("text")
	}

	// Ensure lexer is configured
	lexer = chroma.Coalesce(lexer)

	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		// Fall back to simple green coloring
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Render(code)
	}

	// Format the tokens
	var result strings.Builder
	err = sh.formatter.Format(&result, sh.style, iterator)
	if err != nil {
		// Fall back to simple green coloring
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Render(code)
	}

	return result.String()
}

// HighlightMarkdownLine highlights a single markdown line with minimal styling
func (sh *SyntaxHighlighter) HighlightMarkdownLine(line string) string {
	if sh == nil {
		return line
	}

	trimmed := strings.TrimSpace(line)

	// Headers (with proper hierarchy)
	if strings.HasPrefix(line, "#### ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")). // Green
			Bold(true).
			Render(line)
	}
	if strings.HasPrefix(line, "### ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")). // Yellow
			Bold(true).
			Render(line)
	}
	if strings.HasPrefix(line, "## ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")). // Cyan
			Bold(true).
			Render(line)
	}
	if strings.HasPrefix(line, "# ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")). // Blue
			Bold(true).
			Render(line)
	}

	// Code blocks
	if strings.HasPrefix(trimmed, "```") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Render(line)
	}

	// Blockquotes
	if strings.HasPrefix(trimmed, "> ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")). // Light gray
			Italic(true).
			Render(line)
	}

	// List items (unordered)
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
		prefix := strings.Repeat(" ", len(line)-len(trimmed))
		return prefix + lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")). // Magenta
			Render("• ") + trimmed[2:]
	}

	// Numbered lists
	if len(trimmed) > 2 && trimmed[1] == '.' && trimmed[0] >= '0' && trimmed[0] <= '9' {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")). // Magenta
			Render(line)
	}

	// Horizontal rules
	if trimmed == "---" || trimmed == "***" || strings.HasPrefix(trimmed, "---") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Render(line)
	}

	// Inline code (simple detection)
	if strings.Contains(line, "`") && strings.Count(line, "`") >= 2 {
		return sh.highlightInlineCode(line)
	}

	// Default - just return the line as is
	return line
}

// highlightInlineCode highlights inline code snippets
func (sh *SyntaxHighlighter) highlightInlineCode(line string) string {
	result := ""
	inCode := false
	codeStart := 0

	for i, char := range line {
		if char == '`' {
			if inCode {
				// End of code block
				codeText := line[codeStart:i]
				styledCode := lipgloss.NewStyle().
					Foreground(lipgloss.Color("2")). // Green
					Background(lipgloss.Color("0")). // Black background
					Render(codeText)
				result += styledCode + "`"
				inCode = false
			} else {
				// Start of code block
				result += "`"
				codeStart = i + 1
				inCode = true
			}
		} else if !inCode {
			result += string(char)
		}
	}

	// Handle unclosed code block
	if inCode {
		result += line[codeStart:]
	}

	return result
}
