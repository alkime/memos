package workflow

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/require"
)

func TestRecordingPhase_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "recording.mp3")

	finishCalled := false
	controls := RecordingControls{
		FileSize:       &mockCappedDial{current: 1024, max: 10240},
		StartStopPause: &mockKnob{state: false},
		SampleLevels:   &mockLevels{samples: []int16{100, 200, 300}},
		Finish: func() {
			finishCalled = true
		},
	}

	phase := NewRecording(controls, 10240, outputPath)
	tm := teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Initial state - should show paused
	checker.checkString(t, tm, "Paused")

	// Toggle recording on (space key)
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	checker.checkString(t, tm, "Recording")

	// Toggle recording off
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	checker.checkString(t, tm, "Paused")

	// Finish recording (enter key)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for finish callback to be called
	require.Eventually(t, func() bool {
		return finishCalled
	}, 1*time.Second, 50*time.Millisecond, "Finish callback should be called")
}

func TestRecordingPhase_ExistingOutput(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "recording.mp3")

	// Create existing recording file
	//nolint:gosec // Test file
	require.NoError(t, os.WriteFile(outputPath, []byte("existing audio data"), 0o644))

	controls := RecordingControls{
		FileSize:       &mockCappedDial{current: 0, max: 10240},
		StartStopPause: &mockKnob{state: false},
		SampleLevels:   &mockLevels{samples: []int16{}},
		Finish:         func() {},
	}

	phase := NewRecording(controls, 10240, outputPath)
	tm := teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Should show existing output prompt
	checker.checkString(t, tm, "already exists")

	// Press enter to use existing - sends NextPhaseMsg (handled by container)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Test passes if we got to the existing output view
	// The actual phase transition is handled by the parent container
}
