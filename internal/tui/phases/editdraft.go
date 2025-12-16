package phases

import (
	"context"
	"log/slog"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type editDraftPhase struct {
	draftPath string
	editorCmd string
}

// NewEditDraftPhase creates a new edit draft phase.
func NewEditDraftPhase(draftPath, editorCmd string) tea.Model {
	return &editDraftPhase{
		draftPath: draftPath,
		editorCmd: editorCmd,
	}
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

		// return ep, func() tea.Msg { return phases.NextPhaseMsg{} }
		// FOR NOW, this is the final stage--quit the TUI
		return ep, tea.Quit
	}

	return ep, nil
}

func (ep *editDraftPhase) View() string {
	return "Opening editor..."
}

type editorCompleteMsg struct {
	err error
}

//nolint:gosec // subprocess launching
func (ep *editDraftPhase) openEditorCmd() tea.Cmd {
	var c *exec.Cmd
	if ep.editorCmd == "" {
		// macOS default: blocking open in new window
		c = exec.CommandContext(context.Background(), "open", "-Wn", ep.draftPath)
	} else {
		c = exec.CommandContext(context.Background(), ep.editorCmd, ep.draftPath)
	}

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorCompleteMsg{err: err}
	})
}
