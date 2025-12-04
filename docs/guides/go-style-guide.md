# Go Style Guide

This is a living document that captures coding standards and best practices for the memos project. Guidelines are extracted from PR reviews and updated regularly.

## How to Update This Guide

A major source of updates for this document should come when addressing PR feedback.

When addressing such feedback, once all feedback has been addressed, the final step
in making updates based on the feedback--prior to merging--should be to understand
this document and if you should make any updates to it based on the comments.

This ensures the style guide stays current and captures learnings from code reviews as they happen.

## Core Guidelines

### 1. Always Wrap Errors with Context

**Rule:** Use `fmt.Errorf` with `%w` to add context when returning errors. Never silence the `wrapcheck` linter with `//nolint:wrapcheck`.

**Why:** Adding context to errors makes debugging easier by showing the call chain and providing relevant values (file paths, configuration, etc.). The `%w` verb preserves the error chain for `errors.Is()` and `errors.As()`.

**Don't:**
```go
homeDir, err := os.UserHomeDir()
if err != nil {
    return err //nolint:wrapcheck // Clear error context
}
```

**Do:**
```go
homeDir, err := os.UserHomeDir()
if err != nil {
    return fmt.Errorf("failed to get user home directory: %w", err)
}
```

**Additional Examples:**
```go
// Include relevant values in error messages
return fmt.Errorf("failed to create output directory %s: %w", dir, err)
return fmt.Errorf("failed to start server on port %s: %w", s.config.Port, err)
return fmt.Errorf("failed to read transcript file %s: %w", transcriptPath, err)
```

**Note:** Use `errors.New()` for sentinel errors that don't wrap other errors:
```go
if !r.state.isRecording {
    return errors.New("recorder not started")
}
```

---

### 2. Use Structured Logging (slog) Consistently

**Rule:** Use the `internal/logger` package (based on `log/slog`) for all logging. Avoid `fmt.Printf`, `fmt.Println`, or `log.Print` in production code.

**Why:** Structured logging provides consistent, parseable output with key-value pairs. The logger is environment-aware (debug in development, info in production) and outputs JSON for easy parsing.

**Don't:**
```go
fmt.Printf("Recording... Press Enter to stop. (Max: %s or %d MB)\n",
    maxDuration, maxBytes/(1024*1024)) //nolint:forbidigo // CLI output
```

**Do:**
```go
logger := slog.Default()
logger.Info("recording started",
    "maxDuration", maxDuration,
    "maxBytes", maxBytes/(1024*1024))
```

**Logger Setup:**
```go
// In main.go or initialization
import "github.com/alkime/memos/internal/logger"

// Initialize logger (reads LOG_LEVEL from environment)
if err := logger.Setup(); err != nil {
    return fmt.Errorf("failed to setup logger: %w", err)
}

// Use default logger throughout the application
logger := slog.Default()
logger.Info("server starting", "port", config.Port)
logger.Error("operation failed", "error", err)
logger.Debug("debugging info", "key", value)
```

**Common Patterns:**
```go
// Info level - normal operations
logger.Info("message", "key1", value1, "key2", value2)

// Error level - include the error
logger.Error("operation failed", "error", err, "context", additionalInfo)

// Debug level - verbose diagnostics
logger.Debug("detailed state", "key", value)
```

---

### 3. Prefer io.Reader/io.Writer Over File Paths for Decoupling

**Rule:** When a function only needs to read or write data, accept `io.Reader` or `io.Writer` interfaces instead of file paths. Reserve file path parameters for functions that need file system operations (creating directories, moving files, etc.).

**Why:** Using interfaces like `io.Reader` decouples the code from the file system, making it easier to test and reuse. The caller decides where data comes from (file, network, memory buffer, etc.).

**Don't:**
```go
// Function signature
func (c *Client) TranscribeFile(audioPath string) (string, error) {
    // Validate file exists
    info, err := os.Stat(audioPath)
    if err != nil {
        return "", err
    }

    // Open file
    file, err := os.Open(audioPath)
    // ...
}
```

**Do:**
```go
// Function signature
func (c *Client) TranscribeFile(audioFile io.Reader) (string, error) {
    // Work directly with the reader
    req := openai.AudioRequest{
        Model:  "whisper-1",
        Reader: audioFile,
    }
    // ...
}

// Caller handles file operations
func main() {
    file, err := os.Open(audioPath)
    if err != nil {
        return fmt.Errorf("failed to open audio file %s: %w", audioPath, err)
    }
    defer file.Close()

    transcript, err := client.TranscribeFile(file)
    // ...
}
```

