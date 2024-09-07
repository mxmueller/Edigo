package theme

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type UserTheme struct {
	MainColor    lipgloss.Color
	LighterColor lipgloss.Color
	DarkerColor  lipgloss.Color
}

type Theme struct {
	BaseStyle             lipgloss.Style
	HeaderStyle           lipgloss.Style
	LineNumberStyle       lipgloss.Style
	CursorStyle           lipgloss.Style
	MenuTitleStyle        lipgloss.Style
	MenuItemStyle         lipgloss.Style
	MenuSelectedItemStyle lipgloss.Style
	StatusBarStyle        lipgloss.Style
	UsernameStyle         lipgloss.Style
	ErrorStyle            lipgloss.Style // Neue Stil-Definition für Fehler
	LineNumberPadding     int
	UserThemes            []UserTheme
}

func NewTheme() *Theme {
	primaryColor := lipgloss.Color("#1230AE")
	secondaryColor := lipgloss.Color("6C48C5")
	accentColor := lipgloss.Color("#C68FE6")
	textColor := lipgloss.Color("#FFFFFF")
	mutedTextColor := lipgloss.Color("#4B5563")
	errorColor := lipgloss.Color("#FF0000") // Rote Farbe für Fehler

	baseStyle := lipgloss.NewStyle().Foreground(textColor)

	userThemes := []UserTheme{
		{MainColor: lipgloss.Color("#6C48C5"), LighterColor: lipgloss.Color("#C68FE6"), DarkerColor: lipgloss.Color("#1230AE")},
		{MainColor: lipgloss.Color("#41B3A2"), LighterColor: lipgloss.Color("#BDE8CA"), DarkerColor: lipgloss.Color("#0D7C66")},
		{MainColor: lipgloss.Color("#0000FF"), LighterColor: lipgloss.Color("#6666FF"), DarkerColor: lipgloss.Color("#0000CC")},
		{MainColor: lipgloss.Color("#FFFF00"), LighterColor: lipgloss.Color("#FFFF66"), DarkerColor: lipgloss.Color("#CCCC00")},
		{MainColor: lipgloss.Color("#FF00FF"), LighterColor: lipgloss.Color("#FF66FF"), DarkerColor: lipgloss.Color("#CC00CC")},
		{MainColor: lipgloss.Color("#00FFFF"), LighterColor: lipgloss.Color("#66FFFF"), DarkerColor: lipgloss.Color("#00CCCC")},
		{MainColor: lipgloss.Color("#FF8000"), LighterColor: lipgloss.Color("#FFA64D"), DarkerColor: lipgloss.Color("#CC6600")},
		{MainColor: lipgloss.Color("#8000FF"), LighterColor: lipgloss.Color("#A64DFF"), DarkerColor: lipgloss.Color("#6600CC")},
		{MainColor: lipgloss.Color("#0080FF"), LighterColor: lipgloss.Color("#4DA6FF"), DarkerColor: lipgloss.Color("#0066CC")},
		{MainColor: lipgloss.Color("#FF0080"), LighterColor: lipgloss.Color("#FF4DA6"), DarkerColor: lipgloss.Color("#CC0066")},
	}

	return &Theme{
		BaseStyle: baseStyle,
		HeaderStyle: baseStyle.Copy().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor),
		LineNumberStyle: baseStyle.Copy().
			Foreground(mutedTextColor).
			Width(4).
			Align(lipgloss.Right),
		CursorStyle: lipgloss.NewStyle().
			Foreground(accentColor).
			Background(textColor),
		MenuTitleStyle: baseStyle.Copy().
			Foreground(primaryColor).
			Bold(true).
			Margin(1, 0, 1, 2),
		MenuItemStyle: baseStyle.Copy().
			PaddingLeft(4),
		MenuSelectedItemStyle: baseStyle.Copy().
			Foreground(accentColor).
			Bold(true).
			PaddingLeft(4),
		StatusBarStyle: baseStyle.Copy().
			Foreground(mutedTextColor),
		UsernameStyle: baseStyle.Copy().
			Foreground(accentColor).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1).
			Bold(true),
		ErrorStyle: baseStyle.Copy().
			Foreground(errorColor).
			Bold(true),
		LineNumberPadding: 2,
		UserThemes:        userThemes,
	}
}

func (t *Theme) RenderLineNumber(number string, width int) string {
	renderedNumber := t.LineNumberStyle.Render(fmt.Sprintf("%*s", width, number))
	padding := strings.Repeat(" ", t.LineNumberPadding)
	return renderedNumber + padding
}

func (t *Theme) RenderCursor(isSharedSession bool, themeIndex int) string {
	if !isSharedSession {
		return t.CursorStyle.Render("█")
	}
	userTheme := t.UserThemes[themeIndex%len(t.UserThemes)]
	return lipgloss.NewStyle().Foreground(userTheme.MainColor).Render("█")
}

func (t *Theme) RenderHeader(content string) string {
	return t.HeaderStyle.Render(content)
}

func (t *Theme) RenderMenuTitle(title string) string {
	return t.MenuTitleStyle.Render(title)
}

func (t *Theme) RenderMenuItem(item string, selected bool) string {
	if selected {
		return t.MenuSelectedItemStyle.Render(item)
	}
	return t.MenuItemStyle.Render(item)
}

func (t *Theme) RenderStatusBar(content string) string {
	return t.StatusBarStyle.Render(content)
}

func (t *Theme) RenderUsername(username string, themeIndex int) string {
	userTheme := t.UserThemes[themeIndex%len(t.UserThemes)]
	style := t.UsernameStyle.Copy().
		Foreground(userTheme.DarkerColor).
		Background(userTheme.LighterColor)
	return style.Render(username)
}

func (t *Theme) RenderError(content string) string {
	return t.ErrorStyle.Render(content)
}
