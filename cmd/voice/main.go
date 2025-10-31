package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alkime/memos/internal/cli/audio"
	"github.com/alkime/memos/internal/cli/content"
	"github.com/alkime/memos/internal/cli/transcription"
)

// CLI defines the voice command structure
type CLI struct {
	Record     RecordCmd     `cmd:"" help:"Record audio from microphone"`
	Transcribe TranscribeCmd `cmd:"" help:"Transcribe audio file to text"`
	Process    ProcessCmd    `cmd:"" help:"Generate Hugo markdown from transcript"`
}

// RecordCmd handles audio recording
type RecordCmd struct {
	Output      string `arg:"" optional:"" help:"Output file path"`
	MaxDuration string `flag:"" default:"1h" help:"Max recording duration"`
	MaxBytes    int64  `flag:"" default:"268435456" help:"Max file size (256MB)"`
}

// Run executes the record command
func (r *RecordCmd) Run() error {
	// Determine output path
	outputPath := r.Output
	if outputPath == "" {
		// Default to ~/.memos/recordings/{timestamp}.wav
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		timestamp := time.Now().Format("2006-01-02-150405")
		outputPath = filepath.Join(homeDir, ".memos", "recordings", fmt.Sprintf("%s.wav", timestamp))
	}

	// Parse max duration
	maxDuration, err := time.ParseDuration(r.MaxDuration)
	if err != nil {
		return fmt.Errorf("invalid max duration: %w", err)
	}

	// Create recorder
	recorder := audio.NewRecorder(outputPath, maxDuration, r.MaxBytes)

	// Start recording
	if err := recorder.Start(); err != nil {
		return err
	}

	// Wait for stop condition
	fmt.Printf("Recording... Press Enter to stop. (Max: %s or %d MB)\n", r.MaxDuration, r.MaxBytes/(1024*1024))

	// Read from stdin for Enter key
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	// Stop recording
	if err := recorder.Stop(); err != nil {
		return err
	}

	fmt.Printf("Saved to: %s\n", outputPath)
	return nil
}

// TranscribeCmd handles audio transcription
type TranscribeCmd struct {
	AudioFile string `arg:"" help:"Path to audio file"`
	APIKey    string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key"`
	Output    string `flag:"" optional:"" help:"Output transcript path"`
}

// Run executes the transcribe command
func (t *TranscribeCmd) Run() error {
	// Validate API key
	if t.APIKey == "" {
		return fmt.Errorf("API key required: set OPENAI_API_KEY or use --api-key")
	}

	// Determine output path
	outputPath := t.Output
	if outputPath == "" {
		// Default to same name as audio file, .txt extension
		outputPath = t.AudioFile[:len(t.AudioFile)-len(filepath.Ext(t.AudioFile))] + ".txt"
	}

	// Create transcription client
	client := transcription.NewClient(t.APIKey)

	// Transcribe
	fmt.Println("Transcribing...")
	text, err := client.TranscribeFile(t.AudioFile)
	if err != nil {
		return err
	}

	// Write transcript
	if err := os.WriteFile(outputPath, []byte(text), 0644); err != nil {
		return err
	}

	fmt.Printf("Transcript saved to: %s\n", outputPath)
	return nil
}

// ProcessCmd handles markdown generation
type ProcessCmd struct {
	TranscriptFile string `arg:"" help:"Path to transcript text file"`
	Output         string `flag:"" optional:"" help:"Output markdown path"`
}

// Run executes the process command
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
	fmt.Println("Processing transcript...")
	if err := generator.GeneratePost(p.TranscriptFile, outputPath); err != nil {
		return err
	}

	fmt.Printf("Generated post: %s (draft)\n", outputPath)
	fmt.Println("Note: Raw transcript - Phase 2 will add AI cleanup")
	fmt.Println("Archived: Files moved to ~/.memos/archive/")
	return nil
}

func main() {
	cli := &CLI{}
	ctx := kong.Parse(cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
	os.Exit(0)
}
