# Voice CLI AI Workflow Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor voice CLI to support AI-powered first-draft and copy-edit workflow using Anthropic API.

**Architecture:** Create new `internal/cli/ai` package for Anthropic API client, replace `ProcessCmd` with `FirstDraftCmd` and `CopyEditCmd`, add end-to-end workflow support to main voice command.

**Tech Stack:** Go, Anthropic SDK (github.com/anthropics/anthropic-sdk-go), Kong CLI parser

---

## Task 1: Add Anthropic SDK Dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add Anthropic SDK dependency**

Run:
```bash
go get github.com/anthropics/anthropic-sdk-go
```

Expected: Dependency added to go.mod

**Step 2: Tidy dependencies**

Run:
```bash
go mod tidy
```

Expected: go.sum updated

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add Anthropic SDK for AI content generation"
```

---

## Task 2: Create AI Package Structure

**Files:**
- Create: `internal/cli/ai/client.go`
- Create: `internal/cli/ai/prompts.go`

**Step 1: Create AI package directory**

Run:
```bash
mkdir -p internal/cli/ai
```

**Step 2: Create prompts.go with prompt templates**

Create `internal/cli/ai/prompts.go`:

```go
package ai

// FirstDraftSystemPrompt is the system prompt for generating first drafts.
const FirstDraftSystemPrompt = `You are a first draft writer. Given a raw voice memo transcription, you will:
- Lightly clean it up, removing verbal tics like "um", "and", "like", and similar filler words
- Reword things for clarity, but strive to keep the narrative voice as much as possible
- Organize the ideas, giving them section headings when appropriate, while maintaining the narrative voice
- Output clean markdown with appropriate heading levels (##, ###)
- Do NOT add Hugo frontmatter - just return the content body`

// CopyEditSystemPrompt is the system prompt for copy editing.
const CopyEditSystemPrompt = `You are a copy editor. Given a blog post draft, you will:
- Polish grammar, punctuation, and style consistency
- Fix any typos or awkward phrasing
- Ensure proper markdown formatting
- Generate appropriate Hugo frontmatter with title, date, and draft status
- The frontmatter must include: title, date (RFC3339 format), and draft: true
- Return the complete markdown file including frontmatter`
```

**Step 3: Create client.go with basic structure**

Create `internal/cli/ai/client.go`:

```go
package ai

import (
	"context"
	"errors"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client handles Anthropic API requests for content generation.
type Client struct {
	apiKey string
	model  string
}

// NewClient creates a new AI client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  "claude-sonnet-4-5-20250929",
	}
}

// GenerateFirstDraft creates a lightly edited first draft from raw transcript.
func (c *Client) GenerateFirstDraft(transcript string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("API key required: set ANTHROPIC_API_KEY or use --api-key")
	}

	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))

	params := anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(int64(4096)),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(FirstDraftSystemPrompt),
		}),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(transcript)),
		}),
	}

	ctx := context.Background()
	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to generate first draft via Anthropic API: %w", err)
	}

	// Extract text from response
	if len(resp.Content) == 0 {
		return "", errors.New("empty response from Anthropic API")
	}

	textBlock, ok := resp.Content[0].AsUnion().(anthropic.TextBlock)
	if !ok {
		return "", errors.New("unexpected response type from Anthropic API")
	}

	return textBlock.Text, nil
}

