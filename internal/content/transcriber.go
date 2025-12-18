package content

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Transcriber handles Whisper API transcription requests.
type Transcriber struct {
	apiKey string
}

// NewTranscriber creates a new transcription client.
func NewTranscriber(apiKey string) *Transcriber {
	return &Transcriber{
		apiKey: apiKey,
	}
}

// TranscribeFile transcribes an audio file using Whisper API.
func (t *Transcriber) TranscribeFile(audioFile io.Reader) (string, error) {
	// Validate API key
	if t.apiKey == "" {
		return "", errors.New("API key required: set OPENAI_API_KEY or use --api-key")
	}

	// Create OpenAI client
	client := openai.NewClient(option.WithAPIKey(t.apiKey))

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
