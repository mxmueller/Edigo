package ui

import (
	"edigo/pkg/editor"
	"fmt"

	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
    Network      network.Network
}

func NewUIModel(content string, filePath string) *UIModel {
    update := make(chan struct{}, 1)
	siteID := generateSiteID()
	editorInstance := editor.NewEditor(content, siteID)
	vp := viewport.New(80, 24)
    editorInstance.Viewport = &vp
    editorInstance.Update = update

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
        Network: network.NewNetwork(),
	}
}

func (m *UIModel) Init() tea.Cmd {
    m.Viewport.SetContent(m.Editor.RenderContent())
    	return tea.Batch(
        tea.EnterAltScreen,
		waitForActivity(m.updateEvent),   // wait for activity
	)
}

func (m *UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.ShowMenu {
		return m.updateMenu(msg)
	}
	m.Viewport.SetContent(m.Editor.RenderContent())

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
			m.Viewport.SetContent(m.Editor.RenderContent())
		}

	case tea.WindowSizeMsg:
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height
		for k := range m.Menu.lists {
			m.Menu.lists[k].SetWidth(msg.Width)
			m.Menu.lists[k].SetHeight(msg.Height)
		}
		m.Viewport.SetContent(m.Editor.RenderContent())
	case editor.RemoteChange:
		m.Viewport.SetContent(m.Editor.RenderContent())
        return m, waitForActivity(m.updateEvent)
	}

	return m, nil
}

func waitForActivity(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return editor.RemoteChange(<-sub)
	}
}

func (m *UIModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Menu, cmd = m.Menu.Update(msg, m.Editor.Network)

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
                *m.Editor.RGA = m.Editor.Network.JoinSession(msg.Data) // TODO implement Error handeling
				m.ShowMenu = false
		        m.Viewport.SetContent(m.Editor.RenderContent())
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