// GenerateCopyEdit performs final copy editing and returns markdown with frontmatter and extracted title.
func (c *Client) GenerateCopyEdit(firstDraft string) (markdown string, title string, error error) {
	if c.apiKey == "" {
		return "", "", errors.New("API key required: set ANTHROPIC_API_KEY or use --api-key")
	}

	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))

	params := anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(int64(4096)),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(CopyEditSystemPrompt),
		}),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(firstDraft)),
		}),
	}

	ctx := context.Background()
	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate copy edit via Anthropic API: %w", err)
	}

	// Extract text from response
	if len(resp.Content) == 0 {
		return "", "", errors.New("empty response from Anthropic API")
	}

	textBlock, ok := resp.Content[0].AsUnion().(anthropic.TextBlock)
	if !ok {
		return "", "", errors.New("unexpected response type from Anthropic API")
	}

	markdown = textBlock.Text

	// Extract title from frontmatter
	// Simple parsing: look for 'title: "..."' pattern
	title, err = extractTitleFromFrontmatter(markdown)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract title from frontmatter: %w", err)
	}

	return markdown, title, nil
}
```

**Step 4: Add helper function for title extraction**

Add to `internal/cli/ai/client.go`:

```go
import (
	"regexp"
	"strings"
)

// extractTitleFromFrontmatter parses the title from Hugo frontmatter.
func extractTitleFromFrontmatter(markdown string) (string, error) {
	// Match: title: "Some Title" or title: 'Some Title' or title: Some Title
	titleRegex := regexp.MustCompile(`(?m)^title:\s*["']?([^"'\n]+)["']?`)
	matches := titleRegex.FindStringSubmatch(markdown)
	if len(matches) < 2 {
		return "", errors.New("title not found in frontmatter")
	}

	title := strings.TrimSpace(matches[1])
	if title == "" {
		return "", errors.New("title is empty in frontmatter")
	}

	return title, nil
}
```

**Step 5: Verify code compiles**

Run:
```bash
go build ./internal/cli/ai
```

Expected: No errors

**Step 6: Commit**

```bash
git add internal/cli/ai/
git commit -m "feat(ai): add Anthropic API client for content generation

- Add first-draft generation with light cleanup
- Add copy-edit generation with frontmatter
- Extract title from frontmatter for file naming"
```

---

## Task 3: Add Slug Generation Helper

**Files:**
- Create: `internal/cli/ai/slug.go`

**Step 1: Create slug generation function**

Create `internal/cli/ai/slug.go`:

```go
package ai

import (
	"regexp"
	"strings"
)

// GenerateSlug converts a title to a URL-friendly slug.
// Example: "Voice CLI Improvements" -> "voice-cli-improvements"
func GenerateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters (keep alphanumeric and hyphens)
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	slug = reg.ReplaceAllString(slug, "")

	// Collapse multiple hyphens to single hyphen
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}
```

**Step 2: Create test file**

Create `internal/cli/ai/slug_test.go`:

```go
package ai

