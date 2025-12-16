package phases

import (
	"log/slog"
	"os"

	"github.com/alkime/memos/internal/cli/transcription"
	"github.com/alkime/memos/internal/tui/components/labeledspinner"
	"github.com/alkime/memos/internal/tui/components/phases"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type transcribePhase struct {
	spinner                 labeledspinner.Model
	audioInputPath          string
	transcriptionOutputPath string
	client                  *transcription.Client
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
	}
}

func (tp *transcribePhase) Init() tea.Cmd {
	return tea.Sequence(
		tp.spinner.Init(),
		tp.transcribeCmd(),
	)
}

func (tp *transcribePhase) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	tp.spinner, cmd = tp.spinner.Update(teaMsg)

	return tp, cmd
}

func (tp *transcribePhase) View() string {
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
