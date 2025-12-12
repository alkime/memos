// Package recording provides the TUI model for the recording phase.
package recording

import (
	"fmt"
	"strings"

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
		stopwatch: stopwatch.New(),
		progress:  p,
		maxBytes:  maxBytes,
	}
}

// Init returns the initial command for the recording phase.
func (m Model) Init() tea.Cmd {
	// Only start spinner - stopwatch starts when user toggles recording
	return m.spinner.Tick
}

// Update handles messages for the recording phase.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ToggleMsg:
		m.controls.StartStopPause.Toggle()
		if m.IsRecording() {
			cmds = append(cmds, m.stopwatch.Start())
		} else {
			cmds = append(cmds, m.stopwatch.Stop())
		}
		return m, tea.Batch(cmds...)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	// Always update stopwatch - it needs StartStopMsg, TickMsg, and ResetMsg
	var stopwatchCmd tea.Cmd
	m.stopwatch, stopwatchCmd = m.stopwatch.Update(msg)
	if stopwatchCmd != nil {
		cmds = append(cmds, stopwatchCmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the recording phase UI.
func (m Model) View() string {
	var sb strings.Builder

	// Recording indicator with spinner and stopwatch
	if m.IsRecording() {
		sb.WriteString(m.spinner.View())
		sb.WriteString(" ")
		sb.WriteString(style.Title.Render("Recording"))
		sb.WriteString(" ")
		sb.WriteString(style.Subtitle.Render(m.stopwatch.View()))
	} else {
		sb.WriteString(style.Warning.Render("Paused"))
		sb.WriteString(" ")
		sb.WriteString(style.Subtitle.Render(m.stopwatch.View()))
	}

	sb.WriteString("\n\n")

	// File size progress
	current, maxValue := m.controls.FileSize.Cap()
	percent := float64(0)
	if maxValue > 0 {
		percent = float64(current) / float64(maxValue)
	}

	sb.WriteString(m.progress.ViewAs(percent))
	sb.WriteString("\n")
	sb.WriteString(style.Subtitle.Render(formatBytes(current, maxValue)))
	sb.WriteString("\n\n")

	// Help text
	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render("space"))
	sb.WriteString(style.Help.Render("] start/pause  "))
	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render("enter"))
	sb.WriteString(style.Help.Render("] stop and transcribe  "))

	sb.WriteString(style.Help.Render("["))
	sb.WriteString(style.Key.Render("q"))
	sb.WriteString(style.Help.Render("] quit"))

	return sb.String()
}

// IsRecording returns whether recording is currently active.
func (m Model) IsRecording() bool {
	return m.controls.StartStopPause.Read()
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
