# Voice Recording Limits Implementation

**Date:** 2025-11-18
**Status:** Approved
**Related Files:** `internal/cli/audio/recorder.go`, `cmd/voice/main.go`

## Overview

Implement enforcement of recording limits (max duration and max bytes) for the voice CLI tool. Currently, these limits are configurable but not enforced (see TODO at `recorder.go:78`).

## Requirements

### Functional Requirements

1. **Limit Enforcement**: Stop recording when either limit is reached:
   - `MaxDuration`: Wall clock time since recording started
   - `MaxBytes`: Total bytes written to buffer

2. **Graceful Handling**: When limit reached:
   - Stop audio device cleanly
   - Save MP3 file with captured audio
   - Return sentinel error to stop workflow
   - Display clear message to user

3. **Progress Display**: Replace dot progress indicator with periodic status updates:
   - Show every 5 seconds: elapsed time, buffer size, percentages
   - Highlight (bold) when either metric >= 90%
   - Format: `Recording: HH:MM:SS / HH:MM:SS (XX%) | XXX.X MB / XXX.X MB (XX%)`

4. **Workflow Control**:
   - Normal completion (user pressed Enter) → continue to transcription
   - Limit reached → stop workflow, user manually continues if needed

### Non-Functional Requirements

1. **Performance**: Limit checks should be cheap (inline after each packet)
2. **Precision**: Detect limit immediately when exceeded (not batched)
3. **Thread Safety**: Safe concurrent access to buffer size from ticker goroutine

## Architecture

### Design Choice: Inline Checks with Progress Ticker

**Selected approach:** Inline limit checking with separate progress ticker goroutine

**Rationale:**
- Inline checks provide immediate, precise limit detection
- Limit checks are very cheap (two integer comparisons per packet)
- Separate ticker cleanly handles progress display without coupling
- Integrates well with existing goroutine structure

**Alternatives considered:**
- Dedicated monitor goroutine: Could overshoot limits between checks
- Batched checks: Requires managing limit logic in multiple places

### Component Architecture

```
FileRecorder.Go()
├─ Packet reading goroutine (modified)
│  ├─ Write packet to buffer
│  ├─ Increment atomic byte counter
│  ├─ Check MaxBytes (inline)
│  ├─ Check MaxDuration (inline)
│  └─ Break loop if limit hit → hardStop()
│
├─ Progress ticker goroutine (new)
│  ├─ Every 5 seconds
│  ├─ Read atomic byte counter
│  ├─ Calculate elapsed time
│  ├─ Calculate percentages
│  └─ Display formatted progress (bold if >= 90%)
│
└─ Stop signals goroutine (unchanged)
   └─ Waits for Enter, Ctrl+C, or context cancel
```

## Data Flow

### Normal Recording Flow

1. Start recording → capture `startTime := time.Now()`
2. Launch 3 goroutines:
   - Packet reader (with inline limit checks)
   - Progress ticker (displays status every 5s)
   - Stop signal listener (existing)
3. For each packet:
   - Write to buffer
   - Increment atomic counter: `bytesWritten.Add(int64(len(packet)))`
   - Check: `if bytesWritten.Load() >= MaxBytes` → break
   - Check: `if time.Since(startTime) >= MaxDuration` → break
4. On loop exit: call `hardStop(ctx, dev)` to stop device
5. Flush buffer to MP3 file
6. Return result (nil or sentinel error)

### Limit Detection Flow

When limit exceeded:
```go
// In packet reading loop
if bytesWritten.Load() >= r.config.MaxBytes {
    logger.Info("recording stopped", "reason", "max_bytes_reached",
                "bytes", bytesWritten.Load())
    hardStop(ctx, dev)
    break
}

if time.Since(startTime) >= r.config.MaxDuration {
    logger.Info("recording stopped", "reason", "max_duration_reached",
                "duration", time.Since(startTime))
    hardStop(ctx, dev)
    break
}

// After loop, before flushing MP3
if <duration limit hit> {
    defer func() { err = ErrMaxDurationReached }()
}
if <bytes limit hit> {
    defer func() { err = ErrMaxBytesReached }()
}

// Flush MP3 normally, then return sentinel error
```

## Implementation Details

### Sentinel Errors

