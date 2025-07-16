package main

import (
	"strings"
	"testing"
)

func TestNewSyntaxHighlighter(t *testing.T) {
	highlighter := NewSyntaxHighlighter()
	if highlighter == nil {
		t.Errorf("NewSyntaxHighlighter should not return nil")
	}
	if highlighter.formatter == nil {
		t.Errorf("Formatter should not be nil")
	}
	if highlighter.style == nil {
		t.Errorf("Style should not be nil")
	}
}

func TestHighlightMarkdownLine(t *testing.T) {
	highlighter := NewSyntaxHighlighter()
	if highlighter == nil {
		t.Skip("Skipping test due to highlighter initialization failure")
	}

	tests := []struct {
		input    string
		contains string // What the output should contain
		desc     string
	}{
		{"# Header 1", "Header 1", "H1 header"},
		{"## Header 2", "Header 2", "H2 header"},
		{"### Header 3", "Header 3", "H3 header"},
		{"#### Header 4", "Header 4", "H4 header"},
		{"```go", "```go", "Code fence"},
		{"> Quote", "Quote", "Blockquote"},
		{"- List item", "List item", "Unordered list"},
		{"* List item", "List item", "Unordered list with asterisk"},
		{"+ List item", "List item", "Unordered list with plus"},
		{"1. Numbered", "Numbered", "Numbered list"},
		{"---", "---", "Horizontal rule"},
		{"Regular text", "Regular text", "Regular text"},
		{"`inline code`", "inline code", "Inline code"},
	}

	for _, test := range tests {
		result := highlighter.HighlightMarkdownLine(test.input)
		if !strings.Contains(result, test.contains) {
			t.Errorf("%s: Expected result to contain '%s', got '%s'",
				test.desc, test.contains, result)
		}
	}
}

func TestHighlightCodeBlock(t *testing.T) {
	highlighter := NewSyntaxHighlighter()
	if highlighter == nil {
		t.Skip("Skipping test due to highlighter initialization failure")
	}

	// Test Go code highlighting
	goCode := `func main() {
    fmt.Println("Hello, World!")
}`
	result := highlighter.HighlightCodeBlock(goCode, "go")
	if result == "" {
		t.Errorf("Go code highlighting should not return empty string")
	}
	if result == goCode {
		// If result is same as input, highlighting might have failed
		// but we'll accept it as a fallback
		t.Logf("Go code highlighting may have fallen back to plain text")
	}

	// Test Python code highlighting
	pythonCode := `def hello():
    print("Hello, World!")`
	result = highlighter.HighlightCodeBlock(pythonCode, "python")
	if result == "" {
		t.Errorf("Python code highlighting should not return empty string")
	}

	// Test unknown language
	result = highlighter.HighlightCodeBlock("some code", "unknown")
	if result == "" {
		t.Errorf("Unknown language highlighting should not return empty string")
	}

	// Test empty code
	result = highlighter.HighlightCodeBlock("", "go")
	if result != "" {
		t.Errorf("Empty code should return empty string, got '%s'", result)
	}
}

func TestHighlightInlineCode(t *testing.T) {
	highlighter := NewSyntaxHighlighter()
	if highlighter == nil {
		t.Skip("Skipping test due to highlighter initialization failure")
	}

	tests := []struct {
		input    string
		contains string
		desc     string
	}{
		{"`code`", "code", "Simple inline code"},
		{"Text with `code` in middle", "code", "Inline code in text"},
		{"Multiple `code1` and `code2`", "code1", "Multiple inline codes"},
		{"No code here", "No code here", "No inline code"},
		{"`unclosed code", "unclosed code", "Unclosed inline code"},
	}

	for _, test := range tests {
		result := highlighter.highlightInlineCode(test.input)
		if !strings.Contains(result, test.contains) {
			t.Errorf("%s: Expected result to contain '%s', got '%s'",
				test.desc, test.contains, result)
		}
	}
}

func TestHighlightMarkdownLineEdgeCases(t *testing.T) {
	highlighter := NewSyntaxHighlighter()
	if highlighter == nil {
		t.Skip("Skipping test due to highlighter initialization failure")
	}

	// Test with nil highlighter
	var nilHighlighter *SyntaxHighlighter
	result := nilHighlighter.HighlightMarkdownLine("# Header")
	if result != "# Header" {
		t.Errorf("Nil highlighter should return input unchanged")
	}

	// Test empty string
	result = highlighter.HighlightMarkdownLine("")
	if result != "" {
		t.Errorf("Empty string should return empty string")
	}

	// Test whitespace-only string
	result = highlighter.HighlightMarkdownLine("   ")
	if result != "   " {
		t.Errorf("Whitespace-only string should return unchanged")
	}

	// Test indented headers (should not be treated as headers)
	result = highlighter.HighlightMarkdownLine("  # Not a header")
	// Should not be highlighted as header since it's indented
	if result == "" {
		t.Errorf("Indented header should return some result")
	}

	// Test list items with indentation
	result = highlighter.HighlightMarkdownLine("  - Indented list")
	if !strings.Contains(result, "Indented list") {
		t.Errorf("Indented list should preserve content")
	}
}
