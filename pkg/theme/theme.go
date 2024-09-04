package theme

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Theme struct {
	BaseStyle             lipgloss.Style
	HeaderStyle           lipgloss.Style
	LineNumberStyle       lipgloss.Style
	CursorStyle           lipgloss.Style
	MenuTitleStyle        lipgloss.Style
	MenuItemStyle         lipgloss.Style
	MenuSelectedItemStyle lipgloss.Style
	StatusBarStyle        lipgloss.Style
	LineNumberPadding     int
}

func NewTheme() *Theme {
	primaryColor := lipgloss.Color("#1230AE")
	secondaryColor := lipgloss.Color("6C48C5")
	accentColor := lipgloss.Color("#C68FE6")
	textColor := lipgloss.Color("#FFFFFF")
	mutedTextColor := lipgloss.Color("#4B5563")

	baseStyle := lipgloss.NewStyle().Foreground(textColor)

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
		LineNumberPadding: 2,
	}
}

func (t *Theme) RenderLineNumber(number string, width int) string {
	renderedNumber := t.LineNumberStyle.Render(fmt.Sprintf("%*s", width, number))
	padding := strings.Repeat(" ", t.LineNumberPadding)
	return renderedNumber + padding
}

func (t *Theme) RenderCursor(content string) string {
	return t.CursorStyle.Render(content)
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
