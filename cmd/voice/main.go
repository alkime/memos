package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alkime/memos/internal/cli/ai"
	"github.com/alkime/memos/internal/cli/audio"
	"github.com/alkime/memos/internal/cli/audio/device"
	"github.com/alkime/memos/internal/cli/transcription"
	"github.com/alkime/memos/internal/git"
)

// CLI defines the voice command structure.
type CLI struct {
	Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
	Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
	FirstDraft FirstDraftCmd `cmd:"" help:"Generate AI first draft from transcript"`
	CopyEdit   CopyEditCmd   `cmd:"" help:"Final copy-edit and save to content/posts"`
	Devices    DevicesCmd    `cmd:"" help:"List available audio devices"`
}

// RecordCmd handles audio recording.
type RecordCmd struct {
	Output       string `arg:"" optional:"" help:"Output file path"`
	Name         string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	MaxDuration  string `flag:"" default:"1h" help:"Max recording duration"`
	MaxBytes     int64  `flag:"" default:"268435456" help:"Max file size (256MB)"`
	NoTranscribe bool   `flag:"" help:"Skip automatic transcription after recording"`
	APIKey       string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key for transcription"`
}

// getWorkingName determines the working name for files.
// Priority: explicit name > git branch > timestamp fallback.
func getWorkingName(explicitName string) string {
	// Use explicit name if provided
	if explicitName != "" {
		return git.SanitizeBranchName(explicitName)
	}

	// Try to get git branch
	branch, err := git.GetCurrentBranch()
	if err == nil {
		return git.SanitizeBranchName(branch)
	}

	// Fallback to timestamp if not in git repo
	return time.Now().Format("2006-01-02-150405")
}

// Run executes the record command.
func (r *RecordCmd) Run() error {
	// Determine output path
	outputPath := r.Output
	if outputPath == "" {
		// Default to ~/.memos/work/{name}/recording.mp3
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		workingName := getWorkingName(r.Name)
		outputPath = filepath.Join(homeDir, ".memos", "work", workingName, "recording.mp3")
	}

	// Create parent directory if needed
	parentDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", parentDir, err)
	}

	// Parse max duration
	maxDuration, err := time.ParseDuration(r.MaxDuration)
	if err != nil {
		return fmt.Errorf("invalid max duration: %w", err)
	}

	// Create recorder
	recorder := audio.NewRecorder(audio.FileRecorderConfig{
		OutputPath:  outputPath,
		MaxDuration: maxDuration,
		MaxBytes:    r.MaxBytes,
	})

	err = recorder.Go(context.Background())
	if err != nil {
		return fmt.Errorf("failed to record audio: %w", err)
	}

	// Auto-transcribe unless --no-transcribe flag is set
	if r.NoTranscribe {
		return nil
	}

	// Skip transcription if no API key is provided
	if r.APIKey == "" {
		slog.Info("Skipping transcription (no API key provided)")
		return nil
	}

	// Delegate to transcribe command
	transcribeCmd := &TranscribeCmd{
		AudioFile: outputPath,
		APIKey:    r.APIKey,
		Output:    "", // Let it default to .txt file
		Name:      r.Name,
	}

	// If transcription fails, keep the recording
	if err := transcribeCmd.Run(); err != nil {
		slog.Error("Failed to transcribe recording", "error", err)
		return nil
	}

	return nil
}

// TranscribeCmd handles audio transcription.
type TranscribeCmd struct {
	AudioFile string `arg:"" optional:"" help:"Path to audio file (auto-detects if not provided)"`
	APIKey    string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key"`
	Output    string `flag:"" optional:"" help:"Output transcript path"`
	Name      string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
}

// autoDetectAudioFile determines the audio file path from working directory
// and prompts user for confirmation.
func autoDetectAudioFile(workingName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	audioFilePath := filepath.Join(homeDir, ".memos", "work", workingName, "recording.mp3")

	// Check if file exists
	if _, err := os.Stat(audioFilePath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf(
				"no recording found at %s - please record first or provide explicit path",
				audioFilePath,
			)
		}
		return "", fmt.Errorf("failed to check audio file: %w", err)
	}

	// Prompt user for confirmation
	//nolint:forbidigo // Interactive CLI confirmation requires user input
	fmt.Printf("Transcribe %s? [Y/n] ", audioFilePath)
	var response string
	if _, err := fmt.Scanln(&response); err != nil && err.Error() != "unexpected newline" {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	// Check response (default to yes if empty)
	if response != "" && response != "Y" && response != "y" && response != "yes" {
		return "", fmt.Errorf(
			"if %s is not the file to transcribe, please provide the correct one as an argument",
			audioFilePath,
		)
	}

	return audioFilePath, nil
}

