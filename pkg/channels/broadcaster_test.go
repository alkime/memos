package channels_test

import (
	"context"
	"testing"
	"time"

	"github.com/alkime/memos/pkg/channels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcaster(t *testing.T) {
	t.Run("error cases", func(t *testing.T) {
		t.Run("subscribe with nil channel", func(t *testing.T) {
			fo := channels.NewBroadcaster[int]()
			err := fo.Subscribe(nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be nil")
		})

		t.Run("subscribe with timeout and nil channel", func(t *testing.T) {
			fo := channels.NewBroadcaster[int]()
			err := fo.SubscribeWithTimeout(nil, 1*time.Second)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be nil")
		})

		t.Run("subscribe with timeout and zero timeout", func(t *testing.T) {
			fo := channels.NewBroadcaster[int]()
			ch := make(chan int, 10)
			err := fo.SubscribeWithTimeout(ch, 0)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "must be positive")
		})

		t.Run("subscribe with timeout and negative timeout", func(t *testing.T) {
			fo := channels.NewBroadcaster[int]()
			ch := make(chan int, 10)
			err := fo.SubscribeWithTimeout(ch, -1*time.Second)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "must be positive")
		})

		t.Run("run with no subscribers", func(t *testing.T) {
			ctx := context.Background()
			fo := channels.NewBroadcaster[int]()
			_, err := fo.Run(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "no subscribers")
		})

		t.Run("run twice", func(t *testing.T) {
			ctx := context.Background()
			fo := channels.NewBroadcaster[int]()
			ch := make(chan int, 10)
			require.NoError(t, fo.Subscribe(ch))

			_, err := fo.Run(ctx)
			require.NoError(t, err)

			_, err = fo.Run(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "already started")
		})
	})

	t.Run("basic broadcasting", func(t *testing.T) {
		t.Run("single subscriber receives all messages", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fo := channels.NewBroadcaster[int]()
			sub := make(chan int, 10)
			require.NoError(t, fo.Subscribe(sub))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Send messages
			input <- 1
			input <- 2
			input <- 3

			// Shutdown and collect
			cancel()
			fo.Wait()
			close(sub)

			received := channels.ReceiveAll(sub, 10*time.Millisecond, 0)
			assert.Equal(t, []int{1, 2, 3}, received)
		})

		t.Run("multiple subscribers receive same messages", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fo := channels.NewBroadcaster[int]()
			sub1 := make(chan int, 10)
			sub2 := make(chan int, 10)
			sub3 := make(chan int, 10)
			require.NoError(t, fo.Subscribe(sub1))
			require.NoError(t, fo.Subscribe(sub2))
			require.NoError(t, fo.Subscribe(sub3))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Send messages
			input <- 1
			input <- 2
			input <- 3
			time.Sleep(5 * time.Millisecond) // Let messages propagate

			// Shutdown and collect
			cancel()
			fo.Wait()
			close(sub1)
			close(sub2)
			close(sub3)

			received1 := channels.ReceiveAll(sub1, 10*time.Millisecond, 0)
			received2 := channels.ReceiveAll(sub2, 10*time.Millisecond, 0)
			received3 := channels.ReceiveAll(sub3, 10*time.Millisecond, 0)

			// All subscribers should receive all messages (broadcast)
			assert.Equal(t, []int{1, 2, 3}, received1)
			assert.Equal(t, []int{1, 2, 3}, received2)
			assert.Equal(t, []int{1, 2, 3}, received3)
		})
	})

	t.Run("message dropping", func(t *testing.T) {
		t.Run("non-blocking subscriber drops when full", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fo := channels.NewBroadcaster[int]()
			sub := make(chan int, 1) // Small buffer
			require.NoError(t, fo.Subscribe(sub))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Send 2 messages quickly
			input <- 1
			input <- 2
			time.Sleep(5 * time.Millisecond) // Let sends complete

			// Shutdown and collect
			cancel()
			fo.Wait()
			close(sub)

			received := channels.ReceiveAll(sub, 10*time.Millisecond, 0)
			// Only first message should be received (second dropped due to full buffer)
			assert.Len(t, received, 1)
			assert.Equal(t, 1, received[0])
		})

		t.Run("timeout subscriber drops on timeout", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fo := channels.NewBroadcaster[int]()
			sub := make(chan int, 1) // Small buffer
			require.NoError(t, fo.SubscribeWithTimeout(sub, 1*time.Millisecond))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Send 2 messages quickly
			input <- 1
			input <- 2
			time.Sleep(5 * time.Millisecond) // Let sends complete

			// Shutdown and collect
			cancel()
			fo.Wait()
			close(sub)

			received := channels.ReceiveAll(sub, 10*time.Millisecond, 0)
			// Only first message should be received (second times out)
			assert.Len(t, received, 1)
			assert.Equal(t, 1, received[0])
		})

		t.Run("full subscriber drops while ready subscriber receives", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fo := channels.NewBroadcaster[int]()
			fullSub := make(chan int, 1)
			fullSub <- 99 // Pre-fill to make it full
			readySub := make(chan int, 10)

			require.NoError(t, fo.Subscribe(fullSub))
			require.NoError(t, fo.Subscribe(readySub))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Send multiple messages - broadcast to both subscribers
			for i := 1; i <= 5; i++ {
				input <- i
			}
			time.Sleep(10 * time.Millisecond) // Let sends complete

			// Shutdown and collect
			cancel()
			fo.Wait()

			// Drain fullSub
			<-fullSub // Remove pre-filled value
			close(fullSub)
			receivedFull := channels.ReceiveAll(fullSub, 10*time.Millisecond, 0)

			close(readySub)
			receivedReady := channels.ReceiveAll(readySub, 10*time.Millisecond, 0)

			// With broadcasting: full subscriber drops all, ready subscriber gets all
			assert.Empty(t, receivedFull, "full subscriber should drop all messages")
			assert.Equal(t, []int{1, 2, 3, 4, 5}, receivedReady, "ready subscriber should receive all messages")
		})
	})

	t.Run("stats", func(t *testing.T) {
		t.Run("reports zero stats for active subscribers", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fo := channels.NewBroadcaster[int]()
			sub1 := make(chan int, 10)
			sub2 := make(chan int, 10)
			require.NoError(t, fo.Subscribe(sub1))
			require.NoError(t, fo.Subscribe(sub2))

			_, err := fo.Run(ctx)
			require.NoError(t, err)

			stats := fo.Stats()
			require.Len(t, stats, 2)
			assert.Equal(t, 0, stats[0].Dropped)
			assert.False(t, stats[0].Inactive)
			assert.Equal(t, 0, stats[1].Dropped)
			assert.False(t, stats[1].Inactive)
		})

		t.Run("reports dropped message counts", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fo := channels.NewBroadcaster[int]()
			fullSub := make(chan int, 1)
			fullSub <- 99 // Pre-fill to make it full
			readySub := make(chan int, 10)

			require.NoError(t, fo.Subscribe(fullSub))
			require.NoError(t, fo.Subscribe(readySub))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Send multiple messages
			for i := 1; i <= 5; i++ {
				input <- i
			}
			time.Sleep(10 * time.Millisecond) // Let sends complete

			stats := fo.Stats()
			require.Len(t, stats, 2)

			// First subscriber (full) should have dropped messages
			assert.Equal(t, 5, stats[0].Dropped, "full subscriber should drop all 5 messages")
			assert.False(t, stats[0].Inactive, "full subscriber should still be active")

			// Second subscriber (ready) should have no dropped messages
			assert.Equal(t, 0, stats[1].Dropped, "ready subscriber should drop no messages")
			assert.False(t, stats[1].Inactive, "ready subscriber should still be active")
		})

		t.Run("reports inactive status when channel closed", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fo := channels.NewBroadcaster[int]()
			sub1 := make(chan int, 10)
			sub2 := make(chan int, 10)
			require.NoError(t, fo.Subscribe(sub1))
			require.NoError(t, fo.Subscribe(sub2))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Close first subscriber to trigger inactive state
			close(sub1)

			// Send messages - first subscriber will detect closed channel
			input <- 1
			input <- 2
			time.Sleep(10 * time.Millisecond) // Let sends complete

			stats := fo.Stats()
			require.Len(t, stats, 2)

			// First subscriber should be marked inactive with dropped messages
			assert.Equal(t, 2, stats[0].Dropped, "closed subscriber should count dropped messages")
			assert.True(t, stats[0].Inactive, "closed subscriber should be marked inactive")

			// Second subscriber should be active with no drops
			assert.Equal(t, 0, stats[1].Dropped)
			assert.False(t, stats[1].Inactive)
		})

		t.Run("accumulates dropped message counts", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fo := channels.NewBroadcaster[int]()
			sub := make(chan int, 1) // Small buffer
			require.NoError(t, fo.Subscribe(sub))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Send 2 messages, check stats
			input <- 1
			input <- 2
			time.Sleep(5 * time.Millisecond)

			stats1 := fo.Stats()
			require.Len(t, stats1, 1)
			assert.Equal(t, 1, stats1[0].Dropped, "should have 1 dropped message")

			// Send 2 more messages
			input <- 3
			input <- 4
			time.Sleep(5 * time.Millisecond)

			stats2 := fo.Stats()
			require.Len(t, stats2, 1)
			assert.Equal(t, 3, stats2[0].Dropped, "should accumulate to 3 dropped messages")
		})
	})

	t.Run("lifecycle", func(t *testing.T) {
		t.Run("context cancellation stops processing", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			fo := channels.NewBroadcaster[int]()
			sub := make(chan int, 10)
			require.NoError(t, fo.Subscribe(sub))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Send a message
			input <- 1

			// Cancel context and wait
			cancel()
			fo.Wait()

			// No more messages should be processed after cancellation
			close(sub)
			received := channels.ReceiveAll(sub, 10*time.Millisecond, 0)
			assert.Equal(t, []int{1}, received)
		})

		t.Run("messages in flight are drained", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			fo := channels.NewBroadcaster[int]()
			sub := make(chan int, 10)
			require.NoError(t, fo.Subscribe(sub))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Send messages
			input <- 1
			input <- 2
			input <- 3

			// Cancel immediately
			cancel()
			fo.Wait()
			close(sub)

			// All messages should have been drained to subscriber
			received := channels.ReceiveAll(sub, 10*time.Millisecond, 0)
			assert.Equal(t, []int{1, 2, 3}, received)
		})

		t.Run("wait blocks until complete", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			fo := channels.NewBroadcaster[int]()
			sub := make(chan int, 10)
			require.NoError(t, fo.Subscribe(sub))

			input, err := fo.Run(ctx)
			require.NoError(t, err)

			// Send message
			input <- 42

			// Cancel and measure wait time
			cancel()
			start := time.Now()
			fo.Wait()
			duration := time.Since(start)

			// Wait should return quickly since subscriber has buffer
			assert.Less(t, duration, 100*time.Millisecond)

			close(sub)
			received := channels.ReceiveAll(sub, 10*time.Millisecond, 0)
			assert.Equal(t, []int{42}, received)
		})
	})
}
