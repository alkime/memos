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
	"github.com/gen2brain/malgo"

	"github.com/youpy/go-wav"
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
// and streaming it into a WAV file.
type FileRecorder struct {
	config FileRecorderConfig
}

// recordingState tracks the state of an active recording.
type recordingState struct {
	isRecording bool
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
			fmt.Print(".")
			// todo: check size limits.
			_, err := buf.Write(packet)
			if err != nil {
				slog.Error("failed to write audio packet to buffer. halting....", "error", err)
				break
			}
		}
		fmt.Println("\n[finished audio read loop]")
	})
	wg.Go(func() {
		select {
		case <-catchSignals():
			slog.Info("received stop signal, stopping recording")
			hardStop(ctx, dev)
		case <-ctx.Done():
			slog.Info("context done, stopping recording")
			hardStop(ctx, dev)
		}
	})

	slog.Info("running... waiting for recording to finish")
	wg.Wait()
	slog.Info("recording finished. buffer size bytes", "size_bytes", buf.Len())

	// flush buffer to WAV file.
	err = r.flushWAVFile(r.config.OutputPath, buf)
	if err != nil {
		return fmt.Errorf("failed to flush WAV file: %w", err)
	}

	return nil
}

func (r *FileRecorder) flushWAVFile(wavFilePath string, buf *bytes.Buffer) error {
	fd, err := os.Create(wavFilePath)
	if err != nil {
		return fmt.Errorf("failed to create WAV file %s: %w", wavFilePath, err)
	}
	defer closeFd(fd)

	// Convert buffer bytes to []wav.Sample
	rawBytes := buf.Bytes()
	numSamples := len(rawBytes) / 2 // 2 bytes per int16 sample
	samples := make([]wav.Sample, numSamples)

	for i := range numSamples {
		// Read 2 bytes and convert to int16 (little-endian)
		sampleInt16 := int16(binary.LittleEndian.Uint16(rawBytes[i*2 : i*2+2]))

		// Convert to int and store in both channels (mono audio)
		sampleValue := int(sampleInt16)
		samples[i] = wav.Sample{
			Values: [2]int{sampleValue, sampleValue},
		}
	}

	wavWriter := wav.NewWriter(fd, uint32(len(samples)), uint16(defaultChannels), uint32(defaultSampleRate), 16)
	err = wavWriter.WriteSamples(samples)
	if err != nil {
		return fmt.Errorf("failed to write samples to WAV file %s: %w", wavFilePath, err)
	}

	slog.Info("WAV file saved", "path", wavFilePath)
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

func catchSignals() <-chan os.Signal {
	stopC := make(chan os.Signal, 2)
	signal.Notify(stopC, os.Interrupt, syscall.SIGTERM)
	return stopC
}