// Run executes the transcribe command.
func (t *TranscribeCmd) Run() error {
	// Validate API key
	if t.APIKey == "" {
		return fmt.Errorf("API key required: set OPENAI_API_KEY or use --api-key")
	}

	// Determine audio file path
	audioFilePath := t.AudioFile
	if audioFilePath == "" {
		// Auto-detect audio file from working directory
		workingName := getWorkingName(t.Name)
		detectedPath, err := autoDetectAudioFile(workingName)
		if err != nil {
			return err
		}
		audioFilePath = detectedPath
	}

	// Determine output path
	outputPath := t.Output
	if outputPath == "" {
		// Default to same directory as audio file, .txt extension
		outputPath = audioFilePath[:len(audioFilePath)-len(filepath.Ext(audioFilePath))] + ".txt"
	}

	// Open audio file
	audioFile, err := os.Open(audioFilePath)
	if err != nil {
		return fmt.Errorf("failed to open audio file %s: %w", audioFilePath, err)
	}
	defer audioFile.Close()

	// Validate file is not empty
	info, err := audioFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat audio file: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("audio file is empty")
	}

	// Create transcription client
	client := transcription.NewClient(t.APIKey)

	// Transcribe
	slog.Info("Transcribing audio file...")
	text, err := client.TranscribeFile(audioFile)
	if err != nil {
		return fmt.Errorf("failed to transcribe audio file: %w", err)
	}

	// Write transcript
	//nolint:gosec // Transcript files need to be readable
	if err := os.WriteFile(outputPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write transcript to %s: %w", outputPath, err)
	}

	slog.Info("Transcript saved", "path", outputPath)

	// Print the transcript to screen
	//nolint:forbidigo // CLI output for transcript display
	fmt.Printf("\n--- Transcript ---\n\n")
	//nolint:forbidigo // CLI output for transcript display
	fmt.Println(text)
	//nolint:forbidigo // CLI output for transcript display
	fmt.Println("\n------------------")

	return nil
}

// FirstDraftCmd handles AI-powered first draft generation.
type FirstDraftCmd struct {
	TranscriptFile string `arg:"" optional:"" help:"Path to transcript file (auto-detects if not provided)"`
	APIKey         string `flag:"" env:"ANTHROPIC_API_KEY" help:"Anthropic API key"`
	Output         string `flag:"" optional:"" help:"Output markdown path"`
	Name           string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	NoEdit         bool   `flag:"" help:"Skip opening editor after generation"`
}

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

// CopyEditCmd handles AI-powered copy editing and final post generation.
type CopyEditCmd struct {
	FirstDraftFile string `arg:"" optional:"" help:"Path to first draft file (auto-detects if not provided)"`
	APIKey         string `flag:"" env:"ANTHROPIC_API_KEY" help:"Anthropic API key"`
	Output         string `flag:"" optional:"" help:"Output path (defaults to content/posts/)"`
	Name           string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
}

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

type DevicesCmd struct{}

func (dcmd *DevicesCmd) Run() error {
	slog.Info("Enumerating audio devices...")

	adev := device.NewAudioDevice(nil)
	devices, err := adev.EnumerateDevices(context.Background())
	if err != nil {
		return fmt.Errorf("failed to enumerate audio devices: %w", err)
	}

	for _, dev := range devices {
		slog.Info("Audio Device",
			"name", dev.Name,
			"isDefault", dev.IsDefault,
			"formatCount", dev.FormatCount,
			"formats", dev.Formats,
		)
	}

	return nil
}

func main() {
	// Set up text-based logger for CLI output
	//nolint:exhaustruct // Using default values for other HandlerOptions fields
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	cli := &CLI{} //nolint:exhaustruct // Kong fills in command fields
	ctx := kong.Parse(cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
	os.Exit(0)
}
