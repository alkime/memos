// Package generating provides the TUI model for the first draft generation phase.
package generating

import (
	"fmt"
	"os"

	"github.com/alkime/memos/internal/cli/ai"
	"github.com/alkime/memos/internal/tui/msg"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the generating phase UI state.
type Model struct {
	spinner    spinner.Model
	transcript string
	mode       ai.Mode
	outputPath string
	client     *ai.Client
}

// New creates a new generating phase model.
func New(transcript string, mode ai.Mode, apiKey, outputPath string) Model {
	s := spinner.New()
	s.Spinner = spinner.Pulse

	return Model{
		spinner:    s,
		transcript: transcript,
		mode:       mode,
		outputPath: outputPath,
		client:     ai.NewClient(apiKey),
	}
}

// Init returns the initial command for the generating phase.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.generateCmd(),
	)
}

// Update handles messages for the generating phase.
func (m Model) Update(teaMsg tea.Msg) (Model, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(teaMsg)
		return m, cmd
	}

	return m, nil
}

// View renders the generating phase UI.
func (m Model) View() string {
	var s string

	s += m.spinner.View() + " "
	s += style.TitleStyle.Render("Generating first draft...")
	s += "\n\n"

	s += style.SubtitleStyle.Render(fmt.Sprintf("Mode: %s", m.mode))
	s += "\n\n"

	s += style.HelpStyle.Render("Claude is processing your transcript")

	return s
}

// generateCmd returns a command that performs the first draft generation.
func (m Model) generateCmd() tea.Cmd {
	return func() tea.Msg {
		draft, err := m.client.GenerateFirstDraft(m.transcript, m.mode)
		if err != nil {
			return msg.FirstDraftErrorMsg{
				Err: fmt.Errorf("failed to generate first draft: %w", err),
			}
		}

		// Save draft to file
		//nolint:gosec // Draft files need to be readable
		if err := os.WriteFile(m.outputPath, []byte(draft), 0644); err != nil {
			return msg.FirstDraftErrorMsg{
				Err: fmt.Errorf("failed to save first draft to %s: %w", m.outputPath, err),
			}
		}

		return msg.FirstDraftCompleteMsg{
			DraftPath: m.outputPath,
			Content:   draft,
		}
	}
}
