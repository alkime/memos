package channels_test

import (
	"testing"
	"time"

	"github.com/alkime/memos/pkg/channels"
	"github.com/stretchr/testify/assert"
)

func TestReceiveAll(t *testing.T) {
	t.Run("receives all messages until channel closes", func(t *testing.T) {
		ch := make(chan int, 5)
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)

		received := channels.ReceiveAll(ch, 100*time.Millisecond, 0)
		assert.Equal(t, []int{1, 2, 3}, received)
	})

	t.Run("stops at maxItems limit", func(t *testing.T) {
		ch := make(chan int, 5)
		ch <- 1
		ch <- 2
		ch <- 3
		ch <- 4
		ch <- 5

		received := channels.ReceiveAll(ch, 100*time.Millisecond, 3)
		assert.Equal(t, []int{1, 2, 3}, received)
		assert.Len(t, received, 3)

		// Channel should still have remaining messages
		assert.Equal(t, 4, <-ch)
		assert.Equal(t, 5, <-ch)
	})

	t.Run("maxItems=0 means unlimited", func(t *testing.T) {
		ch := make(chan int, 10)
		for i := 1; i <= 10; i++ {
			ch <- i
		}
		close(ch)

		received := channels.ReceiveAll(ch, 100*time.Millisecond, 0)
		assert.Len(t, received, 10)
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, received)
	})

	t.Run("stops on timeout", func(t *testing.T) {
		ch := make(chan int, 5)
		ch <- 1
		ch <- 2

		// Use short timeout, no more messages coming
		received := channels.ReceiveAll(ch, 10*time.Millisecond, 0)
		assert.Equal(t, []int{1, 2}, received)
	})

	t.Run("returns empty for closed channel", func(t *testing.T) {
		ch := make(chan int)
		close(ch)

		received := channels.ReceiveAll(ch, 100*time.Millisecond, 0)
		assert.Empty(t, received)
	})

	t.Run("returns empty for timeout on empty channel", func(t *testing.T) {
		ch := make(chan int, 5)

		received := channels.ReceiveAll(ch, 10*time.Millisecond, 0)
		assert.Empty(t, received)
	})

	t.Run("maxItems takes precedence over timeout", func(t *testing.T) {
		ch := make(chan int, 5)
		ch <- 1
		ch <- 2
		ch <- 3

		// Short timeout but maxItems=2 should return first
		start := time.Now()
		received := channels.ReceiveAll(ch, 100*time.Millisecond, 2)
		duration := time.Since(start)

		assert.Equal(t, []int{1, 2}, received)
		// Should return quickly due to maxItems, not wait for timeout
		assert.Less(t, duration, 50*time.Millisecond)
	})

	t.Run("receives messages as they arrive until timeout", func(t *testing.T) {
		ch := make(chan int, 5)

		// Send messages in background with delays
		go func() {
			for i := 1; i <= 3; i++ {
				ch <- i
				time.Sleep(5 * time.Millisecond)
			}
		}()

		// Should receive all 3 messages before timeout
		received := channels.ReceiveAll(ch, 50*time.Millisecond, 0)
		assert.Len(t, received, 3)
		assert.Equal(t, []int{1, 2, 3}, received)
	})
}
