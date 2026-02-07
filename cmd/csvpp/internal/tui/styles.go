// Package tui provides TUI components for the csvpp CLI.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette for the TUI theme.
var (
	colorAccent  = lipgloss.Color("229") // yellow – header/selected foreground, filter prompt
	colorPrimary = lipgloss.Color("57")  // purple – header/selected background
	colorMuted   = lipgloss.Color("241") // gray   – help/status text
	colorFilter  = lipgloss.Color("86")  // green  – active filter indicator
)

// Styles holds the styles for the TUI components.
type Styles struct {
	Header       lipgloss.Style
	Cell         lipgloss.Style
	Selected     lipgloss.Style
	Help         lipgloss.Style
	Status       lipgloss.Style
	FilterPrompt lipgloss.Style
	FilterActive lipgloss.Style
}

// DefaultStyles returns the default styles for the TUI.
func DefaultStyles() Styles {
	return Styles{
		Header:       lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Background(colorPrimary).Padding(0, 1),
		Cell:         lipgloss.NewStyle().Padding(0, 1),
		Selected:     lipgloss.NewStyle().Foreground(colorAccent).Background(colorPrimary).Padding(0, 1),
		Help:         lipgloss.NewStyle().Foreground(colorMuted),
		Status:       lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 1),
		FilterPrompt: lipgloss.NewStyle().Bold(true).Foreground(colorAccent),
		FilterActive: lipgloss.NewStyle().Foreground(colorFilter),
	}
}
