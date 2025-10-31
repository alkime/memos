package audio

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	_ "github.com/gen2brain/malgo" // Audio recording library
)

// Recorder handles audio recording from microphone.
type Recorder struct {
	sampleRate  uint32
	channels    uint32
	outputPath  string
	maxDuration time.Duration
	maxBytes    int64
	state       *recordingState
}

// recordingState tracks the state of an active recording.
type recordingState struct {
	isRecording bool
}

// NewRecorder creates a new audio recorder.
func NewRecorder(outputPath string, maxDuration time.Duration, maxBytes int64) *Recorder {
	return &Recorder{ //nolint:exhaustruct // state initialized on Start()
		sampleRate:  16000, // Whisper native sample rate
		channels:    1,     // Mono
		outputPath:  outputPath,
		maxDuration: maxDuration,
		maxBytes:    maxBytes,
	}
}

// Start initializes the recorder and creates output directory.
func (r *Recorder) Start() error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(r.outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err //nolint:wrapcheck // Clear error from MkdirAll
	}

	r.state = &recordingState{isRecording: true}
	return nil
}

// Stop stops the recording.
func (r *Recorder) Stop() error {
	if r.state == nil || !r.state.isRecording {
		return errors.New("recorder not started")
	}
	r.state.isRecording = false
	return nil
}
