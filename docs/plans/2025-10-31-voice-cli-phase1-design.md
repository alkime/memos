# Voice CLI Phase 1 Design

**Date:** 2025-10-31
**Status:** Approved for Implementation
**Phase:** 1 - MVP Foundation

---

## Overview

Build a Go CLI tool (`voice`) that automates the voice-to-blog workflow for Phase 1 MVP. Proves end-to-end workflow with minimal complexity: record audio → transcribe with Whisper API → generate Hugo markdown draft.

**Key Principle:** Manual workflow (run each command separately). No automation, no LLM cleanup, no cloud storage. Just the essential pipeline.

---

## Architecture

### Project Structure

```
cmd/voice/                      # CLI entry point (new)
  main.go                       # Kong CLI setup, command routing

internal/cli/                   # CLI-specific packages (new namespace)
  audio/                        # Audio recording with malgo
    recorder.go
    recorder_test.go
  transcription/                # Whisper API client
    client.go
    client_test.go
  content/                      # Hugo markdown generation
    generator.go
    generator_test.go
```

**User Directories:**
```
~/.memos/
  recordings/                   # Active recordings (temp storage)
  archive/                      # Processed recordings moved here after success
```

### Design Rationale

**Why `internal/cli/` namespace?**
- Separates CLI code from existing server code (`internal/config`, `internal/logger`, `internal/server`)
- Clear ownership boundaries
- Scales cleanly when Phase 2+ adds more CLI packages

**Why separate `cmd/voice/`?**
- Independent binary from `cmd/server`
- Different concerns: local CLI vs web server
- Can be built and versioned separately if needed

---

## Dependencies

### New Dependencies to Add

```go
// CLI framework
github.com/alecthomas/kong

// Audio recording (zero system dependencies)
github.com/gen2brain/malgo

// Whisper API (research exact package during implementation)
// Likely: OpenAI Go SDK or direct HTTP client
```

### Testing Dependencies

Already in go.mod:
- `github.com/stretchr/testify` (including `testify/mock`)

---

## Command Interface

### CLI Structure (Kong)

```go
type CLI struct {
    Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
    Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
    Process    ProcessCmd    `cmd:"" help:"Generate Hugo markdown from transcript"`
}

type RecordCmd struct {
    Output      string        `arg:"" optional:"" help:"Output file path"`
    MaxDuration time.Duration `flag:"" default:"1h" help:"Max recording duration"`
    MaxBytes    int64         `flag:"" default:"268435456" help:"Max file size (256MB)"`
}

type TranscribeCmd struct {
    AudioFile string `arg:"" help:"Path to audio file"`
    APIKey    string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key"`
    Output    string `flag:"" optional:"" help:"Output transcript path"`
}

type ProcessCmd struct {
    TranscriptFile string `arg:"" help:"Path to transcript text file"`
    Output         string `flag:"" optional:"" help:"Output markdown path"`
}
```

### Example Workflow

```bash
# Step 1: Record
$ voice record
Recording... Press Enter to stop. (Max: 1h or 256MB)
Saved to: ~/.memos/recordings/2025-10-31-143052.wav (2.4 MB, 30s)

# Step 2: Transcribe
$ voice transcribe ~/.memos/recordings/2025-10-31-143052.wav
Transcribing...
Transcript saved to: ~/.memos/recordings/2025-10-31-143052.txt

# Step 3: Generate markdown
$ voice process ~/.memos/recordings/2025-10-31-143052.txt
Processing transcript...
Generated post: content/posts/2025-10-31-143052.md (draft)
Note: Raw transcript - Phase 2 will add AI cleanup
Archived: ~/.memos/archive/2025-10-31-143052.wav
Archived: ~/.memos/archive/2025-10-31-143052.txt
```

---

## Component Design

### 1. Audio Recording (`internal/cli/audio/`)

**Core Type:**
```go
type Recorder struct {
    sampleRate   uint32
    channels     uint32
    outputPath   string
    maxDuration  time.Duration
    maxBytes     int64
}

