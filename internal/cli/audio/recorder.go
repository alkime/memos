package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alkime/memos/internal/cli/audio/device"
	mp3 "github.com/braheezy/shine-mp3/pkg/mp3"
	"github.com/gen2brain/malgo"
)

const (
	defaultSampleRate = 16_000 // Whisper native sample rate is 16kHz
	defaultChannels   = 1      // Whisper native audio is mono
)

type FileRecorderConfig struct {
	OutputPath  string
	MaxDuration time.Duration
	MaxBytes    int64
}

// FileRecorder handles audio recording from microphone
// and streaming it into an MP3 file.
type FileRecorder struct {
	config FileRecorderConfig
}

// NewRecorder creates a new audio recorder.
func NewRecorder(conf FileRecorderConfig) *FileRecorder {
	return &FileRecorder{
		config: conf,
	}
}

func (r *FileRecorder) Go(ctx context.Context) (err error) {
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

	// spawn 2 goroutines to:
	// -- read the data channel.
	// -- a task for listening for "finished" signals either from
	//    ^C or reading \n from stdin.
	wg := new(sync.WaitGroup)

	buf := bytes.NewBuffer(nil)

	wg.Go(func() {
		for packet := range dataC {
			// todo write a packet to WAV file.
			fmt.Print(".") //nolint:forbidigo // CLI progress indicator
			// todo: check size limits.
			_, err := buf.Write(packet)
			if err != nil {
				slog.Error("failed to write audio packet to buffer. halting....", "error", err)
				break
			}
		}
		fmt.Println("\n[finished audio read loop]") //nolint:forbidigo // CLI status message
	})
	wg.Go(func() {
		<-catchStopSignals(ctx)
		slog.Info("received stop signal, stopping recording")
		hardStop(ctx, dev)
	})

	slog.Info("running... waiting for recording to finish")
	wg.Wait()
	slog.Info("recording finished. buffer size bytes", "size_bytes", buf.Len())

	// flush buffer to MP3 file.
	err = r.flushMP3File(r.config.OutputPath, buf)
	if err != nil {
		return fmt.Errorf("failed to flush MP3 file: %w", err)
	}

	return nil
}

func (r *FileRecorder) flushMP3File(mp3FilePath string, buf *bytes.Buffer) error {
	fd, err := os.Create(mp3FilePath)
	if err != nil {
		return fmt.Errorf("failed to create MP3 file %s: %w", mp3FilePath, err)
	}
	defer closeFd(fd)

	// Convert buffer bytes to []int16
	// The buffer contains S16LE (16-bit signed little-endian) PCM data
	numSamples := buf.Len() / 2 // 2 bytes per int16 sample
	monoSamples := make([]int16, numSamples)

	// Read raw bytes directly into int16 slice using binary.Read
	reader := bytes.NewReader(buf.Bytes())
	err = binary.Read(reader, binary.LittleEndian, monoSamples)
	if err != nil {
		return fmt.Errorf("failed to read PCM samples: %w", err)
	}

	// WORKAROUND: shine-mp3 Write() has a bug for mono (always increments by samples_per_pass * 2)
	// Convert mono to stereo by duplicating samples (L=R)
	stereoSamples := make([]int16, numSamples*2)
	for i, sample := range monoSamples {
		stereoSamples[i*2] = sample   // Left channel
		stereoSamples[i*2+1] = sample // Right channel (duplicate)
	}

	slog.Info("encoding MP3", "monoSamples", numSamples, "stereoSamples", len(stereoSamples))

	// Create MP3 encoder as STEREO (workaround for mono bug)
	encoder := mp3.NewEncoder(defaultSampleRate, 2) // 2 channels

	// Write stereo PCM samples to MP3 file
	err = encoder.Write(fd, stereoSamples)
	if err != nil {
		return fmt.Errorf("failed to encode audio to MP3 %s: %w", mp3FilePath, err)
	}

	slog.Info("MP3 file saved", "path", mp3FilePath)

	return nil
}

func hardStop(ctx context.Context, dev device.AudioDevice) {
	slog.Info("hard stopping audio device")
	err := dev.Stop(ctx, true)
	if err != nil {
		slog.Warn("failed to hard stop audio device", "error", err)
	}
}

func closeFd(fd *os.File) {
	err := fd.Close()
	if err != nil {
		slog.Warn("failed to close file descriptor", "error", err)
	}
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