import "testing"

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "simple title",
			title:    "Voice CLI Improvements",
			expected: "voice-cli-improvements",
		},
		{
			name:     "title with special characters",
			title:    "AI-Powered Content: First Draft!",
			expected: "ai-powered-content-first-draft",
		},
		{
			name:     "title with multiple spaces",
			title:    "Multiple    Spaces   Here",
			expected: "multiple-spaces-here",
		},
		{
			name:     "title with leading/trailing spaces",
			title:    "  Trimmed Title  ",
			expected: "trimmed-title",
		},
		{
			name:     "title already lowercase",
			title:    "already-lowercase",
			expected: "already-lowercase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.title)
			if got != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		})
	}
}
```

**Step 3: Run test to verify it passes**

Run:
```bash
go test ./internal/cli/ai -v
```

Expected: PASS for all test cases

**Step 4: Commit**

```bash
git add internal/cli/ai/slug.go internal/cli/ai/slug_test.go
git commit -m "feat(ai): add slug generation for file naming"
```

---

## Task 4: Add FirstDraftCmd to main.go

**Files:**
- Modify: `cmd/voice/main.go`

**Step 1: Add FirstDraftCmd struct**

Add after `TranscribeCmd` in `cmd/voice/main.go`:

```go
// FirstDraftCmd handles AI-powered first draft generation.
type FirstDraftCmd struct {
	TranscriptFile string `arg:"" optional:"" help:"Path to transcript file (auto-detects if not provided)"`
	APIKey         string `flag:"" env:"ANTHROPIC_API_KEY" help:"Anthropic API key"`
	Output         string `flag:"" optional:"" help:"Output markdown path"`
	Name           string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	NoEdit         bool   `flag:"" help:"Skip opening editor after generation"`
}
```

**Step 2: Add import for ai package**

Add to imports in `cmd/voice/main.go`:

```go
"github.com/alkime/memos/internal/cli/ai"
"os/exec"
```

**Step 3: Implement FirstDraftCmd.Run()**

Add to `cmd/voice/main.go`:

```go
// Run executes the first-draft command.
func (f *FirstDraftCmd) Run() error {
	// Validate API key
	if f.APIKey == "" {
		return fmt.Errorf("API key required: set ANTHROPIC_API_KEY or use --api-key")
	}

	// Determine transcript file path
	transcriptPath := f.TranscriptFile
	if transcriptPath == "" {
		// Auto-detect transcript from working directory
		workingName := getWorkingName(f.Name)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		transcriptPath = filepath.Join(homeDir, ".memos", "work", workingName, "transcript.txt")

		// Check if file exists
		if _, err := os.Stat(transcriptPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf(
					"no transcript found at %s - please transcribe first or provide explicit path",
					transcriptPath,
				)
			}
			return fmt.Errorf("failed to check transcript file: %w", err)
		}

		// Prompt user for confirmation
		//nolint:forbidigo // Interactive CLI confirmation
		fmt.Printf("Generate first draft from %s? [Y/n] ", transcriptPath)
		var response string
		if _, err := fmt.Scanln(&response); err != nil && err.Error() != "unexpected newline" {
			return fmt.Errorf("failed to read user input: %w", err)
		}

		// Check response (default to yes if empty)
		if response != "" && response != "Y" && response != "y" && response != "yes" {
			return fmt.Errorf(
				"if %s is not the transcript to use, please provide the correct one as an argument",
				transcriptPath,
			)
		}
	}

	// Determine output path
	outputPath := f.Output
	if outputPath == "" {
		// Default to same directory as transcript, first-draft.md
		outputPath = filepath.Join(filepath.Dir(transcriptPath), "first-draft.md")
	}

	// Read transcript
	transcriptBytes, err := os.ReadFile(transcriptPath)
	if err != nil {
		return fmt.Errorf("failed to read transcript file %s: %w", transcriptPath, err)
	}
	transcript := string(transcriptBytes)

	// Create AI client
	client := ai.NewClient(f.APIKey)

	// Generate first draft
	slog.Info("Generating first draft with AI...")
	firstDraft, err := client.GenerateFirstDraft(transcript)
	if err != nil {
		// On API failure, save raw transcript as fallback
		slog.Error("Failed to generate first draft with AI", "error", err)
		slog.Info("Falling back to raw transcript")
		firstDraft = transcript
	}

	// Write first draft
	//nolint:gosec // Markdown files need to be readable
	if err := os.WriteFile(outputPath, []byte(firstDraft), 0644); err != nil {
		return fmt.Errorf("failed to write first draft to %s: %w", outputPath, err)
	}

	slog.Info("First draft saved", "path", outputPath)

	// Open in editor unless --no-edit flag is set
	if !f.NoEdit {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		slog.Info("Opening first draft in editor", "editor", editor)
		cmd := exec.Command(editor, outputPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			slog.Error("Failed to open editor", "error", err)
			slog.Info("You can manually edit the file", "path", outputPath)
		}
	}

	return nil
}
```

**Step 4: Add FirstDraft to CLI struct**

Modify the `CLI` struct in `cmd/voice/main.go`:

```go
type CLI struct {
	Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
	Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
	FirstDraft FirstDraftCmd `cmd:"" help:"Generate AI first draft from transcript"`
	Process    ProcessCmd    `cmd:"" help:"Generate Hugo markdown from transcript"`
	Devices    DevicesCmd    `cmd:"" help:"List available audio devices"`
}
```

**Step 5: Test compilation**

Run:
```bash
go build ./cmd/voice
```

Expected: No errors

**Step 6: Commit**

```bash
git add cmd/voice/main.go
git commit -m "feat(voice): add first-draft command with AI generation

