// Package viewtranscript provides the TUI model for viewing the transcript.
package viewtranscript

import (
	"github.com/alkime/memos/internal/tui/msg"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the view transcript phase UI state.
type Model struct {
	viewport   viewport.Model
	transcript string
	ready      bool
}

// New creates a new view transcript phase model.
func New(transcript string, width, height int) Model {
	// Reserve space for header and footer
	headerHeight := 3
	footerHeight := 3
	viewportHeight := height - headerHeight - footerHeight

	if viewportHeight < 5 {
		viewportHeight = 5
	}

	vp := viewport.New(width-4, viewportHeight) // -4 for border padding
	vp.SetContent(transcript)

	return Model{
		viewport:   vp,
		transcript: transcript,
		ready:      true,
	}
}

// Init returns the initial command for the view transcript phase.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the view transcript phase.
func (m Model) Update(teaMsg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 3
		footerHeight := 3
		viewportHeight := teaMsg.Height - headerHeight - footerHeight
		if viewportHeight < 5 {
			viewportHeight = 5
		}

		m.viewport.Width = teaMsg.Width - 4
		m.viewport.Height = viewportHeight
		m.ready = true

	case tea.KeyMsg:
		switch teaMsg.String() {
		case "y", "enter":
			return m, func() tea.Msg { return msg.ProceedToFirstDraftMsg{} }
		case "n", "s":
			return m, func() tea.Msg { return msg.SkipFirstDraftMsg{} }
		}
	}

	// Handle viewport scrolling
	m.viewport, cmd = m.viewport.Update(teaMsg)

	return m, cmd
}

// View renders the view transcript phase UI.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var s string

	// Header
	s += style.TitleStyle.Render("=== Transcript ===")
	s += "\n\n"

	// Viewport with border
	s += style.ViewportStyle.Render(m.viewport.View())
	s += "\n\n"

	// Footer with keyboard hints
	s += style.HelpStyle.Render("[")
	s += style.KeyStyle.Render("y") + style.HelpStyle.Render("/")
	s += style.KeyStyle.Render("enter") + style.HelpStyle.Render("] generate first draft  ")
	s += style.HelpStyle.Render("[")
	s += style.KeyStyle.Render("n") + style.HelpStyle.Render("/")
	s += style.KeyStyle.Render("s") + style.HelpStyle.Render("] skip  ")
	s += style.HelpStyle.Render("[") + style.KeyStyle.Render("q") + style.HelpStyle.Render("] quit")

	return s
}

// Transcript returns the transcript text.
func (m Model) Transcript() string {
	return m.transcript
}
