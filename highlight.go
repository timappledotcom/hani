package main

import (
	"strings"
	"github.com/charmbracelet/lipgloss"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/alecthomas/chroma/v2/formatters"
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
		formatter = formatters.Fallback
	}
	
	// Use a dark theme that works well in terminals
	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}
	
	return &SyntaxHighlighter{
		formatter: formatter,
		style:     style,
	}
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
	// Headers
	if strings.HasPrefix(line, "# ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")).  // Blue
			Bold(true).
			Render(line)
	}
	if strings.HasPrefix(line, "## ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).  // Cyan
			Bold(true).
			Render(line)
	}
	if strings.HasPrefix(line, "### ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).  // Yellow
			Bold(true).
			Render(line)
	}
	
	// Code blocks
	if strings.HasPrefix(line, "```") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).  // Gray
			Render(line)
	}
	
	// List items
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).  // Magenta
			Render("• ") + line[2:]
	}
	
	// Default - just return the line as is
	return line
}
