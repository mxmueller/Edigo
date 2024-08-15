package input

import (
	"edigo/pkg/file"
	"edigo/pkg/notify"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"os"
)

func HandleKey(ev *tcell.EventKey, screen *tview.TextView, lines []string, cursorX *int, cursorY *int, unsavedChanges bool) ([]string, bool) {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		os.Exit(0)
	case tcell.KeyCtrlS:
		if unsavedChanges {
			err := file.Save("example.txt", lines)
			if err != nil {
				log.Fatalf("Error saving file: %v", err)
			}
			unsavedChanges = false
			notify.ClearMessage(screen)
		}
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
			unsavedChanges = true
		} else if *cursorY > 0 {
			previousLine := lines[*cursorY-1]
			currentLine := lines[*cursorY]
			lines = append(lines[:*cursorY], lines[*cursorY+1:]...)
			*cursorY--
			*cursorX = len(previousLine)
			lines[*cursorY] = previousLine + currentLine
			unsavedChanges = true
		}
	case tcell.KeyDelete:
		if *cursorX < len(lines[*cursorY]) {
			line := lines[*cursorY]
			lines[*cursorY] = line[:*cursorX] + line[*cursorX+1:]
			unsavedChanges = true
		} else if *cursorY < len(lines)-1 {
			lines[*cursorY] += lines[*cursorY+1]
			lines = append(lines[:*cursorY+1], lines[*cursorY+2:]...)
			unsavedChanges = true
		}
	case tcell.KeyRune:
		if *cursorY >= len(lines) {
			lines = append(lines, "")
		}
		if *cursorX <= len(lines[*cursorY]) {
			lines[*cursorY] = lines[*cursorY][:*cursorX] + string(ev.Rune()) + lines[*cursorY][*cursorX:]
			*cursorX++
			unsavedChanges = true
		}
	}

	if *cursorY >= len(lines) {
		*cursorY = len(lines) - 1
	}
	if *cursorX > len(lines[*cursorY]) {
		*cursorX = len(lines[*cursorY])
	}

	return lines, unsavedChanges
}
