package ui

import (
	"edigo/pkg/editor"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"time"
)

type UIModel struct {
	Editor         *editor.Editor
	InputHandler   *editor.InputHandler
	Viewport       viewport.Model
	SaveKey        key.Binding
	MenuKey        key.Binding
	FilePath       string
	Menu           MenuModel
	ShowMenu       bool
	UnsavedChanges bool
}

func NewUIModel(content string, filePath string) *UIModel {
	siteID := generateSiteID()
	editorInstance := editor.NewEditor(content, siteID)
	vp := viewport.New(80, 24)

	return &UIModel{
		Editor:       editorInstance,
		InputHandler: editor.NewInputHandler(editorInstance),
		Viewport:     vp,
		SaveKey: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		MenuKey: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "menu"),
		),
		FilePath:       filePath,
		Menu:           NewMenuModel(),
		ShowMenu:       false,
		UnsavedChanges: false,
	}
}

func (m *UIModel) Init() tea.Cmd {
	m.Viewport.SetContent(m.renderContent())
	return tea.EnterAltScreen
}

func (m *UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.ShowMenu {
		return m.updateMenu(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.SaveKey):
			m.saveFile()
			return m, nil
		case key.Matches(msg, m.MenuKey):
			m.ShowMenu = true
			m.Menu.current = "main"
			return m, nil
		default:
			m.InputHandler.HandleKeyMsg(msg)
			m.UnsavedChanges = true
			m.Viewport.SetContent(m.renderContent())
		}

	case tea.WindowSizeMsg:
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height
		for k := range m.Menu.lists {
			m.Menu.lists[k].SetWidth(msg.Width)
			m.Menu.lists[k].SetHeight(msg.Height)
		}
		m.Viewport.SetContent(m.renderContent())
	}

	return m, nil
}

func (m *UIModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Menu, cmd = m.Menu.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			if m.Menu.current == "main" {
				m.ShowMenu = false
			} else {
				m.Menu.current = "main"
			}
			return m, nil
		}
	case MenuMsg:
		switch msg.Action {
		case SaveAction:
			m.saveFile()
			m.ShowMenu = false
		case QuitAction:
			if m.UnsavedChanges {
				// Here you might want to add a confirmation dialog
				// For now, we'll just print a warning
				fmt.Println("Warning: You have unsaved changes!")
			}
			return m, tea.Quit
		case JoinSessionAction:
			if msg.Data != "Back to Main Menu" && msg.Data != "Back to Editor" && msg.Data != "Quit" {
				fmt.Printf("Joining session: %s\n", msg.Data)
				m.ShowMenu = false
			}
		case CreatePublicSessionAction:
			fmt.Println("Creating public session...")
			m.ShowMenu = false
		case CreatePrivateSessionAction:
			fmt.Println("Creating private session...")
			m.ShowMenu = false
		case BackToEditorAction:
			m.ShowMenu = false
		}
		return m, nil
	}

	return m, cmd
}

func (m *UIModel) View() string {
	if m.ShowMenu {
		return m.Menu.View()
	}
	return m.Viewport.View()
}

func (m *UIModel) renderContent() string {
	lineNumbers := m.Editor.GetLineNumbers()
	document := m.Editor.RenderDocument()

	lineNumberWidth := len(fmt.Sprintf("%d", strings.Count(document, "\n")+1))

	var output strings.Builder

	lines := strings.Split(document, "\n")
	numberLines := strings.Split(lineNumbers, "\n")

	totalLines := m.Viewport.Height - 3 // Subtracting 3 for filename, empty line, and status bar

	for i := 0; i < totalLines; i++ {
		if i < len(lines) {
			if i < len(numberLines) {
				output.WriteString(RenderLineNumber(fmt.Sprintf("%*s", lineNumberWidth, numberLines[i])))
			} else {
				output.WriteString(RenderLineNumber(fmt.Sprintf("%*s", lineNumberWidth, "")))
			}
			output.WriteString(" " + lines[i])
		} else {
			output.WriteString(RenderLineNumber(fmt.Sprintf("%*s", lineNumberWidth, "~")))
		}
		output.WriteString("\n")
	}

	header := RenderHeader(fmt.Sprintf("File: %s", m.FilePath))
	statusBar := m.renderStatusBar()

	return fmt.Sprintf("%s\n%s\n%s", header, output.String(), statusBar)
}

func (m *UIModel) renderStatusBar() string {
	unsavedIndicator := " "
	if m.UnsavedChanges {
		unsavedIndicator = "*"
	}

	leftStatus := fmt.Sprintf("%s%s", unsavedIndicator, m.FilePath)
	rightStatus := "Press ESC for menu"

	padding := strings.Repeat(" ", m.Viewport.Width-lipgloss.Width(leftStatus)-lipgloss.Width(rightStatus))

	return lipgloss.JoinHorizontal(lipgloss.Left,
		RenderStatusBar(leftStatus),
		padding,
		RenderStatusBar(rightStatus),
	)
}

func (m *UIModel) saveFile() {
	content := m.Editor.RenderDocumentWithoutLineNumbers()
	err := os.WriteFile(m.FilePath, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error saving file: %v\n", err)
	} else {
		m.UnsavedChanges = false
	}
}

func generateSiteID() string {
	return fmt.Sprintf("site-%d", time.Now().UnixNano())
}
