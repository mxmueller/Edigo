package main

import (
	"edigo/pkg/network"
	"edigo/pkg/ui"
	"fmt"
	"io/ioutil"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please provide the file path as an argument.")
		os.Exit(1)
	}

	filePath := os.Args[1]
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

    go network.ListenForBroadcasts()
    go network.BroadcastSession()

	model := ui.NewUIModel(string(content), filePath) // Pass file content and path to the model
	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		panic(err)
	}
}
