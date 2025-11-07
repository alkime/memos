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
