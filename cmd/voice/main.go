package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alkime/memos/internal/audiofile"
	"github.com/alkime/memos/internal/cli/ai"
	"github.com/alkime/memos/internal/cli/audio"
	"github.com/alkime/memos/internal/cli/audio/device"
	"github.com/alkime/memos/internal/cli/editor"
	"github.com/alkime/memos/internal/cli/transcription"
	"github.com/alkime/memos/internal/git"
	"github.com/alkime/memos/internal/tui"
	"github.com/alkime/memos/internal/tui/phase/msg"
	tui_recording "github.com/alkime/memos/internal/tui/phase/recording"
	tuiPhases "github.com/alkime/memos/internal/tui/phases"
	"github.com/alkime/memos/internal/workdir"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/malgo"
)

// CLI defines the voice command structure.
type CLI struct {
	// Default workflow command (hidden from help, runs when no subcommand given)
	// Run RunCmd `cmd:"" default:"1" hidden:"" help:"Run end-to-end workflow"`
	Run RunCmd `cmd:"" help:"Run end-to-end workflow"`

	// Commands
	Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
	Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
	FirstDraft FirstDraftCmd `cmd:"" help:"Generate AI first draft from transcript"`
	CopyEdit   CopyEditCmd   `cmd:"" help:"Final copy-edit and save to content/posts"`
	Devices    DevicesCmd    `cmd:"" help:"List available audio devices"`
	Tui        TuiCmd        `cmd:"" help:"Launch terminal UI for recording"`
}

// RunCmd executes the end-to-end workflow: record -> transcribe -> first-draft -> editor.
type RunCmd struct {
	Mode string `flag:"" default:"memos" help:"Content mode: memos (full) or journal (minimal)"`
}

// Run executes the end-to-end workflow when no subcommand is provided: record -> transcribe -> first-draft -> editor.
func (r *RunCmd) Run() error {
	// Get API keys from environment
	openAIKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	// Validate API keys
	if openAIKey == "" {
		slog.Warn("OPENAI_API_KEY not set, transcription will be skipped")
	}
	if anthropicKey == "" {
		slog.Warn("ANTHROPIC_API_KEY not set, first draft generation will be skipped")
	}

	// Step 1: Record
	slog.Info("Starting end-to-end workflow: Record -> Transcribe -> First Draft")
	recordCmd := &RecordCmd{
		Output:       "",
		Name:         "", // Auto-detect from git branch
		MaxDuration:  "1h",
		MaxBytes:     268435456, // 256MB
		NoTranscribe: true,      // We'll handle transcription manually
		OpenAIAPIKey: openAIKey,
	}

	if err := recordCmd.Run(); err != nil {
		return fmt.Errorf("failed to record audio: %w", err)
	}

	// Skip transcription if no OpenAI key
	if openAIKey == "" {
		slog.Info("Skipping transcription (no OpenAI API key)")
		return nil
	}

	// Step 2: Transcribe
	transcribeCmd := &TranscribeCmd{
		AudioFile:    "",
		OpenAIAPIKey: openAIKey,
		Output:       "",
		Name:         "",   // Auto-detect from git branch
		SkipPrompt:   true, // Skip prompt in end-to-end workflow
	}

	if err := transcribeCmd.Run(); err != nil {
		return fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Skip first draft if no Anthropic key
	if anthropicKey == "" {
		slog.Info("Skipping first draft generation (no Anthropic API key)")
		return nil
	}

	// Parse and validate mode
	mode := ai.Mode(r.Mode)
	if mode != ai.ModeMemos && mode != ai.ModeJournal {
		return fmt.Errorf("invalid mode %q: must be 'memos' or 'journal'", r.Mode)
	}

	// Step 3: First Draft
	firstDraftCmd := &FirstDraftCmd{
		TranscriptFile:  "",
		AnthropicAPIKey: anthropicKey,
		Output:          "",
		Name:            "",     // Auto-detect from git branch
		Mode:            r.Mode, // Pass mode through
		NoEdit:          false,  // Always open editor in end-to-end workflow
	}

	if err := firstDraftCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate first draft: %w", err)
	}

	slog.Info("Workflow complete", "mode", mode,
		"next", "Review the first draft, then run 'voice copy-edit --mode="+string(mode)+"' when ready")

	return nil
}