- Auto-detect transcript from working directory
- Generate first draft using Anthropic API
- Fallback to raw transcript on API failure
- Open in editor by default (--no-edit to skip)"
```

---

## Task 5: Add CopyEditCmd to main.go

**Files:**
- Modify: `cmd/voice/main.go`

**Step 1: Add CopyEditCmd struct**

Add after `FirstDraftCmd` in `cmd/voice/main.go`:

```go
// CopyEditCmd handles AI-powered copy editing and final post generation.
type CopyEditCmd struct {
	FirstDraftFile string `arg:"" optional:"" help:"Path to first draft file (auto-detects if not provided)"`
	APIKey         string `flag:"" env:"ANTHROPIC_API_KEY" help:"Anthropic API key"`
	Output         string `flag:"" optional:"" help:"Output path (defaults to content/posts/)"`
	Name           string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
}
```

**Step 2: Add import for time package**

Already imported, verify `time` is in imports.

**Step 3: Implement CopyEditCmd.Run()**

Add to `cmd/voice/main.go`:

```go
// Run executes the copy-edit command.
func (c *CopyEditCmd) Run() error {
	// Validate API key
	if c.APIKey == "" {
		return fmt.Errorf("API key required: set ANTHROPIC_API_KEY or use --api-key")
	}

	// Determine first draft file path
	firstDraftPath := c.FirstDraftFile
	if firstDraftPath == "" {
		// Auto-detect first draft from working directory
		workingName := getWorkingName(c.Name)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		firstDraftPath = filepath.Join(homeDir, ".memos", "work", workingName, "first-draft.md")

		// Check if file exists
		if _, err := os.Stat(firstDraftPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf(
					"no first draft found at %s - please run 'voice first-draft' first",
					firstDraftPath,
				)
			}
			return fmt.Errorf("failed to check first draft file: %w", err)
		}
	}

	// Read first draft
	firstDraftBytes, err := os.ReadFile(firstDraftPath)
	if err != nil {
		return fmt.Errorf("failed to read first draft file %s: %w", firstDraftPath, err)
	}
	firstDraft := string(firstDraftBytes)

	// Create AI client
	client := ai.NewClient(c.APIKey)

	// Generate copy edit
	slog.Info("Generating copy edit with AI...")
	markdown, title, err := client.GenerateCopyEdit(firstDraft)
	if err != nil {
		return fmt.Errorf("failed to generate copy edit: %w", err)
	}

	// Determine output path
	outputPath := c.Output
	if outputPath == "" {
		// Generate filename from title
		slug := ai.GenerateSlug(title)
		now := time.Now()
		filename := fmt.Sprintf("%s-%s.md", now.Format("2006-01"), slug)

		// Check if file exists, add numeric suffix if needed
		outputPath = filepath.Join("content", "posts", filename)
		suffix := 2
		for {
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				break
			}
			// File exists, try with suffix
			filename = fmt.Sprintf("%s-%s-%d.md", now.Format("2006-01"), slug, suffix)
			outputPath = filepath.Join("content", "posts", filename)
			suffix++
		}
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", dir, err)
	}

	// Write final post
	//nolint:gosec // Markdown files need to be readable
	if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write final post to %s: %w", outputPath, err)
	}

	slog.Info("Final post saved", "path", outputPath, "title", title)

	return nil
}
```

**Step 4: Add CopyEdit to CLI struct**

Modify the `CLI` struct in `cmd/voice/main.go`:

```go
type CLI struct {
	Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
	Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
	FirstDraft FirstDraftCmd `cmd:"" help:"Generate AI first draft from transcript"`
	CopyEdit   CopyEditCmd   `cmd:"" help:"Final copy-edit and save to content/posts"`
	Process    ProcessCmd    `cmd:"" help:"Generate Hugo markdown from transcript"`
	Devices    DevicesCmd    `cmd:"" help:"List available audio devices"`
}
```

**Step 5: Test compilation**

Run:
```bash
go build ./cmd/voice
```

Expected: No errors

**Step 6: Commit**

```bash
git add cmd/voice/main.go
git commit -m "feat(voice): add copy-edit command with AI polishing

