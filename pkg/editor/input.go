package editor

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type InputHandler struct {
	Editor *Editor
}

func NewInputHandler(editor *Editor) *InputHandler {
	return &InputHandler{
		Editor: editor,
	}
}

func (ih *InputHandler) HandleKeyMsg(msg tea.KeyMsg) {
	switch {
	// Move cursor left
	case key.Matches(msg, key.NewBinding(key.WithKeys("left"))):
		if ih.Editor.CursorX > 0 {
			ih.Editor.CursorX--
		} else if ih.Editor.CursorY > 0 {
			ih.Editor.CursorY--
			ih.Editor.CursorX = len(ih.Editor.getLine(ih.Editor.CursorY))
		}
	// Move cursor right
	case key.Matches(msg, key.NewBinding(key.WithKeys("right"))):
		if ih.Editor.CursorX < len(ih.Editor.getLine(ih.Editor.CursorY)) {
			ih.Editor.CursorX++
		} else if ih.Editor.CursorY < len(ih.Editor.getLines())-1 {
			ih.Editor.CursorY++
			ih.Editor.CursorX = 0
		}
	// Move cursor up
	case key.Matches(msg, key.NewBinding(key.WithKeys("up"))):
		if ih.Editor.CursorY > 0 {
			ih.Editor.CursorY--
			ih.Editor.CursorX = min(ih.Editor.CursorX, len(ih.Editor.getLine(ih.Editor.CursorY)))
		}
	// Move cursor down
	case key.Matches(msg, key.NewBinding(key.WithKeys("down"))):
		if ih.Editor.CursorY < len(ih.Editor.getLines())-1 {
			ih.Editor.CursorY++
			ih.Editor.CursorX = min(ih.Editor.CursorX, len(ih.Editor.getLine(ih.Editor.CursorY)))
		}
	// Handle backspace (delete character before cursor)
	case key.Matches(msg, key.NewBinding(key.WithKeys("backspace", "ctrl+h"))):
		if ih.Editor.CursorX > 0 {
			ih.Editor.DeleteCharacterBeforeCursor()
		} else if ih.Editor.CursorY > 0 {
			prevLineLen := len(ih.Editor.getLine(ih.Editor.CursorY - 1))
			ih.Editor.JoinWithPreviousLine()
			ih.Editor.CursorY--
			ih.Editor.CursorX = prevLineLen
		}
	// Handle delete (delete character under the cursor)
	case key.Matches(msg, key.NewBinding(key.WithKeys("delete"))):
		ih.Editor.DeleteCharacterAtCursor()
	// Handle text input
	default:
		if len(msg.String()) == 1 { // Only handle single characters
			ih.Editor.InsertCharacter(msg.String())
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
