package editor

import (
	"strings"
)

type Editor struct {
	Content []string
	CursorX int
	CursorY int
}

// NewEditor initializes a new text editor with the given content.
func NewEditor(content string) *Editor {
	lines := strings.Split(content, "\n")
	return &Editor{
		Content: lines,
		CursorX: 0,
		CursorY: 0,
	}
}

// InsertCharacter inserts a character at the current cursor position.
func (e *Editor) InsertCharacter(ch string) {
	line := e.Content[e.CursorY]
	e.Content[e.CursorY] = line[:e.CursorX] + ch + line[e.CursorX:]
	e.CursorX++
}

// DeleteCharacterBeforeCursor deletes the character before the cursor.
func (e *Editor) DeleteCharacterBeforeCursor() {
	if e.CursorX > 0 {
		line := e.Content[e.CursorY]
		e.Content[e.CursorY] = line[:e.CursorX-1] + line[e.CursorX:]
		e.CursorX--
	}
}

// DeleteCharacterAtCursor deletes the character under the cursor.
func (e *Editor) DeleteCharacterAtCursor() {
	line := e.Content[e.CursorY]
	if e.CursorX < len(line) {
		e.Content[e.CursorY] = line[:e.CursorX] + line[e.CursorX+1:]
	}
}

// JoinWithPreviousLine joins the current line with the previous one.
func (e *Editor) JoinWithPreviousLine() {
	if e.CursorY > 0 {
		e.Content[e.CursorY-1] += e.Content[e.CursorY]
		e.Content = append(e.Content[:e.CursorY], e.Content[e.CursorY+1:]...)
		e.CursorY--
		e.CursorX = len(e.Content[e.CursorY])
	}
}

// getLine returns the content of a specific line by index.
func (e *Editor) getLine(lineIndex int) string {
	if lineIndex >= 0 && lineIndex < len(e.Content) {
		return e.Content[lineIndex]
	}
	return ""
}

// getLines returns all lines of the document as a slice of strings.
func (e *Editor) getLines() []string {
	return e.Content
}

// RenderDocument renders the document with line numbers and the cursor.
func (e *Editor) RenderDocument() string {
	var result strings.Builder
	for i, line := range e.Content {
		if i == e.CursorY {
			if e.CursorX < len(line) {
				result.WriteString(line[:e.CursorX] + "█" + line[e.CursorX:] + "\n")
			} else {
				result.WriteString(line + "█\n")
			}
		} else {
			result.WriteString(line + "\n")
		}
	}
	return result.String()
}

// RenderDocumentWithoutLineNumbers renders the document without line numbers.
func (e *Editor) RenderDocumentWithoutLineNumbers() string {
	return strings.Join(e.Content, "\n")
}
