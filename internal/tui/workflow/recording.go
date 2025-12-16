// Package workflow provides TUI workflow step implementations.
package workflow

import (
	"fmt"
	"strings"

	"github.com/alkime/memos/internal/tui/components/phases"
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

// recordingPhase represents the recording phase UI state.
type recordingPhase struct {
	keys           recordingKeyMap
	controls       RecordingControls
	spinner        spinner.Model
	stopwatch      stopwatch.Model
	progress       progress.Model
	maxBytes       int64
	outputPath     string
	existingOutput existingOutputState
}

// NewRecording creates a new recording phase model.
func NewRecording(controls RecordingControls, maxBytes int64, outputPath string) tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Points

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return &recordingPhase{
		keys:           defaultRecordingKeyMap(),
		controls:       controls,
		spinner:        s,
		stopwatch:      stopwatch.New(),
		progress:       p,
		maxBytes:       maxBytes,
		outputPath:     outputPath,
		existingOutput: newExistingOutputState(outputPath),
	}
}

// Init returns the initial command for the recording phase.
func (r *recordingPhase) Init() tea.Cmd {
	return r.spinner.Tick
}

// Update handles messages for the recording phase.
func (r *recordingPhase) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch typedMsg := teaMsg.(type) {
	case tea.KeyMsg:
		// Handle existing output keybindings first
		if r.existingOutput.found {
			switch {
			case key.Matches(typedMsg, r.existingOutput.keys.UseExisting):
				return r, phases.NextPhaseCmd
			case key.Matches(typedMsg, r.existingOutput.keys.Redo):
				r.existingOutput.found = false

				return r, r.spinner.Tick
			}

			return r, nil
		}

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

	case AudioFinalizingCompleteMsg:
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

	// Always update stopwatch (but only if not showing existing output)
	if !r.existingOutput.found {
		var stopwatchCmd tea.Cmd
		r.stopwatch, stopwatchCmd = r.stopwatch.Update(teaMsg)
		if stopwatchCmd != nil {
			cmds = append(cmds, stopwatchCmd)
		}
	}

	return r, tea.Batch(cmds...)
}

// View renders the recording phase UI.
func (r *recordingPhase) View() string {
	// Show existing output view if recording already exists
	if r.existingOutput.found {
		return renderExistingOutputView(r.existingOutput, "Recording")
	}

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
	sb.WriteString(renderKeyHelp(r.keys.Toggle, " "))
	sb.WriteString(renderKeyHelp(r.keys.Finish, "\n"))
	sb.WriteString(renderGlobalKeyHelp())

	return sb.String()
}

// IsRecording returns whether recording is currently active.
func (r *recordingPhase) IsRecording() bool {
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