**When to Use File Paths:**

Use file paths when the function needs file system operations:
```go
// This function creates directories and manages multiple files
func (g *Generator) GeneratePost(transcriptPath string, outputPath string) error {
    // Create output directory
    if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
        return fmt.Errorf("failed to create output directory: %w", err)
    }

    // Read transcript
    content, err := os.ReadFile(transcriptPath)
    // ...

    // Write output
    if err := os.WriteFile(outputPath, rendered, 0644); err != nil {
        return fmt.Errorf("failed to write output: %w", err)
    }

    return nil
}
```

**Benefits:**
- **Testing:** Easy to test with `bytes.Buffer` or `strings.Reader` without creating temp files
- **Flexibility:** Works with any data source (files, HTTP requests, in-memory data, stdin)
- **Composition:** Functions can be chained using `io.Pipe()` or similar patterns

---

### 4. Prefer SDK-Provided Types Over Primitives

**Rule:** When an SDK provides typed constants (enums, iotas, const groups), use those types instead of primitive types like `string` or `int`.

**Why:** SDK types provide compile-time type safety, prevent invalid values, and make code more maintainable when APIs evolve.

**Don't:**
```go
type Client struct {
    apiKey string
    model  string  // Can be any string, including invalid ones
}

func NewClient(apiKey string) *Client {
    return &Client{
        apiKey: apiKey,
        model:  "claude-sonnet-4-5-20250929",  // Typos won't be caught
    }
}
```

**Do:**
```go
type Client struct {
    apiKey string
    model  anthropic.Model  // Type-safe, validated by SDK
}

func NewClient(apiKey string) *Client {
    return &Client{
        apiKey: apiKey,
        model:  anthropic.ModelClaudeSonnet4_5_20250929,  // Compile-time checked
    }
}
```

**Benefits:**
- Compile-time validation of values
- IDE autocomplete for valid options
- Refactoring safety when SDK updates
- Self-documenting code

---

### 5. Use Explicit, Non-Ambiguous CLI Flag Names

**Rule:** CLI flags should have explicit, descriptive names that clearly indicate their purpose. Avoid generic names like `--api-key` when different API keys serve different purposes.

**Why:** Ambiguous flag names create confusion and collision problems when multiple similar resources are needed in the same command invocation.

**Don't:**
```go
type RecordCmd struct {
    APIKey string `flag:"" env:"OPENAI_API_KEY"`  // Ambiguous
}

type FirstDraftCmd struct {
    APIKey string `flag:"" env:"ANTHROPIC_API_KEY"`  // Same name, different meaning
}

// Problem: In RunCmd, both are needed but flags collide
```

**Do:**
```go
type RecordCmd struct {
    OpenAIAPIKey string `flag:"" env:"OPENAI_API_KEY"`  // Explicit
}

type FirstDraftCmd struct {
    AnthropicAPIKey string `flag:"" env:"ANTHROPIC_API_KEY"`  // Explicit
}

// Clear which key is for which service
```

**Common patterns:**
- Service-specific: `--openai-api-key`, `--anthropic-api-key`, `--github-token`
- Resource-specific: `--source-file`, `--dest-file` (not `--file1`, `--file2`)
- Context-specific: `--input-format`, `--output-format` (not `--format`)

---

### 6. Limit Function Return Parameters

**Rule:** When a function returns more than 2-3 values (excluding error), wrap them in a struct. This improves readability and makes the API easier to evolve.

**Why:** Multiple return values make function signatures harder to read, reduce type safety, and make it difficult to add new return values without breaking all callers.

**Don't:**
```go
// 4 return values is too many
func (c *Client) GenerateCopyEdit(
    firstDraft string,
    currentDate string,
) (markdown string, title string, changes []string, err error) {
    // ...
    return markdown, title, changes, nil
}

// Caller must handle all values in correct order
markdown, title, changes, err := client.GenerateCopyEdit(draft, date)
```

**Do:**
```go
// CopyEditResult wraps related return values
type CopyEditResult struct {
    Title    string
    Markdown string
    Changes  []string
}

func (c *Client) GenerateCopyEdit(
    firstDraft string,
    currentDate string,
) (*CopyEditResult, error) {
    // ...
    return &CopyEditResult{
        Title:    title,
        Markdown: markdown,
        Changes:  changes,
    }, nil
}

// Caller gets named fields and clear structure
result, err := client.GenerateCopyEdit(draft, date)
if err != nil {
    return err
}
fmt.Println(result.Title)
```

