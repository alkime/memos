// Package recording provides the TUI model for the recording phase.
package recording

import (
	"fmt"

	"github.com/alkime/memos/internal/tui/style"
	"github.com/alkime/memos/pkg/uictl"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
)

// ToggleMsg signals a toggle of recording state.
type ToggleMsg struct{}

// Model represents the recording phase UI state.
type Model struct {
	controls  Controls
	spinner   spinner.Model
	stopwatch stopwatch.Model
	progress  progress.Model
	maxBytes  int64
	started   bool
}

// Controls provides read/write access to recording hardware.
type Controls struct {
	FileSize       uictl.CappedDial[int64]
	StartStopPause uictl.Knob
}

// New creates a new recording phase model.
func New(controls Controls, maxBytes int64) Model {
	s := spinner.New()
	s.Spinner = spinner.Points

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return Model{
		controls:  controls,
		spinner:   s,
		stopwatch: stopwatch.NewWithInterval(0), // Default interval
		progress:  p,
		maxBytes:  maxBytes,
		started:   false,
	}
}

// Init returns the initial command for the recording phase.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.stopwatch.Init(),
	)
}

// Update handles messages for the recording phase.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ToggleMsg:
		m.controls.StartStopPause.Toggle()
		if !m.started {
			m.started = true
			cmds = append(cmds, m.stopwatch.Start())
		}
		return m, tea.Batch(cmds...)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case stopwatch.TickMsg:
		var cmd tea.Cmd
		m.stopwatch, cmd = m.stopwatch.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the recording phase UI.
func (m Model) View() string {
	var s string

	// Recording indicator with spinner and stopwatch
	switch {
	case m.controls.StartStopPause.Read():
		s += m.spinner.View() + " "
		s += style.TitleStyle.Render("Recording") + " "
		s += style.SubtitleStyle.Render(m.stopwatch.View())
	case m.started:
		s += style.WarningStyle.Render("Paused") + " "
		s += style.SubtitleStyle.Render(m.stopwatch.View())
	default:
		s += style.SubtitleStyle.Render("Press space to start recording")
	}
	s += "\n\n"

	// File size progress
	current, maxValue := m.controls.FileSize.Cap()
	percent := float64(0)
	if maxValue > 0 {
		percent = float64(current) / float64(maxValue)
	}

	s += m.progress.ViewAs(percent) + "\n"
	s += style.SubtitleStyle.Render(formatBytes(current, maxValue))
	s += "\n\n"

	// Help text
	if m.started {
		s += style.HelpStyle.Render("[") + style.KeyStyle.Render("space") + style.HelpStyle.Render("] pause/resume  ")
		s += style.HelpStyle.Render("[") + style.KeyStyle.Render("enter") + style.HelpStyle.Render("] stop  ")
	} else {
		s += style.HelpStyle.Render("[") + style.KeyStyle.Render("space") + style.HelpStyle.Render("] start  ")
	}
	s += style.HelpStyle.Render("[") + style.KeyStyle.Render("q") + style.HelpStyle.Render("] quit")

	return s
}

// IsRecording returns whether recording is currently active.
func (m Model) IsRecording() bool {
	return m.controls.StartStopPause.Read()
}

// HasStarted returns whether recording has been started at least once.
func (m Model) HasStarted() bool {
	return m.started
}

// formatBytes formats bytes as a human-readable string.
func formatBytes(current, maxBytes int64) string {
	currentMB := float64(current) / (1024 * 1024)
	maxMB := float64(maxBytes) / (1024 * 1024)

	if maxBytes == 0 {
		return fmt.Sprintf("%.1f MB / unlimited", currentMB)
	}

	percent := int(float64(current) / float64(maxBytes) * 100)

	return fmt.Sprintf("%.1f MB / %.1f MB (%d%%)", currentMB, maxMB, percent)
}
