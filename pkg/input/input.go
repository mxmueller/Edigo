package input

import (
	"edigo/pkg/file"
	"edigo/pkg/notify"
	"edigo/pkg/render"
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
)

var unsavedChanges bool = false

func HandleKey(ev *tcell.EventKey, screen tcell.Screen, lines []string, cursorX *int, cursorY *int, style tcell.Style, tildeColor tcell.Color) []string {
	_, screenHeight := screen.Size()

	// handle keychange bindings
	// @Tim dont ask me whats happening here...
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		screen.Fini()
		os.Exit(0)
	case tcell.KeyCtrlS:
		// file save, this should be done in a seperate module in the future
		err := file.Save("example.txt", lines)
		if err != nil {
			screen.Fini()
			log.Fatalf("Error saving file: %v", err)
		}
		unsavedChanges = false
		notify.ClearMessage(screen)
	case tcell.KeyUp:
		if *cursorY > 0 {
			*cursorY--
			if *cursorX > len(lines[*cursorY]) {
				*cursorX = len(lines[*cursorY])
			}
		}
	case tcell.KeyDown:
		if *cursorY < len(lines)-1 {
			*cursorY++
			if *cursorX > len(lines[*cursorY]) {
				*cursorX = len(lines[*cursorY])
			}
		} else {
			lines = append(lines, "")
			*cursorY++
		}
	case tcell.KeyLeft:
		if *cursorX > 0 {
			*cursorX--
		} else if *cursorY > 0 {
			*cursorY--
			*cursorX = len(lines[*cursorY])
		}
	case tcell.KeyRight:
		if *cursorX < len(lines[*cursorY]) {
			*cursorX++
		} else if *cursorY < len(lines)-1 {
			*cursorY++
			*cursorX = 0
		}
	case tcell.KeyBackspace2, tcell.KeyBackspace:
		if *cursorX > 0 {
			line := lines[*cursorY]
			lines[*cursorY] = line[:*cursorX-1] + line[*cursorX:]
			*cursorX--
		} else if *cursorY > 0 {
			previousLine := lines[*cursorY-1]
			currentLine := lines[*cursorY]
			lines = append(lines[:*cursorY], lines[*cursorY+1:]...)
			*cursorY--
			*cursorX = len(previousLine)
			lines[*cursorY] = previousLine + currentLine
		}
	case tcell.KeyDelete:
		if *cursorX < len(lines[*cursorY]) {
			line := lines[*cursorY]
			lines[*cursorY] = line[:*cursorX] + line[*cursorX+1:]
		} else if *cursorY < len(lines)-1 {
			lines[*cursorY] += lines[*cursorY+1]
			lines = append(lines[:*cursorY+1], lines[*cursorY+2:]...)
		}
	case tcell.KeyRune:
		if *cursorY >= len(lines) {
			lines = append(lines, "")
		}
		if *cursorX <= len(lines[*cursorY]) {
			lines[*cursorY] = lines[*cursorY][:*cursorX] + string(ev.Rune()) + lines[*cursorY][*cursorX:]
			*cursorX++
		}
		unsavedChanges = true
		notify.ShowMessage(screen, "Press Ctrl+S to save", style)
	}

	if *cursorY >= len(lines) {
		*cursorY = len(lines) - 1
	}
	if *cursorX > len(lines[*cursorY]) {
		*cursorX = len(lines[*cursorY])
	}

	screen.Clear()
	screen.Sync()

	render.Content(screen, lines, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault), tildeColor)

	// @ SaveMessage
	if unsavedChanges {
		message := "Press Ctrl+S to save"
		for i, r := range message {
			screen.SetContent(i, screenHeight-1, r, nil, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDefault))
		}
	}

	screen.ShowCursor(*cursorX, *cursorY)
	screen.Sync()

	return lines
}
