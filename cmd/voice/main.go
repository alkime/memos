package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alkime/memos/internal/cli/audio"
	"github.com/alkime/memos/internal/cli/audio/device"
	"github.com/alkime/memos/internal/cli/content"
	"github.com/alkime/memos/internal/cli/transcription"
	"github.com/alkime/memos/internal/git"
)

// CLI defines the voice command structure.
type CLI struct {
	Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
	Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
	Process    ProcessCmd    `cmd:"" help:"Generate Hugo markdown from transcript"`
	Devices    DevicesCmd    `cmd:"" help:"List available audio devices"`
}

// RecordCmd handles audio recording.
type RecordCmd struct {
	Output      string `arg:"" optional:"" help:"Output file path"`
	Name        string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	MaxDuration string `flag:"" default:"1h" help:"Max recording duration"`
	MaxBytes    int64  `flag:"" default:"268435456" help:"Max file size (256MB)"`
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
		// Default to ~/.memos/work/{name}/recording.wav
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		workingName := getWorkingName(r.Name)
		outputPath = filepath.Join(homeDir, ".memos", "work", workingName, "recording.wav")
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
	audioFilePath := filepath.Join(homeDir, ".memos", "work", workingName, "recording.wav")

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

	return nil
}

// ProcessCmd handles markdown generation.
type ProcessCmd struct {
	TranscriptFile string `arg:"" help:"Path to transcript text file"`
	Output         string `flag:"" optional:"" help:"Output markdown path"`
}

// Run executes the process command.
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
	slog.Info("Processing transcript...")
	if err := generator.GeneratePost(p.TranscriptFile, outputPath); err != nil {
		return fmt.Errorf("failed to generate post: %w", err)
	}

	slog.Info("Generated post (draft)", "path", outputPath)
	slog.Info("Note: Raw transcript - Phase 2 will add AI cleanup")
	slog.Info("Archived: Files moved to ~/.memos/archive/")

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
