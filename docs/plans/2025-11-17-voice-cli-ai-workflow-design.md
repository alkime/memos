# Voice CLI AI Workflow Design

**Date:** 2025-11-17
**Status:** Approved
**Replaces:** Process command (placeholder implementation)

## Overview

Refactor the voice CLI to incorporate AI-powered content generation using the Anthropic API. The new workflow splits content creation into two AI-assisted steps (first-draft and copy-edit) with manual review in between, and provides a unified command for the complete end-to-end workflow.

## Problem Statement

The current `voice process` command is a placeholder that simply wraps raw transcripts in Hugo frontmatter. The actual workflow requires:

1. AI-powered first draft generation (light cleanup, organization, preserving voice)
2. Manual review and editing by the user
3. AI-powered copy-editing (grammar, style, final polish)
4. Proper file naming and Hugo frontmatter generation

## Design Goals

1. **Preserve narrative voice** - First draft should clean up verbal tics while maintaining the author's voice
2. **Manual control** - User reviews/edits between AI steps, no fully automated publication
3. **Workflow flexibility** - Support both complete end-to-end workflow and individual steps
4. **Consistent patterns** - Follow existing codebase patterns (similar to transcription package)
5. **Clear file organization** - Logical progression from recording → transcript → first-draft → final post

## Architecture

### Package Structure

**New Package: `internal/cli/ai`**

```
internal/cli/ai/
├── client.go       # Anthropic API client
└── prompts.go      # Prompt templates
```

Follows the same pattern as `internal/cli/transcription` - encapsulates API interaction, provides clean interface to commands.

**Modified Files:**

- `cmd/voice/main.go` - Replace `ProcessCmd` with `FirstDraftCmd` and `CopyEditCmd`, add end-to-end workflow
- `go.mod` - Add `github.com/anthropics/anthropic-sdk-go`

**Removed:**

- `ProcessCmd` in `cmd/voice/main.go`
- Potentially simplify or remove `internal/cli/content/generator.go` (currently just wraps in frontmatter)

### Command Structure

```go
type CLI struct {
    Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
    Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
    FirstDraft FirstDraftCmd `cmd:"" help:"Generate AI first draft from transcript"`
    CopyEdit   CopyEditCmd   `cmd:"" help:"Final copy-edit and save to content/posts"`
    Devices    DevicesCmd    `cmd:"" help:"List available audio devices"`
}
```

**No subcommand (default workflow):**

When `voice` is run without arguments, it executes:

1. Record audio → `~/.memos/work/{branch}/recording.mp3`
2. Auto-transcribe → `~/.memos/work/{branch}/transcript.txt`
3. Generate first draft → `~/.memos/work/{branch}/first-draft.md`
4. Open `first-draft.md` in `$EDITOR` (blocking)
5. Exit (user reviews, then manually runs `voice copy-edit`)

**Individual Subcommands:**

- `voice record [--no-transcribe]` - Record only, optionally skip auto-transcribe
- `voice transcribe [audio-file]` - Transcribe only, auto-detects file if not provided
- `voice first-draft [transcript-file]` - Generate first draft, auto-detects transcript, opens editor
- `voice copy-edit` - Auto-detects first-draft, generates final post in `content/posts/`

### File Flow

```
~/.memos/work/{branch}/
├── recording.mp3          # Step 1: Record
├── transcript.txt         # Step 2: Transcribe
└── first-draft.md         # Step 3: First Draft (user edits this)

content/posts/
└── {YYYY-MM}-{slug}.md    # Step 4: Copy Edit (final output)
```

## AI Integration

### Anthropic API Client

```go
// internal/cli/ai/client.go
package ai

type Client struct {
    apiKey string
    model  string  // Default: "claude-sonnet-4-5-20250929"
}

func NewClient(apiKey string) *Client

func (c *Client) GenerateFirstDraft(transcript string) (string, error)

func (c *Client) GenerateCopyEdit(firstDraft string) (markdown string, title string, error)
```

