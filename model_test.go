package main

import (
	"os"
	"testing"
	"time"
)

func TestNewModel(t *testing.T) {
	// Test with empty filename
	m := NewModel("")
	if len(m.content) != 1 || m.content[0] != "" {
		t.Errorf("Expected empty content, got %v", m.content)
	}
	if !m.saved {
		t.Errorf("Expected new empty model to be saved")
	}

	// Test with non-existent file
	m = NewModel("nonexistent.md")
	if m.filename != "nonexistent.md" {
		t.Errorf("Expected filename to be set")
	}
	if m.saved {
		t.Errorf("Expected new file model to be unsaved")
	}
}

func TestIsBinaryFile(t *testing.T) {
	// Test empty file
	if isBinaryFile([]byte{}) {
		t.Errorf("Empty file should not be binary")
	}

	// Test text file
	textData := []byte("Hello, world!\nThis is a text file.")
	if isBinaryFile(textData) {
		t.Errorf("Text file should not be binary")
	}

	// Test binary file (with null bytes)
	binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	if !isBinaryFile(binaryData) {
		t.Errorf("Binary file should be detected as binary")
	}

	// Test file with high ratio of non-printable characters
	nonPrintableData := make([]byte, 100)
	for i := 0; i < 50; i++ {
		nonPrintableData[i] = 1 // Non-printable
	}
	for i := 50; i < 100; i++ {
		nonPrintableData[i] = 65 // 'A'
	}
	if !isBinaryFile(nonPrintableData) {
		t.Errorf("File with high non-printable ratio should be binary")
	}
}

func TestSetStatusMsg(t *testing.T) {
	m := NewModel("")

	// Test normal message
	m.setStatusMsg("Test message", false)
	if m.statusMsg != "Test message" {
		t.Errorf("Expected status message to be set")
	}
	if time.Now().After(m.statusMsgTimeout) {
		t.Errorf("Status message timeout should be in the future")
	}

	// Test error message
	m.setStatusMsg("Error message", true)
	if m.statusMsg != "Error message" {
		t.Errorf("Expected error message to be set")
	}
}

func TestEnsureCursorBounds(t *testing.T) {
	m := NewModel("")
	m.content = []string{"Hello", "World", "Test"}

	// Test cursor beyond content
	m.cursor.row = 5
	m.cursor.col = 10
	m.ensureCursorBounds()

	if m.cursor.row != 2 {
		t.Errorf("Expected cursor row to be bounded to 2, got %d", m.cursor.row)
	}
	if m.cursor.col > len(m.content[m.cursor.row]) {
		t.Errorf("Expected cursor col to be bounded")
	}

	// Test negative cursor
	m.cursor.row = -1
	m.cursor.col = -1
	m.ensureCursorBounds()

	if m.cursor.row != 0 {
		t.Errorf("Expected cursor row to be bounded to 0, got %d", m.cursor.row)
	}
	if m.cursor.col != 0 {
		t.Errorf("Expected cursor col to be bounded to 0, got %d", m.cursor.col)
	}
}

func TestRebuildCodeBlocks(t *testing.T) {
	m := NewModel("")
	m.content = []string{
		"# Header",
		"```go",
		"func main() {",
		"    fmt.Println(\"Hello\")",
		"}",
		"```",
		"More text",
		"```python",
		"print('Hello')",
		"```",
	}
	m.codeBlocksDirty = true
	m.rebuildCodeBlocks()

	if len(m.codeBlocks) != 2 {
		t.Errorf("Expected 2 code blocks, got %d", len(m.codeBlocks))
	}

	// Check first code block
	if m.codeBlocks[0].start != 1 || m.codeBlocks[0].end != 5 {
		t.Errorf("First code block bounds incorrect: start=%d, end=%d",
			m.codeBlocks[0].start, m.codeBlocks[0].end)
	}
	if m.codeBlocks[0].lang != "go" {
		t.Errorf("Expected first code block language to be 'go', got '%s'", m.codeBlocks[0].lang)
	}

	// Check second code block
	if m.codeBlocks[1].start != 7 || m.codeBlocks[1].end != 9 {
		t.Errorf("Second code block bounds incorrect: start=%d, end=%d",
			m.codeBlocks[1].start, m.codeBlocks[1].end)
	}
	if m.codeBlocks[1].lang != "python" {
		t.Errorf("Expected second code block language to be 'python', got '%s'", m.codeBlocks[1].lang)
	}
}

