package channels

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// subscriber holds a channel and its send timeout configuration.
type subscriber[T any] struct {
	ch      chan<- T
	timeout *time.Duration // nil means non-blocking
}

// FanOut broadcasts messages from a single input channel to multiple subscriber channels.
// It owns the input channel and handles graceful shutdown via context cancellation.
//
// Messages are sent to subscribers using the configured send strategy:
// - Non-blocking (default): Messages are dropped if channel is full
// - With timeout: Messages are dropped if send times out
//
// On context cancellation, the input channel is closed and all remaining messages
// are drained to subscribers before shutdown completes.
type FanOut[T any] struct {
	subscribers []subscriber[T]
	input       chan T
	started     atomic.Bool
	wg          sync.WaitGroup
}

// NewFanOut creates a new FanOut instance with subscribers for the give type T.
func NewFanOut[T any]() *FanOut[T] {
	return &FanOut[T]{}
}

// Subscribe adds a channel to receive broadcasted messages in non-blocking mode.
// If the channel is full, messages will be dropped for that subscriber.
// Must be called before Run(). Not safe for concurrent use with Run().
func (f *FanOut[T]) Subscribe(ch chan<- T) {
	f.subscribers = append(f.subscribers, subscriber[T]{
		ch:      ch,
		timeout: nil,
	})
}

// SubscribeWithTimeout adds a channel to receive broadcasted messages with a send timeout.
// If the send times out, messages will be dropped for that subscriber.
// Must be called before Run(). Not safe for concurrent use with Run().
func (f *FanOut[T]) SubscribeWithTimeout(ch chan<- T, timeout time.Duration) {
	f.subscribers = append(f.subscribers, subscriber[T]{
		ch:      ch,
		timeout: &timeout,
	})
}

// Run starts the fan-out and returns the input channel for sending messages.
//
// The returned channel is owned by FanOut and will be closed on context cancellation.
// After closure, all remaining messages are drained to subscribers.
//
// Returns error if already started or no subscribers exist.
func (f *FanOut[T]) Run(ctx context.Context) (chan<- T, error) {
	if f.started.Load() {
		return nil, fmt.Errorf("fan out already started")
	}

	if len(f.subscribers) == 0 {
		return nil, fmt.Errorf("no subscribers available")
	}

	f.input = make(chan T, len(f.subscribers)*2)

	// Start broadcaster goroutine
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()

		// Read each message from input
		for msg := range f.input {
			// Broadcast to all subscribers
			for _, sub := range f.subscribers {
				if sub.timeout != nil {
					// Send with timeout - errors (timeout/closed) mean message is dropped
					_ = SendWithTimeout(sub.ch, msg, *sub.timeout)
				} else {
					// Send non-blocking - errors (full/closed) mean message is dropped
					_ = SendNonBlock(sub.ch, msg)
				}
			}
		}
	}()

	f.started.Store(true)

	// Shutdown handler: close input and wait for drain to complete
	go func() {
		<-ctx.Done()
		close(f.input)
		f.wg.Wait()
	}()

	return f.input, nil
}

// Wait blocks until all subscribers have finished processing messages.
// This is useful for waiting for graceful shutdown to complete after
// the context is cancelled. Multiple goroutines can safely call Wait().
func (f *FanOut[T]) Wait() {
	f.wg.Wait()
}
