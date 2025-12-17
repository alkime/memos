package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/alkime/memos/internal/audiofile"
	"github.com/alkime/memos/internal/cli/ai"
	"github.com/alkime/memos/internal/cli/audio/device"
	"github.com/alkime/memos/internal/git"
	"github.com/alkime/memos/internal/keyring"
	"github.com/alkime/memos/internal/tui"
	"github.com/alkime/memos/internal/tui/workflow"
	"github.com/alkime/memos/internal/workdir"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/malgo"
)

// CLI defines the voice command structure.
type CLI struct {
	// Default TUI command (runs when no subcommand given)
	TUI TUICmd `cmd:"" default:"withargs" help:"Launch terminal UI for recording workflow"`

	// Subcommands
	Devices DevicesCmd `cmd:"" help:"List available audio devices"`
	Config  ConfigCmd  `cmd:"" help:"Manage configuration"`
}

// TUICmd is the default command that runs the TUI.
type TUICmd struct {
	Output          string `arg:"" optional:"" help:"Output file path"`
	Name            string `flag:"" optional:"" help:"Working name (overrides git branch detection)"`
	MaxDuration     string `flag:"" default:"1h" help:"Max recording duration"`
	MaxBytes        int64  `flag:"" default:"268435456" help:"Max file size (256MB)"`
	Mode            string `flag:"" default:"memos" help:"Content mode: memos (full) or journal (minimal)"`
	OpenAIAPIKey    string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key for transcription"`
	AnthropicAPIKey string `flag:"" env:"ANTHROPIC_API_KEY" help:"Anthropic API key for first draft"`
}

// Run executes the TUI command.
//
//nolint:funlen // CLI command with multiple setup steps
func (c *TUICmd) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}

	// Parse and validate mode
	mode := ai.Mode(c.Mode)
	if mode != ai.ModeMemos && mode != ai.ModeJournal {
		return fmt.Errorf("invalid mode %q: must be 'memos' or 'journal'", c.Mode)
	}

	// Resolve API keys: environment variables take priority, fallback to keychain
	if c.OpenAIAPIKey == "" {
		if key, err := keyring.Get(keyring.OpenAIKey); err == nil {
			c.OpenAIAPIKey = key
		}
	}

	if c.AnthropicAPIKey == "" {
		if key, err := keyring.Get(keyring.AnthropicKey); err == nil {
			c.AnthropicAPIKey = key
		}
	}

	if c.OpenAIAPIKey == "" || c.AnthropicAPIKey == "" {
		return errors.New("missing API keys: set via environment variables or run 'voice config set-key'")
	}

	// Determine working paths
	workingName := getWorkingName(c.Name)

	// Ensure working directory exists
	if err := workdir.Prep(workingName); err != nil {
		return fmt.Errorf("failed to prepare working directory: %w", err)
	}

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

	outputPath, err := workdir.FilePath(workingName, workdir.MP3File)
	if err != nil {
		return fmt.Errorf("failed to determine output path: %w", err)
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
		WorkingName:     workingName,
		OpenAIAPIKey:    c.OpenAIAPIKey,
		AnthropicAPIKey: c.AnthropicAPIKey,
		Mode:            mode,
		MaxBytes:        c.MaxBytes,
		EditorCmd:       os.Getenv("MEMOS_EDITOR"),
	}

	ctrls := makeRecordingControls(ctx, dev, recorder, dataC, c.MaxBytes)
	p := tea.NewProgram(tui.New(config, ctrls))

	// Audio recorder goroutine (waits for channel close, MP3 conversion, cleanup)
	wg.Go(func() {
		if err := recorder.Start(ctx); err != nil {
			slog.Error("Audio recorder error", "error", err)
		}

		// Wait for recorder to finish (triggered by Finish() closing dataC)
		if err := recorder.Wait(); err != nil {
			slog.Error("Audio recorder error", "error", err)
		}

		p.Send(workflow.AudioFinalizingCompleteMsg{})
	})

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	wg.Wait()

	//nolint:forbidigo // CLI output for completion notification
	fmt.Println("\nfinished. bye!")

	return nil
}

// DevicesCmd lists available audio devices.
type DevicesCmd struct{}

// Run executes the devices command.
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

// ConfigCmd groups configuration-related subcommands.
type ConfigCmd struct {
	SetKey   SetKeyCmd   `cmd:"" help:"Store an API key in system keychain"`
	ListKeys ListKeysCmd `cmd:"" name:"list-keys" help:"Show which API keys are configured"`
}

// SetKeyCmd stores an API key in the system keychain.
type SetKeyCmd struct {
	Service string `arg:"" enum:"openai,anthropic" help:"Service name (openai or anthropic)"`
	Key     string `arg:"" help:"API key value"`
}

// Run executes the set-key command.
func (c *SetKeyCmd) Run() error {
	key, err := keyring.KeyFromServiceName(c.Service)
	if err != nil {
		return fmt.Errorf("invalid service: %w", err)
	}

	if err := keyring.Set(key, c.Key); err != nil {
		return fmt.Errorf("failed to store key: %w", err)
	}

	//nolint:forbidigo // CLI output
	fmt.Printf("%s API key stored in keychain\n", c.Service)

	return nil
}

// ListKeysCmd shows which API keys are configured.
type ListKeysCmd struct{}

// Run executes the list-keys command.
//
//nolint:unparam // error return required by Kong interface
func (c *ListKeysCmd) Run() error {
	allSet := true

	for _, key := range keyring.AllKeys() {
		if keyring.IsSet(key) {
			//nolint:forbidigo // CLI output
			fmt.Printf("%s: configured\n", key.DisplayName())
		} else {
			//nolint:forbidigo // CLI output
			fmt.Printf("%s: not set\n", key.DisplayName())
			allSet = false
		}
	}

	if !allSet {
		//nolint:forbidigo // CLI output
		fmt.Println("\nRun 'voice config set-key <service> <key>' to configure.")
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
	return "recording"
}

func makeRecordingControls(
	ctx context.Context,
	dev device.AudioDevice,
	recorder *audiofile.Recorder,
	dataC chan []byte,
	maxBytes int64,
) workflow.RecordingControls {
	return workflow.RecordingControls{
		StartStopPause: audioDevKnob{
			ctx: ctx,
			dev: dev,
		},
		FileSize: audioFileDial{
			ctx:      ctx,
			recorder: recorder,
			maxBytes: maxBytes,
		},
		SampleLevels: audioSampleLevels{
			recorder: recorder,
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

// audioSampleLevels implements uictl.Levels[int16] for waveform visualization.
type audioSampleLevels struct {
	recorder *audiofile.Recorder
}

// Read returns recent audio samples for visualization.
// Returns approximately 50ms of samples at 16kHz (800 samples).
func (asl audioSampleLevels) Read() []int16 {
	return asl.recorder.ReadSamples(800)
}
