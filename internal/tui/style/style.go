// Package style defines lipgloss styles for the TUI.
package style

import "github.com/charmbracelet/lipgloss"

// UI styles using lipgloss.
// These are package-level for convenience; lipgloss styles are value types
// and safe for concurrent use.
//
//nolint:gochecknoglobals // Lipgloss styles are idiomatic as package-level vars
var (
	// TitleStyle is used for phase titles and headers.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	// SubtitleStyle is used for secondary text.
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// SuccessStyle is used for success messages.
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	// ErrorStyle is used for error messages.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	// WarningStyle is used for warning messages.
	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	// ViewportStyle is used for the transcript viewport border.
	ViewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	// HelpStyle is used for keyboard shortcut hints.
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// KeyStyle is used for highlighting keyboard keys.
	KeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	// ProgressStyle is used for progress indicators.
	ProgressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63"))
)