- Auto-detect first draft from working directory
- Generate final post with frontmatter
- Save to content/posts/ with date-slug filename
- Handle filename conflicts with numeric suffixes"
```

---

## Task 6: Remove ProcessCmd

**Files:**
- Modify: `cmd/voice/main.go`

**Step 1: Remove ProcessCmd from CLI struct**

Modify the `CLI` struct in `cmd/voice/main.go`:

```go
type CLI struct {
	Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
	Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
	FirstDraft FirstDraftCmd `cmd:"" help:"Generate AI first draft from transcript"`
	CopyEdit   CopyEditCmd   `cmd:"" help:"Final copy-edit and save to content/posts"`
	Devices    DevicesCmd    `cmd:"" help:"List available audio devices"`
}
```

**Step 2: Delete ProcessCmd struct and Run method**

Remove the `ProcessCmd` struct and its `Run()` method from `cmd/voice/main.go` (lines 240-269).

**Step 3: Test compilation**

Run:
```bash
go build ./cmd/voice
```

Expected: No errors

**Step 4: Commit**

```bash
git add cmd/voice/main.go
git commit -m "refactor(voice): remove process command

Replaced by first-draft and copy-edit workflow."
```

---

## Task 7: Add End-to-End Workflow Support

**Files:**
- Modify: `cmd/voice/main.go`

**Step 1: Add default command struct**

Add after `CLI` struct in `cmd/voice/main.go`:

```go
// DefaultCmd handles the end-to-end workflow when no subcommand is provided.
type DefaultCmd struct {
	Name         string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	MaxDuration  string `flag:"" default:"1h" help:"Max recording duration"`
	MaxBytes     int64  `flag:"" default:"268435456" help:"Max file size (256MB)"`
	OpenAIKey    string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key for transcription"`
	AnthropicKey string `flag:"" env:"ANTHROPIC_API_KEY" help:"Anthropic API key for AI generation"`
}
```

**Step 2: Implement DefaultCmd.Run()**

Add to `cmd/voice/main.go`:

```go
// Run executes the end-to-end workflow: record -> transcribe -> first-draft -> editor.
func (d *DefaultCmd) Run() error {
	// Validate API keys
	if d.OpenAIKey == "" {
		slog.Warn("OPENAI_API_KEY not set, transcription will be skipped")
	}
	if d.AnthropicKey == "" {
		slog.Warn("ANTHROPIC_API_KEY not set, first draft generation will be skipped")
	}

	// Step 1: Record
	slog.Info("Starting end-to-end workflow: Record -> Transcribe -> First Draft")
	recordCmd := &RecordCmd{
		Output:       "",
		Name:         d.Name,
		MaxDuration:  d.MaxDuration,
		MaxBytes:     d.MaxBytes,
		NoTranscribe: true, // We'll handle transcription manually
		APIKey:       d.OpenAIKey,
	}

	if err := recordCmd.Run(); err != nil {
		return fmt.Errorf("failed to record audio: %w", err)
	}

	// Skip transcription if no OpenAI key
	if d.OpenAIKey == "" {
		slog.Info("Skipping transcription (no OpenAI API key)")
		return nil
	}

	// Step 2: Transcribe
	transcribeCmd := &TranscribeCmd{
		AudioFile: "",
		APIKey:    d.OpenAIKey,
		Output:    "",
		Name:      d.Name,
	}

	if err := transcribeCmd.Run(); err != nil {
		return fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Skip first draft if no Anthropic key
	if d.AnthropicKey == "" {
		slog.Info("Skipping first draft generation (no Anthropic API key)")
		return nil
	}

	// Step 3: First Draft
	firstDraftCmd := &FirstDraftCmd{
		TranscriptFile: "",
		APIKey:         d.AnthropicKey,
		Output:         "",
		Name:           d.Name,
		NoEdit:         false, // Always open editor in end-to-end workflow
	}

	if err := firstDraftCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate first draft: %w", err)
	}

	slog.Info("Workflow complete. Review the first draft, then run 'voice copy-edit' when ready.")

	return nil
}
```

**Step 3: Modify CLI struct to support default command**

Replace the `CLI` struct in `cmd/voice/main.go`:

```go
type CLI struct {
	Default    DefaultCmd    `cmd:"" default:"1" help:"Run end-to-end workflow (record, transcribe, first-draft)"`
	Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
	Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
	FirstDraft FirstDraftCmd `cmd:"" help:"Generate AI first draft from transcript"`
	CopyEdit   CopyEditCmd   `cmd:"" help:"Final copy-edit and save to content/posts"`
	Devices    DevicesCmd    `cmd:"" help:"List available audio devices"`
}
```

**Step 4: Test compilation**

Run:
```bash
go build ./cmd/voice
```

Expected: No errors

**Step 5: Test help output**

Run:
```bash
./bin/voice --help
```

Expected: Shows all commands including default workflow

**Step 6: Commit**

```bash
git add cmd/voice/main.go
git commit -m "feat(voice): add end-to-end workflow as default command

