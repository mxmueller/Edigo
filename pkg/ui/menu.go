package ui

import (
	"edigo/pkg/network"
	"edigo/pkg/theme"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type MenuItem struct {
	title string
	desc  string
}

func (i MenuItem) Title() string       { return i.title }
func (i MenuItem) Description() string { return i.desc }
func (i MenuItem) FilterValue() string { return i.title }

type MenuModel struct {
	lists    map[string]*list.Model
	current  string
	quitting bool
	Theme    *theme.Theme
}

type MenuAction string

const (
	NoAction                   MenuAction = "no_action"
	SaveAction                 MenuAction = "save"
	QuitAction                 MenuAction = "quit"
	CreateSessionAction        MenuAction = "create_session"
	CreatePublicSessionAction  MenuAction = "create_public_session"
	CreatePrivateSessionAction MenuAction = "create_private_session"
	JoinSessionAction          MenuAction = "join_session"
	BackToEditorAction         MenuAction = "back_to_editor"
	BackToMainMenuAction       MenuAction = "back_to_main_menu"
)

type MenuMsg struct {
	Action MenuAction
	Data   string
}

func NewMenuModel(theme *theme.Theme) MenuModel {
	mainItems := []list.Item{
		MenuItem{title: "Create Session", desc: "Start a new editing session"},
		MenuItem{title: "Join Session", desc: "Join an existing editing session"},
		MenuItem{title: "Save", desc: "Save the current file"},
		MenuItem{title: "Back to Editor", desc: "Return to the editor"},
		MenuItem{title: "Quit", desc: "Exit the editor"},
	}

	joinItems := []list.Item{}

	createItems := []list.Item{
		MenuItem{title: "Create Public Session", desc: "Create a session without a password"},
		MenuItem{title: "Create Private Session", desc: "Create a session with a password"},
		MenuItem{title: "Back to Main Menu", desc: "Return to main menu"},
		MenuItem{title: "Back to Editor", desc: "Return to the editor"},
		MenuItem{title: "Quit", desc: "Exit the editor"},
	}

	mainList := createList("Main Menu", mainItems, theme)
	joinList := createList("Join Session", joinItems, theme)
	createList := createList("Create Session", createItems, theme)

	return MenuModel{
		lists: map[string]*list.Model{
			"main":   mainList,
			"join":   joinList,
			"create": createList,
		},
		current: "main",
		Theme:   theme,
	}
}

func createList(title string, items []list.Item, theme *theme.Theme) *list.Model {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = theme.MenuTitleStyle
	l.Styles.PaginationStyle = theme.MenuItemStyle.Copy().PaddingLeft(4)
	l.Styles.HelpStyle = theme.MenuItemStyle.Copy().PaddingBottom(1)

	return &l
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg, network *network.Network) (MenuModel, tea.Cmd) {
	m.setSessions(network)

	var cmd tea.Cmd
	*m.lists[m.current], cmd = m.lists[m.current].Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for k := range m.lists {
			m.lists[k].SetWidth(msg.Width)
			m.lists[k].SetHeight(msg.Height)
		}
	case tea.KeyMsg:
		if msg.String() == "enter" {
			return m.handleEnter()
		}
	}
	return m, cmd
}

func (m MenuModel) handleEnter() (MenuModel, tea.Cmd) {
	item, ok := m.lists[m.current].SelectedItem().(MenuItem)
	if !ok {
		return m, nil
	}

	switch item.title {
	case "Create Session":
		m.current = "create"
		return m, nil
	case "Join Session":
		m.current = "join"
		return m, nil
	case "Save":
		return m, func() tea.Msg { return MenuMsg{Action: SaveAction} }
	case "Back to Editor":
		return m, func() tea.Msg { return MenuMsg{Action: BackToEditorAction} }
	case "Quit":
		return m, func() tea.Msg { return MenuMsg{Action: QuitAction} }
	case "Back to Main Menu":
		m.current = "main"
		return m, nil
	case "Create Public Session":
		return m, func() tea.Msg { return MenuMsg{Action: CreatePublicSessionAction} }
	case "Create Private Session":
		return m, func() tea.Msg { return MenuMsg{Action: CreatePrivateSessionAction} }
	default:
		if m.current == "join" {
			return m, func() tea.Msg { return MenuMsg{Action: JoinSessionAction, Data: item.title} }
		}
	}

	return m, nil
}

func (m MenuModel) View() string {
	return m.lists[m.current].View()
}

func (m *MenuModel) setSessions(network *network.Network) {
	sessionItems := []list.Item{}

	for name, session := range network.Sessions {
		sessionItems = append(sessionItems, MenuItem{title: name, desc: "IP: " + session.IP + ":" + strconv.Itoa(session.Port)})
	}

	sessionItems = append(sessionItems, MenuItem{title: "Back to Main Menu", desc: "Return to main menu"})
	sessionItems = append(sessionItems, MenuItem{title: "Back to Editor", desc: "Return to the editor"})

	m.lists["join"].SetItems(sessionItems)
}
