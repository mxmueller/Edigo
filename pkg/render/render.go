package render

import (
	"github.com/gdamore/tcell/v2"
)

func Content(screen tcell.Screen, lines []string, style tcell.Style, tildeColor tcell.Color) {
	screen.Clear()
	width, height := screen.Size()

	for y := 0; y < height && y < len(lines); y++ {
		line := lines[y]
		for x, ch := range line {
			if x < width {
				screen.SetContent(x, y, ch, nil, style)
			}
		}
	}

	// Lines, that are not in the file are displayed with a tilde until they are edited
	for y := len(lines); y < height; y++ {
		screen.SetContent(0, y, '~', nil, style.Foreground(tildeColor))
	}
	screen.Show()
}
