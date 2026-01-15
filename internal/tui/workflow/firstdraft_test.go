package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alkime/memos/internal/content"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirstDraftPhase_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.txt")
	outputPath := filepath.Join(tmpDir, "first-draft.md")

	// Create transcript file
	//nolint:gosec // Test file
	require.NoError(t, os.WriteFile(transcriptPath, []byte("This is my transcript content."), 0o644))

	writer := &mockWriter{firstDraftResult: "# My First Draft\n\nThis is the generated content."}
	phase := NewFirstDraftPhase(writer, transcriptPath, outputPath, content.ModeMemos)

	tm := teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Should show generating spinner
	checker.checkString(t, tm, "Generating first draft")

	// Eventually the generation completes - verify by checking file creation
	require.Eventually(t, func() bool {
		_, err := os.Stat(outputPath)
		return err == nil
	}, checker.timeout, checker.intervl, "Draft file should be created")

	// Verify writer was called
	assert.True(t, writer.firstDraftCalled, "Writer.GenerateFirstDraft should be called")

	// Verify output file was created
	generatedContent, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "# My First Draft\n\nThis is the generated content.", string(generatedContent))
}

func TestFirstDraftPhase_ExistingOutput(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptPath := filepath.Join(tmpDir, "transcript.txt")
	outputPath := filepath.Join(tmpDir, "first-draft.md")

	// Create existing files
	//nolint:gosec // Test files
	require.NoError(t, os.WriteFile(transcriptPath, []byte("transcript"), 0o644))
	//nolint:gosec // Test files
	require.NoError(t, os.WriteFile(outputPath, []byte("existing draft content"), 0o644))

	writer := &mockWriter{firstDraftResult: "new draft"}
	phase := NewFirstDraftPhase(writer, transcriptPath, outputPath, content.ModeMemos)

	tm := teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Should show existing output prompt
	checker.checkString(t, tm, "already exists")

	// Press enter to use existing
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify writer was NOT called
	assert.False(t, writer.firstDraftCalled, "Writer should not be called when using existing")

	// Verify original content was preserved
	existingContent, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "existing draft content", string(existingContent))
}
