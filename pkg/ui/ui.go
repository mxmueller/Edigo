package ui

import (
	"edigo/pkg/editor"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

type UIModel struct {
	Editor       *editor.Editor
	InputHandler *editor.InputHandler
	Viewport     viewport.Model
	QuitKey      key.Binding
	SaveKey      key.Binding
	FilePath     string
}

func NewUIModel(content string, filePath string) *UIModel {
	editorInstance := editor.NewEditor(content)
	vp := viewport.New(80, 24) // Default size; changes on window resize

	return &UIModel{
		Editor:       editorInstance,
		InputHandler: editor.NewInputHandler(editorInstance),
		Viewport:     vp,
		QuitKey: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		SaveKey: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		FilePath: filePath,
	}
}

func (m *UIModel) Init() tea.Cmd {
	m.Viewport.SetContent(m.renderContent())
	return tea.EnterAltScreen
}

func (m *UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.QuitKey) {
			return m, tea.Quit
		}
		if key.Matches(msg, m.SaveKey) {
			m.saveFile()
			return m, nil
		}
		m.InputHandler.HandleKeyMsg(msg) // Handle key input for navigation
		m.Viewport.SetContent(m.renderContent())

	case tea.WindowSizeMsg:
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height
		m.Viewport.SetContent(m.renderContent())
	}

	return m, nil
}

func (m *UIModel) View() string {
	return m.Viewport.View()
}

func (m *UIModel) renderContent() string {
	return fmt.Sprintf("File: %s\n\n%s", m.FilePath, m.Editor.RenderDocument())
}

func (m *UIModel) saveFile() {
	content := m.Editor.RenderDocumentWithoutLineNumbers()
	err := os.WriteFile(m.FilePath, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error saving file: %v\n", err)
	}
}
