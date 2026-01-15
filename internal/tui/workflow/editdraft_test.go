package workflow

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditDraftPhase_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	draftPath := filepath.Join(tmpDir, "first-draft.md")

	launcher := &mockEditorLauncher{}
	phase := NewEditDraftPhase(launcher, draftPath)

	tm := teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Should show opening editor message
	checker.checkString(t, tm, "Opening editor")

	// Wait for editor to be launched (after the 250ms delay)
	require.Eventually(t, func() bool {
		return launcher.launched
	}, 1*time.Second, 50*time.Millisecond, "Editor launcher should be called")

	// Verify launcher was called with correct path
	assert.Equal(t, draftPath, launcher.filePath, "Editor should be launched with correct file path")
}
