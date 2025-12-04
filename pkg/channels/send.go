package channels

import "time"

// SendNonBlock attempts to send a message without blocking.
// Returns error if the channel is full or closed.
func SendNonBlock[T any](ch chan<- T, msg T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrChannelClosed
		}
	}()

	select {
	case ch <- msg:
		return nil
	default:
		return ErrChannelFull
	}
}

// SendWithTimeout sends a message with a timeout.
// Returns error if the timeout expires or channel is closed.
func SendWithTimeout[T any](ch chan<- T, msg T, timeout time.Duration) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrChannelClosed
		}
	}()

	select {
	case ch <- msg:
		return nil
	case <-time.After(timeout):
		return ErrChannelTimeout
	}
}