func NewRecorder(outputPath string, maxDuration time.Duration, maxBytes int64) *Recorder
func (r *Recorder) Start() error
func (r *Recorder) Stop() error
func (r *Recorder) WaitForStopCondition() error
```

**Implementation:**
- Use malgo for cross-platform audio capture
- Configure: 16kHz sample rate (Whisper native), mono, 16-bit PCM
- Monitor stdin for Enter key press
- Monitor elapsed time and buffer size against limits
- Stop on whichever happens first: Enter, maxDuration, or maxBytes
- Write WAV file to ~/.memos/recordings/{timestamp}.wav

**Safety Limits:**
- Default max duration: 1 hour
- Default max bytes: 256 MB
- Configurable via CLI flags
- Prevents accidental runaway recordings

**File Naming:**
- Format: `2025-10-31-143052.wav` (timestamp when recording starts)
- Auto-create directories if missing

---

### 2. Transcription (`internal/cli/transcription/`)

**Core Type:**
```go
type Client struct {
    apiKey     string
    httpClient *http.Client
}

func NewClient(apiKey string) *Client
func (c *Client) TranscribeFile(audioPath string) (string, error)
```

**Implementation:**
- POST to OpenAI Whisper API: `https://api.openai.com/v1/audio/transcriptions`
- Model: `whisper-1`
- Send audio file as multipart/form-data
- Response format: JSON with `text` field
- Write transcript to `.txt` file (same name as audio, different extension)

**API Key Handling:**
- Load from `--api-key` flag (highest priority)
- Fallback to `OPENAI_API_KEY` environment variable
- Kong handles this pattern automatically

**Error Handling:**
- Missing API key → clear error message
- API errors (401, 429, 500) → return status code and message
- Network timeout → 60s timeout for large files
- No retry logic (Phase 1 simplicity)

---

### 3. Content Generation (`internal/cli/content/`)

**Core Type:**
```go
type Generator struct {
    contentDir string  // e.g., "content/posts"
}

func NewGenerator(contentDir string) *Generator
func (g *Generator) GeneratePost(transcriptPath string, outputPath string) error
```

**⚠️ Phase 1: Raw Transcript Only**

Phase 2 will add Claude API integration here to:
- Clean up filler words and verbal tics
- Add proper punctuation and paragraph breaks
- Auto-generate meaningful title from content
- Suggest tags/categories

For now, we're proving the end-to-end workflow with raw transcripts.

**Markdown Generation:**

1. Read transcript text from `.txt` file
2. Generate timestamp-based title: `"Voice Memo 2025-10-31 14:30"`
3. Create minimal frontmatter:
   ```yaml
   ---
   title: "Voice Memo 2025-10-31 14:30"
   date: 2025-10-31T14:30:52-07:00
   draft: true
   ---
   ```
4. Body: Raw transcript text (no processing in Phase 1)
5. Write to `content/posts/{timestamp}.md`
6. Move audio and transcript to `~/.memos/archive/`

**Archive Behavior:**
- Only archive on successful markdown generation
- Move both `.wav` and `.txt` to archive
- Keeps files for debugging but organized
- User can bulk-delete old archives later

**Assumptions:**
- CLI runs from project root
- `content/posts/` directory exists
- User has write permissions

---

## Testing Strategy

### Unit Tests Only (Phase 1)

Using `testify/mock` (already in go.mod) for interface mocking.

#### `internal/cli/audio/` Tests
- Define `AudioDevice` interface wrapping malgo operations
- Mock device for: initialization, capture, stop
- Test WAV file generation with sample PCM data
- Test max duration/bytes limits trigger correctly
- Test file path generation and directory creation
- Test Enter key detection

#### `internal/cli/transcription/` Tests
- Mock HTTP client for API calls
- Test request formation (multipart, headers, auth)
- Test response parsing (JSON → string)
- Test error handling (401, 429, network errors)
- Test API key precedence (flag vs env var)

#### `internal/cli/content/` Tests
- Mock filesystem operations
- Test frontmatter generation with various timestamps
- Test markdown structure (YAML + body)
- Test file path generation
- Test archive move behavior
- Test error cases (missing dirs, permission denied)

### Test Coverage Goals
- Core logic: >80% coverage
- All major error paths tested
- Happy path testable unit-by-unit

