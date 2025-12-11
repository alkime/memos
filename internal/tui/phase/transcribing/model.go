// Package transcribing provides the TUI model for the transcription phase.
package transcribing

import (
	"fmt"
	"os"

	"github.com/alkime/memos/internal/cli/transcription"
	"github.com/alkime/memos/internal/tui/component"
	"github.com/alkime/memos/internal/tui/phase/msg"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the transcription phase UI state.
type Model struct {
	spinner    component.LabeledSpinner
	audioPath  string
	outputPath string
	client     *transcription.Client
}

// New creates a new transcription phase model.
func New(audioPath, outputPath, apiKey string) Model {
	return Model{
		spinner: component.NewLabeledSpinner(
			spinner.Dot,
			"Transcribing audio...",
			"Sending to Whisper API",
			"This may take a moment depending on audio length",
		),
		audioPath:  audioPath,
		outputPath: outputPath,
		client:     transcription.NewClient(apiKey),
	}
}

// Init returns the initial command for the transcription phase.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Init(),
		m.transcribeCmd(),
	)
}

// Update handles messages for the transcription phase.
func (m Model) Update(teaMsg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(teaMsg)

	return m, cmd
}

// View renders the transcription phase UI.
func (m Model) View() string {
	return m.spinner.View()
}

// transcribeCmd returns a command that performs the transcription.
func (m Model) transcribeCmd() tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(m.audioPath)
		if err != nil {
			return msg.TranscriptionErrorMsg{
				Err: fmt.Errorf("failed to open audio file %s: %w", m.audioPath, err),
			}
		}
		defer file.Close()

		text, err := m.client.TranscribeFile(file)
		if err != nil {
			return msg.TranscriptionErrorMsg{
				Err: fmt.Errorf("transcription failed: %w", err),
			}
		}

		// Save transcript to file
		//nolint:gosec // Transcript files need to be readable
		if err := os.WriteFile(m.outputPath, []byte(text), 0644); err != nil {
			return msg.TranscriptionErrorMsg{
				Err: fmt.Errorf("failed to save transcript to %s: %w", m.outputPath, err),
			}
		}

		return msg.TranscriptionCompleteMsg{
			Transcript: text,
			OutputPath: m.outputPath,
		}
	}
}
