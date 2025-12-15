package phases

import (
	tea "github.com/charmbracelet/bubbletea"
)

// NextPhaseMsg signals the phases container to advance to the next phase.
type NextPhaseMsg struct{}

// PrevPhaseMsg signals the phases container to go back to the previous phase.
type PrevPhaseMsg struct{}

type Phase struct {
	Name string
	mdl  tea.Model
}

func (p Phase) Init() tea.Cmd {
	return p.mdl.Init()
}

func (p Phase) Update(msg tea.Msg) (Phase, tea.Cmd) {
	updatedMdl, cmd := p.mdl.Update(msg)
	p.mdl = updatedMdl
	return p, cmd
}

func (p Phase) View() string {
	return p.mdl.View()
}

func NewPhase(name string, mdl tea.Model) Phase {
	return Phase{
		Name: name,
		mdl:  mdl,
	}
}

type Model struct {
	phases []Phase
	curr   int
}

func New(phases []Phase) Model {
	return Model{
		phases: phases,
		curr:   0,
	}
}

func (m Model) currentPhase() Phase {
	return m.phases[m.curr]
}

func (m Model) Init() tea.Cmd {
	return m.currentPhase().Init()
}

func (m Model) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch teaMsg.(type) {
	case NextPhaseMsg:
		if m.curr >= len(m.phases)-1 {
			return m, nil
		}
		m.curr++
		initCmd := m.currentPhase().Init()
		return m, initCmd

	case PrevPhaseMsg:
		if m.curr <= 0 {
			return m, nil
		}
		m.curr--
		return m, m.currentPhase().Init()
	}

	ph, cmd := m.currentPhase().Update(teaMsg)
	m.phases[m.curr] = ph

	return m, cmd
}

func (m Model) View() string {
	return m.currentPhase().View()
}

// CurrentPhaseName returns the name of the current phase.
func (m Model) CurrentPhaseName() string {
	return m.currentPhase().Name
}
