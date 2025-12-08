package mp3_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/alkime/memos/internal/mp3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoderConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      mp3.EncoderConfig
		expectError string
	}{
		{
			name: "valid config",
			config: mp3.EncoderConfig{
				SampleRate:      16000,
				Channels:        1,
				BufferThreshold: 4096,
			},
			expectError: "",
		},
		{
			name: "zero sample rate",
			config: mp3.EncoderConfig{
				SampleRate:      0,
				Channels:        1,
				BufferThreshold: 4096,
			},
			expectError: "sample rate must be positive",
		},
		{
			name: "invalid channels",
			config: mp3.EncoderConfig{
				SampleRate:      16000,
				Channels:        2,
				BufferThreshold: 4096,
			},
			expectError: "only mono (1 channel) is supported",
		},
		{
			name: "zero buffer threshold",
			config: mp3.EncoderConfig{
				SampleRate:      16000,
				Channels:        1,
				BufferThreshold: 0,
			},
			expectError: "buffer threshold must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()

			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEncoderConfig_WithDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    mp3.EncoderConfig
		expected mp3.EncoderConfig
	}{
		{
			name:  "empty config gets all defaults",
			input: mp3.EncoderConfig{},
			expected: mp3.EncoderConfig{
				SampleRate:      mp3.DefaultSampleRate,
				Channels:        mp3.DefaultChannels,
				BufferThreshold: mp3.DefaultBufferThreshold,
			},
		},
		{
			name: "partial config preserves custom values",
			input: mp3.EncoderConfig{
				SampleRate: 44100,
			},
			expected: mp3.EncoderConfig{
				SampleRate:      44100,
				Channels:        mp3.DefaultChannels,
				BufferThreshold: mp3.DefaultBufferThreshold,
			},
		},
		{
			name: "complete config unchanged",
			input: mp3.EncoderConfig{
				SampleRate:      48000,
				Channels:        1,
				BufferThreshold: 8192,
			},
			expected: mp3.EncoderConfig{
				SampleRate:      48000,
				Channels:        1,
				BufferThreshold: 8192,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.input.WithDefaults()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewStreamingEncoder_ValidatesInputs(t *testing.T) {
	t.Parallel()

	validConfig := mp3.EncoderConfig{
		SampleRate:      16000,
		Channels:        1,
		BufferThreshold: 4096,
	}

	tests := []struct {
		name        string
		config      mp3.EncoderConfig
		input       <-chan []byte
		output      io.Writer
		expectError string
	}{
		{
			name:        "valid inputs",
			config:      validConfig,
			input:       make(chan []byte),
			output:      bytes.NewBuffer(nil),
			expectError: "",
		},
		{
			name: "invalid config",
			config: mp3.EncoderConfig{
				SampleRate:      0,
				Channels:        1,
				BufferThreshold: 4096,
			},
			input:       make(chan []byte),
			output:      bytes.NewBuffer(nil),
			expectError: "invalid encoder config",
		},
		{
			name:        "nil input channel",
			config:      validConfig,
			input:       nil,
			output:      bytes.NewBuffer(nil),
			expectError: "input channel cannot be nil",
		},
		{
			name:        "nil output writer",
			config:      validConfig,
			input:       make(chan []byte),
			output:      nil,
			expectError: "output writer cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			encoder, err := mp3.NewStreamingEncoder(tt.config, tt.input, tt.output)

			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
				assert.Nil(t, encoder)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, encoder)
			}
		})
	}
}

func TestStreamingEncoder_EncodesData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	input := make(chan []byte, 10)
	output := bytes.NewBuffer(nil)

	config := mp3.EncoderConfig{
		SampleRate:      16000,
		Channels:        1,
		BufferThreshold: 100, // Small threshold for testing
	}.WithDefaults()

	encoder, err := mp3.NewStreamingEncoder(config, input, output)
	require.NoError(t, err)

	err = encoder.Start(ctx)
	require.NoError(t, err)

	// Send some test data (100 bytes = 50 int16 samples)
	testData := make([]byte, 100)
	for i := range testData {
		testData[i] = byte(i)
	}

	input <- testData
	close(input) // Signal completion

	err = encoder.Wait()
	require.NoError(t, err)

	// Verify MP3 data was written
	assert.Greater(t, output.Len(), 0, "expected MP3 data to be written")
}

func TestStreamingEncoder_HandlesContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	input := make(chan []byte, 10)
	output := bytes.NewBuffer(nil)

	config := mp3.EncoderConfig{
		SampleRate:      16000,
		Channels:        1,
		BufferThreshold: 4096,
	}.WithDefaults()

	encoder, err := mp3.NewStreamingEncoder(config, input, output)
	require.NoError(t, err)

	err = encoder.Start(ctx)
	require.NoError(t, err)

	// Cancel context immediately
	cancel()

	err = encoder.Wait()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}

func TestStreamingEncoder_HandlesChannelClose(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	input := make(chan []byte, 10)
	output := bytes.NewBuffer(nil)

	config := mp3.EncoderConfig{
		SampleRate:      16000,
		Channels:        1,
		BufferThreshold: 4096,
	}.WithDefaults()

	encoder, err := mp3.NewStreamingEncoder(config, input, output)
	require.NoError(t, err)

	err = encoder.Start(ctx)
	require.NoError(t, err)

	// Close input channel immediately
	close(input)

	err = encoder.Wait()
	require.NoError(t, err, "encoder should handle channel close gracefully")
}

func TestStreamingEncoder_CannotStartTwice(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	input := make(chan []byte, 10)
	output := bytes.NewBuffer(nil)

	config := mp3.EncoderConfig{
		SampleRate:      16000,
		Channels:        1,
		BufferThreshold: 4096,
	}.WithDefaults()

	encoder, err := mp3.NewStreamingEncoder(config, input, output)
	require.NoError(t, err)

	// First start should succeed
	err = encoder.Start(ctx)
	require.NoError(t, err)

	// Second start should fail
	err = encoder.Start(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "encoder already started")

	// Clean up
	close(input)
	_ = encoder.Wait()
}

func TestStreamingEncoder_MultipleDataChunks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	input := make(chan []byte, 100)
	output := bytes.NewBuffer(nil)

	config := mp3.EncoderConfig{
		SampleRate:      16000,
		Channels:        1,
		BufferThreshold: 200, // Small threshold to trigger multiple encodes
	}.WithDefaults()

	encoder, err := mp3.NewStreamingEncoder(config, input, output)
	require.NoError(t, err)

	err = encoder.Start(ctx)
	require.NoError(t, err)

	// Send multiple chunks that will trigger multiple encode batches
	for i := 0; i < 10; i++ {
		testData := make([]byte, 100)
		for j := range testData {
			testData[j] = byte(j + i)
		}
		input <- testData

		// Small delay to ensure encoder processes
		time.Sleep(10 * time.Millisecond)
	}

	close(input)

	err = encoder.Wait()
	require.NoError(t, err)

	// Verify MP3 data was written
	assert.Greater(t, output.Len(), 0, "expected MP3 data to be written")
}
