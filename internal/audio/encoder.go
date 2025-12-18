package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"

	mp3encoder "github.com/braheezy/shine-mp3/pkg/mp3"
)

// StreamingEncoder reads raw PCM bytes from a channel, buffers to a threshold,
// then batch-encodes to MP3 and writes to an io.Writer.
//
// The encoder runs in a goroutine and handles graceful shutdown when the input
// channel is closed or the context is cancelled.
type StreamingEncoder struct {
	config EncoderConfig
	input  <-chan []byte
	output io.Writer

	encoder *mp3encoder.Encoder
	buffer  []byte

	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

// NewStreamingEncoder creates a new streaming MP3 encoder.
//
// Parameters:
//   - config: Encoder configuration (sample rate, channels, buffer threshold)
//   - input: Channel of raw PCM bytes (S16LE format)
//   - output: Writer where MP3 frames are written
//
// Returns error if config is invalid or parameters are nil.
func NewStreamingEncoder(
	config EncoderConfig,
	input <-chan []byte,
	output io.Writer,
) (*StreamingEncoder, error) {
	if input == nil {
		return nil, errors.New("input channel cannot be nil")
	}

	if output == nil {
		return nil, errors.New("output writer cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid encoder config: %w", err)
	}

	return &StreamingEncoder{ //nolint:exhaustruct // wg, errOnce, err initialized on Start()
		config: config,
		input:  input,
		output: output,
		buffer: make([]byte, 0, config.BufferThreshold),
	}, nil
}

// Start begins the encoding goroutine. Must be called before any data is sent
// to the input channel. Returns error if already started.
func (e *StreamingEncoder) Start(ctx context.Context) error {
	if e.encoder != nil {
		return errors.New("encoder already started")
	}

	// Create shine-mp3 encoder as STEREO (workaround for mono bug)
	e.encoder = mp3encoder.NewEncoder(e.config.SampleRate, 2)

	// todo: figure out logging w/ tui bubbletea...
	// slog.Info("starting MP3 encoder",
	// 	"sampleRate", e.config.SampleRate,
	// 	"channels", e.config.Channels,
	// 	"bufferThreshold", e.config.BufferThreshold)

	e.wg.Go(func() {
		defer func() {
			if err := e.Flush(); err != nil {
				e.setError(fmt.Errorf("failed to flush encoder on shutdown: %w", err))
			}
		}()

		for {
			select {
			case data, ok := <-e.input:
				if !ok {
					// Channel closed, finish encoding
					return
				}

				// Append to buffer
				e.buffer = append(e.buffer, data...)

				// Encode when threshold reached
				if len(e.buffer) >= e.config.BufferThreshold {
					if err := e.encodeBatch(); err != nil {
						e.setError(err)
						return
					}
				}

			case <-ctx.Done():
				e.setError(fmt.Errorf("encoder context cancelled: %w", ctx.Err()))
				return
			}
		}
	})

	return nil
}

// encodeBatch converts buffered PCM data to MP3 and writes to output.
// Clears the buffer after successful encoding.
func (e *StreamingEncoder) encodeBatch() error {
	if len(e.buffer) == 0 {
		return nil
	}

	// Convert buffer bytes to []int16 (S16LE PCM)
	numSamples := len(e.buffer) / 2 // 2 bytes per int16 sample
	monoSamples := make([]int16, numSamples)

	reader := bytes.NewReader(e.buffer)

	if err := binary.Read(reader, binary.LittleEndian, monoSamples); err != nil {
		return fmt.Errorf("failed to read PCM samples: %w", err)
	}

	// WORKAROUND: shine-mp3 Write() has a bug for mono (always increments by samples_per_pass * 2)
	// Convert mono to stereo by duplicating samples (L=R)
	stereoSamples := make([]int16, numSamples*2)
	for i, sample := range monoSamples {
		stereoSamples[i*2] = sample   // Left channel
		stereoSamples[i*2+1] = sample // Right channel (duplicate)
	}

	slog.Debug("encoding MP3 batch",
		"monoSamples", numSamples,
		"stereoSamples", len(stereoSamples))

	// Encode to MP3 and write to output
	if err := e.encoder.Write(e.output, stereoSamples); err != nil {
		return fmt.Errorf("failed to encode audio to MP3: %w", err)
	}

	// Clear buffer (reuse allocated memory)
	e.buffer = e.buffer[:0]

	return nil
}

// Flush encodes any remaining buffered data. Safe to call multiple times.
func (e *StreamingEncoder) Flush() error {
	if err := e.encodeBatch(); err != nil {
		return fmt.Errorf("failed to flush MP3 encoder: %w", err)
	}

	return nil
}

// Wait blocks until encoding completes and returns any error that occurred.
func (e *StreamingEncoder) Wait() error {
	e.wg.Wait()

	return e.err
}

// setError records the first error that occurs (subsequent calls are no-ops).
func (e *StreamingEncoder) setError(err error) {
	e.errOnce.Do(func() {
		e.err = err
		slog.Debug("streaming encoder error", "error", err)
	})
}
