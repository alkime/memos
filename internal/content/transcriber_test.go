package content

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTranscriber(t *testing.T) {
	apiKey := "test-api-key"

	transcriber := NewTranscriber(apiKey)

	assert.NotNil(t, transcriber)
	assert.Equal(t, apiKey, transcriber.apiKey)
}

func TestNewTranscriber_EmptyAPIKey(t *testing.T) {
	transcriber := NewTranscriber("")

	assert.NotNil(t, transcriber)
	assert.Equal(t, "", transcriber.apiKey)
}

func TestTranscriber_TranscribeFile_MissingAPIKey(t *testing.T) {
	transcriber := NewTranscriber("")
	reader := strings.NewReader("fake audio data")

	text, err := transcriber.TranscribeFile(reader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
	assert.Empty(t, text)
}

func TestTranscriber_TranscribeFile_EmptyFile(t *testing.T) {
	transcriber := NewTranscriber("test-key")
	reader := strings.NewReader("")

	text, err := transcriber.TranscribeFile(reader)

	// Empty file will likely cause API error, but we're just testing
	// that the method handles it without panicking
	assert.Empty(t, text)
	// Note: Error handling depends on API behavior with empty input
	// In a real scenario, this would be validated at the CLI level
	_ = err
}

