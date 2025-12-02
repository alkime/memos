package tui

import (
	"context"

	"github.com/alkime/memos/internal/tui/recording"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	cancel      context.CancelFunc
	recordingUI recording.Model
}

func New(cancel context.CancelFunc) tea.Model {
	return model{
		cancel:      cancel,
		recordingUI: recording.Model{},
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		}
	}

	// delegate to sub-models
	m.recordingUI, cmd = m.recordingUI.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return m.recordingUI.View()
}
