package main

import (
	"edigo/pkg/file"
	"edigo/pkg/input"
	"edigo/pkg/render"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"os"
)

func main() {
	app := tview.NewApplication()

	lines, err := file.Load("example.txt")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
		os.Exit(1)
	}

	cursorX, cursorY := 0, 0
	unsavedChanges := false

	// TextView
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	// Footer
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetText("ESC: Menü öffnen").
		SetTextAlign(tview.AlignCenter)

	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, true). // Editor
		AddItem(footer, 1, 1, false)   // Footer

	// Overlay
	menu := tview.NewModal().
		SetText("Menü:\n\n[1] Option 1\n[2] Option 2\n[ESC] Schließen").
		AddButtons([]string{"Option 1", "Option 2", "Schließen"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Schließen" {
				app.SetRoot(mainFlex, true)
			} else {
				// more menu
			}
		})

	render.Content(textView, lines, cursorX, cursorY, unsavedChanges, footer)

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			app.SetRoot(menu, true) // show menu
		default:
			lines, unsavedChanges = input.HandleKey(event, textView, lines, &cursorX, &cursorY, unsavedChanges)
			render.Content(textView, lines, cursorX, cursorY, unsavedChanges, footer)
		}
		return event
	})

	if err := app.SetRoot(mainFlex, true).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
