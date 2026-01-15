package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alkime/memos/internal/content"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyEditPhase_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "first-draft.md")
	outputDir := filepath.Join(tmpDir, "posts")

	// Create input draft file
	//nolint:gosec // Test file
	require.NoError(t, os.WriteFile(inputPath, []byte("# Draft Content\n\nThis is my draft."), 0o644))
	require.NoError(t, os.MkdirAll(outputDir, 0o755))

	writer := &mockWriter{
		copyEditResult: &content.CopyEditResult{
			Title:    "My Final Post",
			Markdown: "---\ntitle: My Final Post\n---\n\n# My Final Post\n\nPolished content here.",
			Changes:  []string{"Fixed grammar", "Added frontmatter"},
		},
	}
	phase := NewCopyEditPhase(writer, inputPath, content.ModeMemos, outputDir)

	_ = teatest.NewTestModel(t, phase, teatest.WithInitialTermSize(80, 24))
	checker := defaultChecker()

	// Wait for completion and verify output file is created
	// (the mock returns immediately so we skip checking for spinner)
	require.Eventually(t, func() bool {
		files, _ := os.ReadDir(outputDir)
		return len(files) > 0
	}, checker.timeout, checker.intervl, "Output file should be created")

	// Verify writer was called
	assert.True(t, writer.copyEditCalled, "Writer.GenerateCopyEdit should be called")

	// Verify output file was created with expected slug pattern
	files, err := os.ReadDir(outputDir)
	require.NoError(t, err)
	require.Len(t, files, 1, "Should create exactly one output file")

	// File should have date prefix and slug
	fileName := files[0].Name()
	assert.True(t, strings.HasSuffix(fileName, ".md"), "Output should be markdown file")
	assert.True(t, strings.Contains(fileName, "my-final-post"), "Filename should contain slug")

	// Verify file content
	outputContent, err := os.ReadFile(filepath.Join(outputDir, fileName))
	require.NoError(t, err)
	assert.Contains(t, string(outputContent), "My Final Post")
}
