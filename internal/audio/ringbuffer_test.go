package audio_test

import (
	"context"
	"testing"
	"time"

	"github.com/alkime/memos/internal/audio"
	"github.com/stretchr/testify/require"
)

func TestSampleRingBuffer_Write(t *testing.T) {
	t.Parallel()

	buf := audio.NewSampleRingBuffer(10)

	// Write 5 samples
	buf.Write([]int16{1, 2, 3, 4, 5})

	got := buf.ReadSamples(5)
	require.Equal(t, []int16{1, 2, 3, 4, 5}, got)
	require.Equal(t, 5, buf.Count())
}

func TestSampleRingBuffer_WriteEmpty(t *testing.T) {
	t.Parallel()

	buf := audio.NewSampleRingBuffer(10)
	buf.Write([]int16{})

	require.Equal(t, 0, buf.Count())
	require.Nil(t, buf.ReadSamples(5))
}

func TestSampleRingBuffer_Wraparound(t *testing.T) {
	t.Parallel()

	buf := audio.NewSampleRingBuffer(5)

	// Write 7 samples (wraps around, overwrites first 2)
	buf.Write([]int16{1, 2, 3, 4, 5, 6, 7})

	// Should return last 5: [3, 4, 5, 6, 7]
	got := buf.ReadSamples(5)
	require.Equal(t, []int16{3, 4, 5, 6, 7}, got)
	require.Equal(t, 5, buf.Count())
}

func TestSampleRingBuffer_MultipleWrites(t *testing.T) {
	t.Parallel()

	buf := audio.NewSampleRingBuffer(5)

	// Write in batches
	buf.Write([]int16{1, 2})
	buf.Write([]int16{3, 4})
	buf.Write([]int16{5, 6})

	// Should have last 5: [2, 3, 4, 5, 6]
	got := buf.ReadSamples(5)
	require.Equal(t, []int16{2, 3, 4, 5, 6}, got)
}

func TestSampleRingBuffer_ReadLessThanAvailable(t *testing.T) {
	t.Parallel()

	buf := audio.NewSampleRingBuffer(10)
	buf.Write([]int16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	// Read only last 3
	got := buf.ReadSamples(3)
	require.Equal(t, []int16{8, 9, 10}, got)
}

func TestSampleRingBuffer_ReadMoreThanAvailable(t *testing.T) {
	t.Parallel()

	buf := audio.NewSampleRingBuffer(10)
	buf.Write([]int16{1, 2, 3})

	// Request more than available
	got := buf.ReadSamples(10)
	require.Equal(t, []int16{1, 2, 3}, got)
}

func TestSampleRingBuffer_ReadZero(t *testing.T) {
	t.Parallel()

	buf := audio.NewSampleRingBuffer(10)
	buf.Write([]int16{1, 2, 3})

	got := buf.ReadSamples(0)
	require.Nil(t, got)
}

func TestSampleRingBuffer_ReadNegative(t *testing.T) {
	t.Parallel()

	buf := audio.NewSampleRingBuffer(10)
	buf.Write([]int16{1, 2, 3})

	got := buf.ReadSamples(-1)
	require.Nil(t, got)
}

func TestSampleRingBuffer_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	buf := audio.NewSampleRingBuffer(1000)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Writer goroutine
	go func() {
		counter := int16(0)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				buf.Write([]int16{counter, counter + 1, counter + 2})
				counter += 3
			}
		}
	}()

	// Reader goroutine - should not panic or race
	for {
		select {
		case <-ctx.Done():
			return
		default:
			samples := buf.ReadSamples(10)
			// Just verify we got something or nil
			_ = samples
		}
	}
}

func TestBytesToInt16(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []byte
		expected []int16
	}{
		{
			name:     "empty",
			input:    []byte{},
			expected: nil,
		},
		{
			name:     "single sample",
			input:    []byte{0x00, 0x01}, // 256 in little-endian
			expected: []int16{256},
		},
		{
			name:     "multiple samples",
			input:    []byte{0x01, 0x00, 0x02, 0x00, 0x03, 0x00}, // 1, 2, 3
			expected: []int16{1, 2, 3},
		},
		{
			name:     "negative sample",
			input:    []byte{0xFF, 0xFF}, // -1 in little-endian signed
			expected: []int16{-1},
		},
		{
			name:     "max positive",
			input:    []byte{0xFF, 0x7F}, // 32767
			expected: []int16{32767},
		},
		{
			name:     "max negative",
			input:    []byte{0x00, 0x80}, // -32768
			expected: []int16{-32768},
		},
		{
			name:     "odd byte count truncates",
			input:    []byte{0x01, 0x00, 0x02}, // Only first 2 bytes form a sample
			expected: []int16{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := audio.BytesToInt16(tt.input)
			require.Equal(t, tt.expected, got)
		})
	}
}
