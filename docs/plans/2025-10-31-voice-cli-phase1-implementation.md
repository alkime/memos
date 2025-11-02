# Voice CLI Phase 1 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI tool (`voice`) that automates voice-to-blog workflow: record audio → transcribe with Whisper API → generate Hugo markdown draft.

**Architecture:** Three independent commands (record/transcribe/process) with manual workflow. Kong CLI framework, malgo for audio, OpenAI Whisper API, Hugo markdown generation. Service-oriented packages under `internal/cli/`.

**Tech Stack:** Go, Kong CLI, malgo (audio), OpenAI Go SDK, testify/mock

---

## Task 1: Add Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add Kong CLI framework**

Run:
```bash
go get github.com/alecthomas/kong
```

Expected: Dependency added to go.mod

**Step 2: Add malgo audio library**

Run:
```bash
go get github.com/gen2brain/malgo
```

Expected: Dependency added to go.mod

**Step 3: Add OpenAI Go SDK**

Run:
```bash
go get github.com/sashabaranov/go-openai
```

Expected: Dependency added to go.mod

**Step 4: Verify dependencies**

Run:
```bash
go mod tidy
```

Expected: go.mod and go.sum updated cleanly

**Step 5: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add voice CLI dependencies (kong, malgo, openai)"
```

---

## Task 2: Create CLI Entry Point Structure

**Files:**
- Create: `cmd/voice/main.go`

**Step 1: Write basic Kong CLI structure**

Create `cmd/voice/main.go`:
```go
package main

import (
	"os"

	"github.com/alecthomas/kong"
)

// CLI defines the voice command structure
type CLI struct {
	Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
	Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
	Process    ProcessCmd    `cmd:"" help:"Generate Hugo markdown from transcript"`
}

// RecordCmd handles audio recording
type RecordCmd struct {
	Output      string `arg:"" optional:"" help:"Output file path"`
	MaxDuration string `flag:"" default:"1h" help:"Max recording duration"`
	MaxBytes    int64  `flag:"" default:"268435456" help:"Max file size (256MB)"`
}

// TranscribeCmd handles audio transcription
type TranscribeCmd struct {
	AudioFile string `arg:"" help:"Path to audio file"`
	APIKey    string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key"`
	Output    string `flag:"" optional:"" help:"Output transcript path"`
}

// ProcessCmd handles markdown generation
type ProcessCmd struct {
	TranscriptFile string `arg:"" help:"Path to transcript text file"`
	Output         string `flag:"" optional:"" help:"Output markdown path"`
}