Run 'voice' without subcommand to execute:
record -> transcribe -> first-draft -> editor"
```

---

## Task 8: Update Documentation

**Files:**
- Modify: `cmd/voice/README.md`
- Modify: `CLAUDE.md`

**Step 1: Update cmd/voice/README.md**

Replace content in `cmd/voice/README.md`:

```markdown
# Voice CLI

Voice-to-blog workflow with AI-powered content generation.

## Quick Start

### End-to-End Workflow

Run the complete workflow from recording to first draft:

\`\`\`bash
# Set API keys
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."

# Run complete workflow (record -> transcribe -> first-draft -> editor)
voice

# After reviewing and editing first-draft.md:
voice copy-edit
\`\`\`

### Step-by-Step Workflow

Or run individual commands:

\`\`\`bash
# 1. Record audio
voice record

# 2. Transcribe to text (auto-detects recording.mp3)
voice transcribe

# 3. Generate AI first draft (auto-detects transcript.txt)
voice first-draft

# ... edit first-draft.md manually ...

# 4. Generate final post (auto-detects first-draft.md)
voice copy-edit
\`\`\`

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

\`\`\`
~/.memos/work/{branch}/
├── recording.mp3      # Audio recording
├── transcript.txt     # Raw transcription
└── first-draft.md     # AI-generated first draft (edit this!)

content/posts/
└── {YYYY-MM}-{slug}.md   # Final published post
\`\`\`

## Configuration

### API Keys

Required for different steps:

- `OPENAI_API_KEY` - For transcription (Whisper)
- `ANTHROPIC_API_KEY` - For AI content generation (Claude)

Set via environment variables or `--api-key` flags.

### Editor

Set your preferred editor:

\`\`\`bash
export EDITOR=code  # VS Code
export EDITOR=vim   # Vim
export EDITOR=nano  # Nano
\`\`\`

Default: `vi`

## Examples

### Blog Post Workflow

\`\`\`bash
# Record voice memo
voice record

# Review transcript, then generate first draft
voice first-draft

# ... manually review and edit first-draft.md ...

# Generate final post
voice copy-edit

# Post is now in content/posts/ ready for Hugo
\`\`\`

### Quick Voice Note

\`\`\`bash
# Just record and transcribe (no AI)
voice record --no-transcribe
voice transcribe

# Transcript saved to ~/.memos/work/{branch}/transcript.txt
\`\`\`

### Custom Workflow

\`\`\`bash
# Record with custom name
voice record --name "my-idea"

# Transcribe specific file
voice transcribe ~/Downloads/interview.mp3

# Generate first draft from custom transcript
voice first-draft ~/Downloads/interview.txt

# Copy-edit specific first draft
voice copy-edit ~/Documents/edited-draft.md --output content/posts/2025-11-interview.md
\`\`\`

## Troubleshooting

### "API key required"

Make sure environment variables are set:

\`\`\`bash
echo $OPENAI_API_KEY
echo $ANTHROPIC_API_KEY
\`\`\`

Or pass explicitly:

\`\`\`bash
voice transcribe --api-key sk-...
voice first-draft --api-key sk-ant-...
\`\`\`

### "No recording found"

Check working directory:

\`\`\`bash
ls ~/.memos/work/$(git rev-parse --abbrev-ref HEAD)/
\`\`\`

Or provide explicit path:

\`\`\`bash
voice transcribe path/to/recording.mp3
\`\`\`

### Editor doesn't open

Set `EDITOR` environment variable:

\`\`\`bash
export EDITOR=vim
\`\`\`

Or skip editor:

\`\`\`bash
voice first-draft --no-edit
\`\`\`
\`\`\`

**Step 2: Update CLAUDE.md voice CLI section**

Find the "Voice Recording & Transcription" section in `CLAUDE.md` and update it:

```markdown
## Voice Recording & Transcription

