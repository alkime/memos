package tui

import (
	"strings"

	"github.com/alkime/memos/internal/tui/components/phases"
	tuiPhases "github.com/alkime/memos/internal/tui/phases"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/alkime/memos/internal/workdir"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// model2 is the new TUI model using the phases component.
type model2 struct {
	config       Config
	keys         tuiPhases.KeyMap
	phases       phases.Model
	windowWidth  int
	windowHeight int
}

// New2 creates a new TUI model using the phases component.
func New2(config Config, recordingControls tuiPhases.RecordingControls) tea.Model {
	var phs []phases.Phase

	phs = append(phs, phases.NewPhase("Recording", tuiPhases.NewRecording(recordingControls, config.MaxBytes)))
	phs = append(phs, phases.NewPhase("Transcribing", tuiPhases.NewTranscribePhase(
		workdir.MustFilePath(config.WorkingName, workdir.MP3File),
		workdir.MustFilePath(config.WorkingName, workdir.TranscriptFile),
		config.OpenAIAPIKey)))

	phs = append(phs, phases.NewPhase("View Transcript", tuiPhases.NewViewTranscriptPhase(
		workdir.MustFilePath(config.WorkingName, workdir.TranscriptFile),
	)))

	phs = append(phs, phases.NewPhase("First Draft", tuiPhases.NewFirstDraftPhase(
		workdir.MustFilePath(config.WorkingName, workdir.TranscriptFile),
		workdir.MustFilePath(config.WorkingName, workdir.FirstDraftFile),
		config.AnthropicAPIKey,
		config.Mode,
	)))

	phs = append(phs, phases.NewPhase("Edit Draft", tuiPhases.NewEditDraftPhase(
		workdir.MustFilePath(config.WorkingName, workdir.FirstDraftFile),
		config.EditorCmd,
	)))
	// TODO: Add remaining phases as we migrate them

	return &model2{
		config:       config,
		keys:         tuiPhases.DefaultKeyMap(),
		phases:       phases.New(phs),
		windowWidth:  80,
		windowHeight: 24,
	}
}

// Init returns the initial command.
func (m *model2) Init() tea.Cmd {
	return m.phases.Init()
}

// Update handles all messages.
func (m *model2) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size for all phases
	if wsm, ok := teaMsg.(tea.WindowSizeMsg); ok {
		m.windowWidth = wsm.Width
		m.windowHeight = wsm.Height
	}

	// Global key handling (quit from any phase)
	if km, ok := teaMsg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(km, m.keys.ForceQuit):
			if m.config.Cancel != nil {
				m.config.Cancel()
			}

			return m, tea.Quit

		case key.Matches(km, m.keys.Quit):
			if m.config.Cancel != nil {
				m.config.Cancel()
			}

			return m, tea.Quit
		}
	}

	// Delegate to phases container
	updatedPhases, cmd := m.phases.Update(teaMsg)
	m.phases = updatedPhases.(phases.Model) //nolint:forcetypeassert // phases.Model always returns phases.Model

	return m, cmd
}

// View renders the current UI.
func (m *model2) View() string {
	var sb strings.Builder

	// Add header with current phase name
	sb.WriteString(style.Subtitle.Render("Phase: " + m.phases.CurrentPhaseName()))
	sb.WriteString("\n\n")

	// Render current phase
	sb.WriteString(m.phases.View())

	return sb.String()
}
