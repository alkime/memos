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

type viewTranscriptPhase struct {
	transcriptPath string
	viewport       viewport.Model
	keys           viewTranscriptKeyMap
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
	switch teaMsg := teaMsg.(type) {
	case tea.WindowSizeMsg:
		return vtp, vtp.reloadViewportCommand(teaMsg.Width, teaMsg.Height)
	case tea.KeyMsg:
		switch {
		case key.Matches(teaMsg, vtp.keys.Proceed):
			return vtp, phases.NextPhaseCmd
		}
	}

	var cmd tea.Cmd
	vtp.viewport, cmd = vtp.viewport.Update(teaMsg)

	return vtp, cmd
}

func (vtp *viewTranscriptPhase) View() string {
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

func (vtp *viewTranscriptPhase) reloadViewportCommand(width, height int) tea.Cmd {
	return func() tea.Msg {
		transcript, err := vtp.readTranscriptFile()
		if err != nil {
			slog.Error("Failed to read transcript file", "error", err)
			return tea.Quit
		}

		headerHeight := 3
		footerHeight := 3
		viewportHeight := height - headerHeight - footerHeight
		if viewportHeight < 5 {
			viewportHeight = 5
		}

		viewportWidth := width - 4
		vtp.viewport.Width = viewportWidth
		vtp.viewport.Height = viewportHeight

		// Re-wrap content for new width
		wrappedContent := wrapText(transcript, viewportWidth)
		vtp.viewport.SetContent(wrappedContent)

		return nil
	}
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
