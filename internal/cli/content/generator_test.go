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
