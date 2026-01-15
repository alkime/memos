package workflow

import (
	"context"
	"log/slog"
	"os/exec"
	"time"

	"github.com/alkime/memos/internal/tui/components/phases"
	tea "github.com/charmbracelet/bubbletea"
)

type editDraftPhase struct {
	draftPath string
	launcher  EditorLauncher
}

// NewEditDraftPhase creates a new edit draft phase.
func NewEditDraftPhase(launcher EditorLauncher, draftPath string) tea.Model {
	return &editDraftPhase{
		draftPath: draftPath,
		launcher:  launcher,
	}
}

// DefaultEditorLauncher provides the default editor launching behavior.
type DefaultEditorLauncher struct {
	EditorCmd string
}

// Launch opens the file in the configured editor.
//
//nolint:gosec // subprocess launching
func (d *DefaultEditorLauncher) Launch(filePath string) tea.Cmd {
	var c *exec.Cmd
	if d.EditorCmd == "" {
		// macOS default: blocking open in new window
		c = exec.CommandContext(context.Background(), "open", "-Wn", filePath)
	} else {
		c = exec.CommandContext(context.Background(), d.EditorCmd, filePath)
	}

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorCompleteMsg{err: err}
	})
}

type startEditorMsg struct{}

func (ep *editDraftPhase) Init() tea.Cmd {
	// Use a tick to allow the view to render before launching editor.
	// This is a workaround for a Bubble Tea limitation where the ticker-based
	// renderer may not have rendered before tea.ExecProcess suspends the TUI.
	// See: https://github.com/charmbracelet/bubbletea/pull/1429
	return tea.Tick(250*time.Millisecond, func(time.Time) tea.Msg {
		return startEditorMsg{}
	})
}

func (ep *editDraftPhase) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := teaMsg.(type) {
	case startEditorMsg:
		return ep, ep.openEditorCmd()
	case editorCompleteMsg:
		if msg.err != nil {
			slog.Error("Editor closed with error", "error", msg.err)
		}

		return ep, func() tea.Msg { return phases.NextPhaseMsg{} }
	}

	return ep, nil
}

func (ep *editDraftPhase) View() string {
	return "Opening editor..."
}

type editorCompleteMsg struct {
	err error
}

func (ep *editDraftPhase) openEditorCmd() tea.Cmd {
	return ep.launcher.Launch(ep.draftPath)
}
