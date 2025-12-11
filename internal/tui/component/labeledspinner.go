// Package component provides reusable TUI components.
package component

import (
	"strings"

	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// LabeledSpinner displays a spinner with title, subtitle, and help text.
// This pattern is used across multiple TUI phases (finalizing, transcribing, generating).
type LabeledSpinner struct {
	Spinner  spinner.Model
	Title    string
	Subtitle string
	Help     string
}

// NewLabeledSpinner creates a new labeled spinner with the given configuration.
func NewLabeledSpinner(s spinner.Spinner, title, subtitle, help string) LabeledSpinner {
	sp := spinner.New()
	sp.Spinner = s

	return LabeledSpinner{
		Spinner:  sp,
		Title:    title,
		Subtitle: subtitle,
		Help:     help,
	}
}

// Init returns the initial command for the spinner.
func (ls LabeledSpinner) Init() tea.Cmd {
	return ls.Spinner.Tick
}

// Update handles spinner tick messages.
func (ls LabeledSpinner) Update(teaMsg tea.Msg) (LabeledSpinner, tea.Cmd) {
	if tickMsg, ok := teaMsg.(spinner.TickMsg); ok {
		var cmd tea.Cmd
		ls.Spinner, cmd = ls.Spinner.Update(tickMsg)

		return ls, cmd
	}

	return ls, nil
}

// View renders the labeled spinner with static help text.
func (ls LabeledSpinner) View() string {
	return ls.ViewWithHelp(ls.Help)
}

// ViewWithHelp renders the labeled spinner with dynamic help text.
// Use this when help text needs to be computed at render time (e.g., elapsed time).
func (ls LabeledSpinner) ViewWithHelp(help string) string {
	var sb strings.Builder

	sb.WriteString(ls.Spinner.View())
	sb.WriteString(" ")
	sb.WriteString(style.Title.Render(ls.Title))
	sb.WriteString("\n\n")

	sb.WriteString(style.Subtitle.Render(ls.Subtitle))
	sb.WriteString("\n\n")

	sb.WriteString(style.Help.Render(help))

	return sb.String()
}
