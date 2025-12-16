// Package generating provides the TUI model for the first draft generation phase.
package generating

import (
	"fmt"
	"os"

	"github.com/alkime/memos/internal/cli/ai"
	"github.com/alkime/memos/internal/tui/components/labeledspinner"
	"github.com/alkime/memos/internal/tui/phase/msg"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the generating phase UI state.
type Model struct {
	spinner    labeledspinner.Model
	transcript string
	mode       ai.Mode
	outputPath string
	client     *ai.Client
}

// New creates a new generating phase model.
func New(transcript string, mode ai.Mode, apiKey, outputPath string) Model {
	return Model{
		spinner: labeledspinner.New(
			spinner.Pulse,
			"Generating first draft...",
			fmt.Sprintf("Mode: %s", mode),
			"Claude is processing your transcript",
		),
		transcript: transcript,
		mode:       mode,
		outputPath: outputPath,
		client:     ai.NewClient(apiKey),
	}
}

// Init returns the initial command for the generating phase.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Init(),
		m.generateCmd(),
	)
}

// Update handles messages for the generating phase.
func (m Model) Update(teaMsg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(teaMsg)

	return m, cmd
}

// View renders the generating phase UI.
func (m Model) View() string {
	return m.spinner.View()
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