// RecordCmd handles audio recording.
type RecordCmd struct {
	Output       string `arg:"" optional:"" help:"Output file path"`
	Name         string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	MaxDuration  string `flag:"" default:"1h" help:"Max recording duration"`
	MaxBytes     int64  `flag:"" default:"268435456" help:"Max file size (256MB)"`
	NoTranscribe bool   `flag:"" help:"Skip automatic transcription after recording"`
	OpenAIAPIKey string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key for transcription"`
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

func prepRecordingOutputPath(outputPath, nameParam string) (string, error) {
	if outputPath == "" {
		workingName := getWorkingName(nameParam)
		var err error
		outputPath, err = workdir.FilePath(workingName, "recording.mp3")
		if err != nil {
			return "", fmt.Errorf("failed to determine output path: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	return outputPath, nil
}

// Run executes the record command.
func (r *RecordCmd) Run() error {
	// Determine output path
	outputPath, err := prepRecordingOutputPath(r.Output, r.Name)
	if err != nil {
		return fmt.Errorf("failed to prepare output path: %w", err)
	}

	// Parse max duration
	maxDuration, err := time.ParseDuration(r.MaxDuration)
	if err != nil {
		return fmt.Errorf("invalid max duration: %w", err)
	}

	// Create recorder
	recorder, err := audio.NewRecorder(audio.FileRecorderConfig{
		OutputPath:        outputPath,
		MaxDuration:       maxDuration,
		MaxBytes:          r.MaxBytes,
		IgnoreStopSignals: false,
		DisplayProgress:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to create recorder: %w", err)
	}

	err = recorder.Go(context.Background(), nil)
	if err != nil {
		// Check for limit errors - these are not failures
		if errors.Is(err, audio.ErrMaxDurationReached) {
			slog.Info("recording stopped due to max duration limit")
			//nolint:forbidigo // CLI output for limit notification
			fmt.Printf("\nRecording stopped: Maximum duration (%s) reached\n", r.MaxDuration)
			//nolint:forbidigo // CLI output for limit notification
			fmt.Println("Audio file saved. Run 'voice transcribe' to continue manually.")
			return nil // Stop workflow, but exit successfully
		}
		if errors.Is(err, audio.ErrMaxBytesReached) {
			slog.Info("recording stopped due to max bytes limit")
			//nolint:forbidigo // CLI output for limit notification
			fmt.Printf("\nRecording stopped: Maximum size (%d MB) reached\n", r.MaxBytes/(1024*1024))
			//nolint:forbidigo // CLI output for limit notification
			fmt.Println("Audio file saved. Run 'voice transcribe' to continue manually.")
			return nil // Stop workflow, but exit successfully
		}

		// Actual error
		return fmt.Errorf("failed to record audio: %w", err)
	}

	// Auto-transcribe unless --no-transcribe flag is set
	if r.NoTranscribe {
		return nil
	}

	// Skip transcription if no API key is provided
	if r.OpenAIAPIKey == "" {
		slog.Info("Skipping transcription (no API key provided)")
		return nil
	}

	// Delegate to transcribe command
	transcribeCmd := &TranscribeCmd{
		AudioFile:    outputPath,
		OpenAIAPIKey: r.OpenAIAPIKey,
		Output:       "", // Let it default to transcript.txt in working directory
		Name:         r.Name,
		SkipPrompt:   true, // Skip prompt when auto-transcribing after recording
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
	AudioFile    string `arg:"" optional:"" help:"Path to audio file (auto-detects if not provided)"`
	OpenAIAPIKey string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key"`
	Output       string `flag:"" optional:"" help:"Output transcript path"`
	Name         string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	SkipPrompt   bool   `flag:"" help:"Skip confirmation prompt for auto-detected files"`
}

// autoDetectAudioFile determines the audio file path from working directory
// and optionally prompts user for confirmation.
func autoDetectAudioFile(workingName string, skipPrompt bool) (string, error) {
	audioFilePath, err := workdir.FilePath(workingName, "recording.mp3")
	if err != nil {
		return "", fmt.Errorf("failed to determine audio file path: %w", err)
	}

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

	// Skip prompt if requested (for end-to-end workflow)
	if skipPrompt {
		return audioFilePath, nil
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
	if t.OpenAIAPIKey == "" {
		return fmt.Errorf("API key required: set OPENAI_API_KEY or use --openai-api-key")
	}

	// Determine audio file path
	audioFilePath := t.AudioFile
	workingName := getWorkingName(t.Name)
	if audioFilePath == "" {
		// Auto-detect audio file from working directory
		detectedPath, err := autoDetectAudioFile(workingName, t.SkipPrompt)
		if err != nil {
			return err
		}
		audioFilePath = detectedPath
	}

	// Determine output path
	outputPath := t.Output
	if outputPath == "" {
		// Default to transcript.txt in working directory
		var err error
		outputPath, err = workdir.FilePath(workingName, "transcript.txt")
		if err != nil {
			return fmt.Errorf("failed to determine output path: %w", err)
		}
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
	client := transcription.NewClient(t.OpenAIAPIKey)

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
	TranscriptFile  string `arg:"" optional:"" help:"Path to transcript file (auto-detects if not provided)"`
	AnthropicAPIKey string `flag:"" env:"ANTHROPIC_API_KEY" help:"Anthropic API key"`
	Output          string `flag:"" optional:"" help:"Output markdown path"`
	Name            string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	Mode            string `flag:"" default:"memos" help:"Content mode: memos (full) or journal (minimal)"`
	NoEdit          bool   `flag:"" help:"Skip opening editor after generation"`
}

// Run executes the first-draft command.
//
//nolint:funlen // Function length justified by sequential steps in a CLI command.
func (f *FirstDraftCmd) Run() error {
	// Validate API key
	if f.AnthropicAPIKey == "" {
		return fmt.Errorf("API key required: set ANTHROPIC_API_KEY or use --anthropic-api-key")
	}

	// Determine transcript file path
	transcriptPath := f.TranscriptFile
	if transcriptPath == "" {
		// Auto-detect transcript from working directory
		workingName := getWorkingName(f.Name)
		var err error
		transcriptPath, err = workdir.FilePath(workingName, "transcript.txt")
		if err != nil {
			return fmt.Errorf("failed to determine transcript path: %w", err)
		}

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
	client := ai.NewClient(f.AnthropicAPIKey)

	// Parse and validate mode
	mode := ai.Mode(f.Mode)
	if mode != ai.ModeMemos && mode != ai.ModeJournal {
		return fmt.Errorf("invalid mode %q: must be 'memos' or 'journal'", f.Mode)
	}

	// Generate first draft
	slog.Info("Generating first draft with AI...", "mode", mode)
	firstDraft, err := client.GenerateFirstDraft(transcript, mode)
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
		// Ignore editor errors - user can manually edit if needed
		err = editor.Open(context.Background(), outputPath)
		if err != nil {
			slog.Warn("error calling editor.Open", "error", err)
		}
	}

	return nil
}

// CopyEditCmd handles AI-powered copy editing and final post generation.
type CopyEditCmd struct {
	FirstDraftFile  string `arg:"" optional:"" help:"Path to first draft file (auto-detects if not provided)"`
	AnthropicAPIKey string `flag:"" env:"ANTHROPIC_API_KEY" help:"Anthropic API key"`
	Output          string `flag:"" optional:"" help:"Output path (defaults to content/posts/)"`
	Name            string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	Mode            string `flag:"" default:"memos" help:"Content mode: memos (full) or journal (minimal)"`
}

// Run executes the copy-edit command.
//
//nolint:funlen // Function length justified by sequential steps in a CLI command.
func (c *CopyEditCmd) Run() error {
	// Validate API key
	if c.AnthropicAPIKey == "" {
		return fmt.Errorf("API key required: set ANTHROPIC_API_KEY or use --anthropic-api-key")
	}

	// Determine first draft file path
	firstDraftPath := c.FirstDraftFile
	if firstDraftPath == "" {
		// Auto-detect first draft from working directory
		workingName := getWorkingName(c.Name)
		var err error
		firstDraftPath, err = workdir.FilePath(workingName, "first-draft.md")
		if err != nil {
			return fmt.Errorf("failed to determine first draft path: %w", err)
		}

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
	client := ai.NewClient(c.AnthropicAPIKey)

	// Get current date for both AI prompt and filename
	now := time.Now()
	currentDate := now.Format(time.RFC3339)

	// Parse and validate mode
	mode := ai.Mode(c.Mode)
	if mode != ai.ModeMemos && mode != ai.ModeJournal {
		return fmt.Errorf("invalid mode %q: must be 'memos' or 'journal'", c.Mode)
	}

	// Generate copy edit
	slog.Info("Generating copy edit with AI...", "mode", mode)
	result, err := client.GenerateCopyEdit(firstDraft, currentDate, mode)
	if err != nil {
		return fmt.Errorf("failed to generate copy edit: %w", err)
	}

	// Save and display changes
	if len(result.Changes) > 0 {
		// Format changes for terminal display
		//nolint:forbidigo // CLI output for changes display
		fmt.Println("\n--- Changes Made ---")
		for _, change := range result.Changes {
			//nolint:forbidigo // CLI output for changes display
			fmt.Printf("  • %s\n", change)
		}
		//nolint:forbidigo // CLI output for changes display
		fmt.Println("--------------------")

		// Save changes to working directory
		workingName := getWorkingName(c.Name)
		changesPath, err := workdir.FilePath(workingName, "changes.txt")
		if err == nil {
			changesContent := "Changes made during copy-edit:\n\n"
			for _, change := range result.Changes {
				changesContent += fmt.Sprintf("• %s\n", change)
			}
			//nolint:gosec // Changes file needs to be readable
			if err := os.WriteFile(changesPath, []byte(changesContent), 0644); err != nil {
				slog.Warn("Failed to save changes file", "path", changesPath, "error", err)
			} else {
				slog.Info("Changes saved", "path", changesPath)
			}
		}
	}

	// Determine output path
	outputPath := c.Output
	if outputPath == "" {
		// Generate filename from title
		slug := ai.GenerateSlug(result.Title)
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
	if err := os.WriteFile(outputPath, []byte(result.Markdown), 0644); err != nil {
		return fmt.Errorf("failed to write final post to %s: %w", outputPath, err)
	}

	slog.Info("Final post saved", "path", outputPath, "title", result.Title)

	// Open in editor for review
	// Ignore editor errors - user can manually edit if needed
	err = editor.Open(context.Background(), outputPath)
	if err != nil {
		slog.Warn("error calling editor.Open", "error", err)
	}

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

type TuiCmd struct {
	Output          string `arg:"" optional:"" help:"Output file path"`
	Name            string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	MaxDuration     string `flag:"" default:"1h" help:"Max recording duration"`
	MaxBytes        int64  `flag:"" default:"268435456" help:"Max file size (256MB)"`
	Mode            string `flag:"" default:"memos" help:"Content mode: memos (full) or journal (minimal)"`
	OpenAIAPIKey    string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key for transcription"`
	AnthropicAPIKey string `flag:"" env:"ANTHROPIC_API_KEY" help:"Anthropic API key for first draft"`
}

//nolint:funlen // CLI command with multiple setup steps
func (tc *TuiCmd) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}

	// Parse and validate mode
	mode := ai.Mode(tc.Mode)
	if mode != ai.ModeMemos && mode != ai.ModeJournal {
		return fmt.Errorf("invalid mode %q: must be 'memos' or 'journal'", tc.Mode)
	}

	// Determine working paths
	workingName := getWorkingName(tc.Name)

	// input

	var (
		defaultSampleRate = 16_000
		defaultChannels   = 1
	)

	dataC := make(chan []byte, 64)

	dev := device.NewAudioDevice(&device.AudioDeviceConfig{
		Format:          malgo.FormatS16,
		SampleRate:      defaultSampleRate,
		CaptureChannels: defaultChannels,
	})

	err := dev.CaptureInto(ctx, dataC)
	if err != nil {
		return fmt.Errorf("failed to start audio capture: %w", err)
	}

	// always dealloc when we're done
	defer func() {
		dev.Dealloc(ctx)
		slog.Debug("Audio device deallocated")
	}()

	// Output paths

	outputPath, err := prepRecordingOutputPath(tc.Output, tc.Name)
	if err != nil {
		return fmt.Errorf("failed to prepare output path: %w", err)
	}

	transcriptPath, err := workdir.FilePath(workingName, "transcript.txt")
	if err != nil {
		return fmt.Errorf("failed to determine transcript path: %w", err)
	}

	draftPath, err := workdir.FilePath(workingName, "first-draft.md")
	if err != nil {
		return fmt.Errorf("failed to determine draft path: %w", err)
	}

	// Create audio file recorder
	recorder, err := audiofile.NewRecorder(audiofile.Config{
		SampleRate: defaultSampleRate,
		Channels:   defaultChannels,
		MP3Path:    outputPath,
	}, dataC)
	if err != nil {
		return fmt.Errorf("failed to create audio recorder: %w", err)
	}

	// Build TUI config
	config := tui.Config{
		Cancel:          cancel,
		AudioPath:       outputPath,
		TranscriptPath:  transcriptPath,
		DraftPath:       draftPath,
		WorkingName:     workingName,
		OpenAIAPIKey:    tc.OpenAIAPIKey,
		AnthropicAPIKey: tc.AnthropicAPIKey,
		Mode:            mode,
		MaxBytes:        tc.MaxBytes,
		EditorCmd:       os.Getenv("MEMOS_EDITOR"),
	}

	ctrls := makeRecordingControls2(ctx, dev, recorder, dataC, tc.MaxBytes)
	p := tea.NewProgram(tui.New2(config, ctrls))

	// Audio recorder goroutine (waits for channel close, MP3 conversion, cleanup)
	wg.Go(func() {
		if err := recorder.Start(ctx); err != nil {
			slog.Error("Audio recorder error", "error", err)
		}

		// Wait for recorder to finish (triggered by Finish() closing dataC)
		if err := recorder.Wait(); err != nil {
			slog.Error("Audio recorder error", "error", err)
		}

		p.Send(msg.AudioFinalizingCompleteMsg{AudioPath: outputPath})
	})

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	wg.Wait()

	//nolint:forbidigo // CLI output for completion notification
	fmt.Println("\nfinished. bye!")

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

//nolint:unused // Kept for old TUI model during migration
func makeRecordingControls(
	ctx context.Context,
	dev device.AudioDevice,
	recorder *audiofile.Recorder,
	maxBytes int64,
) tui_recording.Controls {
	return tui_recording.Controls{
		StartStopPause: audioDevKnob{
			ctx: ctx,
			dev: dev,
		},
		FileSize: audioFileDial{
			ctx:      ctx,
			recorder: recorder,
			maxBytes: maxBytes,
		},
	}
}

func makeRecordingControls2(
	ctx context.Context,
	dev device.AudioDevice,
	recorder *audiofile.Recorder,
	dataC chan []byte,
	maxBytes int64,
) tuiPhases.RecordingControls {
	return tuiPhases.RecordingControls{
		StartStopPause: audioDevKnob{
			ctx: ctx,
			dev: dev,
		},
		FileSize: audioFileDial{
			ctx:      ctx,
			recorder: recorder,
			maxBytes: maxBytes,
		},
		Finish: func() {
			if err := dev.Stop(ctx); err != nil {
				slog.Error("Failed to stop audio device", "error", err)
			}
			close(dataC)
		},
	}
}

type audioDevKnob struct {
	ctx context.Context
	dev device.AudioDevice
}

func (adk audioDevKnob) Read() bool {
	return adk.dev.IsStarted()
}

func (adk audioDevKnob) On() {
	err := adk.dev.Start(adk.ctx)
	if err != nil {
		slog.Error("audioDevKnob On error", "error", err)
	}
}

func (adk audioDevKnob) Off() {
	err := adk.dev.Stop(adk.ctx)
	if err != nil {
		slog.Error("audioDevKnob Off error", "error", err)
	}
}

func (adk audioDevKnob) Toggle() {
	err := adk.dev.Toggle(adk.ctx)
	if err != nil {
		slog.Error("audioDevKnob Toggle error", "error", err)
	}
}

type audioFileDial struct {
	ctx      context.Context
	recorder *audiofile.Recorder
	maxBytes int64
}

func (afd audioFileDial) Read() int64 {
	return afd.recorder.BytesWritten()
}

func (afd audioFileDial) Cap() (int64, int64) {
	return afd.Read(), afd.maxBytes
}
