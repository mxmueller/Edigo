package notify

import (
	"github.com/gdamore/tcell/v2"
)

func ShowMessage(screen tcell.Screen, message string, style tcell.Style) {
	width, height := screen.Size()

	xPos := (width - len(message)) / 2
	yPos := height - 1

	// show message
	for i, r := range message {
		screen.SetContent(xPos+i, yPos, r, nil, style)
	}
	screen.Show()
}

func ClearMessage(screen tcell.Screen) {
	width, height := screen.Size()

	yPos := height - 1
	for x := 0; x < width; x++ {
		screen.SetContent(x, yPos, ' ', nil, tcell.StyleDefault)
	}
	screen.Show()
}
