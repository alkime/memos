package channels_test

import (
	"testing"
	"time"

	"github.com/alkime/memos/pkg/channels"
	"github.com/stretchr/testify/assert"
)

func TestSendFunctions(t *testing.T) {

	t.Run("send non-blocking", func(t *testing.T) {
		t.Run("success - buffered channel with capacity", func(t *testing.T) {
			ch := make(chan int, 2)
			err := channels.SendNonBlock(ch, 42)
			assert.NoError(t, err)
			assert.Equal(t, 42, <-ch) // Verify message was sent
		})

		t.Run("full - buffered channel", func(t *testing.T) {
			ch := make(chan int, 1)
			ch <- 1 // Fill buffer
			err := channels.SendNonBlock(ch, 42)
			assert.ErrorIs(t, err, channels.ErrChannelFull)
		})

		t.Run("full - unbuffered with no receiver", func(t *testing.T) {
			ch := make(chan int)
			err := channels.SendNonBlock(ch, 42)
			assert.ErrorIs(t, err, channels.ErrChannelFull)
		})

		t.Run("closed channel - empty", func(t *testing.T) {
			ch := make(chan int)
			close(ch)
			err := channels.SendNonBlock(ch, 42)
			assert.ErrorIs(t, err, channels.ErrChannelClosed)
		})

		t.Run("closed channel - with buffered data", func(t *testing.T) {
			ch := make(chan int, 2)
			ch <- 1 // Write data before closing
			close(ch)
			err := channels.SendNonBlock(ch, 42)
			assert.ErrorIs(t, err, channels.ErrChannelClosed)
			// Verify original data still readable
			assert.Equal(t, 1, <-ch)
		})
	})

	t.Run("send with timeout", func(t *testing.T) {
		t.Run("success - buffered channel with capacity", func(t *testing.T) {
			ch := make(chan int, 2)
			err := channels.SendWithTimeout(ch, 42, 10*time.Millisecond)
			assert.NoError(t, err)
			assert.Equal(t, 42, <-ch)
		})

		t.Run("success - unbuffered with receiver", func(t *testing.T) {
			ch := make(chan int)
			go func() { <-ch }()
			err := channels.SendWithTimeout(ch, 42, 10*time.Millisecond)
			assert.NoError(t, err)
		})

		t.Run("timeout - buffered channel full", func(t *testing.T) {
			ch := make(chan int, 1)
			ch <- 1 // Fill buffer
			err := channels.SendWithTimeout(ch, 42, 1*time.Millisecond)
			assert.ErrorIs(t, err, channels.ErrChannelTimeout)
		})

		t.Run("timeout - unbuffered with no receiver", func(t *testing.T) {
			ch := make(chan int)
			err := channels.SendWithTimeout(ch, 42, 1*time.Millisecond)
			assert.ErrorIs(t, err, channels.ErrChannelTimeout)
		})

		t.Run("closed channel - empty", func(t *testing.T) {
			ch := make(chan int)
			close(ch)
			err := channels.SendWithTimeout(ch, 42, 10*time.Millisecond)
			assert.ErrorIs(t, err, channels.ErrChannelClosed)
		})

		t.Run("closed channel - with buffered data", func(t *testing.T) {
			ch := make(chan int, 2)
			ch <- 1 // Write data before closing
			close(ch)
			err := channels.SendWithTimeout(ch, 42, 10*time.Millisecond)
			assert.ErrorIs(t, err, channels.ErrChannelClosed)
			// Verify original data still readable
			assert.Equal(t, 1, <-ch)
		})
	})
}