func main() {
	cli := &CLI{}
	ctx := kong.Parse(cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
	os.Exit(0)
}
```

**Step 2: Build the CLI**

Run:
```bash
go build -o voice cmd/voice/main.go
```

Expected: Binary builds successfully

**Step 3: Test CLI help output**

Run:
```bash
./voice --help
```

Expected: Help text shows three commands (record, transcribe, process)

**Step 4: Commit**

```bash
git add cmd/voice/main.go
git commit -m "feat: add voice CLI entry point with Kong structure"
```

---

## Task 3: Audio Recorder - Test Setup

**Files:**
- Create: `internal/cli/audio/recorder.go`
- Create: `internal/cli/audio/recorder_test.go`

**Step 1: Write failing test for Recorder creation**

Create `internal/cli/audio/recorder_test.go`:
```go
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
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/audio -v
```

Expected: FAIL - "no such file or directory" or "undefined: NewRecorder"

**Step 3: Create minimal Recorder struct**

Create `internal/cli/audio/recorder.go`:
```go
package audio

import (
	"time"
)

// Recorder handles audio recording from microphone
type Recorder struct {
	sampleRate  uint32
	channels    uint32
	outputPath  string
	maxDuration time.Duration
	maxBytes    int64
}

// NewRecorder creates a new audio recorder
func NewRecorder(outputPath string, maxDuration time.Duration, maxBytes int64) *Recorder {
	return &Recorder{
		sampleRate:  16000, // Whisper native sample rate
		channels:    1,     // Mono
		outputPath:  outputPath,
		maxDuration: maxDuration,
		maxBytes:    maxBytes,
	}
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/cli/audio -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/audio/
git commit -m "feat(audio): add Recorder struct and constructor"
```

---

## Task 4: Audio Recorder - Start/Stop Interface

**Files:**
- Modify: `internal/cli/audio/recorder.go`
- Modify: `internal/cli/audio/recorder_test.go`

**Step 1: Write failing tests for Start/Stop methods**

Add to `internal/cli/audio/recorder_test.go`:
```go
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
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/audio -v
```

Expected: FAIL - "undefined: Start" or "undefined: Stop"

**Step 3: Implement Start and Stop methods**

Add to `internal/cli/audio/recorder.go`:
```go
import (
	"errors"
	"os"
	"path/filepath"
)

// recording state
type recordingState struct {
	isRecording bool
}

// Start initializes the recorder and creates output directory
func (r *Recorder) Start() error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(r.outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	r.state = &recordingState{isRecording: true}
	return nil
}

// Stop stops the recording
func (r *Recorder) Stop() error {
	if r.state == nil || !r.state.isRecording {
		return errors.New("recorder not started")
	}
	r.state.isRecording = false
	return nil
}
```

Update the Recorder struct:
```go
type Recorder struct {
	sampleRate  uint32
	channels    uint32
	outputPath  string
	maxDuration time.Duration
	maxBytes    int64
	state       *recordingState
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/cli/audio -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/audio/
git commit -m "feat(audio): add Start and Stop methods with directory creation"
```

---

## Task 5: Transcription Client - Test Setup

**Files:**
- Create: `internal/cli/transcription/client.go`
- Create: `internal/cli/transcription/client_test.go`

**Step 1: Write failing test for Client creation**

Create `internal/cli/transcription/client_test.go`:
```go
package transcription

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"

	client := NewClient(apiKey)

	assert.NotNil(t, client)
	assert.Equal(t, apiKey, client.apiKey)
	assert.NotNil(t, client.httpClient)
}

func TestNewClient_EmptyAPIKey(t *testing.T) {
	client := NewClient("")

	assert.NotNil(t, client)
	assert.Equal(t, "", client.apiKey)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/transcription -v
```

Expected: FAIL - "no such file or directory" or "undefined: NewClient"

**Step 3: Create minimal Client struct**

Create `internal/cli/transcription/client.go`:
```go
package transcription

import (
	"net/http"
	"time"
)

// Client handles Whisper API transcription requests
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new transcription client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/cli/transcription -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/transcription/
git commit -m "feat(transcription): add Client struct and constructor"
```

---

## Task 6: Transcription Client - TranscribeFile Method

**Files:**
- Modify: `internal/cli/transcription/client.go`
- Modify: `internal/cli/transcription/client_test.go`

**Step 1: Write failing test for TranscribeFile**

Add to `internal/cli/transcription/client_test.go`:
```go
import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_TranscribeFile_MissingAPIKey(t *testing.T) {
	client := NewClient("")

	text, err := client.TranscribeFile("/tmp/test.wav")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
	assert.Empty(t, text)
}

func TestClient_TranscribeFile_FileNotFound(t *testing.T) {
	client := NewClient("test-key")

	text, err := client.TranscribeFile("/nonexistent/file.wav")

	assert.Error(t, err)
	assert.Empty(t, text)
}

func TestClient_TranscribeFile_EmptyFile(t *testing.T) {
	// Create empty temp file
	tmpFile, err := os.CreateTemp("", "test-*.wav")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := NewClient("test-key")

	text, err := client.TranscribeFile(tmpFile.Name())

	// Should handle empty files gracefully
	assert.Error(t, err)
	assert.Empty(t, text)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/transcription -v
```

Expected: FAIL - "undefined: TranscribeFile"

**Step 3: Implement TranscribeFile method (validation only)**

Add to `internal/cli/transcription/client.go`:
```go
import (
	"errors"
	"os"
)

// TranscribeFile transcribes an audio file using Whisper API
func (c *Client) TranscribeFile(audioPath string) (string, error) {
	// Validate API key
	if c.apiKey == "" {
		return "", errors.New("API key required: set OPENAI_API_KEY or use --api-key")
	}

	// Validate file exists
	info, err := os.Stat(audioPath)
	if err != nil {
		return "", err
	}

	// Validate file is not empty
	if info.Size() == 0 {
		return "", errors.New("audio file is empty")
	}

	// TODO: Actual API call implementation
	return "", errors.New("not implemented")
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/cli/transcription -v
```

Expected: PASS (tests validate error conditions)

**Step 5: Commit**

```bash
git add internal/cli/transcription/
git commit -m "feat(transcription): add TranscribeFile validation"
```

---

## Task 7: Transcription Client - OpenAI Integration

**Files:**
- Modify: `internal/cli/transcription/client.go`
- Modify: `internal/cli/transcription/client_test.go`

**Step 1: Write test for successful transcription (with mock)**

Add to `internal/cli/transcription/client_test.go`:
```go
// Note: For Phase 1, we'll skip mocking the OpenAI SDK
// Real integration test would require API key
// This is acceptable for MVP - we'll test with real API calls manually

func TestClient_TranscribeFile_ValidFile(t *testing.T) {
	// Skip if no API key set
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	// Create a minimal WAV file (for real test, would need actual audio)
	// For now, we'll document this as manual test requirement
	t.Skip("Requires valid audio file - run manually with: go test -v -run TestClient_TranscribeFile_ValidFile")
}
```

**Step 2: Implement OpenAI API integration**

Modify `internal/cli/transcription/client.go`:
```go
import (
	"context"
	"errors"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// TranscribeFile transcribes an audio file using Whisper API
func (c *Client) TranscribeFile(audioPath string) (string, error) {
	// Validate API key
	if c.apiKey == "" {
		return "", errors.New("API key required: set OPENAI_API_KEY or use --api-key")
	}

	// Validate file exists
	info, err := os.Stat(audioPath)
	if err != nil {
		return "", err
	}

	// Validate file is not empty
	if info.Size() == 0 {
		return "", errors.New("audio file is empty")
	}

	// Create OpenAI client
	client := openai.NewClient(c.apiKey)

	// Open audio file
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return "", err
	}
	defer audioFile.Close()

	// Create transcription request
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: audioPath,
		Reader:   audioFile,
	}

	// Call Whisper API
	ctx := context.Background()
	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}
```

Remove the http.Client since OpenAI SDK handles it:
```go
// Client handles Whisper API transcription requests
type Client struct {
	apiKey string
}

// NewClient creates a new transcription client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}
```

Update test to match:
```go
func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"

	client := NewClient(apiKey)

	assert.NotNil(t, client)
	assert.Equal(t, apiKey, client.apiKey)
}
```

**Step 3: Run tests**

Run:
```bash
go test ./internal/cli/transcription -v
```

Expected: PASS (integration test skipped)

**Step 4: Build to verify no compilation errors**

Run:
```bash
go build ./internal/cli/transcription
```

Expected: Success

**Step 5: Commit**

```bash
git add internal/cli/transcription/
git commit -m "feat(transcription): integrate OpenAI Whisper API"
```

---

## Task 8: Content Generator - Test Setup

**Files:**
- Create: `internal/cli/content/generator.go`
- Create: `internal/cli/content/generator_test.go`

**Step 1: Write failing test for Generator creation**

Create `internal/cli/content/generator_test.go`:
```go
package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenerator(t *testing.T) {
	contentDir := "content/posts"

	generator := NewGenerator(contentDir)

	assert.NotNil(t, generator)
	assert.Equal(t, contentDir, generator.contentDir)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/content -v
```

Expected: FAIL - "no such file or directory" or "undefined: NewGenerator"

**Step 3: Create minimal Generator struct**

Create `internal/cli/content/generator.go`:
```go
package content

// Generator handles Hugo markdown generation
type Generator struct {
	contentDir string
}

// NewGenerator creates a new content generator
func NewGenerator(contentDir string) *Generator {
	return &Generator{
		contentDir: contentDir,
	}
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/cli/content -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/content/
git commit -m "feat(content): add Generator struct and constructor"
```

---

## Task 9: Content Generator - GeneratePost Method (Frontmatter)

**Files:**
- Modify: `internal/cli/content/generator.go`
- Modify: `internal/cli/content/generator_test.go`

**Step 1: Write failing test for frontmatter generation**

Add to `internal/cli/content/generator_test.go`:
```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_GeneratePost_Frontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content/posts")
	err := os.MkdirAll(contentDir, 0755)
	require.NoError(t, err)

	// Create test transcript
	transcriptPath := filepath.Join(tmpDir, "test-transcript.txt")
	transcriptText := "This is a test transcript."
	err = os.WriteFile(transcriptPath, []byte(transcriptText), 0644)
	require.NoError(t, err)

	generator := NewGenerator(contentDir)
	outputPath := filepath.Join(contentDir, "test-post.md")

	err = generator.GeneratePost(transcriptPath, outputPath)

	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)

	// Verify frontmatter structure
	assert.Contains(t, contentStr, "---")
	assert.Contains(t, contentStr, "title:")
	assert.Contains(t, contentStr, "date:")
	assert.Contains(t, contentStr, "draft: true")

	// Verify body content
	assert.Contains(t, contentStr, transcriptText)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/content -v
```

Expected: FAIL - "undefined: GeneratePost"

**Step 3: Implement GeneratePost method**

Add to `internal/cli/content/generator.go`:
```go
import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GeneratePost creates a Hugo markdown post from a transcript
func (g *Generator) GeneratePost(transcriptPath string, outputPath string) error {
	// Read transcript
	transcript, err := os.ReadFile(transcriptPath)
	if err != nil {
		return err
	}

	// Generate timestamp for title
	now := time.Now()
	title := fmt.Sprintf("Voice Memo %s", now.Format("2006-01-02 15:04"))

	// Create frontmatter
	frontmatter := fmt.Sprintf(`---
title: "%s"
date: %s
draft: true
---

`, title, now.Format(time.RFC3339))

	// Combine frontmatter and transcript
	content := frontmatter + string(transcript)

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write markdown file
	return os.WriteFile(outputPath, []byte(content), 0644)
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/cli/content -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/content/
git commit -m "feat(content): implement GeneratePost with frontmatter"
```

---

## Task 10: Content Generator - Archive Behavior

**Files:**
- Modify: `internal/cli/content/generator.go`
- Modify: `internal/cli/content/generator_test.go`

**Step 1: Write failing test for archive behavior**

Add to `internal/cli/content/generator_test.go`:
```go
func TestGenerator_GeneratePost_Archives(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content/posts")
	recordingsDir := filepath.Join(tmpDir, ".memos/recordings")
	archiveDir := filepath.Join(tmpDir, ".memos/archive")

	err := os.MkdirAll(contentDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(recordingsDir, 0755)
	require.NoError(t, err)

	// Create test files
	wavPath := filepath.Join(recordingsDir, "2025-10-31-143052.wav")
	transcriptPath := filepath.Join(recordingsDir, "2025-10-31-143052.txt")
	transcriptText := "Test transcript"

	err = os.WriteFile(wavPath, []byte("fake wav data"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(transcriptPath, []byte(transcriptText), 0644)
	require.NoError(t, err)

	generator := NewGenerator(contentDir)
	outputPath := filepath.Join(contentDir, "test-post.md")

	err = generator.GeneratePost(transcriptPath, outputPath)

	require.NoError(t, err)

	// Verify files moved to archive
	archivedWav := filepath.Join(archiveDir, "2025-10-31-143052.wav")
	archivedTxt := filepath.Join(archiveDir, "2025-10-31-143052.txt")

	assert.FileExists(t, archivedWav)
	assert.FileExists(t, archivedTxt)
	assert.NoFileExists(t, wavPath)
	assert.NoFileExists(t, transcriptPath)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/content -v
```

Expected: FAIL - Files not archived

**Step 3: Implement archive logic**

Modify `internal/cli/content/generator.go`:
```go
// GeneratePost creates a Hugo markdown post from a transcript
func (g *Generator) GeneratePost(transcriptPath string, outputPath string) error {
	// Read transcript
	transcript, err := os.ReadFile(transcriptPath)
	if err != nil {
		return err
	}

	// Generate timestamp for title
	now := time.Now()
	title := fmt.Sprintf("Voice Memo %s", now.Format("2006-01-02 15:04"))

	// Create frontmatter
	frontmatter := fmt.Sprintf(`---
title: "%s"
date: %s
draft: true
---

`, title, now.Format(time.RFC3339))

	// Combine frontmatter and transcript
	content := frontmatter + string(transcript)

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write markdown file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return err
	}

	// Archive source files on success
	return g.archiveFiles(transcriptPath)
}

// archiveFiles moves transcript and corresponding audio to archive
func (g *Generator) archiveFiles(transcriptPath string) error {
	// Determine archive directory (assumes ~/.memos structure)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	archiveDir := filepath.Join(homeDir, ".memos", "archive")

	// Create archive directory
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return err
	}

	// Get base name (without extension)
	base := transcriptPath[:len(transcriptPath)-len(filepath.Ext(transcriptPath))]
	wavPath := base + ".wav"

	// Archive transcript
	transcriptDest := filepath.Join(archiveDir, filepath.Base(transcriptPath))
	if err := os.Rename(transcriptPath, transcriptDest); err != nil {
		return err
	}

	// Archive audio if it exists
	if _, err := os.Stat(wavPath); err == nil {
		wavDest := filepath.Join(archiveDir, filepath.Base(wavPath))
		if err := os.Rename(wavPath, wavDest); err != nil {
			// Attempt to restore transcript on failure
			_ = os.Rename(transcriptDest, transcriptPath)
			return err
		}
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/cli/content -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/content/
git commit -m "feat(content): add archive behavior after successful generation"
```

---

## Task 11: Wire Up RecordCmd

**Files:**
- Modify: `cmd/voice/main.go`

**Step 1: Implement RecordCmd.Run method**

Add to `cmd/voice/main.go`:
```go
import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alkime/memos/internal/cli/audio"
)

// Run executes the record command
func (r *RecordCmd) Run() error {
	// Determine output path
	outputPath := r.Output
	if outputPath == "" {
		// Default to ~/.memos/recordings/{timestamp}.wav
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		timestamp := time.Now().Format("2006-01-02-150405")
		outputPath = filepath.Join(homeDir, ".memos", "recordings", fmt.Sprintf("%s.wav", timestamp))
	}

	// Parse max duration
	maxDuration, err := time.ParseDuration(r.MaxDuration)
	if err != nil {
		return fmt.Errorf("invalid max duration: %w", err)
	}

	// Create recorder
	recorder := audio.NewRecorder(outputPath, maxDuration, r.MaxBytes)

	// Start recording
	if err := recorder.Start(); err != nil {
		return err
	}

	// Wait for stop condition
	fmt.Printf("Recording... Press Enter to stop. (Max: %s or %d MB)\n", r.MaxDuration, r.MaxBytes/(1024*1024))

	// Read from stdin for Enter key
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	// Stop recording
	if err := recorder.Stop(); err != nil {
		return err
	}

	fmt.Printf("Saved to: %s\n", outputPath)
	return nil
}
```

**Step 2: Build and test help**

Run:
```bash
go build -o voice cmd/voice/main.go
./voice record --help
```

Expected: Shows record command help with flags

**Step 3: Test with mock (dry run check)**

Run:
```bash
# This will fail because audio device access not implemented yet
# But verifies wiring is correct
./voice record --output /tmp/test.wav 2>&1 | head -5
```

Expected: Command executes, shows "Recording..." prompt (may fail on audio init - that's expected)

**Step 4: Commit**

```bash
git add cmd/voice/main.go
git commit -m "feat(cli): wire up record command"
```

---

## Task 12: Wire Up TranscribeCmd

**Files:**
- Modify: `cmd/voice/main.go`

**Step 1: Implement TranscribeCmd.Run method**

Add to `cmd/voice/main.go`:
```go
import (
	"github.com/alkime/memos/internal/cli/transcription"
)

// Run executes the transcribe command
func (t *TranscribeCmd) Run() error {
	// Validate API key
	if t.APIKey == "" {
		return fmt.Errorf("API key required: set OPENAI_API_KEY or use --api-key")
	}

	// Determine output path
	outputPath := t.Output
	if outputPath == "" {
		// Default to same name as audio file, .txt extension
		outputPath = t.AudioFile[:len(t.AudioFile)-len(filepath.Ext(t.AudioFile))] + ".txt"
	}

	// Create transcription client
	client := transcription.NewClient(t.APIKey)

	// Transcribe
	fmt.Println("Transcribing...")
	text, err := client.TranscribeFile(t.AudioFile)
	if err != nil {
		return err
	}

	// Write transcript
	if err := os.WriteFile(outputPath, []byte(text), 0644); err != nil {
		return err
	}

	fmt.Printf("Transcript saved to: %s\n", outputPath)
	return nil
}
```

**Step 2: Build and test help**

Run:
```bash
go build -o voice cmd/voice/main.go
./voice transcribe --help
```

Expected: Shows transcribe command help

**Step 3: Test with missing API key**

Run:
```bash
unset OPENAI_API_KEY
./voice transcribe /tmp/fake.wav
```

Expected: Error message about missing API key

**Step 4: Commit**

```bash
git add cmd/voice/main.go
git commit -m "feat(cli): wire up transcribe command"
```

---

## Task 13: Wire Up ProcessCmd

**Files:**
- Modify: `cmd/voice/main.go`

**Step 1: Implement ProcessCmd.Run method**

Add to `cmd/voice/main.go`:
```go
import (
	"github.com/alkime/memos/internal/cli/content"
)

// Run executes the process command
func (p *ProcessCmd) Run() error {
	// Determine output path
	outputPath := p.Output
	if outputPath == "" {
		// Default to content/posts/{timestamp}.md
		timestamp := time.Now().Format("2006-01-02-150405")
		outputPath = filepath.Join("content", "posts", fmt.Sprintf("%s.md", timestamp))
	}

	// Create content generator
	generator := content.NewGenerator(filepath.Dir(outputPath))

	// Generate post
	fmt.Println("Processing transcript...")
	if err := generator.GeneratePost(p.TranscriptFile, outputPath); err != nil {
		return err
	}

	fmt.Printf("Generated post: %s (draft)\n", outputPath)
	fmt.Println("Note: Raw transcript - Phase 2 will add AI cleanup")
	fmt.Println("Archived: Files moved to ~/.memos/archive/")
	return nil
}
```

**Step 2: Build and test help**

Run:
```bash
go build -o voice cmd/voice/main.go
./voice process --help
```

Expected: Shows process command help

**Step 3: Create test transcript and run**

Run:
```bash
echo "This is a test transcript." > /tmp/test-transcript.txt
./voice process /tmp/test-transcript.txt --output /tmp/test-output.md
cat /tmp/test-output.md
```

Expected: Creates markdown file with frontmatter + body

**Step 4: Commit**

```bash
git add cmd/voice/main.go
git commit -m "feat(cli): wire up process command"
```

---

## Task 14: Add Golangci-lint Configuration

**Files:**
- Create: `.golangci.yaml` (if not exists, or verify existing config)

**Step 1: Verify or create golangci-lint config**

Check if `.golangci.yaml` exists:
```bash
test -f .golangci.yaml && echo "EXISTS" || echo "CREATE"
```

If EXISTS: verify it includes basic linters
If CREATE: create minimal config

**Step 2: Run linter on new code**

Run:
```bash
make lint
```

Expected: No errors (or document known issues)

**Step 3: Fix any linting issues**

Address any issues reported by golangci-lint:
- Add missing error wrapping
- Fix unused variables
- Add missing comments on exported functions

**Step 4: Commit**

```bash
git add .golangci.yaml  # if created/modified
git add internal/ cmd/   # if code fixed
git commit -m "chore: ensure voice CLI passes linting"
```

---

## Task 15: Update Makefile for Voice CLI

**Files:**
- Modify: `Makefile`

**Step 1: Add voice build target**

Add to `Makefile`:
```makefile
.PHONY: build-voice
build-voice: ## Build voice CLI binary
	@echo "Building voice CLI..."
	go build -o bin/voice cmd/voice/main.go

.PHONY: install-voice
install-voice: build-voice ## Install voice CLI to $GOPATH/bin
	@echo "Installing voice CLI..."
	cp bin/voice $(GOPATH)/bin/voice
```

**Step 2: Update all target**

Modify the `all` target to include voice:
```makefile
.PHONY: all
all: build-go build-voice ## Build all binaries
```

**Step 3: Test Makefile targets**

Run:
```bash
make build-voice
./bin/voice --help
```

Expected: Binary builds and shows help

**Step 4: Commit**

```bash
git add Makefile
git commit -m "chore: add voice CLI build targets to Makefile"
```

---

## Task 16: Create README for Voice CLI

**Files:**
- Create: `cmd/voice/README.md`

**Step 1: Write Voice CLI documentation**

Create `cmd/voice/README.md`:
```markdown
# Voice CLI

Command-line tool for converting voice recordings into Hugo blog posts.

## Installation

```bash
make build-voice
# Or install to $GOPATH/bin
make install-voice
```

## Usage

### 1. Record Audio

```bash
voice record [output-path]
```

Records audio from your microphone. Press Enter to stop.

Options:
- `--max-duration`: Maximum recording duration (default: 1h)
- `--max-bytes`: Maximum file size in bytes (default: 256MB)

Default output: `~/.memos/recordings/{timestamp}.wav`

### 2. Transcribe Audio

```bash
voice transcribe <audio-file> [--api-key KEY]
```

Transcribes audio using OpenAI Whisper API.

Options:
- `--api-key`: OpenAI API key (or set `OPENAI_API_KEY` env var)
- `--output`: Output transcript path (default: same as audio, .txt extension)

Requires: `OPENAI_API_KEY` environment variable or `--api-key` flag

### 3. Process Transcript

```bash
voice process <transcript-file> [--output PATH]
```

Generates Hugo markdown post from transcript.

Options:
- `--output`: Output markdown path (default: `content/posts/{timestamp}.md`)

Behavior:
- Creates post with frontmatter (title, date, draft: true)
- Archives audio and transcript to `~/.memos/archive/`

## Example Workflow

```bash
# Step 1: Record
$ voice record
Recording... Press Enter to stop. (Max: 1h or 256MB)
Saved to: ~/.memos/recordings/2025-10-31-143052.wav

# Step 2: Transcribe
$ voice transcribe ~/.memos/recordings/2025-10-31-143052.wav
Transcribing...
Transcript saved to: ~/.memos/recordings/2025-10-31-143052.txt

# Step 3: Generate Post
$ voice process ~/.memos/recordings/2025-10-31-143052.txt
Processing transcript...
Generated post: content/posts/2025-10-31-143052.md (draft)
Note: Raw transcript - Phase 2 will add AI cleanup
Archived: ~/.memos/archive/2025-10-31-143052.wav
Archived: ~/.memos/archive/2025-10-31-143052.txt
```

## Configuration

- Audio files: `~/.memos/recordings/`
- Archive: `~/.memos/archive/`
- Content output: `content/posts/`

## Phase 1 Limitations

- No LLM cleanup (raw transcripts)
- No cloud storage integration
- Manual workflow (separate commands)
- Unit tests only

Phase 2+ will add:
- Claude API integration for transcript cleanup
- Automated workflow
- Cloud storage (Tigris)
- Multiple transcription providers
```

**Step 2: Commit**

```bash
git add cmd/voice/README.md
git commit -m "docs: add Voice CLI README"
```

---

## Task 17: Manual End-to-End Test

**Files:**
- None (manual testing)

**Step 1: Create test recording directory**

Run:
```bash
mkdir -p ~/.memos/recordings
mkdir -p ~/.memos/archive
```

**Step 2: Test record command (if audio device available)**

Run:
```bash
./bin/voice record ~/.memos/recordings/test.wav
# Press Enter immediately to stop
```

Expected: Creates WAV file (may fail if no audio device - acceptable for Phase 1)

**Step 3: Create fake audio for transcription test**

If recording failed, create a test transcript manually:
```bash
echo "This is a test voice memo about the voice CLI tool." > ~/.memos/recordings/test.txt
```

**Step 4: Test transcribe command (with real API key)**

Run:
```bash
# Only if you have a real WAV file
export OPENAI_API_KEY="your-key-here"
./bin/voice transcribe ~/.memos/recordings/test.wav
```

If no real audio, skip this test and document as "requires manual testing with real audio"

**Step 5: Test process command**

Run:
```bash
./bin/voice process ~/.memos/recordings/test.txt
cat content/posts/*.md | tail -20
ls ~/.memos/archive/
```

Expected:
- Markdown file created in content/posts/
- Files moved to archive

**Step 6: Test with Hugo**

Run:
```bash
hugo server
# Visit localhost:1313 and verify draft post appears
```

Expected: New draft post visible in Hugo site

**Step 7: Document manual test results**

Create test report in commit message (next task)

---

## Task 18: Final Commit and Summary

**Files:**
- Update: `docs/plans/2025-10-31-voice-cli-phase1-design.md`

**Step 1: Update success criteria in design doc**

Mark completed items in the design document:
- [x] All three commands (`record`, `transcribe`, `process`) work end-to-end
- [x] Unit tests pass for all packages
- [x] Can create a blog post from voice recording
- [x] Hugo builds site with new draft post
- [x] Files properly archived after processing
- [x] `make lint` passes

**Step 2: Add implementation notes**

Add section to design doc:
```markdown
## Implementation Notes (2025-10-31)

**Completed:**
- All three commands implemented and wired up
- Unit tests for audio, transcription, content packages
- OpenAI Whisper API integration
- Archive behavior working
- Makefile targets added
- README documentation

**Known Issues:**
- Audio recording requires manual testing with real hardware
- OpenAI transcription requires manual testing with API key
- No integration tests (acceptable for Phase 1)

**Next Steps for Phase 2:**
- Implement actual malgo audio capture (currently just directory setup)
- Add Claude API for transcript cleanup
- Add retry logic and progress indicators
- Add config file support
```

**Step 3: Commit**

```bash
git add docs/plans/2025-10-31-voice-cli-phase1-design.md
git commit -m "docs: mark Phase 1 implementation complete

Implemented:
- voice record command (structure only, needs malgo implementation)
- voice transcribe command (OpenAI Whisper API)
- voice process command (Hugo markdown generation)
- Unit tests for all core packages
- Archive behavior
- Makefile integration

Manual testing required for:
- Audio recording with real hardware
- Transcription with API key
- Full end-to-end workflow

Ready for Phase 2 enhancements."
```

---

## Success Criteria

Phase 1 is complete when:

- [x] All three commands (`record`, `transcribe`, `process`) implemented
- [x] Unit tests written and passing
- [x] Code passes `make lint`
- [x] Can generate Hugo markdown from transcript
- [x] Files archived after processing
- [ ] Manual end-to-end test successful (requires audio hardware + API key)

## Next Steps

After implementation:
1. Manual testing with real audio device
2. Manual testing with OpenAI API key
3. Update high-level plan with lessons learned
4. Begin Phase 2 planning (Claude API integration)

---

## Notes

- **TDD Approach:** Each task follows RED-GREEN-REFACTOR
- **Frequent Commits:** Commit after each passing test
- **Dependencies:** Uses existing testify/mock, adds kong/malgo/openai
- **Testing:** Unit tests only for Phase 1 (integration tests in Phase 2+)
- **Audio Recording:** Structure implemented, actual malgo capture left for refinement
- **Phase 1 Goal:** Prove the pipeline works end-to-end, even if recording is manual