func TestIsInCodeBlock(t *testing.T) {
	m := NewModel("")
	m.content = []string{
		"# Header",
		"```go",
		"func main() {",
		"}",
		"```",
		"More text",
	}
	m.codeBlocksDirty = true
	m.rebuildCodeBlocks()

	// Test line inside code block
	inBlock, lang := m.isInCodeBlock(2)
	if !inBlock {
		t.Errorf("Line 2 should be in code block")
	}
	if lang != "go" {
		t.Errorf("Expected language 'go', got '%s'", lang)
	}

	// Test line outside code block
	inBlock, _ = m.isInCodeBlock(0)
	if inBlock {
		t.Errorf("Line 0 should not be in code block")
	}

	// Test fence lines
	inBlock, _ = m.isInCodeBlock(1) // Opening fence
	if inBlock {
		t.Errorf("Fence line should not be considered inside code block")
	}
}

func TestInsertCursor(t *testing.T) {
	m := NewModel("")

	// Test cursor at end of line
	result := m.insertCursor("Hello", "Hello", 5)
	if result != "Hello█" {
		t.Errorf("Expected 'Hello█', got '%s'", result)
	}

	// Test cursor in middle of line
	result = m.insertCursor("Hello", "Hello", 2)
	if result != "He█llo" {
		t.Errorf("Expected 'He█llo', got '%s'", result)
	}

	// Test cursor at beginning
	result = m.insertCursor("Hello", "Hello", 0)
	if result != "█Hello" {
		t.Errorf("Expected '█Hello', got '%s'", result)
	}
}

func TestWordMovement(t *testing.T) {
	m := NewModel("")
	m.content = []string{"Hello world test", "Another line"}
	m.cursor = Position{row: 0, col: 0}

	// Test next word
	pos := m.nextWord()
	if pos.row != 0 || pos.col != 6 {
		t.Errorf("Expected next word at (0,6), got (%d,%d)", pos.row, pos.col)
	}

	// Test previous word
	m.cursor = Position{row: 0, col: 6}
	pos = m.prevWord()
	if pos.row != 0 || pos.col != 0 {
		t.Errorf("Expected previous word at (0,0), got (%d,%d)", pos.row, pos.col)
	}

	// Test end of word
	m.cursor = Position{row: 0, col: 0}
	pos = m.endOfWord()
	if pos.row != 0 || pos.col != 4 {
		t.Errorf("Expected end of word at (0,4), got (%d,%d)", pos.row, pos.col)
	}
}

func TestSaveFile(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := "test_save.md"
	defer os.Remove(tmpFile)

	m := NewModel(tmpFile)
	m.content = []string{"# Test", "This is a test"}
	m.saved = false

	// Test save
	newModel, _ := m.saveFile()
	m = newModel.(Model)

	if !m.saved {
		t.Errorf("Expected file to be marked as saved")
	}

	// Verify file was actually written
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Errorf("Failed to read saved file: %v", err)
	}

	expected := "# Test\nThis is a test"
	if string(data) != expected {
		t.Errorf("Expected file content '%s', got '%s'", expected, string(data))
	}

	// Check backup was created
	backupFile := tmpFile + ".bak"
	defer os.Remove(backupFile)

	// Save again to test backup creation
	m.content = []string{"# Modified", "Content changed"}
	m.saved = false
	m.saveFile()

	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Errorf("Backup file should have been created")
	}
}

func TestMinFunction(t *testing.T) {
	if min(5, 3) != 3 {
		t.Errorf("min(5, 3) should be 3")
	}
	if min(1, 10) != 1 {
		t.Errorf("min(1, 10) should be 1")
	}
	if min(5, 5) != 5 {
		t.Errorf("min(5, 5) should be 5")
	}
}

func TestIsWhitespace(t *testing.T) {
	if !isWhitespace(' ') {
		t.Errorf("Space should be whitespace")
	}
	if !isWhitespace('\t') {
		t.Errorf("Tab should be whitespace")
	}
	if !isWhitespace('\n') {
		t.Errorf("Newline should be whitespace")
	}
	if isWhitespace('a') {
		t.Errorf("Letter 'a' should not be whitespace")
	}
}
