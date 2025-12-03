package channels

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// FanOut broadcasts messages from a single input channel to multiple subscriber channels.
// It owns the input channel and handles graceful shutdown via context cancellation.
//
// Messages are sent to subscribers in a non-blocking manner - if a subscriber's channel
// is full, the message is dropped for that subscriber.
//
// On context cancellation, the input channel is closed and all remaining messages
// are drained to subscribers before shutdown completes.
type FanOut[T any] struct {
	subscribers []chan<- T
	input       chan T
	started     atomic.Bool
	wg          sync.WaitGroup
}

// Subscribe adds a channel to receive broadcasted messages.
// Must be called before Run(). Not safe for concurrent use with Run().
func (f *FanOut[T]) Subscribe(ch chan<- T) {
	f.subscribers = append(f.subscribers, ch)
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

	// Start a goroutine for each subscriber
	for _, sub := range f.subscribers {
		f.wg.Add(1)
		sub := sub // Create per-iteration copy for Go < 1.22 compatibility
		go func() {
			defer f.wg.Done()
			// Drain channel until closed
			for msg := range f.input {
				select {
				case sub <- msg:
				default:
					// Drop message if subscriber channel is full
				}
			}
		}()
	}

	f.started.Store(true)

	// Shutdown handler: close input and wait for drain to complete
	go func() {
		<-ctx.Done()
		close(f.input)
		f.wg.Wait()
	}()

	return f.input, nil
}

// SendNonBlock attempts to send a message without blocking.
// Returns error if the channel is full or closed.
func SendNonBlock[T any](ch chan<- T, msg T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("channel closed")
		}
	}()

	select {
	case ch <- msg:
		return nil
	default:
		return errors.New("channel full")
	}
}

// SendWithTimeout sends a message with a timeout.
// Returns error if the timeout expires or channel is closed.
func SendWithTimeout[T any](ch chan<- T, msg T, timeout time.Duration) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("channel closed")
		}
	}()

	select {
	case ch <- msg:
		return nil
	case <-time.After(timeout):
		return errors.New("send timeout")
	}
}
