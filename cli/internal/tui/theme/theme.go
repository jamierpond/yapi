// Package theme defines colors and styles for the TUI.
package theme

import "github.com/charmbracelet/lipgloss"

// Shared Color Palette (extracted from webapp/tailwind.config.js)
var (
	Bg         = lipgloss.Color("#1a1b26")
	BgElevated = lipgloss.Color("#2a2d3b")
	Fg         = lipgloss.Color("#a9b1d6")
	FgMuted    = lipgloss.Color("#565f89")
	Accent     = lipgloss.Color("#ff9e64") // Orange
	Primary    = lipgloss.Color("#7D56F4") // Purple
	Border     = lipgloss.Color("#414868")
	Green      = lipgloss.Color("#69DB7C")
	Red        = lipgloss.Color("#FF6B6B")
	Yellow     = lipgloss.Color("#FFE066")
)

// Shared Styles
var (
	App = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border)

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(Primary).
		Padding(0, 1)

	TitleAccent = lipgloss.NewStyle().
			Foreground(Bg).
			Background(Accent).
			Padding(0, 1).
			Bold(true)

	Item = lipgloss.NewStyle().
		PaddingLeft(2)

	SelectedItem = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(Accent).
			Bold(true)

	Footer = lipgloss.NewStyle().
		Foreground(FgMuted).
		Padding(0, 1).
		MarginTop(1)

	ViewportContent = lipgloss.NewStyle().
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border)

	Error   = lipgloss.NewStyle().Foreground(Red)
	Success = lipgloss.NewStyle().Foreground(Green)
	Warn    = lipgloss.NewStyle().Foreground(Yellow)
	Info    = lipgloss.NewStyle().Foreground(FgMuted)

	BorderedBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary)

	Help = lipgloss.NewStyle().
		Foreground(FgMuted)
)