### First Draft Prompt

**System Prompt:**
```
You are a first draft writer. Given a raw voice memo transcription, you will:
- Lightly clean it up, removing verbal tics like "um", "and", "like", and similar filler words
- Reword things for clarity, but strive to keep the narrative voice as much as possible
- Organize the ideas, giving them section headings when appropriate, while maintaining the narrative voice
- Output clean markdown with appropriate heading levels (##, ###)
- Do NOT add Hugo frontmatter - just return the content body
```

**Input:** Raw transcript text
**Output:** Cleaned markdown content (no frontmatter)

### Copy Edit Prompt

**System Prompt:**
```
You are a copy editor. Given a blog post draft, you will:
- Polish grammar, punctuation, and style consistency
- Fix any typos or awkward phrasing
- Ensure proper markdown formatting
- Generate appropriate Hugo frontmatter with title, date, and draft status
- Return the complete markdown file including frontmatter
```

**Input:** First draft markdown (user-edited)
**Output:** Complete markdown file with frontmatter

**Frontmatter Format:**
```yaml
---
title: "Extracted Title"
date: 2025-11-17T14:30:00-08:00
draft: true
---
```

### Environment Variables

- `ANTHROPIC_API_KEY` - Required for AI commands
- Can also use `--api-key` flag (consistent with existing OpenAI integration)

## Data Flow

### End-to-End Workflow (No Subcommand)

```
User runs: voice

1. RecordCmd.Run()
   ├─ Create ~/.memos/work/{branch}/recording.mp3
   └─ Auto-call TranscribeCmd.Run() (unless --no-transcribe)

2. TranscribeCmd.Run()
   ├─ Call OpenAI Whisper API
   └─ Create ~/.memos/work/{branch}/transcript.txt

3. FirstDraftCmd.Run()
   ├─ Read transcript.txt
   ├─ Call Anthropic API with first-draft prompt
   ├─ Write ~/.memos/work/{branch}/first-draft.md
   └─ Open first-draft.md in $EDITOR (blocking)

4. Exit
   User reviews first-draft.md, makes edits

5. User runs: voice copy-edit

6. CopyEditCmd.Run()
   ├─ Auto-detect ~/.memos/work/{branch}/first-draft.md
   ├─ Call Anthropic API with copy-edit prompt
   ├─ Parse frontmatter to extract title
   ├─ Generate slug from title
   └─ Write content/posts/{YYYY-MM}-{slug}.md
```

### Individual Command Workflow

Each command can be run independently:

```bash
# Step-by-step workflow
voice record                    # Creates recording.mp3 + transcript.txt
voice first-draft              # Creates first-draft.md, opens editor
# ... user edits first-draft.md ...
voice copy-edit                # Creates content/posts/{date}-{slug}.md

# Or skip recording if you have an audio file
voice transcribe my-recording.mp3
voice first-draft
voice copy-edit

# Or just transcribe and manually write
voice transcribe
# ... manually write post ...
```

## Error Handling

### API Failures

**First Draft API Failure:**
- Save raw transcript to `first-draft.md` as fallback
- Log error with API key troubleshooting suggestion
- Still open editor (user can manually edit transcript)
- Non-fatal - degraded but usable workflow

**Copy Edit API Failure:**
- Don't overwrite any files
- Return clear error message
- User can retry or manually move first-draft to content/posts
- Fatal error - must succeed to proceed

### Missing Files

**Auto-detection for first-draft:**
- Look for `~/.memos/work/{branch}/transcript.txt`
- Prompt user for confirmation (like transcribe command)
- Clear error if file doesn't exist

**Auto-detection for copy-edit:**
- Look for `~/.memos/work/{branch}/first-draft.md`
- Clear error if file doesn't exist
- Suggest running `voice first-draft` first

### Editor Handling

- Use `$EDITOR` environment variable
- Fallback to `vi` if `$EDITOR` not set
- Block until editor process exits (standard behavior)
- If editor command fails, log error but don't fail operation
- No validation of user changes - user has full control

