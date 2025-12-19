package tui

import (
	"context"
	"strings"

	"github.com/alkime/memos/internal/content"
	"github.com/alkime/memos/internal/platform/workdir"
	"github.com/alkime/memos/internal/tui/components/phases"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/alkime/memos/internal/tui/workflow"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Config holds TUI configuration.
type Config struct {
	Cancel          context.CancelFunc
	WorkingName     string
	OpenAIAPIKey    string
	AnthropicAPIKey string
	Mode            content.Mode
	MaxBytes        int64
	EditorCmd       string
	OutputDir       string
}

// model is the TUI model using the phases component.
type model struct {
	config       Config
	keys         workflow.KeyMap
	phases       phases.Model
	windowWidth  int
	windowHeight int
}

// New creates a new TUI model using the phases component.
func New(config Config, recordingControls workflow.RecordingControls) tea.Model {
	var phs []phases.Phase

	phs = append(phs, phases.NewPhase("Recording", workflow.NewRecording(
		recordingControls,
		config.MaxBytes,
		workdir.MustFilePath(config.WorkingName, workdir.MP3File),
	)))
	phs = append(phs, phases.NewPhase("Transcribing", workflow.NewTranscribePhase(
		workdir.MustFilePath(config.WorkingName, workdir.MP3File),
		workdir.MustFilePath(config.WorkingName, workdir.TranscriptFile),
		config.OpenAIAPIKey)))

	phs = append(phs, phases.NewPhase("View Transcript", workflow.NewViewTranscriptPhase(
		workdir.MustFilePath(config.WorkingName, workdir.TranscriptFile),
	)))

	phs = append(phs, phases.NewPhase("First Draft", workflow.NewFirstDraftPhase(
		workdir.MustFilePath(config.WorkingName, workdir.TranscriptFile),
		workdir.MustFilePath(config.WorkingName, workdir.FirstDraftFile),
		config.AnthropicAPIKey,
		config.Mode,
	)))

	phs = append(phs, phases.NewPhase("Edit Draft", workflow.NewEditDraftPhase(
		workdir.MustFilePath(config.WorkingName, workdir.FirstDraftFile),
		config.EditorCmd,
	)))

	phs = append(phs, phases.NewPhase("Copy Edit", workflow.NewCopyEditPhase(
		workdir.MustFilePath(config.WorkingName, workdir.FirstDraftFile),
		config.AnthropicAPIKey,
		config.Mode,
		config.OutputDir,
	)))

	return &model{
		config:       config,
		keys:         workflow.DefaultKeyMap(),
		phases:       phases.New(phs),
		windowWidth:  80,
		windowHeight: 24,
	}
}

// Init returns the initial command.
func (m *model) Init() tea.Cmd {
	return m.phases.Init()
}

// Update handles all messages.
func (m *model) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *model) View() string {
	var sb strings.Builder

	// Add header with current phase name
	sb.WriteString(style.Subtitle.Render("Phase: " + m.phases.CurrentPhaseName()))
	sb.WriteString("\n\n")

	// Render current phase
	sb.WriteString(m.phases.View())

	return sb.String()
}
