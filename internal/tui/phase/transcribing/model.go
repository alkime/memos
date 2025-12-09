// Package transcribing provides the TUI model for the transcription phase.
package transcribing

import (
	"fmt"
	"os"

	"github.com/alkime/memos/internal/cli/transcription"
	"github.com/alkime/memos/internal/tui/msg"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the transcription phase UI state.
type Model struct {
	spinner    spinner.Model
	audioPath  string
	outputPath string
	client     *transcription.Client
}

// New creates a new transcription phase model.
func New(audioPath, outputPath, apiKey string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{
		spinner:    s,
		audioPath:  audioPath,
		outputPath: outputPath,
		client:     transcription.NewClient(apiKey),
	}
}

// Init returns the initial command for the transcription phase.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.transcribeCmd(),
	)
}

// Update handles messages for the transcription phase.
func (m Model) Update(teaMsg tea.Msg) (Model, tea.Cmd) {
	switch teaMsg := teaMsg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(teaMsg)
		return m, cmd
	}

	return m, nil
}

// View renders the transcription phase UI.
func (m Model) View() string {
	var s string

	s += m.spinner.View() + " "
	s += style.TitleStyle.Render("Transcribing audio...")
	s += "\n\n"

	s += style.SubtitleStyle.Render("Sending to Whisper API")
	s += "\n\n"

	s += style.HelpStyle.Render("This may take a moment depending on audio length")

	return s
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
