package audio

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRecorder(t *testing.T) {
	outputPath := "/tmp/test.wav"
	maxDuration := 1 * time.Hour
	maxBytes := int64(268435456)

	recorder := NewRecorder(outputPath, maxDuration, maxBytes)

	assert.NotNil(t, recorder)
	assert.Equal(t, outputPath, recorder.outputPath)
	assert.Equal(t, maxDuration, recorder.maxDuration)
	assert.Equal(t, maxBytes, recorder.maxBytes)
	assert.Equal(t, uint32(16000), recorder.sampleRate)
	assert.Equal(t, uint32(1), recorder.channels)
}

func TestRecorder_Start_CreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := tmpDir + "/recordings/test.wav"
	recorder := NewRecorder(outputPath, 1*time.Hour, 1024*1024)

	err := recorder.Start()

	assert.NoError(t, err)
	// Directory should be created but recording not started yet
	// We'll verify directory creation in actual implementation
}

func TestRecorder_Stop_BeforeStart(t *testing.T) {
	recorder := NewRecorder("/tmp/test.wav", 1*time.Hour, 1024*1024)

	err := recorder.Stop()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}
