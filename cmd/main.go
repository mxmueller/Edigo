package main

import (
	"edigo/pkg/file"
	"edigo/pkg/input"
	"edigo/pkg/render"
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to start screen: %v", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("Failed to initialize screen: %v", err)
		os.Exit(1)
	}
	defer screen.Fini()

	// important for gray background bug
	screen.Clear()
	screen.Sync()

	defStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)
	tildeColor := tcell.ColorRed
	cursorX, cursorY := 0, 0

	lines, err := file.Load("example.txt")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
		os.Exit(1)
	}

	render.Content(screen, lines, defStyle, tildeColor)
	screen.ShowCursor(cursorX, cursorY)
	screen.Sync()

	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			lines = input.HandleKey(ev, screen, lines, &cursorX, &cursorY, defStyle, tildeColor)
		case *tcell.EventResize:
			screen.Clear()
			render.Content(screen, lines, defStyle, tildeColor)
			screen.Sync()
		}
	}
}
