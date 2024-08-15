package notify

import (
	"github.com/rivo/tview"
)

func ShowMessage(screen *tview.TextView, message string) {
	screen.Clear()
	screen.SetText(message)
}

func ClearMessage(screen *tview.TextView) {
	screen.Clear()
}
