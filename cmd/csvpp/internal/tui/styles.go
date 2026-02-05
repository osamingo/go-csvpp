// Package tui provides TUI components for the csvpp CLI.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles holds the styles for the TUI components.
type Styles struct {
	Header   lipgloss.Style
	Cell     lipgloss.Style
	Selected lipgloss.Style
	Help     lipgloss.Style
	Status   lipgloss.Style
}

// DefaultStyles returns the default styles for the TUI.
func DefaultStyles() Styles {
	return Styles{
		Header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
		Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Padding(0, 1),
		Help:     lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		Status:   lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1),
	}
}
