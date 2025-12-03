package channels

import (
	"errors"
)

var (
	ErrChannelClosed  = errors.New("channel closed")
	ErrChannelTimeout = errors.New("send timeout")
	ErrChannelFull    = errors.New("channel full")
)
