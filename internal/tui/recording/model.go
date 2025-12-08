package recording

import (
	"fmt"

	"github.com/alkime/memos/pkg/uictl"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type ToggleMsg struct{}

type Model struct {
	controls RecordingControls
	spinner  spinner.Model
}

func New(controls RecordingControls) Model {
	return Model{
		controls: controls,
		spinner:  spinner.New(spinner.WithSpinner(spinner.Points)),
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ToggleMsg:
		m.controls.StartStopPause.Toggle()
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	s := ""
	if m.controls.StartStopPause.Read() {
		s += m.spinner.View() + " "
	}

	num, max := m.controls.FileSize.Cap()
	s += formatBytes(num, max, 90)
	return s
}

// format functions copy n pasted from recorder.go.

// formatWithBold wraps text in ANSI bold codes if shouldBold is true.
// this is leftover, we should definitely move to using bubbletea style system (called lipgloss ofc)
func formatWithBold(text string, shouldBold bool) string {
	if shouldBold {
		return fmt.Sprintf("\033[1m%s\033[0m", text)
	}

	return text
}

// formatBytes formats bytes in MB with optional bold.
func formatBytes(current, maxBytes int64, thresholdPerc int) string {
	currentMB := float64(current) / (1024 * 1024)
	maxMB := float64(maxBytes) / (1024 * 1024)
	percent := int(float64(current) / float64(maxBytes) * 100)

	text := fmt.Sprintf("%.1f MB / %.1f MB (%d%%)", currentMB, maxMB, percent)

	return formatWithBold(text, percent >= thresholdPerc)
}

type RecordingControls struct {
	FileSize       uictl.CappedDial[int64]
	StartStopPause uictl.Knob
}
