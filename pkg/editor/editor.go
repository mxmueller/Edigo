package editor

import (
	"edigo/pkg/crdt"
	"edigo/pkg/network"
	"fmt"
	"strings"
)

type Editor struct {
	RGA *crdt.RGA
}

func NewEditor(content string, siteID string) *Editor {
	rga := crdt.NewRGA(siteID)
	for _, char := range content {
		rga.LocalInsert(char)
	}
	return &Editor{
		RGA: rga,
	}
}

func (e *Editor) InsertCharacter(ch rune) {
	e.RGA.LocalInsert(ch)
}

func (e *Editor) DeleteCharacterBeforeCursor() {
	e.RGA.LocalDelete()
}

func (e *Editor) MoveCursorLeft() {
	e.RGA.MoveCursorLeft()
}

func (e *Editor) MoveCursorRight() {
	e.RGA.MoveCursorRight()
}

func (e *Editor) MoveCursorUp() {
	e.RGA.MoveCursorUp()
}

func (e *Editor) MoveCursorDown() {
	e.RGA.MoveCursorDown()
}

func (e *Editor) RenderDocument() string {
	var result strings.Builder
	content := e.RGA.GetText()

	for i, ch := range content {
		if i == e.RGA.CursorPosition {
			result.WriteRune('█') // Cursor
		}
		result.WriteRune(ch)
	}

	if e.RGA.CursorPosition == len(content) {
		result.WriteRune('█') // Cursor End
	}

	return result.String()
}

func (e *Editor) GetLineNumbers() string {
	var lineNumbers strings.Builder
	lineNumber := 1
	lineNumbers.WriteString(fmt.Sprintf("%d\n", lineNumber))

	for _, char := range e.RGA.GetText() {
		if char == '\n' {
			lineNumber++
			lineNumbers.WriteString(fmt.Sprintf("%d\n", lineNumber))
		}
	}

	return lineNumbers.String()
}

func (e *Editor) RenderDocumentWithoutLineNumbers() string {
	return e.RGA.GetText()
}