Define in `internal/cli/audio/recorder.go`:

```go
var (
    ErrMaxDurationReached = errors.New("max duration reached")
    ErrMaxBytesReached    = errors.New("max bytes reached")
)
```

### Progress Display

**Format:**
```
Recording: 00:45:23 / 01:00:00 (75%) | 128.5 MB / 256.0 MB (50%)
```

**With warning (>= 90%):**
```
Recording: 00:54:12 / 01:00:00 (90%) | 128.5 MB / 256.0 MB (50%)
           ^^^^^^^^^^^^^^^^^^^^^^ bold
```

**ANSI codes:**
- Bold: `\033[1m{text}\033[0m`
- Helper: `formatWithWarning(text string, isWarning bool) string`

**Ticker implementation:**
```go
func (r *FileRecorder) displayProgress(
    startTime time.Time,
    bytesWritten *atomic.Int64,
) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            elapsed := time.Since(startTime)
            bytes := bytesWritten.Load()

            timePercent := int(float64(elapsed) / float64(r.config.MaxDuration) * 100)
            bytesPercent := int(float64(bytes) / float64(r.config.MaxBytes) * 100)

            // Format and display
            fmt.Printf("\rRecording: %s | %.1f MB / %.1f MB (%d%%)",
                formatDuration(elapsed, r.config.MaxDuration, timePercent >= 90),
                formatBytes(bytes, r.config.MaxBytes, bytesPercent >= 90),
                ...)
        case <-ctx.Done():
            return
        }
    }
}
```

### Thread Safety

**Problem**: Ticker goroutine reads buffer size while packet goroutine writes to it.

**Solution**: Use `atomic.Int64` counter instead of calling `buf.Len()`:
```go
var bytesWritten atomic.Int64

// In packet loop
n, err := buf.Write(packet)
if err == nil {
    bytesWritten.Add(int64(n))
}

// In ticker
currentBytes := bytesWritten.Load()
```

### Calling Code Integration

In `cmd/voice/main.go`, handle sentinel errors:

```go
err := recorder.Go(ctx)
if err != nil {
    if errors.Is(err, audio.ErrMaxDurationReached) {
        fmt.Printf("\nRecording stopped: Maximum duration (%s) reached\n",
                   r.MaxDuration)
        fmt.Println("Audio file saved. Run 'voice transcribe' to continue.")
        return nil // Stop workflow, exit successfully
    }
    if errors.Is(err, audio.ErrMaxBytesReached) {
        fmt.Printf("\nRecording stopped: Maximum size (%d MB) reached\n",
                   r.MaxBytes/(1024*1024))
        fmt.Println("Audio file saved. Run 'voice transcribe' to continue.")
        return nil // Stop workflow, exit successfully
    }

    return fmt.Errorf("recording failed: %w", err)
}
// Normal completion - continue workflow
```

## Edge Cases

### Input Validation

In `NewRecorder()`:
```go
if conf.MaxDuration <= 0 {
    return nil, errors.New("MaxDuration must be positive")
}
if conf.MaxBytes <= 0 {
    return nil, errors.New("MaxBytes must be positive")
}
```

### Very Small Limits

If limit is reached before first packet or very quickly:
- Still save the MP3 file with whatever was captured
- Return appropriate sentinel error
- Display message to user

### Concurrent Stop Signals

If user presses Enter while limit is being checked:
- Both mechanisms call `hardStop(ctx, dev)`
- Second call is idempotent (device already stopped)
- Return the first detected reason (limit takes precedence over manual stop)

## Testing Considerations

1. **Unit tests** for limit checking logic:
   - Mock audio device
   - Inject packets until limit reached
   - Verify sentinel error returned
   - Verify file saved

2. **Progress display tests**:
   - Capture stdout
   - Verify format and timing
   - Verify bold formatting at 90%

3. **Edge case tests**:
   - Zero/negative limits
   - Very small limits
   - Concurrent stop signals

## Success Criteria

1. Recording stops immediately when either limit reached
2. MP3 file saved successfully with captured audio
3. Clear, informative progress updates every 5 seconds
4. Visual warning (bold) when approaching limits (>= 90%)
5. Workflow stops gracefully, user can manually continue
6. No performance degradation from limit checks
7. All edge cases handled safely
