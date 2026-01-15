package workflow

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranscribePhase_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "recording.mp3")
	transcriptPath := filepath.Join(tmpDir, "transcript.txt")

	// Create fake audio file
	//nolint:gosec // Test file
	require.NoError(t, os.WriteFile(audioPath, []byte("fake audio data"), 0o644))

	transcriber := &mockTranscriber{result: "Hello world, this is my transcript."}
	phase := NewTranscribePhase(transcriber, audioPath, transcriptPath)

	tm := teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Should show transcribing spinner initially
	checker.checkString(t, tm, "Transcribing")

	// Eventually the transcription completes - verify by checking file creation
	// (the phase will emit NextPhaseMsg which we can't catch in isolated testing)
	require.Eventually(t, func() bool {
		_, err := os.Stat(transcriptPath)
		return err == nil
	}, checker.timeout, checker.intervl, "Transcript file should be created")

	// Verify transcriber was called
	assert.True(t, transcriber.called, "Transcriber should be called")

	// Verify output file was created with correct content
	content, err := os.ReadFile(transcriptPath)
	require.NoError(t, err)
	assert.Equal(t, "Hello world, this is my transcript.", string(content))
}

func TestTranscribePhase_ExistingOutput(t *testing.T) {
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "recording.mp3")
	transcriptPath := filepath.Join(tmpDir, "transcript.txt")

	// Create existing transcript file
	//nolint:gosec // Test files
	require.NoError(t, os.WriteFile(audioPath, []byte("fake audio"), 0o644))
	//nolint:gosec // Test files
	require.NoError(t, os.WriteFile(transcriptPath, []byte("existing transcript"), 0o644))

	transcriber := &mockTranscriber{result: "new transcript"}
	phase := NewTranscribePhase(transcriber, audioPath, transcriptPath)

	tm := teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Should show existing output prompt
	checker.checkString(t, tm, "already exists")

	// Press enter to use existing - this sends NextPhaseMsg (not Quit)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify transcriber was NOT called (using existing)
	assert.False(t, transcriber.called, "Transcriber should not be called when using existing")

	// Verify original content was preserved
	content, err := os.ReadFile(transcriptPath)
	require.NoError(t, err)
	assert.Equal(t, "existing transcript", string(content))
}
