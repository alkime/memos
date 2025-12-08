package audio

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/alkime/memos/internal/cli/audio/device"
	"github.com/alkime/memos/internal/mp3"
	"github.com/gen2brain/malgo"
)

const (
	defaultSampleRate = 16_000 // Whisper native sample rate is 16kHz
	defaultChannels   = 1      // Whisper native audio is mono
)

// Sentinel errors for limit detection.
var (
	ErrMaxDurationReached = errors.New("max duration reached")
	ErrMaxBytesReached    = errors.New("max bytes reached")
)

type FileRecorderConfig struct {
	OutputPath        string
	MaxDuration       time.Duration
	MaxBytes          int64
	IgnoreStopSignals bool
	DisplayProgress   bool
}

// FileRecorder handles audio recording from microphone
// and streaming it into an MP3 file.
type FileRecorder struct {
	config FileRecorderConfig
}

// NewRecorder creates a new audio recorder.
func NewRecorder(conf FileRecorderConfig) (*FileRecorder, error) {
	if conf.MaxDuration <= 0 {
		return nil, errors.New("MaxDuration must be positive")
	}
	if conf.MaxBytes <= 0 {
		return nil, errors.New("MaxBytes must be positive")
	}

	return &FileRecorder{
		config: conf,
	}, nil
}

//nolint:funlen // Complex goroutine coordination
func (r *FileRecorder) Go(ctx context.Context, callback bufferdPacketCallback) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	dev := device.NewAudioDevice(&device.AudioDeviceConfig{
		Format:          malgo.FormatS16,
		SampleRate:      defaultSampleRate,
		CaptureChannels: defaultChannels,
	})

	dataC, err := dev.Capture(ctx)
	if err != nil {
		return fmt.Errorf("failed to start audio capture: %w", err)
	}
	// start the device. this will not block as the underlying device
	// handles os-level threading.
	if err = dev.Start(ctx); err != nil {
		return fmt.Errorf("failed to start audio device: %w", err)
	}
	// always hard stop when we return in this func.
	defer hardStop(ctx, dev)

	// Track start time for duration limit
	startTime := time.Now()

	// Track bytes written atomically for thread-safe access
	var bytesWritten atomic.Int64

	// Track which limit was hit (if any)
	var limitReached atomic.Value

	// Create MP3 file and encoder
	mp3File, err := os.Create(r.config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create MP3 file %s: %w", r.config.OutputPath, err)
	}
	defer closeFd(mp3File)

	encoderInput := make(chan []byte, 64)

	encoderConfig := mp3.EncoderConfig{
		SampleRate:      defaultSampleRate,
		Channels:        defaultChannels,
		BufferThreshold: 4096,
	}.WithDefaults()

	encoder, err := mp3.NewStreamingEncoder(encoderConfig, encoderInput, mp3File)
	if err != nil {
		return fmt.Errorf("failed to create MP3 encoder: %w", err)
	}

	if err := encoder.Start(ctx); err != nil {
		return fmt.Errorf("failed to start encoder: %w", err)
	}

	// spawn some worker goroutines, some of which are optional:
	// -- read the data channel with limit checking
	// -- (optionally) display progress periodically
	// -- listen for "finished" signals from ^C, Enter, or context (optionally skip this if other method of halting)
	wg := new(sync.WaitGroup)

	// Packet reading goroutine with inline limit checks
	wg.Go(func() {
		defer close(encoderInput) // Signal encoder to finish

	loop:
		for {
			var packet device.DataPacket
			select {
			case p := <-dataC:
				packet = p
			case <-ctx.Done():
				break loop
			}

			// Send to encoder
			encoderInput <- packet

			if callback != nil {
				// Invoke callback with the captured packet
				callback(packet)
			}

			// Update atomic counter
			bytesWritten.Add(int64(len(packet)))

			// Inline limit checks
			if bytesWritten.Load() >= r.config.MaxBytes {
				slog.Info("recording stopped", "reason", "max_bytes_reached",
					"bytes", bytesWritten.Load())
				limitReached.Store(ErrMaxBytesReached)
				break
			}

			elapsed := time.Since(startTime)
			if elapsed >= r.config.MaxDuration {
				slog.Info("recording stopped", "reason", "max_duration_reached",
					"duration", elapsed)
				limitReached.Store(ErrMaxDurationReached)
				break
			}
		}
		cancel() // Ensure context is cancelled
		slog.Info("data reading stopped")
	})

	// Progress display goroutine
	if r.config.DisplayProgress {
		wg.Go(func() {
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					elapsed := time.Since(startTime)
					bytes := bytesWritten.Load()

					timePercent := int(float64(elapsed) / float64(r.config.MaxDuration) * 100)
					bytesPercent := int(float64(bytes) / float64(r.config.MaxBytes) * 100)

					// Show bold if either >= 90%
					timeWarning := timePercent >= 90
					bytesWarning := bytesPercent >= 90

					fmt.Printf("\rRecording: %s | %s\n", //nolint:forbidigo // CLI progress
						formatDuration(elapsed, r.config.MaxDuration, timeWarning),
						formatBytes(bytes, r.config.MaxBytes, bytesWarning))
				case <-ctx.Done():
					return
				}
			}
		})
	}

	// Catch signals if we're not ignoring
	if !r.config.IgnoreStopSignals {
		wg.Go(func() {
			<-catchStopSignals(ctx)
			slog.Info("received stop signal")
			cancel()
		})
	}

	slog.Info("running... waiting for recording to finish")
	wg.Wait()
	slog.Info("recording finished")

	// stop the device which should block until all data
	// has been written to the channel.
	if err = dev.Stop(ctx); err != nil {
		return fmt.Errorf("unable to stop audio device, unable to flush: %w", err)
	}

	defer dev.Dealloc(ctx)

	// Wait for encoder to finish processing all data
	if err := encoder.Wait(); err != nil {
		return fmt.Errorf("failed to encode MP3: %w", err)
	}

	slog.Info("MP3 encoding completed", "path", r.config.OutputPath)

	// Return sentinel error if limit was reached
	if err := limitReached.Load(); err != nil {
		return err.(error)
	}

	return nil
}

