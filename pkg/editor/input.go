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
	case key.Matches(msg, key.NewBinding(key.WithKeys("left"))):
		ih.Editor.MoveCursorLeft()
	case key.Matches(msg, key.NewBinding(key.WithKeys("right"))):
		ih.Editor.MoveCursorRight()
	case key.Matches(msg, key.NewBinding(key.WithKeys("up"))):
		ih.Editor.MoveCursorUp()
	case key.Matches(msg, key.NewBinding(key.WithKeys("down"))):
		ih.Editor.MoveCursorDown()
	case key.Matches(msg, key.NewBinding(key.WithKeys("backspace", "ctrl+h"))):
		ih.Editor.DeleteCharacterBeforeCursor()
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		ih.Editor.InsertCharacter('\n')
	case key.Matches(msg, key.NewBinding(key.WithKeys("delete"))):
	default:
		if len(msg.String()) == 1 { // Only handle single characters
			ih.Editor.InsertCharacter(rune(msg.String()[0]))
		}
	}
}
