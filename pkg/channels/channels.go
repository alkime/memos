package channels

import (
	"errors"
	"time"
)

var (
	ErrChannelClosed  = errors.New("channel closed")
	ErrChannelTimeout = errors.New("send timeout")
	ErrChannelFull    = errors.New("channel full")
)

// ReceiveAll drains a channel into a slice with timeout and optional item limit.
//
// Parameters:
//   - ch: The channel to receive from
//   - timeout: Maximum duration to wait for messages (required)
//   - maxItems: Maximum number of items to collect (0 = unlimited)
//
// Returns when:
//   - maxItems is reached (if maxItems > 0)
//   - Channel is closed
//   - Timeout expires
func ReceiveAll[T any](ch <-chan T, timeout time.Duration, maxItems int) []T {
	var results []T
	deadline := time.After(timeout)
	for {
		// Check if we've reached the item limit
		if maxItems > 0 && len(results) >= maxItems {
			return results
		}

		select {
		case msg, ok := <-ch:
			if !ok {
				return results
			}
			results = append(results, msg)
		case <-deadline:
			return results
		}
	}
}