The platform includes a voice CLI tool (`cmd/voice`) for converting voice recordings into blog posts with AI assistance.

**Complete Workflow:**

\`\`\`bash
# End-to-end: record -> transcribe -> first-draft -> editor
voice

# After editing first draft:
voice copy-edit
\`\`\`

**Audio Format:**
- **Container:** MP3 (encoded with shine-mp3, pure Go implementation)
- **Sample Rate:** 16,000 Hz (16 kHz) - Whisper's native sample rate
- **Channels:** Mono (1 channel)
- **Compression:** ~5-8x vs WAV (approximately 0.2-0.4 MB per minute)
- **OpenAI API Limit:** 25 MB per request (~1-2 hours of recording possible)

**AI Processing:**
1. **First Draft** (Anthropic Claude Sonnet 4.5):
   - Light cleanup: removes verbal tics (um, like, and)
   - Clarity: rewords for readability while preserving voice
   - Organization: adds section headings, maintains narrative flow
   - Output: Clean markdown (no frontmatter)

2. **Copy Edit** (Anthropic Claude Sonnet 4.5):
   - Polish: grammar, punctuation, style consistency
   - Formatting: proper markdown and Hugo frontmatter
   - Output: Final post in `content/posts/{YYYY-MM}-{slug}.md`

**File Flow:**
\`\`\`
~/.memos/work/{branch}/
├── recording.mp3       # Step 1: Record
├── transcript.txt      # Step 2: Transcribe (OpenAI Whisper)
└── first-draft.md      # Step 3: First Draft (Anthropic Claude)

content/posts/
└── {YYYY-MM}-{slug}.md # Step 4: Copy Edit (Anthropic Claude)
\`\`\`

**API Keys:**
- `OPENAI_API_KEY` - For transcription (Whisper)
- `ANTHROPIC_API_KEY` - For AI content generation (Claude)

**Individual Commands:**
- `voice record` - Record audio (auto-transcribes by default)
- `voice transcribe` - Transcribe audio to text
- `voice first-draft` - Generate AI first draft, open in `$EDITOR`
- `voice copy-edit` - Final polish and save to content/posts
- `voice devices` - List available audio devices

See `cmd/voice/README.md` for detailed usage.
\`\`\`

**Step 3: Commit documentation updates**

```bash
git add cmd/voice/README.md CLAUDE.md
git commit -m "docs: update voice CLI documentation for AI workflow

