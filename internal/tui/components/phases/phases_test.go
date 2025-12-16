//nolint:funlen // Test file
package phases_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/alkime/memos/internal/tui/components/phases"
	"github.com/alkime/memos/pkg/collections"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoinits // recommend for CI by bubbletea folks
func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

func TestPhases(t *testing.T) {
	checker := outputChecker{
		intervl: 100 * time.Millisecond,
		timeout: 1 * time.Second,
	}

	p1 := &modelMock{t: t, name: "phase1"}
	p2 := &modelMock{t: t, name: "phase2"}
	p3 := &modelMock{t: t, name: "phase3"}
	p4 := &modelMock{t: t, name: "phase4"}

	ph := phases.New([]phases.Phase{
		phases.NewPhase("one", p1),
		phases.NewPhase("two", p2),
		phases.NewPhase("three", p3),
		phases.NewPhase("four", p4),
	})

	tm := teatest.NewTestModel(t, ph, teatest.WithInitialTermSize(300, 100))

	t.Run("initial phase is phase1", func(t *testing.T) {
		checker.CheckString(t, tm, "phase1")
		t.Run("phase init state checks", func(t *testing.T) {
			checks := collections.ApplyVariadic(func(m *modelMock) bool {
				return m.initCalled
			}, p1, p2, p3, p4)
			require.Equal(t, []bool{true, false, false, false}, checks, "check state of inits across phases")
		})

		t.Run("phase updated state checks", func(t *testing.T) {
			checks := collections.ApplyVariadic(func(m *modelMock) bool {
				return m.updated
			}, p1, p2, p3, p4)
			require.Equal(t, []bool{false, false, false, false}, checks, "check state of updates across phases")
		})
	})

	t.Run("advance a phase", func(t *testing.T) {
		tm.Send(phases.NextPhaseMsg{})
		checker.CheckString(t, tm, "phase2")
		t.Run("phase init state checks", func(t *testing.T) {
			checks := collections.ApplyVariadic(func(m *modelMock) bool {
				return m.initCalled
			}, p1, p2, p3, p4)
			require.Equal(t, []bool{true, true, false, false}, checks, "check state of inits across phases")
		})

		t.Run("phase updated state checks", func(t *testing.T) {
			checks := collections.ApplyVariadic(func(m *modelMock) bool {
				return m.updated
			}, p1, p2, p3, p4)
			require.Equal(t, []bool{false, false, false, false}, checks, "check state of updates across phases")
		})
	})

	t.Run("send mockMsg that triggers a forward advance", func(t *testing.T) {
		tm.Send(mockMsg{triggerForward: true})
		checker.CheckString(t, tm, "phase3")
		t.Run("phase init state checks", func(t *testing.T) {
			checks := collections.ApplyVariadic(func(m *modelMock) bool {
				return m.initCalled
			}, p1, p2, p3, p4)
			require.Equal(t, []bool{true, true, true, false}, checks, "check state of inits across phases")
		})

		t.Run("phase updated state checks", func(t *testing.T) {
			checks := collections.ApplyVariadic(func(m *modelMock) bool {
				return m.updated
			}, p1, p2, p3, p4)
			require.Equal(t, []bool{false, true, false, false}, checks, "check state of updates across phases")
		})
	})
}

type modelMock struct {
	t          *testing.T
	name       string
	updated    bool
	initCalled bool
}

func (m *modelMock) Init() tea.Cmd {
	m.initCalled = true
	return nil
}
func (m *modelMock) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.t.Logf("modelMock Update called: %s, msg: %#v\n", m.name, msg)

	switch msg := msg.(type) {
	case mockMsg:
		m.updated = true
		if msg.triggerForward {
			return m, m.commandNextPhase()
		}
	}

	return m, nil
}

func (m *modelMock) commandNextPhase() tea.Cmd {
	return func() tea.Msg {
		return phases.NextPhaseMsg{}
	}
}

func (m *modelMock) View() string { return m.name }

type outputChecker struct {
	intervl, timeout time.Duration
}

func (o outputChecker) Check(t *testing.T, tm *teatest.TestModel, check func(buf []byte) bool) {
	teatest.WaitFor(t, tm.Output(), check,
		teatest.WithCheckInterval(o.intervl),
		teatest.WithDuration(o.timeout))
}

func (o outputChecker) CheckString(t *testing.T, tm *teatest.TestModel, substr string) {
	o.Check(t, tm, func(buf []byte) bool {
		return bytes.Contains(buf, []byte(substr))
	})
}

type mockMsg struct {
	triggerForward bool
}
