package transcription

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
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
	client := openai.NewClient(option.WithAPIKey(c.apiKey))

	// Create transcription request
	params := openai.AudioTranscriptionNewParams{
		File:  audioFile,
		Model: openai.AudioModelWhisper1,
	}

	// Call Whisper API
	ctx := context.Background()
	resp, err := client.Audio.Transcriptions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create transcription via Whisper API: %w", err)
	}

	return resp.Text, nil
}