### Manual Testing (Success Criteria)
```bash
# Full workflow test:
1. voice record → produces WAV file
2. voice transcribe {wav} → produces .txt with actual speech
3. voice process {txt} → produces content/posts/*.md with draft: true
4. hugo server → new draft appears in local site
5. Files moved to ~/.memos/archive/
```

---

## Error Handling

**Philosophy:** Simple error returns (Go-idiomatic). No fancy error types or rich terminal UI for Phase 1.

**Pattern:**
- Services return `error`
- Commands print to stderr
- Exit with non-zero status code on failure

**Example Error Messages:**
```
Error: no audio input device found
Error: API key required: set OPENAI_API_KEY or use --api-key
Error: content/posts/ not found - run from project root
Error: recording stopped: max duration (1h) reached
```

Clear, actionable error messages without requiring structured error types.

---

## Success Criteria

Phase 1 is complete when:

- [ ] All three commands (`record`, `transcribe`, `process`) work end-to-end
- [ ] Unit tests pass for all packages
- [ ] Can create a blog post from voice recording
- [ ] Hugo builds site with new draft post
- [ ] Files properly archived after processing
- [ ] `make lint` passes

---

## Known Limitations (Phase 1)

**Out of Scope:**
- Config files
- Cloud storage (Tigris)
- LLM cleanup (Claude API)
- Multiple transcription providers (Deepgram)
- Retry logic
- Progress indicators / spinners
- Automated workflow
- Integration tests
- CI/CD

These limitations are intentional. Phase 1 proves the core workflow. Phase 2+ will add intelligence and robustness.

---

## Future Enhancement Points

### Phase 2 Integration Points

1. **Content Generator (`internal/cli/content/`)**
   - Add Claude API call between transcript read and markdown write
   - Transform raw transcript into polished prose
   - Function signature stays the same, just internal processing changes

2. **Transcription Client (`internal/cli/transcription/`)**
   - Add provider interface to support Deepgram
   - Add retry logic with exponential backoff
   - Add progress indicators for long transcriptions

3. **Configuration**
   - Add `~/.memos/config` file support
   - Store API keys, preferences, defaults
   - Keep CLI flags for overrides

### Phase 3 Integration Points

1. **Workflow Command**
   - Add `voice workflow` to run all three steps
   - State persistence between steps
   - Better error recovery

2. **Preview Mode**
   - Preview generated markdown before writing
   - Interactive editing prompt

---

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| Kong over Cobra/urfave | Lightweight, tag-based, no IoC complexity |
| malgo for audio | Zero system dependencies, cross-platform |
| Whisper API (not Deepgram) | Lower cost for Phase 1 testing |
| `internal/cli/` namespace | Separates CLI from server code |
| Manual workflow | Prove pipeline before automation |
| Unit tests only | Fast feedback, sufficient for Phase 1 |
| testify/mock | Already in deps, zero new dependencies |
| Raw transcript | Phase 1 proves workflow, Phase 2 adds polish |
| `~/.memos/` storage | Standard CLI pattern, persistent across sessions |
| Archive moved files | Safety net without clutter |

---

## Questions Resolved

- ✅ Content location: `content/posts/` (memos as drafts)
- ✅ Title generation: Timestamp-based (auto)
- ✅ Audio storage: `~/.memos/` (user home directory)
- ✅ File cleanup: Move to archive (not delete)
- ✅ API key config: Flag with env fallback
- ✅ Stop recording: Press Enter key
- ✅ File reference: Explicit paths (no magic)
- ✅ Frontmatter: Minimal (title, date, draft)
- ✅ CLI name: `voice` (cmd/voice)
- ✅ Architecture: Service-oriented in `internal/cli/`
- ✅ Error handling: Simple error returns
- ✅ Testing: Unit tests with testify/mock
- ✅ Recording limits: 1h / 256MB (configurable)

---

## Next Steps

1. Set up git worktree for isolated development
2. Create detailed implementation plan (task breakdown)
3. Implement Phase 1 following TDD approach
4. Manual testing of full workflow
5. Update high-level plan with lessons learned
