// Package finalizing provides the TUI model for waiting on MP3 conversion.
package finalizing

import (
	"time"

	"github.com/alkime/memos/internal/tui/components/labeledspinner"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the finalizing phase UI state.
// It displays a spinner while waiting for an external signal
// (AudioFinalizingCompleteMsg) that the MP3 conversion is complete.
type Model struct {
	spinner   labeledspinner.Model
	startTime time.Time
}

// New creates a new finalizing phase model.
func New() Model {
	return Model{
		spinner: labeledspinner.New(
			spinner.Dot,
			"Finalizing audio...",
			"Converting to MP3 format",
			"", // Help text is dynamic (elapsed time)
		),
		startTime: time.Now(),
	}
}

// Init returns the initial command for the finalizing phase.
func (m Model) Init() tea.Cmd {
	return m.spinner.Init()
}

// Update handles messages for the finalizing phase.
func (m Model) Update(teaMsg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(teaMsg)

	return m, cmd
}

// View renders the finalizing phase UI.
func (m Model) View() string {
	elapsed := time.Since(m.startTime).Round(time.Second)

	return m.spinner.ViewWithHelp("Elapsed: " + elapsed.String())
}
