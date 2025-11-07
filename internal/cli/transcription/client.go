package transcription

import (
	"context"
	"errors"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

// Client handles Whisper API transcription requests.
type Client struct {
	apiKey string
}

// NewClient creates a new transcription client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

// TranscribeFile transcribes an audio file using Whisper API.
func (c *Client) TranscribeFile(audioFile io.Reader) (string, error) {
	// Validate API key
	if c.apiKey == "" {
		return "", errors.New("API key required: set OPENAI_API_KEY or use --api-key")
	}

	// Create OpenAI client
	client := openai.NewClient(c.apiKey)

	// Create transcription request
	req := openai.AudioRequest{ //nolint:exhaustruct // Only Model and Reader required
		Model:  openai.Whisper1,
		Reader: audioFile,
	}

	// Call Whisper API
	ctx := context.Background()
	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create transcription via Whisper API: %w", err)
	}

	return resp.Text, nil
}
