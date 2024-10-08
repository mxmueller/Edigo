package ui

import (
	"edigo/pkg/editor"
	"edigo/pkg/theme"
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
	Menu           MenuModel
	ShowMenu       bool
	UnsavedChanges bool
	Theme          *theme.Theme
	ErrorMsg       string
}

func NewUIModel(content string, filePath string) *UIModel {
	siteID := generateSiteID()
	theme := theme.NewTheme()
	editorInstance := editor.NewEditor(content, filePath, siteID, theme)
	vp := viewport.New(80, 24)
	editorInstance.Viewport = vp
	editorInstance.FilePath = filePath

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
		Menu:           NewMenuModel(theme),
		ShowMenu:       false,
		UnsavedChanges: false,
		Theme:          theme,
		ErrorMsg:       "",
	}
}

func (m *UIModel) Init() tea.Cmd {
	m.Viewport.SetContent(m.Editor.RenderContent())
	return tea.Batch(
		tea.EnterAltScreen,
		waitForActivity(m.Editor.Update),
	)
}

func (m *UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.ShowMenu {
		return m.updateMenu(msg)
	}

	var cmd tea.Cmd
	m.Viewport, cmd = m.Viewport.Update(msg)

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
		}

	case tea.WindowSizeMsg:
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height - 2 // Reserve space for header and footer
		m.Editor.Viewport.Width = msg.Width
		m.Editor.Viewport.Height = msg.Height - 2
		for k := range m.Menu.lists {
			m.Menu.lists[k].SetWidth(msg.Width)
			m.Menu.lists[k].SetHeight(msg.Height)
		}

	case editor.RemoteChange:
		if !m.Editor.RGA.VerifyIntegrity() {
			m.ErrorMsg = "Data integrity check failed. Please refresh the session."
		} else {
			m.ErrorMsg = ""
		}
	}

	m.Viewport.SetContent(m.Editor.RenderContent())
	return m, tea.Batch(cmd, waitForActivity(m.Editor.Update))
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
				fmt.Println("Warning: You have unsaved changes!")
			}
			m.Editor.Stop()
			return m, tea.Quit
		case JoinSessionAction:
			if msg.Data != "Back to Main Menu" && msg.Data != "Back to Editor" && msg.Data != "Quit" {
				m.ShowMenu = false
				if m.Editor.Network.IsHost || m.Editor.Network.Host != nil {
					m.Editor.Error = "Already in a Session"
					m.Viewport.SetContent(m.Editor.RenderContent())
					break
				}

				go m.Editor.HandleConnections()
				*m.Editor.RGA = m.Editor.Network.JoinSession(msg.Data)
				m.Viewport.SetContent(m.Editor.RenderContent())
				m.Editor.SendCursorUpdate()
			}
		case CreatePublicSessionAction:
			fmt.Println("Creating public session...")

			m.ShowMenu = false
			if m.Editor.Network.IsHost || m.Editor.Network.Host != nil {
				m.Editor.Error = "Already in a Session"
				m.Viewport.SetContent(m.Editor.RenderContent())
				break
			}

			go m.Editor.HandleConnections()
			go m.Editor.Network.BroadcastSession(m.Editor.RGA)
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
		return m.Theme.RenderMenuTitle("Menu") + "\n" + m.Menu.View()
	}

	headerContent := m.Editor.FilePath
	if m.UnsavedChanges {
		headerContent += " [Unsaved Changes]"
	}
	header := m.Theme.RenderHeader(headerContent)

	content := m.Viewport.View()

	footerContent := "Press ESC for menu"
	if m.ErrorMsg != "" {
		footerContent = m.Theme.RenderError(m.ErrorMsg)
	}
	footer := m.Theme.RenderFooter(footerContent)

	return header + "\n" + content + "\n" + footer
}

func (m *UIModel) saveFile() {
	if m.Editor.Network.CurrentSession != "" && !m.Editor.Network.IsHost {
		return
	}

	content := m.Editor.RenderDocumentWithoutLineNumbers()
	err := os.WriteFile(m.Editor.FilePath, []byte(content), 0644)
	if err != nil {
		m.ErrorMsg = fmt.Sprintf("Error saving file: %v", err)
	} else {
		m.UnsavedChanges = false
		m.ErrorMsg = "File saved successfully"
	}
	m.Viewport.SetContent(m.Editor.RenderContent())
}

func generateSiteID() string {
	return fmt.Sprintf("site-%d", time.Now().UnixNano())
}
