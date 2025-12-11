// Package viewtranscript provides the TUI model for viewing the transcript.
package viewtranscript

import (
	"strings"

	"github.com/alkime/memos/internal/tui/phase/msg"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the view transcript phase UI state.
type Model struct {
	keys       KeyMap
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

	viewportWidth := width - 4 // -4 for border padding
	vp := viewport.New(viewportWidth, viewportHeight)

	// Wrap the transcript text to fit the viewport width using lipgloss
	wrappedContent := wrapText(transcript, viewportWidth)
	vp.SetContent(wrappedContent)

	return Model{
		keys:       DefaultKeyMap(),
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

		viewportWidth := teaMsg.Width - 4
		m.viewport.Width = viewportWidth
		m.viewport.Height = viewportHeight

		// Re-wrap content for new width
		wrappedContent := wrapText(m.transcript, viewportWidth)
		m.viewport.SetContent(wrappedContent)
		m.ready = true

	case tea.KeyMsg:
		switch {
		case key.Matches(teaMsg, m.keys.Proceed):
			return m, func() tea.Msg { return msg.ProceedToFirstDraftMsg{} }
		case key.Matches(teaMsg, m.keys.Skip):
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

	var sb strings.Builder

	// Header
	sb.WriteString(style.Title.Render("=== Transcript ==="))
	sb.WriteString("\n\n")

	// Viewport with border
	sb.WriteString(style.Viewport.Render(m.viewport.View()))
	sb.WriteString("\n\n")

	// Footer with keyboard hints from KeyMap
	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render(m.keys.Proceed.Help().Key))
	sb.WriteString(style.Help.Render("] "))
	sb.WriteString(style.Help.Render(m.keys.Proceed.Help().Desc))
	sb.WriteString("  ")
	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render(m.keys.Skip.Help().Key))
	sb.WriteString(style.Help.Render("] "))
	sb.WriteString(style.Help.Render(m.keys.Skip.Help().Desc))
	sb.WriteString("  ")
	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render("q"))
	sb.WriteString(style.Help.Render("] quit"))

	return sb.String()
}

// Transcript returns the transcript text.
func (m Model) Transcript() string {
	return m.transcript
}

// wrapText wraps the given text to fit within the specified width using lipgloss.
// This ensures long lines wrap properly instead of being truncated in the viewport.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	// Use lipgloss.NewStyle().Width() to perform word wrapping
	wrapper := lipgloss.NewStyle().Width(width)

	return wrapper.Render(text)
}
