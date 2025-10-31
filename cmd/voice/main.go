package main

import (
	"os"

	"github.com/alecthomas/kong"
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

// TranscribeCmd handles audio transcription
type TranscribeCmd struct {
	AudioFile string `arg:"" help:"Path to audio file"`
	APIKey    string `flag:"" env:"OPENAI_API_KEY" help:"OpenAI API key"`
	Output    string `flag:"" optional:"" help:"Output transcript path"`
}

// ProcessCmd handles markdown generation
type ProcessCmd struct {
	TranscriptFile string `arg:"" help:"Path to transcript text file"`
	Output         string `flag:"" optional:"" help:"Output markdown path"`
}

func main() {
	cli := &CLI{}
	ctx := kong.Parse(cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
	os.Exit(0)
}
