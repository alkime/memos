// Package phases provides TUI phase implementations.
package phases

import (
	"fmt"
	"strings"

	"github.com/alkime/memos/internal/tui/components/phases"
	"github.com/alkime/memos/internal/tui/phase/msg"
	"github.com/alkime/memos/internal/tui/style"
	"github.com/alkime/memos/pkg/uictl"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
)

// RecordingControls provides read/write access to recording hardware.
type RecordingControls struct {
	FileSize       uictl.CappedDial[int64]
	StartStopPause uictl.Knob
	Finish         func()
}

// recordingKeyMap defines the key bindings for the recording phase.
type recordingKeyMap struct {
	Toggle key.Binding
	Finish key.Binding
}

func defaultRecordingKeyMap() recordingKeyMap {
	return recordingKeyMap{
		Toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "start/stop recording"),
		),
		Finish: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "finish recording"),
		),
	}
}

// Recording represents the recording phase UI state.
type Recording struct {
	keys      recordingKeyMap
	controls  RecordingControls
	spinner   spinner.Model
	stopwatch stopwatch.Model
	progress  progress.Model
	maxBytes  int64
}

// NewRecording creates a new recording phase model.
func NewRecording(controls RecordingControls, maxBytes int64) *Recording {
	s := spinner.New()
	s.Spinner = spinner.Points

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return &Recording{
		keys:      defaultRecordingKeyMap(),
		controls:  controls,
		spinner:   s,
		stopwatch: stopwatch.New(),
		progress:  p,
		maxBytes:  maxBytes,
	}
}

// Init returns the initial command for the recording phase.
func (r *Recording) Init() tea.Cmd {
	return r.spinner.Tick
}

// Update handles messages for the recording phase.
func (r *Recording) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch typedMsg := teaMsg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(typedMsg, r.keys.Toggle):
			r.controls.StartStopPause.Toggle()
			if r.IsRecording() {
				cmds = append(cmds, r.stopwatch.Start())
			} else {
				cmds = append(cmds, r.stopwatch.Stop())
			}

			return r, tea.Batch(cmds...)

		case key.Matches(typedMsg, r.keys.Finish):
			if r.controls.Finish != nil {
				r.controls.Finish()
			}

			return r, nil
		}

	case msg.AudioFinalizingCompleteMsg:
		return r, func() tea.Msg { return phases.NextPhaseMsg{} }
	case spinner.TickMsg:
		var cmd tea.Cmd
		r.spinner, cmd = r.spinner.Update(typedMsg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := r.progress.Update(typedMsg)
		r.progress = progressModel.(progress.Model) //nolint:forcetypeassert // bubbles library contract
		cmds = append(cmds, cmd)
	}

	// Always update stopwatch
	var stopwatchCmd tea.Cmd
	r.stopwatch, stopwatchCmd = r.stopwatch.Update(teaMsg)
	if stopwatchCmd != nil {
		cmds = append(cmds, stopwatchCmd)
	}

	return r, tea.Batch(cmds...)
}

// View renders the recording phase UI.
func (r *Recording) View() string {
	var sb strings.Builder

	// Recording indicator with spinner and stopwatch
	if r.IsRecording() {
		sb.WriteString(r.spinner.View())
		sb.WriteString(" ")
		sb.WriteString(style.Title.Render("Recording"))
		sb.WriteString(" ")
		sb.WriteString(style.Subtitle.Render(r.stopwatch.View()))
	} else {
		sb.WriteString(style.Warning.Render("Paused"))
		sb.WriteString(" ")
		sb.WriteString(style.Subtitle.Render(r.stopwatch.View()))
	}

	sb.WriteString("\n\n")

	// File size progress
	current, maxValue := r.controls.FileSize.Cap()
	percent := float64(0)
	if maxValue > 0 {
		percent = float64(current) / float64(maxValue)
	}

	sb.WriteString(r.progress.ViewAs(percent))
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
func (r *Recording) IsRecording() bool {
	return r.controls.StartStopPause.Read()
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
