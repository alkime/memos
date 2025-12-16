package workflow

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alkime/memos/internal/tui/components/phases"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewTranscriptKeyMap struct {
	Proceed key.Binding
}

func defaultViewTranscriptKeyMap() viewTranscriptKeyMap {
	return viewTranscriptKeyMap{
		Proceed: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y/enter", "generate first draft"),
		),
	}
}

// viewportReadyMsg signals that viewport content has been loaded.
type viewportReadyMsg struct {
	content string
	width   int
	height  int
}

type viewTranscriptPhase struct {
	transcriptPath string
	viewport       viewport.Model
	keys           viewTranscriptKeyMap
	ready          bool
	width          int
	height         int
}

func NewViewTranscriptPhase(transcriptPath string) tea.Model {
	return &viewTranscriptPhase{
		transcriptPath: transcriptPath,
		keys:           defaultViewTranscriptKeyMap(),
	}
}

func (vtp *viewTranscriptPhase) Init() tea.Cmd {
	return tea.WindowSize()
}

func (vtp *viewTranscriptPhase) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		vtp.width = msg.Width
		vtp.height = msg.Height

		return vtp, vtp.loadViewportCmd()

	case viewportReadyMsg:
		vtp.setupViewport(msg)
		vtp.ready = true

		return vtp, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, vtp.keys.Proceed):
			return vtp, phases.NextPhaseCmd
		}
	}

	var cmd tea.Cmd
	vtp.viewport, cmd = vtp.viewport.Update(teaMsg)

	return vtp, cmd
}

func (vtp *viewTranscriptPhase) View() string {
	if !vtp.ready {
		return "Loading transcript..."
	}

	var sb strings.Builder

	// Header
	sb.WriteString(style.Title.Render("=== Transcript ==="))
	sb.WriteString("\n\n")

	// Viewport with border
	sb.WriteString(style.Viewport.Render(vtp.viewport.View()))
	sb.WriteString("\n\n")

	// Help text
	sb.WriteString(renderKeyHelp(vtp.keys.Proceed, "\n"))
	sb.WriteString(renderGlobalKeyHelp())

	return sb.String()
}

func (vtp *viewTranscriptPhase) loadViewportCmd() tea.Cmd {
	return func() tea.Msg {
		transcript, err := vtp.readTranscriptFile()
		if err != nil {
			slog.Error("Failed to read transcript file", "error", err)

			return tea.Quit
		}

		return viewportReadyMsg{
			content: transcript,
			width:   vtp.width,
			height:  vtp.height,
		}
	}
}

func (vtp *viewTranscriptPhase) setupViewport(msg viewportReadyMsg) {
	headerHeight := 3
	footerHeight := 3
	viewportHeight := msg.height - headerHeight - footerHeight
	if viewportHeight < 5 {
		viewportHeight = 5
	}

	viewportWidth := msg.width - 4
	if viewportWidth < 10 {
		viewportWidth = 10
	}

	vtp.viewport = viewport.New(viewportWidth, viewportHeight)
	vtp.viewport.SetContent(wrapText(msg.content, viewportWidth))
}

func (vtp *viewTranscriptPhase) readTranscriptFile() (string, error) {
	content, err := os.ReadFile(vtp.transcriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcript file: %w", err)
	}

	return string(content), nil
}

// wrapText wraps the given text to fit within the specified width using lipgloss.
// This ensures long lines wrap properly instead of being truncated in the viewport.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	// Use lipgloss.NewStyle().Width() to perform word wrapping
	wrapper := lipgloss.NewStyle().Width(width)

	return wrapper.Render(text)
}
