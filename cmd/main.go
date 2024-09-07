package main

import (
	"edigo/pkg/ui"
	"fmt"
	"io/ioutil"
	"log"
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
		log.Fatalf("Error reading file: %v\n", err)
	}

	model := ui.NewUIModel(string(content), filePath)

	go func() {
		model.Editor.Network.ListenForBroadcasts()
	}()

	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		log.Fatalf("Error starting the program: %v\n", err)
	}
}
