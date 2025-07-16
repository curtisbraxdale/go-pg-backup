package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	pink  = lipgloss.Color("205")
	green = lipgloss.Color("78")
	red   = lipgloss.Color("196")
	grey  = lipgloss.Color("240")
	white = lipgloss.Color("255")
	amber = lipgloss.Color("214")

	// Text styles
	pinkTextPrompt  = lipgloss.NewStyle().Foreground(pink).Bold(true).Italic(true)
	greenTextPrompt = lipgloss.NewStyle().Foreground(green).Bold(true).Italic(true)
	greenTextValue  = lipgloss.NewStyle().Foreground(green).Italic(true)
	greyText        = lipgloss.NewStyle().Foreground(grey)
	whiteText       = lipgloss.NewStyle().Foreground(white)
	welcomeStyle    = lipgloss.NewStyle().Foreground(pink).Bold(true).Italic(true)
	summaryStyle    = lipgloss.NewStyle().Foreground(green).Bold(true).Italic(true)
	errorStyle      = lipgloss.NewStyle().Foreground(red).Bold(true).Italic(true)
	cancelledStyle  = lipgloss.NewStyle().Foreground(amber).Bold(true).Italic(true)

	// Button styles
	focusedButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("231")).
			Background(pink).
			Padding(0, 1)
	blurredButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("231")).
			Background(lipgloss.Color("237")).
			Padding(0, 1)

	// Help style
	helpStyle = lipgloss.NewStyle().Foreground(grey)
)
