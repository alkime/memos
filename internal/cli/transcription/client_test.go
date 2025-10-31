package transcription

import (
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
