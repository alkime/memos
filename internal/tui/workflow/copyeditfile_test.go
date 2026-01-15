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

func TestCopyEditFilePhase_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "my-post.md")

	// Create original file
	originalContent := "# Original Title\n\nOriginal content here."
	require.NoError(t, os.WriteFile(filePath, []byte(originalContent), 0o644)) //nolint:gosec // Test file

	writer := &mockWriter{
		copyEditResult: &content.CopyEditResult{
			Title:    "Polished Title",
			Markdown: "---\ntitle: Polished Title\n---\n\n# Polished Title\n\nPolished content here.",
			Changes:  []string{"Improved title", "Added frontmatter"},
		},
	}
	phase := NewCopyEditFilePhase(writer, filePath, content.ModeMemos)

	tm := teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Wait for review state to appear (mock returns immediately, skip spinner check)
	checker.checkString(t, tm, "Copy Edit Complete")

	// Verify writer was called
	assert.True(t, writer.copyEditCalled, "Writer.GenerateCopyEdit should be called")

	// Press enter to apply changes
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Should quit after applying
	tm.WaitFinished(t, teatest.WithFinalTimeout(checker.timeout))

	// Verify file was updated
	updatedContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Contains(t, string(updatedContent), "Polished Title")
	assert.Contains(t, string(updatedContent), "---") // Has frontmatter delimiter
}

func TestCopyEditFilePhase_Discard(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "my-post.md")

	// Create original file
	originalContent := "# Original Title\n\nOriginal content here."
	require.NoError(t, os.WriteFile(filePath, []byte(originalContent), 0o644)) //nolint:gosec // Test file

	writer := &mockWriter{
		copyEditResult: &content.CopyEditResult{
			Title:    "Polished Title",
			Markdown: "---\ntitle: Polished Title\n---\n\n# Polished Title\n\nPolished content here.",
			Changes:  []string{"Improved title"},
		},
	}
	phase := NewCopyEditFilePhase(writer, filePath, content.ModeMemos)

	tm := teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Wait for review state
	checker.checkString(t, tm, "Copy Edit Complete")

	// Press q to discard
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Should quit after discarding
	tm.WaitFinished(t, teatest.WithFinalTimeout(checker.timeout))

	// Verify original content was preserved
	preservedContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, originalContent, string(preservedContent), "Original content should be preserved when discarding")
}
