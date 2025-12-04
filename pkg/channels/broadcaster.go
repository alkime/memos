package channels

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// subscriber holds a channel and its send timeout configuration.
type subscriber[T any] struct {
	ch       chan<- T
	timeout  *time.Duration // nil means non-blocking
	inactive atomic.Bool
	dropped  atomic.Int32
}

func (s *subscriber[T]) send(msg T) {
	if s.inactive.Load() {
		s.dropped.Add(1)
		return
	}
	var err error
	if s.timeout != nil {
		// Send with timeout
		err = SendWithTimeout(s.ch, msg, *s.timeout)
	} else {
		// Non-blocking send.
		err = SendNonBlock(s.ch, msg)
	}
	if err != nil {
		// if channel is closed, mark inactive
		// otherwise just count dropped messages
		s.dropped.Add(1)
		if errors.Is(err, ErrChannelClosed) {
			s.inactive.Store(true)
		}
	}
}

// Broadcaster broadcasts messages from a single input channel to multiple subscriber channels.
// It owns the input channel and handles graceful shutdown via context cancellation.
//
// Messages are sent to subscribers using the configured send strategy:
// - Non-blocking (default): Messages are dropped if channel is full
// - With timeout: Messages are dropped if send times out
//
// On context cancellation, the input channel is closed and all remaining messages
// are drained to subscribers before shutdown completes.
type Broadcaster[T any] struct {
	subscribers []subscriber[T]
	input       chan T
	started     atomic.Bool
	wg          sync.WaitGroup
}

// NewBroadcaster creates a new Broadcaster instance with subscribers for the given type T.
func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{}
}

// Subscribe adds a channel to receive broadcasted messages in non-blocking mode.
// If the channel is full, messages will be dropped for that subscriber.
// Must be called before Run(). Not safe for concurrent use with Run().
func (f *Broadcaster[T]) Subscribe(ch chan<- T) {
	f.subscribers = append(f.subscribers, subscriber[T]{
		ch:      ch,
		timeout: nil,
	})
}

// SubscribeWithTimeout adds a channel to receive broadcasted messages with a send timeout.
// If the send times out, messages will be dropped for that subscriber.
// Must be called before Run(). Not safe for concurrent use with Run().
func (f *Broadcaster[T]) SubscribeWithTimeout(ch chan<- T, timeout time.Duration) {
	f.subscribers = append(f.subscribers, subscriber[T]{
		ch:      ch,
		timeout: &timeout,
	})
}

// Run starts the broadcaster and returns the input channel for sending messages.
//
// The returned channel is owned by Broadcaster and will be closed on context cancellation.
// After closure, all remaining messages are drained to subscribers.
//
// Returns error if already started or no subscribers exist.
func (f *Broadcaster[T]) Run(ctx context.Context) (chan<- T, error) {
	if f.started.Load() {
		return nil, fmt.Errorf("broadcaster already started")
	}

	if len(f.subscribers) == 0 {
		return nil, fmt.Errorf("no subscribers available")
	}

	f.input = make(chan T, len(f.subscribers)*2)

	// Start broadcaster goroutine
	f.wg.Go(func() {
		// Read each message from input
		for msg := range f.input {
			// Broadcast to all subscribers
			for i := range f.subscribers {
				f.subscribers[i].send(msg)
			}
		}
	})

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
func (f *Broadcaster[T]) Wait() {
	f.wg.Wait()
}

type SubscriberStats struct {
	Dropped  int
	Inactive bool
}

func (f *Broadcaster[T]) Stats() []SubscriberStats {
	stats := make([]SubscriberStats, 0, len(f.subscribers))
	for i := range f.subscribers {
		stats = append(stats, SubscriberStats{
			Dropped:  int(f.subscribers[i].dropped.Load()),
			Inactive: f.subscribers[i].inactive.Load(),
		})
	}
	return stats
}
