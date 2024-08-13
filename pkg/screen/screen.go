package screen

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
)

func Initialize() tcell.Screen {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Error creating screen: %v", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("Error initializing screen: %v", err)
		os.Exit(1)
	}
	return screen
}
