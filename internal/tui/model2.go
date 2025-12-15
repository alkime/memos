package tui

import (
	"strings"

	"github.com/alkime/memos/internal/tui/components/phases"
	tuiPhases "github.com/alkime/memos/internal/tui/phases"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// model2 is the new TUI model using the phases component.
type model2 struct {
	config       Config
	keys         KeyMap
	phases       phases.Model
	windowWidth  int
	windowHeight int
}

// New2 creates a new TUI model using the phases component.
func New2(config Config, recordingControls tuiPhases.RecordingControls) tea.Model {
	recordingPhase := tuiPhases.NewRecording(recordingControls, config.MaxBytes)

	return &model2{
		config: config,
		keys:   DefaultKeyMap(),
		phases: phases.New([]phases.Phase{
			phases.NewPhase("Recording", recordingPhase),
			// TODO: Add remaining phases as we migrate them
		}),
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
