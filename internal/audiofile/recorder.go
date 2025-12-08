package audiofile

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	mp3encoder "github.com/braheezy/shine-mp3/pkg/mp3"
)

// Recorder reads raw PCM audio data from a channel and buffers it to disk.
// After recording completes, it can convert the buffered PCM to MP3 format.
type Recorder struct {
	sampleRate int
	channels   int
	input      <-chan []byte
	pcmPath    string
	mp3Path    string

	pcmFile      *os.File
	bytesWritten int64
	mu           sync.RWMutex
	wg           sync.WaitGroup
	errOnce      sync.Once
	err          error
}

// Config holds configuration for the audio recorder.
type Config struct {
	SampleRate int    // Sample rate in Hz (e.g., 16000)
	Channels   int    // Number of channels (1 for mono, 2 for stereo)
	MP3Path    string // Final MP3 output path
}

// NewRecorder creates a new audio file recorder.
//
// Parameters:
//   - config: Recording configuration (sample rate, channels, output path)
//   - input: Channel of raw PCM bytes (S16LE format)
//
// Returns error if parameters are invalid.
func NewRecorder(config Config, input <-chan []byte) (*Recorder, error) {
	if input == nil {
		return nil, errors.New("input channel cannot be nil")
	}

	if config.SampleRate <= 0 {
		return nil, errors.New("sample rate must be positive")
	}

	if config.Channels <= 0 {
		return nil, errors.New("channels must be positive")
	}

	if config.MP3Path == "" {
		return nil, errors.New("MP3 path cannot be empty")
	}

	// Create temporary PCM file path
	pcmPath := config.MP3Path + ".tmp.pcm"

	return &Recorder{ //nolint:exhaustruct // wg, errOnce, err initialized later
		sampleRate: config.SampleRate,
		channels:   config.Channels,
		input:      input,
		pcmPath:    pcmPath,
		mp3Path:    config.MP3Path,
	}, nil
}

// Start begins recording PCM data from the input channel to disk.
// Must be called before any data is sent to the input channel.
func (r *Recorder) Start(ctx context.Context) error {
	if r.pcmFile != nil {
		return errors.New("recorder already started")
	}

	// Create temporary PCM file
	pcmFile, err := os.Create(r.pcmPath)
	if err != nil {
		return fmt.Errorf("failed to create PCM file %s: %w", r.pcmPath, err)
	}

	r.pcmFile = pcmFile

	r.wg.Go(func() {
		defer func() {
			// Close PCM file
			if err := r.pcmFile.Close(); err != nil {
				r.setError(fmt.Errorf("failed to close PCM file: %w", err))
				return
			}

			// Convert PCM to MP3
			slog.Info("converting PCM to MP3")
			if err := r.convertToMP3(); err != nil {
				r.setError(fmt.Errorf("failed to convert to MP3: %w", err))
				return
			}

			// Cleanup temporary PCM file
			if err := r.cleanup(); err != nil {
				slog.Warn("failed to cleanup temporary PCM file", "error", err)
			}

			slog.Info("recording complete", "output", r.mp3Path)
		}()

		for {
			select {
			case data, ok := <-r.input:
				if !ok {
					// Channel closed, finish recording
					return
				}

				n, err := r.pcmFile.Write(data)
				if err != nil {
					r.setError(fmt.Errorf("failed to write PCM data: %w", err))
					return
				}

				// Track bytes written
				r.mu.Lock()
				r.bytesWritten += int64(n)
				r.mu.Unlock()

			case <-ctx.Done():
				return
			}
		}
	})

	return nil
}

// Wait blocks until recording completes (including conversion and cleanup).
// Returns any error that occurred during the entire process.
func (r *Recorder) Wait() error {
	r.wg.Wait()
	return r.err
}

// convertToMP3 converts the buffered PCM data to MP3 format.
// This is called automatically by the recording goroutine.
func (r *Recorder) convertToMP3() error {
	// Read raw PCM bytes
	pcmData, err := os.ReadFile(r.pcmPath)
	if err != nil {
		return fmt.Errorf("failed to read PCM file: %w", err)
	}

	// Convert bytes to int16 samples (S16LE format)
	numSamples := len(pcmData) / 2
	monoSamples := make([]int16, numSamples)

	reader := bytes.NewReader(pcmData)
	if err := binary.Read(reader, binary.LittleEndian, monoSamples); err != nil {
		return fmt.Errorf("failed to read PCM samples: %w", err)
	}

	// Convert mono to stereo (shine-mp3 works better with stereo)
	var samples []int16
	channels := r.channels

	if r.channels == 1 {
		samples = make([]int16, numSamples*2)
		for i, sample := range monoSamples {
			samples[i*2] = sample   // Left
			samples[i*2+1] = sample // Right (duplicate)
		}
		channels = 2
	} else {
		samples = monoSamples
	}

	slog.Info("converting PCM to MP3",
		"pcmPath", r.pcmPath,
		"mp3Path", r.mp3Path,
		"samples", numSamples,
		"sampleRate", r.sampleRate,
		"channels", channels)

	// Create MP3 encoder
	encoder := mp3encoder.NewEncoder(r.sampleRate, channels)

	// Create output file
	mp3File, err := os.Create(r.mp3Path)
	if err != nil {
		return fmt.Errorf("failed to create MP3 file %s: %w", r.mp3Path, err)
	}
	defer mp3File.Close()

	// Encode to MP3 in one shot (not streaming)
	if err := encoder.Write(mp3File, samples); err != nil {
		return fmt.Errorf("failed to encode MP3: %w", err)
	}

	return nil
}

// cleanup removes the temporary PCM file.
// This is called automatically by the recording goroutine.
func (r *Recorder) cleanup() error {
	if err := os.Remove(r.pcmPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove temporary PCM file %s: %w", r.pcmPath, err)
	}

	return nil
}

// GetPCMPath returns the path to the temporary PCM file (for TUI display).
func (r *Recorder) GetPCMPath() string {
	return r.pcmPath
}

// BytesWritten returns the number of bytes written to the PCM file.
// This method is safe to call concurrently from multiple goroutines.
func (r *Recorder) BytesWritten() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.bytesWritten
}

// setError records the first error that occurs (subsequent calls are no-ops).
func (r *Recorder) setError(err error) {
	r.errOnce.Do(func() {
		r.err = err
		slog.Error("audio recorder error", "error", err)
	})
}