**Benefits:**
- Self-documenting: Field names make purpose clear
- Easier evolution: Add fields without breaking callers
- Better IDE support: Autocomplete shows field names
- Fewer ordering errors: Access by name instead of position

**Guidelines:**
- 2 values + error: Generally acceptable (e.g., `(string, error)`)
- 3 values + error: Consider a struct, especially if values are related
- 4+ values + error: Always use a struct

---

### 7. Verify Standard Library Best Practices for Current Go Version

**Rule:** Before applying or accepting advice about Go standard library usage, verify it's current for the project's Go version. The standard library evolves, and advice that was correct in older versions may be outdated or unnecessarily complex.

**Why:** Go improvements can make previously recommended workarounds obsolete. Following outdated advice adds unnecessary complexity and maintenance burden. This is especially important when receiving code review feedback from AI tools or older documentation.

**How to verify:**

1. **Check project Go version:**
   ```bash
   # In go.mod
   grep "^go " go.mod
   # Returns: go 1.23.0
   ```

2. **Check current standard library documentation:**
   ```bash
   # View official docs for current Go installation
   go doc time.After
   go doc sync.WaitGroup

   # Or visit: https://pkg.go.dev/time@go1.23.0
   ```

3. **Look for version-specific notes:**
   - Release notes often mention performance improvements
   - Package documentation includes "As of Go X.Y" notes
   - Search for phrases like "Before Go 1.X" or "As of Go 1.Y"

**Example 1: time.After() Evolution**

**Outdated advice (pre-Go 1.23):**
```go
// DON'T: Unnecessarily complex in Go 1.23+
timer := time.NewTimer(timeout)
defer timer.Stop()
select {
case <-timer.C:
    return ErrTimeout
}
```

**Current best practice (Go 1.23+):**
```go
// DO: Simple and correct in Go 1.23+
select {
case <-time.After(timeout):
    return ErrTimeout
}
```

**Why the change:** Go 1.23 improved the garbage collector to recover unreferenced `time.After()` timers automatically. The old "memory leak" concern no longer applies, making `time.After()` the preferred choice for its simplicity.

From Go 1.23 docs:
> "Before Go 1.23, this documentation warned that the underlying Timer would not be recovered by the garbage collector until the timer fired... As of Go 1.23, the garbage collector can recover unreferenced, unstopped timers. There is no reason to prefer NewTimer when After will do."

**Example 2: sync.WaitGroup.Go() Addition**

**Outdated advice (pre-Go 1.23):**
```go
// DON'T: Verbose pattern no longer necessary in Go 1.23+
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    doWork()
}()
wg.Wait()
```

**Current best practice (Go 1.23+):**
```go
// DO: Use WaitGroup.Go() for cleaner code
var wg sync.WaitGroup
wg.Go(doWork)
wg.Wait()
```

**Why the change:** Go 1.23 added the `WaitGroup.Go()` method, which handles `Add(1)`, launching the goroutine, and calling `Done()` automatically. This eliminates boilerplate and prevents forgetting to call `Done()`.

From Go 1.23 docs:
> "WaitGroup.Go calls Add(1), then runs fn in a new goroutine, and calls Done when fn returns."

**When reviewing code:**
- If advice seems to add complexity without clear benefit, check if it's version-specific
- AI code reviewers may give outdated advice based on older training data
- Cross-reference with current official Go documentation before accepting

---

## Additional Standards

### Linter Configuration

The project uses `golangci-lint` with extensive checks enabled. See `.golangci.yaml` for the complete configuration.

**Common Suppression Patterns:**

Suppression comments are allowed when intentional, but must include a reason:

```go
// Partial struct initialization (common in config, middleware)
return &Recorder{ //nolint:exhaustruct // state initialized on Start()
    sampleRate: 16000,
    channels:   1,
}

// Intentional security exception
os.MkdirAll(dir, 0755) //nolint:gosec // User directory, standard permissions
```

**Never suppress `wrapcheck`** - always wrap errors instead (see guideline #1).

### Running the Linter

```bash
# Check code quality before committing
make lint

# Or run directly
golangci-lint run
```

All code must pass `make lint` before merging.

---

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Project CLAUDE.md](../CLAUDE.md) - Architecture and development workflow
