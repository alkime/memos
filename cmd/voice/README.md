# Voice CLI

Voice-to-blog workflow with AI-powered content generation.

## Quick Start

### End-to-End Workflow

Run the complete workflow from recording to first draft:

```bash
# Set API keys
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."

# Run complete workflow (record -> transcribe -> first-draft -> editor)
voice

# After reviewing and editing first-draft.md:
voice copy-edit
```

### Step-by-Step Workflow

Or run individual commands:

```bash
# 1. Record audio
voice record

# 2. Transcribe to text (auto-detects recording.mp3)
voice transcribe

# 3. Generate AI first draft (auto-detects transcript.txt)
voice first-draft

# ... edit first-draft.md manually ...

# 4. Generate final post (auto-detects first-draft.md)
voice copy-edit
```

## Commands

### `voice` (default)

Run end-to-end workflow: record → transcribe → first-draft → editor

Requires both `OPENAI_API_KEY` and `ANTHROPIC_API_KEY`.

### `voice record`

Record audio from microphone.

- Output: `~/.memos/work/{branch}/recording.mp3`
- Auto-transcribes by default (use `--no-transcribe` to skip)
- Detects git branch for working directory name

Options:
- `--name` - Override working directory name
- `--max-duration` - Max recording length (default: 1h)
- `--max-bytes` - Max file size (default: 256MB)
- `--no-transcribe` - Skip automatic transcription

### `voice transcribe [audio-file]`

Transcribe audio to text using OpenAI Whisper.

- Input: Auto-detects `recording.mp3` or provide explicit path
- Output: `transcript.txt` in same directory

Requires `OPENAI_API_KEY`.

### `voice first-draft [transcript-file]`

Generate AI first draft from transcript.

- Input: Auto-detects `transcript.txt` or provide explicit path
- Output: `first-draft.md` in same directory
- Opens in `$EDITOR` by default (use `--no-edit` to skip)

Requires `ANTHROPIC_API_KEY`.

**What it does:**
- Removes verbal tics (um, like, and)
- Rewords for clarity while preserving your voice
- Organizes ideas with section headings
- Outputs clean markdown

### `voice copy-edit [first-draft-file]`

Final copy-edit and publish to content/posts.

- Input: Auto-detects `first-draft.md` or provide explicit path
- Output: `content/posts/{YYYY-MM}-{slug}.md`

Requires `ANTHROPIC_API_KEY`.

**What it does:**
- Polishes grammar and style
- Fixes typos and awkward phrasing
- Generates Hugo frontmatter
- Saves with date-slug filename

### `voice devices`

List available audio input devices.

## File Structure

```
~/.memos/work/{branch}/
├── recording.mp3      # Audio recording
├── transcript.txt     # Raw transcription
└── first-draft.md     # AI-generated first draft (edit this!)

content/posts/
└── {YYYY-MM}-{slug}.md   # Final published post
```

## Configuration

### API Keys

Required for different steps:

- `OPENAI_API_KEY` - For transcription (Whisper)
- `ANTHROPIC_API_KEY` - For AI content generation (Claude)

Set via environment variables or explicit flags:
- `--openai-api-key` for transcription commands
- `--anthropic-api-key` for AI generation commands

### Editor

Set your preferred editor:

```bash
export EDITOR=code  # VS Code
export EDITOR=vim   # Vim
export EDITOR=nano  # Nano
```

Default: `vi`

## Examples

### Blog Post Workflow

```bash
# Record voice memo
voice record

# Review transcript, then generate first draft
voice first-draft

# ... manually review and edit first-draft.md ...

# Generate final post
voice copy-edit

# Post is now in content/posts/ ready for Hugo
```

### Quick Voice Note

```bash
# Just record and transcribe (no AI)
voice record --no-transcribe
voice transcribe

# Transcript saved to ~/.memos/work/{branch}/transcript.txt
```

### Custom Workflow

```bash
# Record with custom name
voice record --name "my-idea"

# Transcribe specific file
voice transcribe ~/Downloads/interview.mp3

# Generate first draft from custom transcript
voice first-draft ~/Downloads/interview.txt

# Copy-edit specific first draft
voice copy-edit ~/Documents/edited-draft.md --output content/posts/2025-11-interview.md
```

## Troubleshooting

### "API key required"

Make sure environment variables are set:

```bash
echo $OPENAI_API_KEY
echo $ANTHROPIC_API_KEY
```

Or pass explicitly:

```bash
voice transcribe --openai-api-key sk-...
voice first-draft --anthropic-api-key sk-ant-...
```

### "No recording found"

Check working directory:

```bash
ls ~/.memos/work/$(git rev-parse --abbrev-ref HEAD)/
```

Or provide explicit path:

```bash
voice transcribe path/to/recording.mp3
```

### Editor doesn't open

Set `EDITOR` environment variable:

```bash
export EDITOR=vim
```

Or skip editor:

```bash
voice first-draft --no-edit
```