- Add end-to-end workflow examples
- Document first-draft and copy-edit commands
- Update file flow diagrams
- Add troubleshooting section"
```

---

## Task 9: Manual End-to-End Test

**Files:**
- None (testing only)

**Step 1: Build the binary**

Run:
```bash
make build-go
```

Expected: Binary created at `./bin/voice`

**Step 2: Set API keys**

Run:
```bash
export OPENAI_API_KEY="sk-..."  # Your OpenAI key
export ANTHROPIC_API_KEY="sk-ant-..."  # Your Anthropic key
```

**Step 3: Test end-to-end workflow**

Run:
```bash
./bin/voice
```

Expected:
1. Prompts to start recording
2. Records audio until Enter pressed
3. Transcribes with OpenAI Whisper
4. Generates first draft with Claude
5. Opens first-draft.md in editor

**Step 4: Edit first draft**

Manually edit the file, save and close editor.

**Step 5: Run copy-edit**

Run:
```bash
./bin/voice copy-edit
```

Expected:
1. Reads first-draft.md
2. Generates final post with Claude
3. Saves to `content/posts/{YYYY-MM}-{slug}.md`
4. Logs output path and title

**Step 6: Verify output file**

Run:
```bash
ls -la content/posts/
cat content/posts/2025-11-*.md  # Check frontmatter and content
```

Expected: File exists with proper frontmatter and polished content

**Step 7: Test individual commands**

Test each command separately:

```bash
# Test record only
./bin/voice record --no-transcribe

# Test transcribe only
./bin/voice transcribe

# Test first-draft only
./bin/voice first-draft

# Test copy-edit only
./bin/voice copy-edit
```

**Step 8: Document test results**

No commit needed, but note any issues for future fixes.

---

## Task 10: Optional Cleanup - Remove content package

**Files:**
- Delete: `internal/cli/content/` (optional)

**Note:** This task is optional. The `internal/cli/content/generator.go` package is no longer used, but can be kept for backwards compatibility or removed to clean up the codebase.

**Step 1: Check if content package is referenced anywhere**

Run:
```bash
grep -r "internal/cli/content" .
```

Expected: Only in git history, not in current code

**Step 2: Remove content package directory**

Run:
```bash
git rm -r internal/cli/content/
```

**Step 3: Verify build still works**

Run:
```bash
go build ./cmd/voice
```

Expected: No errors

**Step 4: Commit (if removed)**

```bash
git commit -m "refactor: remove unused content package

Generator functionality replaced by AI workflow."
```

---

## Testing Checklist

- [ ] `go build ./cmd/voice` compiles without errors
- [ ] `go test ./internal/cli/ai` passes all tests
- [ ] `voice --help` shows all commands
- [ ] `voice record` creates recording.mp3
- [ ] `voice transcribe` creates transcript.txt
- [ ] `voice first-draft` creates first-draft.md and opens editor
- [ ] `voice copy-edit` creates content/posts/{date}-{slug}.md
- [ ] End-to-end workflow runs: `voice` -> edit -> `voice copy-edit`
- [ ] File auto-detection works (branch-based working directory)
- [ ] API failures are handled gracefully with fallbacks
- [ ] Duplicate filenames get numeric suffixes
- [ ] Slug generation handles special characters
- [ ] Frontmatter is properly generated with title/date/draft

## Common Issues

**Import errors:**
- Run `go mod tidy` after adding Anthropic SDK

**Editor doesn't open:**
- Set `export EDITOR=vim` or similar
- Use `--no-edit` flag to skip

**API key errors:**
- Verify keys are set: `echo $ANTHROPIC_API_KEY`
- Pass explicitly with `--api-key` flag

**Title extraction fails:**
- Check AI response includes frontmatter with `title:` field
- Regex matches: `title: "..."`, `title: '...'`, or `title: ...`

## Next Steps

After implementation:

1. **Test with real recordings** - Try full workflow with actual voice memos
2. **Tune prompts** - Adjust `internal/cli/ai/prompts.go` based on output quality
3. **Add tests** - Write integration tests for command workflow
4. **Update README** - Add examples and troubleshooting to main README.md
5. **Consider archiving** - Add optional archiving of working files after copy-edit
