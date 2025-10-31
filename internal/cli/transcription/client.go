package transcription

import (
	"errors"
	"os"

	_ "github.com/sashabaranov/go-openai" // OpenAI API client
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

	// TODO: Actual API call implementation
	return "", errors.New("not implemented")
}
