package workflow

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/require"
)

func TestViewTranscriptPhase_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.txt")

	// Create transcript file
	transcriptContent := "This is my test transcript. It contains several sentences about testing."
	//nolint:gosec // Test file
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcriptContent), 0o644))

	phase := NewViewTranscriptPhase(transcriptPath)
	tm := teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Send window size to trigger viewport loading
	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Should show transcript content after loading
	checker.checkString(t, tm, "Transcript")

	// Press enter to proceed - sends NextPhaseMsg (handled by parent container)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Test passes if we can view the transcript
	// The actual phase transition is handled by the parent container
}
