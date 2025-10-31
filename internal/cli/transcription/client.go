package transcription

import (
	"context"
	"errors"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// Client handles Whisper API transcription requests
type Client struct {
	apiKey string
}

// NewClient creates a new transcription client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

// TranscribeFile transcribes an audio file using Whisper API
func (c *Client) TranscribeFile(audioPath string) (string, error) {
	// Validate API key
	if c.apiKey == "" {
		return "", errors.New("API key required: set OPENAI_API_KEY or use --api-key")
	}

	// Validate file exists
	info, err := os.Stat(audioPath)
	if err != nil {
		return "", err
	}

	// Validate file is not empty
	if info.Size() == 0 {
		return "", errors.New("audio file is empty")
	}

	// Create OpenAI client
	client := openai.NewClient(c.apiKey)

	// Create transcription request
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: audioPath,
	}

	// Call Whisper API
	ctx := context.Background()
	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}
