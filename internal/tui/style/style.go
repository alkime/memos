// Package style defines lipgloss styles for the TUI.
package style

import "github.com/charmbracelet/lipgloss"

// UI styles using lipgloss.
// These are package-level for convenience; lipgloss styles are value types
// and safe for concurrent use.
//
// Variable names intentionally omit "Style" suffix since they're accessed
// via the style package (e.g., style.Title reads better than style.TitleStyle).
var (
	// Title is used for phase titles and headers.
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	// Subtitle is used for secondary text.
	Subtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Success is used for success messages.
	Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("42"))

	// Error is used for error messages.
	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	// Warning is used for warning messages.
	Warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	// Viewport is used for the transcript viewport border.
	Viewport = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	// Help is used for keyboard shortcut hints.
	Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Key is used for highlighting keyboard keys.
	Key = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	// Progress is used for progress indicators.
	Progress = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63"))

	// Label is used for inline labels (e.g., "Title:", "Saved:").
	Label = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255"))

	// Muted is used for de-emphasized text (e.g., file paths).
	Muted = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	// Bullet is used for list item markers.
	Bullet = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))
)
