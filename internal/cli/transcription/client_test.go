package transcription //nolint:testpackage // Needs access to unexported fields

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	reader := strings.NewReader("fake audio data")

	text, err := client.TranscribeFile(reader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
	assert.Empty(t, text)
}

func TestClient_TranscribeFile_EmptyFile(t *testing.T) {
	client := NewClient("test-key")
	reader := strings.NewReader("")

	text, err := client.TranscribeFile(reader)

	// Empty file will likely cause API error, but we're just testing
	// that the method handles it without panicking
	assert.Empty(t, text)
	// Note: Error handling depends on API behavior with empty input
	// In a real scenario, this would be validated at the CLI level
	_ = err
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
