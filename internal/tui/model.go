package tui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/alkime/memos/internal/cli/ai"
	"github.com/alkime/memos/internal/tui/phase"
	"github.com/alkime/memos/internal/tui/phase/finalizing"
	"github.com/alkime/memos/internal/tui/phase/generating"
	"github.com/alkime/memos/internal/tui/phase/msg"
	"github.com/alkime/memos/internal/tui/phase/recording"
	"github.com/alkime/memos/internal/tui/phase/transcribing"
	"github.com/alkime/memos/internal/tui/phase/viewtranscript"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// editorContext is the context used for editor commands.
// It uses the background context since editor processes should
// continue even if the TUI is shutting down.
var editorContext = context.Background() //nolint:gochecknoglobals // Package-level for exec context

// Config holds TUI configuration.
type Config struct {
	Cancel          context.CancelFunc
	AudioPath       string
	TranscriptPath  string
	DraftPath       string
	OpenAIAPIKey    string
	AnthropicAPIKey string
	Mode            ai.Mode
	MaxBytes        int64
	EditorCmd       string
}

type model struct {
	config Config
	keys   KeyMap
	phase  phase.Phase

	// Sub-models (only one active at a time)
	recordingUI      recording.Model
	finalizingUI     finalizing.Model
	transcribingUI   transcribing.Model
	viewTranscriptUI viewtranscript.Model
	generatingUI     generating.Model

	// Shared state
	transcript   string
	draftPath    string
	err          error
	windowWidth  int
	windowHeight int
}

// New creates a new TUI model.
func New(config Config, recordingControls recording.Controls) tea.Model {
	return model{
		config:       config,
		keys:         DefaultKeyMap(),
		phase:        phase.PhaseRecording,
		recordingUI:  recording.New(recordingControls, config.MaxBytes),
		windowWidth:  80, // Default
		windowHeight: 24, // Default
	}
}

// Init returns the initial command.
func (m model) Init() tea.Cmd {
	return m.recordingUI.Init()
}

// Update handles all messages.
func (m model) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
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

	// Phase-specific handling
	switch m.phase {
	case phase.PhaseRecording:
		return m.updateRecording(teaMsg)
	case phase.PhaseFinalizingAudio:
		return m.updateFinalizing(teaMsg)
	case phase.PhaseTranscribing:
		return m.updateTranscribing(teaMsg)
	case phase.PhaseViewTranscript:
		return m.updateViewTranscript(teaMsg)
	case phase.PhaseGeneratingDraft:
		return m.updateGenerating(teaMsg)
	case phase.PhaseComplete:
		return m.updateComplete(teaMsg)
	case phase.PhaseError:
		return m.updateError(teaMsg)
	}

	return m, nil
}

func (m model) updateRecording(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := teaMsg.(type) {
	case tea.KeyMsg:
		switch typedMsg.String() {
		case " ":
			// Toggle recording
			return m, func() tea.Msg { return recording.ToggleMsg{} }
		case "enter":
			// Stop recording and transition to finalizing (wait for MP3)
			if m.recordingUI.HasStarted() {
				if m.config.Cancel != nil {
					m.config.Cancel()
				}

				// Check if we have API key for transcription
				if m.config.OpenAIAPIKey == "" {
					m.err = fmt.Errorf("OPENAI_API_KEY not set, cannot transcribe")
					m.phase = phase.PhaseError

					return m, nil
				}

				m.phase = phase.PhaseFinalizingAudio
				m.finalizingUI = finalizing.New()

				return m, m.finalizingUI.Init()
			}
		}
	}

	var cmd tea.Cmd
	m.recordingUI, cmd = m.recordingUI.Update(teaMsg)

	return m, cmd
}

func (m model) updateFinalizing(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := teaMsg.(type) {
	case msg.AudioFinalizingCompleteMsg:
		m.phase = phase.PhaseTranscribing
		m.transcribingUI = transcribing.New(
			typedMsg.AudioPath,
			m.config.TranscriptPath,
			m.config.OpenAIAPIKey,
		)

		return m, m.transcribingUI.Init()

	case msg.AudioFinalizingErrorMsg:
		m.err = typedMsg.Err
		m.phase = phase.PhaseError

		return m, nil
	}

	var cmd tea.Cmd
	m.finalizingUI, cmd = m.finalizingUI.Update(teaMsg)

	return m, cmd
}

func (m model) updateTranscribing(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := teaMsg.(type) {
	case msg.TranscriptionCompleteMsg:
		m.transcript = typedMsg.Transcript
		m.phase = phase.PhaseViewTranscript
		m.viewTranscriptUI = viewtranscript.New(
			m.transcript,
			m.windowWidth,
			m.windowHeight,
		)

		return m, m.viewTranscriptUI.Init()

	case msg.TranscriptionErrorMsg:
		m.err = typedMsg.Err
		m.phase = phase.PhaseError
		return m, nil
	}

	var cmd tea.Cmd
	m.transcribingUI, cmd = m.transcribingUI.Update(teaMsg)

	return m, cmd
}

