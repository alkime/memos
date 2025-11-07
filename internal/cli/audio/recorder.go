package audio

import (
	"errors"
	"fmt"
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
//
// TODO: Implement actual audio recording functionality using malgo library.
// Current implementation is a stub that only creates the output directory.
// Future implementation will:
//   - Initialize malgo device context
//   - Configure audio capture device (16kHz mono for Whisper)
//   - Start recording to WAV file
//   - Handle maxDuration and maxBytes limits
//   - Manage recording state and cleanup
func (r *Recorder) Start() error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(r.outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", dir, err)
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
