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

func TestNewRecorder_ValidatesConfig(t *testing.T) { //nolint:funlen // Table-driven test with many cases
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

func TestFormatWithBold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		text       string
		shouldBold bool
		want       string
	}{
		{
			name:       "not bold",
			text:       "hello",
			shouldBold: false,
			want:       "hello",
		},
		{
			name:       "bold",
			text:       "hello",
			shouldBold: true,
			want:       "\033[1mhello\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatWithBold(tt.text, tt.shouldBold)
			if got != tt.want {
				t.Errorf("formatWithBold() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		elapsed    time.Duration
		max        time.Duration
		shouldBold bool
		wantPlain  string // expected without ANSI codes
	}{
		{
			name:       "50 percent not bold",
			elapsed:    30 * time.Minute,
			max:        60 * time.Minute,
			shouldBold: false,
			wantPlain:  "00:30:00 / 01:00:00 (50%)",
		},
		{
			name:       "90 percent bold",
			elapsed:    54 * time.Minute,
			max:        60 * time.Minute,
			shouldBold: true,
			wantPlain:  "00:54:00 / 01:00:00 (90%)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatDuration(tt.elapsed, tt.max, tt.shouldBold)

			if tt.shouldBold {
				// Should have ANSI codes
				expected := "\033[1m" + tt.wantPlain + "\033[0m"
				if got != expected {
					t.Errorf("formatDuration() = %q, want %q", got, expected)
				}
			} else if got != tt.wantPlain {
				// Should not have ANSI codes
				t.Errorf("formatDuration() = %q, want %q", got, tt.wantPlain)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		current    int64
		max        int64
		shouldBold bool
		wantPlain  string
	}{
		{
			name:       "50 percent not bold",
			current:    128 * 1024 * 1024, // 128 MB
			max:        256 * 1024 * 1024, // 256 MB
			shouldBold: false,
			wantPlain:  "128.0 MB / 256.0 MB (50%)",
		},
		{
			name:       "90 percent bold",
			current:    230 * 1024 * 1024, // ~230 MB
			max:        256 * 1024 * 1024, // 256 MB
			shouldBold: true,
			wantPlain:  "230.0 MB / 256.0 MB (89%)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatBytes(tt.current, tt.max, tt.shouldBold)

			if tt.shouldBold {
				// Should have ANSI codes
				expected := "\033[1m" + tt.wantPlain + "\033[0m"
				if got != expected {
					t.Errorf("formatBytes() = %q, want %q", got, expected)
				}
			} else if got != tt.wantPlain {
				// Should not have ANSI codes
				t.Errorf("formatBytes() = %q, want %q", got, tt.wantPlain)
			}
		})
	}
}
