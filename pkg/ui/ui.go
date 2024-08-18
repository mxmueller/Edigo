package ui

import (
	"edigo/pkg/editor"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"time"
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
	siteID := generateSiteID()
	editorInstance := editor.NewEditor(content, siteID)
	vp := viewport.New(80, 24)

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
		m.InputHandler.HandleKeyMsg(msg)
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
	// Zeilennummern und Dokumentinhalt separat rendern
	lineNumbers := m.Editor.GetLineNumbers()
	document := m.Editor.RenderDocument()

	// Spaltenbreite für Zeilennummern festlegen
	lineNumberWidth := len(fmt.Sprintf("%d", strings.Count(document, "\n")+1)) + 2

	// Baue die Ausgabe mit zwei Spalten: Zeilennummern und Dokumentinhalt
	var output strings.Builder

	lines := strings.Split(document, "\n")
	numberLines := strings.Split(lineNumbers, "\n")

	// Bestimme die Höhe des Viewports für die Darstellung der Tilden
	totalLines := m.Viewport.Height - 2 // Abzug für Dateiname und Leerzeile

	for i := 0; i < totalLines; i++ {
		if i < len(lines) {
			// Zeige die Zeilennummer und den entsprechenden Text an
			if i < len(numberLines) {
				output.WriteString(fmt.Sprintf("%-*s", lineNumberWidth, numberLines[i]))
			} else {
				output.WriteString(fmt.Sprintf("%-*s", lineNumberWidth, ""))
			}
			output.WriteString(lines[i])
		} else {
			// Zeige eine Tilde in der Zeilennummernspalte an
			output.WriteString(fmt.Sprintf("%-*s", lineNumberWidth, "~"))
		}
		output.WriteString("\n")
	}

	return fmt.Sprintf("File: %s\n\n%s", m.FilePath, output.String())
}

func (m *UIModel) saveFile() {
	content := m.Editor.RenderDocumentWithoutLineNumbers()
	err := os.WriteFile(m.FilePath, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error saving file: %v\n", err)
	}
}

func generateSiteID() string {
	return fmt.Sprintf("site-%d", time.Now().UnixNano())
}
