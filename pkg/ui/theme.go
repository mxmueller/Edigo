package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#F25D94")
	accentColor    = lipgloss.Color("#10B981")
	textColor      = lipgloss.Color("#FFFFFF")
	mutedTextColor = lipgloss.Color("#4B5563")

	// Base Styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Header Styles
	HeaderStyle = BaseStyle.Copy().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor)

	// Line Number Styles
	LineNumberStyle = BaseStyle.Copy().
			Foreground(mutedTextColor).
			Width(4).
			Align(lipgloss.Right)

	// Cursor Style
	CursorStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Background(textColor)

	// Menu Styles
	MenuTitleStyle = BaseStyle.Copy().
			Foreground(primaryColor).
			Bold(true).
			Margin(1, 0, 1, 2)

	MenuItemStyle = BaseStyle.Copy().
			PaddingLeft(4)

	MenuSelectedItemStyle = MenuItemStyle.Copy().
				Foreground(accentColor).
				Bold(true)

	// Status Bar Style
	StatusBarStyle = BaseStyle.Copy().
			Foreground(mutedTextColor)
)

// RenderLineNumber formats and styles the line number
func RenderLineNumber(number string) string {
	return LineNumberStyle.Render(number)
}

// RenderCursor styles the cursor
func RenderCursor(content string) string {
	return CursorStyle.Render(content)
}

// RenderHeader styles the header (file name)
func RenderHeader(content string) string {
	return HeaderStyle.Render(content)
}

// RenderMenuTitle styles the menu title
func RenderMenuTitle(title string) string {
	return MenuTitleStyle.Render(title)
}

// RenderMenuItem styles a menu item
func RenderMenuItem(item string, selected bool) string {
	if selected {
		return MenuSelectedItemStyle.Render(item)
	}
	return MenuItemStyle.Render(item)
}

// RenderStatusBar styles the status bar
func RenderStatusBar(content string) string {
	return StatusBarStyle.Render(content)
}