### File Conflicts

**Duplicate filenames in content/posts/:**
- If `{YYYY-MM}-{slug}.md` exists, append numeric suffix
- Try `-2.md`, `-3.md`, etc. until unique filename found
- Log warning about duplicate filename
- Continue operation with new filename

### API Key Validation

- Check for `ANTHROPIC_API_KEY` before making API calls
- Clear error message if missing
- Consistent with existing `OPENAI_API_KEY` pattern

## File Naming Convention

**Working Directory Files:**
- `recording.mp3` - Fixed name, overwritten each time
- `transcript.txt` - Fixed name, overwritten each time
- `first-draft.md` - Fixed name, user edits this file

**Final Output:**
- Format: `content/posts/{YYYY-MM}-{slug}.md`
- `{YYYY-MM}` - Current year and month (e.g., `2025-11`)
- `{slug}` - Lowercase, hyphenated version of title from frontmatter
- Example: `content/posts/2025-11-voice-workflow-improvements.md`

**Slug Generation:**
1. Extract title from frontmatter
2. Convert to lowercase
3. Replace spaces with hyphens
4. Remove special characters (keep alphanumeric and hyphens)
5. Collapse multiple hyphens to single hyphen

## Migration Strategy

**Breaking Changes:**
- Remove `voice process` command entirely
- Update documentation and README

**Backwards Compatibility:**
- Existing `voice record` and `voice transcribe` commands unchanged
- Working directory structure (`~/.memos/work/{branch}/`) unchanged
- File formats (MP3, TXT) unchanged

**Documentation Updates:**
- Update `cmd/voice/README.md` with new workflow
- Update main `CLAUDE.md` with new command structure
- Add examples for both end-to-end and step-by-step workflows

## Future Enhancements

**Not in Scope (YAGNI):**

- Archive working files after copy-edit (can add later if needed)
- Custom prompts via config file (use hardcoded prompts initially)
- Multiple draft iterations (user can manually re-run commands)
- Template selection for different post types (single template for now)
- Automatic tag/category generation (manual in frontmatter for now)

**Potential Future Work:**

- Add `--model` flag to select different Claude models
- Add `--no-edit` flag to skip editor opening
- Support custom prompt templates in `~/.memos/prompts/`
- Integrate with Hugo's `draft: false` workflow for publishing

## Dependencies

**New:**
- `github.com/anthropics/anthropic-sdk-go` - Anthropic API client

**Existing:**
- `github.com/openai/openai-go` - OpenAI Whisper (unchanged)
- `github.com/alecthomas/kong` - CLI parsing (unchanged)
- All existing audio recording dependencies (unchanged)

## Testing Strategy

**Unit Tests:**
- `internal/cli/ai/client_test.go` - Mock API responses, test prompt formatting
- `internal/cli/ai/prompts_test.go` - Validate prompt templates

**Integration Tests:**
- Test file auto-detection logic
- Test editor opening (mock `$EDITOR`)
- Test slug generation from various titles
- Test file conflict resolution

**Manual Testing:**
- End-to-end workflow with real API calls
- Test with missing API keys
- Test with malformed files
- Test editor workflow interruption (Ctrl-C)

## Security Considerations

- API keys read from environment variables (secure, not in code)
- File permissions: `0644` for generated markdown (readable)
- Directory permissions: `0755` for `~/.memos/` (standard)
- No user input passed directly to shell commands
- AI-generated content written to predictable locations (no path injection)

## Success Criteria

1. `voice` command runs complete workflow: record → transcribe → first-draft → editor
2. `voice copy-edit` generates properly formatted Hugo posts in `content/posts/`
3. First draft preserves narrative voice while cleaning up verbal tics
4. Copy edit polishes grammar/style and generates correct frontmatter
5. All commands handle missing files and API failures gracefully
6. File naming follows `{YYYY-MM}-{slug}.md` convention
7. Existing `record` and `transcribe` commands remain unchanged
