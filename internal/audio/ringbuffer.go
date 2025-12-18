package audio

import (
	"encoding/binary"
	"sync"
)

// SampleRingBuffer is a thread-safe circular buffer for audio samples.
// It stores int16 samples and allows concurrent reads while writing.
type SampleRingBuffer struct {
	samples []int16
	head    int // Next write position
	count   int // Number of valid samples (up to capacity)
	mu      sync.RWMutex
}

// NewSampleRingBuffer creates a ring buffer with the given capacity.
func NewSampleRingBuffer(capacity int) *SampleRingBuffer {
	return &SampleRingBuffer{
		samples: make([]int16, capacity),
		head:    0,
		count:   0,
		mu:      sync.RWMutex{},
	}
}

// Write appends samples to the buffer, overwriting oldest if full.
// This method is safe to call from a single writer goroutine.
func (b *SampleRingBuffer) Write(samples []int16) {
	if len(samples) == 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	capacity := len(b.samples)

	for _, sample := range samples {
		b.samples[b.head] = sample
		b.head = (b.head + 1) % capacity

		if b.count < capacity {
			b.count++
		}
	}
}

// ReadSamples returns up to n most recent samples in chronological order.
// Returns fewer samples if the buffer contains less than n.
// This method is safe to call concurrently from multiple goroutines.
func (b *SampleRingBuffer) ReadSamples(n int) []int16 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.count == 0 || n <= 0 {
		return nil
	}

	// Limit n to available samples
	if n > b.count {
		n = b.count
	}

	result := make([]int16, n)
	capacity := len(b.samples)

	// Calculate start position (oldest sample to return)
	// head points to next write position, so oldest is at (head - count)
	// We want the last n samples, so start at (head - n)
	start := (b.head - n + capacity) % capacity

	for i := 0; i < n; i++ {
		result[i] = b.samples[(start+i)%capacity]
	}

	return result
}

// Count returns the number of valid samples in the buffer.
func (b *SampleRingBuffer) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.count
}

// BytesToInt16 converts S16LE (signed 16-bit little-endian) bytes to int16 samples.
func BytesToInt16(data []byte) []int16 {
	numSamples := len(data) / 2
	if numSamples == 0 {
		return nil
	}

	samples := make([]int16, numSamples)

	for i := 0; i < numSamples; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(data[i*2:]))
	}

	return samples
}
