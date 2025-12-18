package audio_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alkime/memos/internal/audio"
)

// TestRecorder_BytesWritten verifies that the recorder tracks bytes written to the PCM file.
func TestRecorder_BytesWritten(t *testing.T) {
	t.Parallel()

	// Setup: Create temp directory for test files
	tmpDir := t.TempDir()
	mp3Path := filepath.Join(tmpDir, "test.mp3")

	// Create input channel and recorder
	input := make(chan []byte)
	config := audio.Config{
		SampleRate: 16000,
		Channels:   1,
		MP3Path:    mp3Path,
	}

	recorder, err := audio.NewRecorder(config, input)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}

	// Start recording
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := recorder.Start(ctx); err != nil {
		t.Fatalf("failed to start recorder: %v", err)
	}

	// Initially, no bytes should be written
	if got := recorder.BytesWritten(); got != 0 {
		t.Errorf("BytesWritten() = %d, want 0 before any data", got)
	}

	// Send test data (100 bytes)
	testData1 := make([]byte, 100)
	input <- testData1

	// Give recorder time to process
	time.Sleep(50 * time.Millisecond)

	// Check bytes written
	if got := recorder.BytesWritten(); got != 100 {
		t.Errorf("BytesWritten() = %d, want 100 after first write", got)
	}

	// Send more data (200 bytes)
	testData2 := make([]byte, 200)
	input <- testData2

	// Give recorder time to process
	time.Sleep(50 * time.Millisecond)

	// Check total bytes written
	if got := recorder.BytesWritten(); got != 300 {
		t.Errorf("BytesWritten() = %d, want 300 after second write", got)
	}

	// Cleanup: Close channel and wait for recording to finish
	close(input)
	if err := recorder.Wait(); err != nil {
		t.Fatalf("recording failed: %v", err)
	}

	// Final check - bytes should still be 300
	if got := recorder.BytesWritten(); got != 300 {
		t.Errorf("BytesWritten() = %d, want 300 after recording complete", got)
	}

	// Verify MP3 file was created
	if _, err := os.Stat(mp3Path); os.IsNotExist(err) {
		t.Errorf("MP3 file was not created at %s", mp3Path)
	}
}
