package tui

import (
	"context"

	"github.com/alkime/memos/internal/tui/recording"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	cancel      context.CancelFunc
	recordingUI recording.Model
	outputFile  string
}

func New(cancel context.CancelFunc, outputFile string, recordingControls recording.Controls) tea.Model {
	return model{
		cancel:      cancel,
		recordingUI: recording.New(recordingControls),
		outputFile:  outputFile,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		// only one for now abut we'll have more...
		m.recordingUI.Init(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch typedMsg := msg.(type) {
	case tea.KeyMsg:
		switch typedMsg.String() {
		case "ctrl+c", "q":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		case " ", "enter":
			msg = recording.ToggleMsg{}
		}
	}

	// delegate to sub-models
	m.recordingUI, cmd = m.recordingUI.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	s := "recording to: " + m.outputFile + "\n\n"
	s += "> ctrl+c or q to quit.\n"
	s += "> space or enter to start/stop/pause.\n\n"
	s += m.recordingUI.View() + "\n"

	return s
}
