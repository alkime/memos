package phases

import (
	"log/slog"
	"os"

	"github.com/alkime/memos/internal/cli/ai"
	"github.com/alkime/memos/internal/tui/components/labeledspinner"
	"github.com/alkime/memos/internal/tui/components/phases"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type firstDraftPhase struct {
	spinner        labeledspinner.Model
	transcriptPath string
	outputPath     string
	mode           ai.Mode
	client         *ai.Client
}

// NewFirstDraftPhase creates a new first draft generation phase.
func NewFirstDraftPhase(transcriptPath, outputPath, apiKey string, mode ai.Mode) tea.Model {
	return &firstDraftPhase{
		spinner: labeledspinner.New(
			spinner.Pulse,
			"Generating first draft...",
			"Claude is processing your transcript",
			"This may take a moment",
		),
		transcriptPath: transcriptPath,
		outputPath:     outputPath,
		mode:           mode,
		client:         ai.NewClient(apiKey),
	}
}

func (fp *firstDraftPhase) Init() tea.Cmd {
	return tea.Sequence(
		fp.spinner.Init(),
		fp.generateCmd(),
	)
}

func (fp *firstDraftPhase) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	fp.spinner, cmd = fp.spinner.Update(teaMsg)

	return fp, cmd
}

func (fp *firstDraftPhase) View() string {
	return fp.spinner.View()
}

func (fp *firstDraftPhase) generateCmd() tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(fp.transcriptPath)
		if err != nil {
			slog.Error("Failed to read transcript file", "error", err)
			return tea.Quit
		}

		draft, err := fp.client.GenerateFirstDraft(string(content), fp.mode)
		if err != nil {
			slog.Error("First draft generation failed", "error", err)
			return tea.Quit
		}

		//nolint:gosec // Transcript files need to be readable
		if err := os.WriteFile(fp.outputPath, []byte(draft), 0o644); err != nil {
			slog.Error("Failed to write first draft", "error", err)
			return tea.Quit
		}

		return phases.NextPhaseMsg{}
	}
}
