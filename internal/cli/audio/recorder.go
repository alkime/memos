package audio

import (
	"time"

	_ "github.com/gen2brain/malgo" // Audio recording library
)

// Recorder handles audio recording from microphone
type Recorder struct {
	sampleRate  uint32
	channels    uint32
	outputPath  string
	maxDuration time.Duration
	maxBytes    int64
}

// NewRecorder creates a new audio recorder
func NewRecorder(outputPath string, maxDuration time.Duration, maxBytes int64) *Recorder {
	return &Recorder{
		sampleRate:  16000, // Whisper native sample rate
		channels:    1,     // Mono
		outputPath:  outputPath,
		maxDuration: maxDuration,
		maxBytes:    maxBytes,
	}
}
