package audio //nolint:testpackage // Testing package-private function

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

// Test that catchStopSignals returns a channel that closes when context is cancelled.
func TestCatchStopSignals_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	stopC := catchStopSignals(ctx)

	// Cancel the context
	cancel()

	// Channel should close/receive signal
	select {
	case <-stopC:
		// Expected: channel should receive signal
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected stop signal when context cancelled, but timed out")
	}
}

// Test that catchStopSignals returns a channel that closes when OS signal is received.
func TestCatchStopSignals_OSSignal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	stopC := catchStopSignals(ctx)

	// Send SIGINT to current process
	err := syscall.Kill(os.Getpid(), syscall.SIGINT)
	if err != nil {
		t.Fatalf("failed to send signal: %v", err)
	}

	// Channel should close/receive signal
	select {
	case <-stopC:
		// Expected: channel should receive signal
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected stop signal when OS signal received, but timed out")
	}
}
