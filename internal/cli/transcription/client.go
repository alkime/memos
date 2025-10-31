package transcription

import (
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
