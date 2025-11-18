# Voice Recording Limits Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement enforcement of MaxDuration and MaxBytes limits with progress display for voice recording.

**Architecture:** Inline limit checking in packet loop + separate progress ticker goroutine. Uses atomic.Int64 for thread-safe byte counting and sentinel errors for workflow control.

**Tech Stack:** Go 1.21+, sync/atomic, time, ANSI escape codes for formatting

---

## Task 1: Add Sentinel Errors

**Files:**
- Modify: `internal/cli/audio/recorder.go:1-25`

**Step 1: Add sentinel error declarations**

Add after the constants block (after line 23):

```go
// Sentinel errors for limit detection
var (
	ErrMaxDurationReached = errors.New("max duration reached")
	ErrMaxBytesReached    = errors.New("max bytes reached")
)
```

**Step 2: Verify it compiles**

Run: `go build ./internal/cli/audio`
Expected: SUCCESS (no output)

**Step 3: Commit**

```bash
git add internal/cli/audio/recorder.go
git commit -m "feat(audio): add sentinel errors for recording limits

Add ErrMaxDurationReached and ErrMaxBytesReached to communicate
when recording stops due to limits.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 2: Add Config Validation

**Files:**
- Modify: `internal/cli/audio/recorder.go:38-42`
- Test: `internal/cli/audio/recorder_test.go`

**Step 1: Write failing test**

Add to `recorder_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/audio -run TestNewRecorder_ValidatesConfig -v`
Expected: FAIL (NewRecorder doesn't return error yet)

**Step 3: Update NewRecorder signature and add validation**

Modify `NewRecorder()` function (lines 38-42):

```go
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/cli/audio -run TestNewRecorder_ValidatesConfig -v`
Expected: PASS

**Step 5: Fix existing caller in main.go**

Modify `cmd/voice/main.go` (around line 165):

Change:
```go
recorder := audio.NewRecorder(audio.FileRecorderConfig{
```

To:
```go
recorder, err := audio.NewRecorder(audio.FileRecorderConfig{
```

Add error check after:
```go
if err != nil {
	return fmt.Errorf("failed to create recorder: %w", err)
}
```

**Step 6: Verify it builds**

Run: `go build ./cmd/voice`
Expected: SUCCESS

**Step 7: Commit**

```bash
git add internal/cli/audio/recorder.go internal/cli/audio/recorder_test.go cmd/voice/main.go
git commit -m "feat(audio): validate recorder config

Validate MaxDuration and MaxBytes are positive in NewRecorder.
Returns error for invalid config values.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 3: Add Progress Display Helpers

**Files:**
- Modify: `internal/cli/audio/recorder.go` (add at end before catchStopSignals)

**Step 1: Add formatting helper functions**

Add before `catchStopSignals()` function:

```go
// formatWithBold wraps text in ANSI bold codes if shouldBold is true
func formatWithBold(text string, shouldBold bool) string {
	if shouldBold {
		return fmt.Sprintf("\033[1m%s\033[0m", text)
	}
	return text
}

// formatDuration formats elapsed and max duration with optional bold
func formatDuration(elapsed, max time.Duration, shouldBold bool) string {
	// Format as HH:MM:SS
	formatTime := func(d time.Duration) string {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}

	elapsedStr := formatTime(elapsed)
	maxStr := formatTime(max)
	percent := int(float64(elapsed) / float64(max) * 100)

	text := fmt.Sprintf("%s / %s (%d%%)", elapsedStr, maxStr, percent)
	return formatWithBold(text, shouldBold)
}

// formatBytes formats bytes in MB with optional bold
func formatBytes(current, max int64, shouldBold bool) string {
	currentMB := float64(current) / (1024 * 1024)
	maxMB := float64(max) / (1024 * 1024)
	percent := int(float64(current) / float64(max) * 100)

	text := fmt.Sprintf("%.1f MB / %.1f MB (%d%%)", currentMB, maxMB, percent)
	return formatWithBold(text, shouldBold)
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/cli/audio`
Expected: SUCCESS

**Step 3: Add import for fmt if needed**

Ensure `fmt` is in the import block at the top of the file.

**Step 4: Commit**

```bash
git add internal/cli/audio/recorder.go
git commit -m "feat(audio): add progress display formatting helpers

Add formatWithBold, formatDuration, and formatBytes for displaying
recording progress with ANSI bold highlighting.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 4: Implement Limit Checking in Go() Method

**Files:**
- Modify: `internal/cli/audio/recorder.go:44-104`

**Step 1: Add imports**

Ensure these imports are present:
```go
import (
	// ... existing imports
	"sync/atomic"
)
```

**Step 2: Modify Go() method to add limit checking**

Replace the `Go()` method (lines 44-104) with:

```go
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

	// Track start time for duration limit
	startTime := time.Now()

	// Track bytes written atomically for thread-safe access
	var bytesWritten atomic.Int64

	// Track which limit was hit (if any)
	var limitReached error

	// spawn 3 goroutines:
	// -- read the data channel with limit checking
	// -- display progress periodically
	// -- listen for "finished" signals from ^C, Enter, or context
	wg := new(sync.WaitGroup)

	buf := bytes.NewBuffer(nil)

	// Packet reading goroutine with inline limit checks
	wg.Go(func() {
		for packet := range dataC {
			// Write packet to buffer
			n, err := buf.Write(packet)
			if err != nil {
				slog.Error("failed to write audio packet to buffer. halting....", "error", err)
				break
			}

			// Update atomic counter
			bytesWritten.Add(int64(n))

			// Inline limit checks
			if bytesWritten.Load() >= r.config.MaxBytes {
				slog.Info("recording stopped", "reason", "max_bytes_reached",
					"bytes", bytesWritten.Load())
				limitReached = ErrMaxBytesReached
				hardStop(ctx, dev)
				break
			}

			elapsed := time.Since(startTime)
			if elapsed >= r.config.MaxDuration {
				slog.Info("recording stopped", "reason", "max_duration_reached",
					"duration", elapsed)
				limitReached = ErrMaxDurationReached
				hardStop(ctx, dev)
				break
			}
		}
		fmt.Println() //nolint:forbidigo // Clear progress line
	})

	// Progress display goroutine
	wg.Go(func() {
		ticker := time.NewTicker(5 * time.Second)
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

				fmt.Printf("\rRecording: %s | %s", //nolint:forbidigo // CLI progress
					formatDuration(elapsed, r.config.MaxDuration, timeWarning),
					formatBytes(bytes, r.config.MaxBytes, bytesWarning))
			case <-ctx.Done():
				return
			}
		}
	})

	// Stop signals goroutine (existing)
	wg.Go(func() {
		<-catchStopSignals(ctx)
		slog.Info("received stop signal, stopping recording")
		hardStop(ctx, dev)
	})

	slog.Info("running... waiting for recording to finish")
	wg.Wait()
	slog.Info("recording finished", "buffer_size_bytes", buf.Len())

	// Flush buffer to MP3 file
	err = r.flushMP3File(r.config.OutputPath, buf)
	if err != nil {
		return fmt.Errorf("failed to flush MP3 file: %w", err)
	}

	// Return sentinel error if limit was reached
	if limitReached != nil {
		return limitReached
	}

	return nil
}
```

**Step 3: Verify it compiles**

Run: `go build ./internal/cli/audio`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/cli/audio/recorder.go
git commit -m "feat(audio): implement recording limit enforcement

Add inline limit checking for MaxDuration and MaxBytes with atomic
byte counting. Display progress every 5s with bold warning at 90%.
Return sentinel errors when limits reached.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 5: Update Calling Code to Handle Sentinel Errors

**Files:**
- Modify: `cmd/voice/main.go` (RecordCmd.Run method, around line 170)

**Step 1: Add imports**

Ensure these imports are present in `cmd/voice/main.go`:
```go
import (
	// ... existing imports
	"errors"
)
```

**Step 2: Update RecordCmd.Run to handle sentinel errors**

Find the line where `recorder.Go(ctx)` is called (around line 170) and update:

Change from:
```go
if err := recorder.Go(ctx); err != nil {
	return fmt.Errorf("failed to record audio: %w", err)
}
```

To:
```go
err = recorder.Go(ctx)
if err != nil {
	// Check for limit errors - these are not failures
	if errors.Is(err, audio.ErrMaxDurationReached) {
		logger.Info("recording stopped due to max duration limit")
		fmt.Printf("\nRecording stopped: Maximum duration (%s) reached\n", r.MaxDuration) //nolint:forbidigo // CLI output
		fmt.Println("Audio file saved. Run 'voice transcribe' to continue manually.") //nolint:forbidigo // CLI output
		return nil // Stop workflow, but exit successfully
	}
	if errors.Is(err, audio.ErrMaxBytesReached) {
		logger.Info("recording stopped due to max bytes limit")
		fmt.Printf("\nRecording stopped: Maximum size (%d MB) reached\n", r.MaxBytes/(1024*1024)) //nolint:forbidigo // CLI output
		fmt.Println("Audio file saved. Run 'voice transcribe' to continue manually.") //nolint:forbidigo // CLI output
		return nil // Stop workflow, but exit successfully
	}

	// Actual error
	return fmt.Errorf("failed to record audio: %w", err)
}
```

**Step 3: Verify it builds**

Run: `go build ./cmd/voice`
Expected: SUCCESS

**Step 4: Test manually (optional)**

Run: `./bin/voice record --max-duration 5s`
Expected: Recording stops after 5 seconds with appropriate message

**Step 5: Commit**

```bash
git add cmd/voice/main.go
git commit -m "feat(voice): handle recording limit errors

Detect ErrMaxDurationReached and ErrMaxBytesReached sentinel errors.
Display clear message and stop workflow gracefully when limits hit.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 6: Add Tests for Limit Enforcement

**Files:**
- Test: `internal/cli/audio/recorder_test.go`

**Step 1: Write test for formatting helpers**

Add to `recorder_test.go`:

```go
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
			} else {
				// Should not have ANSI codes
				if got != tt.wantPlain {
					t.Errorf("formatDuration() = %q, want %q", got, tt.wantPlain)
				}
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
			} else {
				// Should not have ANSI codes
				if got != tt.wantPlain {
					t.Errorf("formatBytes() = %q, want %q", got, tt.wantPlain)
				}
			}
		})
	}
}
```

**Step 2: Run tests to verify they pass**

Run: `go test ./internal/cli/audio -run TestFormat -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/cli/audio/recorder_test.go
git commit -m "test(audio): add tests for progress formatting

Test formatWithBold, formatDuration, and formatBytes helpers.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 7: Run Full Test Suite and Lint

**Step 1: Run all tests**

Run: `go test ./...`
Expected: All tests PASS

**Step 2: Run linter**

Run: `make lint`
Expected: No errors (or only acceptable nolint directives)

**Step 3: If linter fails, fix issues**

Common fixes:
- Add `//nolint:forbidigo // CLI progress` for fmt.Printf in progress display
- Ensure all errors are wrapped with context

**Step 4: Re-run tests after fixes**

Run: `go test ./...`
Expected: PASS

**Step 5: Commit any linter fixes**

```bash
git add .
git commit -m "fix(audio): address linter feedback

Add appropriate nolint directives for CLI output.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 8: Manual Testing

**Step 1: Build the voice binary**

Run: `go build -o bin/voice ./cmd/voice`
Expected: Binary created at `bin/voice`

**Step 2: Test with short duration limit**

Run: `./bin/voice record --max-duration 10s`
Expected:
- Recording starts
- Progress updates every 5s
- Recording stops after 10s with message: "Recording stopped: Maximum duration (10s) reached"
- File saved to `~/.memos/work/<branch>/recording.mp3`

**Step 3: Test with small byte limit**

Run: `./bin/voice record --max-bytes 1048576` (1 MB)
Expected:
- Recording starts
- Progress updates show bytes increasing
- Recording stops when ~1 MB reached
- Message: "Recording stopped: Maximum size (1 MB) reached"

**Step 4: Test normal stop (Enter key)**

Run: `./bin/voice record`
Expected:
- Recording starts
- Progress updates display
- Press Enter
- Recording stops normally
- Continues to transcription (if API key configured)

**Step 5: Verify bold formatting at 90%**

Run: `./bin/voice record --max-duration 20s`
Expected:
- At ~18 seconds, duration display shows bold (if terminal supports ANSI)
- Recording continues until 20s then stops

---

## Task 9: Update README Documentation

**Files:**
- Modify: `cmd/voice/README.md`

**Step 1: Document limit behavior**

Add section after "Recording Limits" (around line 60):

```markdown
#### Limit Behavior

When a recording limit is reached:

1. **Recording stops automatically** - The audio device is stopped gracefully
2. **File is saved** - The MP3 file is saved with all audio captured up to the limit
3. **Workflow stops** - The voice tool exits successfully without continuing to transcription
4. **Clear message** - A message indicates which limit was reached

To continue processing after a limit is reached, manually run:
```bash
voice transcribe
voice first-draft
voice copy-edit
```

#### Progress Display

During recording, progress updates appear every 5 seconds showing:
- Elapsed time and percentage of max duration
- Bytes recorded and percentage of max size
- **Bold highlighting** when either metric reaches 90% of the limit

Example:
```
Recording: 00:54:00 / 01:00:00 (90%) | 128.5 MB / 256.0 MB (50%)
```

The time portion is shown in bold to warn you're approaching the 1-hour limit.
```

**Step 2: Commit**

```bash
git add cmd/voice/README.md
git commit -m "docs(voice): document recording limit behavior

Explain what happens when limits are reached and document
progress display format.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Success Criteria

✅ Recording stops immediately when MaxDuration reached
✅ Recording stops immediately when MaxBytes reached
✅ MP3 file saved successfully when limits hit
✅ Sentinel errors returned (ErrMaxDurationReached, ErrMaxBytesReached)
✅ Calling code handles sentinel errors and stops workflow
✅ Progress display updates every 5 seconds
✅ Bold formatting appears at 90% of limits
✅ Config validation prevents invalid limits
✅ All tests pass
✅ Linter passes
✅ Manual testing confirms expected behavior
✅ README documents limit behavior

## Notes

- **Testing limits:** Use small values (e.g., `--max-duration 10s`) for faster testing
- **ANSI codes:** Bold formatting may not appear in all terminals (VS Code integrated terminal, basic terminals)
- **Thread safety:** atomic.Int64 ensures safe concurrent access without mutex overhead
- **Error handling:** Limit errors are expected behavior, not failures - workflow stops gracefully
