package main

import (
	"edigo/pkg/ui"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io/ioutil"
	"os"
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

	model := ui.NewUIModel(string(content), filePath) // Pass file content and path to the model
	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		panic(err)
	}
}
