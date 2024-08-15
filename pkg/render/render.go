package render

import (
	"github.com/rivo/tview"
)

func Content(screen *tview.TextView, lines []string, cursorX, cursorY int, unsavedChanges bool, footer *tview.TextView) {
	screen.Clear()
	_, _, height, _ := screen.GetInnerRect()

	for y := 0; y < height; y++ {
		if y < len(lines) {
			line := lines[y]
			if y == cursorY {
				// Cursor
				beforeCursor := line[:cursorX]
				cursorChar := " "
				if cursorX < len(line) {
					cursorChar = string(line[cursorX])
				}
				afterCursor := ""
				if cursorX+1 < len(line) {
					afterCursor = line[cursorX+1:]
				}
				screen.Write([]byte(beforeCursor))
				screen.Write([]byte("[black:white]" + cursorChar + "[white:black]"))
				screen.Write([]byte(afterCursor + "\n"))
			} else {
				screen.Write([]byte(line + "\n"))
			}
		}
	}

	if unsavedChanges {
		footer.SetText("ESC: Menü öffnen | Ctrl+S: Speichern")
	} else {
		footer.SetText("ESC: Menü öffnen")
	}
}