func hardStop(ctx context.Context, dev device.AudioDevice) {
	slog.Info("hard stopping audio device")
	err := dev.Stop(ctx)
	if err != nil {
		slog.Warn("failed to hard stop audio device", "error", err)
	}

	dev.Dealloc(ctx)
}

func closeFd(fd *os.File) {
	err := fd.Close()
	if err != nil {
		slog.Warn("failed to close file descriptor", "error", err)
	}
}

// formatWithBold wraps text in ANSI bold codes if shouldBold is true.
func formatWithBold(text string, shouldBold bool) string {
	if shouldBold {
		return fmt.Sprintf("\033[1m%s\033[0m", text)
	}

	return text
}

// formatDuration formats elapsed and maxDuration duration with optional bold.
func formatDuration(elapsed, maxDuration time.Duration, shouldBold bool) string {
	// Format as HH:MM:SS
	formatTime := func(d time.Duration) string {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}

	elapsedStr := formatTime(elapsed)
	maxStr := formatTime(maxDuration)
	percent := int(float64(elapsed) / float64(maxDuration) * 100)

	text := fmt.Sprintf("%s / %s (%d%%)", elapsedStr, maxStr, percent)

	return formatWithBold(text, shouldBold)
}

// formatBytes formats bytes in MB with optional bold.
func formatBytes(current, maxBytes int64, shouldBold bool) string {
	currentMB := float64(current) / (1024 * 1024)
	maxMB := float64(maxBytes) / (1024 * 1024)
	percent := int(float64(current) / float64(maxBytes) * 100)

	text := fmt.Sprintf("%.1f MB / %.1f MB (%d%%)", currentMB, maxMB, percent)

	return formatWithBold(text, shouldBold)
}

// catchStopSignals listens for a few things:
//   - OS signals: SIGINT, SIGTERM
//   - Context's Done channel
//   - User inputting newline or space into stdin.
func catchStopSignals(ctx context.Context) <-chan struct{} {
	stopC := make(chan struct{})
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)

	// Channel to signal when user presses Enter or Space
	stdinC := make(chan struct{})

	// Goroutine to watch for stdin input
	// Note: This goroutine may leak if stdin input never arrives and another
	// stop signal is received first. This is acceptable because:
	// 1. os.Stdin.Read() cannot be cancelled by context
	// 2. The goroutine will be cleaned up when the program exits
	// 3. This is a short-lived CLI tool, not a long-running server
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				return
			}
			// Check for newline (Enter) or space
			if buf[0] == '\n' || buf[0] == ' ' {
				close(stdinC)
				return
			}
		}
	}()

	go func() {
		defer close(stopC)
		defer signal.Stop(sigC)

		select {
		case <-ctx.Done():
		case <-sigC:
		case <-stdinC:
		}
	}()

	return stopC
}

type bufferdPacketCallback func(packet device.DataPacket)
