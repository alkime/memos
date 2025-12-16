package workflow

import (
	"log/slog"
	"os"

	"github.com/alkime/memos/internal/cli/transcription"
	"github.com/alkime/memos/internal/tui/components/labeledspinner"
	"github.com/alkime/memos/internal/tui/components/phases"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type transcribePhase struct {
	spinner                 labeledspinner.Model
	audioInputPath          string
	transcriptionOutputPath string
	client                  *transcription.Client
	existingOutput          existingOutputState
}

func NewTranscribePhase(audioInputPath, transcriptionOutputPath, apiKey string) tea.Model {
	return &transcribePhase{
		spinner: labeledspinner.New(
			spinner.Dot,
			"Transcribing audio...",
			"Sending to Whisper API",
			"This may take a moment depending on audio length",
		),
		audioInputPath:          audioInputPath,
		transcriptionOutputPath: transcriptionOutputPath,
		client:                  transcription.NewClient(apiKey),
		existingOutput:          newExistingOutputState(transcriptionOutputPath),
	}
}

func (tp *transcribePhase) Init() tea.Cmd {
	// Skip transcription if output already exists
	if tp.existingOutput.found {
		return nil
	}

	return tea.Sequence(
		tp.spinner.Init(),
		tp.transcribeCmd(),
	)
}

func (tp *transcribePhase) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle existing output keybindings
	if tp.existingOutput.found {
		if keyMsg, ok := teaMsg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, tp.existingOutput.keys.UseExisting):
				return tp, phases.NextPhaseCmd
			case key.Matches(keyMsg, tp.existingOutput.keys.Redo):
				tp.existingOutput.found = false

				return tp, tea.Sequence(tp.spinner.Init(), tp.transcribeCmd())
			}
		}

		return tp, nil
	}

	var cmd tea.Cmd
	tp.spinner, cmd = tp.spinner.Update(teaMsg)

	return tp, cmd
}

func (tp *transcribePhase) View() string {
	if tp.existingOutput.found {
		return renderExistingOutputView(tp.existingOutput, "Transcript")
	}

	return tp.spinner.View()
}

func (tp *transcribePhase) transcribeCmd() tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(tp.audioInputPath)
		if err != nil {
			slog.Error("Failed to open audio file for transcription", "error", err)
			return tea.Quit
		}
		defer file.Close()

		text, err := tp.client.TranscribeFile(file)
		if err != nil {
			slog.Error("Transcription failed", "error", err)
			return tea.Quit
		}

		//nolint:gosec // Transcript files need to be readable
		if err := os.WriteFile(tp.transcriptionOutputPath, []byte(text), 0o644); err != nil {
			slog.Error("Failed to write transcription output", "error", err)
			return tea.Quit
		}

		return phases.NextPhaseMsg{}
	}
}