func (m model) updateViewTranscript(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch teaMsg.(type) {
	case msg.ProceedToFirstDraftMsg:
		// Check if we have API key for generation
		if m.config.AnthropicAPIKey == "" {
			m.err = fmt.Errorf("ANTHROPIC_API_KEY not set, cannot generate first draft")
			m.phase = phase.PhaseError
			return m, nil
		}

		m.phase = phase.PhaseGeneratingDraft
		m.generatingUI = generating.New(
			m.transcript,
			m.config.Mode,
			m.config.AnthropicAPIKey,
			m.config.DraftPath,
		)

		return m, m.generatingUI.Init()

	case msg.SkipFirstDraftMsg:
		m.phase = phase.PhaseComplete
		return m, nil
	}

	var cmd tea.Cmd
	m.viewTranscriptUI, cmd = m.viewTranscriptUI.Update(teaMsg)

	return m, cmd
}

func (m model) updateGenerating(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := teaMsg.(type) {
	case msg.FirstDraftCompleteMsg:
		m.draftPath = typedMsg.DraftPath
		m.phase = phase.PhaseComplete
		return m, m.openEditorCmd()

	case msg.FirstDraftErrorMsg:
		m.err = typedMsg.Err
		m.phase = phase.PhaseError
		return m, nil
	}

	var cmd tea.Cmd
	m.generatingUI, cmd = m.generatingUI.Update(teaMsg)

	return m, cmd
}

func (m model) updateComplete(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch teaMsg.(type) {
	case msg.OpenEditorCompleteMsg:
		return m, tea.Quit
	case tea.KeyMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateError(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := teaMsg.(type) {
	case tea.KeyMsg:
		switch typedMsg.String() {
		case "r":
			// Retry - go back to appropriate phase
			return m.retryFromError()
		case "s":
			// Skip to complete
			m.phase = phase.PhaseComplete

			return m, nil
		}
	}

	return m, nil
}

func (m model) retryFromError() (tea.Model, tea.Cmd) {
	// Determine which phase to retry based on error context
	// For now, just go to complete
	m.phase = phase.PhaseComplete
	m.err = nil
	return m, nil
}

func (m model) openEditorCmd() tea.Cmd {
	if m.draftPath == "" {
		return nil
	}

	editorCmd := m.config.EditorCmd
	if editorCmd == "" {
		editorCmd = "open" // macOS default
	}

	c := exec.CommandContext(editorContext, editorCmd, m.draftPath)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return msg.OpenEditorCompleteMsg{FilePath: m.draftPath, Err: err}
	})
}

// View renders the current UI.
func (m model) View() string {
	var sb strings.Builder

	// Add header with current phase
	sb.WriteString(style.Subtitle.Render(fmt.Sprintf("Phase: %s", m.phase.String())))
	sb.WriteString("\n\n")

	switch m.phase {
	case phase.PhaseRecording:
		sb.WriteString(m.viewRecording())
	case phase.PhaseFinalizingAudio:
		sb.WriteString(m.finalizingUI.View())
	case phase.PhaseTranscribing:
		sb.WriteString(m.transcribingUI.View())
	case phase.PhaseViewTranscript:
		sb.WriteString(m.viewTranscriptUI.View())
	case phase.PhaseGeneratingDraft:
		sb.WriteString(m.generatingUI.View())
	case phase.PhaseComplete:
		sb.WriteString(m.viewComplete())
	case phase.PhaseError:
		sb.WriteString(m.viewError())
	}

	return sb.String()
}

func (m model) viewRecording() string {
	var sb strings.Builder

	sb.WriteString(style.Title.Render("Recording to: "))
	sb.WriteString(m.config.AudioPath)
	sb.WriteString("\n\n")
	sb.WriteString(m.recordingUI.View())

	return sb.String()
}

func (m model) viewComplete() string {
	var sb strings.Builder

	sb.WriteString(style.Success.Render("Workflow complete!"))
	sb.WriteString("\n\n")

	if m.draftPath != "" {
		sb.WriteString("First draft saved to: ")
		sb.WriteString(m.draftPath)
		sb.WriteString("\n\n")
		sb.WriteString(style.Help.Render("Opening in editor..."))
	} else if m.transcript != "" {
		sb.WriteString("Transcript saved to: ")
		sb.WriteString(m.config.TranscriptPath)
		sb.WriteString("\n\n")
		sb.WriteString(style.Help.Render("First draft generation was skipped"))
	}

	sb.WriteString("\n\n")
	sb.WriteString(style.Help.Render("Press any key to exit"))

	return sb.String()
}

func (m model) viewError() string {
	var sb strings.Builder

	sb.WriteString(style.Error.Render("Error occurred:"))
	sb.WriteString("\n\n")

	if m.err != nil {
		sb.WriteString(m.err.Error())
	} else {
		sb.WriteString("Unknown error")
	}

	sb.WriteString("\n\n")
	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render("r"))
	sb.WriteString(style.Help.Render("] retry  "))
	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render("s"))
	sb.WriteString(style.Help.Render("] skip  "))
	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render("q"))
	sb.WriteString(style.Help.Render("] quit"))

	return sb.String()
}
