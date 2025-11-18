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

func TestNewRecorder_ValidatesConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      FileRecorderConfig
		expectError string
	}{
		{
			name: "zero max duration",
			config: FileRecorderConfig{
				OutputPath:  "/tmp/test.mp3",
				MaxDuration: 0,
				MaxBytes:    1024,
			},
			expectError: "MaxDuration must be positive",
		},
		{
			name: "negative max duration",
			config: FileRecorderConfig{
				OutputPath:  "/tmp/test.mp3",
				MaxDuration: -1 * time.Second,
				MaxBytes:    1024,
			},
			expectError: "MaxDuration must be positive",
		},
		{
			name: "zero max bytes",
			config: FileRecorderConfig{
				OutputPath:  "/tmp/test.mp3",
				MaxDuration: 1 * time.Minute,
				MaxBytes:    0,
			},
			expectError: "MaxBytes must be positive",
		},
		{
			name: "negative max bytes",
			config: FileRecorderConfig{
				OutputPath:  "/tmp/test.mp3",
				MaxDuration: 1 * time.Minute,
				MaxBytes:    -1,
			},
			expectError: "MaxBytes must be positive",
		},
		{
			name: "valid config",
			config: FileRecorderConfig{
				OutputPath:  "/tmp/test.mp3",
				MaxDuration: 1 * time.Minute,
				MaxBytes:    1024,
			},
			expectError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder, err := NewRecorder(tt.config)

			if tt.expectError != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.expectError)
				}
				if err.Error() != tt.expectError {
					t.Fatalf("expected error %q, got %q", tt.expectError, err.Error())
				}
				if recorder != nil {
					t.Fatal("expected nil recorder when error occurs")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if recorder == nil {
					t.Fatal("expected non-nil recorder for valid config")
				}
			}
		})
	}
}
