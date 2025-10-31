package transcription

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"

	client := NewClient(apiKey)

	assert.NotNil(t, client)
	assert.Equal(t, apiKey, client.apiKey)
}

func TestNewClient_EmptyAPIKey(t *testing.T) {
	client := NewClient("")

	assert.NotNil(t, client)
	assert.Equal(t, "", client.apiKey)
}

func TestClient_TranscribeFile_MissingAPIKey(t *testing.T) {
	client := NewClient("")

	text, err := client.TranscribeFile("/tmp/test.wav")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
	assert.Empty(t, text)
}

func TestClient_TranscribeFile_FileNotFound(t *testing.T) {
	client := NewClient("test-key")

	text, err := client.TranscribeFile("/nonexistent/file.wav")

	assert.Error(t, err)
	assert.Empty(t, text)
}

func TestClient_TranscribeFile_EmptyFile(t *testing.T) {
	// Create empty temp file
	tmpFile, err := os.CreateTemp("", "test-*.wav")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := NewClient("test-key")

	text, err := client.TranscribeFile(tmpFile.Name())

	// Should handle empty files gracefully
	assert.Error(t, err)
	assert.Empty(t, text)
}

func TestClient_TranscribeFile_ValidFile(t *testing.T) {
	// Skip if no API key set
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	// Skip - requires valid audio file for real test
	t.Skip("Requires valid audio file - run manually with real audio")
}
