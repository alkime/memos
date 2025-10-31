package content

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	contentDir := "content/posts"

	generator := NewGenerator(contentDir)

	assert.NotNil(t, generator)
	assert.Equal(t, contentDir, generator.contentDir)
}

func TestGenerator_GeneratePost_Frontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content/posts")
	err := os.MkdirAll(contentDir, 0755)
	require.NoError(t, err)

	// Create test transcript
	transcriptPath := filepath.Join(tmpDir, "test-transcript.txt")
	transcriptText := "This is a test transcript."
	err = os.WriteFile(transcriptPath, []byte(transcriptText), 0644)
	require.NoError(t, err)

	generator := NewGenerator(contentDir)
	outputPath := filepath.Join(contentDir, "test-post.md")

	err = generator.GeneratePost(transcriptPath, outputPath)

	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)

	// Verify frontmatter structure
	assert.Contains(t, contentStr, "---")
	assert.Contains(t, contentStr, "title:")
	assert.Contains(t, contentStr, "date:")
	assert.Contains(t, contentStr, "draft: true")

	// Verify body content
	assert.Contains(t, contentStr, transcriptText)
}

func TestGenerator_GeneratePost_Archives(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content/posts")
	recordingsDir := filepath.Join(tmpDir, ".memos/recordings")
	archiveDir := filepath.Join(tmpDir, ".memos/archive")

	err := os.MkdirAll(contentDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(recordingsDir, 0755)
	require.NoError(t, err)

	// Create test files
	wavPath := filepath.Join(recordingsDir, "2025-10-31-143052.wav")
	transcriptPath := filepath.Join(recordingsDir, "2025-10-31-143052.txt")
	transcriptText := "Test transcript"

	err = os.WriteFile(wavPath, []byte("fake wav data"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(transcriptPath, []byte(transcriptText), 0644)
	require.NoError(t, err)

	generator := NewGenerator(contentDir)
	outputPath := filepath.Join(contentDir, "test-post.md")

	err = generator.GeneratePost(transcriptPath, outputPath)

	require.NoError(t, err)

	// Verify files moved to archive
	archivedWav := filepath.Join(archiveDir, "2025-10-31-143052.wav")
	archivedTxt := filepath.Join(archiveDir, "2025-10-31-143052.txt")

	assert.FileExists(t, archivedWav)
	assert.FileExists(t, archivedTxt)
	assert.NoFileExists(t, wavPath)
	assert.NoFileExists(t, transcriptPath)
}
