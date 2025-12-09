// Package finalizing provides the TUI model for waiting on MP3 conversion.
package finalizing

import (
	"time"

	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the finalizing phase UI state.
// It displays a spinner while waiting for an external signal
// (AudioFinalizingCompleteMsg) that the MP3 conversion is complete.
type Model struct {
	spinner   spinner.Model
	startTime time.Time
}

// New creates a new finalizing phase model.
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{
		spinner:   s,
		startTime: time.Now(),
	}
}

// Init returns the initial command for the finalizing phase.
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages for the finalizing phase.
func (m Model) Update(teaMsg tea.Msg) (Model, tea.Cmd) {
	if tickMsg, ok := teaMsg.(spinner.TickMsg); ok {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(tickMsg)

		return m, cmd
	}

	return m, nil
}

// View renders the finalizing phase UI.
func (m Model) View() string {
	var s string

	s += m.spinner.View() + " "
	s += style.TitleStyle.Render("Finalizing audio...")
	s += "\n\n"

	s += style.SubtitleStyle.Render("Converting to MP3 format")
	s += "\n\n"

	elapsed := time.Since(m.startTime).Round(time.Second)
	s += style.HelpStyle.Render("Elapsed: " + elapsed.String())

	return s
}
